package platformscoped

import (
	"context"
	"errors"
	"testing"
	"time"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSkipCloudPermissionValidation(t *testing.T) {
	t.Setenv("KLOUDLITE_SKIP_CLOUD_PERMISSION_VALIDATION", "true")

	if !skipCloudPermissionValidation() {
		t.Fatal("expected cloud permission validation to be skipped")
	}
}

func TestSkipCloudPermissionValidationDefaultsToFalse(t *testing.T) {
	if skipCloudPermissionValidation() {
		t.Fatal("expected cloud permission validation to run by default")
	}
}

func TestSetupCloudProviderRejectsUnsupportedProvider(t *testing.T) {
	_, err := setupCloudProvider(context.Background(), Env{CloudProvider: workmachinev1.CloudProvider("unknown")})
	if err == nil {
		t.Fatal("expected unsupported cloud provider error")
	}
}

func TestValidateProviderPermissionsHonorsSkipToggle(t *testing.T) {
	t.Setenv("KLOUDLITE_SKIP_CLOUD_PERMISSION_VALIDATION", "true")
	provider := &recordingProvider{validateErr: errors.New("should be skipped")}

	if err := validateProviderPermissions(context.Background(), provider); err != nil {
		t.Fatalf("expected validation to be skipped, got %v", err)
	}
	if provider.validateCalls != 0 {
		t.Fatalf("expected no validation calls, got %d", provider.validateCalls)
	}
}

func TestPlatformScopedLifecycleStepNamesPreserveOrder(t *testing.T) {
	r := &PlatformScopedReconciler{}
	steps := r.lifecycleSteps()

	got := make([]string, 0, len(steps))
	for _, step := range steps {
		got = append(got, step.Name)
	}

	want := []string{
		"handle-machine-type-change",
		"handle-node-reboot-request",
		"cleanup-workmachine-namespace",
		"setup-cloud-machine",
		"ensure-workmachine-manager",
	}
	wantConditions := []string{
		workmachineshared.ConditionCloudMachineProvisioned,
		workmachineshared.ConditionCloudMachineRunning,
		"",
		workmachineshared.ConditionNodeJoined,
		"",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d steps, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d: expected %q, got %q", i, want[i], got[i])
		}
		if steps[i].Condition != wantConditions[i] {
			t.Fatalf("step %d condition: expected %q, got %q", i, wantConditions[i], steps[i].Condition)
		}
	}
}

func TestEnsureWorkMachineManagerCreatesPerWorkMachineStatefulSet(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}
	wm.Status.MachineID = "machine-id"
	session := workmachineshared.NewStatusSession(wm)
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
		Scheme: scheme,
		env:    Env{PodNamespace: "kloudlite", WorkMachineManagerImage: "workmachine-manager:test"},
	}

	result, err := r.ensureWorkMachineManager(context.Background(), session)
	if err != nil {
		t.Fatalf("ensureWorkMachineManager returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("ensureWorkMachineManager result = %#v, want no requeue", result)
	}

	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: workMachineManagerName("wm"), Namespace: "kloudlite"}, statefulSet); err != nil {
		t.Fatalf("expected per-workmachine statefulset: %v", err)
	}
	if statefulSet.Spec.Replicas == nil || *statefulSet.Spec.Replicas != 1 {
		t.Fatalf("replicas = %#v, want 1", statefulSet.Spec.Replicas)
	}
	if statefulSet.Spec.Template.Spec.NodeSelector["kloudlite.io/workmachine"] != "wm" {
		t.Fatalf("nodeSelector = %#v, want kloudlite.io/workmachine=wm", statefulSet.Spec.Template.Spec.NodeSelector)
	}
	container := statefulSet.Spec.Template.Spec.Containers[0]
	if container.Name != "workmachine-manager" {
		t.Fatalf("container name = %q, want workmachine-manager", container.Name)
	}
	if len(container.Command) != 1 || container.Command[0] != "/app/workmachine-manager" {
		t.Fatalf("container command = %#v, want /app/workmachine-manager", container.Command)
	}
	if len(container.Args) != 2 || container.Args[0] != "server" || container.Args[1] != "workmachine-manager" {
		t.Fatalf("container args = %#v, want server workmachine-manager", container.Args)
	}
	if container.Env[0].Name != "NODE_NAME" {
		t.Fatalf("expected manager to expose NODE_NAME env, got %#v", statefulSet.Spec.Template.Spec.Containers[0].Env)
	}

	deployment := &appsv1.Deployment{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: machineScopedControllerName("wm"), Namespace: "kloudlite"}, deployment); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected no per-workmachine deployment, got err=%v", err)
	}
}

