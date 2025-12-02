package composition

import (
	"context"
	"reflect"
	"time"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// updateStatusWithRetry updates the status of the Composition with retry logic for conflicts
func (r *CompositionReconciler) updateStatusWithRetry(ctx context.Context, composition *compositionsv1.Composition, environment *compositionsv1.Environment, state compositionsv1.CompositionState, message string, logger *zap.Logger) (reconcile.Result, error) {
	backoff := wait.Backoff{
		Steps:    5,
		Duration: 10 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
	}

	var lastErr error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		if err := r.Status().Update(ctx, composition); err != nil {
			lastErr = err
			if apierrors.IsConflict(err) {
				// Resource version conflict - fetch the latest version and retry
				logger.Debug("Status update conflict, refetching and retrying", zap.Error(err))
				if err := r.Get(ctx, types.NamespacedName{Name: composition.Name, Namespace: composition.Namespace}, composition); err != nil {
					logger.Error("Failed to refetch composition after conflict", zap.Error(err))
					return false, err
				}

				// Re-apply the status changes on the latest version
				composition.Status.State = state
				composition.Status.Message = message
				composition.Status.ObservedGeneration = composition.Generation

				// Update environment activation state
				if environment != nil {
					composition.Status.EnvironmentActivated = environment.Spec.Activated
				}

				now := metav1.Now()
				if state == compositionsv1.CompositionStateRunning {
					composition.Status.LastDeployedTime = &now
				}

				// Create ready condition
				readyCondition := metav1.Condition{
					Type:               readyConditionType,
					Status:             metav1.ConditionFalse,
					ObservedGeneration: composition.Generation,
					LastTransitionTime: now,
					Reason:             string(state),
					Message:            message,
				}

				// Set condition status to true when running
				if state == compositionsv1.CompositionStateRunning {
					readyCondition.Status = metav1.ConditionTrue
				}

				// Update or add condition
				r.setCondition(composition, readyCondition)

				// Retry the update
				return false, nil
			}
			// For non-conflict errors, don't retry
			return false, err
		}
		// Success
		return true, nil
	})

	if err != nil {
		if err == wait.ErrWaitTimeout {
			logger.Error("Failed to update status after maximum retries", zap.Error(lastErr))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, lastErr
		}
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Updated Composition status",
		zap.String("state", string(state)),
		zap.String("message", message))

	// Requeue to continue monitoring for deploying/failed/degraded states
	switch state {
	case compositionsv1.CompositionStateDeploying:
		return reconcile.Result{RequeueAfter: time.Duration(deployingRequeueInterval) * time.Nanosecond}, nil
	case compositionsv1.CompositionStateFailed, compositionsv1.CompositionStateDegraded:
		// Requeue to re-check health in case pods recover or status changes
		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

// updateStatus updates the status of the Composition (legacy method for compatibility)
func (r *CompositionReconciler) updateStatus(ctx context.Context, composition *compositionsv1.Composition, environment *compositionsv1.Environment, state compositionsv1.CompositionState, message string, logger *zap.Logger) (reconcile.Result, error) {
	// Track if any status field actually changed
	needsUpdate := false
	oldStatus := composition.Status.DeepCopy()

	if composition.Status.State != state {
		composition.Status.State = state
		needsUpdate = true
	}

	if composition.Status.Message != message {
		composition.Status.Message = message
		needsUpdate = true
	}

	if composition.Status.ObservedGeneration != composition.Generation {
		composition.Status.ObservedGeneration = composition.Generation
		needsUpdate = true
	}

	// Update environment activation state
	if environment != nil {
		if composition.Status.EnvironmentActivated != environment.Spec.Activated {
			composition.Status.EnvironmentActivated = environment.Spec.Activated
			needsUpdate = true
		}
	}

	now := metav1.Now()
	if state == compositionsv1.CompositionStateRunning {
		if composition.Status.LastDeployedTime == nil || !composition.Status.LastDeployedTime.Equal(&now) {
			composition.Status.LastDeployedTime = &now
			needsUpdate = true
		}
	}

	// Create ready condition
	readyCondition := metav1.Condition{
		Type:               readyConditionType,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: composition.Generation,
		LastTransitionTime: now,
		Reason:             string(state),
		Message:            message,
	}

	// Set condition status to true when running
	if state == compositionsv1.CompositionStateRunning {
		readyCondition.Status = metav1.ConditionTrue
	}

	// Update or add condition (returns true if changed)
	if r.setCondition(composition, readyCondition) {
		needsUpdate = true
	}

	// Only update status if something actually changed
	if !needsUpdate && reflect.DeepEqual(oldStatus, &composition.Status) {
		logger.Debug("Composition status unchanged, skipping status update")
		return reconcile.Result{}, nil
	}

	return r.updateStatusWithRetry(ctx, composition, environment, state, message, logger)
}

// setCondition updates or adds a condition to the composition status
// Returns true if the condition was added or changed
func (r *CompositionReconciler) setCondition(composition *compositionsv1.Composition, condition metav1.Condition) bool {
	for i, existingCondition := range composition.Status.Conditions {
		if existingCondition.Type == condition.Type {
			// Only update if something actually changed
			if existingCondition.Status != condition.Status ||
				existingCondition.Reason != condition.Reason ||
				existingCondition.Message != condition.Message {
				composition.Status.Conditions[i] = condition
				return true
			}
			return false
		}
	}
	// Condition not found, add it
	composition.Status.Conditions = append(composition.Status.Conditions, condition)
	return true
}
