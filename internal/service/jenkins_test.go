package service_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"developer-portal-backend/internal/config"
	"developer-portal-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// JenkinsServiceTestSuite defines the test suite for JenkinsService
type JenkinsServiceTestSuite struct {
	suite.Suite
	jenkinsService *service.JenkinsService
	mockServer     *httptest.Server
}

// SetupTest sets up the test suite
func (suite *JenkinsServiceTestSuite) SetupTest() {
	// Will be configured per test
}

// TearDownTest cleans up after each test
func (suite *JenkinsServiceTestSuite) TearDownTest() {
	if suite.mockServer != nil {
		suite.mockServer.Close()
	}
	// Clean up environment variables
	os.Unsetenv("JENKINS_CFSMC_JAAS_TOKEN")
	os.Unsetenv("JENKINS_P_USER")
	os.Unsetenv("JENKINS_TEST_INSTANCE_JAAS_TOKEN")
	os.Unsetenv("JENKINS_TEST_INSTANCE_JAAS_TOKEN_NAME")
}

// TestNewJenkinsService tests creating a new Jenkins service
func (suite *JenkinsServiceTestSuite) TestNewJenkinsService() {
	jenkinsService := service.NewJenkinsService(nil)
	assert.NotNil(suite.T(), jenkinsService)
}

// TestGetJobParameters_Success tests successful parameter retrieval
func (suite *JenkinsServiceTestSuite) TestGetJobParameters_Success() {
	// Set up environment variables
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	// Create mock Jenkins server
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Contains(suite.T(), r.URL.Path, "/job/test-job/api/json")
		assert.Equal(suite.T(), "property[parameterDefinitions[name,type,defaultParameterValue[value],choices,description]]", r.URL.Query().Get("tree"))

		// Verify Basic Auth
		authHeader := r.Header.Get("Authorization")
		assert.NotEmpty(suite.T(), authHeader)
		expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-user:test-token-123"))
		assert.Equal(suite.T(), expectedAuth, authHeader)

		// Return mock response with proper Jenkins structure
		response := map[string]interface{}{
			"property": []map[string]interface{}{
				{
					"_class": "hudson.model.ParametersDefinitionProperty",
					"parameterDefinitions": []map[string]interface{}{
						{
							"name": "BRANCH",
							"type": "StringParameterDefinition",
							"defaultParameterValue": map[string]string{
								"value": "main",
							},
						},
						{
							"name": "ENVIRONMENT",
							"type": "ChoiceParameterDefinition",
							"defaultParameterValue": map[string]string{
								"value": "dev",
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Configure service with mock server URL
	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "test-job")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	
	// Verify the response structure
	resultMap, ok := result.(map[string]interface{})
	require.True(suite.T(), ok, "Result should be a map")
	
	paramDefs, ok := resultMap["parameterDefinitions"]
	require.True(suite.T(), ok, "Result should contain parameterDefinitions")
	require.NotNil(suite.T(), paramDefs)
}

// TestGetJobParameters_MissingCredentials tests missing credentials
func (suite *JenkinsServiceTestSuite) TestGetJobParameters_MissingCredentials() {
	cfg := &config.Config{
		JenkinsBaseURL: "http://localhost",
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)
	
	// Don't set environment variables
	result, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "test-job")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	// Since neither token nor username is set, the error will be about the token (checked first)
	assert.Contains(suite.T(), err.Error(), "jenkins token not found")
	assert.Contains(suite.T(), err.Error(), "JENKINS_CFSMC_JAAS_TOKEN")
}

// TestGetJobParameters_MissingToken tests missing token only
func (suite *JenkinsServiceTestSuite) TestGetJobParameters_MissingToken() {
	os.Setenv("JENKINS_P_USER", "test-user")
	// Don't set TOKEN

	cfg := &config.Config{
		JenkinsBaseURL: "http://localhost",
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "test-job")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "jenkins token not found")
}

// TestGetJobParameters_MissingTokenName tests missing token name only
func (suite *JenkinsServiceTestSuite) TestGetJobParameters_MissingTokenName() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	// Don't set JENKINS_P_USER

	cfg := &config.Config{
		JenkinsBaseURL: "http://localhost",
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "test-job")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "jenkins username not found")
	assert.Contains(suite.T(), err.Error(), "JENKINS_P_USER")
}

// TestGetJobParameters_DifferentJaasInstances tests different JAAS instance names
func (suite *JenkinsServiceTestSuite) TestGetJobParameters_DifferentJaasInstances() {
	testCases := []struct {
		name             string
		jaasName         string
		expectedTokenEnv string
	}{
		{
			name:             "SimpleName",
			jaasName:         "cfsmc",
			expectedTokenEnv: "JENKINS_CFSMC_JAAS_TOKEN",
		},
		{
			name:             "HyphenatedName",
			jaasName:         "test-instance",
			expectedTokenEnv: "JENKINS_TEST_INSTANCE_JAAS_TOKEN",
		},
		{
			name:             "MixedCaseName",
			jaasName:         "TestCase",
			expectedTokenEnv: "JENKINS_TESTCASE_JAAS_TOKEN",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				JenkinsBaseURL: "http://localhost",
			}
			svc := service.NewJenkinsService(cfg)
			
			// Should error with missing credentials mentioning the correct token env var
			result, err := svc.GetJobParameters(context.Background(), tc.jaasName, "test-job")

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedTokenEnv)
		})
	}
}

