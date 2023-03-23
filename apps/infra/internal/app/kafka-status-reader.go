package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/status-n-billing/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/apps/infra/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type StatusUpdatesConsumer redpanda.Consumer

func processStatusUpdates(consumer StatusUpdatesConsumer, d domain.Domain, logger logging.Logger) {
	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
		logger.Debugf("processing offset %d timestamp %s", offset, timeStamp)

		var su types.StatusUpdate
		if err := json.Unmarshal(msg, &su); err != nil {
			logger.Errorf(err, "parsing into status update")
			return nil
			// return err
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
		case "CloudProvider":
			{
				var cp entities.CloudProvider
				if err := fn.JsonConversion(su.Object, &cp); err != nil {
					return err
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteCloudProviderMessage(ctx, cp)
				}
				return d.OnUpdateCloudProviderMessage(ctx, cp)
			}
		case "Edge":
			{
				var edge entities.Edge
				if err := fn.JsonConversion(su.Object, &edge); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteEdgeMessage(ctx, edge)
				}
				return d.OnUpdateEdgeMessage(ctx, edge)
			}
		case "WorkerNode":
			{
				var wNode entities.WorkerNode
				if err := fn.JsonConversion(su.Object, &wNode); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteWorkerNodeMessage(ctx, wNode)
				}
				return d.OnUpdateWorkerNodeMessage(ctx, wNode)
			}
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
		default:
			{
				mLogger.Infof("infra status updates consumer does not acknowledge the kind %s", kind)
				return nil
			}
		}
	})
}
