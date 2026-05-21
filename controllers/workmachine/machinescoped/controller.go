package machinescoped

import (
	"context"
	"fmt"
	"time"

	"github.com/codingconcepts/env"
	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	"github.com/kloudlite/kloudlite/pkg/errors"
	fn "github.com/kloudlite/kloudlite/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/pkg/operator-toolkit/reconciler"
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type machineLifecycleStep struct {
	Name      string
	Condition string
	ShouldRun func(*v1.WorkMachine) bool
	OnCreate  func(context.Context, *workmachineshared.StatusSession) (ctrl.Result, error)
	OnSkip    func(context.Context, *workmachineshared.StatusSession) (ctrl.Result, error)
	OnDelete  func(context.Context, *workmachineshared.StatusSession) (ctrl.Result, error)
}

// Reconcile handles machine-scoped WorkMachine resources.
func (r *MachineScopedReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	obj := &v1.WorkMachine{}
	if err := r.Get(ctx, request.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !r.shouldReconcileWorkMachine(ctx, obj) {
		return reconcile.Result{}, nil
	}

	original := obj.DeepCopy()
	session := workmachineshared.NewStatusSession(obj)
	session.Touch()

	result, err := r.reconcileMachine(ctx, session)
	workmachineshared.MarkAggregate(session, workmachineshared.ConditionMachineWorkloadsReady, workmachineshared.MachineSteps(), workmachineshared.ReasonReconciled, workmachineshared.ReasonMachineWorkloadsNotReady, "machine-scoped workloads are ready", "machine-scoped workloads are not ready")
	session.ComputeReady()

	patchResult, patchErr := workmachineshared.PatchStatus(ctx, r.Client, original, session.Object())
	if finalResult, finalErr := reconcileMachineReturn(result, err, patchResult, patchErr); finalErr != nil || !isZeroMachineResult(finalResult) {
		return finalResult, finalErr
	}

	return ctrl.Result{}, nil
}

func (r *MachineScopedReconciler) shouldReconcileWorkMachine(ctx context.Context, obj *v1.WorkMachine) bool {
	if r.env.WorkMachineName != "" {
		return obj.Name == r.env.WorkMachineName
	}
	if r.env.NodeName == "" {
		return true
	}

	node := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: r.env.NodeName}, node); err != nil {
		return false
	}
	return node.Labels["kloudlite.io/workmachine"] == obj.Name
}

func (r *MachineScopedReconciler) lifecycleSteps() []machineLifecycleStep {
	return []machineLifecycleStep{
		{
			Name:      "setup-namespace",
			Condition: workmachineshared.ConditionMachineNamespaceReady,
			OnCreate:  r.createNamespace,
			OnDelete:  r.deleteNamespace,
		},
		{
			Name:      "ensure-network-policy",
			Condition: workmachineshared.ConditionNetworkPolicyReady,
			OnCreate:  r.ensureNetworkPolicy,
			OnDelete:  r.cleanupNetworkPolicy,
		},
		{
			Name:      "sync-wildcard-cert-secret",
			Condition: workmachineshared.ConditionWildcardCertSynced,
			OnCreate:  r.syncWildcardCertSecret,
		},
		{
			Name:      "setup-host-manager-rbac",
			Condition: workmachineshared.ConditionHostManagerRBACReady,
			OnCreate:  r.createHostManagerRBAC,
		},
		{
			Name:      "ensure-ssh",
			Condition: workmachineshared.ConditionSSHReady,
			OnCreate:  r.ensureSSH,
		},
		{
			Name:      "ensure-wm-ingress-controller",
			Condition: workmachineshared.ConditionIngressControllerReady,
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateRunning
			},
			OnCreate: r.ensureWorkmachineIngressController,
			OnSkip:   r.cleanupWorkmachineIngressController,
			OnDelete: r.cleanupWorkmachineIngressController,
		},
		{
			Name:      "mark-integrated-host-manager",
			Condition: workmachineshared.ConditionHostManagerReady,
			OnCreate:  r.ensureIntegratedHostManager,
			OnDelete:  r.cleanupHostManagerPod,
		},
		{
			Name:      "ensure-buildkit",
			Condition: workmachineshared.ConditionBuildKitReady,
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateRunning
			},
			OnCreate: r.ensureBuildKit,
			OnSkip:   r.cleanupBuildKit,
			OnDelete: r.cleanupBuildKit,
		},
		{
			Name:      "ensure-tunnel-server",
			Condition: workmachineshared.ConditionTunnelServerReady,
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateRunning && obj.Status.PublicIP != ""
			},
			OnCreate: r.ensureTunnelServer,
			OnSkip:   r.cleanupTunnelServer,
			OnDelete: r.cleanupTunnelServer,
		},
		{
			Name:      "ensure-code-analyzer",
			Condition: workmachineshared.ConditionCodeAnalyzerReady,
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateRunning
			},
			OnCreate: r.ensureCodeAnalyzer,
			OnSkip:   r.cleanupCodeAnalyzer,
			OnDelete: r.cleanupCodeAnalyzer,
		},
		{
			Name:      "check-auto-shutdown",
			Condition: workmachineshared.ConditionAutoShutdownChecked,
			ShouldRun: func(obj *v1.WorkMachine) bool {
				return obj.Spec.State == v1.MachineStateRunning &&
					obj.Status.State == v1.MachineStateRunning &&
					obj.Spec.AutoShutdown != nil &&
					obj.Spec.AutoShutdown.Enabled
			},
			OnCreate: r.checkAutoShutdown,
		},
	}
}