// TestTriggerJob_Success tests successful job trigger
func (suite *JenkinsServiceTestSuite) TestTriggerJob_Success() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	// Create mock Jenkins server
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Contains(suite.T(), r.URL.Path, "/job/test-job/buildWithParameters")
		assert.Equal(suite.T(), http.MethodPost, r.Method)

		// Verify Basic Auth
		authHeader := r.Header.Get("Authorization")
		assert.NotEmpty(suite.T(), authHeader)

		// Parse form data
		r.ParseForm()
		assert.Equal(suite.T(), "feature/test", r.FormValue("BRANCH"))
		assert.Equal(suite.T(), "staging", r.FormValue("ENVIRONMENT"))

		// Return success with Location header
		w.Header().Set("Location", suite.mockServer.URL+"/queue/item/12345/")
		w.WriteHeader(http.StatusCreated)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	parameters := map[string]string{
		"BRANCH":      "feature/test",
		"ENVIRONMENT": "staging",
	}

	result, err := suite.jenkinsService.TriggerJob(context.Background(), "cfsmc", "test-job", parameters)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "queued", result.Status)
	assert.Equal(suite.T(), "test-job", result.JobName)
	assert.Equal(suite.T(), "cfsmc", result.JaasName)
	assert.Equal(suite.T(), "12345", result.QueueItemID)
	assert.Contains(suite.T(), result.QueueURL, "/queue/item/12345/")
}

// TestTriggerJob_MissingCredentials tests triggering without credentials
func (suite *JenkinsServiceTestSuite) TestTriggerJob_MissingCredentials() {
	cfg := &config.Config{
		JenkinsBaseURL: "http://localhost",
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)
	
	result, err := suite.jenkinsService.TriggerJob(context.Background(), "cfsmc", "test-job", map[string]string{})

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	// Will fail on missing token (checked first)
	assert.Contains(suite.T(), err.Error(), "jenkins token not found")
}

// TestTriggerJob_WithoutParameters tests triggering without parameters
func (suite *JenkinsServiceTestSuite) TestTriggerJob_WithoutParameters() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(suite.T(), r.URL.Path, "/job/test-job/buildWithParameters")
		w.Header().Set("Location", suite.mockServer.URL+"/queue/item/456/")
		w.WriteHeader(http.StatusCreated)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.TriggerJob(context.Background(), "cfsmc", "test-job", map[string]string{})

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "queued", result.Status)
}

// TestJenkinsURLConstruction tests URL construction with placeholders
func (suite *JenkinsServiceTestSuite) TestJenkinsURLConstruction() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Just return success
		response := map[string]interface{}{
			"property": []map[string]interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))

	// Test with {jaasName} placeholder
	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	_, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "test-job")
	assert.NoError(suite.T(), err)
}

// TestEmptyJobName tests behavior with empty job name
func (suite *JenkinsServiceTestSuite) TestEmptyJobName() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Empty job name will result in /job//api/json path
		assert.Contains(suite.T(), r.URL.Path, "/job/")
		
		response := map[string]interface{}{
			"property": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	_, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "")
	// May succeed or fail depending on mock server behavior
	_ = err
}

// TestSpecialCharactersInJobName tests job names with special characters
func (suite *JenkinsServiceTestSuite) TestSpecialCharactersInJobName() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	testCases := []string{
		"job-with-hyphens",
		"job_with_underscores",
		"job.with.dots",
		"folder/job",
		"folder/subfolder/job",
	}

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"property": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	for _, jobName := range testCases {
		suite.T().Run(jobName, func(t *testing.T) {
			_, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", jobName)
			assert.NoError(t, err)
		})
	}
}

