package watch

import (
	"context"
	"fmt"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Manager struct {
	registry *registry.Registry
	store    *store.Store
	client   client.WithWatch
	logger   *zap.Logger
}

func NewManager(reg *registry.Registry, st *store.Store, kube client.WithWatch, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Manager{registry: reg, store: st, client: kube, logger: logger}
}

func (m *Manager) SyncOnce(ctx context.Context, resource registry.Resource, namespace string) error {
	list := resource.NewList()
	objectList, ok := list.(client.ObjectList)
	if !ok {
		return fmt.Errorf("resource %q list is not a client object list", resource.Alias)
	}

	storeNamespace := ""
	options := []client.ListOption{}
	if resource.Scope == registry.Namespaced {
		storeNamespace = namespace
		options = append(options, client.InNamespace(namespace))
	}

	if err := m.client.List(ctx, objectList, options...); err != nil {
		return err
	}

	items, err := meta.ExtractList(list)
	if err != nil {
		return err
	}

	objects := make([]client.Object, 0, len(items))
	for _, item := range items {
		object, ok := item.(client.Object)
		if !ok {
			return fmt.Errorf("resource %q list item is not a client object", resource.Alias)
		}
		objects = append(objects, object)
	}

	m.store.ReplaceScope(resource.Alias, storeNamespace, objects)
	return nil
}

func (m *Manager) StartClusterScoped(ctx context.Context) {
	for _, resource := range m.registry.All() {
		if resource.Scope != registry.Cluster {
			continue
		}
		if err := m.SyncOnce(ctx, resource, ""); err != nil {
			m.logger.Error("sync cluster resource", zap.String("resource", resource.Alias), zap.Error(err))
		}
	}
}

func (m *Manager) EnsureNamespaced(ctx context.Context, alias string, namespace string) error {
	resource, ok := m.registry.Get(alias)
	if !ok {
		return fmt.Errorf("resource %q is not registered", alias)
	}
	if resource.Scope != registry.Namespaced {
		return fmt.Errorf("resource %q is %s-scoped, not %s-scoped", alias, resource.Scope, registry.Namespaced)
	}
	if m.store.Ready(alias, namespace) {
		return nil
	}
	return m.SyncOnce(ctx, resource, namespace)
}
