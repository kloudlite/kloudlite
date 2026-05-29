package environment

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kloudlite/kloudlite/controllers/composition"
	"github.com/kloudlite/kloudlite/controllers/testutil"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestMakeStringSet tests the makeStringSet helper
func TestMakeStringSet(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected map[string]bool
	}{
		{
			name:     "empty slice",
			items:    []string{},
			expected: map[string]bool{},
		},
		{
			name:     "single item",
			items:    []string{"item1"},
			expected: map[string]bool{"item1": true},
		},
		{
			name:     "multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: map[string]bool{"item1": true, "item2": true, "item3": true},
		},
		{
			name:     "duplicate items",
			items:    []string{"item1", "item1", "item2"},
			expected: map[string]bool{"item1": true, "item2": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeStringSet(tt.items)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s: %v, got %v", k, v, result[k])
				}
			}
		})
	}
}

func TestFindEnvironmentForComposeResourceUsesOwnershipLabels(t *testing.T) {
	reconciler := &EnvironmentReconciler{}
	resource := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-config",
			Namespace: "env-dev",
			Labels: map[string]string{
				composition.DockerCompositionLabel:    "dev",
				composition.EnvironmentNamespaceLabel: "wm-alice",
				composition.ManagedLabel:              "true",
			},
		},
	}

	requests := reconciler.findEnvironmentForComposeResource(context.Background(), resource)

	if len(requests) != 1 {
		t.Fatalf("expected one reconcile request, got %#v", requests)
	}
	if requests[0] != (reconcile.Request{NamespacedName: types.NamespacedName{Name: "dev", Namespace: "wm-alice"}}) {
		t.Fatalf("expected request for owning environment, got %#v", requests[0])
	}
}

func TestFindEnvironmentForComposeResourceIgnoresUnownedResources(t *testing.T) {
	reconciler := &EnvironmentReconciler{}

	tests := []struct {
		name     string
		resource *corev1.ConfigMap
	}{
		{
			name: "no labels",
			resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "api-config",
				Namespace: "env-dev",
			}},
		},
		{
			name: "missing environment namespace label",
			resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "api-config",
				Namespace: "env-dev",
				Labels: map[string]string{
					composition.DockerCompositionLabel: "dev",
					composition.ManagedLabel:           "true",
				},
			}},
		},
		{
			name: "missing managed label",
			resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "api-config",
				Namespace: "env-dev",
				Labels: map[string]string{
					composition.DockerCompositionLabel:    "dev",
					composition.EnvironmentNamespaceLabel: "wm-alice",
				},
			}},
		},
		{
			name: "managed false",
			resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "api-config",
				Namespace: "env-dev",
				Labels: map[string]string{
					composition.DockerCompositionLabel:    "dev",
					composition.EnvironmentNamespaceLabel: "wm-alice",
					composition.ManagedLabel:              "false",
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := reconciler.findEnvironmentForComposeResource(context.Background(), tt.resource)
			if len(requests) != 0 {
				t.Fatalf("expected no reconcile requests, got %#v", requests)
			}
		})
	}
}

func TestReconcileComposeRemovesComposeResourcesAndClearsStatusWhenComposeRemoved(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = nil
	env.Status.ComposeStatus = &environmentsv1.CompositionStatus{
		State: environmentsv1.CompositionStateRunning,
		DeployedResources: &environmentsv1.DeployedResources{
			StatefulSets: []string{"api"},
			Services:     []string{"api"},
			PVCs:         []string{"api-data"},
		},
	}
	sts := composeTestStatefulSet(env, "api")
	svc := composeTestService(env, "api")
	pvc := composeTestPVC(env, "api-data")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), sts, svc, pvc)

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if reconciled {
		t.Fatalf("reconciled = true, want false")
	}
	assertComposeStatusCleared(t, ctx, reconciler, env)
	assertObjectDeleted(t, ctx, reconciler, sts)
	assertObjectDeleted(t, ctx, reconciler, svc)
	assertObjectDeleted(t, ctx, reconciler, pvc)
}

