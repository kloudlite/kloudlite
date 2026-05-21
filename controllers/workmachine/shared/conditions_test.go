package shared

import (
	"context"
	"testing"

	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestStatusSessionPreservesOtherOwnerConditions(t *testing.T) {
	wm := &v1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}
	wm.Generation = 7
	wm.Status.Conditions = []metav1.Condition{{
		Type:               ConditionMachineNamespaceReady,
		Status:             metav1.ConditionTrue,
		Reason:             ReasonReconciled,
		Message:            "namespace ready",
		ObservedGeneration: 7,
	}}

	session := NewStatusSession(wm)
	session.MarkTrue(ConditionCloudMachineProvisioned, ReasonReconciled, "cloud machine exists")

	if got := GetCondition(session.Object().Status.Conditions, ConditionMachineNamespaceReady); got == nil || got.Status != metav1.ConditionTrue {
		t.Fatalf("machine-scoped condition was not preserved: %#v", session.Object().Status.Conditions)
	}
	if got := GetCondition(session.Object().Status.Conditions, ConditionCloudMachineProvisioned); got == nil || got.ObservedGeneration != 7 {
		t.Fatalf("platform condition missing or wrong generation: %#v", got)
	}
}

func TestReadyFalseWhenPlatformReadyMissing(t *testing.T) {
	wm := &v1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}
	wm.Generation = 3
	session := NewStatusSession(wm)

	session.ComputeReady()

	ready := GetCondition(session.Object().Status.Conditions, ConditionReady)
	if ready == nil || ready.Status != metav1.ConditionFalse || ready.Reason != ReasonPlatformNotReady {
		t.Fatalf("Ready = %#v, want False/%s", ready, ReasonPlatformNotReady)
	}
}

func TestStepDefinitionsMapStepsToConditions(t *testing.T) {
	platformSteps := PlatformSteps()
	assertStep(t, platformSteps, 0, "handle-machine-type-change", ConditionCloudMachineProvisioned)
	assertStep(t, platformSteps, 1, "handle-node-reboot-request", ConditionCloudMachineRunning)
	assertStep(t, platformSteps, 2, "setup-cloud-machine", ConditionNodeJoined)
	if len(platformSteps) != 3 {
		t.Fatalf("platform steps = %d, want 3; machine-scoped controller bootstrap must not gate platform readiness", len(platformSteps))
	}

	machineSteps := MachineSteps()
	assertStep(t, machineSteps, 0, "setup-namespace", ConditionMachineNamespaceReady)
	assertStep(t, machineSteps, 1, "ensure-network-policy", ConditionNetworkPolicyReady)
	assertStep(t, machineSteps, 2, "sync-wildcard-cert-secret", ConditionWildcardCertSynced)
	assertStep(t, machineSteps, 3, "setup-host-manager-rbac", ConditionHostManagerRBACReady)
	assertStep(t, machineSteps, 4, "ensure-ssh", ConditionSSHReady)
	assertStep(t, machineSteps, 5, "ensure-wm-ingress-controller", ConditionIngressControllerReady)
	assertStep(t, machineSteps, 6, "ensure-host-manager", ConditionHostManagerReady)
	assertStep(t, machineSteps, 7, "ensure-buildkit", ConditionBuildKitReady)
	assertStep(t, machineSteps, 8, "ensure-tunnel-server", ConditionTunnelServerReady)
	assertStep(t, machineSteps, 9, "ensure-code-analyzer", ConditionCodeAnalyzerReady)
	assertStep(t, machineSteps, 10, "check-auto-shutdown", ConditionAutoShutdownChecked)
}

func TestPlatformReadyRequiresRunningNodeOnly(t *testing.T) {
	wm := &v1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}
	wm.Generation = 5
	session := NewStatusSession(wm)
	session.MarkTrue(ConditionCloudMachineProvisioned, ReasonReconciled, "cloud machine exists")

	MarkAggregate(session, ConditionPlatformReady, PlatformSteps(), ReasonReconciled, ReasonPlatformNotReady, "platform ready", "platform blocked")

	platformReady := GetCondition(session.Object().Status.Conditions, ConditionPlatformReady)
	if platformReady == nil || platformReady.Status != metav1.ConditionFalse || platformReady.Reason != ReasonPlatformNotReady {
		t.Fatalf("PlatformReady = %#v, want False/%s when running/node/controller conditions are missing", platformReady, ReasonPlatformNotReady)
	}

	session.MarkTrue(ConditionCloudMachineRunning, ReasonReconciled, "cloud machine running")
	session.MarkTrue(ConditionNodeJoined, ReasonReconciled, "node joined")

	MarkAggregate(session, ConditionPlatformReady, PlatformSteps(), ReasonReconciled, ReasonPlatformNotReady, "platform ready", "platform blocked")

	platformReady = GetCondition(session.Object().Status.Conditions, ConditionPlatformReady)
	if platformReady == nil || platformReady.Status != metav1.ConditionTrue || platformReady.Reason != ReasonReconciled {
		t.Fatalf("PlatformReady = %#v, want True/%s after all platform conditions are true", platformReady, ReasonReconciled)
	}
}