func reconcileMachineReturn(reconcileResult ctrl.Result, reconcileErr error, patchResult ctrl.Result, patchErr error) (ctrl.Result, error) {
	if reconcileErr != nil {
		return reconcileResult, reconcileErr
	}
	if patchErr != nil || !isZeroMachineResult(patchResult) {
		return patchResult, patchErr
	}
	if !isZeroMachineResult(reconcileResult) {
		return reconcileResult, nil
	}
	return ctrl.Result{}, nil
}

func (r *MachineScopedReconciler) reconcileMachine(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	steps := r.lifecycleSteps()
	if session.Object().GetDeletionTimestamp() != nil {
		return reconcileMachineDeletion(ctx, session, steps)
	}

	for _, step := range steps {
		if step.ShouldRun != nil && !step.ShouldRun(session.Object()) {
			if step.OnSkip == nil {
				continue
			}
			result, err := step.OnSkip(ctx, session)
			if err != nil || !isZeroMachineResult(result) {
				return result, err
			}
			continue
		}

		result, err := step.OnCreate(ctx, session)
		if err != nil || !isZeroMachineResult(result) {
			return result, err
		}
	}
	return ctrl.Result{}, nil
}

func reconcileMachineDeletion(ctx context.Context, session *workmachineshared.StatusSession, steps []machineLifecycleStep) (ctrl.Result, error) {
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].OnDelete == nil {
			continue
		}
		result, err := steps[i].OnDelete(ctx, session)
		if err != nil {
			session.MarkFalse(workmachineshared.ConditionMachineScopedCleanupComplete, workmachineshared.ReasonError, err.Error())
			return result, err
		}
		if !isZeroMachineResult(result) {
			session.MarkFalse(workmachineshared.ConditionMachineScopedCleanupComplete, workmachineshared.ReasonWaiting, fmt.Sprintf("machine-scoped cleanup is waiting for %s", steps[i].Name))
			return result, nil
		}
	}

	session.MarkTrue(workmachineshared.ConditionMachineScopedCleanupComplete, workmachineshared.ReasonReconciled, "machine-scoped cleanup is complete")
	return ctrl.Result{}, nil
}

func isZeroMachineResult(result ctrl.Result) bool {
	return result == (ctrl.Result{})
}

func markMachineReady(session *workmachineshared.StatusSession, conditionType, message string) (ctrl.Result, error) {
	session.MarkTrue(conditionType, workmachineshared.ReasonReconciled, message)
	return ctrl.Result{}, nil
}

func markMachineFailed(session *workmachineshared.StatusSession, conditionType string, err error) (ctrl.Result, error) {
	session.MarkFalse(conditionType, workmachineshared.ReasonError, err.Error())
	return ctrl.Result{}, err
}

func markMachineError(session *workmachineshared.StatusSession, conditionType string, err error) (ctrl.Result, error) {
	session.MarkFalse(conditionType, workmachineshared.ReasonError, err.Error())
	return ctrl.Result{}, err
}

func markMachineWaiting(session *workmachineshared.StatusSession, conditionType, message string, after time.Duration) (ctrl.Result, error) {
	if after <= 0 {
		after = 2 * time.Second
	}
	session.MarkFalse(conditionType, workmachineshared.ReasonWaiting, message)
	return ctrl.Result{RequeueAfter: after}, nil
}

