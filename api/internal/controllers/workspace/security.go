package workspace

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// validateCommandForExec validates that commands are safe for execution in pods
func (r *WorkspaceReconciler) validateCommandForExec(command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("command cannot be empty")
	}

	// Define allowed commands and their safe patterns
	allowedCommands := map[string]bool{
		"sh":   true,
		"awk":  true,
		"wc":   true,
		"cat":  true,
		"grep": true,
	}

	// Check first argument is an allowed command
	if !allowedCommands[command[0]] {
		return fmt.Errorf("command not allowed: %s", command[0])
	}

	// For shell commands, validate the script content
	if command[0] == "sh" && len(command) > 2 && command[1] == "-c" {
		script := command[2]

		// Check for potentially dangerous shell constructs
		dangerousPatterns := []string{
			";rm", ";:", ";&&", ";||",
			">/dev/", "< /dev/",
			"&& rm", "|| rm",
			"$(", "`", "eval", "exec",
			"wget ", "curl ", "nc ", "netcat",
		}

		scriptLower := strings.ToLower(script)
		for _, pattern := range dangerousPatterns {
			if strings.Contains(scriptLower, pattern) {
				return fmt.Errorf("potentially dangerous command pattern detected: %s", pattern)
			}
		}

		// Allow specific safe patterns for DNS and connection checking
		allowedPatterns := []string{
			"awk '$4 == \"01\"'",          // connection counting
			"/proc/net/tcp",               // network stats
			"/proc/net/tcp6",              // IPv6 network stats
			"wc -l",                       // line counting
			"cat >",                       // file writing for DNS config
			"/etc/resolv.conf",            // DNS config file
			"/tmp/kloudlite-context.json", // Kloudlite context file for Starship prompt
		}

		isAllowed := false
		for _, pattern := range allowedPatterns {
			if strings.Contains(script, pattern) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return fmt.Errorf("command pattern not allowed: %s", script[:min(100, len(script))])
		}
	}

	return nil
}

// execInPod executes a command in a pod container and returns the stdout output
func (r *WorkspaceReconciler) execInPod(ctx context.Context, pod *corev1.Pod, containerName string, command []string) (string, error) {
	// Validate command before execution
	if err := r.validateCommandForExec(command); err != nil {
		return "", fmt.Errorf("command validation failed: %w", err)
	}

	// This function requires the reconciler to have Config and Clientset
	// These should be added to the WorkspaceReconciler struct
	if r.Config == nil || r.Clientset == nil {
		return "", fmt.Errorf("workspace reconciler missing Config or Clientset for pod execution")
	}

	req := r.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
		Stdin:     false,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(r.Config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", fmt.Errorf("failed to exec command: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

// validateHostPath validates that a host path is safe for cleanup operations
func (r *WorkspaceReconciler) validateHostPath(hostPath string, workspaceName string) error {
	// Use the enhanced validation function from utils
	return utils.ValidateHostPathForWorkspace(hostPath, workspaceName)
}

// validateEnvironmentConnection validates environment reference and returns environment details
func (r *WorkspaceReconciler) validateEnvironmentConnection(ctx context.Context, workspace *workspacev1.Workspace) (*environmentv1.Environment, error) {
	if workspace.Spec.EnvironmentConnection == nil {
		return nil, nil // No environment connection is valid
	}

	env := &environmentv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{
		Name: workspace.Spec.EnvironmentConnection.EnvironmentRef.Name,
	}, env)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("environment '%s' not found", workspace.Spec.EnvironmentConnection.EnvironmentRef.Name)
		}
		return nil, fmt.Errorf("failed to get environment '%s': %w", workspace.Spec.EnvironmentConnection.EnvironmentRef.Name, err)
	}

	if !env.Spec.Activated {
		return nil, fmt.Errorf("environment '%s' is not activated", env.Name)
	}

	// Validate target namespace
	if err := utils.ValidateKubernetesNamespace(env.Spec.TargetNamespace); err != nil {
		return nil, fmt.Errorf("environment '%s' has invalid target namespace: %w", env.Name, err)
	}

	return env, nil
}

// buildDNSSearchDomains builds and validates DNS search domains based on environment connection
func (r *WorkspaceReconciler) buildDNSSearchDomains(env *environmentv1.Environment) (string, error) {
	var domains []string

	if env != nil && env.Spec.Activated {
		// Include validated environment namespace in search domains
		envDomain := fmt.Sprintf("%s.svc.cluster.local", env.Spec.TargetNamespace)
		domains = []string{envDomain, "svc.cluster.local", "cluster.local"}
	} else {
		// Default domains
		domains = []string{"svc.cluster.local", "cluster.local"}
	}

	// Sanitize search domains to prevent DNS injection
	searchDomains, err := utils.SanitizeSearchDomains(domains)
	if err != nil {
		return "", fmt.Errorf("failed to sanitize search domains: %w", err)
	}

	return searchDomains, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
