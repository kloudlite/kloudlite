package environment

import (
	"context"
	"errors"
	"testing"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestEnvironmentStatusSessionSetsObservedGeneration(t *testing.T) {
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Generation: 7}}
	session := NewEnvironmentStatusSession(env)

	session.MarkTrue(EnvironmentConditionNamespaceReady, EnvironmentReasonReconciled, "namespace ready")

	condition := meta.FindStatusCondition(env.Status.Conditions, EnvironmentConditionNamespaceReady)
	if condition == nil {
		t.Fatalf("condition not found")
	}
	if condition.ObservedGeneration != 7 {
		t.Fatalf("ObservedGeneration = %d, want 7", condition.ObservedGeneration)
	}
}

func TestEnvironmentStatusSessionComputeReady(t *testing.T) {
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Generation: 3}}
	session := NewEnvironmentStatusSession(env)
	session.MarkTrue(EnvironmentConditionNamespaceReady, EnvironmentReasonReconciled, "namespace ready")
	session.MarkTrue(EnvironmentConditionNetworkPolicyReady, EnvironmentReasonReconciled, "network policy ready")
	session.MarkTrue(EnvironmentConditionActivationReady, EnvironmentReasonReconciled, "activation ready")
	session.MarkTrue(EnvironmentConditionComposeReady, EnvironmentReasonReconciled, "compose ready")
	session.ComputeReady()

	ready := meta.FindStatusCondition(env.Status.Conditions, EnvironmentConditionReady)
	if ready == nil || ready.Status != metav1.ConditionTrue {
		t.Fatalf("Ready condition = %#v, want true", ready)
	}
}

func TestEnvironmentStatusSessionComputeReadyMissingPrerequisite(t *testing.T) {
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Generation: 3}}
	session := NewEnvironmentStatusSession(env)
	session.MarkTrue(EnvironmentConditionNamespaceReady, EnvironmentReasonReconciled, "namespace ready")
	session.MarkTrue(EnvironmentConditionActivationReady, EnvironmentReasonReconciled, "activation ready")

	session.ComputeReady()

	assertReadyFalse(t, env)
}

func TestEnvironmentStatusSessionComputeReadyStalePrerequisite(t *testing.T) {
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Generation: 3}}
	session := NewEnvironmentStatusSession(env)
	session.MarkTrue(EnvironmentConditionNamespaceReady, EnvironmentReasonReconciled, "namespace ready")
	session.MarkTrue(EnvironmentConditionNetworkPolicyReady, EnvironmentReasonReconciled, "network policy ready")
	session.MarkTrue(EnvironmentConditionActivationReady, EnvironmentReasonReconciled, "activation ready")
	env.Status.Conditions[1].ObservedGeneration = 2

	session.ComputeReady()

	assertReadyFalse(t, env)
}

func TestEnvironmentStatusSessionComputeReadyRequiresComposeWhenEnabled(t *testing.T) {
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dev", Generation: 3},
		Spec:       environmentsv1.EnvironmentSpec{Compose: &environmentsv1.CompositionSpec{ComposeContent: "services:\n  app:\n    image: nginx"}},
	}
	session := NewEnvironmentStatusSession(env)
	markBaseReadyPrerequisites(session)

	session.ComputeReady()

	assertReadyFalse(t, env)
	session.MarkTrue(EnvironmentConditionComposeReady, EnvironmentReasonReconciled, "compose ready")
	session.ComputeReady()

	assertReadyTrue(t, env)
}

func TestPatchEnvironmentStatusReturnsRequeueOnConflict(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := environmentsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Namespace: "wm-alice"}}
	original := env.DeepCopy()
	env.Status.State = environmentsv1.EnvironmentStateActive
	conflict := apierrors.NewConflict(schema.GroupResource{Group: environmentsv1.SchemeGroupVersion.Group, Resource: "environments"}, "dev", errors.New("stale object"))
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(original).WithStatusSubresource(&environmentsv1.Environment{}).WithInterceptorFuncs(interceptor.Funcs{
		SubResourcePatch: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
			return conflict
		},
	}).Build()

	result, err := PatchEnvironmentStatus(context.Background(), c, original, env)
	if err != nil {
		t.Fatalf("PatchEnvironmentStatus error = %v, want nil conflict requeue", err)
	}
	if result.RequeueAfter != 500*time.Millisecond {
		t.Fatalf("RequeueAfter = %s, want 500ms", result.RequeueAfter)
	}
}

func markBaseReadyPrerequisites(session *EnvironmentStatusSession) {
	session.MarkTrue(EnvironmentConditionNamespaceReady, EnvironmentReasonReconciled, "namespace ready")
	session.MarkTrue(EnvironmentConditionNetworkPolicyReady, EnvironmentReasonReconciled, "network policy ready")
	session.MarkTrue(EnvironmentConditionActivationReady, EnvironmentReasonReconciled, "activation ready")
}

func assertReadyFalse(t *testing.T, env *environmentsv1.Environment) {
	t.Helper()
	ready := meta.FindStatusCondition(env.Status.Conditions, EnvironmentConditionReady)
	if ready == nil || ready.Status != metav1.ConditionFalse || ready.Reason != EnvironmentReasonBlocked {
		t.Fatalf("Ready condition = %#v, want false/%s", ready, EnvironmentReasonBlocked)
	}
}

func assertReadyTrue(t *testing.T, env *environmentsv1.Environment) {
	t.Helper()
	ready := meta.FindStatusCondition(env.Status.Conditions, EnvironmentConditionReady)
	if ready == nil || ready.Status != metav1.ConditionTrue || ready.Reason != EnvironmentReasonReconciled {
		t.Fatalf("Ready condition = %#v, want true/%s", ready, EnvironmentReasonReconciled)
	}
}
