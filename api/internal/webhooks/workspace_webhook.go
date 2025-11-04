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
	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkspaceWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
}

func NewWorkspaceWebhook(logger logger.Logger, k8sClient client.Client) *WorkspaceWebhook {
	return &WorkspaceWebhook{
		logger:    logger,
		k8sClient: k8sClient,
	}
}

// ValidateWorkspace handles validation webhook for Workspace CRD
func (w *WorkspaceWebhook) ValidateWorkspace(c *gin.Context) {
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

// MutateWorkspace handles mutation webhook for Workspace CRD
func (w *WorkspaceWebhook) MutateWorkspace(c *gin.Context) {
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

func (w *WorkspaceWebhook) handleValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the workspace object
	var workspace workspacesv1.Workspace
	if err := json.Unmarshal(req.Object.Raw, &workspace); err != nil {
		w.logger.Error("Failed to unmarshal workspace: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal workspace object",
			},
		}
	}

	// Perform validation
	if err := w.validateWorkspace(&workspace, req.Operation); err != nil {
		w.logger.Warn("Workspace validation failed: " + err.Error())
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

func (w *WorkspaceWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the workspace object
	var workspace workspacesv1.Workspace
	if err := json.Unmarshal(req.Object.Raw, &workspace); err != nil {
		w.logger.Error("Failed to unmarshal workspace: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal workspace object",
			},
		}
	}

	// Create patches for mutations
	var patches []map[string]interface{}

	// Set default status if not specified
	if workspace.Spec.Status == "" {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/status",
			"value": "active",
		})
	}

	// Set default storage size if not specified
	if workspace.Spec.StorageSize == "" {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/storageSize",
			"value": "10Gi",
		})
	}

	// Set default workspace path if not specified
	if workspace.Spec.WorkspacePath == "" {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/workspacePath",
			"value": "/workspace",
		})
	}

	// Set default VS Code version if not specified
	if workspace.Spec.VSCodeVersion == "" {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/vscodeVersion",
			"value": "latest",
		})
	}

	// Ensure labels map exists
	if workspace.Labels == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels",
			"value": map[string]string{},
		})
	}

	// Add workspace-name label
	workspaceNameLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1workspace-name",
		"value": workspace.Name,
	}
	patches = append(patches, workspaceNameLabelPatch)

	// Look up user to get email and user ID
	owner := workspace.Spec.Owner
	var userName string
	var userEmail string

	ctx := context.Background()
	if strings.Contains(owner, "@") {
		// Owner is an email, find the user
		userEmail = owner
		userList := &platformv1alpha1.UserList{}
		if err := w.k8sClient.List(ctx, userList); err == nil {
			for _, u := range userList.Items {
				if u.Spec.Email == owner {
					userName = u.Name
					break
				}
			}
		}
		if userName == "" {
			// If no user found, use sanitized email as username
			userName = strings.ReplaceAll(strings.Split(owner, "@")[0], ".", "-")
		}
	} else {
		// Owner is a username
		userName = owner
		var user platformv1alpha1.User
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: userName}, &user); err == nil {
			userEmail = user.Spec.Email
		}
	}

	// Add owned-by label with username
	ownerLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1owned-by",
		"value": userName,
	}
	patches = append(patches, ownerLabelPatch)

	// Add base64 encoded email as a label
	if userEmail != "" {
		encodedEmail := fn.LabelValueEncoder(userEmail)
		emailLabelPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/kloudlite.io~1owner-email",
			"value": encodedEmail,
		}
		patches = append(patches, emailLabelPatch)
	}

	// Ensure finalizers array exists
	if workspace.Finalizers == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/finalizers",
			"value": []string{},
		})
	}

	// Add finalizer for cleanup
	if req.Operation == admissionv1.Create {
		finalizerExists := false
		for _, f := range workspace.Finalizers {
			if f == "workspaces.kloudlite.io/finalizer" {
				finalizerExists = true
				break
			}
		}

		if !finalizerExists {
			// Append finalizer to the array
			newFinalizers := append(workspace.Finalizers, "workspaces.kloudlite.io/finalizer")
			patches = append(patches, map[string]interface{}{
				"op":    "replace",
				"path":  "/metadata/finalizers",
				"value": newFinalizers,
			})
		}
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

func (w *WorkspaceWebhook) validateWorkspace(workspace *workspacesv1.Workspace, operation admissionv1.Operation) error {
	ctx := context.Background()

	// Validate that owner exists
	owner := workspace.Spec.Owner
	if owner == "" {
		return fmt.Errorf("workspace owner is required")
	}

	// Check if owner is an email or username and verify user exists
	var foundUser *platformv1alpha1.User
	if strings.Contains(owner, "@") {
		// Check by email
		userList := &platformv1alpha1.UserList{}
		if err := w.k8sClient.List(ctx, userList); err != nil {
			return fmt.Errorf("failed to list users: %v", err)
		}

		for _, user := range userList.Items {
			if user.Spec.Email == owner {
				foundUser = &user
				break
			}
		}
	} else {
		// Check by username
		user := &platformv1alpha1.User{}
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: owner}, user); err == nil {
			foundUser = user
		}
	}

	if foundUser == nil {
		return fmt.Errorf("owner %s does not exist", owner)
	}

	// Validate that the user has a WorkMachine
	workMachineList := &machinesv1.WorkMachineList{}
	if err := w.k8sClient.List(ctx, workMachineList); err != nil {
		return fmt.Errorf("failed to list work machines: %v", err)
	}

	hasWorkMachine := false
	for _, wm := range workMachineList.Items {
		if wm.Spec.OwnedBy == owner || wm.Spec.OwnedBy == foundUser.Spec.Email || wm.Spec.OwnedBy == foundUser.Name {
			hasWorkMachine = true
			break
		}
	}

	if !hasWorkMachine {
		return fmt.Errorf("user %s does not have a WorkMachine. Please create a WorkMachine first", owner)
	}

	// Validate displayName
	if workspace.Spec.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}

	if len(workspace.Spec.DisplayName) > 100 {
		return fmt.Errorf("displayName must be no more than 100 characters")
	}

	// Validate description length
	if len(workspace.Spec.Description) > 500 {
		return fmt.Errorf("description must be no more than 500 characters")
	}

	// Validate status enum
	if workspace.Spec.Status != "" {
		validStatuses := map[string]bool{
			"active":    true,
			"suspended": true,
			"archived":  true,
		}
		if !validStatuses[workspace.Spec.Status] {
			return fmt.Errorf("invalid status: %s. Must be one of: active, suspended, archived", workspace.Spec.Status)
		}
	}

	// Validate workspace name format (DNS-1123 label)
	if workspace.Name != "" {
		if err := validateDNS1123Label(workspace.Name); err != nil {
			return fmt.Errorf("invalid workspace name: %w", err)
		}
	}

	// Validate resource quota if specified
	if workspace.Spec.ResourceQuota != nil {
		if err := validateResourceQuota(workspace.Spec.ResourceQuota); err != nil {
			return fmt.Errorf("invalid resource quota: %w", err)
		}
	}

	// Validate settings if specified
	if workspace.Spec.Settings != nil {
		if err := validateWorkspaceSettings(workspace.Spec.Settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
	}

	// Validate service intercepts - ensure no service is intercepted by multiple workspaces
	if workspace.Spec.EnvironmentConnection != nil && len(workspace.Spec.EnvironmentConnection.Intercepts) > 0 {
		if err := w.validateServiceIntercepts(ctx, workspace); err != nil {
			return fmt.Errorf("invalid service intercepts: %w", err)
		}
	}

	// Validate state transitions on UPDATE
	if operation == admissionv1.Update {
		// For updates, we would need the old object to validate transitions
		// This is left as a future enhancement
	}

	// Prevent deletion of active workspaces
	if operation == admissionv1.Delete && workspace.Spec.Status == "active" {
		return fmt.Errorf("cannot delete an active workspace. Please suspend or archive it first")
	}

	return nil
}

