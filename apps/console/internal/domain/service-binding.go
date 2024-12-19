package domain

import (
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateServiceIntercept implements Domain.
func (d *domain) CreateServiceIntercept(ctx ConsoleContext, envName string, serviceName string, interceptTo string, portMappings []*crdsv1.SvcInterceptPortMappings) (*entities.ServiceBinding, error) {
	filters := ctx.DBFilters()

	env, err := d.environmentRepo.FindOne(ctx, ctx.DBFilters().Add(fc.MetadataName, envName))
	if err != nil {
		return nil, err
	}

	if env == nil {
		return nil, fmt.Errorf("environment not found")
	}

	sb, err := d.serviceBindingRepo.FindOne(ctx, filters.Add(fc.ServiceBindingSpecServiceRefName, serviceName).Add(fc.ClusterName, env.ClusterName))
	if err != nil {
		return nil, err
	}

	if sb == nil {
		return nil, fmt.Errorf("no service binding found")
	}

	pm := make([]crdsv1.SvcInterceptPortMappings, len(portMappings))
	for i := range portMappings {
		pm[i] = *portMappings[i]
	}

	serviceIntercept := &crdsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: env.Spec.TargetNamespace,
		},
		Spec: crdsv1.ServiceInterceptSpec{
			ToAddr:       interceptTo,
			PortMappings: pm,
		},
	}

	serviceIntercept.EnsureGVK()

	usb, err := d.serviceBindingRepo.PatchById(ctx, sb.Id, repos.Document{
		fc.EnvironmentName:      envName,
		fc.EnvironmentNamespace: env.Spec.TargetNamespace,
		fc.ServiceBindingInterceptStatus: &entities.InterceptStatus{
			Intercepted:  fn.New(true),
			ToAddr:       interceptTo,
			PortMappings: pm,
		},
	})
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, usb.EnvironmentName, serviceIntercept, usb.RecordVersion); err != nil {
		return nil, err
	}

	return usb, nil
}

// DeleteServiceIntercept implements Domain.
func (d *domain) DeleteServiceIntercept(ctx ConsoleContext, envName string, serviceName string) error {
	filters := ctx.DBFilters()

	env, err := d.environmentRepo.FindOne(ctx, ctx.DBFilters().Add(fc.MetadataName, envName))
	if err != nil {
		return err
	}

	if env == nil {
		return fmt.Errorf("environment not found")
	}

	sb, err := d.serviceBindingRepo.FindOne(ctx, filters.Add(fc.ServiceBindingSpecServiceRefName, serviceName).Add(fc.ClusterName, env.ClusterName))
	if err != nil {
		return err
	}

	serviceIntercept := &crdsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: sb.EnvironmentNamespace,
		},
	}

	serviceIntercept.EnsureGVK()

	if _, err := d.serviceBindingRepo.PatchById(ctx, sb.Id, repos.Document{
		fc.ServiceBindingInterceptStatus: &entities.InterceptStatus{
			Intercepted:  fn.New(false),
			ToAddr:       "",
			PortMappings: nil,
		},
	}); err != nil {
		return err
	}

	if err := d.deleteK8sResource(ctx, sb.EnvironmentName, serviceIntercept); err != nil {
		return err
	}

	return nil
}

// ListServiceBindings implements Domain.
func (d *domain) ListServiceBindings(ctx ConsoleContext, envName string, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ServiceBinding], error) {
	filters := ctx.DBFilters()

	env, err := d.environmentRepo.FindOne(ctx, filters.Add(fc.MetadataName, envName))
	if err != nil {
		return nil, err
	}

	if env == nil {
		return nil, fmt.Errorf("environment not found")
	}

	return d.serviceBindingRepo.FindPaginated(ctx, filters.Add(fc.ServiceBindingSpecServiceRefNamespace, env.Spec.TargetNamespace), pagination)
}

// OnServiceBindingDeleteMessage implements Domain.
func (d *domain) OnServiceBindingDeleteMessage(ctx ConsoleContext, svcb *networkingv1.ServiceBinding) error {
	if svcb == nil {
		return errors.Newf("no service binding found")
	}

	if svcb.Spec.Hostname == "" {
		return nil
	}

	if err := d.serviceBindingRepo.DeleteOne(ctx, repos.Filter{fc.AccountName: ctx.AccountName, fc.ServiceBindingSpecHostname: svcb.Spec.Hostname}); err != nil {
		return err
	}

	return nil
}

// OnServiceBindingUpdateMessage implements Domain.
func (d *domain) OnServiceBindingUpdateMessage(ctx ConsoleContext, svcb *networkingv1.ServiceBinding, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	if svcb == nil {
		return errors.Newf("no service binding found")
	}

	filter := ctx.DBFilters().Add(fc.MetadataName, svcb.Name)

	if svcb.Spec.ServiceIP == nil || svcb.Spec.Hostname == "" {
		d.logger.Infof("filters: %#v\n", filter)
		// INFO: it means that service binding has been de-allocated
		if err := d.serviceBindingRepo.DeleteOne(ctx, filter); err != nil {
			if !errors.Is(err, repos.ErrNoDocuments) {
				return err
			}
		}
		return nil
	}

	var environmentName string
	if svcb.Spec.ServiceRef != nil {
		env, err := d.environmentRepo.FindOne(ctx, ctx.DBFilters().Add(fc.EnvironmentSpecTargetNamespace, svcb.Spec.ServiceRef.Namespace))
		if err != nil {
			return err
		}

		if env == nil {
			return fmt.Errorf("environment not found")
		}

		environmentName = env.Name
	}

	if _, err := d.serviceBindingRepo.Upsert(ctx, filter, &entities.ServiceBinding{
		ServiceBinding:  *svcb,
		AccountName:     ctx.AccountName,
		ClusterName:     opts.ClusterName,
		EnvironmentName: environmentName,
	}); err != nil {
		return errors.NewE(err)
	}

	// d.resourceEventPublisher.PublishResourceEvent(ctx, urouter.GetResourceType(), urouter.GetName(), PublishUpdate)
	return nil
}
