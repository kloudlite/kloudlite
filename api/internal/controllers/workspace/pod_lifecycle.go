package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// updateDNSConfigInRunningPod updates /etc/resolv.conf in a running workspace pod
// when the environment connection changes
func (r *WorkspaceReconciler) updateDNSConfigInRunningPod(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Get the target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to get target namespace: %w", err)
	}

	// Get the pod
	pod := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Only update if pod is running
	if pod.Status.Phase != corev1.PodRunning {
		logger.Info("Skipping DNS update - pod is not running",
			zap.String("phase", string(pod.Status.Phase)))
		return nil
	}

	// Build search domains based on environment connection with validation
	var domains []string
	if workspace.Spec.EnvironmentConnection != nil {
		env := &environmentv1.Environment{}
		err := r.Get(ctx, client.ObjectKey{
			Name: workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
		}, env)
		if err == nil && env.Spec.Activated {
			// Validate environment namespace for security
			// Include validated environment namespace in search domains
			envDomain := fmt.Sprintf("%s.svc.cluster.local", env.Spec.TargetNamespace)
			domains = []string{envDomain}
			logger.Info("Environment connection detected for DNS update",
				zap.String("environment", env.Name),
				zap.String("targetNamespace", env.Spec.TargetNamespace))
		}
	}

	domains = append(domains, "svc.cluster.local", "cluster.local")

	// Build new resolv.conf content with validated domains
	resolvConf := fmt.Sprintf("nameserver 10.43.0.10\nsearch %s\noptions ndots:5\n", strings.Join(domains, " "))

	// Exec into pod and update /etc/resolv.conf
	// Note: /etc/resolv.conf is mounted from EmptyDir with ReadOnly: false, so it's writable
	command := []string{"sh", "-c", fmt.Sprintf("cat > /etc/resolv.conf << 'EOFR'\n%sEOFR\n", resolvConf)}
	_, err = r.execInPod(ctx, pod, "workspace", command)
	if err != nil {
		return fmt.Errorf("failed to update DNS config: %w", err)
	}

	logger.Info("Successfully updated DNS configuration in running pod", zap.String("workspace", workspace.Name))

	return nil
}

// updateKloudliteContextFile writes the Kloudlite context state to a file in the running pod
// This file is used by kloudlite-context.sh for fast prompt rendering without API calls
func (r *WorkspaceReconciler) updateKloudliteContextFile(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Get the target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to get target namespace: %w", err)
	}

	// Get the pod
	pod := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}

	// Only update if pod is running
	if pod.Status.Phase != corev1.PodRunning {
		logger.Info("Skipping context file update - pod is not running",
			zap.String("phase", string(pod.Status.Phase)))
		return nil
	}

	// Get environment name from spec
	envName := ""
	if workspace.Spec.EnvironmentConnection != nil {
		// Fetch environment to validate it exists and is activated
		env := &environmentv1.Environment{}
		err := r.Get(ctx, client.ObjectKey{
			Name: workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
		}, env)
		if err == nil && env.Spec.Activated {
			envName = env.Name
		}
	}

	// Get active service intercepts from workspace status
	// The status is populated by the collectActiveIntercepts function during status updates
	intercepts := []string{}
	for _, interceptStatus := range workspace.Status.ActiveIntercepts {
		intercepts = append(intercepts, interceptStatus.ServiceName)
	}

	// Build JSON content
	contextData := map[string]interface{}{
		"environment": envName,
		"intercepts":  intercepts,
	}

	jsonBytes, err := json.Marshal(contextData)
	if err != nil {
		return fmt.Errorf("failed to marshal context data: %w", err)
	}

	// Write to /tmp/kloudlite-context.json in the pod
	command := []string{"sh", "-c", fmt.Sprintf("cat > /tmp/kloudlite-context.json << 'EOF'\n%s\nEOF\n", string(jsonBytes))}
	_, err = r.execInPod(ctx, pod, "workspace", command)
	if err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	logger.Info("Successfully updated Kloudlite context file in running pod",
		zap.String("workspace", workspace.Name),
		zap.String("environment", envName),
		zap.Strings("intercepts", intercepts))

	return nil
}

// getWorkMachine fetches the WorkMachine resource by name
func (r *WorkspaceReconciler) getWorkMachine(ctx context.Context, name string) (*machinesv1.WorkMachine, error) {
	wm := &machinesv1.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: name}, wm); err != nil {
		return nil, fmt.Errorf("failed to get WorkMachine %s: %w", name, err)
	}
	return wm, nil
}
