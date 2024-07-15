package app

import (
	"context"
	"encoding/json"

	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	msgOfficeT "github.com/kloudlite/api/apps/message-office/types"
	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	"github.com/kloudlite/api/pkg/messaging/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ErrorOnApplyConsumer messaging.Consumer

func ProcessErrorOnApply(consumer ErrorOnApplyConsumer, logger logging.Logger, d domain.Domain) {
	counter := 0
	processMsg := func(msg *types.ConsumeMsg) error {
		counter += 1

		em, err := msgOfficeT.UnmarshalErrMessage(msg.Payload)
		if err != nil {
			return errors.NewE(err)
		}

		var errObj t.AgentErrMessage
		if err := json.Unmarshal(em.Error, &errObj); err != nil {
			return errors.NewE(err)
		}

		obj := unstructured.Unstructured{Object: errObj.Object}

		mLogger := logger.WithKV(
			"gvk", obj.GroupVersionKind(),
			"accountName", em.AccountName,
			"clusterName", em.ClusterName,
		)

		mLogger.Infof("[%d] received message", counter)
		defer func() {
			mLogger.Infof("[%d] processed message", counter)
		}()

		dctx := domain.InfraContext{
			Context:     context.TODO(),
			UserId:      "sys-user:error-on-apply-worker",
			UserEmail:   "",
			UserName:    "",
			AccountName: em.AccountName,
		}

		opts := domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp}

		gvkstr := obj.GroupVersionKind().String()
		switch gvkstr {
		case globalVpnGVK.String():
			{
				cc, err := fn.JsonConvert[entities.GlobalVPNConnection](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnGlobalVPNConnectionApplyError(dctx, em.ClusterName, obj.GetName(), errObj.Error, opts)
				}
				return d.OnGlobalVPNConnectionDeleteMessage(dctx, em.ClusterName, cc)
			}
		case nodepoolGVK.String():
			{
				nodepool, err := fn.JsonConvert[entities.NodePool](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnNodepoolApplyError(dctx, em.ClusterName, obj.GetName(), errObj.Error, opts)
				}
				return d.OnNodePoolDeleteMessage(dctx, em.ClusterName, nodepool)
			}
		case helmreleaseGVK.String():
			{
				helmRelease, err := fn.JsonConvert[entities.HelmRelease](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnHelmReleaseApplyError(dctx, em.ClusterName, obj.GetName(), errObj.Error, opts)
				}
				return d.OnHelmReleaseDeleteMessage(dctx, em.ClusterName, helmRelease)
			}
		default:
			{
				return errors.Newf("infra error-on-apply reader does not acknowledge resource with gvk (%s)", gvkstr)
			}
		}
	}

	if err := consumer.Consume(processMsg, types.ConsumeOpts{
		OnError: func(err error) error {
			return nil
		},
	}); err != nil {
		logger.Errorf(err, "when setting up error-on-apply consumer")
	}
}
