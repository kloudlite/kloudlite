package environment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/controllers/composition"
	"github.com/kloudlite/kloudlite/controllers/testutil"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type lifecycleFailingClient struct {
	client.Client
	failCreate bool
	failDelete bool
	failUpdate bool
}

func (c *lifecycleFailingClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if c.failCreate {
		return errors.New("create failed")
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c *lifecycleFailingClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if c.failDelete {
		return errors.New("delete failed")
	}
	return c.Client.Delete(ctx, obj, opts...)
}

func (c *lifecycleFailingClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if c.failUpdate {
		return errors.New("update failed")
	}
	return c.Client.Update(ctx, obj, opts...)
}

func TestReconcileEnvironmentReturnPreservesReconcileError(t *testing.T) {
	reconcileErr := errors.New("reconcile failed")
	reconcileResult := ctrl.Result{RequeueAfter: time.Second}
	patchResult := ctrl.Result{RequeueAfter: 500 * time.Millisecond}

	result, err := reconcileEnvironmentReturn(reconcileResult, reconcileErr, patchResult, nil)

	if !errors.Is(err, reconcileErr) {
		t.Fatalf("error = %v, want %v", err, reconcileErr)
	}
	if result != reconcileResult {
		t.Fatalf("result = %#v, want %#v", result, reconcileResult)
	}
}

func TestReconcileEnvironmentReturnUsesPatchRequeue(t *testing.T) {
	reconcileResult := ctrl.Result{}
	patchResult := ctrl.Result{RequeueAfter: 500 * time.Millisecond}

	result, err := reconcileEnvironmentReturn(reconcileResult, nil, patchResult, nil)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != patchResult {
		t.Fatalf("result = %#v, want %#v", result, patchResult)
	}
}

func TestLifecycleStepsReturnsRequiredStepsInOrder(t *testing.T) {
	reconciler := &EnvironmentReconciler{}

	steps := reconciler.lifecycleSteps()

	got := make([]string, 0, len(steps))
	for _, step := range steps {
		got = append(got, step.Name)
	}
	want := []string{
		"ensure-target-namespace",
		"ensure-network-policy",
		"reconcile-compose",
		"reconcile-activation",
	}
	if !equalStringSlices(got, want) {
		t.Fatalf("steps = %#v, want %#v", got, want)
	}

	for _, step := range steps {
		if step.Condition == "" {
			t.Fatalf("step %s condition is empty", step.Name)
		}
		if step.OnCreate == nil {
			t.Fatalf("step %s OnCreate is nil", step.Name)
		}
	}
	if steps[0].OnDelete == nil {
		t.Fatalf("step %s OnDelete is nil", steps[0].Name)
	}
	if steps[2].OnDelete == nil {
		t.Fatalf("step %s OnDelete is nil", steps[2].Name)
	}
	if steps[3].OnDelete == nil {
		t.Fatalf("step %s OnDelete is nil", steps[3].Name)
	}
}

func TestLifecycleCreateRunsOnCreateStepsInOrderAndStopsOnResult(t *testing.T) {
	var calls []string
	stopResult := ctrl.Result{RequeueAfter: time.Second}
	steps := []environmentLifecycleStep{
		{Name: "first", OnCreate: lifecycleTestStep("first", &calls, ctrl.Result{}, nil)},
		{Name: "second", OnCreate: lifecycleTestStep("second", &calls, stopResult, nil)},
		{Name: "third", OnCreate: lifecycleTestStep("third", &calls, ctrl.Result{}, nil)},
	}
	session := NewEnvironmentStatusSession(&environmentsv1.Environment{})

	result, err := reconcileEnvironmentLifecycle(context.Background(), session, steps)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != stopResult {
		t.Fatalf("result = %#v, want %#v", result, stopResult)
	}
	wantCalls := []string{"first", "second"}
	if !equalStringSlices(calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", calls, wantCalls)
	}
}

func TestLifecycleCreateStopsOnError(t *testing.T) {
	var calls []string
	stepErr := errors.New("step failed")
	steps := []environmentLifecycleStep{
		{Name: "first", OnCreate: lifecycleTestStep("first", &calls, ctrl.Result{}, nil)},
		{Name: "second", OnCreate: lifecycleTestStep("second", &calls, ctrl.Result{}, stepErr)},
		{Name: "third", OnCreate: lifecycleTestStep("third", &calls, ctrl.Result{}, nil)},
	}
	session := NewEnvironmentStatusSession(&environmentsv1.Environment{})

	_, err := reconcileEnvironmentLifecycle(context.Background(), session, steps)

	if !errors.Is(err, stepErr) {
		t.Fatalf("error = %v, want %v", err, stepErr)
	}
	wantCalls := []string{"first", "second"}
	if !equalStringSlices(calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", calls, wantCalls)
	}
}

func TestLifecycleDeletionRunsOnDeleteStepsInReverseOrderAndMarksCleanupComplete(t *testing.T) {
	var calls []string
	steps := []environmentLifecycleStep{
		{Name: "first", OnDelete: lifecycleTestStep("first", &calls, ctrl.Result{}, nil)},
		{Name: "second", OnDelete: lifecycleTestStep("second", &calls, ctrl.Result{}, nil)},
		{Name: "third", OnDelete: lifecycleTestStep("third", &calls, ctrl.Result{}, nil)},
	}
	now := metav1.Now()
	session := NewEnvironmentStatusSession(&environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &now},
	})

	result, err := reconcileEnvironmentLifecycle(context.Background(), session, steps)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want zero", result)
	}
	wantCalls := []string{"third", "second", "first"}
	if !equalStringSlices(calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", calls, wantCalls)
	}
	condition := meta.FindStatusCondition(session.Object().Status.Conditions, EnvironmentConditionCleanupComplete)
	if condition == nil {
		t.Fatalf("CleanupComplete condition missing")
	}
	if condition.Status != metav1.ConditionTrue {
		t.Fatalf("CleanupComplete status = %s, want %s", condition.Status, metav1.ConditionTrue)
	}
}

