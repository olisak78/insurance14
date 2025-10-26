package handlers

import (
	"context"
	"net/http"

	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// JenkinsHandler handles Jenkins-related HTTP requests
type JenkinsHandler struct {
	service service.JenkinsServiceInterface
}

// NewJenkinsHandler creates a new Jenkins handler
func NewJenkinsHandler(s service.JenkinsServiceInterface) *JenkinsHandler {
	return &JenkinsHandler{service: s}
}

// JenkinsParametersResponse represents the response structure for job parameters
type JenkinsParametersResponse struct {
	ParameterDefinitions []JenkinsParameterDefinition `json:"parameterDefinitions"`
}

// JenkinsParameterDefinition represents a single parameter definition
type JenkinsParameterDefinition struct {
	Class                 string                        `json:"_class" example:"hudson.model.StringParameterDefinition"`
	DefaultParameterValue *JenkinsDefaultParameterValue `json:"defaultParameterValue,omitempty"`
	Description           string                        `json:"description" example:"Branch to deploy"`
	Name                  string                        `json:"name" example:"BRANCH"`
	Type                  string                        `json:"type" example:"StringParameterDefinition"`
	Choices               []string                      `json:"choices,omitempty" example:"dev,staging,production"`
}

// JenkinsDefaultParameterValue represents a default parameter value
type JenkinsDefaultParameterValue struct {
	Class string `json:"_class" example:"hudson.model.StringParameterValue"`
	Value string `json:"value" example:"main"`
}

// JenkinsTriggerResponse represents the response when triggering a job
type JenkinsTriggerResponse struct {
	Message string `json:"message" example:"job triggered successfully"`
}

// GetJobParameters retrieves the parameters definition for a Jenkins job
// @Summary Get Jenkins job parameters
// @Description Retrieves available parameters for a Jenkins job from the specified JAAS instance. Returns only parameters from hudson.model.ParametersDefinitionProperty.
// @Tags jenkins
// @Produce json
// @Param jaasName path string true "JAAS instance name (e.g., 'cfsmc')"
// @Param jobName path string true "Jenkins job name"
// @Success 200 {object} handlers.JenkinsParametersResponse "Filtered parameter definitions containing parameterDefinitions array with name, type, defaultParameterValue, choices, and description"
// @Failure 400 {object} map[string]string "Missing path parameter"
// @Failure 502 {object} map[string]string "Jenkins request failed"
// @Security BearerAuth
// @Router /self-service/jenkins/{jaasName}/{jobName}/parameters [get]
func (h *JenkinsHandler) GetJobParameters(c *gin.Context) {
	jaasName := c.Param("jaasName")
	if jaasName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path parameter: jaasName"})
		return
	}

	jobName := c.Param("jobName")
	if jobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path parameter: jobName"})
		return
	}

	// Create context with user information from auth claims
	ctx := createContextWithUser(c)

	result, err := h.service.GetJobParameters(ctx, jaasName, jobName)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "jenkins request failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// createContextWithUser extracts user information from gin context and adds it to a new context
func createContextWithUser(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	// Try to get auth claims from context
	if claims, exists := c.Get("auth_claims"); exists {
		if authClaims, ok := claims.(*auth.AuthClaims); ok {
			// Add user information to context
			ctx = context.WithValue(ctx, "email", authClaims.Email)
			ctx = context.WithValue(ctx, "username", authClaims.Username)
			ctx = context.WithValue(ctx, "user", authClaims.Email) // Prefer email as primary identifier
			return ctx
		}
	}

	return ctx
}

// TriggerJob triggers a Jenkins job with optional parameters
// @Summary Trigger Jenkins job
// @Description Triggers a Jenkins job on the specified JAAS instance with optional parameters. If no parameters are provided or body is empty, Jenkins will use the default values defined in the job configuration. You can override specific parameters while letting others use their defaults.
// @Tags jenkins
// @Accept json
// @Produce json
// @Param jaasName path string true "JAAS instance name (e.g., 'cfsmc')"
// @Param jobName path string true "Jenkins job name"
// @Param parameters body map[string]string false "Optional job parameters as key-value pairs. Omitted parameters will use their default values from the job configuration."
// @Success 200 {object} handlers.JenkinsTriggerResponse "Job triggered successfully"
// @Failure 400 {object} map[string]string "Missing path parameter or invalid request body"
// @Failure 502 {object} map[string]string "Jenkins trigger failed"
// @Security BearerAuth
// @Router /self-service/jenkins/{jaasName}/{jobName}/trigger [post]
func (h *JenkinsHandler) TriggerJob(c *gin.Context) {
	jaasName := c.Param("jaasName")
	if jaasName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path parameter: jaasName"})
		return
	}

	jobName := c.Param("jobName")
	if jobName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path parameter: jobName"})
		return
	}

	// Parse optional parameters from request body
	var parameters map[string]string
	if err := c.ShouldBindJSON(&parameters); err != nil {
		// If body is empty or invalid, proceed with no parameters
		parameters = make(map[string]string)
	}

	// Create context with user information from auth claims
	ctx := createContextWithUser(c)

	err := h.service.TriggerJob(ctx, jaasName, jobName, parameters)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "jenkins trigger failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "job triggered successfully"})
}
