package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kloudlite/operator/operators/status-n-billing/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/apps/infra/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

func processStatusUpdates(consumer redpanda.Consumer, domain domain.Domain, logger logging.Logger) {
	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
		logger.Infof("processing offset %d timestamp %s", offset, timeStamp)
		logger.Debugf("%s", msg)

		var statusUpdate types.StatusUpdate
		if err := json.Unmarshal(msg, &statusUpdate); err != nil {
			logger.Errorf(err, "parsing into status update")
			return nil
			// return err
		}

		obj := unstructured.Unstructured{Object: statusUpdate.Object}

		kind := obj.GetObjectKind().GroupVersionKind().Kind
		switch kind {
		case "CloudProvider":
			{
				var cp entities.CloudProvider
				if err := fn.JsonConversion(statusUpdate.Object, &cp); err != nil {
					return err
				}

				if obj.GetDeletionTimestamp() != nil {
					return domain.OnDeleteCloudProviderMessage(context.TODO(), cp)
				}
				return domain.OnUpdateCloudProviderMessage(context.TODO(), cp)
			}
		default:
			{
				logger.Infof("infra status updates consumer does not acknowledge the kind %s", kind)
			}
		}

		return nil
	})
}
