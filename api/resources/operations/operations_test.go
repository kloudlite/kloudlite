package operations

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/api/resources/registry"
	resourceservice "github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestListEnsuresNamespacedScopeAndReturnsDirtyState(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	st := store.New()
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", Labels: map[string]string{"app": "demo"}}},
	).Build()
	st.MarkDirty("configmaps", "default", "app", "pending_patch_confirmation", map[string]string{"app": "demo"})
	ensurer := &recordingEnsurer{}
	ops := New(resourceservice.New(registry.Default(), st, kube), registry.Default(), st, ensurer)

	result, err := ops.List(context.Background(), ListRequest{Resource: "configmaps", Namespace: "default", Scope: registry.Namespaced})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if ensurer.alias != "configmaps" || ensurer.namespace != "default" {
		t.Fatalf("expected namespace ensurer to be called, got alias=%q namespace=%q", ensurer.alias, ensurer.namespace)
	}
	if len(result.Items) != 1 || result.Items[0].GetName() != "app" {
		t.Fatalf("unexpected items: %#v", result.Items)
	}
	if len(result.Dirty) != 1 || result.Dirty[0].Name != "app" {
		t.Fatalf("expected dirty state for app, got %#v", result.Dirty)
	}
}

func TestCreateMarksObjectDirtyThroughOperation(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	st := store.New()
	ops := New(resourceservice.New(registry.Default(), st, fake.NewClientBuilder().WithScheme(scheme).Build()), registry.Default(), st, nil)

	result, err := ops.Create(context.Background(), MutateRequest{Resource: "configmaps", Namespace: "default", Scope: registry.Namespaced, Object: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app"}}})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if result.Object.GetNamespace() != "default" {
		t.Fatalf("expected namespace default, got %q", result.Object.GetNamespace())
	}
	if result.Dirty == nil || result.Dirty.Name != "app" {
		t.Fatalf("expected dirty state for app, got %#v", result.Dirty)
	}
}

func TestMachineTypesAreCodeManagedReadOnlyOperations(t *testing.T) {
	ops := New(resourceservice.New(registry.Default(), store.New(), fake.NewClientBuilder().Build()), registry.Default(), store.New(), nil, WithCloudProvider(func() string { return "aws" }))

	list, err := ops.List(context.Background(), ListRequest{Resource: "machinetypes", Scope: registry.Cluster})
	if err != nil {
		t.Fatalf("list machine types failed: %v", err)
	}
	if len(list.Items) == 0 {
		t.Fatal("expected code-managed machine types")
	}
	get, err := ops.Get(context.Background(), GetRequest{Resource: "machinetypes", Name: list.Items[0].GetName(), Scope: registry.Cluster})
	if err != nil {
		t.Fatalf("get machine type failed: %v", err)
	}
	if get.Object.GetName() != list.Items[0].GetName() {
		t.Fatalf("unexpected machine type: %s", get.Object.GetName())
	}
	_, err = ops.Create(context.Background(), MutateRequest{Resource: "machinetypes", Scope: registry.Cluster, Object: list.Items[0]})
	if !resourceservice.IsKind(err, resourceservice.ErrNotImplemented) {
		t.Fatalf("expected read-only machine types error, got %v", err)
	}
}

type recordingEnsurer struct {
	alias     string
	namespace string
}

func (e *recordingEnsurer) EnsureNamespaced(_ context.Context, alias string, namespace string) error {
	e.alias = alias
	e.namespace = namespace
	return nil
}
