package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	platformv1alpha1 "github.com/kloudlite/api/v2/pkg/apis/platform/v1alpha1"
	"github.com/kloudlite/api/v2/internal/repository"
	"github.com/kloudlite/api/v2/internal/services"
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

// CreateUser handles POST /users
func (h *UserHandlers) CreateUser(c *gin.Context) {
	var userSpec platformv1alpha1.UserSpec
	if err := c.ShouldBindJSON(&userSpec); err != nil {
		h.logger.Error("Failed to parse create user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get namespace from query param or header, default to "default"
	namespace := getNamespace(c)

	// Create User object
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
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
	namespace := getNamespace(c)

	user, err := h.userService.GetUser(c.Request.Context(), name, namespace)
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

// UpdateUser handles PUT /users/:name
func (h *UserHandlers) UpdateUser(c *gin.Context) {
	name := c.Param("name")
	namespace := getNamespace(c)

	var userSpec platformv1alpha1.UserSpec
	if err := c.ShouldBindJSON(&userSpec); err != nil {
		h.logger.Error("Failed to parse update user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create User object with updated spec
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: userSpec,
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), user)
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

// DeleteUser handles DELETE /users/:name
func (h *UserHandlers) DeleteUser(c *gin.Context) {
	name := c.Param("name")
	namespace := getNamespace(c)

	err := h.userService.DeleteUser(c.Request.Context(), name, namespace)
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

// ListUsers handles GET /users
func (h *UserHandlers) ListUsers(c *gin.Context) {
	namespace := getNamespace(c)

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

	users, err := h.userService.ListUsers(c.Request.Context(), namespace, opts...)
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