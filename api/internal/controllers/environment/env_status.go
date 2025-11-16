package environment

import (
	"context"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// updateEnvironmentStatus safely updates environment status with retry logic
func (r *EnvironmentReconciler) updateEnvironmentStatus(ctx context.Context, environment *environmentsv1.Environment, state environmentsv1.EnvironmentState, message string, logger *zap.Logger) error {
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.State = state
		environment.Status.Message = message

		now := metav1.Now()
		if state == environmentsv1.EnvironmentStateActive {
			environment.Status.LastActivatedTime = &now
		} else if state == environmentsv1.EnvironmentStateInactive {
			environment.Status.LastDeactivatedTime = &now
		}

		return nil
	}, logger)
}

// addOrUpdateCondition adds or updates a condition in the environment status
func (r *EnvironmentReconciler) addOrUpdateCondition(environment *environmentsv1.Environment, conditionType environmentsv1.EnvironmentConditionType, status metav1.ConditionStatus, reason, message string) {
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	now := metav1.Now()
	newCondition := environmentsv1.EnvironmentCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: &now,
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or add new one
	found := false
	for i, condition := range environment.Status.Conditions {
		if condition.Type == conditionType {
			environment.Status.Conditions[i] = newCondition
			found = true
			break
		}
	}
	if !found {
		environment.Status.Conditions = append(environment.Status.Conditions, newCondition)
	}
}

// updateCloningStatus updates the cloning status phase and message
func (r *EnvironmentReconciler) updateCloningStatus(ctx context.Context, environment *environmentsv1.Environment, phase environmentsv1.CloningPhase, message string, logger *zap.Logger) {
	if environment.Status.CloningStatus == nil {
		now := metav1.Now()
		environment.Status.CloningStatus = &environmentsv1.CloningStatus{
			StartTime: &now,
		}
	}

	environment.Status.CloningStatus.Phase = phase
	environment.Status.CloningStatus.Message = message

	// Set completion time if phase is completed or failed
	if phase == environmentsv1.CloningPhaseCompleted || phase == environmentsv1.CloningPhaseFailed {
		now := metav1.Now()
		environment.Status.CloningStatus.CompletionTime = &now
	}

	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
	}
}
