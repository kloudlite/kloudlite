package app

import (
	"context"
	"encoding/json"
	"fmt"

	t "github.com/kloudlite/operator/agent/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/messaging"
	msgTypes "kloudlite.io/pkg/messaging/types"
)

type ErrorOnApplyConsumer messaging.Consumer

func ProcessErrorOnApply(consumer ErrorOnApplyConsumer, d domain.Domain, logger logging.Logger) {
	counter := 0

	msgReader := func(msg *msgTypes.ConsumeMsg) error {
		counter += 1
		logger.Debugf("received message [%d]", counter)
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

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		kind := obj.GroupVersionKind().Kind
		dctx := domain.NewConsoleContext(context.TODO(), "sys-user:apply-on-error-worker", errMsg.AccountName, errMsg.ClusterName)

		switch kind {
		case "Project":
			{
				return d.OnApplyProjectError(dctx, errMsg.Error, obj.GetName())
			}
		case "Env":
			{
				return d.OnApplyWorkspaceError(dctx, errMsg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "App":
			{
				return d.OnApplyAppError(dctx, errMsg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "Config":
			{
				return d.OnApplyConfigError(dctx, errMsg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "Secret":
			{
				return d.OnApplySecretError(dctx, errMsg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "Router":
			{
				return d.OnApplyRouterError(dctx, errMsg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "ManagedService":
			{
				return d.OnApplyManagedServiceError(dctx, errMsg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "ManagedResource":
			{
				return d.OnApplyManagedResourceError(dctx, errMsg.Error, obj.GetNamespace(), obj.GetName())
			}
		default:
			{
				return fmt.Errorf("console apply error reader does not acknowledge resource with kind (%s)", kind)
			}
		}
	}

	consumer.Consume(msgReader, msgTypes.ConsumeOpts{
		OnError: func(err error) error {
			logger.Errorf(err, "received while reading messages, ignoring it")
			return nil
		},
	})
}
