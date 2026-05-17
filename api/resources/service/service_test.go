package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListRequiresReadyCache(t *testing.T) {
	svc := New(registry.Default(), store.New(), fake.NewClientBuilder().Build())
	_, err := svc.List(context.Background(), "configmaps", "default", store.Selector{})
	if !IsKind(err, ErrCacheNotReady) {
		t.Fatalf("expected cache not ready error, got %v", err)
	}
}

func TestCreateWritesToKubernetesWithoutOptimisticStoreMutation(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	st := store.New()
	svc := New(registry.Default(), st, fake.NewClientBuilder().WithScheme(scheme).Build())
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}}

	created, err := svc.Create(context.Background(), "configmaps", "default", cm)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.GetName() != "app" {
		t.Fatalf("unexpected created object: %#v", created)
	}
	if _, ok := st.Get("configmaps", "default", "app"); ok {
		t.Fatal("store should not be updated optimistically")
	}
}

func TestListReturnsCachedObjectsWhenScopeIsReady(t *testing.T) {
	st := store.New()
	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}},
	})
	svc := New(registry.Default(), st, fake.NewClientBuilder().Build())

	items, err := svc.List(context.Background(), "configmaps", "default", store.Selector{})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(items) != 1 || items[0].GetName() != "app" {
		t.Fatalf("unexpected list result: %#v", items)
	}
}

func TestGetRequiresReadyCacheAndReturnsCachedObject(t *testing.T) {
	st := store.New()
	svc := New(registry.Default(), st, fake.NewClientBuilder().Build())

	_, err := svc.Get(context.Background(), "configmaps", "default", "app")
	if !IsKind(err, ErrCacheNotReady) {
		t.Fatalf("expected cache not ready error, got %v", err)
	}

	st.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}},
	})
	object, err := svc.Get(context.Background(), "configmaps", "default", "app")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if object.GetName() != "app" {
		t.Fatalf("unexpected object: %#v", object)
	}
}

func TestGetReturnsNotFoundWhenCachedObjectIsAbsent(t *testing.T) {
	st := store.New()
	st.ReplaceScope("configmaps", "default", nil)
	svc := New(registry.Default(), st, fake.NewClientBuilder().Build())

	_, err := svc.Get(context.Background(), "configmaps", "default", "missing")
	if !IsKind(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
	if got := HTTPStatus(err); got != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, got)
	}
}

func TestUnknownResourceAliasReturnsTypedError(t *testing.T) {
	svc := New(registry.Default(), store.New(), fake.NewClientBuilder().Build())

	if _, err := svc.List(context.Background(), "missing", "default", store.Selector{}); !IsKind(err, ErrUnknownResource) {
		t.Fatalf("expected unknown resource error from list, got %v", err)
	}
	if _, err := svc.Get(context.Background(), "missing", "default", "app"); !IsKind(err, ErrUnknownResource) {
		t.Fatalf("expected unknown resource error from get, got %v", err)
	}
	if _, err := svc.Create(context.Background(), "missing", "default", &corev1.ConfigMap{}); !IsKind(err, ErrUnknownResource) {
		t.Fatalf("expected unknown resource error from create, got %v", err)
	}
}

func TestCreateNamespacedResourceUsesPathNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	kube := fake.NewClientBuilder().WithScheme(scheme).Build()
	svc := New(registry.Default(), store.New(), kube)
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "body"}}

	created, err := svc.Create(context.Background(), "configmaps", "path", cm)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.GetNamespace() != "path" {
		t.Fatalf("expected created namespace from path, got %q", created.GetNamespace())
	}

	var stored corev1.ConfigMap
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "path", Name: "app"}, &stored); err != nil {
		t.Fatalf("expected object in path namespace: %v", err)
	}
}

func TestCreateClusterScopedResourceDoesNotForceNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
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
	})
	if err != nil {
		t.Fatal(err)
	}
	svc := New(reg, store.New(), fake.NewClientBuilder().WithScheme(scheme).Build())
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: "body"}}

	created, err := svc.Create(context.Background(), "nodes", "path", node)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.GetNamespace() != "" {
		t.Fatalf("expected cluster-scoped resource namespace to be empty, got %q", created.GetNamespace())
	}
}

func TestCreateSetsGVKWhenMissing(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	svc := New(registry.Default(), store.New(), fake.NewClientBuilder().WithScheme(scheme).Build())
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app"}}

	created, err := svc.Create(context.Background(), "configmaps", "default", cm)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if got, want := created.GetObjectKind().GroupVersionKind(), (schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}); got != want {
		t.Fatalf("expected GVK %s, got %s", want, got)
	}
}

func TestCreateRejectsObjectWithMismatchedGVK(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	svc := New(registry.Default(), store.New(), fake.NewClientBuilder().WithScheme(scheme).Build())
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "app"}}
	secret.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Version: "v1", Kind: "Secret"})

	_, err := svc.Create(context.Background(), "configmaps", "default", secret)
	if !IsKind(err, ErrBadRequest) {
		t.Fatalf("expected bad request error, got %v", err)
	}
}

func TestCreateRejectsNilObject(t *testing.T) {
	svc := New(registry.Default(), store.New(), fake.NewClientBuilder().Build())

	_, err := svc.Create(context.Background(), "configmaps", "default", nil)
	if !IsKind(err, ErrBadRequest) {
		t.Fatalf("expected bad request error, got %v", err)
	}
}

func TestPatchAndDeleteReturnNotImplementedErrors(t *testing.T) {
	svc := New(registry.Default(), store.New(), fake.NewClientBuilder().Build())

	if _, err := svc.Patch(context.Background(), "configmaps", "default", "app", &corev1.ConfigMap{}); !IsKind(err, ErrNotImplemented) {
		t.Fatalf("expected not implemented error from patch, got %v", err)
	}
	if err := svc.Delete(context.Background(), "configmaps", "default", "app"); !IsKind(err, ErrNotImplemented) {
		t.Fatalf("expected not implemented error from delete, got %v", err)
	}
}

func TestHTTPStatusMapsServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "unknown resource", err: UnknownResource("missing"), want: http.StatusNotFound},
		{name: "wrong scope", err: NewError(ErrWrongScope, "wrong scope", nil), want: http.StatusBadRequest},
		{name: "not found", err: NewError(ErrNotFound, "not found", nil), want: http.StatusNotFound},
		{name: "bad request", err: NewError(ErrBadRequest, "bad request", nil), want: http.StatusBadRequest},
		{name: "cache not ready", err: NewError(ErrCacheNotReady, "cache not ready", nil), want: http.StatusServiceUnavailable},
		{name: "not implemented", err: NewError(ErrNotImplemented, "not implemented", nil), want: http.StatusNotImplemented},
		{name: "unknown error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatus(tt.err); got != tt.want {
				t.Fatalf("expected status %d, got %d", tt.want, got)
			}
		})
	}
}
