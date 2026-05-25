package environment

import (
	"context"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EnvironmentConditionReady                = "Ready"
	EnvironmentConditionNamespaceReady       = "NamespaceReady"
	EnvironmentConditionNetworkPolicyReady   = "NetworkPolicyReady"
	EnvironmentConditionComposeReady         = "ComposeReady"
	EnvironmentConditionActivationReady      = "ActivationReady"
	EnvironmentConditionSnapshotRestoreReady = "SnapshotRestoreReady"
	EnvironmentConditionCleanupComplete      = "CleanupComplete"

	EnvironmentReasonReconciled        = "Reconciled"
	EnvironmentReasonBlocked           = "Blocked"
	EnvironmentReasonError             = "Error"
	EnvironmentReasonWaiting           = "Waiting"
	EnvironmentReasonNamespaceUnsafe   = "NamespaceUnsafe"
	EnvironmentReasonComposeFailed     = "ComposeFailed"
	EnvironmentReasonActivationPending = "ActivationPending"
)

type EnvironmentStatusSession struct {
	obj *environmentsv1.Environment
}

func NewEnvironmentStatusSession(obj *environmentsv1.Environment) *EnvironmentStatusSession {
	return &EnvironmentStatusSession{obj: obj}
}

func (s *EnvironmentStatusSession) Object() *environmentsv1.Environment { return s.obj }

func (s *EnvironmentStatusSession) MarkTrue(conditionType, reason, message string) {
	s.set(conditionType, metav1.ConditionTrue, reason, message)
}

func (s *EnvironmentStatusSession) MarkFalse(conditionType, reason, message string) {
	s.set(conditionType, metav1.ConditionFalse, reason, message)
}

func (s *EnvironmentStatusSession) MarkUnknown(conditionType, reason, message string) {
	s.set(conditionType, metav1.ConditionUnknown, reason, message)
}

func (s *EnvironmentStatusSession) set(conditionType string, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&s.obj.Status.Conditions, metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: s.obj.Generation,
	})
}

func (s *EnvironmentStatusSession) Touch() {
	s.obj.Status.ObservedGeneration = s.obj.Generation
}

func (s *EnvironmentStatusSession) ComputeReady() {
	required := []string{
		EnvironmentConditionNamespaceReady,
		EnvironmentConditionNetworkPolicyReady,
		EnvironmentConditionActivationReady,
	}
	if s.obj.Spec.Compose != nil && s.obj.Spec.Compose.ComposeContent != "" {
		required = append(required, EnvironmentConditionComposeReady)
	}
	for _, conditionType := range required {
		condition := meta.FindStatusCondition(s.obj.Status.Conditions, conditionType)
		if condition == nil || condition.Status != metav1.ConditionTrue || condition.ObservedGeneration != s.obj.Generation {
			s.MarkFalse(EnvironmentConditionReady, EnvironmentReasonBlocked, "environment prerequisites are not ready")
			return
		}
	}
	s.MarkTrue(EnvironmentConditionReady, EnvironmentReasonReconciled, "Environment is ready")
}

func PatchEnvironmentStatus(ctx context.Context, c client.Client, original *environmentsv1.Environment, updated *environmentsv1.Environment) (ctrl.Result, error) {
	if equality.Semantic.DeepEqual(original.Status, updated.Status) {
		return ctrl.Result{}, nil
	}
	if err := c.Status().Patch(ctx, updated, client.MergeFrom(original)); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		if apierrors.IsConflict(err) {
			return ctrl.Result{RequeueAfter: 500 * time.Millisecond}, nil
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
