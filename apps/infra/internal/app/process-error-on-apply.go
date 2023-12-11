package app

import (
	"context"
	"encoding/json"
	"fmt"

	t "github.com/kloudlite/operator/agent/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/messaging/types"
)

type ErrorOnApplyConsumer messaging.Consumer

func ProcessErrorOnApply(consumer ErrorOnApplyConsumer, logger logging.Logger, d domain.Domain) {
	counter := 0
	processMsg := func(msg *types.ConsumeMsg) error {
		counter += 1

		var errMsg t.AgentErrMessage
		if err := json.Unmarshal(msg.Payload, &errMsg); err != nil {
			return err
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
				return d.OnNodepoolApplyError(dctx, errMsg.ClusterName, obj.GetName(), errMsg.Error)
			}
		case deviceGVK.String():
			{
				return d.OnVPNDeviceApplyError(dctx, errMsg.ClusterName, obj.GetName(), errMsg.Error)
			}
		default:
			{
				return fmt.Errorf("infra error-on-apply reader does not acknowledge resource with kind (%s)", kind)
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
