package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	t "github.com/kloudlite/operator/agent/types"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type ApplyOnErrorConsumer redpanda.Consumer

func ProcessApplyOnError(consumer ApplyOnErrorConsumer, d domain.Domain, logr logging.Logger) {
	counter := 0
	logger := logr.WithName("apply-on-error")
	consumer.StartConsuming(func(m []byte, timeStamp time.Time, offset int64) error {
		counter += 1
		logger.Debugf("received message [%d]", counter)
		var msg t.AgentErrMessage
		if err := json.Unmarshal(m, &msg); err != nil {
			return err
		}

		mLogger := logger.WithKV(
			"gvk", msg.Object.GetObjectKind().GroupVersionKind(),
			"accountName", msg.AccountName,
			"clusterName", msg.ClusterName,
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		kind := msg.Object.GetObjectKind().GroupVersionKind().Kind
		ctx := domain.NewConsoleContext(context.TODO(), msg.AccountName, msg.ClusterName)

		switch kind {
		case "Project":
			{
				return d.OnApplyProjectError(ctx, msg.Error, msg.Object.GetName())
			}
		case "App":
			{
				return d.OnApplyAppError(ctx, msg.Error, msg.Object.GetNamespace(), msg.Object.GetName())
			}
		case "Config":
			{
				return d.OnApplyConfigError(ctx, msg.Error, msg.Object.GetNamespace(), msg.Object.GetName())
			}
		case "Secret":
			{
				return d.OnApplySecretError(ctx, msg.Error, msg.Object.GetNamespace(), msg.Object.GetName())
			}
		case "Router":
			{
				return d.OnApplyRouterError(ctx, msg.Error, msg.Object.GetNamespace(), msg.Object.GetName())
			}
		case "ManagedService":
			{
				return d.OnApplyManagedServiceError(ctx, msg.Error, msg.Object.GetNamespace(), msg.Object.GetName())
			}
		case "ManagedResource":
			{
				return d.OnApplyManagedResourceError(ctx, msg.Error, msg.Object.GetNamespace(), msg.Object.GetName())
			}
		default:
			{
				return fmt.Errorf("console apply error reader does not acknowledge resource with kind (%s)", kind)
			}
		}
	})
}
