package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DashboardHandlers handles HTTP requests for the dashboard
type DashboardHandlers struct {
	machineTypeRepo     repository.MachineTypeRepository
	workMachineRepo     repository.WorkMachineRepository
	userPreferencesRepo repository.UserPreferencesRepository
	workspaceRepo       repository.WorkspaceRepository
	environmentRepo     repository.EnvironmentRepository
	logger              *zap.Logger
}

// NewDashboardHandlers creates a new DashboardHandlers
func NewDashboardHandlers(
	machineTypeRepo repository.MachineTypeRepository,
	workMachineRepo repository.WorkMachineRepository,
	userPreferencesRepo repository.UserPreferencesRepository,
	workspaceRepo repository.WorkspaceRepository,
	environmentRepo repository.EnvironmentRepository,
	logger *zap.Logger,
) *DashboardHandlers {
	return &DashboardHandlers{
		machineTypeRepo:     machineTypeRepo,
		workMachineRepo:     workMachineRepo,
		userPreferencesRepo: userPreferencesRepo,
		workspaceRepo:       workspaceRepo,
		environmentRepo:     environmentRepo,
		logger:              logger,
	}
}

// DashboardResponse represents the dashboard data response
type DashboardResponse struct {
	MachineTypes       []machinesv1.MachineType         `json:"machineTypes"`
	WorkMachines       []machinesv1.WorkMachine         `json:"workMachines"`
	Preferences        *platformv1alpha1.UserPreferences `json:"preferences,omitempty"`
	PinnedWorkspaces   []workspacesv1.Workspace         `json:"pinnedWorkspaces"`
	PinnedEnvironments []environmentsv1.Environment     `json:"pinnedEnvironments"`
	IsAdmin            bool                             `json:"isAdmin"`
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
