package app

// import (
// 	"context"
// 	"encoding/json"
// 	"strings"
// 	"time"
//
// 	clusterv1 "github.com/kloudlite/operator/apis/clusters/v1"
// 	"github.com/kloudlite/operator/operators/resource-watcher/types"
// 	"kloudlite.io/apps/infra/internal/domain"
// 	"kloudlite.io/pkg/logging"
// 	"kloudlite.io/pkg/redpanda"
// )

// type ByocClientUpdatesConsumer redpanda.Consumer
//
// func processByocClientUpdates(consumer ByocClientUpdatesConsumer, d domain.DomainName, logger logging.Logger) {
// 	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
// 		logger.Debugf("processing offset %d timestamp %s", offset, timeStamp)
//
// 		var su types.ResourceUpdate
// 		if err := json.Unmarshal(msg, &su); err != nil {
// 			logger.Errorf(err, "parsing into status update")
// 			return nil
// 			// return err
// 		}
//
// 		if len(strings.TrimSpace(su.AccountName)) == 0 {
// 			logger.Infof("message does not contain 'accountName', so won't be able to find a resource uniquely, thus ignoring ...")
// 			return nil
// 		}
//
// 		b, err := json.Marshal(su.Object)
// 		if err != nil {
// 			return err
// 		}
// 		var obj clusterv1.BYOC
// 		if err := json.Unmarshal(b, &obj); err != nil {
// 			return err
// 		}
//
// 		// obj := unstructured.Unstructured{Object: su.Object}
//
// 		mLogger := logger.WithKV(
// 			"gvk", obj.GetObjectKind().GroupVersionKind().String(),
// 			"accountName", su.AccountName,
// 			"clusterName", su.ClusterName,
// 		)
//
// 		mLogger.Infof("received message")
// 		defer func() {
// 			mLogger.Infof("processed message")
// 		}()
//
// 		ctx := domain.InfraContext{Context: context.TODO(), AccountName: su.AccountName}
// 		byocCluster, err := d.GetBYOCCluster(ctx, su.ClusterName)
// 		if err != nil {
// 			logger.Infof("error: %s, commiting anyway ...", err.Error())
// 			return nil
// 		}
//
// 		if byocCluster == nil {
// 			return nil
// 		}
//
// 		if byocCluster.Generation > obj.Generation {
// 			// this message belongs to previous generation, so we can ignore it
// 			return nil
// 		}
//
// 		byocCluster.Spec = obj.Spec
// 		byocCluster.Status = obj.Status
//
// 		if err := d.OnBYOCClusterHelmUpdates(ctx, *byocCluster); err != nil {
// 			return err
// 		}
//
// 		return nil
// 	})
// }
