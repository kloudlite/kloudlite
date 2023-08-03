package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/apps/infra/internal/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type InfraUpdatesConsumer redpanda.Consumer

func processInfraUpdates(consumer InfraUpdatesConsumer, d domain.Domain, logger logging.Logger) {
	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
		logger.Debugf("processing offset %d timestamp %s", offset, timeStamp)

		var su types.ResourceUpdate
		if err := json.Unmarshal(msg, &su); err != nil {
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
			"accountName", su.AccountName,
			"clusterName", su.ClusterName,
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		ctx := domain.InfraContext{Context: context.TODO(), AccountName: su.AccountName}

		kind := obj.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case "Cluster":
			{
				var clus entities.Cluster
				if err := fn.JsonConversion(su.Object, &clus); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteClusterMessage(ctx, clus)
				}
				return d.OnUpdateClusterMessage(ctx, clus)
			}
		case "NodePool":
			{
				var np entities.NodePool
				if err := fn.JsonConversion(su.Object, &np); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteNodePoolMessage(ctx, su.ClusterName, np)
				}
				return d.OnUpdateNodePoolMessage(ctx, su.ClusterName, np)
			}
		default:
			{
				mLogger.Infof("infra status updates consumer does not acknowledge the kind %s", kind)
				return nil
			}
		}
	})
}
