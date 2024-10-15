package resource_updates_receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"github.com/kloudlite/api/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/apps/console/internal/entities"
	msgOfficeT "github.com/kloudlite/api/apps/message-office/types"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
)

type ErrorOnApplyConsumer messaging.Consumer

func ProcessErrorOnApply(consumer ErrorOnApplyConsumer, d domain.Domain, logger *slog.Logger) {
	getEnvironmentResourceContext := func(ctx domain.ConsoleContext, resType entities.ResourceType, clusterName string, obj unstructured.Unstructured) (domain.ResourceContext, error) {
		mapping, err := d.GetEnvironmentResourceMapping(ctx, resType, clusterName, obj.GetNamespace(), obj.GetName())
		if err != nil {
			return domain.ResourceContext{}, err
		}
		if mapping == nil {
			return domain.ResourceContext{}, fmt.Errorf("resource mapping could not be found")
		}
		return newResourceContext(ctx, mapping.EnvironmentName), nil
	}

	counter := 0
	mu := sync.Mutex{}

	msgReader := func(msg *msgTypes.ConsumeMsg) error {
		mu.Lock()
		counter += 1
		mu.Unlock()

		start := time.Now()

		logger := logger.With("subject", msg.Subject, "counter", counter)
		logger.Debug("INCOMING message", "counter", counter)

		em, err := msgOfficeT.UnmarshalErrMessage(msg.Payload)
		if err != nil {
			return errors.NewE(err)
		}

		var errObj t.AgentErrMessage
		if err := json.Unmarshal(em.Error, &errObj); err != nil {
			return errors.NewE(err)
		}

		obj := unstructured.Unstructured{Object: errObj.Object}
		gvkStr := obj.GroupVersionKind().String()

		mlogger := logger.With(
			"GVK", gvkStr,
			"NN", fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()),
			"accountName", em.AccountName,
			"clusterName", em.ClusterName,
		)

		mlogger.Info("validated message")
		defer func() {
			mlogger.Info("PROCESSED message", "took", fmt.Sprintf("%.2fs", time.Since(start).Seconds()))
		}()

		dctx := domain.NewConsoleContext(context.TODO(), "sys-user:apply-on-error-worker", em.AccountName)

		opts := domain.UpdateAndDeleteOpts{MessageTimestamp: msg.Timestamp}

		switch gvkStr {
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
		case appsGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeApp, em.ClusterName, obj)
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
		case externalAppsGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeExternalApp, em.ClusterName, obj)
				if err != nil {
					return errors.NewE(err)
				}

				externalApp, err := fn.JsonConvert[entities.ExternalApp](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnExternalAppApplyError(rctx, errObj.Error, obj.GetName(), opts)
				}

				return d.OnExternalAppDeleteMessage(rctx, externalApp)
			}
		case configGVK.String():
			{
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeConfig, em.ClusterName, obj)
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
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeSecret, em.ClusterName, obj)
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
				rctx, err := getEnvironmentResourceContext(dctx, entities.ResourceTypeRouter, em.ClusterName, obj)
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
				mres, err := fn.JsonConvert[crdsv1.ManagedResource](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnManagedResourceApplyError(dctx, errObj.Error, mres.Spec.ResourceTemplate.MsvcRef.Name, obj.GetName(), opts)
				}
				return d.OnManagedResourceDeleteMessage(dctx, mres.Spec.ResourceTemplate.MsvcRef.Name, mres)
			}
		case clusterMsvcGVK.String():
			{
				cmsvc, err := fn.JsonConvert[entities.ClusterManagedService](obj.Object)
				if err != nil {
					return err
				}

				if errObj.Action == t.ActionApply {
					return d.OnClusterManagedServiceApplyError(dctx, em.ClusterName, obj.GetName(), errObj.Error, opts)
				}
				return d.OnClusterManagedServiceDeleteMessage(dctx, em.ClusterName, cmsvc)
			}

		default:
			{
				return errors.Newf("console apply error reader does not acknowledge resource with GVK (%s)", gvkStr)
			}
		}
	}

	if err := consumer.Consume(msgReader, msgTypes.ConsumeOpts{
		OnError: func(err error) error {
			logger.Error("while reading messages, got", "err", err)
			return nil
		},
	}); err != nil {
		logger.Error("while consuming messages, got", "err", err)
	}
}
