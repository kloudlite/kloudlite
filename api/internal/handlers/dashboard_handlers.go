package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DashboardHandlers handles HTTP requests for the dashboard
type DashboardHandlers struct {
	machineTypeRepo     repository.MachineTypeRepository
	workMachineRepo     repository.WorkMachineRepository
	userPreferencesRepo repository.UserPreferencesRepository
	workspaceRepo       repository.WorkspaceRepository
	environmentRepo     repository.EnvironmentRepository
	compositionRepo     repository.CompositionRepository
	k8sClient           client.Client
	logger              *zap.Logger
}

// NewDashboardHandlers creates a new DashboardHandlers
func NewDashboardHandlers(
	machineTypeRepo repository.MachineTypeRepository,
	workMachineRepo repository.WorkMachineRepository,
	userPreferencesRepo repository.UserPreferencesRepository,
	workspaceRepo repository.WorkspaceRepository,
	environmentRepo repository.EnvironmentRepository,
	compositionRepo repository.CompositionRepository,
	k8sClient client.Client,
	logger *zap.Logger,
) *DashboardHandlers {
	return &DashboardHandlers{
		machineTypeRepo:     machineTypeRepo,
		workMachineRepo:     workMachineRepo,
		userPreferencesRepo: userPreferencesRepo,
		workspaceRepo:       workspaceRepo,
		environmentRepo:     environmentRepo,
		compositionRepo:     compositionRepo,
		k8sClient:           k8sClient,
		logger:              logger,
	}
}

// DashboardResponse represents the dashboard data response
type DashboardResponse struct {
	MachineTypes       []machinesv1.MachineType          `json:"machineTypes"`
	WorkMachines       []machinesv1.WorkMachine          `json:"workMachines"`
	Preferences        *platformv1alpha1.UserPreferences `json:"preferences,omitempty"`
	PinnedWorkspaces   []workspacesv1.Workspace          `json:"pinnedWorkspaces"`
	PinnedEnvironments []environmentsv1.Environment      `json:"pinnedEnvironments"`
	IsAdmin            bool                              `json:"isAdmin"`
}

