package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

// OnServiceBindingDeleteMessage implements Domain.
func (d *domain) OnServiceBindingDeleteMessage(ctx ConsoleContext, svcb *networkingv1.ServiceBinding) error {
	panic("unimplemented")
}

// OnServiceBindingUpdateMessage implements Domain.
func (d *domain) OnServiceBindingUpdateMessage(ctx ConsoleContext, svcb *networkingv1.ServiceBinding, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	if svcb == nil {
		return errors.Newf("no service binding found")
	}

	if svcb.Spec.Hostname == "" {
		return nil
	}

	if _, err := d.serviceBindingRepo.Upsert(ctx, repos.Filter{
		fc.AccountName: ctx.AccountName,
		// fc.ClusterName:                opts.ClusterName,
		// fc.MetadataName:               svcb.Name,
		fc.ServiceBindingSpecHostname: svcb.Spec.Hostname,
	}, &entities.ServiceBinding{
		ServiceBinding: *svcb,
		AccountName:    ctx.AccountName,
		ClusterName:    opts.ClusterName,
	}); err != nil {
		return errors.NewE(err)
	}

	// d.resourceEventPublisher.PublishResourceEvent(ctx, urouter.GetResourceType(), urouter.GetName(), PublishUpdate)
	return nil
}
