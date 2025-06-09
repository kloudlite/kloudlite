package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/operator/operators/resource-watcher/types"

	msgOfficeT "github.com/kloudlite/api/apps/message-office/types"
	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ReceiveResourceUpdatesConsumer messaging.Consumer

func gvk(obj client.Object) string {
	val := obj.GetObjectKind().GroupVersionKind().String()
	return val
}

var buildRunGVK = fn.GVK("distribution.kloudlite.io/v1", "BuildRun")

func processResourceUpdates(consumer ReceiveResourceUpdatesConsumer, d domain.Domain, logger logging.Logger) {
	readMsg := func(msg *msgTypes.ConsumeMsg) error {
		logger.Debugf("processing msg timestamp %s", msg.Timestamp.Format(time.RFC3339))

		ru, err := msgOfficeT.UnmarshalResourceUpdate(msg.Payload)
		if err != nil {
			logger.Errorf(err, "unmarshaling resource update")
			return nil
		}

		var su types.ResourceUpdate
		if err := json.Unmarshal(msg.Payload, &su); err != nil {
			logger.Errorf(err, "parsing into status update")
			return nil
		}

		if su.Object == nil {
			logger.Infof("message does not contain 'object', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		if len(strings.TrimSpace(ru.AccountName)) == 0 {
			logger.Infof("message does not contain 'accountName', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		obj := su.Object
		mLogger := logger.WithKV(
			"gvk", obj.GetObjectKind().GroupVersionKind(),
			"accountName/clusterName", fmt.Sprintf("%s/%s", ru.AccountName, ru.ClusterName),
		)

		resStatus, err := func() (types.ResourceStatus, error) {
			v, ok := obj.Object[types.ResourceStatusKey]
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

		opts := domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp}

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		dctx := domain.RegistryContext{Context: context.TODO(), UserId: "sys-user-process-infra-updates", AccountName: ru.AccountName}

		gvkStr := obj.GetObjectKind().GroupVersionKind().String()

		switch gvkStr {
		case buildRunGVK.String():
			{
				var buildRun entities.BuildRun
				if err := fn.JsonConversion(su.Object, &buildRun); err != nil {
					return err
				}

				buildRun.AccountName = ru.AccountName
				buildRun.ClusterName = ru.ClusterName

				if obj.GetDeletionTimestamp() != nil {
					return d.OnBuildRunDeleteMessage(dctx, buildRun)
				}

				return d.OnBuildRunUpdateMessage(dctx, buildRun, resStatus, opts)
			}

		default:
			{
				mLogger.Infof("container registry status updates consumer does not acknowledge the gvk %s", gvk(obj))
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