// GetDashboard handles GET /dashboard
// Returns all data needed for the homepage in a single request
func (h *DashboardHandlers) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current user from context
	username, _, roles, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if user is admin
	isAdmin := h.hasRole(roles, platformv1alpha1.RoleAdmin) || h.hasRole(roles, platformv1alpha1.RoleSuperAdmin)

	// Response data
	var response DashboardResponse
	response.IsAdmin = isAdmin

	// Use WaitGroup for parallel fetching
	var wg sync.WaitGroup
	var mu sync.Mutex
	var fetchErrors []error

	// Fetch machine types
	wg.Add(1)
	go func() {
		defer wg.Done()
		machineTypes, err := h.machineTypeRepo.ListActive(ctx)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			h.logger.Error("Failed to fetch machine types", zap.Error(err))
			fetchErrors = append(fetchErrors, err)
			return
		}
		response.MachineTypes = machineTypes.Items
	}()

	// Fetch work machines (all if admin, or just user's own)
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		defer mu.Unlock()

		if isAdmin {
			workMachines, err := h.workMachineRepo.List(ctx)
			if err != nil {
				h.logger.Error("Failed to fetch all work machines", zap.Error(err))
				fetchErrors = append(fetchErrors, err)
				return
			}
			response.WorkMachines = workMachines.Items
		} else {
			workMachine, err := h.workMachineRepo.GetByOwner(ctx, username)
			if err != nil {
				if client.IgnoreNotFound(err) != nil {
					h.logger.Error("Failed to fetch user's work machine", zap.Error(err), zap.String("username", username))
					fetchErrors = append(fetchErrors, err)
				}
				// Not found is ok - user may not have a work machine yet
				return
			}
			response.WorkMachines = []machinesv1.WorkMachine{*workMachine}
		}
	}()

	// Fetch user preferences
	wg.Add(1)
	go func() {
		defer wg.Done()
		prefs, err := h.userPreferencesRepo.GetOrCreate(ctx, username)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			h.logger.Error("Failed to fetch user preferences", zap.Error(err), zap.String("username", username))
			fetchErrors = append(fetchErrors, err)
			return
		}
		response.Preferences = prefs
	}()

	// Wait for initial data to be fetched
	wg.Wait()

	// If we couldn't fetch preferences, we can't fetch pinned resources
	if response.Preferences == nil {
		c.JSON(http.StatusOK, response)
		return
	}

	// Fetch pinned workspaces in parallel
	pinnedWsRefs := response.Preferences.Spec.PinnedWorkspaces
	if len(pinnedWsRefs) > 0 {
		var wsWg sync.WaitGroup
		pinnedWorkspaces := make([]workspacesv1.Workspace, 0, len(pinnedWsRefs))
		var wsMu sync.Mutex

		for _, wsRef := range pinnedWsRefs {
			wsWg.Add(1)
			go func(ref platformv1alpha1.ResourceReference) {
				defer wsWg.Done()
				ws, err := h.workspaceRepo.Get(ctx, ref.Namespace, ref.Name)
				if err != nil {
					// Workspace may have been deleted - skip silently
					if client.IgnoreNotFound(err) != nil {
						h.logger.Warn("Failed to fetch pinned workspace",
							zap.Error(err),
							zap.String("namespace", ref.Namespace),
							zap.String("name", ref.Name),
						)
					}
					return
				}
				wsMu.Lock()
				pinnedWorkspaces = append(pinnedWorkspaces, *ws)
				wsMu.Unlock()
			}(wsRef)
		}
		wsWg.Wait()
		response.PinnedWorkspaces = pinnedWorkspaces
	}

	// Fetch pinned environments in parallel
	pinnedEnvNames := response.Preferences.Spec.PinnedEnvironments
	if len(pinnedEnvNames) > 0 {
		var envWg sync.WaitGroup
		pinnedEnvironments := make([]environmentsv1.Environment, 0, len(pinnedEnvNames))
		var envMu sync.Mutex

		for _, envName := range pinnedEnvNames {
			envWg.Add(1)
			go func(name string) {
				defer envWg.Done()
				env, err := h.environmentRepo.Get(ctx, name)
				if err != nil {
					// Environment may have been deleted - skip silently
					if client.IgnoreNotFound(err) != nil {
						h.logger.Warn("Failed to fetch pinned environment",
							zap.Error(err),
							zap.String("name", name),
						)
					}
					return
				}
				envMu.Lock()
				pinnedEnvironments = append(pinnedEnvironments, *env)
				envMu.Unlock()
			}(envName)
		}
		envWg.Wait()
		response.PinnedEnvironments = pinnedEnvironments
	}

	c.JSON(http.StatusOK, response)
}

// hasRole checks if the user has a specific role
func (h *DashboardHandlers) hasRole(roles []platformv1alpha1.RoleType, role platformv1alpha1.RoleType) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// userHasAccessToEnvironment checks if a user has access to view an environment
func (h *DashboardHandlers) userHasAccessToEnvironment(username string, env *environmentsv1.Environment) bool {
	// Owner always has access
	if env.Spec.OwnedBy == username {
		return true
	}

	visibility := env.Spec.Visibility
	if visibility == "" {
		visibility = "private"
	}

	switch visibility {
	case "private":
		return false
	case "shared":
		for _, sharedUser := range env.Spec.SharedWith {
			if sharedUser == username {
				return true
			}
		}
		return false
	case "open":
		return true
	default:
		return false
	}
}

// userHasAccessToWorkspace checks if a user has access to view a workspace
func (h *DashboardHandlers) userHasAccessToWorkspace(username string, ws *workspacesv1.Workspace) bool {
	// Owner always has access
	if ws.Spec.OwnedBy == username {
		return true
	}

	visibility := string(ws.Spec.Visibility)
	if visibility == "" {
		visibility = "private"
	}

	switch visibility {
	case "private":
		return false
	case "shared":
		for _, sharedUser := range ws.Spec.SharedWith {
			if sharedUser == username {
				return true
			}
		}
		return false
	case "open":
		return true
	default:
		return false
	}
}

