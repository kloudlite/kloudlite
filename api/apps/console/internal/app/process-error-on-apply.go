package app

import (
	"context"
	"encoding/json"
	"fmt"

	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"github.com/kloudlite/api/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/apps/console/internal/entities"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
)

type ErrorOnApplyConsumer messaging.Consumer

func ProcessErrorOnApply(consumer ErrorOnApplyConsumer, d domain.Domain, logger logging.Logger) {
	counter := 0

	getEnvironmentResourceContext := func(ctx domain.ConsoleContext, resType entities.ResourceType, clusterName string, obj unstructured.Unstructured) (domain.ResourceContext, error) {
		mapping, err := d.GetEnvironmentResourceMapping(ctx, resType, clusterName, obj.GetNamespace(), obj.GetName())
		if err != nil {
			return domain.ResourceContext{}, err
		}
		if mapping == nil {
			return domain.ResourceContext{}, fmt.Errorf("resource mapping could not be found")
		}
		return newResourceContext(ctx, mapping.ProjectName, mapping.EnvironmentName), nil
	}

	msgReader := func(msg *msgTypes.ConsumeMsg) error {
		counter += 1
		logger.Debugf("received message [%d]", counter)
		var errObj t.AgentErrMessage
		if err := json.Unmarshal(msg.Payload, &errObj); err != nil {
			return errors.NewE(err)
		}

		obj := unstructured.Unstructured{Object: errObj.Object}

		mLogger := logger.WithKV(
			"gvk", obj.GroupVersionKind(),
			"accountName", errObj.AccountName,
			"clusterName", errObj.ClusterName,
		)

		mLogger.Infof("received message")
		defer func() {
			mLogger.Infof("processed message")
		}()

		dctx := domain.NewConsoleContext(context.TODO(), "sys-user:apply-on-error-worker", errObj.AccountName)

		opts := domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp}

		gvkStr := obj.GroupVersionKind().String()

		switch gvkStr {
		case deviceGVK.String():
			{
				if errObj.Action == t.ActionApply {
					return d.OnVPNDeviceApplyError(dctx, errObj.Error, obj.GetName(), opts)
				}

				p, err := fn.JsonConvert[entities.ConsoleVPNDevice](obj.Object)
				if err != nil {
					return err
				}

				return d.OnVPNDeviceDeleteMessage(dctx, p)
			}
		case projectGVK.String():
			{
				if errObj.Action == t.ActionApply {
					return d.OnProjectApplyError(dctx, errObj.Error, obj.GetName(), opts)
				}

				p, err := fn.JsonConvert[entities.Project](obj.Object)
				if err != nil {
					return err
				}

				return d.OnProjectDeleteMessage(dctx, p)
			}
		case environmentGVK.String():
			{
				if errObj.Action == t.ActionApply {
					return d.OnEnvironmentApplyError(dctx, errObj.Error, obj.GetNamespace(), obj.GetName(), opts)
				}

				p, err := fn.JsonConvert[entities.Environment](obj.Object)
				if err != nil {
					return err
				}

				return d.OnEnvironmentDeleteMessage(dctx, p)
			}
		case projectManagedServiceGVK.String():
			{
				mapping, err := d.GetProjectResourceMapping(dctx, entities.ResourceTypeProjectManagedService, errObj.ClusterName, obj.GetName())
				if err != nil {
					return err
				}
				if mapping == nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnProjectManagedServiceApplyError(dctx, mapping.ProjectName, obj.GetName(), errObj.Error, opts)
				}

				pmsvc, err := fn.JsonConvert[entities.ProjectManagedService](obj.Object)
				if err != nil {
					return err
				}

				return d.OnProjectManagedServiceDeleteMessage(dctx, mapping.ProjectName, pmsvc)
			}

		case appsGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeApp, errObj.ClusterName, obj)
				if err != nil {
					return errors.NewE(err)
				}

				app, err := fn.JsonConvert[entities.App](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnAppApplyError(rctx, errObj.Error, obj.GetName(), opts)
				}

				return d.OnAppDeleteMessage(rctx, app)
			}
		case configGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeConfig, errObj.ClusterName, obj)
				if err != nil {
					return errors.NewE(err)
				}

				config, err := fn.JsonConvert[entities.Config](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnConfigApplyError(rctx, errObj.Error, obj.GetName(), opts)
				}
				return d.OnConfigDeleteMessage(rctx, config)
			}
		case secretGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeSecret, errObj.ClusterName, obj)
				if err != nil {
					return errors.NewE(err)
				}

				secret, err := fn.JsonConvert[entities.Secret](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnSecretApplyError(rctx, errObj.Error, obj.GetName(), opts)
				}
				return d.OnSecretDeleteMessage(rctx, secret)
			}
		case routerGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeRouter, errObj.ClusterName, obj)
				if err != nil {
					return errors.NewE(err)
				}

				router, err := fn.JsonConvert[entities.Router](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnRouterApplyError(rctx, errObj.Error, obj.GetName(), opts)
				}
				return d.OnRouterDeleteMessage(rctx, router)
			}
		case managedResourceGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeManagedResource, errObj.ClusterName, obj)
				if err != nil {
					return errors.NewE(err)
				}

				mres, err := fn.JsonConvert[entities.ManagedResource](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnManagedResourceApplyError(rctx, errObj.Error, obj.GetName(), opts)
				}
				return d.OnManagedResourceDeleteMessage(rctx, mres)
			}

		default:
			{
				return errors.Newf("console apply error reader does not acknowledge resource with GVK (%s)", gvkStr)
			}
		}
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
