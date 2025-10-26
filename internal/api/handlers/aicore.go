package handlers

import (
	"net/http"

	"developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AICoreHandler handles HTTP requests for AI Core operations
type AICoreHandler struct {
	aicoreService service.AICoreServiceInterface
	validator     *validator.Validate
}

// NewAICoreHandler creates a new AI Core handler
func NewAICoreHandler(aicoreService service.AICoreServiceInterface, validator *validator.Validate) *AICoreHandler {
	return &AICoreHandler{
		aicoreService: aicoreService,
		validator:     validator,
	}
}

// handleAICoreError handles common AI Core service errors and returns appropriate HTTP responses
func (h *AICoreHandler) handleAICoreError(c *gin.Context, err error) {
	switch {
	case errors.IsAuthentication(err):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
	case errors.IsAuthorization(err):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.IsConfiguration(err):
		c.JSON(http.StatusForbidden, gin.H{"error": "No AI Core credentials configured for your team"})
	case errors.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetDeployments handles GET /ai-core/deployments
// @Summary Get AI Core deployments
// @Description Get all deployments from AI Core for the authenticated user's team
// @Tags ai-core
// @Accept json
// @Produce json
// @Success 200 {object} service.AICoreDeploymentsResponse "Successfully retrieved deployments"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/deployments [get]
func (h *AICoreHandler) GetDeployments(c *gin.Context) {
	deployments, err := h.aicoreService.GetDeployments(c)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// GetModels handles GET /ai-core/models
// @Summary Get AI Core models
// @Description Get all available models from AI Core for a specific scenario
// @Tags ai-core
// @Accept json
// @Produce json
// @Param scenarioId query string true "Scenario ID to get models for"
// @Success 200 {object} service.AICoreModelsResponse "Successfully retrieved models"
// @Failure 400 {object} map[string]interface{} "Bad request - missing scenarioId parameter"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/models [get]
func (h *AICoreHandler) GetModels(c *gin.Context) {
	scenarioID := c.Query("scenarioId")
	if scenarioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scenarioId query parameter is required"})
		return
	}

	models, err := h.aicoreService.GetModels(c, scenarioID)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, models)
}

// GetConfigurations handles GET /ai-core/configurations
// @Summary Get AI Core configurations
// @Description Get all configurations from AI Core for the authenticated user's team
// @Tags ai-core
// @Accept json
// @Produce json
// @Success 200 {object} service.AICoreConfigurationsResponse "Successfully retrieved configurations"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/configurations [get]
func (h *AICoreHandler) GetConfigurations(c *gin.Context) {
	configurations, err := h.aicoreService.GetConfigurations(c)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, configurations)
}

// CreateConfiguration handles POST /ai-core/configurations
// @Summary Create AI Core configuration
// @Description Create a new configuration in AI Core for the authenticated user's team
// @Tags ai-core
// @Accept json
// @Produce json
// @Param configuration body service.AICoreConfigurationRequest true "Configuration data"
// @Success 201 {object} service.AICoreConfigurationResponse "Successfully created configuration"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/configurations [post]
func (h *AICoreHandler) CreateConfiguration(c *gin.Context) {
	var req service.AICoreConfigurationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	configuration, err := h.aicoreService.CreateConfiguration(c, &req)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusCreated, configuration)
}

// CreateDeployment handles POST /ai-core/deployments
// @Summary Create AI Core deployment
// @Description Create a new deployment in AI Core using a configuration
// @Tags ai-core
// @Accept json
// @Produce json
// @Param deployment body service.AICoreDeploymentRequest true "Deployment data"
// @Success 202 {object} service.AICoreDeploymentResponse "Successfully scheduled deployment"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/deployments [post]
func (h *AICoreHandler) CreateDeployment(c *gin.Context) {
	var req service.AICoreDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deployment, err := h.aicoreService.CreateDeployment(c, &req)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, deployment)
}

// UpdateDeployment handles PATCH /ai-core/deployments/{deploymentId}
// @Summary Update AI Core deployment
// @Description Update target status or configuration of a deployment in AI Core
// @Tags ai-core
// @Accept json
// @Produce json
// @Param deploymentId path string true "Deployment ID"
// @Param modification body service.AICoreDeploymentModificationRequest true "Deployment modification data"
// @Success 202 {object} service.AICoreDeploymentModificationResponse "Successfully scheduled deployment modification"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 404 {object} map[string]interface{} "Deployment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/deployments/{deploymentId} [patch]
func (h *AICoreHandler) UpdateDeployment(c *gin.Context) {
	deploymentID := c.Param("deploymentId")
	if deploymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deploymentId parameter is required"})
		return
	}

	var req service.AICoreDeploymentModificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// At least one field should be provided
	if req.TargetStatus == "" && req.ConfigurationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one of targetStatus or configurationId must be provided"})
		return
	}

	response, err := h.aicoreService.UpdateDeployment(c, deploymentID, &req)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, response)
}

// GetDeploymentDetails handles GET /ai-core/deployments/{deploymentId}
// @Summary Get AI Core deployment details
// @Description Get detailed information about a specific deployment from AI Core
// @Tags ai-core
// @Accept json
// @Produce json
// @Param deploymentId path string true "Deployment ID"
// @Success 200 {object} service.AICoreDeploymentDetailsResponse "Successfully retrieved deployment details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 404 {object} map[string]interface{} "Deployment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/deployments/{deploymentId} [get]
func (h *AICoreHandler) GetDeploymentDetails(c *gin.Context) {
	deploymentID := c.Param("deploymentId")
	if deploymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deploymentId parameter is required"})
		return
	}

	response, err := h.aicoreService.GetDeploymentDetails(c, deploymentID)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteDeployment handles DELETE /ai-core/deployments/{deploymentId}
// @Summary Delete AI Core deployment
// @Description Mark deployment as deleted in AI Core
// @Tags ai-core
// @Accept json
// @Produce json
// @Param deploymentId path string true "Deployment ID"
// @Success 202 {object} service.AICoreDeploymentDeletionResponse "Successfully scheduled deployment deletion"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "User not assigned to team or team credentials not found"
// @Failure 404 {object} map[string]interface{} "Deployment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /ai-core/deployments/{deploymentId} [delete]
func (h *AICoreHandler) DeleteDeployment(c *gin.Context) {
	deploymentID := c.Param("deploymentId")
	if deploymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deploymentId parameter is required"})
		return
	}

	response, err := h.aicoreService.DeleteDeployment(c, deploymentID)
	if err != nil {
		h.handleAICoreError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, response)
}
