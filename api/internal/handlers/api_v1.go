package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/config"
	"github.com/kloudlite/kloudlite/api/internal/dto"
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
	c.JSON(http.StatusOK, dto.InfoResponse{
		Version:     "v2.0.0",
		Environment: h.config.Environment,
	})
}
