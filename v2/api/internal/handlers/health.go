package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/internal/dto"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status: "healthy",
		Time:   time.Now().Unix(),
	})
}

func ReadinessCheck(c *gin.Context) {
	// TODO: Add actual readiness checks (database connection, etc.)
	c.JSON(http.StatusOK, dto.HealthResponse{
		Status: "ready",
		Time:   time.Now().Unix(),
	})
}