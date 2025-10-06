package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WorkMachineReconciler reconciles a WorkMachine object
type WorkMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const WorkMachineFinalizerName = "workmachine.machines.kloudlite.io/cleanup"

// Reconcile handles WorkMachine CR reconciliation
func (r *WorkMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("controller", "workmachine", "workmachine", req.Name)
	logger.Info("Reconciling WorkMachine")

	// Fetch the WorkMachine instance
	workMachine := &machinesv1.WorkMachine{}
	if err := r.Get(ctx, req.NamespacedName, workMachine); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("WorkMachine not found, probably deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get WorkMachine")
		return ctrl.Result{}, err
	}

	// Check if WorkMachine is being deleted
	if workMachine.DeletionTimestamp != nil {
		logger.Info("WorkMachine is being deleted, starting cleanup")
		return r.handleWorkMachineDeletion(ctx, workMachine, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(workMachine, WorkMachineFinalizerName) {
		logger.Info("Adding finalizer to WorkMachine")
		controllerutil.AddFinalizer(workMachine, WorkMachineFinalizerName)
		if err := r.Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Ensure target namespace exists
	if err := r.ensureNamespace(ctx, workMachine.Spec.TargetNamespace, logger); err != nil {
		logger.Error(err, "Failed to ensure namespace exists")
		return ctrl.Result{}, err
	}

	// Ensure RBAC resources exist for package-manager
	if err := r.ensurePackageManagerRBAC(ctx, workMachine.Spec.TargetNamespace, logger); err != nil {
		logger.Error(err, "Failed to ensure package-manager RBAC")
		return ctrl.Result{}, err
	}

	// Check if namespace is being deleted (but WorkMachine is not)
	namespace := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: workMachine.Spec.TargetNamespace}, namespace); err == nil {
		if namespace.DeletionTimestamp != nil {
			logger.Info("Namespace is being deleted, but WorkMachine is not - recreating finalizer protection")
			// This shouldn't happen normally because namespace has our finalizer
			// But if it does, we requeue and the deletion will be blocked by the finalizer
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Initialize status if it doesn't exist
	if workMachine.Status.State == "" {
		// First time - set current state to desired state
		workMachine.Status.State = workMachine.Spec.DesiredState
		now := metav1.Now()
		if workMachine.Spec.DesiredState == machinesv1.MachineStateRunning {
			workMachine.Status.StartedAt = &now
		}

		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to initialize WorkMachine status")
			return ctrl.Result{}, err
		}
		logger.Info("Initialized WorkMachine status", "state", workMachine.Status.State)
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	// Check if state transition is needed
	currentState := workMachine.Status.State
	desiredState := workMachine.Spec.DesiredState

	// Ensure package manager deployment exists (regardless of machine state)
	if err := r.ensurePackageManagerDeployment(ctx, workMachine, logger); err != nil {
		logger.Error(err, "Failed to ensure package-manager deployment")
		return ctrl.Result{}, err
	}

	if currentState == desiredState {
		// No transition needed, machine is in desired state
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Handle state transitions
	logger.Info("State transition detected", "current", currentState, "desired", desiredState)

	switch currentState {
	case machinesv1.MachineStateStopped:
		if desiredState == machinesv1.MachineStateRunning {
			// Transition to starting
			workMachine.Status.State = machinesv1.MachineStateStarting
			if err := r.Status().Update(ctx, workMachine); err != nil {
				logger.Error(err, "Failed to update status to starting")
				return ctrl.Result{}, err
			}
			logger.Info("Machine transitioning to starting")
			// Requeue quickly to move to running
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}

	case machinesv1.MachineStateStarting:
		// Transition to running
		workMachine.Status.State = machinesv1.MachineStateRunning
		now := metav1.Now()
		workMachine.Status.StartedAt = &now
		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status to running")
			return ctrl.Result{}, err
		}
		logger.Info("Machine is now running")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil

	case machinesv1.MachineStateRunning:
		if desiredState == machinesv1.MachineStateStopped {
			// Transition to stopping
			workMachine.Status.State = machinesv1.MachineStateStopping
			if err := r.Status().Update(ctx, workMachine); err != nil {
				logger.Error(err, "Failed to update status to stopping")
				return ctrl.Result{}, err
			}
			logger.Info("Machine transitioning to stopping")
			// Requeue quickly to move to stopped
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}

	case machinesv1.MachineStateStopping:
		// Transition to stopped
		workMachine.Status.State = machinesv1.MachineStateStopped
		now := metav1.Now()
		workMachine.Status.StoppedAt = &now
		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status to stopped")
			return ctrl.Result{}, err
		}
		logger.Info("Machine is now stopped")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// handleWorkMachineDeletion handles cleanup when WorkMachine is being deleted
func (r *WorkMachineReconciler) handleWorkMachineDeletion(ctx context.Context, workMachine *machinesv1.WorkMachine, logger logr.Logger) (ctrl.Result, error) {
	namespaceName := workMachine.Spec.TargetNamespace

	// Check for active Workspaces in the target namespace
	workspaceList := &workspacesv1.WorkspaceList{}
	if err := r.List(ctx, workspaceList, client.InNamespace(namespaceName)); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Failed to list Workspaces", "namespace", namespaceName)
			return ctrl.Result{}, err
		}
	}

	// Block deletion if there are active workspaces
	if len(workspaceList.Items) > 0 {
		logger.Info("Deletion blocked: active Workspaces exist in namespace",
			"namespace", namespaceName,
			"workspaceCount", len(workspaceList.Items))

		// Update status with DeletionBlocked condition
		now := metav1.Now()
		workspaceNames := make([]string, len(workspaceList.Items))
		for i, ws := range workspaceList.Items {
			workspaceNames[i] = ws.Name
		}

		message := fmt.Sprintf("Cannot delete WorkMachine: %d active workspace(s) exist: %v", len(workspaceList.Items), workspaceNames)

		// Check if condition already exists
		conditionExists := false
		for i, condition := range workMachine.Status.Conditions {
			if condition.Type == machinesv1.WorkMachineConditionDeletionBlocked {
				workMachine.Status.Conditions[i].Status = metav1.ConditionTrue
				workMachine.Status.Conditions[i].Reason = "ActiveWorkspacesExist"
				workMachine.Status.Conditions[i].Message = message
				workMachine.Status.Conditions[i].LastTransitionTime = &now
				conditionExists = true
				break
			}
		}

		if !conditionExists {
			workMachine.Status.Conditions = append(workMachine.Status.Conditions, machinesv1.WorkMachineCondition{
				Type:               machinesv1.WorkMachineConditionDeletionBlocked,
				Status:             metav1.ConditionTrue,
				LastTransitionTime: &now,
				Reason:             "ActiveWorkspacesExist",
				Message:            message,
			})
		}

		if err := r.Status().Update(ctx, workMachine); err != nil {
			logger.Error(err, "Failed to update status with DeletionBlocked condition")
			return ctrl.Result{}, err
		}

		// Requeue to check again later
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// No workspaces, proceed with namespace deletion
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace)
	if err == nil {
		// Namespace still exists
		if namespace.DeletionTimestamp != nil {
			// Namespace is being deleted - remove our finalizer to allow it to complete
			if controllerutil.ContainsFinalizer(namespace, WorkMachineFinalizerName) {
				logger.Info("Removing finalizer from namespace to allow deletion", "namespace", namespaceName)
				controllerutil.RemoveFinalizer(namespace, WorkMachineFinalizerName)
				if err := r.Update(ctx, namespace); err != nil {
					logger.Error(err, "Failed to remove finalizer from namespace")
					return ctrl.Result{}, err
				}
			}
			logger.Info("Namespace is being deleted, waiting for it to be removed", "namespace", namespaceName)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// Delete the namespace
		logger.Info("Deleting WorkMachine namespace", "namespace", namespaceName)
		if err := r.Delete(ctx, namespace); err != nil && !errors.IsNotFound(err) {
			logger.Error(err, "Failed to delete namespace", "namespace", namespaceName)
			return ctrl.Result{}, err
		}

		// Requeue to wait for namespace deletion to complete
		logger.Info("Namespace deletion initiated, waiting for completion", "namespace", namespaceName)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get namespace", "namespace", namespaceName)
		return ctrl.Result{}, err
	}

	// Namespace is deleted, remove finalizer
	logger.Info("Namespace is deleted, removing finalizer from WorkMachine")
	controllerutil.RemoveFinalizer(workMachine, WorkMachineFinalizerName)
	if err := r.Update(ctx, workMachine); err != nil {
		logger.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	logger.Info("WorkMachine cleanup completed successfully")
	return ctrl.Result{}, nil
}

// ensureNamespace creates the namespace if it doesn't exist and adds finalizer
func (r *WorkMachineReconciler) ensureNamespace(ctx context.Context, namespaceName string, logger logr.Logger) error {
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: namespaceName}, namespace)

	if err == nil {
		// Namespace already exists, ensure it has the finalizer
		if !controllerutil.ContainsFinalizer(namespace, WorkMachineFinalizerName) {
			logger.Info("Adding finalizer to existing namespace", "namespace", namespaceName)
			controllerutil.AddFinalizer(namespace, WorkMachineFinalizerName)
			if err := r.Update(ctx, namespace); err != nil {
				logger.Error(err, "Failed to add finalizer to namespace")
				return err
			}
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		return err
	}

	// Create the namespace with finalizer
	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				"kloudlite.io/managed":     "true",
				"kloudlite.io/workmachine": "true",
			},
			Finalizers: []string{WorkMachineFinalizerName},
		},
	}

	if err := r.Create(ctx, namespace); err != nil {
		return err
	}

	logger.Info("Created namespace for WorkMachine with finalizer", "namespace", namespaceName)
	return nil
}

