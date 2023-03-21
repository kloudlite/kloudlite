package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kloudlite/operator/operators/status-n-billing/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	domain "kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
)

type StatusUpdateConsumer redpanda.Consumer

func ProcessStatusUpdates(consumer StatusUpdateConsumer, d domain.Domain, logr logging.Logger) {
	counter := 0
	logger := logr.WithName("status-updates")
	consumer.StartConsuming(func(msg []byte, timeStamp time.Time, offset int64) error {
		counter += 1
		logger.Debugf("received message [%d]", counter)

		var su types.StatusUpdate
		if err := json.Unmarshal(msg, &su); err != nil {
			logger.Errorf(err, "parsing into status update")
			return nil
		}

		if su.Object == nil {
			logger.Infof("msg.Object is nil, so could not extract any info from message, ignoring ...")
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

		kind := obj.GetObjectKind().GroupVersionKind().Kind
		ctx := domain.NewConsoleContext(context.TODO(), su.AccountName, su.ClusterName)

		switch kind {
		case "Project":
			{
				var p entities.Project
				if err := fn.JsonConversion(su.Object, &p); err != nil {
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
				if err := fn.JsonConversion(su.Object, &a); err != nil {
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
				if err := fn.JsonConversion(su.Object, &c); err != nil {
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
				if err := fn.JsonConversion(su.Object, &s); err != nil {
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
				if err := fn.JsonConversion(su.Object, &r); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteRouterMessage(ctx, r)
				}
				return d.OnUpdateRouterMessage(ctx, r)
			}
		case "ManagedService":
			{
				var msvc entities.MSvc
				if err := fn.JsonConversion(su.Object, &msvc); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteManagedServiceMessage(ctx, msvc)
				}
				return d.OnUpdateManagedServiceMessage(ctx, msvc)
			}
		case "ManagedResource":
			{
				var mres entities.MRes
				if err := fn.JsonConversion(su.Object, &mres); err != nil {
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
