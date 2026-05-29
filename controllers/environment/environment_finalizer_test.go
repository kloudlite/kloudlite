package environment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/controllers/testutil"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnvironmentReconciler_Reconcile_AddFinalizerConflictRequeuesWithoutError(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-env", Namespace: "wm-alice"},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "test@example.com",
		},
	}
	conflict := apierrors.NewConflict(schema.GroupResource{Group: environmentsv1.SchemeGroupVersion.Group, Resource: "environments"}, "test-env", errors.New("stale object"))
	k8sClient := testutil.NewFakeClient(scheme, env).WithInterceptorFuncs(interceptor.Funcs{
		Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
			return conflict
		},
	}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	result, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-alice"}})
	if err != nil {
		t.Fatalf("Reconcile error = %v, want nil for conflict requeue", err)
	}
	if result.RequeueAfter != 500*time.Millisecond {
		t.Fatalf("RequeueAfter = %s, want 500ms", result.RequeueAfter)
	}
}

func TestEnvironmentReconciler_Reconcile_RemoveFinalizerConflictRequeuesWithoutError(t *testing.T) {
	scheme := testutil.NewTestScheme()
	now := metav1.Now()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-env",
			Namespace:         "wm-alice",
			DeletionTimestamp: &now,
			Finalizers:        []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "deleted-namespace",
			OwnedBy:         "test@example.com",
		},
		Status: environmentsv1.EnvironmentStatus{State: environmentsv1.EnvironmentStateDeleting},
	}
	conflict := apierrors.NewConflict(schema.GroupResource{Group: environmentsv1.SchemeGroupVersion.Group, Resource: "environments"}, "test-env", errors.New("stale object"))
	k8sClient := testutil.NewFakeClient(scheme, env).WithInterceptorFuncs(interceptor.Funcs{
		Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			return conflict
		},
	}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	result, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-alice"}})
	if err != nil {
		t.Fatalf("Reconcile error = %v, want nil for conflict requeue", err)
	}
	if result.RequeueAfter != 500*time.Millisecond {
		t.Fatalf("RequeueAfter = %s, want 500ms", result.RequeueAfter)
	}
}
