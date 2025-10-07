package helmchart

import (
	"context"
	"fmt"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	helmChartFinalizer = "helmcharts.environments.kloudlite.io/finalizer"
	helmJobImage       = "alpine/helm:latest" // Helm 3 image
)

// Reconciler reconciles HelmChart objects
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// Reconcile handles HelmChart events
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx).WithValues("helmchart", req.NamespacedName)

	zapLogger := r.Logger.With(
		zap.String("helmchart", req.Name),
		zap.String("namespace", req.Namespace),
	)

	zapLogger.Info("Reconciling HelmChart")

	// Fetch the HelmChart instance
	helmChart := &environmentsv1.HelmChart{}
	err := r.Get(ctx, req.NamespacedName, helmChart)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("HelmChart not found, likely deleted")
			return reconcile.Result{}, nil
		}
		zapLogger.Error("Failed to get HelmChart", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if helm chart is being deleted
	if helmChart.DeletionTimestamp != nil {
		zapLogger.Info("HelmChart is being deleted, starting cleanup")
		return r.handleDeletion(ctx, helmChart, zapLogger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(helmChart, helmChartFinalizer) {
		zapLogger.Info("Adding finalizer to HelmChart")
		controllerutil.AddFinalizer(helmChart, helmChartFinalizer)
		if err := r.Update(ctx, helmChart); err != nil {
			zapLogger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Set release name if not already set
	releaseName := helmChart.Status.ReleaseName
	if releaseName == "" {
		releaseName = fmt.Sprintf("%s-%s", helmChart.Name, helmChart.Namespace)
		helmChart.Status.ReleaseName = releaseName
	}

	// Check if installation job already exists and is running
	existingJob, err := r.getHelmJob(ctx, helmChart, "install")
	if err == nil && existingJob != nil {
		// Job exists, check its status
		return r.checkJobStatus(ctx, helmChart, existingJob, zapLogger)
	}

	// Install or upgrade the helm chart
	if err := r.installOrUpgradeHelmChart(ctx, helmChart, zapLogger); err != nil {
		zapLogger.Error("Failed to install/upgrade helm chart", zap.Error(err))
		return r.updateStatus(ctx, helmChart, environmentsv1.HelmChartStateFailed, err.Error(), zapLogger)
	}

	// Requeue to check job status
	return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
}

// installOrUpgradeHelmChart creates a Kubernetes Job to install or upgrade the helm chart
func (r *Reconciler) installOrUpgradeHelmChart(ctx context.Context, helmChart *environmentsv1.HelmChart, logger *zap.Logger) error {
	logger.Info("Installing or upgrading HelmChart")

	// Check if this is an upgrade (chart already installed)
	isUpgrade := helmChart.Status.State == environmentsv1.HelmChartStateInstalled

	// Create values ConfigMap
	valuesConfigMap, err := r.createValuesConfigMap(ctx, helmChart, logger)
	if err != nil {
		return fmt.Errorf("failed to create values ConfigMap: %w", err)
	}

	// Create Helm job
	job := r.createHelmJob(helmChart, valuesConfigMap.Name, isUpgrade)

	// Set controller ownership
	if err := controllerutil.SetControllerReference(helmChart, job, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	// Create the job
	if err := r.Create(ctx, job); err != nil {
		if apierrors.IsAlreadyExists(err) {
			logger.Info("Helm job already exists")
			return nil
		}
		return fmt.Errorf("failed to create helm job: %w", err)
	}

	logger.Info("Created Helm installation job", zap.String("job", job.Name))

	// Update status to installing/upgrading
	if isUpgrade {
		helmChart.Status.State = environmentsv1.HelmChartStateUpgrading
		helmChart.Status.Message = "Upgrading helm chart"
	} else {
		helmChart.Status.State = environmentsv1.HelmChartStateInstalling
		helmChart.Status.Message = "Installing helm chart"
	}

	if err := r.Status().Update(ctx, helmChart); err != nil {
		logger.Warn("Failed to update status", zap.Error(err))
	}

	return nil
}

// createValuesConfigMap creates a ConfigMap containing the Helm values
func (r *Reconciler) createValuesConfigMap(ctx context.Context, helmChart *environmentsv1.HelmChart, logger *zap.Logger) (*corev1.ConfigMap, error) {
	// Convert helmValues map to YAML string
	valuesYAML := ""
	for key, value := range helmChart.Spec.HelmValues {
		// Simple YAML generation - in production, use a proper YAML library
		valuesYAML += fmt.Sprintf("%s: %s\n", key, string(value.Raw))
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-helm-values", helmChart.Name),
			Namespace: helmChart.Namespace,
			Labels: map[string]string{
				"kloudlite.io/helm-chart": helmChart.Name,
				"kloudlite.io/managed-by": "helmchart-controller",
			},
		},
		Data: map[string]string{
			"values.yaml": valuesYAML,
		},
	}

	// Set controller ownership
	if err := controllerutil.SetControllerReference(helmChart, configMap, r.Scheme); err != nil {
		return nil, err
	}

	// Try to get existing ConfigMap
	existing := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKeyFromObject(configMap), existing)
	if err == nil {
		// Update existing ConfigMap
		existing.Data = configMap.Data
		if err := r.Update(ctx, existing); err != nil {
			return nil, err
		}
		logger.Info("Updated values ConfigMap", zap.String("name", existing.Name))
		return existing, nil
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	// Create new ConfigMap
	if err := r.Create(ctx, configMap); err != nil {
		return nil, err
	}

	logger.Info("Created values ConfigMap", zap.String("name", configMap.Name))
	return configMap, nil
}

// createHelmJob creates a Kubernetes Job that runs helm install/upgrade
func (r *Reconciler) createHelmJob(helmChart *environmentsv1.HelmChart, valuesConfigMapName string, isUpgrade bool) *batchv1.Job {
	// Build helm command
	helmCommand := "helm repo add chart-repo " + helmChart.Spec.Chart.URL + " && "

	// Pre-install hook
	if helmChart.Spec.PreInstall != "" && !isUpgrade {
		helmCommand += helmChart.Spec.PreInstall + " && "
	}

	// Main helm command
	action := "install"
	if isUpgrade {
		action = "upgrade --install"
	}

	helmCommand += fmt.Sprintf("helm %s %s chart-repo/%s",
		action,
		helmChart.Status.ReleaseName,
		helmChart.Spec.Chart.Name)

	// Add version if specified
	if helmChart.Spec.Chart.Version != "" {
		helmCommand += fmt.Sprintf(" --version %s", helmChart.Spec.Chart.Version)
	}

	// Add values file
	helmCommand += " --values /values/values.yaml"

	// Add namespace
	helmCommand += fmt.Sprintf(" --namespace %s", helmChart.Namespace)

	// Wait for installation
	helmCommand += " --wait --timeout 10m"

	// Post-install hook
	if helmChart.Spec.PostInstall != "" && !isUpgrade {
		helmCommand += " && " + helmChart.Spec.PostInstall
	}

	// Default job vars
	jobVars := &environmentsv1.HelmJobVars{}
	if helmChart.Spec.HelmJobVars != nil {
		jobVars = helmChart.Spec.HelmJobVars
	}

	// Set default resources if not specified
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
	if jobVars.Resources.Requests != nil || jobVars.Resources.Limits != nil {
		resources = jobVars.Resources
	}

	backoffLimit := int32(3)
	ttlSecondsAfterFinished := int32(3600) // Clean up after 1 hour

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-helm-install-%d", helmChart.Name, time.Now().Unix()),
			Namespace: helmChart.Namespace,
			Labels: map[string]string{
				"kloudlite.io/helm-chart": helmChart.Name,
				"kloudlite.io/job-type":   "install",
				"kloudlite.io/managed-by": "helmchart-controller",
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kloudlite.io/helm-chart": helmChart.Name,
						"kloudlite.io/job-type":   "install",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyNever,
					NodeSelector:       jobVars.NodeSelector,
					Tolerations:        jobVars.Tolerations,
					Affinity:           jobVars.Affinity,
					ServiceAccountName: "default",
					Containers: []corev1.Container{
						{
							Name:      "helm",
							Image:     helmJobImage,
							Command:   []string{"/bin/sh", "-c"},
							Args:      []string{helmCommand},
							Resources: resources,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "values",
									MountPath: "/values",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "values",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: valuesConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return job
}

// getHelmJob retrieves the helm job for the given chart
func (r *Reconciler) getHelmJob(ctx context.Context, helmChart *environmentsv1.HelmChart, jobType string) (*batchv1.Job, error) {
	jobList := &batchv1.JobList{}
	err := r.List(ctx, jobList,
		client.InNamespace(helmChart.Namespace),
		client.MatchingLabels{
			"kloudlite.io/helm-chart": helmChart.Name,
			"kloudlite.io/job-type":   jobType,
		})
	if err != nil {
		return nil, err
	}

	if len(jobList.Items) == 0 {
		return nil, apierrors.NewNotFound(batchv1.Resource("job"), "helm-job")
	}

	// Return the most recent job
	mostRecentJob := &jobList.Items[0]
	for i := range jobList.Items {
		if jobList.Items[i].CreationTimestamp.After(mostRecentJob.CreationTimestamp.Time) {
			mostRecentJob = &jobList.Items[i]
		}
	}

	return mostRecentJob, nil
}

// checkJobStatus checks the status of the helm job and updates the HelmChart status
func (r *Reconciler) checkJobStatus(ctx context.Context, helmChart *environmentsv1.HelmChart, job *batchv1.Job, logger *zap.Logger) (reconcile.Result, error) {
	// Check if job succeeded
	if job.Status.Succeeded > 0 {
		logger.Info("Helm installation job succeeded")
		return r.updateStatus(ctx, helmChart, environmentsv1.HelmChartStateInstalled,
			"Helm chart installed successfully", logger)
	}

	// Check if job failed
	if job.Status.Failed > 0 {
		logger.Error("Helm installation job failed",
			zap.Int32("failed", job.Status.Failed))

		// Get pod logs for debugging
		message := fmt.Sprintf("Helm installation failed after %d attempts", job.Status.Failed)

		return r.updateStatus(ctx, helmChart, environmentsv1.HelmChartStateFailed, message, logger)
	}

	// Job is still running
	logger.Info("Helm installation job is still running",
		zap.Int32("active", job.Status.Active))

	// Requeue to check again
	return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
}

// handleDeletion handles cleanup when HelmChart is being deleted
func (r *Reconciler) handleDeletion(ctx context.Context, helmChart *environmentsv1.HelmChart, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Cleaning up HelmChart resources")

	// Update status to uninstalling if not already set
	if helmChart.Status.State != environmentsv1.HelmChartStateUninstalling &&
		helmChart.Status.State != environmentsv1.HelmChartStateDeleting {
		helmChart.Status.State = environmentsv1.HelmChartStateUninstalling
		helmChart.Status.Message = "Uninstalling helm chart"
		if err := r.Status().Update(ctx, helmChart); err != nil {
			logger.Error("Failed to update status to uninstalling", zap.Error(err))
		}
	}

	// Check if uninstall job exists
	uninstallJob, err := r.getHelmJob(ctx, helmChart, "uninstall")
	if err == nil && uninstallJob != nil {
		// Check uninstall job status
		if uninstallJob.Status.Succeeded > 0 {
			logger.Info("Helm uninstall job succeeded")
			// Continue to remove finalizer
		} else if uninstallJob.Status.Failed > 0 {
			logger.Warn("Helm uninstall job failed, continuing with cleanup")
			// Continue anyway
		} else {
			// Job still running
			logger.Info("Waiting for helm uninstall job to complete")
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	} else {
		// Create uninstall job
		if err := r.createUninstallJob(ctx, helmChart, logger); err != nil {
			logger.Error("Failed to create uninstall job", zap.Error(err))
			// Continue with cleanup even if uninstall fails
		} else {
			// Wait for uninstall to complete
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}
	}

	// Clean up jobs and ConfigMaps
	if err := r.cleanupResources(ctx, helmChart, logger); err != nil {
		logger.Error("Failed to cleanup resources", zap.Error(err))
		// Continue anyway
	}

	// Remove finalizer
	logger.Info("Removing finalizer from HelmChart")
	controllerutil.RemoveFinalizer(helmChart, helmChartFinalizer)
	if err := r.Update(ctx, helmChart); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("HelmChart cleanup completed successfully")
	return reconcile.Result{}, nil
}

// createUninstallJob creates a job to uninstall the helm release
func (r *Reconciler) createUninstallJob(ctx context.Context, helmChart *environmentsv1.HelmChart, logger *zap.Logger) error {
	// Pre-uninstall hook
	preUninstall := ""
	if helmChart.Spec.PreUninstall != "" {
		preUninstall = helmChart.Spec.PreUninstall + " && "
	}

	// Helm uninstall command
	helmCommand := fmt.Sprintf("%shelm uninstall %s --namespace %s --wait",
		preUninstall,
		helmChart.Status.ReleaseName,
		helmChart.Namespace)

	// Post-uninstall hook
	if helmChart.Spec.PostUninstall != "" {
		helmCommand += " && " + helmChart.Spec.PostUninstall
	}

	// Default job vars
	jobVars := &environmentsv1.HelmJobVars{}
	if helmChart.Spec.HelmJobVars != nil {
		jobVars = helmChart.Spec.HelmJobVars
	}

	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
	if jobVars.Resources.Requests != nil || jobVars.Resources.Limits != nil {
		resources = jobVars.Resources
	}

	backoffLimit := int32(3)
	ttlSecondsAfterFinished := int32(600)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-helm-uninstall-%d", helmChart.Name, time.Now().Unix()),
			Namespace: helmChart.Namespace,
			Labels: map[string]string{
				"kloudlite.io/helm-chart": helmChart.Name,
				"kloudlite.io/job-type":   "uninstall",
				"kloudlite.io/managed-by": "helmchart-controller",
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kloudlite.io/helm-chart": helmChart.Name,
						"kloudlite.io/job-type":   "uninstall",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyNever,
					NodeSelector:       jobVars.NodeSelector,
					Tolerations:        jobVars.Tolerations,
					Affinity:           jobVars.Affinity,
					ServiceAccountName: "default",
					Containers: []corev1.Container{
						{
							Name:      "helm",
							Image:     helmJobImage,
							Command:   []string{"/bin/sh", "-c"},
							Args:      []string{helmCommand},
							Resources: resources,
						},
					},
				},
			},
		},
	}

	// Set controller ownership
	if err := controllerutil.SetControllerReference(helmChart, job, r.Scheme); err != nil {
		return err
	}

	// Create the job
	if err := r.Create(ctx, job); err != nil {
		if apierrors.IsAlreadyExists(err) {
			logger.Info("Helm uninstall job already exists")
			return nil
		}
		return err
	}

	logger.Info("Created Helm uninstall job", zap.String("job", job.Name))
	return nil
}

// cleanupResources cleans up Jobs and ConfigMaps created by the controller
func (r *Reconciler) cleanupResources(ctx context.Context, helmChart *environmentsv1.HelmChart, logger *zap.Logger) error {
	// Delete jobs
	jobList := &batchv1.JobList{}
	if err := r.List(ctx, jobList,
		client.InNamespace(helmChart.Namespace),
		client.MatchingLabels{"kloudlite.io/helm-chart": helmChart.Name}); err != nil {
		return err
	}

	for _, job := range jobList.Items {
		if err := r.Delete(ctx, &job); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("Failed to delete job", zap.String("name", job.Name), zap.Error(err))
		}
	}

	// Delete ConfigMaps
	configMapList := &corev1.ConfigMapList{}
	if err := r.List(ctx, configMapList,
		client.InNamespace(helmChart.Namespace),
		client.MatchingLabels{"kloudlite.io/helm-chart": helmChart.Name}); err != nil {
		return err
	}

	for _, cm := range configMapList.Items {
		if err := r.Delete(ctx, &cm); err != nil && !apierrors.IsNotFound(err) {
			logger.Error("Failed to delete ConfigMap", zap.String("name", cm.Name), zap.Error(err))
		}
	}

	return nil
}

// updateStatus updates the status of the HelmChart
func (r *Reconciler) updateStatus(ctx context.Context, helmChart *environmentsv1.HelmChart, state environmentsv1.HelmChartState, message string, logger *zap.Logger) (reconcile.Result, error) {
	helmChart.Status.State = state
	helmChart.Status.Message = message
	helmChart.Status.ObservedGeneration = helmChart.Generation

	now := metav1.Now()
	if state == environmentsv1.HelmChartStateInstalled {
		helmChart.Status.LastInstallTime = &now
		helmChart.Status.InstalledVersion = helmChart.Spec.Chart.Version
	}

	// Update condition
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		ObservedGeneration: helmChart.Generation,
		LastTransitionTime: now,
		Reason:             string(state),
		Message:            message,
	}

	if state == environmentsv1.HelmChartStateInstalled {
		readyCondition.Status = metav1.ConditionTrue
	}

	// Update or add condition
	found := false
	for i, condition := range helmChart.Status.Conditions {
		if condition.Type == "Ready" {
			helmChart.Status.Conditions[i] = readyCondition
			found = true
			break
		}
	}
	if !found {
		helmChart.Status.Conditions = append(helmChart.Status.Conditions, readyCondition)
	}

	if err := r.Status().Update(ctx, helmChart); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Updated HelmChart status",
		zap.String("state", string(state)),
		zap.String("message", message))

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.HelmChart{}).
		Owns(&batchv1.Job{}).      // Watch Jobs owned by HelmChart
		Owns(&corev1.ConfigMap{}). // Watch ConfigMaps owned by HelmChart
		Complete(r)
}
