package watch

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSyncOncePopulatesAndMarksReadyForNamespacedConfigMaps(t *testing.T) {
	kube := newFakeClient(t,
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "other"}},
	)
	counting := &countingClient{WithWatch: kube}
	st := store.New()
	resource := mustResource(t, registry.Default(), "configmaps")
	mgr := NewManager(registry.Default(), st, counting, zap.NewNop())

	if err := mgr.SyncOnce(context.Background(), resource, "default"); err != nil {
		t.Fatalf("sync once failed: %v", err)
	}

	if counting.listCalls != 1 {
		t.Fatalf("expected one list call, got %d", counting.listCalls)
	}
	if !st.Ready("configmaps", "default") {
		t.Fatal("expected configmaps/default scope to be ready")
	}
	items := st.List("configmaps", "default", store.Selector{})
	if len(items) != 1 || items[0].GetName() != "app" {
		t.Fatalf("unexpected cached configmaps: %#v", items)
	}
}

func TestSyncOnceReplacesOldStoreContents(t *testing.T) {
	kube := newFakeClient(t, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "new", Namespace: "default"}})
	st := store.New()
	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "old", Namespace: "default"}},
	})
	mgr := NewManager(registry.Default(), st, kube, zap.NewNop())
	resource := mustResource(t, registry.Default(), "configmaps")

	if err := mgr.SyncOnce(context.Background(), resource, "default"); err != nil {
		t.Fatalf("sync once failed: %v", err)
	}

	if _, ok := st.Get("configmaps", "default", "old"); ok {
		t.Fatal("expected old configmap to be replaced")
	}
	if _, ok := st.Get("configmaps", "default", "new"); !ok {
		t.Fatal("expected new configmap to be cached")
	}
}

func TestEnsureNamespacedStartsOneContinuousWatchForReadyScope(t *testing.T) {
	fakeWatcher := k8swatch.NewFake()
	st := store.New()
	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cached", Namespace: "default"}},
	})
	watching := &watchingClient{WithWatch: newFakeClient(t), watcher: fakeWatcher}
	mgr := NewManager(registry.Default(), st, watching, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mgr.EnsureNamespaced(ctx, "configmaps", "default"); err != nil {
		t.Fatalf("ensure namespaced failed: %v", err)
	}
	if err := mgr.EnsureNamespaced(ctx, "configmaps", "default"); err != nil {
		t.Fatalf("ensure namespaced failed second time: %v", err)
	}
	waitFor(t, func() bool { return watching.watchCalls.Load() == 1 })
	assertStable(t, func() bool { return watching.watchCalls.Load() == 1 })

	fakeWatcher.Add(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}})
	waitFor(t, func() bool {
		_, ok := st.Get("configmaps", "default", "app")
		return ok
	})
	if _, ok := st.Get("configmaps", "default", "cached"); ok {
		t.Fatal("expected continuous watch startup sync to replace stale cached object")
	}
}

func TestNamespacedSyncRejectsEmptyNamespaceBeforeList(t *testing.T) {
	st := store.New()
	counting := &countingClient{WithWatch: newFakeClient(t, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}})}
	reg := registry.Default()
	mgr := NewManager(reg, st, counting, zap.NewNop())
	resource := mustResource(t, reg, "configmaps")

	if err := mgr.SyncOnce(context.Background(), resource, ""); err == nil {
		t.Fatal("expected empty namespace sync to fail")
	}
	if err := mgr.EnsureNamespaced(context.Background(), "configmaps", ""); err == nil {
		t.Fatal("expected empty namespace ensure to fail")
	}

	if counting.listCalls != 0 {
		t.Fatalf("expected no list calls, got %d", counting.listCalls)
	}
	if st.Ready("configmaps", "") {
		t.Fatal("expected configmaps empty namespace scope to remain not ready")
	}
}

func TestEnsureNamespacedRejectsUnknownAliasAndClusterScopedAlias(t *testing.T) {
	reg := newTestRegistry(t)
	mgr := NewManager(reg, store.New(), newFakeClient(t), zap.NewNop())

	if err := mgr.EnsureNamespaced(context.Background(), "missing", "default"); err == nil {
		t.Fatal("expected unknown alias error")
	}
	if err := mgr.EnsureNamespaced(context.Background(), "nodes", "default"); err == nil {
		t.Fatal("expected cluster-scoped alias error")
	}
}

func TestStartClusterScopedSyncsClusterScopedResources(t *testing.T) {
	reg := newTestRegistry(t)
	st := store.New()
	mgr := NewManager(reg, st, newFakeClient(t, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}}), zap.NewNop())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr.StartClusterScoped(ctx)
	waitFor(t, func() bool { return st.Ready("nodes", "") })

	items := st.List("nodes", "", store.Selector{})
	if len(items) != 1 || items[0].GetName() != "worker" {
		t.Fatalf("unexpected cached nodes: %#v", items)
	}
	if st.Ready("configmaps", "") {
		t.Fatal("did not expect namespaced configmaps to be synced as cluster-scoped")
	}
}

