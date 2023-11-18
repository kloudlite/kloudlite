package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/apps/infra/internal/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/kafka"
)

type ReceiveInfraUpdatesConsumer kafka.Consumer

func processInfraUpdates(consumer ReceiveInfraUpdatesConsumer, d domain.Domain) {
	consumer.StartConsuming(func(ctx kafka.ConsumerContext, topic string, value []byte, metadata kafka.RecordMetadata) error {
		logger := ctx.Logger
		logger.Debugf("processing msg timestamp %s", metadata.Timestamp.Format(time.RFC3339))

		var su types.ResourceUpdate
		if err := json.Unmarshal(value, &su); err != nil {
			logger.Errorf(err, "parsing into status update")
			return nil
		}

		if su.Object == nil {
			logger.Infof("message does not contain 'object', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		if len(strings.TrimSpace(su.AccountName)) == 0 {
			logger.Infof("message does not contain 'accountName', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		obj := unstructured.Unstructured{Object: su.Object}
		mLogger := logger.WithKV(
			"gvk", obj.GetObjectKind().GroupVersionKind(),
			"accountName/clusterName", fmt.Sprintf("%s/%s", su.AccountName, su.ClusterName),
			"partition/offset", fmt.Sprintf("%d/%d", metadata.Partition, metadata.Offset),
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		dctx := domain.InfraContext{Context: ctx, UserId: "sys-user-process-infra-updates", AccountName: su.AccountName}

		kind := obj.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case "Cluster":
			{
				var clus entities.Cluster
				if err := fn.JsonConversion(su.Object, &clus); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteClusterMessage(dctx, clus)
				}
				return d.OnUpdateClusterMessage(dctx, clus)
			}
		case "NodePool":
			{
				var np entities.NodePool
				if err := fn.JsonConversion(su.Object, &np); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteNodePoolMessage(dctx, su.ClusterName, np)
				}
				return d.OnUpdateNodePoolMessage(dctx, su.ClusterName, np)
			}
		case "VPNDevice":
			{
				var device entities.VPNDevice
				if err := fn.JsonConversion(su.Object, &device); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnVPNDeviceDeleteMessage(dctx, su.ClusterName, device)
				}
				return d.OnVPNDeviceUpdateMessage(dctx, su.ClusterName, device)
			}
		default:
			{
				mLogger.Infof("infra status updates consumer does not acknowledge the kind %s", kind)
				return nil
			}
		}
	})
}
