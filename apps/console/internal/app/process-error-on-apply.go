package app

import (
	"encoding/json"
	"fmt"
	t "github.com/kloudlite/operator/agent/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/pkg/kafka"
)

type ErrorOnApplyConsumer kafka.Consumer

func ProcessErrorOnApply(consumer ErrorOnApplyConsumer, d domain.Domain) {
	counter := 0
	consumer.StartConsuming(func(ctx kafka.ConsumerContext, topic string, value []byte, metadata kafka.RecordMetadata) error {
		counter += 1
		logger := ctx.Logger
		logger.Debugf("received message [%d]", counter)
		var msg t.AgentErrMessage
		if err := json.Unmarshal(value, &msg); err != nil {
			return err
		}

		obj := unstructured.Unstructured{Object: msg.Object}

		mLogger := logger.WithKV(
			"gvk", obj.GroupVersionKind(),
			"accountName", msg.AccountName,
			"clusterName", msg.ClusterName,
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		kind := obj.GroupVersionKind().Kind
		dctx := domain.NewConsoleContext(ctx, "sys-user:apply-on-error-worker", msg.AccountName, msg.ClusterName)

		switch kind {
		case "Project":
			{
				return d.OnApplyProjectError(dctx, msg.Error, obj.GetName())
			}
		case "Env":
			{
				return d.OnApplyWorkspaceError(dctx, msg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "App":
			{
				return d.OnApplyAppError(dctx, msg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "Config":
			{
				return d.OnApplyConfigError(dctx, msg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "Secret":
			{
				return d.OnApplySecretError(dctx, msg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "Router":
			{
				return d.OnApplyRouterError(dctx, msg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "ManagedService":
			{
				return d.OnApplyManagedServiceError(dctx, msg.Error, obj.GetNamespace(), obj.GetName())
			}
		case "ManagedResource":
			{
				return d.OnApplyManagedResourceError(dctx, msg.Error, obj.GetNamespace(), obj.GetName())
			}
		default:
			{
				return fmt.Errorf("console apply error reader does not acknowledge resource with kind (%s)", kind)
			}
		}
	})
}
