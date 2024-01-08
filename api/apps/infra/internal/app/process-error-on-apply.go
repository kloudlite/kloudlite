package app

import (
	"context"
	"encoding/json"

	"github.com/kloudlite/api/apps/infra/internal/domain"
	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"github.com/kloudlite/api/pkg/errors"
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

		var errMsg t.AgentErrMessage
		if err := json.Unmarshal(msg.Payload, &errMsg); err != nil {
			return errors.NewE(err)
		}

		obj := unstructured.Unstructured{Object: errMsg.Object}

		mLogger := logger.WithKV(
			"gvk", obj.GroupVersionKind(),
			"accountName", errMsg.AccountName,
			"clusterName", errMsg.ClusterName,
		)

		mLogger.Infof("[%d] received message", counter)
		defer func() {
			mLogger.Infof("[%d] processed message", counter)
		}()

		kind := obj.GroupVersionKind().Kind
		dctx := domain.InfraContext{
			Context:     context.TODO(),
			UserId:      "sys-user:error-on-apply-worker",
			UserEmail:   "",
			UserName:    "",
			AccountName: errMsg.AccountName,
		}

		switch obj.GroupVersionKind().String() {
		case nodepoolGVK.String():
			{
				return d.OnNodepoolApplyError(dctx, errMsg.ClusterName, obj.GetName(), errMsg.Error, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}
		case deviceGVK.String():
			{
				return d.OnVPNDeviceApplyError(dctx, errMsg.ClusterName, obj.GetName(), errMsg.Error, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}
		case clusterMsvcGVK.String():
			{
				return d.OnClusterManagedServiceApplyError(dctx, errMsg.ClusterName, obj.GetName(), errMsg.Error, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}
		case helmreleaseGVK.String():
			{
				return d.OnHelmReleaseApplyError(dctx, errMsg.ClusterName, obj.GetName(), errMsg.Error, domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp})
			}
		default:
			{
				return errors.Newf("infra error-on-apply reader does not acknowledge resource with kind (%s)", kind)
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