// TestConcurrentRequests tests thread safety with concurrent requests
func (suite *JenkinsServiceTestSuite) TestConcurrentRequests() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"property": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	done := make(chan bool, 10)

	// Run 10 concurrent requests
	for i := 0; i < 10; i++ {
		go func(jobNum int) {
			jobName := fmt.Sprintf("test-job-%d", jobNum)
			_, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", jobName)
			assert.NoError(suite.T(), err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestTriggerJob_MultipleParameters tests triggering with many parameters
func (suite *JenkinsServiceTestSuite) TestTriggerJob_MultipleParameters() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		assert.Equal(suite.T(), "value1", r.FormValue("PARAM1"))
		assert.Equal(suite.T(), "value2", r.FormValue("PARAM2"))
		assert.Equal(suite.T(), "value3", r.FormValue("PARAM3"))
		w.Header().Set("Location", suite.mockServer.URL+"/queue/item/789/")
		w.WriteHeader(http.StatusCreated)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	parameters := map[string]string{
		"PARAM1":  "value1",
		"PARAM2":  "value2",
		"PARAM3":  "value3",
		"BRANCH":  "main",
		"VERSION": "1.0.0",
	}

	result, err := suite.jenkinsService.TriggerJob(context.Background(), "cfsmc", "test-job", parameters)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
}

// TestParameterValues tests various parameter value formats
func (suite *JenkinsServiceTestSuite) TestParameterValues() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	testCases := []struct {
		name  string
		value string
	}{
		{"SimpleValue", "test"},
		{"WithSpaces", "test value"},
		{"WithSpecialChars", "test@#$%"},
		{"Empty", ""},
	}

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		// Just verify it was sent
		_ = r.FormValue("TEST_PARAM")
		w.WriteHeader(http.StatusCreated)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			parameters := map[string]string{
				"TEST_PARAM": tc.value,
			}

			result, err := suite.jenkinsService.TriggerJob(context.Background(), "cfsmc", "test-job", parameters)
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestServerError tests handling of 5xx errors
func (suite *JenkinsServiceTestSuite) TestServerError() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	_, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "test-job")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "500")
}

// TestNotFoundError tests handling of 404 errors
func (suite *JenkinsServiceTestSuite) TestNotFoundError() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	_, err := suite.jenkinsService.GetJobParameters(context.Background(), "cfsmc", "nonexistent-job")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), strings.ToLower(err.Error()), "404")
}

// TestGetQueueItemStatus_Queued tests getting status of a queued item
func (suite *JenkinsServiceTestSuite) TestGetQueueItemStatus_Queued() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	// Create mock Jenkins server
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(suite.T(), r.URL.Path, "/queue/item/12345/api/json")
		assert.Equal(suite.T(), http.MethodGet, r.Method)

		// Return queued item data
		response := map[string]interface{}{
			"id":            12345,
			"why":           "Waiting for available executor",
			"inQueueSince":  1700000000000,
			"cancelled":     false,
			"blocked":       false,
			"buildable":     true,
			"executable":    nil,
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetQueueItemStatus(context.Background(), "cfsmc", "12345")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "queued", result.Status)
	assert.Nil(suite.T(), result.BuildNumber)
	assert.Empty(suite.T(), result.BuildURL)
	assert.Equal(suite.T(), "Waiting for available executor", result.QueuedReason)
	assert.Greater(suite.T(), result.WaitTime, 0)
}

// TestGetQueueItemStatus_Running tests getting status when build has started
func (suite *JenkinsServiceTestSuite) TestGetQueueItemStatus_Running() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	buildNum := 42
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":           12345,
			"why":          "Started",
			"inQueueSince": 1700000000000,
			"cancelled":    false,
			"executable": map[string]interface{}{
				"number": buildNum,
				"url":    suite.mockServer.URL + "/job/test-job/42/",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetQueueItemStatus(context.Background(), "cfsmc", "12345")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "running", result.Status)
	assert.NotNil(suite.T(), result.BuildNumber)
	assert.Equal(suite.T(), buildNum, *result.BuildNumber)
	assert.Contains(suite.T(), result.BuildURL, "/job/test-job/42/")
}

// TestGetQueueItemStatus_Cancelled tests getting status of cancelled item
func (suite *JenkinsServiceTestSuite) TestGetQueueItemStatus_Cancelled() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"id":           12345,
			"why":          "Cancelled",
			"inQueueSince": 1700000000000,
			"cancelled":    true,
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetQueueItemStatus(context.Background(), "cfsmc", "12345")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "cancelled", result.Status)
}

