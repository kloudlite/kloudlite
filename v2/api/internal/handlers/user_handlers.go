package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	"github.com/kloudlite/kloudlite/v2/api/internal/middleware"
	"github.com/kloudlite/kloudlite/v2/api/internal/repository"
	"github.com/kloudlite/kloudlite/v2/api/internal/services"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"go.uber.org/zap"
)

// UserHandlers handles HTTP requests for User resources
type UserHandlers struct {
	userService services.UserService
	logger      *zap.Logger
}

// NewUserHandlers creates a new UserHandlers
func NewUserHandlers(userService services.UserService, logger *zap.Logger) *UserHandlers {
	return &UserHandlers{
		userService: userService,
		logger:      logger,
	}
}

// Helper functions for authorization

// getCurrentUserRoles extracts user roles from JWT context (set by auth middleware)
func (h *UserHandlers) getCurrentUserRoles(c *gin.Context) []platformv1alpha1.RoleType {
	_, roles, exists := middleware.GetUserFromContext(c)
	if !exists {
		return []platformv1alpha1.RoleType{}
	}
	return roles
}

// hasRole checks if user has a specific role
func (h *UserHandlers) hasRole(roles []platformv1alpha1.RoleType, targetRole platformv1alpha1.RoleType) bool {
	for _, role := range roles {
		if role == targetRole {
			return true
		}
	}
	return false
}

// canCreateUserWithRoles checks if current user can create a user with the specified roles
func (h *UserHandlers) canCreateUserWithRoles(currentUserRoles []platformv1alpha1.RoleType, targetRoles []platformv1alpha1.RoleType) bool {
	isSuperAdmin := h.hasRole(currentUserRoles, platformv1alpha1.RoleSuperAdmin)
	isAdmin := h.hasRole(currentUserRoles, platformv1alpha1.RoleAdmin)

	// Super admin can create users with any roles
	if isSuperAdmin {
		return true
	}

	// Admin can only create regular users
	if isAdmin {
		for _, role := range targetRoles {
			if role != platformv1alpha1.RoleUser {
				return false // Admin trying to create admin or super-admin user
			}
		}
		return true
	}

	// Regular users cannot create any users
	return false
}

// canModifyUser checks if current user can modify the target user
func (h *UserHandlers) canModifyUser(currentUserRoles []platformv1alpha1.RoleType, targetUser *platformv1alpha1.User) bool {
	isSuperAdmin := h.hasRole(currentUserRoles, platformv1alpha1.RoleSuperAdmin)
	isAdmin := h.hasRole(currentUserRoles, platformv1alpha1.RoleAdmin)

	// Super admin can modify any user
	if isSuperAdmin {
		return true
	}

	// Admin can only modify regular users
	if isAdmin {
		for _, role := range targetUser.Spec.Roles {
			if role == platformv1alpha1.RoleAdmin || role == platformv1alpha1.RoleSuperAdmin {
				return false // Admin trying to modify admin or super-admin user
			}
		}
		return true
	}

	// Regular users cannot modify any users
	return false
}

// CreateUser handles POST /users
func (h *UserHandlers) CreateUser(c *gin.Context) {
	// Check authorization first
	currentUserRoles := h.getCurrentUserRoles(c)
	if len(currentUserRoles) == 0 {
		h.logger.Error("No user roles found in request")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: No user roles found",
		})
		return
	}

	var userSpec platformv1alpha1.UserSpec
	if err := c.ShouldBindJSON(&userSpec); err != nil {
		h.logger.Error("Failed to parse create user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Check if current user can create a user with the specified roles
	if !h.canCreateUserWithRoles(currentUserRoles, userSpec.Roles) {
		h.logger.Error("User lacks permission to create user with specified roles",
			zap.Any("currentUserRoles", currentUserRoles),
			zap.Any("targetRoles", userSpec.Roles))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Insufficient permissions to create user with specified roles",
		})
		return
	}

	// Validate user spec
	if err := h.validateUserSpec(&userSpec); err != nil {
		h.logger.Error("User validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "User validation failed",
			"details": err.Error(),
		})
		return
	}

	// Create User object (cluster-scoped, no namespace)
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "", // Users are cluster-scoped
		},
		Spec: userSpec,
	}

	// Get name from query if provided, otherwise use GenerateName
	if name := c.Query("name"); name != "" {
		user.Name = name
	} else {
		user.GenerateName = "user-"
	}

	createdUser, err := h.userService.CreateUser(c.Request.Context(), user)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