func TestReconcileComposeEmptyContentRemovesComposeResourcesAndClearsStatus(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: ""}
	env.Status.ComposeStatus = &environmentsv1.CompositionStatus{
		State: environmentsv1.CompositionStateRunning,
		DeployedResources: &environmentsv1.DeployedResources{
			StatefulSets: []string{"api"},
			Services:     []string{"api"},
			PVCs:         []string{"api-data"},
		},
	}
	sts := composeTestStatefulSet(env, "api")
	svc := composeTestService(env, "api")
	pvc := composeTestPVC(env, "api-data")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), sts, svc, pvc)

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if reconciled {
		t.Fatalf("reconciled = true, want false")
	}
	assertComposeStatusCleared(t, ctx, reconciler, env)
	assertObjectDeleted(t, ctx, reconciler, sts)
	assertObjectDeleted(t, ctx, reconciler, svc)
	assertObjectDeleted(t, ctx, reconciler, pvc)
}

func TestReconcileComposeRemovedWithNilStatusStillCleansComposeResources(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = nil
	env.Status.ComposeStatus = nil
	sts := composeTestStatefulSet(env, "api")
	svc := composeTestService(env, "api")
	pvc := composeTestPVC(env, "api-data")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), sts, svc, pvc)

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if reconciled {
		t.Fatalf("reconciled = true, want false")
	}
	if env.Status.ComposeStatus != nil {
		t.Fatalf("ComposeStatus = %#v, want nil", env.Status.ComposeStatus)
	}
	assertObjectDeleted(t, ctx, reconciler, sts)
	assertObjectDeleted(t, ctx, reconciler, svc)
	assertObjectDeleted(t, ctx, reconciler, pvc)
}

func TestReconcileComposeRemovedSkipsSameNameResourcesFromOtherEnvironmentNamespace(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = nil
	env.Status.ComposeStatus = nil
	sts := composeTestStatefulSet(env, "api")
	svc := composeTestService(env, "api")
	pvc := composeTestPVC(env, "api-data")
	setComposeResourceEnvironmentNamespace(sts, "wm-bob")
	setComposeResourceEnvironmentNamespace(svc, "wm-bob")
	setComposeResourceEnvironmentNamespace(pvc, "wm-bob")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), sts, svc, pvc)

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if reconciled {
		t.Fatalf("reconciled = true, want false")
	}
	assertObjectExists(t, ctx, reconciler, sts)
	assertObjectExists(t, ctx, reconciler, svc)
	assertObjectExists(t, ctx, reconciler, pvc)
}

func TestCleanupComposeResourcesSkipsSameNameResourcesFromOtherEnvironmentNamespace(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	sts := composeTestStatefulSet(env, "api")
	svc := composeTestService(env, "api")
	pvc := composeTestPVC(env, "api-data")
	setComposeResourceEnvironmentNamespace(sts, "wm-bob")
	setComposeResourceEnvironmentNamespace(svc, "wm-bob")
	setComposeResourceEnvironmentNamespace(pvc, "wm-bob")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), sts, svc, pvc)

	if err := reconciler.cleanupComposeResources(ctx, env, zap.NewNop()); err != nil {
		t.Fatalf("cleanupComposeResources error = %v, want nil", err)
	}

	assertObjectExists(t, ctx, reconciler, sts)
	assertObjectExists(t, ctx, reconciler, svc)
	assertObjectExists(t, ctx, reconciler, pvc)
}

func TestReconcileComposeRemovesServiceWhenComposeShrinks(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: composeContentWithServices("api", "worker")}
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env))

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("initial reconcile error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("initial reconciled = false, want true")
	}
	assertObjectExists(t, ctx, reconciler, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: env.Spec.TargetNamespace}})

	env.Spec.Compose.ComposeContent = composeContentWithServices("api")
	reconciled, err = reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("shrink reconcile error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("shrink reconciled = false, want true")
	}
	assertObjectExists(t, ctx, reconciler, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: env.Spec.TargetNamespace}})
	assertObjectDeleted(t, ctx, reconciler, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: env.Spec.TargetNamespace}})
	assertObjectDeleted(t, ctx, reconciler, &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: env.Spec.TargetNamespace}})
}