func TestMarkAggregateRequiresRunnableStepsCurrentGeneration(t *testing.T) {
	wm := &v1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}
	wm.Generation = 4
	session := NewStatusSession(wm)
	session.MarkTrue(ConditionCloudMachineProvisioned, ReasonReconciled, "cloud machine exists")
	wm.Status.Conditions = append(wm.Status.Conditions, metav1.Condition{
		Type:               ConditionMachineNamespaceReady,
		Status:             metav1.ConditionTrue,
		Reason:             ReasonReconciled,
		Message:            "stale namespace ready",
		ObservedGeneration: 3,
	})

	MarkAggregate(session, ConditionPlatformReady, []Step{
		{Name: "cloud", Condition: ConditionCloudMachineProvisioned},
		{Name: "namespace", Condition: ConditionMachineNamespaceReady},
		{Name: "skipped", Condition: ConditionNetworkPolicyReady, ShouldRun: func(*v1.WorkMachine) bool { return false }},
	}, ReasonReconciled, ReasonPlatformNotReady, "platform ready", "platform blocked")

	platformReady := GetCondition(session.Object().Status.Conditions, ConditionPlatformReady)
	if platformReady == nil || platformReady.Status != metav1.ConditionFalse || platformReady.Reason != ReasonPlatformNotReady {
		t.Fatalf("PlatformReady = %#v, want False/%s", platformReady, ReasonPlatformNotReady)
	}

	session.MarkTrue(ConditionMachineNamespaceReady, ReasonReconciled, "namespace ready")
	MarkAggregate(session, ConditionPlatformReady, []Step{
		{Name: "cloud", Condition: ConditionCloudMachineProvisioned},
		{Name: "namespace", Condition: ConditionMachineNamespaceReady},
		{Name: "skipped", Condition: ConditionNetworkPolicyReady, ShouldRun: func(*v1.WorkMachine) bool { return false }},
	}, ReasonReconciled, ReasonPlatformNotReady, "platform ready", "platform blocked")

	platformReady = GetCondition(session.Object().Status.Conditions, ConditionPlatformReady)
	if platformReady == nil || platformReady.Status != metav1.ConditionTrue || platformReady.Reason != ReasonReconciled {
		t.Fatalf("PlatformReady = %#v, want True/%s", platformReady, ReasonReconciled)
	}
}

func TestPatchStatusConflictReturnsRequeueWithoutError(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	wm := &v1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(wm).WithStatusSubresource(&v1.WorkMachine{}).Build()

	latest := &v1.WorkMachine{}
	if err := c.Get(context.Background(), client.ObjectKey{Name: "wm"}, latest); err != nil {
		t.Fatal(err)
	}
	original := latest.DeepCopy()
	latest.Status.ObservedGeneration = 1

	conflictClient := &statusConflictClient{Client: c}
	result, err := PatchStatus(context.Background(), conflictClient, original, latest)
	if err != nil {
		t.Fatalf("PatchStatus returned error %v, want nil", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatalf("PatchStatus conflict result = %#v, want short requeue", result)
	}
}

type statusConflictClient struct{ client.Client }

func (c *statusConflictClient) Status() client.StatusWriter {
	return conflictStatusWriter{}
}

type conflictStatusWriter struct{}

func (conflictStatusWriter) Create(context.Context, client.Object, client.Object, ...client.SubResourceCreateOption) error {
	return apiErrors.NewConflict(schema.GroupResource{Group: "machines.kloudlite.io", Resource: "workmachines"}, "wm", nil)
}
func (conflictStatusWriter) Update(context.Context, client.Object, ...client.SubResourceUpdateOption) error {
	return apiErrors.NewConflict(schema.GroupResource{Group: "machines.kloudlite.io", Resource: "workmachines"}, "wm", nil)
}
func (conflictStatusWriter) Patch(context.Context, client.Object, client.Patch, ...client.SubResourcePatchOption) error {
	return apiErrors.NewConflict(schema.GroupResource{Group: "machines.kloudlite.io", Resource: "workmachines"}, "wm", nil)
}

func assertStep(t *testing.T, steps []Step, index int, name string, condition string) {
	t.Helper()
	if len(steps) <= index {
		t.Fatalf("steps length = %d, want at least %d", len(steps), index+1)
	}
	if steps[index].Name != name || steps[index].Condition != condition {
		t.Fatalf("step %d = {%q, %q}, want {%q, %q}", index, steps[index].Name, steps[index].Condition, name, condition)
	}
}
