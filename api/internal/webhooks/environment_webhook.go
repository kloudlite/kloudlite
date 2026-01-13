package webhooks

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EnvironmentWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
	clientset *kubernetes.Clientset
}

func NewEnvironmentWebhook(logger logger.Logger, k8sClient client.Client, clientset *kubernetes.Clientset) *EnvironmentWebhook {
	return &EnvironmentWebhook{
		logger:    logger,
		k8sClient: k8sClient,
		clientset: clientset,
	}
}

// ValidateEnvironment handles validation webhook for Environment CRD
func (w *EnvironmentWebhook) ValidateEnvironment(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	// Process the admission request
	response := w.handleValidation(admissionReview.Request)

	// Build the admission review response
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

// MutateEnvironment handles mutation webhook for Environment CRD
func (w *EnvironmentWebhook) MutateEnvironment(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	// Process the admission request
	response := w.handleMutation(admissionReview.Request)

	// Build the admission review response
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *EnvironmentWebhook) handleValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the environment object
	var env environmentsv1.Environment
	if err := json.Unmarshal(req.Object.Raw, &env); err != nil {
		w.logger.Error("Failed to unmarshal environment: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal environment object",
			},
		}
	}

	// Perform validation
	if err := w.validateEnvironment(&env, req.Operation); err != nil {
		w.logger.Warn("Environment validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}

func (w *EnvironmentWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the environment object
	var env environmentsv1.Environment
	if err := json.Unmarshal(req.Object.Raw, &env); err != nil {
		w.logger.Error("Failed to unmarshal environment: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal environment object",
			},
		}
	}

	// Create patches for mutations
	var patches []map[string]interface{}

	// Set default Activated to true for new environments
	if req.Operation == admissionv1.Create && !env.Spec.Activated {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/activated",
			"value": true,
		})
		// Update the env object for subsequent checks
		env.Spec.Activated = true
		w.logger.Info(fmt.Sprintf("Setting activated=true by default for new environment: %s", env.Name))
	}

	// Derive OwnedBy from namespace label if not provided
	// Environment is namespaced, so we can look up the namespace to get the owner
	var userName string
	if env.Spec.OwnedBy != "" {
		ownedBy := env.Spec.OwnedBy
		// Determine if OwnedBy is an email or username
		if strings.Contains(ownedBy, "@") {
			// OwnedBy is an email - find the actual username
			userList := &platformv1alpha1.UserList{}
			if err := w.k8sClient.List(context.Background(), userList); err == nil {
				for _, u := range userList.Items {
					if u.Spec.Email == ownedBy {
						userName = u.Name
						break
					}
				}
			}
			// If no user found, use a sanitized version of email as username
			if userName == "" {
				userName = strings.ReplaceAll(strings.Split(ownedBy, "@")[0], ".", "-")
			}
		} else {
			userName = ownedBy
		}
	} else {
		// Derive owner from namespace label
		ns := &corev1.Namespace{}
		if err := w.k8sClient.Get(context.Background(), client.ObjectKey{Name: env.Namespace}, ns); err == nil {
			if ns.Labels != nil {
				userName = ns.Labels["kloudlite.io/owned-by"]
			}
		}
		// Set OwnedBy in spec if we derived it
		if userName != "" {
			patches = append(patches, map[string]interface{}{
				"op":    "add",
				"path":  "/spec/ownedBy",
				"value": userName,
			})
			env.Spec.OwnedBy = userName
		}
	}

	// Derive WorkMachineName from namespace if not provided
	// The namespace name is the WorkMachine name (e.g., wm-karthik)
	if env.Spec.WorkMachineName == "" && strings.HasPrefix(env.Namespace, "wm-") {
		workMachineName := env.Namespace // namespace IS the workmachine name
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/workmachineName",
			"value": workMachineName,
		})
		env.Spec.WorkMachineName = workMachineName
		w.logger.Info(fmt.Sprintf("Derived WorkMachineName: %s for environment: %s", workMachineName, env.Name))
	}

	// Generate targetNamespace if not provided
	// Format: env-{envName}-{random6} to avoid conflicts
	if env.Spec.TargetNamespace == "" {
		suffix := generateRandomSuffix(6)
		targetNamespace := fmt.Sprintf("env-%s-%s", env.Name, suffix)
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/targetNamespace",
			"value": targetNamespace,
		})
		// Update the env object for subsequent checks
		env.Spec.TargetNamespace = targetNamespace
		w.logger.Info(fmt.Sprintf("Generated targetNamespace: %s for environment: %s", targetNamespace, env.Name))
	}

	// Add default labels if not present
	if env.Spec.Labels == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/labels",
			"value": map[string]string{},
		})
	}

	// Add environment label to identify resources
	labelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/labels/kloudlite.io~1environment-name",
		"value": env.Name,
	}
	patches = append(patches, labelPatch)

	// Add managed-by label
	managedByPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/labels/kloudlite.io~1managed-by",
		"value": "environment-controller",
	}
	patches = append(patches, managedByPatch)

	// Ensure metadata.labels exists
	if env.Labels == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels",
			"value": map[string]string{},
		})
	}

	// Add created-by label to metadata with the actual username
	createdByPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1created-by",
		"value": userName,
	}
	patches = append(patches, createdByPatch)

	// Also add to spec.labels for namespace labeling
	ownerPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/labels/kloudlite.io~1owned-by",
		"value": userName,
	}
	patches = append(patches, ownerPatch)

	// Add targetNamespace label for easy validation and lookup
	if env.Spec.TargetNamespace != "" {
		targetNamespacePatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/kloudlite.io~1target-namespace",
			"value": env.Spec.TargetNamespace,
		}
		patches = append(patches, targetNamespacePatch)
	}

	// Add WorkMachine ownership label
	if env.Spec.WorkMachineName != "" {
		workMachineNamePatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/kloudlite.io~1workmachine-name",
			"value": env.Spec.WorkMachineName,
		}
		patches = append(patches, workMachineNamePatch)
	}

	// Add activated label for efficient filtering
	activatedValue := "false"
	if env.Spec.Activated {
		activatedValue = "true"
	}
	activatedPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1activated",
		"value": activatedValue,
	}
	patches = append(patches, activatedPatch)

	// Add default annotations if not present
	if env.Spec.Annotations == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/annotations",
			"value": map[string]string{},
		})
	}

	// Add creation timestamp annotation
	if req.Operation == admissionv1.Create {
		timestampPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/annotations/kloudlite.io~1created-at",
			"value": metav1.Now().Format("2006-01-02T15:04:05Z07:00"),
		}
		patches = append(patches, timestampPatch)
	}

	// Set default resource quotas if not specified and environment is activated
	if env.Spec.Activated && env.Spec.ResourceQuotas == nil {
		patches = append(patches, map[string]interface{}{
			"op":   "add",
			"path": "/spec/resourceQuotas",
			"value": map[string]string{
				"limits.cpu":             "10",
				"limits.memory":          "10Gi",
				"requests.cpu":           "5",
				"requests.memory":        "5Gi",
				"persistentvolumeclaims": "10",
			},
		})
	}

	// Convert patches to JSON
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		w.logger.Error("Failed to marshal patches: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to create patches",
			},
		}
	}

	// Return response with patches
	patchType := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
}