// validateDNS1123Label validates that a name is a valid DNS-1123 label
func validateDNS1123Label(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("name must be no more than 63 characters")
	}

	dnsLabelRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !dnsLabelRegex.MatchString(name) {
		return fmt.Errorf("name must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character")
	}

	return nil
}

// validateResourceQuota validates resource quota values
func validateResourceQuota(quota *workspacesv1.ResourceQuota) error {
	// Validate CPU format
	if quota.CPU != "" {
		cpuRegex := regexp.MustCompile(`^[0-9]+(\.[0-9]+)?m?$`)
		if !cpuRegex.MatchString(quota.CPU) {
			return fmt.Errorf("invalid CPU format: %s", quota.CPU)
		}
	}

	// Validate memory format
	if quota.Memory != "" {
		memoryRegex := regexp.MustCompile(`^[0-9]+([KMGT]i)?$`)
		if !memoryRegex.MatchString(quota.Memory) {
			return fmt.Errorf("invalid memory format: %s", quota.Memory)
		}
	}

	// Validate storage format
	if quota.Storage != "" {
		storageRegex := regexp.MustCompile(`^[0-9]+([KMGT]i)?$`)
		if !storageRegex.MatchString(quota.Storage) {
			return fmt.Errorf("invalid storage format: %s", quota.Storage)
		}
	}

	// Validate GPU count
	if quota.GPUs < 0 || quota.GPUs > 8 {
		return fmt.Errorf("GPUs must be between 0 and 8")
	}

	return nil
}

