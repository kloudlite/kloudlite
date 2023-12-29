package app

import (
	"context"
	"encoding/json"
	"github.com/kloudlite/api/pkg/errors"
	"strings"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/apps/console/internal/entities"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceUpdateConsumer messaging.Consumer

func ProcessResourceUpdates(consumer ResourceUpdateConsumer, d domain.Domain, logger logging.Logger) {
	counter := 0

	projectGVK := fn.GVK("crds.kloudlite.io/v1", "Project")
	appsGVK := fn.GVK("crds.kloudlite.io/v1", "App")
	workspaceGVK := fn.GVK("crds.kloudlite.io/v1", "Workspace")
	imagePullSecretGVK := fn.GVK("crds.kloudlite.io/v1", "ImagePullSecret")
	configGVK := fn.GVK("crds.kloudlite.io/v1", "Config")
	secretGVK := fn.GVK("crds.kloudlite.io/v1", "Secret")
	routerGVK := fn.GVK("crds.kloudlite.io/v1", "Router")
	managedResourceGVK := fn.GVK("crds.kloudlite.io/v1", "ManagedResource")

	msgReader := func(msg *msgTypes.ConsumeMsg) error {
		logger := logger.WithKV("subject", msg.Subject)

		counter += 1
		logger.Debugf("[%d] received message", counter)

		var ru types.ResourceUpdate
		if err := json.Unmarshal(msg.Payload, &ru); err != nil {
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

		kind := obj.GetObjectKind().GroupVersionKind().String()
		dctx := domain.NewConsoleContext(context.TODO(), "sys-user:status-updater", ru.AccountName, ru.ClusterName)

		switch kind {
		case projectGVK.String():
			{
				var p entities.Project
				if err := fn.JsonConversion(ru.Object, &p); err != nil {
					return errors.NewE(err)
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteProjectMessage(dctx, p)
				}
				return d.OnUpdateProjectMessage(dctx, p)
			}

		case workspaceGVK.String():
			{
				var p entities.Workspace
				if err := fn.JsonConversion(ru.Object, &p); err != nil {
					return errors.NewE(err)
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteWorkspaceMessage(dctx, p)
				}
				return d.OnUpdateWorkspaceMessage(dctx, p)
			}
		case appsGVK.String():
			{
				var a entities.App
				if err := fn.JsonConversion(ru.Object, &a); err != nil {
					return errors.NewE(err)
				}

				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteAppMessage(dctx, a)
				}
				return d.OnUpdateAppMessage(dctx, a)
			}
		case configGVK.String():
			{
				var c entities.Config
				if err := fn.JsonConversion(ru.Object, &c); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteConfigMessage(dctx, c)
				}
				return d.OnUpdateConfigMessage(dctx, c)
			}
		case secretGVK.String():
			{
				var s entities.Secret
				if err := fn.JsonConversion(ru.Object, &s); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteSecretMessage(dctx, s)
				}
				return d.OnUpdateSecretMessage(dctx, s)
			}
		case imagePullSecretGVK.String():
			{
				var s entities.ImagePullSecret
				if err := fn.JsonConversion(ru.Object, &s); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteImagePullSecretMessage(dctx, s)
				}
				return d.OnUpdateImagePullSecretMessage(dctx, s)
			}
		case routerGVK.String():
			{
				var r entities.Router
				if err := fn.JsonConversion(ru.Object, &r); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteRouterMessage(dctx, r)
				}
				return d.OnUpdateRouterMessage(dctx, r)
			}
		case managedResourceGVK.String():
			{
				var mres entities.ManagedResource
				if err := fn.JsonConversion(ru.Object, &mres); err != nil {
					return errors.NewE(err)
				}
				if obj.GetDeletionTimestamp() != nil {
					return d.OnDeleteManagedResourceMessage(dctx, mres)
				}
				return d.OnUpdateManagedResourceMessage(dctx, mres)
			}
		}

		return nil
	}

	if err:=consumer.Consume(msgReader, msgTypes.ConsumeOpts{
		OnError: func(err error) error {
			logger.Errorf(err, "received while reading messages, ignoring it")
			return nil
		},
	}); err != nil {
		logger.Errorf(err, "error while consuming messages")
	}
}
