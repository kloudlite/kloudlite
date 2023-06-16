package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type ResourceUpdateConsumer redpanda.Consumer

func ProcessResourceUpdates(consumer ResourceUpdateConsumer, d domain.Domain, logr logging.Logger) {
	counter := 0
	logger := logr.WithName("resource-updates")
	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
		counter += 1
		logger.Debugf("[%d] received message", counter)

		var ru types.ResourceUpdate
		if err := json.Unmarshal(msg, &ru); err != nil {
			logger.Errorf(err, "parsing into status update")
			return nil
		}

		if ru.Object == nil {
			logger.Infof("msg.Object is nil, so could not extract any info from message, ignoring ...")
			return nil
		}

		obj := unstructured.Unstructured{Object: ru.Object}

		mLogger := logger.WithKV(
			"gvk", obj.GetObjectKind().GroupVersionKind(),
			"accountName", ru.AccountName,
			"clusterName", ru.ClusterName,
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		if len(strings.TrimSpace(ru.AccountName)) == 0 {
			logger.Infof("message does not contain 'accountName', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		if len(strings.TrimSpace(ru.ClusterName)) == 0 {
			logger.Infof("message does not contain 'clusterName', so won't be able to find a resource uniquely, thus ignoring ...")
			return nil
		}

		kind := obj.GetObjectKind().GroupVersionKind().Kind
		ctx := domain.NewConsoleContext(context.TODO(), "sys-user:status-updater", ru.AccountName, ru.ClusterName)

		switch kind {
		case "Project":
			{
				var p entities.Project
				if err := fn.JsonConversion(ru.Object, &p); err != nil {
					return err
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteProjectMessage(ctx, p)
				}
				return d.OnUpdateProjectMessage(ctx, p)
			}
		case "App":
			{
				var a entities.App
				if err := fn.JsonConversion(ru.Object, &a); err != nil {
					return err
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteAppMessage(ctx, a)
				}
				return d.OnUpdateAppMessage(ctx, a)
			}
		case "Config":
			{
				var c entities.Config
				if err := fn.JsonConversion(ru.Object, &c); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteConfigMessage(ctx, c)
				}
				return d.OnUpdateConfigMessage(ctx, c)
			}
		case "Secret":
			{
				var s entities.Secret
				if err := fn.JsonConversion(ru.Object, &s); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteSecretMessage(ctx, s)
				}
				return d.OnUpdateSecretMessage(ctx, s)
			}
		case "Router":
			{
				var r entities.Router
				if err := fn.JsonConversion(ru.Object, &r); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteRouterMessage(ctx, r)
				}
				return d.OnUpdateRouterMessage(ctx, r)
			}
		case "ManagedService":
			{
				var msvc entities.ManagedService
				if err := fn.JsonConversion(ru.Object, &msvc); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteManagedServiceMessage(ctx, msvc)
				}
				return d.OnUpdateManagedServiceMessage(ctx, msvc)
			}
		case "ManagedResource":
			{
				var mres entities.ManagedResource
				if err := fn.JsonConversion(ru.Object, &mres); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteManagedResourceMessage(ctx, mres)
				}
				return d.OnUpdateManagedResourceMessage(ctx, mres)
			}
		}

		return nil
	})
}
