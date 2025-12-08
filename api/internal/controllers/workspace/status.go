package workspace

import (
	"context"
	"fmt"
	"reflect"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// updateStatus updates workspace status with retry logic
// Following Kubernetes best practice: "re-fetch before update" to avoid stale data on conflict retry
func (r *WorkspaceReconciler) updateStatus(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// ConnectedEnvironment is set during THIS reconciliation, preserve it
	connectedEnvironment := workspace.Status.ConnectedEnvironment

	return statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
		// Restore ConnectedEnvironment (set during this reconciliation)
		workspace.Status.ConnectedEnvironment = connectedEnvironment
		return nil
	}, logger)
}

// updateWorkspaceStatus updates the workspace status based on pod state
func (r *WorkspaceReconciler) updateWorkspaceStatus(ctx context.Context, workspace *workspacev1.Workspace, pod *corev1.Pod, phase, message string, logger *zap.Logger) (reconcile.Result, error) {
	// Track if any status field actually changed
	needsUpdate := false

	if workspace.Status.Phase != phase {
		workspace.Status.Phase = phase
		needsUpdate = true
	}

	if workspace.Status.Message != message {
		workspace.Status.Message = message
		needsUpdate = true
	}

	if workspace.Status.PodName != pod.Name {
		workspace.Status.PodName = pod.Name
		needsUpdate = true
	}

	if workspace.Status.PodIP != pod.Status.PodIP {
		workspace.Status.PodIP = pod.Status.PodIP
		needsUpdate = true
	}

	if workspace.Status.NodeName != pod.Spec.NodeName {
		workspace.Status.NodeName = pod.Spec.NodeName
		needsUpdate = true
	}

	// Build access URLs for all services if pod is running
	if pod.Status.PodIP != "" && phase == "Running" {
		accessURLs := make(map[string]string)

		// Try to fetch DomainRequest - it may not exist yet if WorkMachine is still setting up
		// DomainRequest is cluster-scoped with name "installation-domain"
		var domainRequest domainrequestv1.DomainRequest
		domainRequestExists := false
		if err := r.Get(ctx, fn.NN("", "installation-domain"), &domainRequest); err != nil {
			logger.Warn("DomainRequest 'installation-domain' not found, using pod IP fallback URLs",
				zap.Error(err))
		} else {
			domainRequestExists = true
		}

		// Compute hash for DNS-safe hostnames: hash(owner-workspaceName)
		wsHash := generateHash(fmt.Sprintf("%s-%s", workspace.Spec.OwnedBy, workspace.Name))

		// Update hash in status if changed
		if workspace.Status.Hash != wsHash {
			workspace.Status.Hash = wsHash
			needsUpdate = true
		}

		// Update subdomain in status if changed
		subdomain := ""
		if domainRequestExists && domainRequest.Status.Subdomain != "" {
			subdomain = domainRequest.Status.Subdomain
		}
		if workspace.Status.Subdomain != subdomain {
			workspace.Status.Subdomain = subdomain
			needsUpdate = true
		}

		// Update ExposedRoutes in status based on Expose spec
		exposedRoutes := make(map[string]string)
		if subdomain != "" && wsHash != "" {
			for _, exposed := range workspace.Spec.Expose {
				portStr := fmt.Sprintf("%d", exposed.Port)
				exposedRoutes[portStr] = fmt.Sprintf("https://p%d-%s.%s", exposed.Port, wsHash, subdomain)
			}
		}
		if !reflect.DeepEqual(workspace.Status.ExposedRoutes, exposedRoutes) {
			workspace.Status.ExposedRoutes = exposedRoutes
			needsUpdate = true
		}

		// Try to use public domain URLs if DomainRequest is available
		if domainRequestExists && domainRequest.Status.Subdomain != "" {
			// Get WorkMachine to construct domain URLs
			wm, err := r.getWorkMachine(ctx, workspace.Spec.WorkmachineName)
			if err == nil && wm.Status.PublicIP != "" {
				// Use hash-based pattern: {prefix}-{hash}.{subdomain}
				// Example: claude-a1b2c3d4.eastman.khost.dev
				// Use public HTTPS domain URLs
				accessURLs["code-server"] = fmt.Sprintf("https://vscode-%s.%s", wsHash, subdomain)
				accessURLs["ttyd"] = fmt.Sprintf("https://tty-%s.%s", wsHash, subdomain)
				accessURLs["claude-ttyd"] = fmt.Sprintf("https://claude-%s.%s", wsHash, subdomain)
				accessURLs["opencode-ttyd"] = fmt.Sprintf("https://opencode-%s.%s", wsHash, subdomain)
				accessURLs["codex-ttyd"] = fmt.Sprintf("https://codex-%s.%s", wsHash, subdomain)
				// SSH is still via pod IP (not routed through HAProxy)
				accessURLs["ssh"] = fmt.Sprintf("ssh://%s:22", pod.Status.PodIP)
			} else {
				// Fallback to internal pod IPs
				logger.Warn("Failed to get WorkMachine for domain URLs, using pod IPs",
					zap.String("workmachine", workspace.Spec.WorkmachineName),
					zap.Error(err))
				accessURLs["ssh"] = fmt.Sprintf("ssh://%s:22", pod.Status.PodIP)
				accessURLs["code-server"] = fmt.Sprintf("http://%s:8080", pod.Status.PodIP)
				accessURLs["ttyd"] = fmt.Sprintf("http://%s:7681", pod.Status.PodIP)
				accessURLs["vscode-tunnel"] = fmt.Sprintf("http://%s:8000", pod.Status.PodIP)
			}
		} else {
			// Fallback to internal pod IPs (DomainRequest not available or no subdomain)
			accessURLs["ssh"] = fmt.Sprintf("ssh://%s:22", pod.Status.PodIP)
			accessURLs["code-server"] = fmt.Sprintf("http://%s:8080", pod.Status.PodIP)
			accessURLs["ttyd"] = fmt.Sprintf("http://%s:7681", pod.Status.PodIP)
			accessURLs["vscode-tunnel"] = fmt.Sprintf("http://%s:8000", pod.Status.PodIP)
		}

		// Only update if AccessURLs changed
		if !reflect.DeepEqual(workspace.Status.AccessURLs, accessURLs) {
			workspace.Status.AccessURLs = accessURLs
			needsUpdate = true
		}

		// Keep AccessURL for backward compatibility (default to code-server)
		if workspace.Status.AccessURL != accessURLs["code-server"] {
			workspace.Status.AccessURL = accessURLs["code-server"]
			needsUpdate = true
		}
	}

	// Store old ConnectedEnvironment for comparison
	oldConnectedEnvironment := workspace.Status.ConnectedEnvironment

	// Update ConnectedEnvironment status if EnvironmentConnection is set
	if workspace.Spec.EnvironmentConnection != nil {
		env := &environmentv1.Environment{}
		err := r.Get(ctx, client.ObjectKey{
			Name: workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
		}, env)

		if err == nil && env.Spec.Activated {
			// Update connected environment status - only if environment spec.activated is true
			// Check spec.activated instead of status.state to avoid timeout during activation
			// Use display name format: {owner}/{envName}
			displayName := fmt.Sprintf("%s/%s", env.Spec.OwnedBy, env.Spec.Name)
			workspace.Status.ConnectedEnvironment = &workspacev1.ConnectedEnvironmentInfo{
				Name:            displayName,
				TargetNamespace: env.Spec.TargetNamespace,
			}
			logger.Info("Updated ConnectedEnvironment status",
				zap.String("workspace", workspace.Name),
				zap.String("environment", displayName),
				zap.String("targetNamespace", env.Spec.TargetNamespace),
			)
		} else if err != nil {
			// Environment not found or fetch failed - set to nil
			workspace.Status.ConnectedEnvironment = nil
			logger.Warn("Failed to fetch environment for status update",
				zap.String("workspace", workspace.Name),
				zap.String("environment", workspace.Spec.EnvironmentConnection.EnvironmentRef.Name),
				zap.Error(err),
			)
		} else {
			// Environment exists but not activated (spec.activated = false) - set to nil
			workspace.Status.ConnectedEnvironment = nil
			logger.Info("Environment exists but not activated",
				zap.String("workspace", workspace.Name),
				zap.String("environment", env.Name),
			)
		}
	} else {
		// No environment connection, clear connected environment status
		workspace.Status.ConnectedEnvironment = nil
	}

	// Check if ConnectedEnvironment changed
	if !reflect.DeepEqual(oldConnectedEnvironment, workspace.Status.ConnectedEnvironment) {
		needsUpdate = true
	}

	// Only update status if something actually changed
	if !needsUpdate {
		logger.Debug("Workspace status unchanged, skipping status update")
		return reconcile.Result{}, nil
	}

	if err := r.updateStatus(ctx, workspace, logger); err != nil {
		logger.Error("Failed to update workspace status after retries", zap.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// applyLabelsAndAnnotations applies standard labels and annotations to workspace resources
func (r *WorkspaceReconciler) applyLabelsAndAnnotations(obj metav1.Object, workspace *workspacev1.Workspace) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	// Apply standard labels using utils for consistency
	labels["app"] = "workspace"
	labels["workspace"] = workspace.Name
	labels["workspaces.kloudlite.io/workspace-name"] = workspace.Name
	labels["kloudlite.io/workspace-owner"] = workspace.Spec.OwnedBy

	if workspace.Spec.DisplayName != "" {
		labels["kloudlite.io/workspace-display-name"] = utils.SanitizeForLabel(workspace.Spec.DisplayName)
	}

	obj.SetLabels(labels)

	// Apply annotations if provided
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	if workspace.Spec.DisplayName != "" {
		annotations["kloudlite.io/workspace-display-name"] = workspace.Spec.DisplayName
	}
	annotations["kloudlite.io/workspace-owner"] = workspace.Spec.OwnedBy

	obj.SetAnnotations(annotations)
}

// updateWorkspaceStatusWithConditions updates workspace status with proper condition management
func (r *WorkspaceReconciler) updateWorkspaceStatusWithConditions(ctx context.Context, workspace *workspacev1.Workspace, phase, message string, logger *zap.Logger) error {
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
		// Initialize conditions if not present
		if workspace.Status.Conditions == nil {
			workspace.Status.Conditions = []metav1.Condition{}
		}

		// Update phase and message
		workspace.Status.Phase = phase
		workspace.Status.Message = message

		// Update ready condition based on phase
		readyStatus := metav1.ConditionFalse
		reason := "NotReady"
		if phase == "Running" {
			readyStatus = metav1.ConditionTrue
			reason = "WorkspaceRunning"
		} else if phase == "Failed" {
			readyStatus = metav1.ConditionFalse
			reason = "WorkspaceFailed"
		} else if phase == "Creating" {
			readyStatus = metav1.ConditionFalse
			reason = "WorkspaceCreating"
		} else if phase == "Stopping" || phase == "Stopped" {
			readyStatus = metav1.ConditionFalse
			reason = "WorkspaceStopped"
		}

		// Add or update Ready condition
		now := metav1.Now()
		r.addOrUpdateWorkspaceCondition(workspace, "Ready", readyStatus, reason, message, &now)

		return nil
	}, logger)
}

// addOrUpdateWorkspaceCondition adds or updates a workspace condition
func (r *WorkspaceReconciler) addOrUpdateWorkspaceCondition(workspace *workspacev1.Workspace, conditionType string, status metav1.ConditionStatus, reason, message string, transitionTime *metav1.Time) {
	// Find existing condition
	for i, condition := range workspace.Status.Conditions {
		if condition.Type == conditionType {
			// Update existing condition
			workspace.Status.Conditions[i].Status = status
			workspace.Status.Conditions[i].Reason = reason
			workspace.Status.Conditions[i].Message = message
			if condition.Status != status {
				workspace.Status.Conditions[i].LastTransitionTime = *transitionTime
			}
			return
		}
	}

	// Add new condition
	newCondition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: *transitionTime,
		Reason:             reason,
		Message:            message,
	}
	workspace.Status.Conditions = append(workspace.Status.Conditions, newCondition)
}
