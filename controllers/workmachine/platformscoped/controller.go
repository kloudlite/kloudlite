package platformscoped

import (
	"context"
	"fmt"
	"time"

	"github.com/codingconcepts/env"
	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	"github.com/kloudlite/kloudlite/pkg/errors"
	"github.com/kloudlite/kloudlite/pkg/operator-toolkit/reconciler"
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type platformLifecycleStep struct {
	Name      string
	Condition string
	OnCreate  func(context.Context, *workmachineshared.StatusSession) (ctrl.Result, error)
	OnDelete  func(context.Context, *workmachineshared.StatusSession) (ctrl.Result, error)
}

// Reconcile handles WorkMachine CR reconciliation
func (r *PlatformScopedReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	obj := &v1.WorkMachine{}
	if err := r.Get(ctx, request.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if obj.GetDeletionTimestamp() == nil && !controllerutil.ContainsFinalizer(obj, reconciler.Finalizer) {
		controllerutil.AddFinalizer(obj, reconciler.Finalizer)
		if err := r.Update(ctx, obj); err != nil {
			return ctrl.Result{}, err
		}
	}

	original := obj.DeepCopy()
	session := workmachineshared.NewStatusSession(obj)
	session.Touch()

	result, err := r.reconcilePlatform(ctx, session)
	workmachineshared.MarkAggregate(session, workmachineshared.ConditionPlatformReady, workmachineshared.PlatformSteps(), workmachineshared.ReasonReconciled, workmachineshared.ReasonPlatformNotReady, "platform-scoped prerequisites are ready", "platform-scoped prerequisites are not ready")
	session.ComputeReady()

	patchResult, patchErr := workmachineshared.PatchStatus(ctx, r.Client, original, session.Object())
	if finalResult, finalErr := reconcileReturn(result, err, patchResult, patchErr); finalErr != nil || !isZeroResult(finalResult) {
		return finalResult, finalErr
	}

	if obj.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(obj, reconciler.Finalizer) {
		if err := r.removeFinalizer(ctx, obj.Name); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PlatformScopedReconciler) removeFinalizer(ctx context.Context, name string) error {
	latest := &v1.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: name}, latest); err != nil {
		return client.IgnoreNotFound(err)
	}
	if !controllerutil.ContainsFinalizer(latest, reconciler.Finalizer) {
		return nil
	}
	base := latest.DeepCopy()
	controllerutil.RemoveFinalizer(latest, reconciler.Finalizer)
	if err := r.Patch(ctx, latest, client.MergeFrom(base)); err != nil {
		return client.IgnoreNotFound(err)
	}
	return nil
}

func (r *PlatformScopedReconciler) lifecycleSteps() []platformLifecycleStep {
	return []platformLifecycleStep{
		{
			Name:      "handle-machine-type-change",
			Condition: workmachineshared.ConditionCloudMachineProvisioned,
			OnCreate:  r.handleMachineTypeChange,
		},
		{
			Name:      "handle-node-reboot-request",
			Condition: workmachineshared.ConditionCloudMachineRunning,
			OnCreate:  r.handleNodeRebootRequest,
		},
		{
			Name:     "cleanup-workmachine-namespace",
			OnDelete: r.cleanupWorkMachineNamespace,
		},
		{
			Name:      "setup-cloud-machine",
			Condition: workmachineshared.ConditionNodeJoined,
			OnCreate:  r.setupCloudMachine,
			OnDelete:  r.cleanupCloudMachine,
		},
		{
			Name:     "ensure-workmachine-manager",
			OnCreate: r.ensureWorkMachineManager,
			OnDelete: r.cleanupWorkMachineManager,
		},
	}
}

func reconcileReturn(reconcileResult ctrl.Result, reconcileErr error, patchResult ctrl.Result, patchErr error) (ctrl.Result, error) {
	if reconcileErr != nil {
		return reconcileResult, reconcileErr
	}
	if patchErr != nil || !isZeroResult(patchResult) {
		return patchResult, patchErr
	}
	if !isZeroResult(reconcileResult) {
		return reconcileResult, nil
	}
	return ctrl.Result{}, nil
}

func (r *PlatformScopedReconciler) reconcilePlatform(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	steps := r.lifecycleSteps()
	if session.Object().GetDeletionTimestamp() != nil {
		for i := len(steps) - 1; i >= 0; i-- {
			if steps[i].OnDelete == nil {
				continue
			}
			result, err := steps[i].OnDelete(ctx, session)
			if err != nil || !isZeroResult(result) {
				return result, err
			}
		}
		return ctrl.Result{}, nil
	}

	for _, step := range steps {
		if step.OnCreate == nil {
			continue
		}
		result, err := step.OnCreate(ctx, session)
		if err != nil || !isZeroResult(result) {
			return result, err
		}
	}
	return ctrl.Result{}, nil
}

func isZeroResult(result ctrl.Result) bool {
	return result == (ctrl.Result{})
}

func markReady(session *workmachineshared.StatusSession, conditionType, message string) (ctrl.Result, error) {
	session.MarkTrue(conditionType, workmachineshared.ReasonReconciled, message)
	return ctrl.Result{}, nil
}

func markBlocked(session *workmachineshared.StatusSession, conditionType, reason, message string, after time.Duration) (ctrl.Result, error) {
	session.MarkFalse(conditionType, reason, message)
	return ctrl.Result{RequeueAfter: after}, nil
}

func markError(session *workmachineshared.StatusSession, conditionType string, err error) (ctrl.Result, error) {
	session.MarkFalse(conditionType, workmachineshared.ReasonError, err.Error())
	return ctrl.Result{}, err
}

func (r *PlatformScopedReconciler) ensureWorkMachineManager(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	labels := map[string]string{
		"app":                           "workmachine-manager",
		"kloudlite.io/workmachine":      obj.Name,
		"kloudlite.io/controller-scope": "machine",
	}
	name := workMachineManagerName(obj.Name)

	serviceAccount := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: r.env.PodNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, serviceAccount, func() error {
		serviceAccount.Labels = labels
		return controllerutil.SetControllerReference(obj, serviceAccount, r.Scheme)
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to ensure workmachine-manager service account: %w", err)
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, func() error {
		clusterRoleBinding.Labels = labels
		clusterRoleBinding.RoleRef = rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "cluster-admin"}
		clusterRoleBinding.Subjects = []rbacv1.Subject{{Kind: "ServiceAccount", Name: name, Namespace: r.env.PodNamespace}}
		return controllerutil.SetControllerReference(obj, clusterRoleBinding, r.Scheme)
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to ensure workmachine-manager cluster role binding: %w", err)
	}

	statefulSet := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: r.env.PodNamespace}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, statefulSet, func() error {
		replicas := int32(1)
		statefulSet.Labels = labels
		statefulSet.Spec.Replicas = &replicas
		statefulSet.Spec.ServiceName = name
		statefulSet.Spec.PodManagementPolicy = appsv1.ParallelPodManagement
		statefulSet.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		statefulSet.Spec.Template.Labels = labels
		statefulSet.Spec.Template.Spec.ServiceAccountName = name
		statefulSet.Spec.Template.Spec.NodeSelector = workmachineshared.WorkMachineAddOnPlacement(obj.Name).NodeSelector
		statefulSet.Spec.Template.Spec.Tolerations = workmachineshared.WorkMachineAddOnPlacement(obj.Name).Tolerations
		statefulSet.Spec.Template.Spec.Containers = []corev1.Container{{
			Name:            "workmachine-manager",
			Image:           r.env.WorkMachineManagerImage,
			ImagePullPolicy: corev1.PullAlways,
			EnvFrom: []corev1.EnvFromSource{
				{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "api-server-config"}}},
				{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "api-server-secret"}}},
			},
			Command: []string{"/app/workmachine-manager"},
			Args:    []string{"server", "workmachine-manager"},
			Env: []corev1.EnvVar{
				{Name: "NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}}},
				{Name: "POD_NAMESPACE", Value: r.env.PodNamespace},
			},
		}}
		return controllerutil.SetControllerReference(obj, statefulSet, r.Scheme)
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to ensure workmachine-manager statefulset: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *PlatformScopedReconciler) cleanupWorkMachineManager(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	name := workMachineManagerName(session.Object().Name)
	if err := r.Delete(ctx, &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: r.env.PodNamespace}}); err != nil && !apiErrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to delete workmachine-manager statefulset %s: %w", name, err)
	}
	if err := r.Delete(ctx, &rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: name}}); err != nil && !apiErrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to delete workmachine-manager cluster role binding %s: %w", name, err)
	}
	if err := r.Delete(ctx, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: r.env.PodNamespace}}); err != nil && !apiErrors.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to delete workmachine-manager service account %s: %w", name, err)
	}
	return ctrl.Result{}, nil
}

