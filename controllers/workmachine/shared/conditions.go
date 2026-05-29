package shared

import (
	"context"
	"time"

	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConditionReady                            = "Ready"
	ConditionPlatformReady                    = "PlatformReady"
	ConditionCloudMachineProvisioned          = "CloudMachineProvisioned"
	ConditionCloudMachineRunning              = "CloudMachineRunning"
	ConditionNodeJoined                       = "NodeJoined"
	ConditionMachineScopedControllerAvailable = "MachineScopedControllerAvailable"

	ConditionMachineNamespaceReady  = "MachineNamespaceReady"
	ConditionNetworkPolicyReady     = "NetworkPolicyReady"
	ConditionWildcardCertSynced     = "WildcardCertSynced"
	ConditionHostManagerRBACReady   = "HostManagerRBACReady"
	ConditionSSHReady               = "SSHReady"
	ConditionIngressControllerReady = "IngressControllerReady"
	ConditionHostManagerReady       = "HostManagerReady"
	ConditionBuildKitReady          = "BuildKitReady"
	ConditionTunnelServerReady      = "TunnelServerReady"
	ConditionCodeAnalyzerReady      = "CodeAnalyzerReady"
	ConditionAutoShutdownChecked    = "AutoShutdownChecked"
	ConditionMachineWorkloadsReady  = "MachineWorkloadsReady"

	ConditionDeleting                     = "Deleting"
	ConditionMachineScopedCleanupComplete = "MachineScopedCleanupComplete"
	ConditionCloudMachineDeleted          = "CloudMachineDeleted"
)

const (
	ReasonReconciled                         = "Reconciled"
	ReasonWaitingForCloudMachine             = "WaitingForCloudMachine"
	ReasonWaitingForNodeJoin                 = "WaitingForNodeJoin"
	ReasonWaiting                            = "Waiting"
	ReasonMachineScopedControllerUnavailable = "MachineScopedControllerUnavailable"
	ReasonPlatformNotReady                   = "PlatformNotReady"
	ReasonMachineWorkloadsNotReady           = "MachineWorkloadsNotReady"
	ReasonDeleting                           = "Deleting"
	ReasonError                              = "Error"
)

type StatusSession struct {
	obj *v1.WorkMachine
}

func NewStatusSession(obj *v1.WorkMachine) *StatusSession {
	return &StatusSession{obj: obj}
}

func (s *StatusSession) Object() *v1.WorkMachine { return s.obj }

func (s *StatusSession) MarkTrue(conditionType, reason, message string) {
	s.set(conditionType, metav1.ConditionTrue, reason, message)
}

func (s *StatusSession) MarkFalse(conditionType, reason, message string) {
	s.set(conditionType, metav1.ConditionFalse, reason, message)
}

func (s *StatusSession) MarkUnknown(conditionType, reason, message string) {
	s.set(conditionType, metav1.ConditionUnknown, reason, message)
}

func (s *StatusSession) set(conditionType string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&s.obj.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: s.obj.Generation,
	})
}

func (s *StatusSession) ComputeReady() {
	platform := meta.FindStatusCondition(s.obj.Status.Conditions, ConditionPlatformReady)
	if platform == nil || platform.Status != metav1.ConditionTrue || platform.ObservedGeneration != s.obj.Generation {
		s.MarkFalse(ConditionReady, ReasonPlatformNotReady, "platform-scoped prerequisites are not ready")
		return
	}

	workloads := meta.FindStatusCondition(s.obj.Status.Conditions, ConditionMachineWorkloadsReady)
	if workloads == nil || workloads.Status != metav1.ConditionTrue || workloads.ObservedGeneration != s.obj.Generation {
		s.MarkFalse(ConditionReady, ReasonMachineWorkloadsNotReady, "machine-scoped workloads are not ready")
		return
	}

	s.MarkTrue(ConditionReady, ReasonReconciled, "WorkMachine is ready")
}

func (s *StatusSession) Touch() {
	now := metav1.Now()
	s.obj.Status.LastReconcileTime = &now
	s.obj.Status.ObservedGeneration = s.obj.Generation
}

func GetCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	return meta.FindStatusCondition(conditions, conditionType)
}

func PatchStatus(ctx context.Context, c client.Client, original *v1.WorkMachine, updated *v1.WorkMachine) (ctrl.Result, error) {
	if equality.Semantic.DeepEqual(original.Status, updated.Status) {
		return ctrl.Result{}, nil
	}

	if err := c.Status().Patch(ctx, updated, client.MergeFrom(original)); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		if apiErrors.IsConflict(err) {
			return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
