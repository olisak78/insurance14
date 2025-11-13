package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// JenkinsHandlerTestSuite defines the test suite for JenkinsHandler
type JenkinsHandlerTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupTest sets up the test suite
func (suite *JenkinsHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
}

// TearDownTest cleans up after each test
func (suite *JenkinsHandlerTestSuite) TearDownTest() {
	// Cleanup
}

// MockJenkinsService is a mock implementation for testing
type MockJenkinsService struct {
	GetJobParametersFunc    func(ctx context.Context, jaasName, jobName string) (interface{}, error)
	TriggerJobFunc          func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error)
	GetQueueItemStatusFunc  func(ctx context.Context, jaasName, queueItemID string) (*service.JenkinsQueueStatusResult, error)
	GetBuildStatusFunc      func(ctx context.Context, jaasName, jobName string, buildNumber int) (*service.JenkinsBuildStatusResult, error)
}

func (m *MockJenkinsService) GetJobParameters(ctx context.Context, jaasName, jobName string) (interface{}, error) {
	if m.GetJobParametersFunc != nil {
		return m.GetJobParametersFunc(ctx, jaasName, jobName)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockJenkinsService) TriggerJob(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
	if m.TriggerJobFunc != nil {
		return m.TriggerJobFunc(ctx, jaasName, jobName, parameters)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockJenkinsService) GetQueueItemStatus(ctx context.Context, jaasName, queueItemID string) (*service.JenkinsQueueStatusResult, error) {
	if m.GetQueueItemStatusFunc != nil {
		return m.GetQueueItemStatusFunc(ctx, jaasName, queueItemID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockJenkinsService) GetBuildStatus(ctx context.Context, jaasName, jobName string, buildNumber int) (*service.JenkinsBuildStatusResult, error) {
	if m.GetBuildStatusFunc != nil {
		return m.GetBuildStatusFunc(ctx, jaasName, jobName, buildNumber)
	}
	return nil, fmt.Errorf("not implemented")
}

// TestGetJobParameters_Success tests successful parameter retrieval
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_Success() {
	mockService := &MockJenkinsService{
		GetJobParametersFunc: func(ctx context.Context, jaasName, jobName string) (interface{}, error) {
			assert.Equal(suite.T(), "cfsmc", jaasName)
			assert.Equal(suite.T(), "test-job", jobName)

			return map[string]interface{}{
				"parameterDefinitions": []map[string]interface{}{
					{
						"_class": "hudson.model.BooleanParameterDefinition",
						"name":   "DELETE_CLUSTER",
						"type":   "BooleanParameterDefinition",
						"defaultParameterValue": map[string]interface{}{
							"_class": "hudson.model.BooleanParameterValue",
							"value":  false,
						},
						"description": "WARNING!!! If checked, your cluster data will be completely deleted",
					},
				},
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/parameters", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "parameterDefinitions")
}

// TestGetJobParameters_MissingJaasName tests missing jaasName parameter
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_MissingJaasName() {
	mockService := &MockJenkinsService{}
	handler := handlers.NewJenkinsHandler(mockService)

	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

	// Request with empty jaasName
	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins//test-job/parameters", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Gin returns 400 for empty path parameters
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestGetJobParameters_MissingJobName tests missing jobName parameter
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_MissingJobName() {
	mockService := &MockJenkinsService{}
	handler := handlers.NewJenkinsHandler(mockService)

	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

	// Request with empty jobName
	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc//parameters", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Gin returns 400 for empty path parameters
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestGetJobParameters_ServiceError tests service error handling
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_ServiceError() {
	mockService := &MockJenkinsService{
		GetJobParametersFunc: func(ctx context.Context, jaasName, jobName string) (interface{}, error) {
			return nil, fmt.Errorf("jenkins credentials not found")
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/parameters", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "jenkins request failed")
}

// TestGetJobParameters_JenkinsNotFound tests 404 from Jenkins
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_JenkinsNotFound() {
	mockService := &MockJenkinsService{
		GetJobParametersFunc: func(ctx context.Context, jaasName, jobName string) (interface{}, error) {
			return nil, fmt.Errorf("jenkins request failed: status=404")
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/nonexistent-job/parameters", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "404")
}

// TestTriggerJob_Success tests successful job triggering
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_Success() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			assert.Equal(suite.T(), "cfsmc", jaasName)
			assert.Equal(suite.T(), "test-job", jobName)
			assert.Equal(suite.T(), "main", parameters["BRANCH"])
			assert.Equal(suite.T(), "staging", parameters["ENVIRONMENT"])
			return &service.JenkinsTriggerResult{
				Status:      "queued",
				Message:     "Job successfully queued in Jenkins",
				QueueURL:    "https://cfsmc.jaas-gcp.cloud.sap.corp/queue/item/12345/",
				QueueItemID: "12345",
				BaseJobURL:  "https://cfsmc.jaas-gcp.cloud.sap.corp/job/test-job",
				JobName:     "test-job",
				JaasName:    "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	requestBody := map[string]string{
		"BRANCH":      "main",
		"ENVIRONMENT": "staging",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.JenkinsTriggerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "queued", response.Status)
	assert.Equal(suite.T(), "Job successfully queued in Jenkins", response.Message)
	assert.Equal(suite.T(), "12345", response.QueueItemID)
	assert.Equal(suite.T(), "test-job", response.JobName)
	assert.Equal(suite.T(), "cfsmc", response.JaasName)
}

// TestTriggerJob_WithoutParameters tests triggering without parameters
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_WithoutParameters() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			assert.Equal(suite.T(), "cfsmc", jaasName)
			assert.Equal(suite.T(), "test-job", jobName)
			assert.Empty(suite.T(), parameters)
			return &service.JenkinsTriggerResult{
				Status:   "queued",
				Message:  "Job successfully queued in Jenkins",
				JobName:  "test-job",
				JaasName: "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	// Send empty body
	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestTriggerJob_EmptyParameters tests triggering with empty parameter object
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_EmptyParameters() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			assert.Empty(suite.T(), parameters)
			return &service.JenkinsTriggerResult{
				Status:   "queued",
				Message:  "Job successfully queued in Jenkins",
				JobName:  "test-job",
				JaasName: "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	requestBody := map[string]string{}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestTriggerJob_ServiceError tests service error handling
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_ServiceError() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			return nil, fmt.Errorf("jenkins credentials not found")
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "jenkins trigger failed")
}

// TestTriggerJob_InvalidJSON tests invalid JSON in request body
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_InvalidJSON() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			// Should still be called with empty parameters
			assert.Empty(suite.T(), parameters)
			return &service.JenkinsTriggerResult{
				Status:   "queued",
				Message:  "Job successfully queued in Jenkins",
				JobName:  "test-job",
				JaasName: "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	// Invalid JSON
	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should succeed with empty parameters (handler gracefully handles invalid JSON)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestTriggerJob_MissingJaasName tests missing jaasName parameter
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_MissingJaasName() {
	mockService := &MockJenkinsService{}
	handler := handlers.NewJenkinsHandler(mockService)

	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins//test-job/trigger", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Gin returns 400 for empty path parameters
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestTriggerJob_MissingJobName tests missing jobName parameter
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_MissingJobName() {
	mockService := &MockJenkinsService{}
	handler := handlers.NewJenkinsHandler(mockService)

	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc//trigger", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Gin returns 400 for empty path parameters
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestNewJenkinsHandler tests handler creation
func (suite *JenkinsHandlerTestSuite) TestNewJenkinsHandler() {
	mockService := &MockJenkinsService{}
	handler := handlers.NewJenkinsHandler(mockService)

	assert.NotNil(suite.T(), handler)
}

// TestGetJobParameters_DifferentJaasInstances tests different JAAS instances
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_DifferentJaasInstances() {
	testCases := []string{"cfsmc", "test-instance", "another-jaas"}

	for _, jaasName := range testCases {
		suite.T().Run(jaasName, func(t *testing.T) {
			mockService := &MockJenkinsService{
				GetJobParametersFunc: func(ctx context.Context, jn, job string) (interface{}, error) {
					assert.Equal(t, jaasName, jn)
					return map[string]interface{}{}, nil
				},
			}

			handler := handlers.NewJenkinsHandler(mockService)
			router := gin.New()
			router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/self-service/jenkins/%s/test-job/parameters", jaasName), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestGetJobParameters_DifferentJobNames tests different job names
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_DifferentJobNames() {
	testCases := []struct {
		name       string
		urlJobName string
		expected   string
	}{
		{"simple-job", "simple-job", "simple-job"},
		{"job-with-hyphens", "job-with-hyphens", "job-with-hyphens"},
		{"job_with_underscores", "job_with_underscores", "job_with_underscores"},
		{"JOB_WITH_CAPS", "JOB_WITH_CAPS", "JOB_WITH_CAPS"},
		// Note: slashes in job names need special handling in URL paths
		// In real usage, they would be URL-encoded or handled differently
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			mockService := &MockJenkinsService{
				GetJobParametersFunc: func(ctx context.Context, jaas, job string) (interface{}, error) {
					assert.Equal(t, tc.expected, job)
					return map[string]interface{}{}, nil
				},
			}

			handler := handlers.NewJenkinsHandler(mockService)
			router := gin.New()
			router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/self-service/jenkins/cfsmc/%s/parameters", tc.urlJobName), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestTriggerJob_MultipleParameters tests triggering with multiple parameters
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_MultipleParameters() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			assert.Len(suite.T(), parameters, 5)
			assert.Equal(suite.T(), "value1", parameters["PARAM1"])
			assert.Equal(suite.T(), "value2", parameters["PARAM2"])
			return &service.JenkinsTriggerResult{
				Status:   "queued",
				Message:  "Job successfully queued in Jenkins",
				JobName:  "test-job",
				JaasName: "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	requestBody := map[string]string{
		"PARAM1":  "value1",
		"PARAM2":  "value2",
		"PARAM3":  "value3",
		"BRANCH":  "main",
		"VERSION": "1.0.0",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestTriggerJob_SpecialCharactersInParameters tests special characters in parameter values
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_SpecialCharactersInParameters() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			assert.Equal(suite.T(), "value with spaces", parameters["PARAM1"])
			assert.Equal(suite.T(), "value@#$%", parameters["PARAM2"])
			return &service.JenkinsTriggerResult{
				Status:   "queued",
				Message:  "Job successfully queued in Jenkins",
				JobName:  "test-job",
				JaasName: "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	requestBody := map[string]string{
		"PARAM1": "value with spaces",
		"PARAM2": "value@#$%",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestTriggerJob_MixedParameterTypes tests triggering with mixed value types (string, bool, number)
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_MixedParameterTypes() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			// Verify all types are converted to strings
			assert.Equal(suite.T(), "I572719", parameters["ClusterName"])
			assert.Equal(suite.T(), "", parameters["DELETE_CLUSTER"])
			assert.Equal(suite.T(), "true", parameters["FETCH_STAGING_VERSION"])
			assert.Equal(suite.T(), "DEPLOY_CLUSTER", parameters["DEPLOYMENT_OPTION"])
			assert.Equal(suite.T(), "None", parameters["GIT_ORG_REPO"])
			assert.Equal(suite.T(), "", parameters["DEPLOY_PERFORMANCE_MONITORING"])
			return &service.JenkinsTriggerResult{
				Status:   "queued",
				Message:  "Job successfully queued in Jenkins",
				JobName:  "test-job",
				JaasName: "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	// Send mixed types: string, bool, empty string
	requestBody := map[string]interface{}{
		"ClusterName":                  "I572719",
		"DELETE_CLUSTER":               "",
		"FETCH_STAGING_VERSION":        true, // Boolean
		"DEPLOYMENT_OPTION":            "DEPLOY_CLUSTER",
		"GIT_ORG_REPO":                 "None",
		"DEPLOY_PERFORMANCE_MONITORING": "",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestTriggerJob_NumericParameters tests triggering with numeric values
func (suite *JenkinsHandlerTestSuite) TestTriggerJob_NumericParameters() {
	mockService := &MockJenkinsService{
		TriggerJobFunc: func(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*service.JenkinsTriggerResult, error) {
			// Verify numbers are converted to strings
			assert.Equal(suite.T(), "42", parameters["INTEGER_PARAM"])
			assert.Equal(suite.T(), "3.14", parameters["FLOAT_PARAM"])
			assert.Equal(suite.T(), "false", parameters["BOOL_PARAM"])
			return &service.JenkinsTriggerResult{
				Status:   "queued",
				Message:  "Job successfully queued in Jenkins",
				JobName:  "test-job",
				JaasName: "cfsmc",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.POST("/self-service/jenkins/:jaasName/:jobName/trigger", handler.TriggerJob)

	requestBody := map[string]interface{}{
		"INTEGER_PARAM": 42,
		"FLOAT_PARAM":   3.14,
		"BOOL_PARAM":    false,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/self-service/jenkins/cfsmc/test-job/trigger", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestGetJobParameters_ComplexResponse tests complex Jenkins response
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_ComplexResponse() {
	mockService := &MockJenkinsService{
		GetJobParametersFunc: func(ctx context.Context, jaasName, jobName string) (interface{}, error) {
			return map[string]interface{}{
				"parameterDefinitions": []map[string]interface{}{
					{
						"_class": "hudson.model.StringParameterDefinition",
						"name":   "STRING_PARAM",
						"type":   "StringParameterDefinition",
						"defaultParameterValue": map[string]interface{}{
							"value": "default",
						},
						"description": "A string parameter",
					},
					{
						"_class":  "hudson.model.ChoiceParameterDefinition",
						"name":    "CHOICE_PARAM",
						"type":    "ChoiceParameterDefinition",
						"choices": []string{"option1", "option2", "option3"},
						"defaultParameterValue": map[string]interface{}{
							"value": "option1",
						},
					},
					{
						"_class": "hudson.model.BooleanParameterDefinition",
						"name":   "BOOLEAN_PARAM",
						"type":   "BooleanParameterDefinition",
						"defaultParameterValue": map[string]interface{}{
							"value": true,
						},
					},
				},
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/parameters", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "parameterDefinitions")
}

// TestGetJobParameters_EmptyResponse tests empty Jenkins response
func (suite *JenkinsHandlerTestSuite) TestGetJobParameters_EmptyResponse() {
	mockService := &MockJenkinsService{
		GetJobParametersFunc: func(ctx context.Context, jaasName, jobName string) (interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/parameters", handler.GetJobParameters)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/parameters", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
}

// TestGetQueueItemStatus_Success tests successful queue status retrieval
func (suite *JenkinsHandlerTestSuite) TestGetQueueItemStatus_Success() {
	buildNum := 123
	mockService := &MockJenkinsService{
		GetQueueItemStatusFunc: func(ctx context.Context, jaasName, queueItemID string) (*service.JenkinsQueueStatusResult, error) {
			assert.Equal(suite.T(), "cfsmc", jaasName)
			assert.Equal(suite.T(), "12345", queueItemID)
			return &service.JenkinsQueueStatusResult{
				Status:       "queued",
				BuildNumber:  &buildNum,
				BuildURL:     "https://cfsmc.jaas-gcp.cloud.sap.corp/job/test-job/123/",
				QueuedReason: "Waiting for executor",
				WaitTime:     45,
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/queue/:queueItemId/status", handler.GetQueueItemStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/queue/12345/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.JenkinsQueueStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "queued", response.Status)
	assert.NotNil(suite.T(), response.BuildNumber)
	assert.Equal(suite.T(), buildNum, *response.BuildNumber)
	assert.Equal(suite.T(), "Waiting for executor", response.QueuedReason)
	assert.Equal(suite.T(), 45, response.WaitTime)
}

// TestGetQueueItemStatus_NotFound tests queue item not found
func (suite *JenkinsHandlerTestSuite) TestGetQueueItemStatus_NotFound() {
	mockService := &MockJenkinsService{
		GetQueueItemStatusFunc: func(ctx context.Context, jaasName, queueItemID string) (*service.JenkinsQueueStatusResult, error) {
			return nil, fmt.Errorf("jenkins queue item not found: queue item 99999")
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/queue/:queueItemId/status", handler.GetQueueItemStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/queue/99999/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestGetQueueItemStatus_ServiceError tests service error handling
func (suite *JenkinsHandlerTestSuite) TestGetQueueItemStatus_ServiceError() {
	mockService := &MockJenkinsService{
		GetQueueItemStatusFunc: func(ctx context.Context, jaasName, queueItemID string) (*service.JenkinsQueueStatusResult, error) {
			return nil, fmt.Errorf("service error")
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/queue/:queueItemId/status", handler.GetQueueItemStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/queue/12345/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)
}

// TestGetBuildStatus_Success tests successful build status retrieval
func (suite *JenkinsHandlerTestSuite) TestGetBuildStatus_Success() {
	mockService := &MockJenkinsService{
		GetBuildStatusFunc: func(ctx context.Context, jaasName, jobName string, buildNumber int) (*service.JenkinsBuildStatusResult, error) {
			assert.Equal(suite.T(), "cfsmc", jaasName)
			assert.Equal(suite.T(), "test-job", jobName)
			assert.Equal(suite.T(), 42, buildNumber)
			return &service.JenkinsBuildStatusResult{
				Status:   "success",
				Result:   "SUCCESS",
				Building: false,
				Duration: 120000, // milliseconds
				BuildURL: "https://cfsmc.jaas-gcp.cloud.sap.corp/job/test-job/42/",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/:buildNumber/status", handler.GetBuildStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/42/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.JenkinsBuildStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "success", response.Status)
	assert.Equal(suite.T(), "SUCCESS", response.Result)
	assert.False(suite.T(), response.Building)
	assert.Equal(suite.T(), 120, response.Duration) // converted to seconds
	assert.Contains(suite.T(), response.BuildURL, "/job/test-job/42/")
}

// TestGetBuildStatus_Running tests running build
func (suite *JenkinsHandlerTestSuite) TestGetBuildStatus_Running() {
	mockService := &MockJenkinsService{
		GetBuildStatusFunc: func(ctx context.Context, jaasName, jobName string, buildNumber int) (*service.JenkinsBuildStatusResult, error) {
			return &service.JenkinsBuildStatusResult{
				Status:   "running",
				Result:   "",
				Building: true,
				Duration: 0,
				BuildURL: "https://cfsmc.jaas-gcp.cloud.sap.corp/job/test-job/42/",
			}, nil
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/:buildNumber/status", handler.GetBuildStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/42/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.JenkinsBuildStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "running", response.Status)
	assert.True(suite.T(), response.Building)
}

// TestGetBuildStatus_InvalidBuildNumber tests invalid build number
func (suite *JenkinsHandlerTestSuite) TestGetBuildStatus_InvalidBuildNumber() {
	mockService := &MockJenkinsService{}
	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/:buildNumber/status", handler.GetBuildStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/invalid/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestGetBuildStatus_NotFound tests build not found
func (suite *JenkinsHandlerTestSuite) TestGetBuildStatus_NotFound() {
	mockService := &MockJenkinsService{
		GetBuildStatusFunc: func(ctx context.Context, jaasName, jobName string, buildNumber int) (*service.JenkinsBuildStatusResult, error) {
			return nil, fmt.Errorf("jenkins build not found: job test-job build #99999")
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/:buildNumber/status", handler.GetBuildStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/99999/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestGetBuildStatus_ServiceError tests service error handling
func (suite *JenkinsHandlerTestSuite) TestGetBuildStatus_ServiceError() {
	mockService := &MockJenkinsService{
		GetBuildStatusFunc: func(ctx context.Context, jaasName, jobName string, buildNumber int) (*service.JenkinsBuildStatusResult, error) {
			return nil, fmt.Errorf("service error")
		},
	}

	handler := handlers.NewJenkinsHandler(mockService)
	suite.router.GET("/self-service/jenkins/:jaasName/:jobName/:buildNumber/status", handler.GetBuildStatus)

	req, _ := http.NewRequest(http.MethodGet, "/self-service/jenkins/cfsmc/test-job/42/status", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)
}

// Run the test suite
func TestJenkinsHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(JenkinsHandlerTestSuite))
}

// TestJenkinsHandlerCreation tests basic handler creation
func TestJenkinsHandlerCreation(t *testing.T) {
	mockService := &MockJenkinsService{}
	handler := handlers.NewJenkinsHandler(mockService)
	assert.NotNil(t, handler)
}
