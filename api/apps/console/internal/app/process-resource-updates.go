package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ResourceUpdateConsumer messaging.Consumer

func newResourceContext(ctx domain.ConsoleContext, projectName string, environmentName string) domain.ResourceContext {
	return domain.ResourceContext{
		ConsoleContext:  ctx,
		ProjectName:     projectName,
		EnvironmentName: environmentName,
	}
}

func ProcessResourceUpdates(consumer ResourceUpdateConsumer, d domain.Domain, logger logging.Logger) {
	counter := 0

	projectGVK := fn.GVK("crds.kloudlite.io/v1", "Project")
	appsGVK := fn.GVK("crds.kloudlite.io/v1", "App")
	// workspaceGVK := fn.GVK("crds.kloudlite.io/v1", "Environment") //FIXME
	workspaceGVK := fn.GVK("crds.kloudlite.io/v1", "Workspace")
	imagePullSecretGVK := fn.GVK("crds.kloudlite.io/v1", "ImagePullSecret")
	configGVK := fn.GVK("v1", "ConfigMap")
	secretGVK := fn.GVK("v1", "Secret")
	routerGVK := fn.GVK("crds.kloudlite.io/v1", "Router")
	managedResourceGVK := fn.GVK("crds.kloudlite.io/v1", "ManagedResource")

	getResourceContext := func(ctx domain.ConsoleContext, rt entities.ResourceType, obj unstructured.Unstructured) (domain.ResourceContext, error) {
		mapping, err := d.GetResourceMapping(ctx, rt, obj.GetNamespace(), obj.GetName())
		if err != nil {
			return domain.ResourceContext{}, err
		}
		if mapping == nil {
			return domain.ResourceContext{}, errors.Newf("mapping not found for %s %s/%s", rt, obj.GetNamespace(), obj.GetName())
		}

		return newResourceContext(ctx, mapping.ProjectName, mapping.EnvironmentName), nil
	}

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

		dctx := domain.NewConsoleContext(context.TODO(), "sys-user:console-resource-updater", ru.AccountName)

		gvkStr := obj.GetObjectKind().GroupVersionKind().String()

		resStatus, err := func() (types.ResourceStatus, error) {
			v, ok := ru.Object[types.ResourceStatusKey]
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

		switch gvkStr {
		case projectGVK.String():
			{
				var p entities.Project
				if err := fn.JsonConversion(ru.Object, &p); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnProjectDeleteMessage(dctx, p)
				}
				return d.OnProjectUpdateMessage(dctx, p, resStatus, opts)
			}

		case workspaceGVK.String():
			{
				var ws entities.Environment
				if err := fn.JsonConversion(ru.Object, &ws); err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnEnvironmentDeleteMessage(dctx, ws)
				}
				return d.OnEnvironmentUpdateMessage(dctx, ws, resStatus, opts)
			}
		case appsGVK.String():
			{
				var app entities.App
				if err := fn.JsonConversion(ru.Object, &app); err != nil {
					return errors.NewE(err)
				}

				rctx, err := getResourceContext(dctx, entities.ResourceTypeApp, obj)
				if err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnAppDeleteMessage(rctx, app)
				}
				return d.OnAppUpdateMessage(rctx, app, resStatus, opts)
			}
		case configGVK.String():
			{
				var config entities.Config
				if err := fn.JsonConversion(ru.Object, &config); err != nil {
					return errors.NewE(err)
				}

				rctx, err := getResourceContext(dctx, entities.ResourceTypeConfig, obj)
				if err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnConfigDeleteMessage(rctx, config)
				}
				return d.OnConfigUpdateMessage(rctx, config, resStatus, opts)
			}
		case secretGVK.String():
			{
				var secret entities.Secret
				if err := fn.JsonConversion(ru.Object, &secret); err != nil {
					return errors.NewE(err)
				}

				rctx, err := getResourceContext(dctx, entities.ResourceTypeSecret, obj)
				if err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnSecretDeleteMessage(rctx, secret)
				}
				return d.OnSecretUpdateMessage(rctx, secret, resStatus, opts)
			}
		case imagePullSecretGVK.String():
			{
				var ips entities.ImagePullSecret
				if err := fn.JsonConversion(ru.Object, &ips); err != nil {
					return errors.NewE(err)
				}

				rctx, err := getResourceContext(dctx, entities.ResourceTypeImagePullSecret, obj)
				if err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnImagePullSecretDeleteMessage(rctx, ips)
				}
				return d.OnImagePullSecretUpdateMessage(rctx, ips, resStatus, opts)
			}
		case routerGVK.String():
			{
				var router entities.Router
				if err := fn.JsonConversion(ru.Object, &router); err != nil {
					return errors.NewE(err)
				}

				rctx, err := getResourceContext(dctx, entities.ResourceTypeRouter, obj)
				if err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnRouterDeleteMessage(rctx, router)
				}
				return d.OnRouterUpdateMessage(rctx, router, resStatus, opts)
			}
		case managedResourceGVK.String():
			{
				var mres entities.ManagedResource
				if err := fn.JsonConversion(ru.Object, &mres); err != nil {
					return errors.NewE(err)
				}

				rctx, err := getResourceContext(dctx, entities.ResourceTypeManagedResource, obj)
				if err != nil {
					return errors.NewE(err)
				}

				if resStatus == types.ResourceStatusDeleted {
					return d.OnManagedResourceDeleteMessage(rctx, mres)
				}
				return d.OnManagedResourceUpdateMessage(rctx, mres, resStatus, opts)
			}
		}

		return nil
	}

	if err := consumer.Consume(msgReader, msgTypes.ConsumeOpts{
		OnError: func(err error) error {
			logger.Errorf(err, "received while reading messages, ignoring it")
			return nil
		},
	}); err != nil {
		logger.Errorf(err, "error while consuming messages")
	}
}