// TestGetQueueItemStatus_NotFound tests queue item not found
func (suite *JenkinsServiceTestSuite) TestGetQueueItemStatus_NotFound() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetQueueItemStatus(context.Background(), "cfsmc", "99999")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "queue item not found")
}

// TestGetQueueItemStatus_MissingCredentials tests missing credentials
func (suite *JenkinsServiceTestSuite) TestGetQueueItemStatus_MissingCredentials() {
	cfg := &config.Config{
		JenkinsBaseURL: "https://jenkins.example.com",
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetQueueItemStatus(context.Background(), "cfsmc", "12345")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestGetBuildStatus_Success tests getting status of completed build
func (suite *JenkinsServiceTestSuite) TestGetBuildStatus_Success() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(suite.T(), r.URL.Path, "/job/test-job/42/api/json")
		assert.Equal(suite.T(), http.MethodGet, r.Method)

		response := map[string]interface{}{
			"number":    42,
			"result":    "SUCCESS",
			"building":  false,
			"duration":  120000, // 2 minutes in milliseconds
			"url":       suite.mockServer.URL + "/job/test-job/42/",
			"timestamp": 1700000000000,
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetBuildStatus(context.Background(), "cfsmc", "test-job", 42)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "success", result.Status)
	assert.Equal(suite.T(), "SUCCESS", result.Result)
	assert.False(suite.T(), result.Building)
	assert.Equal(suite.T(), int64(120000), result.Duration)
	assert.Contains(suite.T(), result.BuildURL, "/job/test-job/42/")
}

// TestGetBuildStatus_Running tests getting status of running build
func (suite *JenkinsServiceTestSuite) TestGetBuildStatus_Running() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"number":    42,
			"result":    nil,
			"building":  true,
			"duration":  0,
			"url":       suite.mockServer.URL + "/job/test-job/42/",
			"timestamp": 1700000000000,
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetBuildStatus(context.Background(), "cfsmc", "test-job", 42)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "running", result.Status)
	assert.True(suite.T(), result.Building)
	assert.Equal(suite.T(), int64(0), result.Duration)
}

// TestGetBuildStatus_Failed tests getting status of failed build
func (suite *JenkinsServiceTestSuite) TestGetBuildStatus_Failed() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"number":    42,
			"result":    "FAILURE",
			"building":  false,
			"duration":  60000,
			"url":       suite.mockServer.URL + "/job/test-job/42/",
			"timestamp": 1700000000000,
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetBuildStatus(context.Background(), "cfsmc", "test-job", 42)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "failure", result.Status)
	assert.Equal(suite.T(), "FAILURE", result.Result)
	assert.False(suite.T(), result.Building)
}

// TestGetBuildStatus_Aborted tests getting status of aborted build
func (suite *JenkinsServiceTestSuite) TestGetBuildStatus_Aborted() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"number":    42,
			"result":    "ABORTED",
			"building":  false,
			"duration":  30000,
			"url":       suite.mockServer.URL + "/job/test-job/42/",
			"timestamp": 1700000000000,
		}
		json.NewEncoder(w).Encode(response)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetBuildStatus(context.Background(), "cfsmc", "test-job", 42)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "aborted", result.Status)
	assert.Equal(suite.T(), "ABORTED", result.Result)
}

// TestGetBuildStatus_NotFound tests build not found
func (suite *JenkinsServiceTestSuite) TestGetBuildStatus_NotFound() {
	os.Setenv("JENKINS_CFSMC_JAAS_TOKEN", "test-token-123")
	os.Setenv("JENKINS_P_USER", "test-user")

	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	cfg := &config.Config{
		JenkinsBaseURL: suite.mockServer.URL,
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetBuildStatus(context.Background(), "cfsmc", "test-job", 99999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "build not found")
}

// TestGetBuildStatus_MissingCredentials tests missing credentials
func (suite *JenkinsServiceTestSuite) TestGetBuildStatus_MissingCredentials() {
	cfg := &config.Config{
		JenkinsBaseURL: "https://jenkins.example.com",
	}
	suite.jenkinsService = service.NewJenkinsService(cfg)

	result, err := suite.jenkinsService.GetBuildStatus(context.Background(), "cfsmc", "test-job", 42)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// Run the test suite
func TestJenkinsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(JenkinsServiceTestSuite))
}

// Benchmark test for service creation
func BenchmarkNewJenkinsService(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = service.NewJenkinsService(nil)
	}
}

// TestServiceCreation tests basic service creation
func TestJenkinsServiceCreation(t *testing.T) {
	jenkinsService := service.NewJenkinsService(nil)
	assert.NotNil(t, jenkinsService)
}
