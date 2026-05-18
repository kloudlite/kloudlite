package operations

import (
	"context"
	"os"

	"github.com/kloudlite/kloudlite/api/resources/events"
	"github.com/kloudlite/kloudlite/api/resources/machinetypes"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	resourceservice "github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const machineTypesResource = "machinetypes"

type NamespaceEnsurer interface {
	EnsureNamespaced(ctx context.Context, alias string, namespace string) error
}

type CloudProvider func() string

type Option func(*Operations)

func WithCloudProvider(provider CloudProvider) Option {
	return func(o *Operations) {
		o.cloudProvider = provider
	}
}

type Operations struct {
	service       *resourceservice.Service
	registry      *registry.Registry
	store         *store.Store
	ensurer       NamespaceEnsurer
	cloudProvider CloudProvider
}

type ListRequest struct {
	Resource  string
	Namespace string
	Scope     registry.Scope
	Selector  store.Selector
}

type ListResult struct {
	Items []client.Object
	Dirty []events.DirtyObject
}

type GetRequest struct {
	Resource  string
	Namespace string
	Name      string
	Scope     registry.Scope
}

type GetResult struct {
	Object client.Object
	Dirty  *events.DirtyObject
}

type MutateRequest struct {
	Resource  string
	Namespace string
	Name      string
	Scope     registry.Scope
	Object    client.Object
}

type MutateResult struct {
	Object client.Object
	Dirty  *events.DirtyObject
}

type DeleteRequest struct {
	Resource  string
	Namespace string
	Name      string
	Scope     registry.Scope
}

type DeleteResult struct {
	Dirty *events.DirtyObject
}

func New(svc *resourceservice.Service, reg *registry.Registry, st *store.Store, ensurer NamespaceEnsurer, opts ...Option) *Operations {
	if reg == nil {
		reg = registry.Default()
	}
	operation := &Operations{
		service:       svc,
		registry:      reg,
		store:         st,
		ensurer:       ensurer,
		cloudProvider: func() string { return os.Getenv("CLOUD_PROVIDER") },
	}
	for _, opt := range opts {
		opt(operation)
	}
	return operation
}

func (o *Operations) List(ctx context.Context, req ListRequest) (ListResult, error) {
	if req.Resource == machineTypesResource {
		if err := ensureMachineTypesScope(req.Scope); err != nil {
			return ListResult{}, err
		}
		items, err := o.machineTypes()
		return ListResult{Items: items}, err
	}
	resource, err := o.ensure(ctx, req.Resource, req.Namespace, req.Scope)
	if err != nil {
		return ListResult{}, err
	}
	namespace := scopedNamespace(resource, req.Namespace)
	items, err := o.service.List(ctx, req.Resource, namespace, req.Selector)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Items: items, Dirty: o.store.DirtyForList(req.Resource, namespace, req.Selector)}, nil
}

func (o *Operations) Get(ctx context.Context, req GetRequest) (GetResult, error) {
	if req.Resource == machineTypesResource {
		if err := ensureMachineTypesScope(req.Scope); err != nil {
			return GetResult{}, err
		}
		object, err := o.machineType(req.Name)
		return GetResult{Object: object}, err
	}
	resource, err := o.ensure(ctx, req.Resource, req.Namespace, req.Scope)
	if err != nil {
		return GetResult{}, err
	}
	namespace := scopedNamespace(resource, req.Namespace)
	object, err := o.service.Get(ctx, req.Resource, namespace, req.Name)
	if err != nil {
		return GetResult{}, err
	}
	return GetResult{Object: object, Dirty: o.dirtyForObject(req.Resource, namespace, req.Name)}, nil
}

func (o *Operations) Create(ctx context.Context, req MutateRequest) (MutateResult, error) {
	if req.Resource == machineTypesResource {
		if err := ensureMachineTypesScope(req.Scope); err != nil {
			return MutateResult{}, err
		}
		return MutateResult{}, readOnlyMachineTypes()
	}
	resource, err := o.ensure(ctx, req.Resource, req.Namespace, req.Scope)
	if err != nil {
		return MutateResult{}, err
	}
	namespace := scopedNamespace(resource, req.Namespace)
	object, err := o.service.Create(ctx, req.Resource, namespace, req.Object)
	if err != nil {
		return MutateResult{}, err
	}
	return MutateResult{Object: object, Dirty: o.dirtyForObject(req.Resource, object.GetNamespace(), object.GetName())}, nil
}

