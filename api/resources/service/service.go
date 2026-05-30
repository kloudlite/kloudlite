package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kloudlite/kloudlite/api/resources/events"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service struct {
	registry *registry.Registry
	store    *store.Store
	kube     client.Client
	broker   *events.Broker
}

func New(reg *registry.Registry, st *store.Store, kube client.Client) *Service {
	return &Service{registry: reg, store: st, kube: kube}
}

func NewWithEvents(reg *registry.Registry, st *store.Store, kube client.Client, broker *events.Broker) *Service {
	return &Service{registry: reg, store: st, kube: kube, broker: broker}
}

func (s *Service) List(ctx context.Context, alias string, namespace string, _ store.Selector) ([]client.Object, error) {
	resource, err := s.resolve(alias)
	if err != nil {
		return nil, err
	}

	list := resource.NewList()
	objectList, ok := list.(client.ObjectList)
	if !ok {
		return nil, fmt.Errorf("resource %q list is not a client object list", resource.Alias)
	}

	opts := []client.ListOption{}
	if resource.Scope == registry.Namespaced && namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	if err := s.kube.List(ctx, objectList, opts...); err != nil {
		return nil, mapClientError(err)
	}

	items, err := meta.ExtractList(list)
	if err != nil {
		return nil, err
	}

	result := make([]client.Object, 0, len(items))
	for _, item := range items {
		obj, ok := item.(client.Object)
		if !ok {
			return nil, fmt.Errorf("list item is not a client object")
		}
		result = append(result, obj)
	}
	return result, nil
}

func (s *Service) Get(ctx context.Context, alias string, namespace string, name string) (client.Object, error) {
	resource, err := s.resolve(alias)
	if err != nil {
		return nil, err
	}

	obj, err := newClientObject(resource)
	if err != nil {
		return nil, err
	}
	obj.SetName(name)
	if resource.Scope == registry.Namespaced {
		obj.SetNamespace(namespace)
	}
	obj.GetObjectKind().SetGroupVersionKind(resource.GroupVersionKind())

	if err := s.kube.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		return nil, mapClientError(err)
	}
	return obj, nil
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
		return nil, mapClientError(err)
	}
	created.GetObjectKind().SetGroupVersionKind(resource.GroupVersionKind())
	s.markDirty(resource, created.GetNamespace(), created.GetName(), events.DirtyReasonPendingCreate, created.GetLabels())
	s.store.Upsert(resource.Alias, created)
	return created.DeepCopyObject().(client.Object), nil
}

func (s *Service) Patch(ctx context.Context, alias string, namespace string, name string, object client.Object) (client.Object, error) {
	resource, err := s.resolve(alias)
	if err != nil {
		return nil, err
	}
	if object == nil {
		return nil, NewError(ErrBadRequest, "object is required", nil)
	}
	gvk := object.GetObjectKind().GroupVersionKind()
	resourceGVK := resource.GroupVersionKind()
	if !gvk.Empty() && gvk != resourceGVK {
		return nil, NewError(ErrBadRequest, fmt.Sprintf("object GVK %s does not match resource %q GVK %s", gvk, alias, resourceGVK), nil)
	}

	current, err := newClientObject(resource)
	if err != nil {
		return nil, err
	}
	current.SetName(name)
	if resource.Scope == registry.Namespaced {
		current.SetNamespace(namespace)
	} else {
		current.SetNamespace("")
	}
	if err := s.kube.Get(ctx, client.ObjectKeyFromObject(current), current); err != nil {
		return nil, mapClientError(err)
	}

	patchBytes, err := json.Marshal(object)
	if err != nil {
		return nil, NewError(ErrBadRequest, "invalid patch object", err)
	}
	if err := s.kube.Patch(ctx, current, client.RawPatch(types.MergePatchType, patchBytes)); err != nil {
		return nil, mapClientError(err)
	}
	current.GetObjectKind().SetGroupVersionKind(resource.GroupVersionKind())
	s.markDirty(resource, current.GetNamespace(), current.GetName(), events.DirtyReasonPendingPatch, current.GetLabels())
	s.store.Upsert(resource.Alias, current)
	return current.DeepCopyObject().(client.Object), nil
}

func (s *Service) Delete(ctx context.Context, alias string, namespace string, name string) error {
	resource, err := s.resolve(alias)
	if err != nil {
		return err
	}
	object, err := newClientObject(resource)
	if err != nil {
		return err
	}
	object.SetName(name)
	if resource.Scope == registry.Namespaced {
		object.SetNamespace(namespace)
	} else {
		object.SetNamespace("")
	}
	labels := map[string]string(nil)
	if cached, ok := s.store.Get(resource.Alias, object.GetNamespace(), object.GetName()); ok {
		labels = cached.GetLabels()
	}
	if err := s.kube.Delete(ctx, object); err != nil {
		return mapClientError(err)
	}
	s.markDirty(resource, object.GetNamespace(), object.GetName(), events.DirtyReasonPendingDelete, labels)
	s.store.Delete(resource.Alias, object.GetNamespace(), object.GetName())
	return nil
}

func (s *Service) markDirty(resource registry.Resource, namespace string, name string, reason events.DirtyReason, labels map[string]string) {
	if resource.Scope == registry.Cluster {
		namespace = ""
	}
	dirty := s.store.MarkDirty(resource.Alias, namespace, name, reason, labels)
	if s.broker != nil {
		s.broker.Publish(events.Event{Type: events.EventDirty, Resource: resource.Alias, Namespace: namespace, Name: name, Dirty: &dirty})
	}
}

func (s *Service) resolve(alias string) (registry.Resource, error) {
	resource, ok := s.registry.Get(alias)
	if !ok {
		return registry.Resource{}, UnknownResource(alias)
	}
	return resource, nil
}

func newClientObject(resource registry.Resource) (client.Object, error) {
	object, ok := resource.NewObject().(client.Object)
	if !ok {
		return nil, NewError(ErrBadRequest, fmt.Sprintf("resource %q object is not a client object", resource.Alias), nil)
	}
	object.GetObjectKind().SetGroupVersionKind(resource.GroupVersionKind())
	return object, nil
}

func mapClientError(err error) error {
	switch {
	case apierrors.IsNotFound(err):
		return NewError(ErrNotFound, "resource was not found", err)
	case apierrors.IsAlreadyExists(err), apierrors.IsConflict(err):
		return NewError(ErrConflict, "resource update conflict", err)
	case apierrors.IsInvalid(err), apierrors.IsBadRequest(err):
		return NewError(ErrBadRequest, "invalid resource request", err)
	default:
		return err
	}
}