// EnvironmentDetailsResponse represents the environment details response
type EnvironmentDetailsResponse struct {
	Environment *environmentsv1.Environment `json:"environment"`
	Services    []dto.ServiceInfo           `json:"services"`
	Composition *environmentsv1.Composition `json:"composition,omitempty"`
	Namespace   string                      `json:"namespace"`
	EnvHash     string                      `json:"envHash"`
	Subdomain   string                      `json:"subdomain"`
	IsActive    bool                        `json:"isActive"`
}

// GetEnvironmentDetails handles GET /environments/:name/details
// Returns environment, services, and composition in a single request
func (h *DashboardHandlers) GetEnvironmentDetails(c *gin.Context) {
	ctx := c.Request.Context()
	envName := c.Param("name")

	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment name is required"})
		return
	}

	// Get environment first to get the namespace
	env, err := h.environmentRepo.Get(ctx, envName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Environment not found"})
			return
		}
		h.logger.Error("Failed to get environment", zap.Error(err), zap.String("name", envName))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get environment", "details": err.Error()})
		return
	}

	namespace := env.Spec.TargetNamespace
	response := EnvironmentDetailsResponse{
		Environment: env,
		Namespace:   namespace,
		EnvHash:     env.Status.Hash,
		Subdomain:   env.Status.Subdomain,
		IsActive:    env.Status.State == environmentsv1.EnvironmentStateActive,
	}

	// Fetch services and composition in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Fetch services (deployments with kloudlite.io/managed=true label)
	wg.Add(1)
	go func() {
		defer wg.Done()
		services := h.listServicesInNamespace(ctx, namespace)
		mu.Lock()
		response.Services = services
		mu.Unlock()
	}()

	// Fetch main composition
	wg.Add(1)
	go func() {
		defer wg.Done()
		comp, err := h.compositionRepo.Get(ctx, namespace, "main-composition")
		if err != nil {
			// Composition may not exist yet - that's okay
			if client.IgnoreNotFound(err) != nil {
				h.logger.Warn("Failed to fetch composition", zap.Error(err), zap.String("namespace", namespace))
			}
			return
		}
		mu.Lock()
		response.Composition = comp
		mu.Unlock()
	}()

	wg.Wait()
	c.JSON(http.StatusOK, response)
}

// listServicesInNamespace fetches deployments and services in a namespace
func (h *DashboardHandlers) listServicesInNamespace(ctx context.Context, namespace string) []dto.ServiceInfo {
	// List all deployments managed by compositions
	deploymentList := &appsv1.DeploymentList{}
	if err := h.k8sClient.List(ctx, deploymentList,
		client.InNamespace(namespace),
		client.MatchingLabels{"kloudlite.io/managed": "true"},
	); err != nil {
		h.logger.Warn("Failed to list deployments", zap.String("namespace", namespace), zap.Error(err))
		return nil
	}

	// List all services in the namespace to enrich deployment data
	serviceList := &corev1.ServiceList{}
	if err := h.k8sClient.List(ctx, serviceList, client.InNamespace(namespace)); err != nil {
		h.logger.Warn("Failed to list k8s services", zap.String("namespace", namespace), zap.Error(err))
	}

	// Create a map of services by name for quick lookup
	serviceMap := make(map[string]*corev1.Service)
	for i := range serviceList.Items {
		svc := &serviceList.Items[i]
		serviceMap[svc.Name] = svc
	}

	// Transform the deployments to service info format
	services := make([]dto.ServiceInfo, 0, len(deploymentList.Items))
	for _, deploy := range deploymentList.Items {
		svc, hasSvc := serviceMap[deploy.Name]

		var ports []dto.ServicePort
		var clusterIP string
		var svcType string

		if hasSvc {
			ports = make([]dto.ServicePort, 0, len(svc.Spec.Ports))
			for _, port := range svc.Spec.Ports {
				ports = append(ports, dto.ServicePort{
					Name:       port.Name,
					Protocol:   string(port.Protocol),
					Port:       port.Port,
					TargetPort: port.TargetPort.String(),
				})
			}
			clusterIP = svc.Spec.ClusterIP
			svcType = string(svc.Spec.Type)
		} else {
			ports = make([]dto.ServicePort, 0)
			if len(deploy.Spec.Template.Spec.Containers) > 0 {
				for _, port := range deploy.Spec.Template.Spec.Containers[0].Ports {
					ports = append(ports, dto.ServicePort{
						Name:       port.Name,
						Protocol:   string(port.Protocol),
						Port:       port.ContainerPort,
						TargetPort: fmt.Sprintf("%d", port.ContainerPort),
					})
				}
			}
			svcType = "None"
		}

		image := ""
		if len(deploy.Spec.Template.Spec.Containers) > 0 {
			image = deploy.Spec.Template.Spec.Containers[0].Image
		}

		services = append(services, dto.ServiceInfo{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			Type:      svcType,
			ClusterIP: clusterIP,
			Ports:     ports,
			Selector:  deploy.Spec.Selector.MatchLabels,
			Replicas:  deploy.Status.ReadyReplicas,
			Image:     image,
		})
	}

	return services
}

