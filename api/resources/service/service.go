package service

import (
	"context"
	"fmt"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service struct {
	registry *registry.Registry
	store    *store.Store
	kube     client.Client
}

func New(reg *registry.Registry, st *store.Store, kube client.Client) *Service {
	return &Service{registry: reg, store: st, kube: kube}
}

func (s *Service) List(ctx context.Context, alias string, namespace string, selector store.Selector) ([]client.Object, error) {
	_ = ctx
	if _, err := s.resolve(alias); err != nil {
		return nil, err
	}
	if !s.store.Ready(alias, namespace) {
		return nil, NewError(ErrCacheNotReady, fmt.Sprintf("resource %q cache is not ready", alias), nil)
	}
	return s.store.List(alias, namespace, selector), nil
}

func (s *Service) Get(ctx context.Context, alias string, namespace string, name string) (client.Object, error) {
	_ = ctx
	if _, err := s.resolve(alias); err != nil {
		return nil, err
	}
	if !s.store.Ready(alias, namespace) {
		return nil, NewError(ErrCacheNotReady, fmt.Sprintf("resource %q cache is not ready", alias), nil)
	}
	object, ok := s.store.Get(alias, namespace, name)
	if !ok {
		return nil, NewError(ErrNotFound, fmt.Sprintf("resource %q named %q was not found", alias, name), nil)
	}
	return object, nil
}

func (s *Service) Create(ctx context.Context, alias string, namespace string, object client.Object) (client.Object, error) {
	resource, err := s.resolve(alias)
	if err != nil {
		return nil, err
	}
	if object == nil {
		return nil, NewError(ErrBadRequest, "object is required", nil)
	}

	created := object.DeepCopyObject().(client.Object)
	gvk := created.GetObjectKind().GroupVersionKind()
	resourceGVK := resource.GroupVersionKind()
	if gvk.Empty() {
		created.GetObjectKind().SetGroupVersionKind(resource.GroupVersionKind())
	} else if gvk != resourceGVK {
		return nil, NewError(ErrBadRequest, fmt.Sprintf("object GVK %s does not match resource %q GVK %s", gvk, alias, resourceGVK), nil)
	}
	if resource.Scope == registry.Namespaced {
		created.SetNamespace(namespace)
	} else {
		created.SetNamespace("")
	}

	if err := s.kube.Create(ctx, created); err != nil {
		return nil, err
	}
	created.GetObjectKind().SetGroupVersionKind(resource.GroupVersionKind())
	return created.DeepCopyObject().(client.Object), nil
}

func (s *Service) Patch(ctx context.Context, alias string, namespace string, name string, object client.Object) (client.Object, error) {
	_, _, _, _, _ = ctx, alias, namespace, name, object
	return nil, NewError(ErrNotImplemented, "patch is not implemented", nil)
}

func (s *Service) Delete(ctx context.Context, alias string, namespace string, name string) error {
	_, _, _, _ = ctx, alias, namespace, name
	return NewError(ErrNotImplemented, "delete is not implemented", nil)
}

func (s *Service) resolve(alias string) (registry.Resource, error) {
	resource, ok := s.registry.Get(alias)
	if !ok {
		return registry.Resource{}, UnknownResource(alias)
	}
	return resource, nil
}