func (w *EnvironmentWebhook) validateEnvironment(env *environmentsv1.Environment, operation admissionv1.Operation) error {
	ctx := context.Background()

	// Validate environment name (metadata.name) is unique within the namespace
	// Since environments are now namespace-scoped, uniqueness is handled by Kubernetes
	// We just need to validate the name format
	if env.Name == "" {
		return fmt.Errorf("environment name is required")
	}

	// Validate environment namespace is a WorkMachine namespace
	if operation == admissionv1.Create {
		// Check that the namespace is a WorkMachine namespace (starts with wm-)
		if !strings.HasPrefix(env.Namespace, "wm-") {
			return fmt.Errorf("environments must be created in a WorkMachine namespace (wm-*). Got namespace: %s", env.Namespace)
		}

		// Derive WorkMachineName from namespace if not provided
		if env.Spec.WorkMachineName == "" {
			// Derive from namespace: wm-{username} namespace belongs to workmachine owned by {username}
			// The workmachine name is the owner's username
			ns := &corev1.Namespace{}
			if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Namespace}, ns); err != nil {
				return fmt.Errorf("failed to get namespace %s: %v", env.Namespace, err)
			}

			// Check if namespace is a WorkMachine namespace
			if ns.Labels == nil || ns.Labels["kloudlite.io/workmachine"] != "true" {
				return fmt.Errorf("namespace %s is not a WorkMachine namespace", env.Namespace)
			}
		} else {
			// Verify the WorkMachine exists
			var workMachine machinesv1.WorkMachine
			if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Spec.WorkMachineName}, &workMachine); err != nil {
				return fmt.Errorf("referenced WorkMachine '%s' does not exist", env.Spec.WorkMachineName)
			}

			// Verify the WorkMachine owns this namespace
			if workMachine.Spec.TargetNamespace != env.Namespace {
				return fmt.Errorf("environment namespace %s does not match WorkMachine %s's namespace %s",
					env.Namespace, env.Spec.WorkMachineName, workMachine.Spec.TargetNamespace)
			}
		}
	}

	// Prevent activating environment when WorkMachine is stopped
	if env.Spec.Activated && env.Spec.WorkMachineName != "" {
		var workMachine machinesv1.WorkMachine
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Spec.WorkMachineName}, &workMachine); err == nil {
			if workMachine.Spec.State == "stopped" || workMachine.Spec.State == "disabled" {
				return fmt.Errorf("cannot activate environment: WorkMachine '%s' is in '%s' state. Please start the WorkMachine first", env.Spec.WorkMachineName, workMachine.Spec.State)
			}
			// Also check runtime status
			if workMachine.Status.State == machinesv1.MachineStateStopped || workMachine.Status.State == machinesv1.MachineStateStopping {
				return fmt.Errorf("cannot activate environment: WorkMachine '%s' is currently %s. Please wait for the WorkMachine to be running", env.Spec.WorkMachineName, workMachine.Status.State)
			}
		}
	}

	// Validate snapshot exists and is ready when fromSnapshot is set
	if env.Spec.FromSnapshot != nil && operation == admissionv1.Create {
		// Fetch the snapshot to validate it exists and is ready
		var snapshot snapshotv1.Snapshot
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Spec.FromSnapshot.SnapshotName, Namespace: env.Spec.FromSnapshot.SourceNamespace}, &snapshot); err != nil {
			return fmt.Errorf("snapshot '%s' not found in namespace '%s'", env.Spec.FromSnapshot.SnapshotName, env.Spec.FromSnapshot.SourceNamespace)
		}

		// Validate snapshot is ready
		if snapshot.Status.State != snapshotv1.SnapshotStateReady {
			return fmt.Errorf("snapshot '%s' is not ready (current state: %s). Only ready snapshots can be used to create environments", env.Spec.FromSnapshot.SnapshotName, snapshot.Status.State)
		}

		// Validate snapshot is pushed to registry
		if snapshot.Status.Registry == nil || snapshot.Status.Registry.ImageRef == "" {
			return fmt.Errorf("snapshot '%s' is not pushed to registry. Only pushed snapshots can be used to create environments", env.Spec.FromSnapshot.SnapshotName)
		}
	}

	// Validate targetNamespace is unique across Environments and not used by WorkMachines
	if env.Spec.TargetNamespace != "" && (operation == admissionv1.Create || operation == admissionv1.Update) {
		// Check if any other Environment is using this targetNamespace (using label selector)
		environmentList := &environmentsv1.EnvironmentList{}
		if err := w.k8sClient.List(ctx, environmentList, client.MatchingLabels{
			"kloudlite.io/target-namespace": env.Spec.TargetNamespace,
		}); err != nil {
			return fmt.Errorf("failed to list environments: %v", err)
		}

		for _, existingEnv := range environmentList.Items {
			// Skip the current environment being created/updated
			if existingEnv.Name == env.Name {
				continue
			}

			return fmt.Errorf("targetNamespace '%s' is already used by Environment '%s'. Each Environment must have a unique targetNamespace",
				env.Spec.TargetNamespace, existingEnv.Name)
		}

		// Check if any WorkMachine is using this namespace (using label selector)
		workMachineList := &machinesv1.WorkMachineList{}
		if err := w.k8sClient.List(ctx, workMachineList, client.MatchingLabels{
			"kloudlite.io/target-namespace": env.Spec.TargetNamespace,
		}); err != nil {
			return fmt.Errorf("failed to list workmachines: %v", err)
		}

		if len(workMachineList.Items) > 0 {
			return fmt.Errorf("targetNamespace '%s' is already used by WorkMachine '%s'. Environment cannot use a namespace owned by a WorkMachine",
				env.Spec.TargetNamespace, workMachineList.Items[0].Name)
		}
	}

	// Validate namespace name
	if err := w.validateNamespaceName(env.Spec.TargetNamespace); err != nil {
		return fmt.Errorf("invalid target namespace: %w", err)
	}

	// Check if namespace already exists - reject if it does
	if operation == admissionv1.Create {
		ctx := context.Background()
		ns := &corev1.Namespace{}
		err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Spec.TargetNamespace}, ns)
		if err == nil {
			// Namespace exists - this is not allowed for new environments
			// The mutation webhook will have set/generated targetNamespace, and controller will create it
			if ns.Labels != nil {
				if existingEnv, ok := ns.Labels["kloudlite.io/environment"]; ok {
					if existingEnv != env.Name {
						return fmt.Errorf("namespace %s is already managed by environment %s", env.Spec.TargetNamespace, existingEnv)
					}
					// Same environment name - this might be a recreation, allow it
				} else {
					// Namespace exists but not managed by any environment
					return fmt.Errorf("namespace %s already exists and is not managed by Kloudlite", env.Spec.TargetNamespace)
				}
			} else {
				// Namespace exists without labels - not managed by Kloudlite
				return fmt.Errorf("namespace %s already exists and is not managed by Kloudlite", env.Spec.TargetNamespace)
			}
		}
	}

	// Validate resource quotas if specified
	if env.Spec.ResourceQuotas != nil {
		if err := w.validateResourceQuotas(env.Spec.ResourceQuotas); err != nil {
			return fmt.Errorf("invalid resource quotas: %w", err)
		}
	}

	// Validate network policies if specified
	if env.Spec.NetworkPolicies != nil {
		if err := w.validateNetworkPolicies(env.Spec.NetworkPolicies); err != nil {
			return fmt.Errorf("invalid network policies: %w", err)
		}
	}

	// For deletion operations, fetch the current environment to check status
	if operation == admissionv1.Delete {
		// Fetch current environment to check restore status (namespaced lookup)
		var currentEnv environmentsv1.Environment
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Namespace: env.Namespace, Name: env.Name}, &currentEnv); err == nil {
			// Check if environment is being restored from snapshot
			if currentEnv.Status.SnapshotRestoreStatus != nil {
				phase := currentEnv.Status.SnapshotRestoreStatus.Phase
				if phase != environmentsv1.SnapshotRestorePhaseCompleted && phase != environmentsv1.SnapshotRestorePhaseFailed {
					return fmt.Errorf("cannot delete environment during snapshot restore. Current phase: %s. Please wait for restore to complete or fail", phase)
				}
			}
		}
	}

	// targetNamespace uniqueness is ensured by the random suffix generation
	// No need for global uniqueness check since each environment gets a unique random suffix

	return nil
}

