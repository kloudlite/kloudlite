package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/internal/config"
)

type APIHandlers struct {
	config *config.Config
}

func NewAPIHandlers(cfg *config.Config) *APIHandlers {
	return &APIHandlers{
		config: cfg,
	}
}

func (h *APIHandlers) GetInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":     "v2.0.0",
		"environment": h.config.Environment,
	})
}