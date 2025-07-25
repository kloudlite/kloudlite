package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"
	"time"

	"github.com/kloudlite/api/apps/console/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/console/internal/app/graph/model"
	"github.com/kloudlite/api/apps/console/internal/entities"
	fn "github.com/kloudlite/api/pkg/functions"
)

// CreationTime is the resolver for the creationTime field.
func (r *serviceBindingResolver) CreationTime(ctx context.Context, obj *entities.ServiceBinding) (string, error) {
	if obj == nil {
		return "", errNilSecretVariable
	}
	return obj.BaseEntity.CreationTime.Format(time.RFC3339), nil
}

// InterceptStatus is the resolver for the interceptStatus field.
func (r *serviceBindingResolver) InterceptStatus(ctx context.Context, obj *entities.ServiceBinding) (*model.GithubComKloudliteAPIAppsConsoleInternalEntitiesInterceptStatus, error) {
	return fn.JsonConvertP[model.GithubComKloudliteAPIAppsConsoleInternalEntitiesInterceptStatus](obj.InterceptStatus)
}

// Spec is the resolver for the spec field.
func (r *serviceBindingResolver) Spec(ctx context.Context, obj *entities.ServiceBinding) (*model.GithubComKloudliteOperatorApisNetworkingV1ServiceBindingSpec, error) {
	ports := make([]*model.K8sIoAPICoreV1ServicePort, 0, len(obj.Spec.Ports))
	for _, port := range obj.Spec.Ports {
		ports = append(ports, &model.K8sIoAPICoreV1ServicePort{
			AppProtocol: port.AppProtocol,
			Name:        &port.Name,
			NodePort:    fn.New(int(port.NodePort)),
			Port:        int(port.Port),
			Protocol:    (*model.K8sIoAPICoreV1Protocol)(&port.Protocol),
			TargetPort: &model.K8sIoApimachineryPkgUtilIntstrIntOrString{
				StrVal: port.TargetPort.StrVal,
			},
		})
	}

	return &model.GithubComKloudliteOperatorApisNetworkingV1ServiceBindingSpec{
		GlobalIP:  obj.Spec.GlobalIP,
		Hostname:  &obj.Spec.Hostname,
		Ports:     ports,
		ServiceIP: new(string),
		ServiceRef: &model.GithubComKloudliteOperatorApisCommonTypesNamespacedResourceRef{
			Name:      obj.Spec.ServiceRef.Name,
			Namespace: obj.Spec.ServiceRef.Namespace,
		},
	}, nil
}

// Status is the resolver for the status field.
func (r *serviceBindingResolver) Status(ctx context.Context, obj *entities.ServiceBinding) (*model.GithubComKloudliteOperatorPkgOperatorStatus, error) {
	return fn.JsonConvertP[model.GithubComKloudliteOperatorPkgOperatorStatus](obj.Status)
}

// UpdateTime is the resolver for the updateTime field.
func (r *serviceBindingResolver) UpdateTime(ctx context.Context, obj *entities.ServiceBinding) (string, error) {
	if obj == nil {
		return "", errNilSecretVariable
	}
	return obj.BaseEntity.UpdateTime.Format(time.RFC3339), nil
}

// ServiceBinding returns generated.ServiceBindingResolver implementation.
func (r *Resolver) ServiceBinding() generated.ServiceBindingResolver {
	return &serviceBindingResolver{r}
}

type serviceBindingResolver struct{ *Resolver }
