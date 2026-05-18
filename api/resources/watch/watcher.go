package watch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Manager struct {
	registry     *registry.Registry
	store        *store.Store
	client       client.WithWatch
	logger       *zap.Logger
	mu           sync.Mutex
	started      map[string]struct{}
	lifecycleCtx context.Context
}

func NewManager(reg *registry.Registry, st *store.Store, kube client.WithWatch, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Manager{registry: reg, store: st, client: kube, logger: logger, started: map[string]struct{}{}}
}

func (m *Manager) SyncOnce(ctx context.Context, resource registry.Resource, namespace string) error {
	_, err := m.syncOnce(ctx, resource, namespace)
	return err
}

func (m *Manager) syncOnce(ctx context.Context, resource registry.Resource, namespace string) (string, error) {
	if resource.Scope == registry.Namespaced && namespace == "" {
		return "", fmt.Errorf("namespace is required for namespaced resource %q", resource.Alias)
	}

	list := resource.NewList()
	objectList, ok := list.(client.ObjectList)
	if !ok {
		return "", fmt.Errorf("resource %q list is not a client object list", resource.Alias)
	}

	storeNamespace := ""
	options := []client.ListOption{}
	if resource.Scope == registry.Namespaced {
		storeNamespace = namespace
		options = append(options, client.InNamespace(namespace))
	}

	if err := m.client.List(ctx, objectList, options...); err != nil {
		return "", err
	}

	items, err := meta.ExtractList(list)
	if err != nil {
		return "", err
	}

	objects := make([]client.Object, 0, len(items))
	for _, item := range items {
		object, ok := item.(client.Object)
		if !ok {
			return "", fmt.Errorf("resource %q list item is not a client object", resource.Alias)
		}
		objects = append(objects, object)
	}

	m.store.ReplaceScope(resource.Alias, storeNamespace, objects)
	return objectList.GetResourceVersion(), nil
}

func (m *Manager) StartClusterScoped(ctx context.Context) {
	m.mu.Lock()
	m.lifecycleCtx = ctx
	m.mu.Unlock()

	for _, resource := range m.registry.All() {
		if resource.Scope != registry.Cluster {
			continue
		}
		m.startScope(ctx, resource, "")
	}
}

func (m *Manager) RunScope(ctx context.Context, resource registry.Resource, namespace string) {
	if resource.Scope == registry.Namespaced && namespace == "" {
		m.logger.Error("run namespaced resource watch", zap.String("resource", resource.Alias), zap.Error(fmt.Errorf("namespace is required")))
		return
	}

	for ctx.Err() == nil {
		resourceVersion, err := m.syncOnce(ctx, resource, namespace)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			m.logger.Error("sync resource before watch", zap.String("resource", resource.Alias), zap.String("namespace", namespace), zap.Error(err))
			if !sleep(ctx, time.Second) {
				return
			}
			continue
		}

		list, ok := resource.NewList().(client.ObjectList)
		if !ok {
			m.logger.Error("resource list is not a client object list", zap.String("resource", resource.Alias))
			return
		}

		options := []client.ListOption{}
		if resource.Scope == registry.Namespaced {
			options = append(options, client.InNamespace(namespace))
		}
		if resourceVersion != "" {
			options = append(options, &client.ListOptions{Raw: &metav1.ListOptions{ResourceVersion: resourceVersion}})
		}

		watcher, err := m.client.Watch(ctx, list, options...)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			m.logger.Error("watch resource", zap.String("resource", resource.Alias), zap.String("namespace", namespace), zap.Error(err))
			if !sleep(ctx, time.Second) {
				return
			}
			continue
		}

		m.consume(ctx, resource, watcher)
		watcher.Stop()
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
	if namespace == "" {
		return fmt.Errorf("namespace is required for namespaced resource %q", alias)
	}
	if m.store.Ready(alias, namespace) {
		m.startScope(ctx, resource, namespace)
		return nil
	}
	if err := m.SyncOnce(ctx, resource, namespace); err != nil {
		return err
	}
	m.startScope(ctx, resource, namespace)
	return nil
}

func (m *Manager) startScope(ctx context.Context, resource registry.Resource, namespace string) {
	key := resource.Alias + "/" + namespace
	m.mu.Lock()
	if _, ok := m.started[key]; ok {
		m.mu.Unlock()
		return
	}
	watchCtx := ctx
	if m.lifecycleCtx != nil {
		watchCtx = m.lifecycleCtx
	}
	m.started[key] = struct{}{}
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.started, key)
			m.mu.Unlock()
		}()
		m.RunScope(watchCtx, resource, namespace)
	}()
}

func (m *Manager) consume(ctx context.Context, resource registry.Resource, watcher k8swatch.Interface) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				m.logger.Info("resource watch closed", zap.String("resource", resource.Alias))
				return
			}
			if event.Type == k8swatch.Error {
				m.logger.Error("resource watch error", zap.String("resource", resource.Alias), zap.Any("status", event.Object))
				return
			}
			object, ok := event.Object.(client.Object)
			if !ok {
				m.logger.Error("watch event object is not a client object", zap.String("resource", resource.Alias), zap.String("event", string(event.Type)))
				continue
			}
			switch event.Type {
			case k8swatch.Added, k8swatch.Modified:
				m.store.Upsert(resource.Alias, object)
			case k8swatch.Deleted:
				namespace := object.GetNamespace()
				if resource.Scope == registry.Cluster {
					namespace = ""
				}
				m.store.Delete(resource.Alias, namespace, object.GetName())
			}
		}
	}
}

func sleep(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