func TestCleanupWorkMachineNamespaceDeletesNamespaceAndRequeues(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "wm"},
		Spec:       workmachinev1.WorkMachineSpec{TargetNamespace: "wm-wm"},
	}
	r := &PlatformScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "wm-wm"}}).Build(),
		Scheme: scheme,
	}

	result, err := r.cleanupWorkMachineNamespace(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("cleanupWorkMachineNamespace returned error: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatalf("cleanupWorkMachineNamespace result = %#v, want requeue while namespace deletes", result)
	}

	ns := &corev1.Namespace{}
	if err := r.Get(context.Background(), client.ObjectKey{Name: "wm-wm"}, ns); client.IgnoreNotFound(err) != nil || err == nil {
		t.Fatalf("expected namespace deletion to be requested, got err=%v", err)
	}
}

func TestCleanupWorkMachineNamespaceReturnsWhenNamespaceGone(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	wm := &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "wm"},
		Spec:       workmachinev1.WorkMachineSpec{TargetNamespace: "wm-wm"},
	}
	r := &PlatformScopedReconciler{Client: fake.NewClientBuilder().WithScheme(scheme).Build(), Scheme: scheme}

	result, err := r.cleanupWorkMachineNamespace(context.Background(), workmachineshared.NewStatusSession(wm))
	if err != nil {
		t.Fatalf("cleanupWorkMachineNamespace returned error: %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("cleanupWorkMachineNamespace result = %#v, want no requeue", result)
	}
}

func TestReconcileReturnPreservesReconcileErrorOverPatchConflictRequeue(t *testing.T) {
	reconcileErr := errors.New("real reconcile failure")
	conflictRequeue := ctrl.Result{RequeueAfter: 500 * time.Millisecond}

	result, err := reconcileReturn(ctrl.Result{}, reconcileErr, conflictRequeue, nil)

	if !errors.Is(err, reconcileErr) {
		t.Fatalf("error = %v, want reconcile error %v", err, reconcileErr)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want zero result when preserving reconcile error", result)
	}
}

func TestReconcileReturnUsesPatchConflictRequeueWithoutReconcileError(t *testing.T) {
	conflictRequeue := ctrl.Result{RequeueAfter: 500 * time.Millisecond}

	result, err := reconcileReturn(ctrl.Result{}, nil, conflictRequeue, nil)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != conflictRequeue {
		t.Fatalf("result = %#v, want patch conflict requeue %#v", result, conflictRequeue)
	}
}

type recordingProvider struct {
	validateCalls int
	validateErr   error
}

func (p *recordingProvider) ValidatePermissions(context.Context) error {
	p.validateCalls++
	return p.validateErr
}

func (p *recordingProvider) CreateMachine(context.Context, *workmachinev1.WorkMachine) (*workmachinev1.MachineInfo, error) {
	return nil, nil
}

func (p *recordingProvider) GetMachineStatus(context.Context, string) (*workmachinev1.MachineInfo, error) {
	return nil, nil
}

func (p *recordingProvider) StartMachine(context.Context, string) error  { return nil }
func (p *recordingProvider) StopMachine(context.Context, string) error   { return nil }
func (p *recordingProvider) RebootMachine(context.Context, string) error { return nil }
func (p *recordingProvider) IncreaseVolumeSize(context.Context, string, int32) error {
	return nil
}
func (p *recordingProvider) ChangeMachine(context.Context, string, string) error { return nil }
func (p *recordingProvider) DeleteMachine(context.Context, string) error         { return nil }
