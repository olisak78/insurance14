package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockAICoreService is a simple mock implementation
type MockAICoreService struct {
	mock.Mock
}

func (m *MockAICoreService) GetDeployments(c *gin.Context) (*service.AICoreDeploymentsResponse, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreDeploymentsResponse), args.Error(1)
}

func (m *MockAICoreService) GetDeploymentDetails(c *gin.Context, deploymentID string) (*service.AICoreDeploymentDetailsResponse, error) {
	args := m.Called(c, deploymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreDeploymentDetailsResponse), args.Error(1)
}

func (m *MockAICoreService) GetModels(c *gin.Context, scenarioID string) (*service.AICoreModelsResponse, error) {
	args := m.Called(c, scenarioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreModelsResponse), args.Error(1)
}

func (m *MockAICoreService) GetConfigurations(c *gin.Context) (*service.AICoreConfigurationsResponse, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreConfigurationsResponse), args.Error(1)
}

func (m *MockAICoreService) CreateConfiguration(c *gin.Context, req *service.AICoreConfigurationRequest) (*service.AICoreConfigurationResponse, error) {
	args := m.Called(c, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreConfigurationResponse), args.Error(1)
}

func (m *MockAICoreService) CreateDeployment(c *gin.Context, req *service.AICoreDeploymentRequest) (*service.AICoreDeploymentResponse, error) {
	args := m.Called(c, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreDeploymentResponse), args.Error(1)
}

func (m *MockAICoreService) UpdateDeployment(c *gin.Context, deploymentID string, req *service.AICoreDeploymentModificationRequest) (*service.AICoreDeploymentModificationResponse, error) {
	args := m.Called(c, deploymentID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreDeploymentModificationResponse), args.Error(1)
}

func (m *MockAICoreService) DeleteDeployment(c *gin.Context, deploymentID string) (*service.AICoreDeploymentDeletionResponse, error) {
	args := m.Called(c, deploymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreDeploymentDeletionResponse), args.Error(1)
}

func (m *MockAICoreService) GetMe(c *gin.Context) (*service.AICoreMeResponse, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreMeResponse), args.Error(1)
}

func (m *MockAICoreService) ChatInference(c *gin.Context, req *service.AICoreInferenceRequest) (*service.AICoreInferenceResponse, error) {
	args := m.Called(c, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AICoreInferenceResponse), args.Error(1)
}

func (m *MockAICoreService) ChatInferenceStream(c *gin.Context, req *service.AICoreInferenceRequest, writer gin.ResponseWriter) error {
	args := m.Called(c, req, writer)
	return args.Error(0)
}