func TestEnsureNetworkPolicyStepMarksReadyOnSuccess(t *testing.T) {
	env := lifecycleTestEnvironment()
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env))
	session := NewEnvironmentStatusSession(env)

	result, err := reconciler.ensureNetworkPolicyStep(context.Background(), session)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want zero", result)
	}
	assertCondition(t, env, EnvironmentConditionNetworkPolicyReady, metav1.ConditionTrue, EnvironmentReasonReconciled)
}

func TestEnsureNetworkPolicyDoesNotMarkNetworkPolicyReadyDirectly(t *testing.T) {
	env := lifecycleTestEnvironment()
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env))

	err := reconciler.ensureNetworkPolicy(context.Background(), env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if condition := meta.FindStatusCondition(env.Status.Conditions, EnvironmentConditionNetworkPolicyReady); condition != nil {
		t.Fatalf("NetworkPolicyReady condition = %#v, want nil", condition)
	}
}

func TestEnsureNetworkPolicyStepMarksReadyWhenDisabledPolicyCleanupSucceeds(t *testing.T) {
	env := lifecycleTestEnvironment()
	env.Generation = 12
	env.Spec.NetworkPolicies = &environmentsv1.NetworkPolicies{Enabled: false}
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env), lifecycleTestNetworkPolicy(env))
	session := NewEnvironmentStatusSession(env)

	result, err := reconciler.ensureNetworkPolicyStep(context.Background(), session)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want zero", result)
	}
	assertConditionWithObservedGeneration(t, env, EnvironmentConditionNetworkPolicyReady, metav1.ConditionTrue, EnvironmentReasonReconciled, env.Generation)
}

func TestEnsureNetworkPolicyDoesNotMarkNetworkPolicyReadyWhenDisabledPolicyCleanupSucceeds(t *testing.T) {
	env := lifecycleTestEnvironment()
	env.Spec.NetworkPolicies = &environmentsv1.NetworkPolicies{Enabled: false}
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env), lifecycleTestNetworkPolicy(env))

	err := reconciler.ensureNetworkPolicy(context.Background(), env, zap.NewNop())

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if condition := meta.FindStatusCondition(env.Status.Conditions, EnvironmentConditionNetworkPolicyReady); condition != nil {
		t.Fatalf("NetworkPolicyReady condition = %#v, want nil", condition)
	}
}

func TestEnsureNetworkPolicyStepReturnsErrorAndMarksNotReady(t *testing.T) {
	env := lifecycleTestEnvironment()
	env.Generation = 7
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env))
	reconciler.Client = &lifecycleFailingClient{Client: reconciler.Client, failCreate: true}
	session := NewEnvironmentStatusSession(env)

	_, err := reconciler.ensureNetworkPolicyStep(context.Background(), session)

	if err == nil {
		t.Fatalf("error = nil, want create failure")
	}
	assertConditionWithObservedGeneration(t, env, EnvironmentConditionNetworkPolicyReady, metav1.ConditionFalse, EnvironmentReasonError, env.Generation)
}

func TestEnsureNetworkPolicyStepReturnsUpdateErrorAndMarksNotReady(t *testing.T) {
	env := lifecycleTestEnvironment()
	env.Generation = 8
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env), lifecycleTestNetworkPolicy(env))
	reconciler.Client = &lifecycleFailingClient{Client: reconciler.Client, failUpdate: true}
	session := NewEnvironmentStatusSession(env)

	_, err := reconciler.ensureNetworkPolicyStep(context.Background(), session)

	if err == nil {
		t.Fatalf("error = nil, want update failure")
	}
	assertConditionWithObservedGeneration(t, env, EnvironmentConditionNetworkPolicyReady, metav1.ConditionFalse, EnvironmentReasonError, env.Generation)
}

func TestReconcileComposeStepMarksReadyWhenNoCompose(t *testing.T) {
	env := lifecycleTestEnvironment()
	reconciler := lifecycleTestReconciler(t, env)
	session := NewEnvironmentStatusSession(env)

	result, err := reconciler.reconcileComposeStep(context.Background(), session)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want zero", result)
	}
	assertCondition(t, env, EnvironmentConditionComposeReady, metav1.ConditionTrue, EnvironmentReasonReconciled)
}