func TestRunScopeAppliesWatchEventsUntilContextCancellation(t *testing.T) {
	fakeWatcher := k8swatch.NewFake()
	st := store.New()
	watching := &watchingClient{WithWatch: newFakeClient(t), watcher: fakeWatcher}
	mgr := NewManager(registry.Default(), st, watching, zap.NewNop())
	resource := mustResource(t, registry.Default(), "configmaps")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		mgr.RunScope(ctx, resource, "default")
	}()

	waitFor(t, func() bool { return watching.watchCalls.Load() == 1 && st.Ready("configmaps", "default") })
	fakeWatcher.Add(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}, Data: map[string]string{"version": "one"}})
	waitFor(t, func() bool {
		object, ok := st.Get("configmaps", "default", "app")
		return ok && object.(*corev1.ConfigMap).Data["version"] == "one"
	})

	fakeWatcher.Modify(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}, Data: map[string]string{"version": "two"}})
	waitFor(t, func() bool {
		object, ok := st.Get("configmaps", "default", "app")
		return ok && object.(*corev1.ConfigMap).Data["version"] == "two"
	})

	fakeWatcher.Delete(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}})
	waitFor(t, func() bool {
		_, ok := st.Get("configmaps", "default", "app")
		return !ok
	})

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected RunScope to stop after context cancellation")
	}
}

func TestRunScopeRetriesAfterWatchErrorEvent(t *testing.T) {
	firstWatcher := k8swatch.NewFake()
	secondWatcher := k8swatch.NewFake()
	st := store.New()
	watching := newSequenceWatchingClient(t, firstWatcher, secondWatcher)
	mgr := NewManager(registry.Default(), st, watching, zap.NewNop())
	resource := mustResource(t, registry.Default(), "configmaps")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		mgr.RunScope(ctx, resource, "default")
	}()

	waitFor(t, func() bool { return watching.watchCalls.Load() == 1 })
	firstWatcher.Error(&metav1.Status{Reason: metav1.StatusReasonInternalError, Message: "boom"})
	waitFor(t, func() bool { return watching.watchCalls.Load() == 2 && watching.listCalls.Load() >= 2 })

	secondWatcher.Add(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "after-error", Namespace: "default"}})
	waitFor(t, func() bool {
		_, ok := st.Get("configmaps", "default", "after-error")
		return ok
	})

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected RunScope to stop after context cancellation")
	}
}

type countingClient struct {
	client.WithWatch
	listCalls int
}

func (c *countingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	c.listCalls++
	return c.WithWatch.List(ctx, list, opts...)
}

type watchingClient struct {
	client.WithWatch
	watcher    *k8swatch.FakeWatcher
	watchCalls atomic.Int32
}

func (c *watchingClient) Watch(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (k8swatch.Interface, error) {
	_ = ctx
	_ = list
	_ = opts
	c.watchCalls.Add(1)
	return c.watcher, nil
}

type sequenceWatchingClient struct {
	client.WithWatch
	watchers   []*k8swatch.FakeWatcher
	mu         sync.Mutex
	watchCalls atomic.Int32
	listCalls  atomic.Int32
}

func newSequenceWatchingClient(t *testing.T, watchers ...*k8swatch.FakeWatcher) *sequenceWatchingClient {
	t.Helper()
	return &sequenceWatchingClient{WithWatch: newFakeClient(t), watchers: watchers}
}

func (c *sequenceWatchingClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	c.listCalls.Add(1)
	return c.WithWatch.List(ctx, list, opts...)
}

func (c *sequenceWatchingClient) Watch(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (k8swatch.Interface, error) {
	_ = ctx
	_ = list
	_ = opts
	call := int(c.watchCalls.Add(1))
	c.mu.Lock()
	defer c.mu.Unlock()
	if call <= len(c.watchers) {
		return c.watchers[call-1], nil
	}
	return k8swatch.NewFake(), nil
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}

func assertStable(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(50 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !condition() {
			t.Fatal("condition did not remain stable")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func newFakeClient(t *testing.T, objects ...client.Object) client.WithWatch {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func mustResource(t *testing.T, reg *registry.Registry, alias string) registry.Resource {
	t.Helper()
	resource, ok := reg.Get(alias)
	if !ok {
		t.Fatalf("expected resource %q to be registered", alias)
	}
	return resource
}

func newTestRegistry(t *testing.T) *registry.Registry {
	t.Helper()
	reg, err := registry.New([]registry.Resource{
		{
			Alias:     "nodes",
			Group:     corev1.SchemeGroupVersion.Group,
			Version:   corev1.SchemeGroupVersion.Version,
			Kind:      "Node",
			Plural:    "nodes",
			Scope:     registry.Cluster,
			NewObject: func() runtime.Object { return &corev1.Node{} },
			NewList:   func() runtime.Object { return &corev1.NodeList{} },
		},
		{
			Alias:     "configmaps",
			Group:     corev1.SchemeGroupVersion.Group,
			Version:   corev1.SchemeGroupVersion.Version,
			Kind:      "ConfigMap",
			Plural:    "configmaps",
			Scope:     registry.Namespaced,
			NewObject: func() runtime.Object { return &corev1.ConfigMap{} },
			NewList:   func() runtime.Object { return &corev1.ConfigMapList{} },
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return reg
}
