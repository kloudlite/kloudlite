package environment

import (
	"context"
	"fmt"

	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type environmentLifecycleStep struct {
	Name      string
	Condition string
	ShouldRun func(*environmentsv1.Environment) bool
	OnCreate  func(context.Context, *EnvironmentStatusSession) (ctrl.Result, error)
	OnSkip    func(context.Context, *EnvironmentStatusSession) (ctrl.Result, error)
	OnDelete  func(context.Context, *EnvironmentStatusSession) (ctrl.Result, error)
}

func (r *EnvironmentReconciler) lifecycleSteps() []environmentLifecycleStep {
	return []environmentLifecycleStep{
		{
			Name:      "ensure-target-namespace",
			Condition: EnvironmentConditionNamespaceReady,
			OnCreate:  r.ensureTargetNamespaceStep,
			OnDelete:  r.deleteTargetNamespaceStep,
		},
		{
			Name:      "ensure-network-policy",
			Condition: EnvironmentConditionNetworkPolicyReady,
			OnCreate:  r.ensureNetworkPolicyStep,
		},
		{
			Name:      "reconcile-compose",
			Condition: EnvironmentConditionComposeReady,
			OnCreate:  r.reconcileComposeStep,
			OnDelete:  r.cleanupComposeStep,
		},
		{
			Name:      "reconcile-activation",
			Condition: EnvironmentConditionActivationReady,
			OnCreate:  r.reconcileActivationStep,
			OnDelete:  r.cleanupActivationStep,
		},
	}
}

func reconcileEnvironmentReturn(reconcileResult ctrl.Result, reconcileErr error, patchResult ctrl.Result, patchErr error) (ctrl.Result, error) {
	if reconcileErr != nil {
		return reconcileResult, reconcileErr
	}
	if patchErr != nil || !isZeroEnvironmentResult(patchResult) {
		return patchResult, patchErr
	}
	if !isZeroEnvironmentResult(reconcileResult) {
		return reconcileResult, nil
	}
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) reconcileEnvironment(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	return reconcileEnvironmentLifecycle(ctx, session, r.lifecycleSteps())
}

func reconcileEnvironmentLifecycle(ctx context.Context, session *EnvironmentStatusSession, steps []environmentLifecycleStep) (ctrl.Result, error) {
	if session.Object().GetDeletionTimestamp() != nil {
		return reconcileEnvironmentDeletion(ctx, session, steps)
	}

	for _, step := range steps {
		if step.ShouldRun != nil && !step.ShouldRun(session.Object()) {
			if step.OnSkip == nil {
				continue
			}
			result, err := step.OnSkip(ctx, session)
			if err != nil || !isZeroEnvironmentResult(result) {
				return result, err
			}
			continue
		}

		if step.OnCreate == nil {
			continue
		}
		result, err := step.OnCreate(ctx, session)
		if err != nil || !isZeroEnvironmentResult(result) {
			return result, err
		}
	}
	return ctrl.Result{}, nil
}

func reconcileEnvironmentDeletion(ctx context.Context, session *EnvironmentStatusSession, steps []environmentLifecycleStep) (ctrl.Result, error) {
	obj := session.Object()
	if obj.Status.State != environmentsv1.EnvironmentStateDeleting {
		obj.Status.State = environmentsv1.EnvironmentStateDeleting
		obj.Status.Message = "Deleting environment and cleaning up resources"
	}

	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].OnDelete == nil {
			continue
		}
		result, err := steps[i].OnDelete(ctx, session)
		if err != nil {
			session.MarkFalse(EnvironmentConditionCleanupComplete, EnvironmentReasonError, err.Error())
			return result, err
		}
		if !isZeroEnvironmentResult(result) {
			session.MarkFalse(EnvironmentConditionCleanupComplete, EnvironmentReasonWaiting, fmt.Sprintf("environment cleanup is waiting for %s", steps[i].Name))
			return result, nil
		}
	}

	session.MarkTrue(EnvironmentConditionCleanupComplete, EnvironmentReasonReconciled, "Environment cleanup is complete")
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) ensureTargetNamespaceStep(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	environment := session.Object()
	logger := r.environmentLogger(environment)

	setHashAndSubdomain(environment, logger)

	created, err := r.ensureNamespaceExistsForLifecycle(ctx, environment, logger)
	if err != nil {
		session.MarkFalse(EnvironmentConditionNamespaceReady, EnvironmentReasonError, err.Error())
		return ctrl.Result{}, err
	}

	session.MarkTrue(EnvironmentConditionNamespaceReady, EnvironmentReasonReconciled, "Namespace is ready")
	if created || environment.Status.State == "" {
		environment.Status.State = environmentsv1.EnvironmentStateInactive
		environment.Status.Message = "Namespace created successfully"
	}
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) ensureNamespaceExistsForLifecycle(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (bool, error) {
	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace); err == nil {
		return false, r.ensureNamespaceExistsForLifecycleUpdate(ctx, environment, logger)
	} else if !apierrors.IsNotFound(err) {
		logger.Error("failed to check existing namespace", zap.Error(err))
		return false, err
	}

	if environment.Status.State != environmentsv1.EnvironmentStateInactive {
		setEnvironmentStatus(environment, environmentsv1.EnvironmentStateInactive, "Creating namespace for environment")
	}
	return true, r.createNamespace(ctx, environment, logger)
}