func (m *MockAICoreService) UploadAttachment(c *gin.Context, file multipart.File, header *multipart.FileHeader) (map[string]interface{}, error) {
	args := m.Called(c, file, header)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

type AICoreHandlerTestSuite struct {
	suite.Suite
	handler       *AICoreHandler
	aicoreService *MockAICoreService
	validator     *validator.Validate
	router        *gin.Engine
}

func (suite *AICoreHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	suite.aicoreService = new(MockAICoreService)
	suite.validator = validator.New()
	suite.handler = NewAICoreHandler(suite.aicoreService, suite.validator)

	suite.router = gin.New()
	suite.router.GET("/ai-core/deployments", suite.handler.GetDeployments)
	suite.router.GET("/ai-core/deployments/:deploymentId", suite.handler.GetDeploymentDetails)
	suite.router.GET("/ai-core/models", suite.handler.GetModels)
	suite.router.GET("/ai-core/configurations", suite.handler.GetConfigurations)
	suite.router.POST("/ai-core/configurations", suite.handler.CreateConfiguration)
	suite.router.POST("/ai-core/deployments", suite.handler.CreateDeployment)
	suite.router.PATCH("/ai-core/deployments/:deploymentId", suite.handler.UpdateDeployment)
	suite.router.DELETE("/ai-core/deployments/:deploymentId", suite.handler.DeleteDeployment)
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_Success() {
	// Setup
	expectedResponse := &service.AICoreDeploymentsResponse{
		Count: 2,
		Deployments: []service.AICoreTeamDeployments{
			{
				Team: "team-alpha",
				Deployments: []service.AICoreDeployment{
					{
						ID:              "deployment-1",
						ConfigurationID: "config-1",
						Status:          "RUNNING",
						StatusMessage:   "Deployment is running",
						DeploymentURL:   "https://api.example.com/v1/deployments/deployment-1",
						CreatedAt:       "2023-01-01T00:00:00Z",
						ModifiedAt:      "2023-01-01T01:00:00Z",
					},
				},
			},
			{
				Team: "team-beta",
				Deployments: []service.AICoreDeployment{
					{
						ID:              "deployment-2",
						ConfigurationID: "config-2",
						Status:          "STOPPED",
						StatusMessage:   "Deployment is stopped",
						DeploymentURL:   "https://api.example.com/v1/deployments/deployment-2",
						CreatedAt:       "2023-01-01T00:00:00Z",
						ModifiedAt:      "2023-01-01T02:00:00Z",
					},
				},
			},
		},
	}

	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusOK, w.Code)

	var response service.AICoreDeploymentsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Count)
	suite.Len(response.Deployments, 2)
	suite.Equal("team-alpha", response.Deployments[0].Team)
	suite.Equal("team-beta", response.Deployments[1].Team)
	suite.Equal("deployment-1", response.Deployments[0].Deployments[0].ID)
	suite.Equal("RUNNING", response.Deployments[0].Deployments[0].Status)
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_PartialCredentials_Success() {
	// Setup - Only one team has credentials, other is skipped
	expectedResponse := &service.AICoreDeploymentsResponse{
		Count: 1,
		Deployments: []service.AICoreTeamDeployments{
			{
				Team: "team-alpha",
				Deployments: []service.AICoreDeployment{
					{
						ID:              "deployment-1",
						ConfigurationID: "config-1",
						Status:          "RUNNING",
						StatusMessage:   "Deployment is running",
						DeploymentURL:   "https://api.example.com/v1/deployments/deployment-1",
						CreatedAt:       "2023-01-01T00:00:00Z",
						ModifiedAt:      "2023-01-01T01:00:00Z",
					},
				},
			},
		},
	}

	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusOK, w.Code)

	var response service.AICoreDeploymentsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(1, response.Count)
	suite.Len(response.Deployments, 1) // Only team with credentials returned
	suite.Equal("team-alpha", response.Deployments[0].Team)
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_EmptyResult_Success() {
	// Setup
	expectedResponse := &service.AICoreDeploymentsResponse{
		Count:       0,
		Deployments: []service.AICoreTeamDeployments{},
	}

	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusOK, w.Code)

	var response service.AICoreDeploymentsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(0, response.Count)
	suite.Len(response.Deployments, 0)
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_AuthenticationError() {
	// Setup
	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(nil, errors.ErrUserEmailNotFound)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("Authentication required", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_UserNotFoundError() {
	// Setup
	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(nil, errors.ErrUserNotFoundInDB)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusForbidden, w.Code) // AuthorizationError returns 403

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("user not found in database", response["error"]) // Actual error message
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_UserNotAssignedToTeamError() {
	// Setup
	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(nil, errors.ErrUserNotAssignedToTeam)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("user is not assigned to any team", response["error"]) // Exact error message
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_NoCredentialsError() {
	// Setup
	credentialsError := errors.NewAICoreCredentialsNotFoundError("team-alpha")
	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(nil, credentialsError)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("No AI Core credentials configured for your team", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestGetDeployments_InternalServerError() {
	// Setup
	suite.aicoreService.On("GetDeployments", mock.AnythingOfType("*gin.Context")).Return(nil, fmt.Errorf("internal error"))

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/deployments", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("internal error", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestGetModels_Success() {
	// Setup
	scenarioID := "foundation-models"
	expectedResponse := &service.AICoreModelsResponse{
		Count: 1,
		Resources: []service.AICoreModel{
			{
				Model:        "gpt-4",
				ExecutableID: "exec-1",
				Description:  "GPT-4 model",
				DisplayName:  "GPT-4",
				AccessType:   "public",
				Provider:     "openai",
			},
		},
	}

	suite.aicoreService.On("GetModels", mock.AnythingOfType("*gin.Context"), scenarioID).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("GET", fmt.Sprintf("/ai-core/models?scenarioId=%s", scenarioID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusOK, w.Code)

	var response service.AICoreModelsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(1, response.Count)
	suite.Len(response.Resources, 1)
	suite.Equal("gpt-4", response.Resources[0].Model)
}

func (suite *AICoreHandlerTestSuite) TestGetModels_MissingScenarioID() {
	// Execute
	req := httptest.NewRequest("GET", "/ai-core/models", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("scenarioId query parameter is required", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestCreateConfiguration_Success() {
	// Setup
	requestBody := service.AICoreConfigurationRequest{
		Name:         "test-config",
		ExecutableID: "exec-1",
		ScenarioID:   "foundation-models",
	}

	expectedResponse := &service.AICoreConfigurationResponse{
		ID:      "config-1",
		Message: "Configuration created successfully",
	}

	suite.aicoreService.On("CreateConfiguration", mock.AnythingOfType("*gin.Context"), &requestBody).Return(expectedResponse, nil)

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/ai-core/configurations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusCreated, w.Code)

	var response service.AICoreConfigurationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("config-1", response.ID)
	suite.Equal("Configuration created successfully", response.Message)
}

func (suite *AICoreHandlerTestSuite) TestCreateConfiguration_InvalidJSON() {
	// Execute
	req := httptest.NewRequest("POST", "/ai-core/configurations", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response["error"].(string), "invalid character")
}

func (suite *AICoreHandlerTestSuite) TestCreateConfiguration_ValidationError() {
	// Setup - missing required fields
	requestBody := service.AICoreConfigurationRequest{
		Name: "test-config",
		// Missing ExecutableID and ScenarioID
	}

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/ai-core/configurations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response["error"].(string), "required")
}

func (suite *AICoreHandlerTestSuite) TestCreateDeployment_WithConfigurationID_Success() {
	// Setup - Test scenario 1: Direct deployment with configurationId
	configID := "config-1"
	requestBody := service.AICoreDeploymentRequest{
		ConfigurationID: &configID,
		TTL:             "1h",
	}

	expectedResponse := &service.AICoreDeploymentResponse{
		ID:            "deployment-1",
		Message:       "Deployment created successfully",
		DeploymentURL: "https://api.example.com/v1/deployments/deployment-1",
		Status:        "PENDING",
		TTL:           "1h",
	}

	suite.aicoreService.On("CreateDeployment", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(req *service.AICoreDeploymentRequest) bool {
		return req.ConfigurationID != nil && *req.ConfigurationID == "config-1" && req.ConfigurationRequest == nil
	})).Return(expectedResponse, nil)

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/ai-core/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusAccepted, w.Code)

	var response service.AICoreDeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("deployment-1", response.ID)
	suite.Equal("Deployment created successfully", response.Message)
}

func (suite *AICoreHandlerTestSuite) TestCreateDeployment_WithConfigurationRequest_Success() {
	// Setup - Test scenario 2: Deployment with configuration creation
	requestBody := service.AICoreDeploymentRequest{
		ConfigurationRequest: &service.AICoreConfigurationRequest{
			Name:         "my-llm-config",
			ExecutableID: "aicore-llm",
			ScenarioID:   "foundation-models",
			ParameterBindings: []map[string]string{
				{"key": "modelName", "value": "gpt-4"},
				{"key": "modelVersion", "value": "latest"},
			},
		},
		TTL: "2h",
	}

	expectedResponse := &service.AICoreDeploymentResponse{
		ID:            "deployment-2",
		Message:       "Deployment created successfully",
		DeploymentURL: "https://api.example.com/v1/deployments/deployment-2",
		Status:        "PENDING",
		TTL:           "2h",
	}

	suite.aicoreService.On("CreateDeployment", mock.AnythingOfType("*gin.Context"), mock.MatchedBy(func(req *service.AICoreDeploymentRequest) bool {
		return req.ConfigurationID == nil && req.ConfigurationRequest != nil &&
			req.ConfigurationRequest.Name == "my-llm-config" &&
			req.ConfigurationRequest.ExecutableID == "aicore-llm" &&
			req.ConfigurationRequest.ScenarioID == "foundation-models"
	})).Return(expectedResponse, nil)

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/ai-core/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusAccepted, w.Code)

	var response service.AICoreDeploymentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("deployment-2", response.ID)
	suite.Equal("Deployment created successfully", response.Message)
}

func (suite *AICoreHandlerTestSuite) TestCreateDeployment_BothFieldsProvided_Error() {
	// Setup - Test invalid scenario: both configurationId and configurationRequest provided
	configID := "config-1"
	requestBody := service.AICoreDeploymentRequest{
		ConfigurationID: &configID,
		ConfigurationRequest: &service.AICoreConfigurationRequest{
			Name:         "my-llm-config",
			ExecutableID: "aicore-llm",
			ScenarioID:   "foundation-models",
		},
		TTL: "1h",
	}

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/ai-core/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("ConfigurationId and configurationRequest cannot both be provided", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestCreateDeployment_NeitherFieldProvided_Error() {
	// Setup - Test invalid scenario: neither configurationId nor configurationRequest provided
	requestBody := service.AICoreDeploymentRequest{
		TTL: "1h",
	}

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/ai-core/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("Either configurationId or configurationRequest must be provided", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestCreateDeployment_InvalidConfigurationRequest_Error() {
	// Setup - Test invalid scenario: configurationRequest with missing required fields
	requestBody := service.AICoreDeploymentRequest{
		ConfigurationRequest: &service.AICoreConfigurationRequest{
			Name: "my-llm-config",
			// Missing ExecutableID and ScenarioID
		},
		TTL: "1h",
	}

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/ai-core/deployments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response["error"].(string), "required")
}

func (suite *AICoreHandlerTestSuite) TestUpdateDeployment_Success() {
	// Setup
	deploymentID := "deployment-1"
	requestBody := service.AICoreDeploymentModificationRequest{
		TargetStatus: "STOPPED",
	}

	expectedResponse := &service.AICoreDeploymentModificationResponse{
		ID:           "deployment-1",
		Message:      "Deployment updated successfully",
		Status:       "RUNNING",
		TargetStatus: "STOPPED",
	}

	suite.aicoreService.On("UpdateDeployment", mock.AnythingOfType("*gin.Context"), deploymentID, &requestBody).Return(expectedResponse, nil)

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/ai-core/deployments/%s", deploymentID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusAccepted, w.Code)

	var response service.AICoreDeploymentModificationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("deployment-1", response.ID)
	suite.Equal("STOPPED", response.TargetStatus)
}

func (suite *AICoreHandlerTestSuite) TestUpdateDeployment_MissingDeploymentID() {
	// Execute
	requestBody := service.AICoreDeploymentModificationRequest{
		TargetStatus: "STOPPED",
	}
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/ai-core/deployments/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusNotFound, w.Code) // Gin returns 404 for missing path parameters
}

func (suite *AICoreHandlerTestSuite) TestUpdateDeployment_EmptyRequest() {
	// Setup
	deploymentID := "deployment-1"
	requestBody := service.AICoreDeploymentModificationRequest{
		// Both fields empty
	}

	// Execute
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/ai-core/deployments/%s", deploymentID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("At least one of targetStatus or configurationId must be provided", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestDeleteDeployment_Success() {
	// Setup
	deploymentID := "deployment-1"
	expectedResponse := &service.AICoreDeploymentDeletionResponse{
		ID:      "deployment-1",
		Message: "Deployment deleted successfully",
	}

	suite.aicoreService.On("DeleteDeployment", mock.AnythingOfType("*gin.Context"), deploymentID).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/ai-core/deployments/%s", deploymentID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusAccepted, w.Code)

	var response service.AICoreDeploymentDeletionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("deployment-1", response.ID)
	suite.Equal("Deployment deleted successfully", response.Message)
}

func (suite *AICoreHandlerTestSuite) TestGetDeploymentDetails_Success() {
	// Setup
	deploymentID := "deployment-1"
	expectedResponse := &service.AICoreDeploymentDetailsResponse{
		ID:                "deployment-1",
		DeploymentURL:     "https://api.example.com/v1/deployments/deployment-1",
		ConfigurationID:   "config-1",
		ConfigurationName: "test-config",
		ExecutableID:      "exec-1",
		ScenarioID:        "foundation-models",
		Status:            "RUNNING",
		StatusMessage:     "Deployment is running",
		TargetStatus:      "RUNNING",
		LastOperation:     "CREATE",
		TTL:               "1h",
		CreatedAt:         "2023-01-01T00:00:00Z",
		ModifiedAt:        "2023-01-01T01:00:00Z",
	}

	suite.aicoreService.On("GetDeploymentDetails", mock.AnythingOfType("*gin.Context"), deploymentID).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("GET", fmt.Sprintf("/ai-core/deployments/%s", deploymentID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusOK, w.Code)

	var response service.AICoreDeploymentDetailsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("deployment-1", response.ID)
	suite.Equal("RUNNING", response.Status)
	suite.Equal("config-1", response.ConfigurationID)
}

func (suite *AICoreHandlerTestSuite) TestGetDeploymentDetails_NotFound() {
	// Setup
	deploymentID := "nonexistent-deployment"
	suite.aicoreService.On("GetDeploymentDetails", mock.AnythingOfType("*gin.Context"), deploymentID).Return(nil, errors.ErrAICoreDeploymentNotFound)

	// Execute
	req := httptest.NewRequest("GET", fmt.Sprintf("/ai-core/deployments/%s", deploymentID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Contains(response["error"].(string), "not found")
}

func (suite *AICoreHandlerTestSuite) TestGetConfigurations_Success() {
	// Setup
	expectedResponse := &service.AICoreConfigurationsResponse{
		Count: 2,
		Resources: []service.AICoreConfiguration{
			{
				ID:           "config-1",
				Name:         "test-config-1",
				ExecutableID: "exec-1",
				ScenarioID:   "foundation-models",
				CreatedAt:    "2023-01-01T00:00:00Z",
			},
			{
				ID:           "config-2",
				Name:         "test-config-2",
				ExecutableID: "exec-2",
				ScenarioID:   "foundation-models",
				CreatedAt:    "2023-01-02T00:00:00Z",
			},
		},
	}

	suite.aicoreService.On("GetConfigurations", mock.AnythingOfType("*gin.Context")).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/configurations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusOK, w.Code)

	var response service.AICoreConfigurationsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Count)
	suite.Len(response.Resources, 2)
	suite.Equal("config-1", response.Resources[0].ID)
	suite.Equal("test-config-1", response.Resources[0].Name)
	suite.Equal("foundation-models", response.Resources[0].ScenarioID)
}

func (suite *AICoreHandlerTestSuite) TestGetConfigurations_EmptyResult() {
	// Setup
	expectedResponse := &service.AICoreConfigurationsResponse{
		Count:     0,
		Resources: []service.AICoreConfiguration{},
	}

	suite.aicoreService.On("GetConfigurations", mock.AnythingOfType("*gin.Context")).Return(expectedResponse, nil)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/configurations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusOK, w.Code)

	var response service.AICoreConfigurationsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(0, response.Count)
	suite.Len(response.Resources, 0)
}

func (suite *AICoreHandlerTestSuite) TestGetConfigurations_AuthenticationError() {
	// Setup
	suite.aicoreService.On("GetConfigurations", mock.AnythingOfType("*gin.Context")).Return(nil, errors.ErrUserEmailNotFound)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/configurations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("Authentication required", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestGetConfigurations_UserNotAssignedToTeamError() {
	// Setup
	suite.aicoreService.On("GetConfigurations", mock.AnythingOfType("*gin.Context")).Return(nil, errors.ErrUserNotAssignedToTeam)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/configurations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("user is not assigned to any team", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestGetConfigurations_NoCredentialsError() {
	// Setup
	credentialsError := errors.NewAICoreCredentialsNotFoundError("team-alpha")
	suite.aicoreService.On("GetConfigurations", mock.AnythingOfType("*gin.Context")).Return(nil, credentialsError)

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/configurations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("No AI Core credentials configured for your team", response["error"])
}

func (suite *AICoreHandlerTestSuite) TestGetConfigurations_InternalServerError() {
	// Setup
	suite.aicoreService.On("GetConfigurations", mock.AnythingOfType("*gin.Context")).Return(nil, fmt.Errorf("internal error"))

	// Execute
	req := httptest.NewRequest("GET", "/ai-core/configurations", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	suite.Equal(http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("internal error", response["error"])
}

func TestAICoreHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AICoreHandlerTestSuite))
}