// GetUser handles GET /users/:name
func (h *UserHandlers) GetUser(c *gin.Context) {
	name := c.Param("name")

	user, err := h.userService.GetUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserByEmail handles GET /users/by-email?email=xxx
func (h *UserHandlers) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email parameter is required",
		})
		return
	}

	user, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		h.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles PUT /users/:name (full replace)
func (h *UserHandlers) UpdateUser(c *gin.Context) {
	name := c.Param("name")

	// Check authorization first
	currentUserRoles := h.getCurrentUserRoles(c)
	if len(currentUserRoles) == 0 {
		h.logger.Error("No user roles found in request")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: No user roles found",
		})
		return
	}

	// First get the existing user to preserve metadata
	existingUser, err := h.userService.GetUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get user for update", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	// Check if current user can modify this user
	if !h.canModifyUser(currentUserRoles, existingUser) {
		h.logger.Error("User lacks permission to modify target user",
			zap.Any("currentUserRoles", currentUserRoles),
			zap.Any("targetUserRoles", existingUser.Spec.Roles),
			zap.String("targetUserName", name))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Insufficient permissions to modify this user",
		})
		return
	}

	var userSpec platformv1alpha1.UserSpec
	if err := c.ShouldBindJSON(&userSpec); err != nil {
		h.logger.Error("Failed to parse update user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Check if current user can assign the new roles
	if !h.canCreateUserWithRoles(currentUserRoles, userSpec.Roles) {
		h.logger.Error("User lacks permission to assign specified roles",
			zap.Any("currentUserRoles", currentUserRoles),
			zap.Any("targetRoles", userSpec.Roles))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Insufficient permissions to assign specified roles",
		})
		return
	}

	// Validate user spec for full updates
	if err := h.validateUserSpec(&userSpec); err != nil {
		h.logger.Error("User validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "User validation failed",
			"details": err.Error(),
		})
		return
	}

	// Update only the spec, preserve metadata including ResourceVersion
	existingUser.Spec = userSpec

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), existingUser)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// ActivateUser handles POST /users/:name/activate
func (h *UserHandlers) ActivateUser(c *gin.Context) {
	name := c.Param("name")

	// Check authorization first
	currentUserRoles := h.getCurrentUserRoles(c)
	if len(currentUserRoles) == 0 {
		h.logger.Error("No user roles found in request")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: No user roles found",
		})
		return
	}

	// Get the target user to check permissions
	targetUser, err := h.userService.GetUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get user for activation", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	// Check if current user can modify this user
	if !h.canModifyUser(currentUserRoles, targetUser) {
		h.logger.Error("User lacks permission to activate target user",
			zap.Any("currentUserRoles", currentUserRoles),
			zap.Any("targetUserRoles", targetUser.Spec.Roles),
			zap.String("targetUserName", name))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Insufficient permissions to activate this user",
		})
		return
	}

	user, err := h.userService.ActivateUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to activate user", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to activate user",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User activated successfully", zap.String("name", name))
	c.JSON(http.StatusOK, user)
}

// DeactivateUser handles POST /users/:name/deactivate
func (h *UserHandlers) DeactivateUser(c *gin.Context) {
	name := c.Param("name")

	// Check authorization first
	currentUserRoles := h.getCurrentUserRoles(c)
	if len(currentUserRoles) == 0 {
		h.logger.Error("No user roles found in request")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: No user roles found",
		})
		return
	}

	// Get the target user to check permissions
	targetUser, err := h.userService.GetUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get user for deactivation", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	// Check if current user can modify this user
	if !h.canModifyUser(currentUserRoles, targetUser) {
		h.logger.Error("User lacks permission to deactivate target user",
			zap.Any("currentUserRoles", currentUserRoles),
			zap.Any("targetUserRoles", targetUser.Spec.Roles),
			zap.String("targetUserName", name))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Insufficient permissions to deactivate this user",
		})
		return
	}

	user, err := h.userService.DeactivateUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to deactivate user", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to deactivate user",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User deactivated successfully", zap.String("name", name))
	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /users/:name