// validateWorkspaceSettings validates workspace settings
func validateWorkspaceSettings(settings *workspacesv1.WorkspaceSettings) error {
	// Validate idle timeout
	if settings.IdleTimeout < 0 || settings.IdleTimeout > 10080 {
		return fmt.Errorf("idleTimeout must be between 0 and 10080 minutes (7 days)")
	}

	// Validate max runtime
	if settings.MaxRuntime < 0 || settings.MaxRuntime > 43200 {
		return fmt.Errorf("maxRuntime must be between 0 and 43200 minutes (30 days)")
	}

	// Validate dotfiles repo URL if specified
	if settings.DotfilesRepo != "" {
		// Basic URL validation
		if !strings.HasPrefix(settings.DotfilesRepo, "http://") &&
			!strings.HasPrefix(settings.DotfilesRepo, "https://") &&
			!strings.HasPrefix(settings.DotfilesRepo, "git@") {
			return fmt.Errorf("dotfilesRepo must be a valid git repository URL")
		}
	}

	// Validate git config if specified
	if settings.GitConfig != nil {
		if settings.GitConfig.UserEmail != "" && !strings.Contains(settings.GitConfig.UserEmail, "@") {
			return fmt.Errorf("gitConfig.userEmail must be a valid email address")
		}
	}

	return nil
}

// validateServiceIntercepts ensures no service is intercepted by more than one workspace
func (w *WorkspaceWebhook) validateServiceIntercepts(ctx context.Context, workspace *workspacesv1.Workspace) error {
	// Get the environment reference to determine the target namespace
	if workspace.Spec.EnvironmentConnection == nil {
		return nil // No environment connection, no intercepts to validate
	}

	envRef := workspace.Spec.EnvironmentConnection.EnvironmentRef

	// Fetch the Environment to get its target namespace
	env := &environmentv1.Environment{}
	if err := w.k8sClient.Get(ctx, client.ObjectKey{
		Name:      envRef.Name,
		Namespace: envRef.Namespace,
	}, env); err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", envRef.Name, err)
	}

	targetNamespace := env.Spec.TargetNamespace

	// List all existing ServiceIntercept CRs in the workspace namespace
	interceptList := &interceptsv1.ServiceInterceptList{}
	if err := w.k8sClient.List(ctx, interceptList, client.InNamespace(workspace.Namespace)); err != nil {
		return fmt.Errorf("failed to list existing service intercepts: %w", err)
	}

	// Build a map of service name+namespace -> workspace name for existing intercepts
	existingIntercepts := make(map[string]string)
	for _, intercept := range interceptList.Items {
		// Skip intercepts owned by the current workspace (for updates)
		if intercept.Spec.WorkspaceRef.Name == workspace.Name {
			continue
		}

		// Create a key combining service name and namespace
		serviceKey := fmt.Sprintf("%s/%s", intercept.Spec.ServiceRef.Namespace, intercept.Spec.ServiceRef.Name)
		existingIntercepts[serviceKey] = intercept.Spec.WorkspaceRef.Name
	}

	// Check each intercept in the workspace spec
	for _, interceptSpec := range workspace.Spec.EnvironmentConnection.Intercepts {
		// Construct the service key for this intercept
		// The service will be in the environment's target namespace
		serviceKey := fmt.Sprintf("%s/%s", targetNamespace, interceptSpec.ServiceName)

		// Check if this service is already intercepted by another workspace
		if conflictingWorkspace, exists := existingIntercepts[serviceKey]; exists {
			return fmt.Errorf("service '%s' in namespace '%s' is already being intercepted by workspace '%s'. A service can only be intercepted by one workspace at a time",
				interceptSpec.ServiceName, targetNamespace, conflictingWorkspace)
		}
	}

	return nil
}