func (w *EnvironmentWebhook) validateNamespaceName(name string) error {
	if name == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}

	// Check length
	if len(name) > 63 {
		return fmt.Errorf("namespace name must be no more than 63 characters")
	}

	// Check for valid DNS-1123 label
	dnsLabelRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !dnsLabelRegex.MatchString(name) {
		return fmt.Errorf("namespace name must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character")
	}

	// Check for reserved namespaces
	reservedNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"default",
		"kloudlite-system",
	}

	for _, reserved := range reservedNamespaces {
		if name == reserved {
			return fmt.Errorf("cannot use reserved namespace name: %s", reserved)
		}
	}

	// Check for reserved prefixes
	reservedPrefixes := []string{
		"kube-",
		"openshift-",
		"kubernetes-",
	}

	for _, prefix := range reservedPrefixes {
		if strings.HasPrefix(name, prefix) {
			return fmt.Errorf("cannot use namespace name with reserved prefix: %s", prefix)
		}
	}

	return nil
}

func (w *EnvironmentWebhook) validateResourceQuotas(quotas *environmentsv1.ResourceQuotas) error {
	// Validate CPU limits
	if quotas.LimitsCPU != "" {
		if _, err := parseQuantity(quotas.LimitsCPU); err != nil {
			return fmt.Errorf("invalid CPU limit: %w", err)
		}
	}

	// Validate memory limits
	if quotas.LimitsMemory != "" {
		if _, err := parseQuantity(quotas.LimitsMemory); err != nil {
			return fmt.Errorf("invalid memory limit: %w", err)
		}
	}

	// Validate CPU requests
	if quotas.RequestsCPU != "" {
		if _, err := parseQuantity(quotas.RequestsCPU); err != nil {
			return fmt.Errorf("invalid CPU request: %w", err)
		}
	}

	// Validate memory requests
	if quotas.RequestsMemory != "" {
		if _, err := parseQuantity(quotas.RequestsMemory); err != nil {
			return fmt.Errorf("invalid memory request: %w", err)
		}
	}

	// Validate PVC count
	if quotas.PersistentVolumeClaims != "" {
		if _, err := parseQuantity(quotas.PersistentVolumeClaims); err != nil {
			return fmt.Errorf("invalid PVC count: %w", err)
		}
	}

	return nil
}