func (r *EnvironmentReconciler) ensureNamespaceExistsForLifecycleUpdate(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	exists, err := r.ensureNamespaceExists(ctx, environment, logger)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("namespace %s was not ready after ensure", environment.Spec.TargetNamespace)
	}
	return nil
}

func (r *EnvironmentReconciler) deleteTargetNamespaceStep(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	environment := session.Object()
	deleted, err := r.deleteNamespace(ctx, environment, r.environmentLogger(environment))
	if err != nil {
		return ctrl.Result{RequeueAfter: r.Cfg.Environment.DeletionRetryInterval}, err
	}
	if !deleted {
		return ctrl.Result{RequeueAfter: r.Cfg.Environment.PodTerminationRetryInterval}, nil
	}
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) ensureNetworkPolicyStep(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	environment := session.Object()
	logger := r.environmentLogger(environment)
	if err := r.ensureNetworkPolicy(ctx, environment, logger); err != nil {
		logger.Error("failed to ensure network policy", zap.Error(err))
		session.MarkFalse(EnvironmentConditionNetworkPolicyReady, EnvironmentReasonError, err.Error())
		return ctrl.Result{}, err
	}
	session.MarkTrue(EnvironmentConditionNetworkPolicyReady, EnvironmentReasonReconciled, "Network policy is ready")
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) reconcileComposeStep(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	environment := session.Object()
	logger := r.environmentLogger(environment)
	reconciled, err := r.reconcileCompose(ctx, environment, logger)
	if err != nil {
		logger.Error("failed to reconcile compose", zap.Error(err))
		session.MarkFalse(EnvironmentConditionComposeReady, EnvironmentReasonError, err.Error())
		return ctrl.Result{}, err
	}
	if !reconciled {
		session.MarkTrue(EnvironmentConditionComposeReady, EnvironmentReasonReconciled, "No compose spec configured")
	} else {
		session.MarkTrue(EnvironmentConditionComposeReady, EnvironmentReasonReconciled, "Compose is ready")
	}
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) cleanupComposeStep(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	environment := session.Object()
	logger := r.environmentLogger(environment)
	if err := r.cleanupComposeResources(ctx, environment, logger); err != nil {
		logger.Error("failed to cleanup compose resources, will retry", zap.Error(err))
		return ctrl.Result{RequeueAfter: r.Cfg.Environment.DeletionRetryInterval}, fmt.Errorf("failed to cleanup compose resources: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) reconcileActivationStep(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	result, err := r.reconcileActivationState(ctx, session.Object(), r.environmentLogger(session.Object()))
	if err != nil {
		session.MarkFalse(EnvironmentConditionActivationReady, EnvironmentReasonError, err.Error())
		return result, err
	}
	if !isZeroEnvironmentResult(result) {
		session.MarkFalse(EnvironmentConditionActivationReady, EnvironmentReasonActivationPending, "Activation state reconciliation is pending")
		return result, nil
	}
	session.MarkTrue(EnvironmentConditionActivationReady, EnvironmentReasonReconciled, "Activation state reconciled")
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) cleanupActivationStep(ctx context.Context, session *EnvironmentStatusSession) (ctrl.Result, error) {
	environment := session.Object()
	logger := r.environmentLogger(environment)
	if err := r.cleanupWorkspaceConnections(ctx, environment, logger); err != nil {
		logger.Error("failed to cleanup workspace connections, will retry", zap.Error(err))
		return ctrl.Result{RequeueAfter: r.Cfg.Environment.DeletionRetryInterval}, fmt.Errorf("failed to cleanup workspace connections: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *EnvironmentReconciler) environmentLogger(environment *environmentsv1.Environment) *zap.Logger {
	return r.Logger.With(
		zap.String("environment", environment.Name),
		zap.String("namespace", environment.Namespace),
	)
}

func isZeroEnvironmentResult(result ctrl.Result) bool {
	return result == (ctrl.Result{})
}

func cleanupComplete(environment *environmentsv1.Environment) bool {
	for _, condition := range environment.Status.Conditions {
		if condition.Type == EnvironmentConditionCleanupComplete {
			return condition.Status == metav1.ConditionTrue
		}
	}
	return false
}
