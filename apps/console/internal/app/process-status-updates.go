package app

import (
	"encoding/json"
	"strings"

	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/kafka"
)

type ResourceUpdateConsumer kafka.Consumer

func ProcessResourceUpdates(consumer ResourceUpdateConsumer, d domain.Domain) {
	counter := 0
	consumer.StartConsuming(func(ctx kafka.ConsumerContext, topic string, value []byte, metadata kafka.RecordMetadata) error {
		counter += 1
		logger := ctx.Logger
		logger.Debugf("[%d] received message", counter)

		var ru types.ResourceUpdate
		if err := json.Unmarshal(value, &ru); err != nil {
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
		dctx := domain.NewConsoleContext(ctx, "sys-user:status-updater", ru.AccountName, ru.ClusterName)

		switch kind {
		case "Project":
			{
				var p entities.Project
				if err := fn.JsonConversion(ru.Object, &p); err != nil {
					return err
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteProjectMessage(dctx, p)
				}
				return d.OnUpdateProjectMessage(dctx, p)
			}

		case "Workspace":
			{
				var p entities.Workspace
				if err := fn.JsonConversion(ru.Object, &p); err != nil {
					return err
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteWorkspaceMessage(dctx, p)
				}
				return d.OnUpdateWorkspaceMessage(dctx, p)
			}
		case "App":
			{
				var a entities.App
				if err := fn.JsonConversion(ru.Object, &a); err != nil {
					return err
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteAppMessage(dctx, a)
				}
				return d.OnUpdateAppMessage(dctx, a)
			}
		case "Config":
			{
				var c entities.Config
				if err := fn.JsonConversion(ru.Object, &c); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteConfigMessage(dctx, c)
				}
				return d.OnUpdateConfigMessage(dctx, c)
			}
		case "Secret":
			{
				var s entities.Secret
				if err := fn.JsonConversion(ru.Object, &s); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteSecretMessage(dctx, s)
				}
				return d.OnUpdateSecretMessage(dctx, s)
			}
		case "ImagePullSecret":
			{
				var s entities.ImagePullSecret
				if err := fn.JsonConversion(ru.Object, &s); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteImagePullSecretMessage(dctx, s)
				}
				return d.OnUpdateImagePullSecretMessage(dctx, s)
			}
		case "Router":
			{
				var r entities.Router
				if err := fn.JsonConversion(ru.Object, &r); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteRouterMessage(dctx, r)
				}
				return d.OnUpdateRouterMessage(dctx, r)
			}
		case "ManagedService":
			{
				var msvc entities.ManagedService
				if err := fn.JsonConversion(ru.Object, &msvc); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteManagedServiceMessage(dctx, msvc)
				}
				return d.OnUpdateManagedServiceMessage(dctx, msvc)
			}
		case "ManagedResource":
			{
				var mres entities.ManagedResource
				if err := fn.JsonConversion(ru.Object, &mres); err != nil {
					return err
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteManagedResourceMessage(dctx, mres)
				}
				return d.OnUpdateManagedResourceMessage(dctx, mres)
			}
		}

		return nil
	})
}
