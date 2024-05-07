package app

import (
	"context"
	"encoding/json"
	"fmt"
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
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkingv1 "k8s.io/api/networking/v1"
)

type ReceiveResourceUpdatesConsumer messaging.Consumer

func gvk(obj client.Object) string {
	val := obj.GetObjectKind().GroupVersionKind().String()
	return val
}

var (
	clusterGVK = fn.GVK("clusters.kloudlite.io/v1", "Cluster")
	// clusterConnGVK      = fn.GVK("wireguard.kloudlite.io/v1", "ClusterConnection")
	// globalVpnGVK        = fn.GVK("wireguard.kloudlite.io/v1", "GlobalVPNConnection")
	globalVpnGVK        = fn.GVK("wireguard.kloudlite.io/v1", "GlobalVPN")
	nodepoolGVK         = fn.GVK("clusters.kloudlite.io/v1", "NodePool")
	helmreleaseGVK      = fn.GVK("crds.kloudlite.io/v1", "HelmChart")
	pvcGVK              = fn.GVK("v1", "PersistentVolumeClaim")
	pvGVK               = fn.GVK("v1", "PersistentVolume")
	volumeAttachmentGVK = fn.GVK("storage.k8s.io/v1", "VolumeAttachment")
	namespaceGVK        = fn.GVK("v1", "Namespace")
	clusterMsvcGVK      = fn.GVK("crds.kloudlite.io/v1", "ClusterManagedService")
	ingressGVK          = fn.GVK("networking.k8s.io/v1", "Ingress")
	secretGVK           = fn.GVK("v1", "Secret")
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

		dctx := domain.InfraContext{Context: context.TODO(), UserId: "sys-user-process-infra-updates", AccountName: su.AccountName}

		obj := unstructured.Unstructured{Object: su.Object}
		gvkStr := obj.GetObjectKind().GroupVersionKind().String()

		resStatus, err := func() (types.ResourceStatus, error) {
			v, ok := su.Object[types.ResourceStatusKey]
			if !ok {
				return "", errors.NewE(fmt.Errorf("field %s not found in object", types.ResourceStatusKey))
			}
			s, ok := v.(string)
			if !ok {
				return "", errors.NewE(fmt.Errorf("field value %v is not a string", v))
			}

			return types.ResourceStatus(s), nil
		}()
		if err != nil {
			return err
		}

		mLogger := logger.WithKV(
			"gvk", obj.GetObjectKind().GroupVersionKind(),
			"NN", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
			"resource-status", resStatus,
			"accountName/clusterName", fmt.Sprintf("%s/%s", su.AccountName, su.ClusterName),
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		switch gvkStr {
		case clusterGVK.String():
			{
				var clus entities.Cluster
				if err := fn.JsonConversion(su.Object, &clus); err != nil {
					return err
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnClusterDeleteMessage(dctx, clus)
				}
				return d.OnClusterUpdateMessage(dctx, clus, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}
		case globalVpnGVK.String():
			{
				var gvpn entities.GlobalVPNConnection
				if err := fn.JsonConversion(su.Object, &gvpn); err != nil {
					return errors.NewE(err)
				}

				if v, ok := su.Object[types.KeyGlobalVPNWgParams]; ok {
					wp, err := fn.JsonConvertP[wgv1.WgParams](v)
					if err != nil {
						return errors.NewE(err)
					}
					gvpn.ParsedWgParams = wp
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnGlobalVPNConnectionDeleteMessage(dctx, su.ClusterName, gvpn)
				}
				return d.OnGlobalVPNConnectionUpdateMessage(dctx, su.ClusterName, gvpn, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}
		case nodepoolGVK.String():
			{
				var np entities.NodePool
				if err := fn.JsonConversion(su.Object, &np); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnNodePoolDeleteMessage(dctx, su.ClusterName, np)
				}
				return d.OnNodePoolUpdateMessage(dctx, su.ClusterName, np, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}
		case pvcGVK.String():
			{
				var pvc entities.PersistentVolumeClaim
				if err := fn.JsonConversion(su.Object, &pvc); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnPVCDeleteMessage(dctx, su.ClusterName, pvc)
				}
				return d.OnPVCUpdateMessage(dctx, su.ClusterName, pvc, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}

		case pvGVK.String():
			{
				var pv entities.PersistentVolume
				if err := fn.JsonConversion(su.Object, &pv); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnPVDeleteMessage(dctx, su.ClusterName, pv)
				}
				return d.OnPVUpdateMessage(dctx, su.ClusterName, pv, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}

		case volumeAttachmentGVK.String():
			{
				var volatt entities.VolumeAttachment
				if err := fn.JsonConversion(su.Object, &volatt); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnVolumeAttachmentDeleteMessage(dctx, su.ClusterName, volatt)
				}
				return d.OnVolumeAttachmentUpdateMessage(dctx, su.ClusterName, volatt, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}

		case helmreleaseGVK.String():
			{
				var hr entities.HelmRelease
				if err := fn.JsonConversion(su.Object, &hr); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnHelmReleaseDeleteMessage(dctx, su.ClusterName, hr)
				}
				return d.OnHelmReleaseUpdateMessage(dctx, su.ClusterName, hr, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}

		case namespaceGVK.String():
			{
				var ns entities.Namespace

				if err := fn.JsonConversion(su.Object, &ns); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnNamespaceDeleteMessage(dctx, su.ClusterName, ns)
				}
				return d.OnNamespaceUpdateMessage(dctx, su.ClusterName, ns, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}

		case clusterMsvcGVK.String():
			{
				var cmsvc entities.ClusterManagedService
				if err := fn.JsonConversion(su.Object, &cmsvc); err != nil {
					return errors.NewE(err)
				}

				if v, ok := su.Object[types.KeyClusterManagedSvcSecret]; ok {
					if v2, ok := v.(*corev1.Secret); ok {
						cmsvc.SyncedOutputSecretRef = v2
					}
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnClusterManagedServiceDeleteMessage(dctx, su.ClusterName, cmsvc)
				}
				return d.OnClusterManagedServiceUpdateMessage(dctx, su.ClusterName, cmsvc, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}

		case ingressGVK.String():
			{
				var ingress networkingv1.Ingress
				if err := fn.JsonConversion(su.Object, &ingress); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnIngressDeleteMessage(dctx, su.ClusterName, ingress)
				}
				return d.OnIngressUpdateMessage(dctx, su.ClusterName, ingress, resStatus, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}

		case secretGVK.String():
			{
				var secret corev1.Secret
				if err := fn.JsonConversion(su.Object, &secret); err != nil {
					return errors.NewE(err)
				}

				if secret.Name != "byok-kubeconfig" {
					return nil
				}

				if resStatus == types.ResourceStatusDeleted {
					// FIXME: not implemented for now
					return nil
				}
				return d.UpsertBYOKClusterKubeconfig(dctx, su.ClusterName, secret.Data["kubeconfig"])
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