// WorkspacesListResponse represents the workspaces list response
type WorkspacesListResponse struct {
	Workspaces         []workspacesv1.Workspace          `json:"workspaces"`
	WorkMachine        *machinesv1.WorkMachine           `json:"workMachine,omitempty"`
	Preferences        *platformv1alpha1.UserPreferences `json:"preferences,omitempty"`
	PinnedWorkspaceIds []string                          `json:"pinnedWorkspaceIds"`
	WorkMachineRunning bool                              `json:"workMachineRunning"`
}

// GetWorkspacesListFull handles GET /workspaces/list-full
// Returns workspaces, work machine, and preferences in a single request
func (h *DashboardHandlers) GetWorkspacesListFull(c *gin.Context) {
	ctx := c.Request.Context()

	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var response WorkspacesListResponse
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Fetch workspaces and filter by visibility
	wg.Add(1)
	go func() {
		defer wg.Done()
		workspaces, err := h.workspaceRepo.ListAll(ctx)
		if err != nil {
			h.logger.Error("Failed to list workspaces", zap.Error(err))
			return
		}
		// Filter by visibility access
		var accessibleWorkspaces []workspacesv1.Workspace
		for _, ws := range workspaces.Items {
			if h.userHasAccessToWorkspace(username, &ws) {
				accessibleWorkspaces = append(accessibleWorkspaces, ws)
			}
		}
		mu.Lock()
		response.Workspaces = accessibleWorkspaces
		mu.Unlock()
	}()

	// Fetch work machine
	wg.Add(1)
	go func() {
		defer wg.Done()
		wm, err := h.workMachineRepo.GetByOwner(ctx, username)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				h.logger.Error("Failed to get work machine", zap.Error(err))
			}
			return
		}
		mu.Lock()
		response.WorkMachine = wm
		response.WorkMachineRunning = wm.Status.State == machinesv1.MachineStateRunning
		mu.Unlock()
	}()

	// Fetch preferences
	wg.Add(1)
	go func() {
		defer wg.Done()
		prefs, err := h.userPreferencesRepo.GetOrCreate(ctx, username)
		if err != nil {
			h.logger.Error("Failed to get preferences", zap.Error(err))
			return
		}
		mu.Lock()
		response.Preferences = prefs
		pinnedIds := make([]string, 0, len(prefs.Spec.PinnedWorkspaces))
		for _, ref := range prefs.Spec.PinnedWorkspaces {
			pinnedIds = append(pinnedIds, fmt.Sprintf("%s/%s", ref.Namespace, ref.Name))
		}
		response.PinnedWorkspaceIds = pinnedIds
		mu.Unlock()
	}()

	wg.Wait()
	c.JSON(http.StatusOK, response)
}

// EnvironmentsListResponse represents the environments list response
type EnvironmentsListResponse struct {
	Environments         []environmentsv1.Environment      `json:"environments"`
	WorkMachine          *machinesv1.WorkMachine           `json:"workMachine,omitempty"`
	Preferences          *platformv1alpha1.UserPreferences `json:"preferences,omitempty"`
	PinnedEnvironmentIds []string                          `json:"pinnedEnvironmentIds"`
	WorkMachineRunning   bool                              `json:"workMachineRunning"`
}

