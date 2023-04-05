package app

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/status-n-billing/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type ByocHelmStatusUpdates redpanda.Consumer

func processByocHelmUpdates(consumer ByocHelmStatusUpdates, d domain.Domain, logger logging.Logger) {
	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
		logger.Debugf("processing offset %d timestamp %s", offset, timeStamp)

		var su types.StatusUpdate
		if err := json.Unmarshal(msg, &su); err != nil {
			logger.Errorf(err, "parsing into status update")
			return nil
			// return err
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
			"accountName", su.AccountName,
			"clusterName", su.ClusterName,
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		return nil
	})
}