func TestReconcileComposeRetainsPVCWhenServiceRemovedButTopLevelVolumeRemains(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: composeContentWithMongoVolumeService()}
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env))

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("initial reconcile error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("initial reconciled = false, want true")
	}
	assertObjectExists(t, ctx, reconciler, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "mongodb-data", Namespace: env.Spec.TargetNamespace}})

	env.Spec.Compose.ComposeContent = composeContentWithOnlyMongoVolume()
	reconciled, err = reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("shrink reconcile error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("shrink reconciled = false, want true")
	}
	assertObjectDeleted(t, ctx, reconciler, &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "mongodb", Namespace: env.Spec.TargetNamespace}})
	assertObjectDeleted(t, ctx, reconciler, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "mongodb", Namespace: env.Spec.TargetNamespace}})
	assertObjectExists(t, ctx, reconciler, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "mongodb-data", Namespace: env.Spec.TargetNamespace}})
	assertStringSetEquals(t, env.Status.ComposeStatus.DeployedResources.PVCs, []string{"mongodb-data"})
}

func TestReconcileComposeDeletesPVCWhenTopLevelVolumeRemoved(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: composeContentWithMongoVolumeService()}
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env))

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("initial reconcile error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("initial reconciled = false, want true")
	}
	assertObjectExists(t, ctx, reconciler, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "mongodb-data", Namespace: env.Spec.TargetNamespace}})

	env.Spec.Compose.ComposeContent = composeContentWithServices("api")
	reconciled, err = reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("volume removal reconcile error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("volume removal reconciled = false, want true")
	}
	assertObjectDeleted(t, ctx, reconciler, &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "mongodb-data", Namespace: env.Spec.TargetNamespace}})
	assertStringSetEquals(t, env.Status.ComposeStatus.DeployedResources.PVCs, []string{})
}

func TestReconcileComposeSkipsRemovedResourcesWithMismatchedOwnershipLabels(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: composeContentWithServices("api")}
	env.Status.ComposeStatus = &environmentsv1.CompositionStatus{
		State: environmentsv1.CompositionStateRunning,
		DeployedResources: &environmentsv1.DeployedResources{
			StatefulSets: []string{"api", "worker"},
			Services:     []string{"api", "worker"},
			PVCs:         []string{"worker-data"},
		},
	}
	workerSts := composeTestStatefulSet(env, "worker")
	workerSvc := composeTestService(env, "worker")
	workerPVC := composeTestPVC(env, "worker-data")
	setComposeResourceEnvironmentNamespace(workerSts, "wm-bob")
	setComposeResourceEnvironmentNamespace(workerSvc, "wm-bob")
	setComposeResourceEnvironmentNamespace(workerPVC, "wm-bob")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), workerSts, workerSvc, workerPVC)

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("reconciled = false, want true")
	}
	assertObjectExists(t, ctx, reconciler, workerSts)
	assertObjectExists(t, ctx, reconciler, workerSvc)
	assertObjectExists(t, ctx, reconciler, workerPVC)
}

func TestCleanupRemovedComposeResourcesSkipsStatusNamesWithMismatchedOwnershipLabels(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	oldResources := &environmentsv1.DeployedResources{
		StatefulSets: []string{"worker"},
		Services:     []string{"worker"},
		PVCs:         []string{"worker-data"},
	}
	workerSts := composeTestStatefulSet(env, "worker")
	workerSvc := composeTestService(env, "worker")
	workerPVC := composeTestPVC(env, "worker-data")
	setComposeResourceEnvironmentNamespace(workerSts, "wm-bob")
	setComposeResourceEnvironmentNamespace(workerSvc, "wm-bob")
	setComposeResourceEnvironmentNamespace(workerPVC, "wm-bob")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), workerSts, workerSvc, workerPVC)

	if err := reconciler.cleanupRemovedComposeResources(ctx, env, oldResources, nil, nil, nil, zap.NewNop()); err != nil {
		t.Fatalf("cleanupRemovedComposeResources error = %v, want nil", err)
	}

	assertObjectExists(t, ctx, reconciler, workerSts)
	assertObjectExists(t, ctx, reconciler, workerSvc)
	assertObjectExists(t, ctx, reconciler, workerPVC)
}