func (o *Operations) Patch(ctx context.Context, req MutateRequest) (MutateResult, error) {
	if req.Resource == machineTypesResource {
		if err := ensureMachineTypesScope(req.Scope); err != nil {
			return MutateResult{}, err
		}
		return MutateResult{}, readOnlyMachineTypes()
	}
	resource, err := o.ensure(ctx, req.Resource, req.Namespace, req.Scope)
	if err != nil {
		return MutateResult{}, err
	}
	namespace := scopedNamespace(resource, req.Namespace)
	object, err := o.service.Patch(ctx, req.Resource, namespace, req.Name, req.Object)
	if err != nil {
		return MutateResult{}, err
	}
	return MutateResult{Object: object, Dirty: o.dirtyForObject(req.Resource, object.GetNamespace(), object.GetName())}, nil
}

func (o *Operations) Delete(ctx context.Context, req DeleteRequest) (DeleteResult, error) {
	if req.Resource == machineTypesResource {
		if err := ensureMachineTypesScope(req.Scope); err != nil {
			return DeleteResult{}, err
		}
		return DeleteResult{}, readOnlyMachineTypes()
	}
	resource, err := o.ensure(ctx, req.Resource, req.Namespace, req.Scope)
	if err != nil {
		return DeleteResult{}, err
	}
	namespace := scopedNamespace(resource, req.Namespace)
	if err := o.service.Delete(ctx, req.Resource, namespace, req.Name); err != nil {
		return DeleteResult{}, err
	}
	return DeleteResult{Dirty: o.dirtyForObject(req.Resource, namespace, req.Name)}, nil
}

func (o *Operations) ensure(ctx context.Context, alias string, namespace string, expectedScope registry.Scope) (registry.Resource, error) {
	resource, ok := o.registry.Get(alias)
	if !ok {
		return registry.Resource{}, resourceservice.UnknownResource(alias)
	}
	if expectedScope != "" && resource.Scope != expectedScope {
		return registry.Resource{}, resourceservice.NewError(resourceservice.ErrWrongScope, "resource \""+alias+"\" is "+string(resource.Scope)+"-scoped, not "+string(expectedScope)+"-scoped", nil)
	}
	if resource.Scope != registry.Namespaced {
		return resource, nil
	}
	if namespace == "" {
		return registry.Resource{}, resourceservice.NewError(resourceservice.ErrBadRequest, "namespace is required", nil)
	}
	if o.ensurer == nil {
		return resource, nil
	}
	return resource, o.ensurer.EnsureNamespaced(ctx, alias, namespace)
}

func (o *Operations) dirtyForObject(resource string, namespace string, name string) *events.DirtyObject {
	if dirty, ok := o.store.DirtyForObject(resource, namespace, name); ok {
		return &dirty
	}
	return nil
}

func (o *Operations) machineTypes() ([]client.Object, error) {
	items, err := machinetypes.Defaults(o.cloudProvider())
	if err != nil {
		return nil, resourceservice.NewError(resourceservice.ErrBadRequest, err.Error(), err)
	}
	objects := make([]client.Object, 0, len(items))
	for i := range items {
		objects = append(objects, &items[i])
	}
	return objects, nil
}

func (o *Operations) machineType(name string) (client.Object, error) {
	items, err := o.machineTypes()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.GetName() == name {
			return item, nil
		}
	}
	return nil, resourceservice.NewError(resourceservice.ErrNotFound, "machine type not found", nil)
}

func scopedNamespace(resource registry.Resource, namespace string) string {
	if resource.Scope == registry.Cluster {
		return ""
	}
	return namespace
}

func readOnlyMachineTypes() error {
	return resourceservice.NewError(resourceservice.ErrNotImplemented, "machinetypes are managed in code and are read-only", nil)
}

func ensureMachineTypesScope(scope registry.Scope) error {
	if scope == "" || scope == registry.Cluster {
		return nil
	}
	return resourceservice.NewError(resourceservice.ErrWrongScope, "resource \"machinetypes\" is "+string(registry.Cluster)+"-scoped, not "+string(scope)+"-scoped", nil)
}
