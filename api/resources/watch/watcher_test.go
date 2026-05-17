package watch

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

func TestEnsureNamespacedDoesNothingWhenScopeAlreadyReady(t *testing.T) {
	st := store.New()
	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cached", Namespace: "default"}},
	})
	kube := newFakeClient(t, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "live", Namespace: "default"}})
	counting := &countingClient{WithWatch: kube}
	mgr := NewManager(registry.Default(), st, counting, zap.NewNop())

	if err := mgr.EnsureNamespaced(context.Background(), "configmaps", "default"); err != nil {
		t.Fatalf("ensure namespaced failed: %v", err)
	}

	if counting.listCalls != 0 {
		t.Fatalf("expected no list calls, got %d", counting.listCalls)
	}
	if _, ok := st.Get("configmaps", "default", "cached"); !ok {
		t.Fatal("expected existing cached configmap to remain")
	}
	if _, ok := st.Get("configmaps", "default", "live"); ok {
		t.Fatal("expected live configmap not to be loaded")
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

	mgr.StartClusterScoped(context.Background())

	if !st.Ready("nodes", "") {
		t.Fatal("expected nodes cluster scope to be ready")
	}
	items := st.List("nodes", "", store.Selector{})
	if len(items) != 1 || items[0].GetName() != "worker" {
		t.Fatalf("unexpected cached nodes: %#v", items)
	}
	if st.Ready("configmaps", "") {
		t.Fatal("did not expect namespaced configmaps to be synced as cluster-scoped")
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
