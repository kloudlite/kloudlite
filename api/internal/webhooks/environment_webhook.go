package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
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

	// Use the OwnedBy field from the spec to determine ownership first
	// This is needed for generating the targetNamespace with username prefix
	ownedBy := env.Spec.OwnedBy
	var userName string
	var userEmail string

	// Determine if OwnedBy is an email or username
	if strings.Contains(ownedBy, "@") {
		// OwnedBy is an email
		userEmail = ownedBy
		// Find the actual user name
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
		// OwnedBy is a username
		userName = ownedBy
		// Look up the email
		var user platformv1alpha1.User
		if err := w.k8sClient.Get(context.Background(), client.ObjectKey{Name: userName}, &user); err == nil {
			userEmail = user.Spec.Email
		}
	}

	// Generate targetNamespace if not provided
	// Use username prefix to avoid conflicts between users
	if env.Spec.TargetNamespace == "" {
		// Extract the environment name from the full name (which is {username}--{envname})
		// The env.Name at this point is already prefixed by the handler
		targetNamespace := fmt.Sprintf("env-%s", env.Name)
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

	// Add sanitized email as a label in both metadata and spec
	if userEmail != "" {
		// Sanitize email for use as Kubernetes label value
		sanitizedEmail := utils.SanitizeForLabel(userEmail)

		// Metadata label
		metadataEmailPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/kloudlite.io~1owner-email",
			"value": sanitizedEmail,
		}
		patches = append(patches, metadataEmailPatch)

		// Spec label for namespace
		emailPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/labels/kloudlite.io~1owner-email",
			"value": sanitizedEmail,
		}
		patches = append(patches, emailPatch)
	}

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

	// Validate that WorkMachine reference exists (only for CREATE operations)
	if operation == admissionv1.Create {
		if env.Spec.WorkMachineName == "" {
			return fmt.Errorf("WorkMachineName is required. Environment must be associated with a WorkMachine")
		}

		// Check if the referenced WorkMachine exists
		var workMachine machinesv1.WorkMachine
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Spec.WorkMachineName}, &workMachine); err != nil {
			return fmt.Errorf("referenced WorkMachine '%s' does not exist. Please create the WorkMachine first or provide a valid WorkMachine reference", env.Spec.WorkMachineName)
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

	// Prevent cloning from environment whose WorkMachine is stopped
	if env.Spec.CloneFrom != "" && operation == admissionv1.Create {
		// Fetch the source environment
		var sourceEnv environmentsv1.Environment
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Spec.CloneFrom}, &sourceEnv); err != nil {
			return fmt.Errorf("source environment '%s' not found for cloning", env.Spec.CloneFrom)
		}

		// Check if source environment's WorkMachine is running
		if sourceEnv.Spec.WorkMachineName != "" {
			var sourceWorkMachine machinesv1.WorkMachine
			if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: sourceEnv.Spec.WorkMachineName}, &sourceWorkMachine); err == nil {
				// Check spec state
				if sourceWorkMachine.Spec.State == "stopped" || sourceWorkMachine.Spec.State == "disabled" {
					return fmt.Errorf("cannot clone from environment '%s': its WorkMachine '%s' is in '%s' state. Please start the WorkMachine first", env.Spec.CloneFrom, sourceEnv.Spec.WorkMachineName, sourceWorkMachine.Spec.State)
				}
				// Check runtime status
				if sourceWorkMachine.Status.State == machinesv1.MachineStateStopped || sourceWorkMachine.Status.State == machinesv1.MachineStateStopping {
					return fmt.Errorf("cannot clone from environment '%s': its WorkMachine '%s' is currently %s. Please start the WorkMachine and wait for it to be running", env.Spec.CloneFrom, sourceEnv.Spec.WorkMachineName, sourceWorkMachine.Status.State)
				}
			}
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
		// Prevent deletion of activated environments
		if env.Spec.Activated {
			return fmt.Errorf("cannot delete an activated environment, please deactivate it first")
		}

		// Fetch current environment to check cloning status
		var currentEnv environmentsv1.Environment
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: env.Name}, &currentEnv); err == nil {
			// Check if environment is being cloned TO
			if currentEnv.Status.CloningStatus != nil {
				phase := currentEnv.Status.CloningStatus.Phase
				if phase != "Completed" && phase != "Failed" {
					return fmt.Errorf("cannot delete environment during cloning. Current phase: %s. Please wait for cloning to complete or fail", phase)
				}
			}

			// Check if environment is being cloned FROM (used as source)
			if currentEnv.Status.SourceCloningStatus != nil {
				return fmt.Errorf("cannot delete environment while it's being used as cloning source for: %s. Please wait for cloning to complete", currentEnv.Status.SourceCloningStatus.TargetEnvironmentName)
			}
		}
	}

	// Check for conflicting environments with same namespace
	if operation == admissionv1.Create || operation == admissionv1.Update {
		ctx := context.Background()
		envList := &environmentsv1.EnvironmentList{}
		if err := w.k8sClient.List(ctx, envList); err == nil {
			for _, existingEnv := range envList.Items {
				if existingEnv.Name != env.Name && existingEnv.Spec.TargetNamespace == env.Spec.TargetNamespace {
					return fmt.Errorf("target namespace %s is already used by environment %s", env.Spec.TargetNamespace, existingEnv.Name)
				}
			}
		}
	}

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
