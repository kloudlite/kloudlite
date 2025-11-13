package workmachine

import (
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/templates"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	errors "github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ensurePackageManagerDeploymentStep ensures the workmachine-host-manager pod exists
// Pod will be recreated by the controller if it crashes
func (r *WorkMachineReconciler) ensurePackageManagerDeploymentStep(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := hostManagerNamespace
	// Use unique name per WorkMachine since all host managers share the same namespace
	hostManagerName := fmt.Sprintf("hm-%s", obj.Name)

	// Check WorkMachine state - only create pod when running
	if obj.Spec.State != v1.MachineStateRunning {
		// Delete pod if it exists
		pod := &corev1.Pod{}
		err := r.Get(check.Context(), client.ObjectKey{Name: hostManagerName, Namespace: hostManagerNamespace}, pod)
		if err == nil {
			// Pod exists, delete it
			if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete pod for non-running machine: %w", err))
			}
			check.Logger().Info("Deleted host-manager pod because WorkMachine is not running", "state", obj.Spec.State)
		}
		// Also delete service
		svc := &corev1.Service{}
		err = r.Get(check.Context(), client.ObjectKey{Name: hostManagerName, Namespace: hostManagerNamespace}, svc)
		if err == nil {
			if err := r.Delete(check.Context(), svc); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete service for non-running machine: %w", err))
			}
			check.Logger().Info("Deleted host-manager service because WorkMachine is not running", "state", obj.Spec.State)
		}
		return check.Passed()
	}

	pod := &corev1.Pod{}
	err := r.Get(check.Context(), client.ObjectKey{Name: hostManagerName, Namespace: hostManagerNamespace}, pod)

	// Handle errors (except NotFound which we'll handle below)
	if err != nil && !apiErrors.IsNotFound(err) {
		return check.Errored(err)
	}

	// If pod exists (no error), check its status
	if err == nil {
		// Check if pod is in a failed/completed state and needs recreation
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
			// Delete the failed/completed pod so it will be recreated
			if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete failed/completed pod: %w", err))
			}
			// Pod deleted, will be recreated in next reconcile
			return check.UpdateMsg("Recreating failed pod").RequeueAfter(2 * time.Second)
		}

		// Check if pod is ready (all containers ready)
		if pod.Status.Phase != corev1.PodRunning {
			return check.UpdateMsg(fmt.Sprintf("Waiting for host-manager pod to be running (current: %s)", pod.Status.Phase)).RequeueAfter(5 * time.Second)
		}

		// Check all containers are ready
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !containerStatus.Ready {
				return check.UpdateMsg(fmt.Sprintf("Waiting for container %s to be ready", containerStatus.Name)).RequeueAfter(5 * time.Second)
			}
		}

		// Pod is running and all containers are ready
		return check.Passed()
	}

	// Pod doesn't exist, create it
	pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: hostManagerNamespace,
		},
	}

	// Render pod from template
	b, err := templates.WorkMachineHostManagerPod.Render(
		templates.WorkspaceHostManagerValues{
			Namespace:        namespace,
			WorkMachineName:  obj.Name,
			TargetNamespace:  obj.Spec.TargetNamespace,
			SSHUsername:      SSHUserName,
			HostManagerImage: r.env.HostManagerImage,
		},
	)
	if err != nil {
		return check.Failed(errors.Wrap("failed to render workmachine host manager pod template", err))
	}

	if err := yaml.Unmarshal(b, &pod); err != nil {
		return check.Failed(errors.Wrap("failed to unmarshal into pod", err))
	}

	// Set labels
	pod.SetLabels(fn.MapMerge(pod.GetLabels(), map[string]string{
		"app":                       hostManagerName,
		"kloudlite.io/package-mgmt": "true",
		"kloudlite.io/workmachine":  obj.Name,
	}))

	if err := r.Create(check.Context(), pod); err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return check.Passed()
		}
		return check.Failed(err)
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: hostManagerNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, svc, func() error {
		svc.SetLabels(fn.MapMerge(svc.GetLabels(), map[string]string{
			"app":                       hostManagerName,
			"kloudlite.io/package-mgmt": "true",
			"kloudlite.io/workmachine":  obj.Name,
		}))

		svc.Spec.Selector = map[string]string{
			"app": hostManagerName,
		}

		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "ssh",
				Protocol:   corev1.ProtocolTCP,
				Port:       22,
				TargetPort: intstr.FromInt32(2222),
			},
			{
				Name:       "metrics",
				Protocol:   corev1.ProtocolTCP,
				Port:       8081,
				TargetPort: intstr.FromInt32(8081),
			},
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}
