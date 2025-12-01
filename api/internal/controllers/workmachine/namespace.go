package workmachine

import (
	"fmt"

	environmentV1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// createNamespace creates or updates the target namespace for the WorkMachine
func (r *WorkMachineReconciler) createNamespace(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: obj.Spec.TargetNamespace}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, ns, func() error {
		ns.SetLabels(fn.MapMerge(ns.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		if !fn.IsOwner(ns, obj) {
			// Check if namespace already has other owners
			if len(ns.GetOwnerReferences()) > 0 {
				return fmt.Errorf("namespace %s already owned by another resource", ns.Name)
			}
			ns.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		if !controllerutil.ContainsFinalizer(ns, WorkMachineFinalizerName) {
			ns.SetFinalizers(append(ns.GetFinalizers(), WorkMachineFinalizerName))
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

const (
	wildcardCertSecretName      = "kloudlite-wildcard-cert-tls"
	wildcardCertSourceNamespace = "kloudlite-ingress"
)

// syncWildcardCertSecret copies the wildcard TLS certificate secret from kloudlite-ingress
// to the workmachine namespace for use by tunnel-server, wm-ingress-controller, and workspace pods
func (r *WorkMachineReconciler) syncWildcardCertSecret(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	targetNamespace := obj.Spec.TargetNamespace

	// Get the source secret from kloudlite-ingress namespace
	sourceSecret := &corev1.Secret{}
	if err := r.Get(check.Context(), client.ObjectKey{
		Name:      wildcardCertSecretName,
		Namespace: wildcardCertSourceNamespace,
	}, sourceSecret); err != nil {
		if apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("wildcard certificate secret %s not found in namespace %s", wildcardCertSecretName, wildcardCertSourceNamespace))
		}
		return check.Errored(err)
	}

	// Create or update the secret in the target namespace
	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wildcardCertSecretName,
			Namespace: targetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, targetSecret, func() error {
		targetSecret.Type = sourceSecret.Type
		targetSecret.Data = sourceSecret.Data

		targetSecret.SetLabels(fn.MapMerge(targetSecret.GetLabels(), map[string]string{
			"kloudlite.io/managed":       "true",
			"kloudlite.io/synced-from":   wildcardCertSourceNamespace,
			"kloudlite.io/source-secret": wildcardCertSecretName,
		}))

		if !fn.IsOwner(targetSecret, obj) {
			targetSecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to sync wildcard certificate secret: %w", err))
	}

	return check.Passed()
}

// deleteNamespace handles namespace deletion when WorkMachine is being deleted
func (r *WorkMachineReconciler) deleteNamespace(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespaceName := obj.Spec.TargetNamespace

	// Check for active Workspaces in the target namespace
	var envList environmentV1.EnvironmentList
	if err := r.List(check.Context(), &envList); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}
	}

	// Delete workspace pods directly (bypass finalizers for faster cleanup)
	// When WorkMachine is being deleted, we don't need graceful workspace cleanup
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(check.Context(), workspaceList); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}
	}

	// Filter workspaces owned by this WorkMachine
	var ownedWorkspaces []workspacev1.Workspace
	for _, ws := range workspaceList.Items {
		if ws.Spec.WorkmachineName == obj.Name {
			ownedWorkspaces = append(ownedWorkspaces, ws)
		}
	}

	// Directly delete workspace pods to speed up cleanup
	for _, ws := range ownedWorkspaces {
		// Delete the workspace pod directly
		podName := fmt.Sprintf("workspace-%s", ws.Name)
		pod := &corev1.Pod{}
		err := r.Get(check.Context(), client.ObjectKey{Name: podName, Namespace: namespaceName}, pod)
		if err == nil {
			// Pod exists, delete it
			if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete workspace pod %s: %w", podName, err))
			}
		} else if !apiErrors.IsNotFound(err) {
			return check.Errored(err)
		}

		// Remove finalizer from workspace to allow it to be deleted immediately
		if ws.DeletionTimestamp == nil {
			// Workspace not being deleted yet, delete it
			if err := r.Delete(check.Context(), &ws); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete workspace %s: %w", ws.Name, err))
			}
		} else {
			// Workspace is being deleted but stuck on finalizer, remove it
			if controllerutil.ContainsFinalizer(&ws, "workspaces.kloudlite.io/finalizer") {
				controllerutil.RemoveFinalizer(&ws, "workspaces.kloudlite.io/finalizer")
				if err := r.Update(check.Context(), &ws); err != nil && !apiErrors.IsNotFound(err) {
					return check.Failed(fmt.Errorf("failed to remove finalizer from workspace %s: %w", ws.Name, err))
				}
			}
		}
	}

	// Delete environment namespaces directly (bypass finalizers for faster cleanup)
	for _, env := range envList.Items {
		if env.Spec.WorkMachineName == obj.Name {
			// Delete the environment namespace directly if it exists
			if env.Spec.TargetNamespace != "" {
				envNs := &corev1.Namespace{}
				err := r.Get(check.Context(), client.ObjectKey{Name: env.Spec.TargetNamespace}, envNs)
				if err == nil {
					// Namespace exists, delete it
					if err := r.Delete(check.Context(), envNs); err != nil && !apiErrors.IsNotFound(err) {
						return check.Failed(fmt.Errorf("failed to delete environment namespace %s: %w", env.Spec.TargetNamespace, err))
					}
				} else if !apiErrors.IsNotFound(err) {
					return check.Errored(err)
				}
			}

			// Remove finalizer from environment to allow it to be deleted immediately
			if env.DeletionTimestamp == nil {
				// Environment not being deleted yet, delete it
				if err := r.Delete(check.Context(), &env); err != nil && !apiErrors.IsNotFound(err) {
					return check.Failed(fmt.Errorf("failed to delete environment %s: %w", env.Name, err))
				}
			} else {
				// Environment is being deleted but stuck on finalizer, remove it
				if controllerutil.ContainsFinalizer(&env, "environments.kloudlite.io/finalizer") {
					controllerutil.RemoveFinalizer(&env, "environments.kloudlite.io/finalizer")
					if err := r.Update(check.Context(), &env); err != nil && !apiErrors.IsNotFound(err) {
						return check.Failed(fmt.Errorf("failed to remove finalizer from environment %s: %w", env.Name, err))
					}
				}
			}
		}
	}

	// Delete host-manager pod and service in finalizer
	// Cluster-scoped WorkMachine cannot own namespaced resources via owner references
	hostManagerName := fmt.Sprintf("hm-%s", obj.Name)

	// Delete pod
	pod := &corev1.Pod{}
	if err := r.Get(check.Context(), client.ObjectKey{
		Name:      hostManagerName,
		Namespace: hostManagerNamespace,
	}, pod); err == nil {
		if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete host manager pod: %w", err))
		}
	} else if !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Delete service
	service := &corev1.Service{}
	if err := r.Get(check.Context(), client.ObjectKey{
		Name:      hostManagerName,
		Namespace: hostManagerNamespace,
	}, service); err == nil {
		if err := r.Delete(check.Context(), service); err != nil && !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete host manager service: %w", err))
		}
	} else if !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Proceed with namespace deletion
	namespace := &corev1.Namespace{}
	err := r.Get(check.Context(), client.ObjectKey{Name: namespaceName}, namespace)
	if err == nil {
		// Namespace still exists
		if namespace.DeletionTimestamp != nil {
			// Namespace is being deleted - remove our finalizer to allow it to complete
			if controllerutil.RemoveFinalizer(namespace, WorkMachineFinalizerName) {
				if err := r.Update(check.Context(), namespace); err != nil {
					return check.Failed(err)
				}
			}
			return check.UpdateMsg("Namespace is being deleted, waiting for completion")
		}

		// Delete the namespace
		if err := r.Delete(check.Context(), namespace); err != nil && !apiErrors.IsNotFound(err) {
			return check.Failed(err)
		}

		return check.UpdateMsg("Namespace deletion initiated, waiting for completion")
	}

	if !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// Namespace is deleted
	return check.Passed()
}