// ensurePackageManagerRBAC ensures RBAC resources for package-manager exist in the namespace
func (r *WorkMachineReconciler) ensurePackageManagerRBAC(ctx context.Context, namespace string, logger logr.Logger) error {
	// Create ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "package-manager",
			Namespace: namespace,
		},
	}

	if err := r.Get(ctx, client.ObjectKey{Name: sa.Name, Namespace: namespace}, sa); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to check service account: %w", err)
		}

		// Create service account
		if err := r.Create(ctx, sa); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create service account: %w", err)
		}
		logger.Info("Created ServiceAccount for package-manager", "namespace", namespace)
	}

	// Create Role with PackageRequest permissions
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "package-manager",
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
		},
	}

	existingRole := &rbacv1.Role{}
	if err := r.Get(ctx, client.ObjectKey{Name: role.Name, Namespace: namespace}, existingRole); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to check role: %w", err)
		}

		// Create role
		if err := r.Create(ctx, role); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create role: %w", err)
		}
		logger.Info("Created Role for package-manager", "namespace", namespace)
	}

	// Create RoleBinding
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "package-manager",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "package-manager",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "package-manager",
		},
	}

	existingRB := &rbacv1.RoleBinding{}
	if err := r.Get(ctx, client.ObjectKey{Name: rb.Name, Namespace: namespace}, existingRB); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to check role binding: %w", err)
		}

		// Create role binding
		if err := r.Create(ctx, rb); err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create role binding: %w", err)
		}
		logger.Info("Created RoleBinding for package-manager", "namespace", namespace)
	}

	return nil
}

