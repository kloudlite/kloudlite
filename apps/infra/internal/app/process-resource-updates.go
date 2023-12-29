package app

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/utils/strings/slices"
	"strings"
	"time"

	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
	t "github.com/kloudlite/api/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReceiveResourceUpdatesConsumer messaging.Consumer

func gvk(obj client.Object) string {
	val := obj.GetObjectKind().GroupVersionKind().String()
	return val
}

var (
	clusterGVK     = fn.GVK("clusters.kloudlite.io/v1", "Cluster")
	nodepoolGVK    = fn.GVK("clusters.kloudlite.io/v1", "NodePool")
	helmreleaseGVK = fn.GVK("crds.kloudlite.io/v1", "HelmChart")
	deviceGVK      = fn.GVK("wireguard.kloudlite.io/v1", "Device")
	pvcGVK         = fn.GVK("v1", "PersistentVolumeClaim")
	buildrunGVK    = fn.GVK("distribution.kloudlite.io/v1", "BuildRun")
	clusterMsvcGVK = fn.GVK("clusters.kloudlite.io/v1", "ClusterManagedService")
)

func processResourceUpdates(consumer ReceiveResourceUpdatesConsumer, d domain.Domain, logger logging.Logger) {
	readMsg := func(msg *msgTypes.ConsumeMsg) error {
		logger.Debugf("processing msg timestamp %s", msg.Timestamp.Format(time.RFC3339))

		var su types.ResourceUpdate
		if err := json.Unmarshal(msg.Payload, &su); err != nil {
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
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		dctx := domain.InfraContext{Context: context.TODO(), UserId: "sys-user-process-infra-updates", AccountName: su.AccountName}

		gvkStr := obj.GetObjectKind().GroupVersionKind().String()

		switch gvkStr {
		case clusterGVK.String():
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
		case nodepoolGVK.String():
			{
				var np entities.NodePool
				if err := fn.JsonConversion(su.Object, &np); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteNodePoolMessage(dctx, su.ClusterName, np)
				}
				return d.OnUpdateNodePoolMessage(dctx, su.ClusterName, np)
			}
		case deviceGVK.String():
			{
				var device entities.VPNDevice
				if err := fn.JsonConversion(su.Object, &device); err != nil {
					return errors.NewE(err)
				}
				if v, ok := su.Object["resource-watcher-wireguard-config"]; ok {
					b, err := json.Marshal(v)
					if err != nil {
						return errors.NewE(err)
					}
					var encodedStr t.EncodedString
					if err := json.Unmarshal(b, &encodedStr); err != nil {
						return errors.NewE(err)
					}
					device.WireguardConfig = encodedStr
				}
				if obj.GetDeletionTimestamp() != nil && (!slices.Contains(obj.GetFinalizers(), "finalizers.kloudlite.io/watch") && !slices.Contains(obj.GetFinalizers(), "finalizers.kloudlite.io/status-watcher")) {
					return d.OnVPNDeviceDeleteMessage(dctx, su.ClusterName, device)
				}
				return d.OnVPNDeviceUpdateMessage(dctx, su.ClusterName, device)
			}
		case pvcGVK.String():
			{
				var pvc entities.PersistentVolumeClaim
				if err := fn.JsonConversion(su.Object, &pvc); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnPVCDeleteMessage(dctx, su.ClusterName, pvc)
				}
				return d.OnPVCUpdateMessage(dctx, su.ClusterName, pvc)
			}
			// case buildrunGVK.String():
			// 	{
			// 		var buildRun entities.BuildRun
			// 		if err := fn.JsonConversion(su.Object, &buildRun); err != nil {
			// 			return errors.NewE(err)
			// 		}
			// 		if obj.GetDeletionTimestamp() != nil {
			// 			return d.OnBuildRunDeleteMessage(dctx, su.ClusterName, buildRun)
			// 		}
			// 		return d.OnBuildRunUpdateMessage(dctx, su.ClusterName, buildRun)
			// 	}

		case helmreleaseGVK.String():
			{
				var hr entities.HelmRelease
				if err := fn.JsonConversion(su.Object, &hr); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnHelmReleaseDeleteMessage(dctx, su.ClusterName, hr)
				}
				return d.OnHelmReleaseUpdateMessage(dctx, su.ClusterName, hr)
			}

		case clusterMsvcGVK.String():
			{
				var svc entities.ClusterManagedService
				if err := fn.JsonConversion(su.Object, &svc); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnClusterManagedServiceDeleteMessage(dctx, su.ClusterName, svc)
				}
				return d.OnClusterManagedServiceUpdateMessage(dctx, su.ClusterName, svc)
			}

		default:
			{
				mLogger.Infof("infra status updates consumer does not acknowledge the gvk %s", gvk(&obj))
				return nil
			}
		}
	}

	if err := consumer.Consume(readMsg, msgTypes.ConsumeOpts{
		OnError: func(err error) error {
			logger.Errorf(err, "error while consuming message")
			return nil
		},
	}); err != nil {
		logger.Errorf(err, "error while consuming messages")
	}
}