func TestReconcileComposeStepReturnsErrorAndMarksNotReady(t *testing.T) {
	env := lifecycleTestEnvironment()
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: "services:\n  app:\n    image: nginx"}
	env.Status.ComposeStatus = &environmentsv1.CompositionStatus{
		DeployedResources: &environmentsv1.DeployedResources{Services: []string{"removed"}},
	}
	removedService := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name:      "removed",
		Namespace: env.Spec.TargetNamespace,
		Labels: map[string]string{
			composition.DockerCompositionLabel:    env.Name,
			composition.EnvironmentNamespaceLabel: env.Namespace,
		},
	}}
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env), removedService)
	reconciler.Client = &lifecycleFailingClient{Client: reconciler.Client, failDelete: true}
	session := NewEnvironmentStatusSession(env)

	_, err := reconciler.reconcileComposeStep(context.Background(), session)

	if err == nil {
		t.Fatalf("error = nil, want compose cleanup failure")
	}
	assertCondition(t, env, EnvironmentConditionComposeReady, metav1.ConditionFalse, EnvironmentReasonError)
}

func TestReconcileComposeStepPreservesExistingLifecycleConditions(t *testing.T) {
	env := lifecycleTestEnvironment()
	env.Generation = 9
	env.Spec.Compose = &environmentsv1.CompositionSpec{ComposeContent: "services:\n  app:\n    image: nginx"}
	reconciler := lifecycleTestReconciler(t, env, lifecycleTestNamespace(env))
	session := NewEnvironmentStatusSession(env)
	session.MarkTrue(EnvironmentConditionNamespaceReady, EnvironmentReasonReconciled, "Namespace is ready")
	session.MarkTrue(EnvironmentConditionNetworkPolicyReady, EnvironmentReasonReconciled, "Network policy is ready")

	result, err := reconciler.reconcileComposeStep(context.Background(), session)

	if err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want zero", result)
	}
	assertConditionWithObservedGeneration(t, env, EnvironmentConditionNamespaceReady, metav1.ConditionTrue, EnvironmentReasonReconciled, env.Generation)
	assertConditionWithObservedGeneration(t, env, EnvironmentConditionNetworkPolicyReady, metav1.ConditionTrue, EnvironmentReasonReconciled, env.Generation)
	assertConditionWithObservedGeneration(t, env, EnvironmentConditionComposeReady, metav1.ConditionTrue, EnvironmentReasonReconciled, env.Generation)
}

func lifecycleTestStep(name string, calls *[]string, result ctrl.Result, err error) func(context.Context, *EnvironmentStatusSession) (ctrl.Result, error) {
	return func(context.Context, *EnvironmentStatusSession) (ctrl.Result, error) {
		*calls = append(*calls, name)
		return result, err
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func lifecycleTestEnvironment() *environmentsv1.Environment {
	return &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-env", UID: types.UID("test-env-uid")},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "admin@example.com",
		},
	}
}

func lifecycleTestNamespace(env *environmentsv1.Environment) *corev1.Namespace {
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: env.Spec.TargetNamespace}}
	applyEnvironmentNamespaceOwnership(namespace, env)
	return namespace
}

func lifecycleTestNetworkPolicy(env *environmentsv1.Environment) *networkingv1.NetworkPolicy {
	return &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{
		Name:      networkPolicyName,
		Namespace: env.Spec.TargetNamespace,
	}}
}

func lifecycleTestReconciler(t *testing.T, objs ...client.Object) *EnvironmentReconciler {
	t.Helper()
	scheme := testutil.NewTestScheme()
	return &EnvironmentReconciler{
		Client: testutil.NewFakeClient(scheme, objs...).WithStatusSubresource(&environmentsv1.Environment{}).Build(),
		Scheme: scheme,
		Logger: zap.NewNop(),
	}
}

func assertCondition(t *testing.T, env *environmentsv1.Environment, conditionType string, status metav1.ConditionStatus, reason string) {
	t.Helper()
	assertConditionWithObservedGeneration(t, env, conditionType, status, reason, 0)
}

func assertConditionWithObservedGeneration(t *testing.T, env *environmentsv1.Environment, conditionType string, status metav1.ConditionStatus, reason string, observedGeneration int64) {
	t.Helper()
	condition := meta.FindStatusCondition(env.Status.Conditions, conditionType)
	if condition == nil {
		t.Fatalf("condition %s missing", conditionType)
	}
	if condition.Status != status {
		t.Fatalf("condition %s status = %s, want %s", conditionType, condition.Status, status)
	}
	if condition.Reason != reason {
		t.Fatalf("condition %s reason = %s, want %s", conditionType, condition.Reason, reason)
	}
	if condition.ObservedGeneration != observedGeneration {
		t.Fatalf("condition %s observed generation = %d, want %d", conditionType, condition.ObservedGeneration, observedGeneration)
	}
}