func (h *UserHandlers) DeleteUser(c *gin.Context) {
	name := c.Param("name")

	// Check authorization first
	currentUserRoles := h.getCurrentUserRoles(c)
	if len(currentUserRoles) == 0 {
		h.logger.Error("No user roles found in request")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized: No user roles found",
		})
		return
	}

	// Get the target user to check permissions
	targetUser, err := h.userService.GetUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get user for deletion", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"details": err.Error(),
		})
		return
	}

	// Check if current user can modify (delete) this user
	if !h.canModifyUser(currentUserRoles, targetUser) {
		h.logger.Error("User lacks permission to delete target user",
			zap.Any("currentUserRoles", currentUserRoles),
			zap.Any("targetUserRoles", targetUser.Spec.Roles),
			zap.String("targetUserName", name))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Insufficient permissions to delete this user",
		})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ResetUserPassword handles POST /users/:name/reset-password
func (h *UserHandlers) ResetUserPassword(c *gin.Context) {
	name := c.Param("name")

	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	err := h.userService.ResetUserPassword(c.Request.Context(), name, req.NewPassword)
	if err != nil {
		h.logger.Error("Failed to reset user password", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reset password",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User password reset successfully", zap.String("name", name))
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// UpdateUserLastLogin handles POST /users/:name/update-last-login
func (h *UserHandlers) UpdateUserLastLogin(c *gin.Context) {
	name := c.Param("name")

	err := h.userService.UpdateUserLastLogin(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to update user last login", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update last login",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User last login updated successfully", zap.String("name", name))
	c.JSON(http.StatusOK, gin.H{
		"message": "Last login updated successfully",
	})
}


// ListUsers handles GET /users
func (h *UserHandlers) ListUsers(c *gin.Context) {
	var opts []repository.ListOption

	// Add label selector if provided
	if labelSelector := c.Query("labelSelector"); labelSelector != "" {
		opts = append(opts, repository.WithLabelSelector(labelSelector))
	}

	// Add field selector if provided
	if fieldSelector := c.Query("fieldSelector"); fieldSelector != "" {
		opts = append(opts, repository.WithFieldSelector(fieldSelector))
	}

	// Add limit if provided
	if limit := c.Query("limit"); limit != "" {
		var limitVal int64
		if _, err := fmt.Sscanf(limit, "%d", &limitVal); err == nil {
			opts = append(opts, repository.WithLimit(limitVal))
		}
	}

	// Add continue token if provided
	if continueToken := c.Query("continue"); continueToken != "" {
		opts = append(opts, repository.WithContinue(continueToken))
	}

	users, err := h.userService.ListUsers(c.Request.Context(), opts...)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list users",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// validateUserSpec validates the user specification
func (h *UserHandlers) validateUserSpec(userSpec *platformv1alpha1.UserSpec) error {
	// Validate email
	if userSpec.Email == "" {
		return fmt.Errorf("email is required")
	}

	// Basic email validation
	if !strings.Contains(userSpec.Email, "@") || !strings.Contains(userSpec.Email, ".") {
		return fmt.Errorf("invalid email format")
	}

	// Validate roles - must have at least one role
	if len(userSpec.Roles) == 0 {
		return fmt.Errorf("at least one role is required")
	}

	// Validate each role is valid
	validRoles := map[string]bool{
		"super-admin": true,
		"admin":       true,
		"user":        true,
	}

	for _, role := range userSpec.Roles {
		if !validRoles[string(role)] {
			return fmt.Errorf("invalid role '%s'. Valid roles are: super-admin, admin, user", role)
		}
	}

	return nil
}

// getNamespace extracts namespace from query param, header, or uses default
func getNamespace(c *gin.Context) string {
	// Try query parameter first
	if ns := c.Query("namespace"); ns != "" {
		return ns
	}

	// Try header
	if ns := c.GetHeader("X-Namespace"); ns != "" {
		return ns
	}

	// Default namespace
	return "default"
}