package shared

import (
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Step struct {
	Name      string
	Condition string
	ShouldRun func(*v1.WorkMachine) bool
}

func PlatformSteps() []Step {
	return []Step{
		{Name: "handle-machine-type-change", Condition: ConditionCloudMachineProvisioned},
		{Name: "handle-node-reboot-request", Condition: ConditionCloudMachineRunning},
		{Name: "setup-cloud-machine", Condition: ConditionNodeJoined},
	}
}

func MachineSteps() []Step {
	return []Step{
		{Name: "setup-namespace", Condition: ConditionMachineNamespaceReady},
		{Name: "ensure-network-policy", Condition: ConditionNetworkPolicyReady},
		{Name: "sync-wildcard-cert-secret", Condition: ConditionWildcardCertSynced},
		{Name: "setup-host-manager-rbac", Condition: ConditionHostManagerRBACReady},
		{Name: "ensure-ssh", Condition: ConditionSSHReady},
		{Name: "ensure-wm-ingress-controller", Condition: ConditionIngressControllerReady, ShouldRun: whenRunning},
		{Name: "ensure-host-manager", Condition: ConditionHostManagerReady, ShouldRun: whenRunning},
		{Name: "ensure-buildkit", Condition: ConditionBuildKitReady, ShouldRun: whenRunning},
		{Name: "ensure-tunnel-server", Condition: ConditionTunnelServerReady, ShouldRun: func(obj *v1.WorkMachine) bool {
			return obj.Spec.State == v1.MachineStateRunning && obj.Status.PublicIP != ""
		}},
		{Name: "ensure-code-analyzer", Condition: ConditionCodeAnalyzerReady, ShouldRun: whenRunning},
		{Name: "check-auto-shutdown", Condition: ConditionAutoShutdownChecked, ShouldRun: func(obj *v1.WorkMachine) bool {
			return obj.Spec.State == v1.MachineStateRunning && obj.Status.State == v1.MachineStateRunning && obj.Spec.AutoShutdown != nil && obj.Spec.AutoShutdown.Enabled
		}},
	}
}

func MarkAggregate(session *StatusSession, aggregateType string, steps []Step, readyReason, blockedReason, readyMessage, blockedMessage string) {
	obj := session.Object()
	for _, step := range steps {
		if step.ShouldRun != nil && !step.ShouldRun(obj) {
			continue
		}

		condition := GetCondition(obj.Status.Conditions, step.Condition)
		if condition == nil || condition.Status != metav1.ConditionTrue || condition.ObservedGeneration != obj.Generation {
			session.MarkFalse(aggregateType, blockedReason, blockedMessage)
			return
		}
	}
	session.MarkTrue(aggregateType, readyReason, readyMessage)
}

func whenRunning(obj *v1.WorkMachine) bool {
	return obj.Spec.State == v1.MachineStateRunning
}