// GetEnvironmentsListFull handles GET /environments/list-full
// Returns environments, work machine, and preferences in a single request
func (h *DashboardHandlers) GetEnvironmentsListFull(c *gin.Context) {
	ctx := c.Request.Context()

	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var response EnvironmentsListResponse
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Fetch environments and filter by visibility
	wg.Add(1)
	go func() {
		defer wg.Done()
		envs, err := h.environmentRepo.List(ctx)
		if err != nil {
			h.logger.Error("Failed to list environments", zap.Error(err))
			return
		}
		// Filter by visibility access
		var accessibleEnvs []environmentsv1.Environment
		for _, env := range envs.Items {
			if h.userHasAccessToEnvironment(username, &env) {
				accessibleEnvs = append(accessibleEnvs, env)
			}
		}
		mu.Lock()
		response.Environments = accessibleEnvs
		mu.Unlock()
	}()

	// Fetch work machine
	wg.Add(1)
	go func() {
		defer wg.Done()
		wm, err := h.workMachineRepo.GetByOwner(ctx, username)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				h.logger.Error("Failed to get work machine", zap.Error(err))
			}
			return
		}
		mu.Lock()
		response.WorkMachine = wm
		response.WorkMachineRunning = wm.Status.State == machinesv1.MachineStateRunning
		mu.Unlock()
	}()

	// Fetch preferences
	wg.Add(1)
	go func() {
		defer wg.Done()
		prefs, err := h.userPreferencesRepo.GetOrCreate(ctx, username)
		if err != nil {
			h.logger.Error("Failed to get preferences", zap.Error(err))
			return
		}
		mu.Lock()
		response.Preferences = prefs
		response.PinnedEnvironmentIds = prefs.Spec.PinnedEnvironments
		mu.Unlock()
	}()

	wg.Wait()
	c.JSON(http.StatusOK, response)
}

// WorkspaceDetailsResponse represents the workspace details response
type WorkspaceDetailsResponse struct {
	Workspace          *workspacesv1.Workspace    `json:"workspace"`
	WorkMachine        *machinesv1.WorkMachine    `json:"workMachine,omitempty"`
	PackageRequest     *packagesv1.PackageRequest `json:"packageRequest,omitempty"`
	WorkMachineRunning bool                       `json:"workMachineRunning"`
}

// GetWorkspaceDetails handles GET /namespaces/:namespace/workspaces/:name/details
// Returns workspace, work machine, and package request in a single request
func (h *DashboardHandlers) GetWorkspaceDetails(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Namespace and name are required"})
		return
	}

	username, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Fetch workspace first
	ws, err := h.workspaceRepo.Get(ctx, namespace, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
			return
		}
		h.logger.Error("Failed to get workspace", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workspace", "details": err.Error()})
		return
	}

	var response WorkspaceDetailsResponse
	response.Workspace = ws

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Fetch work machine
	wg.Add(1)
	go func() {
		defer wg.Done()
		wm, err := h.workMachineRepo.GetByOwner(ctx, username)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				h.logger.Error("Failed to get work machine", zap.Error(err))
			}
			return
		}
		mu.Lock()
		response.WorkMachine = wm
		response.WorkMachineRunning = wm.Status.State == machinesv1.MachineStateRunning
		mu.Unlock()
	}()

	// Fetch package request
	wg.Add(1)
	go func() {
		defer wg.Done()
		packageRequestName := fmt.Sprintf("%s-packages", name)
		var pkgReq packagesv1.PackageRequest
		err := h.k8sClient.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      packageRequestName,
		}, &pkgReq)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				h.logger.Warn("Failed to get package request", zap.Error(err))
			}
			return
		}
		mu.Lock()
		response.PackageRequest = &pkgReq
		mu.Unlock()
	}()

	wg.Wait()
	c.JSON(http.StatusOK, response)
}
