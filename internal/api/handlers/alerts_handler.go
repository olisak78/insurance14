package handlers

import (
	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlertsHandler struct {
	alertsService *service.AlertsService
}

func NewAlertsHandler(alertsService *service.AlertsService) *AlertsHandler {
	return &AlertsHandler{
		alertsService: alertsService,
	}
}

// GetAlerts godoc
// @Summary Get Prometheus alerts from GitHub repository
// @Description Fetches all Prometheus alert configurations from the configured GitHub repository
// @Tags alerts
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} map[string]interface{} "Alerts data"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/projects/{projectId}/alerts [get]
func (h *AlertsHandler) GetAlerts(c *gin.Context) {
	// Get authenticated user claims
	claimsInterface, exists := c.Get("auth_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	claims, ok := claimsInterface.(*auth.AuthClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authentication claims"})
		return
	}

	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	alerts, err := h.alertsService.GetProjectAlerts(c.Request.Context(), projectID, claims)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		if err.Error() == "alerts repository not configured for this project" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alerts repository not configured for this project"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// CreateAlertPR godoc
// @Summary Create a pull request with alert changes
// @Description Creates a pull request to update Prometheus alert configurations in GitHub
// @Tags alerts
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param body body map[string]interface{} true "Alert changes"
// @Success 200 {object} map[string]interface{} "PR created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/projects/{projectId}/alerts/pr [post]
func (h *AlertsHandler) CreateAlertPR(c *gin.Context) {
	// Get authenticated user claims
	claimsInterface, exists := c.Get("auth_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	claims, ok := claimsInterface.(*auth.AuthClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authentication claims"})
		return
	}

	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	var payload struct {
		FileName    string `json:"fileName"`
		Content     string `json:"content"`
		Message     string `json:"message"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	prURL, err := h.alertsService.CreateAlertPR(c.Request.Context(), projectID, claims, payload.FileName, payload.Content, payload.Message, payload.Description)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pull request created successfully",
		"prUrl":   prURL,
	})
}