func TestReconcileComposeCreatesAddedServiceWhenComposeGrows(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: composeContentWithServices("api", "worker")}
	env.Status.ComposeStatus = &environmentsv1.CompositionStatus{
		State: environmentsv1.CompositionStateRunning,
		DeployedResources: &environmentsv1.DeployedResources{
			StatefulSets: []string{"api"},
			Services:     []string{"api"},
		},
	}
	apiSts := composeTestStatefulSet(env, "api")
	apiSvc := composeTestService(env, "api")
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env), apiSts, apiSvc)

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if !reconciled {
		t.Fatalf("reconciled = false, want true")
	}
	assertObjectExists(t, ctx, reconciler, &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: env.Spec.TargetNamespace}})
	assertObjectExists(t, ctx, reconciler, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: env.Spec.TargetNamespace}})
	assertObjectExists(t, ctx, reconciler, &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: env.Spec.TargetNamespace}})
	assertObjectExists(t, ctx, reconciler, &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: env.Spec.TargetNamespace}})
	if env.Status.ComposeStatus == nil {
		t.Fatalf("ComposeStatus = nil, want status")
	}
	if env.Status.ComposeStatus.ServicesCount != 2 {
		t.Fatalf("ServicesCount = %d, want 2", env.Status.ComposeStatus.ServicesCount)
	}
	assertStringSetEquals(t, env.Status.ComposeStatus.DeployedResources.StatefulSets, []string{"api", "worker"})
	assertStringSetEquals(t, env.Status.ComposeStatus.DeployedResources.Services, []string{"api", "worker"})
}

func TestReconcileComposeApplyFailureReturnsErrorAndMarksFailed(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: composeContentWithServices("api")}
	applyErr := errors.New("create failed")
	reconciler := composeTestReconcilerWithInterceptor(t, interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			if _, ok := obj.(*appsv1.StatefulSet); ok {
				return applyErr
			}
			return c.Create(ctx, obj, opts...)
		},
	}, env, composeTestNamespace(env))

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err == nil {
		t.Fatalf("error = nil, want apply error")
	}
	if !strings.Contains(err.Error(), "failed to apply statefulset api") {
		t.Fatalf("error = %v, want StatefulSet apply context", err)
	}
	if !reconciled {
		t.Fatalf("reconciled = false, want true")
	}
	if env.Status.ComposeStatus == nil {
		t.Fatalf("ComposeStatus = nil, want failed status")
	}
	if env.Status.ComposeStatus.State != environmentsv1.CompositionStateFailed {
		t.Fatalf("ComposeStatus.State = %s, want %s", env.Status.ComposeStatus.State, environmentsv1.CompositionStateFailed)
	}
	if !strings.Contains(env.Status.ComposeStatus.Message, "failed to apply statefulset api") {
		t.Fatalf("ComposeStatus.Message = %q, want failed StatefulSet apply message", env.Status.ComposeStatus.Message)
	}
}

func TestReconcileComposeParseErrorReturnsErrorAndMarksFailed(t *testing.T) {
	ctx := context.Background()
	env := composeTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: "services:\n  api:\n    image: ["}
	reconciler := composeTestReconciler(t, env, composeTestNamespace(env))

	reconciled, err := reconciler.reconcileCompose(ctx, env, zap.NewNop())

	if err == nil {
		t.Fatalf("error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "parse compose") {
		t.Fatalf("error = %v, want parse compose context", err)
	}
	if !reconciled {
		t.Fatalf("reconciled = false, want true")
	}
	if env.Status.ComposeStatus == nil {
		t.Fatalf("ComposeStatus = nil, want failed status")
	}
	if env.Status.ComposeStatus.State != environmentsv1.CompositionStateFailed {
		t.Fatalf("ComposeStatus.State = %s, want %s", env.Status.ComposeStatus.State, environmentsv1.CompositionStateFailed)
	}
	if !strings.Contains(env.Status.ComposeStatus.Message, "Parse error") {
		t.Fatalf("ComposeStatus.Message = %q, want parse error", env.Status.ComposeStatus.Message)
	}
}

func composeTestEnvironment() *environmentsv1.Environment {
	return &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dev", Namespace: "wm-alice"},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "env-dev",
			Activated:       true,
		},
	}
}

func composeTestNamespace(env *environmentsv1.Environment) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: env.Spec.TargetNamespace}}
}