func (r *MachineScopedReconciler) ensureWorkmachineIngressController(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()

	serviceLabels := map[string]string{
		"app":                      "wm-ingress-controller",
		"kloudlite.io/workmachine": obj.Name,
	}
	managerSelector := map[string]string{
		"app":                      "workmachine-manager",
		"kloudlite.io/workmachine": obj.Name,
	}

	legacyStatefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "wm-ingress-controller", Namespace: obj.Spec.TargetNamespace}}
	if err := r.Delete(ctx, legacyStatefulSet); err != nil && !apiErrors.IsNotFound(err) {
		return markMachineFailed(session, workmachineshared.ConditionIngressControllerReady, fmt.Errorf("failed to delete legacy wm-ingress-controller statefulset: %w", err))
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wm-ingress-controller",
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		if !fn.IsOwner(service, obj) {
			service.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		service.Labels = serviceLabels

		service.Spec = corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: managerSelector,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(443),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		}

		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionIngressControllerReady, fmt.Errorf("failed to create/update wm-ingress-controller service: %w", err))
	}

	return markMachineReady(session, workmachineshared.ConditionIngressControllerReady, "workmachine ingress controller is ready")
}

func (r *MachineScopedReconciler) cleanupWorkmachineIngressController(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	deploymentName := "wm-ingress-controller"
	clusterRoleName := fmt.Sprintf("wm-ingress-controller-%s", obj.Name)
	clusterRoleBindingName := fmt.Sprintf("wm-ingress-controller-%s", obj.Name)

	// Delete StatefulSet
	if err := r.Delete(ctx, &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: obj.Spec.TargetNamespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionIngressControllerReady, fmt.Errorf("failed to delete wm-ingress-controller statefulset: %w", err))
		}
	}

	// Delete Service
	if err := r.Delete(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wm-ingress-controller",
			Namespace: obj.Spec.TargetNamespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionIngressControllerReady, fmt.Errorf("failed to delete wm-ingress-controller service: %w", err))
		}
	}

	// Delete ClusterRoleBinding
	if err := r.Delete(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionIngressControllerReady, fmt.Errorf("failed to delete wm-ingress-controller cluster role binding: %w", err))
		}
	}

	// Delete ClusterRole
	if err := r.Delete(ctx, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionIngressControllerReady, fmt.Errorf("failed to delete wm-ingress-controller cluster role: %w", err))
		}
	}

	// Delete ServiceAccount
	if err := r.Delete(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wm-ingress-controller",
			Namespace: obj.Spec.TargetNamespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionIngressControllerReady, fmt.Errorf("failed to delete wm-ingress-controller service account: %w", err))
		}
	}

	return markMachineReady(session, workmachineshared.ConditionIngressControllerReady, "workmachine ingress controller is cleaned up")
}

// SetupWithManager sets up the machine-scoped controller with the Manager.
func (r *MachineScopedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := env.Set(&r.env); err != nil {
		return errors.Wrap("failed to load env vars", err)
	}

	if err := r.initSharedRuntimeState(); err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.WorkMachine{}).Named("workmachine-manager")
	builder.Owns(&corev1.Namespace{})
	builder.Owns(&appsv1.StatefulSet{})
	builder.Owns(&appsv1.Deployment{})
	builder.Owns(&corev1.ServiceAccount{})
	builder.Owns(&rbacv1.ClusterRole{})
	builder.Owns(&rbacv1.ClusterRoleBinding{})
	builder.Owns(&networkingv1.NetworkPolicy{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor("workmachine-manager")))

	// Watch for workspaces and trigger machine-scoped reconciliation of their owning WorkMachine.
	builder.Watches(
		&workspacev1.Workspace{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			workspace, ok := obj.(*workspacev1.Workspace)
			if !ok {
				return nil
			}

			if workspace.Spec.WorkmachineName == "" {
				return nil
			}

			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: workspace.Spec.WorkmachineName}},
			}
		}),
	)

	// Watch for machine-scoped Pods to recreate them if they crash.
	builder.Watches(
		&corev1.Pod{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				return nil
			}

			// Get WorkMachine name from pod label (host-manager pods are labeled with workmachine name)
			workmachineName, exists := pod.Labels["kloudlite.io/workmachine"]
			if !exists {
				return nil
			}

			// Trigger reconciliation to check and recreate pod if needed
			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: workmachineName}},
			}
		}),
	)

	return builder.Complete(r)
}