func (r *PlatformScopedReconciler) cleanupWorkMachineNamespace(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	if obj.Spec.TargetNamespace == "" {
		return ctrl.Result{}, nil
	}

	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Spec.TargetNamespace}, namespace); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get workmachine namespace %s: %w", obj.Spec.TargetNamespace, err)
	}

	if namespace.DeletionTimestamp == nil {
		if err := r.Delete(ctx, namespace); err != nil && !apiErrors.IsNotFound(err) {
			return ctrl.Result{}, fmt.Errorf("failed to delete workmachine namespace %s: %w", obj.Spec.TargetNamespace, err)
		}
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func workMachineManagerName(workMachineName string) string {
	return fmt.Sprintf("workmachine-manager-%s", workMachineName)
}

func machineScopedControllerName(workMachineName string) string {
	return fmt.Sprintf("workmachine-machine-scoped-%s", workMachineName)
}

func machineScopedWorkloadsReady(obj *v1.WorkMachine) bool {
	condition := meta.FindStatusCondition(obj.Status.Conditions, workmachineshared.ConditionMachineWorkloadsReady)
	return condition != nil && condition.Status == metav1.ConditionTrue && condition.ObservedGeneration == obj.Generation
}

// handleNodeRebootRequest checks if the node associated with this WorkMachine has requested a reboot
// (typically for loading NVIDIA drivers after installation) and reboots the instance if needed
func (r *PlatformScopedReconciler) handleNodeRebootRequest(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	// Only handle reboot requests if machine is created and running
	if obj.Status.MachineID == "" {
		return markReady(session, workmachineshared.ConditionCloudMachineRunning, "cloud machine is not created")
	}

	// Get the node with the same name as the WorkMachine
	var node corev1.Node
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Name}, &node); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Node doesn't exist yet, nothing to do
			return markReady(session, workmachineshared.ConditionCloudMachineRunning, "node is not present")
		}
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to get node: %w", err))
	}

	// Check if node has reboot requested annotation
	rebootRequested, exists := node.Annotations["kloudlite.io/workmachine-reboot-requested"]
	if !exists || rebootRequested != "true" {
		// No reboot requested
		return markReady(session, workmachineshared.ConditionCloudMachineRunning, "no reboot requested")
	}

	ctrl.LoggerFrom(ctx).Info("node reboot requested, rebooting instance", "node", node.Name, "machineID", obj.Status.MachineID)

	// Reboot the instance using cloud provider API
	if err := r.cloudProviderAPI.RebootMachine(ctx, obj.Status.MachineID); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to reboot machine: %w", err))
	}

	// Remove the reboot annotation from the node
	delete(node.Annotations, "kloudlite.io/workmachine-reboot-requested")
	if err := r.Update(ctx, &node); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to remove reboot annotation from node: %w", err))
	}

	ctrl.LoggerFrom(ctx).Info("instance rebooted successfully, waiting for node to rejoin", "node", node.Name, "machineID", obj.Status.MachineID)

	return markReady(session, workmachineshared.ConditionCloudMachineRunning, "node reboot request handled")
}