func (w *EnvironmentWebhook) validateNetworkPolicies(policies *environmentsv1.NetworkPolicies) error {
	// Validate allowed namespaces
	for _, ns := range policies.AllowedNamespaces {
		if err := w.validateNamespaceName(ns); err != nil {
			return fmt.Errorf("invalid allowed namespace %s: %w", ns, err)
		}
	}

	// Validate ingress rules
	for i, rule := range policies.IngressRules {
		if len(rule.From) == 0 && len(rule.Ports) == 0 {
			return fmt.Errorf("ingress rule %d must have at least one 'from' or 'ports' specification", i)
		}

		// Validate ports
		for j, port := range rule.Ports {
			if port.Port < 1 || port.Port > 65535 {
				return fmt.Errorf("invalid port number %d in ingress rule %d, port %d", port.Port, i, j)
			}

			if port.Protocol != "" && port.Protocol != "TCP" && port.Protocol != "UDP" {
				return fmt.Errorf("invalid protocol %s in ingress rule %d, port %d (must be TCP or UDP)", port.Protocol, i, j)
			}
		}
	}

	return nil
}

// parseQuantity is a helper function to validate quantity strings
func parseQuantity(quantity string) (int64, error) {
	// Simple validation for common quantity formats
	// This is a simplified version - in production, use k8s.io/apimachinery/pkg/api/resource.ParseQuantity

	// Check for numeric value
	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if numericRegex.MatchString(quantity) {
		return 0, nil
	}

	// Check for CPU units (m for millicores)
	cpuRegex := regexp.MustCompile(`^[0-9]+m?$`)
	if cpuRegex.MatchString(quantity) {
		return 0, nil
	}

	// Check for memory units (Ki, Mi, Gi, Ti)
	memoryRegex := regexp.MustCompile(`^[0-9]+([KMGT]i)?$`)
	if memoryRegex.MatchString(quantity) {
		return 0, nil
	}

	return 0, fmt.Errorf("invalid quantity format: %s", quantity)
}

// generateRandomSuffix generates a random alphanumeric suffix of the specified length
func generateRandomSuffix(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to a fixed character if random fails
			b[i] = 'x'
		} else {
			b[i] = charset[n.Int64()]
		}
	}
	return string(b)
}