// ensurePackageManagerDeployment ensures the package-manager deployment exists in the WorkMachine's namespace
func (r *WorkMachineReconciler) ensurePackageManagerDeployment(ctx context.Context, workMachine *machinesv1.WorkMachine, logger logr.Logger) error {
	namespace := workMachine.Spec.TargetNamespace
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: "package-manager", Namespace: namespace}, deployment)

	if err == nil {
		// Deployment already exists
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check package-manager deployment: %w", err)
	}

	// Create the Deployment
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	privileged := true
	replicas := int32(1)

	deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "package-manager",
			Namespace: namespace,
			Labels: map[string]string{
				"app":                       "package-manager",
				"kloudlite.io/package-mgmt": "true",
				"kloudlite.io/workmachine":  workMachine.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "package-manager",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "package-manager",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "package-manager",
					Containers: []corev1.Container{
						{
							Name:            "package-manager",
							Image:           "kloudlite/package-manager:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: namespace,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "nix-store",
									MountPath: "/var/lib/kloudlite/nix-store",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "nix-store",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite/nix-store",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := r.Create(ctx, deployment); err != nil {
		return fmt.Errorf("failed to create package-manager deployment: %w", err)
	}

	logger.Info("Created package-manager Deployment", "namespace", namespace)
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&machinesv1.WorkMachine{}).
		Complete(r)
}
