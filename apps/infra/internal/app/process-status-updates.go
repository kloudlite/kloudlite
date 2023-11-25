package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	distributionv1 "github.com/kloudlite/operator/apis/distribution/v1"
	wireguardv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/apps/infra/internal/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/kafka"
	t "kloudlite.io/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReceiveInfraUpdatesConsumer kafka.Consumer

func gvk(obj client.Object) string {
	val := obj.GetObjectKind().GroupVersionKind().String()
	return val
}

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

		gvkStr := obj.GetObjectKind().GroupVersionKind().String()

		clusterGVK := func() string {
			cluster := &clustersv1.Cluster{}
			cluster.EnsureGVK()
			return gvk(cluster)
		}()

		nodepoolGVK := func() string {
			np := &clustersv1.NodePool{}
			np.EnsureGVK()
			return gvk(np)
		}()

		deviceGVK := func() string {
			dev := &wireguardv1.Device{}
			dev.EnsureGVK()
			return gvk(dev)
		}()

		pvcGVK := func() string {
			return schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "PersistentVolumeClaim",
			}.String()
		}()

		buildRunGVK := func() string {
			brun := &distributionv1.BuildRun{}
			brun.EnsureGVK()
			return gvk(brun)
		}()

		switch gvkStr {
		case clusterGVK:
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
		case nodepoolGVK:
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
		case deviceGVK:
			{
				var device entities.VPNDevice
				if err := fn.JsonConversion(su.Object, &device); err != nil {
					return err
				}
				if v, ok := su.Object["resource-watcher-wireguard-config"]; ok {
					b, err := json.Marshal(v)
					if err != nil {
						return err
					}
					var encodedStr t.EncodedString
					if err := json.Unmarshal(b, &encodedStr); err != nil {
						return err
					}
					device.WireguardConfig = encodedStr
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnVPNDeviceDeleteMessage(dctx, su.ClusterName, device)
				}
				return d.OnVPNDeviceUpdateMessage(dctx, su.ClusterName, device)
			}
		case pvcGVK:
			{
				var pvc entities.PersistentVolumeClaim
				if err := fn.JsonConversion(su.Object, &pvc); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnPVCDeleteMessage(dctx, su.ClusterName, pvc)
				}
				return d.OnPVCUpdateMessage(dctx, su.ClusterName, pvc)
			}
		case buildRunGVK:
			{
				var buildRun entities.BuildRun
				if err := fn.JsonConversion(su.Object, &buildRun); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnBuildRunDeleteMessage(dctx, su.ClusterName, buildRun)
				}
				return d.OnBuildRunUpdateMessage(dctx, su.ClusterName, buildRun)
			}
		default:
			{
				mLogger.Infof("infra status updates consumer does not acknowledge the gvk %s", gvk(&obj))
				return nil
			}
		}
	})
}
