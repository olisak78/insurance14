package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// ErrorResponse represents a standard API error response
type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// Health returns the health status of the application
// @Summary Health check
// @Description Get the overall health status of the application including database connectivity
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Application is healthy"
// @Failure 503 {object} HealthResponse "Application is unhealthy"
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Services:  make(map[string]string),
	}

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Status = "unhealthy"
		response.Services["database"] = "error: " + err.Error()
	} else {
		if err := sqlDB.Ping(); err != nil {
			response.Status = "unhealthy"
			response.Services["database"] = "error: " + err.Error()
		} else {
			response.Services["database"] = "healthy"
		}
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Ready returns the readiness status of the application
// @Summary Readiness check
// @Description Check if the application is ready to serve requests
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Application is ready"
// @Failure 503 {object} map[string]interface{} "Application is not ready"
// @Router /health/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check if the application is ready to serve requests
	// This could include checking if migrations are complete, external services are available, etc.

	ready := true
	services := make(map[string]string)

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		ready = false
		services["database"] = "not ready: " + err.Error()
	} else {
		if err := sqlDB.Ping(); err != nil {
			ready = false
			services["database"] = "not ready: " + err.Error()
		} else {
			services["database"] = "ready"
		}
	}

	response := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now(),
		"services":  services,
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Live returns the liveness status of the application
// @Summary Liveness check
// @Description Check if the application is alive and responding
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Application is alive"
// @Router /health/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	// Simple liveness check - if we can respond, we're alive
	c.JSON(http.StatusOK, map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now(),
	})
}