// SetupWithManager sets up the controller with the Manager
func (r *PlatformScopedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := env.Set(&r.env); err != nil {
		return errors.Wrap("failed to load env vars", err)
	}

	if err := r.initSharedRuntimeState(); err != nil {
		return err
	}

	provider, err := setupCloudProvider(context.Background(), r.env)
	if err != nil {
		return err
	}
	r.cloudProviderAPI = provider

	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.WorkMachine{}).Named("workmachine-platform-scoped")
	builder.Owns(&appsv1.StatefulSet{})
	builder.Owns(&corev1.ServiceAccount{})
	builder.Owns(&rbacv1.ClusterRoleBinding{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor("workmachine-platform-scoped")))

	// Watch for Nodes to trigger reconciliation when node joins/updates
	// The reconciler will fetch fresh IPs from the configured cloud provider when Node Ready state changes.
	builder.Watches(
		&corev1.Node{},
		handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			node, ok := obj.(*corev1.Node)
			if !ok {
				return nil
			}

			// Node name matches WorkMachine name
			// Trigger reconciliation to update WorkMachine status
			return []reconcile.Request{
				{NamespacedName: client.ObjectKey{Name: node.Name}},
			}
		}),
	)

	// Add indexer for pod.spec.nodeName to efficiently query pods by node name
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, "spec.nodeName", func(obj client.Object) []string {
		pod := obj.(*corev1.Pod)
		return []string{pod.Spec.NodeName}
	}); err != nil {
		return errors.Wrap("failed to setup field indexer for pod.spec.nodeName", err)
	}

	return builder.Complete(r)
}
