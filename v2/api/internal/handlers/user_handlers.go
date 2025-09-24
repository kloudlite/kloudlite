package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/api/v2/internal/services"
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
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get namespace from query param or header, default to "default"
	namespace := getNamespace(c)

	// Set name and namespace in request
	req.Name = c.Query("name")
	if req.Name == "" && req.Username != "" {
		req.Name = req.Username // Use username as name if name not provided
	}
	req.Namespace = namespace

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
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

// UpdateUser handles PUT /users/:name
func (h *UserHandlers) UpdateUser(c *gin.Context) {
	name := c.Param("name")
	namespace := getNamespace(c)

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update user request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), name, namespace, &req)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err), zap.String("name", name))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
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

	req := &services.ListUsersRequest{
		Namespace:     namespace,
		LabelSelector: c.Query("labelSelector"),
		FieldSelector: c.Query("fieldSelector"),
		Continue:      c.Query("continue"),
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			req.Limit = limit
		}
	}

	users, err := h.userService.ListUsers(c.Request.Context(), namespace, req)
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