func composeTestReconciler(t *testing.T, objs ...client.Object) *EnvironmentReconciler {
	t.Helper()
	scheme := testutil.NewTestScheme()
	return &EnvironmentReconciler{
		Client: testutil.NewFakeClient(scheme, objs...).WithStatusSubresource(&environmentsv1.Environment{}).Build(),
		Scheme: scheme,
		Logger: zap.NewNop(),
	}
}

func composeTestReconcilerWithInterceptor(t *testing.T, interceptorFuncs interceptor.Funcs, objs ...client.Object) *EnvironmentReconciler {
	t.Helper()
	scheme := testutil.NewTestScheme()
	return &EnvironmentReconciler{
		Client: testutil.NewFakeClient(scheme, objs...).WithStatusSubresource(&environmentsv1.Environment{}).WithInterceptorFuncs(interceptorFuncs).Build(),
		Scheme: scheme,
		Logger: zap.NewNop(),
	}
}

func composeTestStatefulSet(env *environmentsv1.Environment, name string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{
		Name:      name,
		Namespace: env.Spec.TargetNamespace,
		Labels:    composeTestOwnershipLabels(env),
	}}
}

func composeTestService(env *environmentsv1.Environment, name string) *corev1.Service {
	return &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name:      name,
		Namespace: env.Spec.TargetNamespace,
		Labels:    composeTestOwnershipLabels(env),
	}}
}

func composeTestPVC(env *environmentsv1.Environment, name string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{
		Name:      name,
		Namespace: env.Spec.TargetNamespace,
		Labels:    composeTestOwnershipLabels(env),
	}}
}

func composeTestOwnershipLabels(env *environmentsv1.Environment) map[string]string {
	return map[string]string{
		composition.DockerCompositionLabel:    env.Name,
		composition.EnvironmentNamespaceLabel: env.Namespace,
	}
}

func setComposeResourceEnvironmentNamespace(obj client.Object, namespace string) {
	labels := obj.GetLabels()
	labels[composition.EnvironmentNamespaceLabel] = namespace
	obj.SetLabels(labels)
}

func assertComposeStatusCleared(t *testing.T, ctx context.Context, reconciler *EnvironmentReconciler, env *environmentsv1.Environment) {
	t.Helper()
	if env.Status.ComposeStatus != nil {
		t.Fatalf("in-memory ComposeStatus = %#v, want nil", env.Status.ComposeStatus)
	}
	fetched := &environmentsv1.Environment{}
	if err := reconciler.Get(ctx, client.ObjectKeyFromObject(env), fetched); err != nil {
		t.Fatalf("get environment: %v", err)
	}
	if fetched.Status.ComposeStatus != nil {
		t.Fatalf("persisted ComposeStatus = %#v, want nil", fetched.Status.ComposeStatus)
	}
}

func assertObjectDeleted(t *testing.T, ctx context.Context, reconciler *EnvironmentReconciler, obj client.Object) {
	t.Helper()
	key := client.ObjectKeyFromObject(obj)
	err := reconciler.Get(ctx, key, obj)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("get %T %s error = %v, want not found", obj, key, err)
	}
}

func assertObjectExists(t *testing.T, ctx context.Context, reconciler *EnvironmentReconciler, obj client.Object) {
	t.Helper()
	key := client.ObjectKeyFromObject(obj)
	if err := reconciler.Get(ctx, key, obj); err != nil {
		t.Fatalf("get %T %s: %v", obj, key, err)
	}
}

func assertStringSetEquals(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("items = %#v, want %#v", got, want)
	}
	gotSet := make(map[string]bool, len(got))
	for _, item := range got {
		gotSet[item] = true
	}
	for _, item := range want {
		if !gotSet[item] {
			t.Fatalf("items = %#v, want %#v", got, want)
		}
	}
}

func composeContentWithServices(names ...string) string {
	content := "services:\n"
	for _, name := range names {
		content += "  " + name + ":\n    image: nginx\n"
	}
	return content
}

func composeContentWithMongoVolumeService() string {
	return "services:\n" +
		"  mongodb:\n" +
		"    image: mongo:7\n" +
		"    volumes:\n" +
		"      - mongodb-data:/data/db\n" +
		"volumes:\n" +
		"  mongodb-data: {}\n"
}

func composeContentWithOnlyMongoVolume() string {
	return "services: {}\n" +
		"volumes:\n" +
		"  mongodb-data: {}\n"
}
