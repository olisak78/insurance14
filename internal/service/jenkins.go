package service

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"developer-portal-backend/internal/config"
	"developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/logger"
)

// JenkinsService provides methods to interact with Jenkins JAAS instances
type JenkinsService struct {
	cfg        *config.Config
	httpClient *http.Client
}

// JenkinsTriggerResult holds the result of triggering a Jenkins job
type JenkinsTriggerResult struct {
	Status      string `json:"status"`      // "queued", "triggered", "failed"
	Message     string `json:"message"`     // Human-readable message
	QueueURL    string `json:"queueUrl"`    // URL to track the queued item (poll this to get build URL when job starts)
	QueueItemID string `json:"queueItemId"` // Queue item ID extracted from URL
	BaseJobURL  string `json:"baseJobUrl"`  // Base URL to the job definition (not specific build)
	JobName     string `json:"jobName"`     // Name of the triggered job
	JaasName    string `json:"jaasName"`    // JAAS instance name
}

// JenkinsQueueStatusResult holds the status of a queued Jenkins job
type JenkinsQueueStatusResult struct {
	Status       string `json:"status"`       // "queued", "running", "complete", "cancelled"
	BuildNumber  *int   `json:"buildNumber"`  // Build number if job has started (nullable)
	BuildURL     string `json:"buildUrl"`     // URL to the build if started
	QueuedReason string `json:"queuedReason"` // Reason item is in queue
	WaitTime     int    `json:"waitTime"`     // Time in seconds the item has been in queue
}

// JenkinsBuildStatusResult holds the status of a Jenkins build
type JenkinsBuildStatusResult struct {
	Status   string `json:"status"`   // "running", "success", "failure", "aborted", "unstable"
	Result   string `json:"result"`   // Jenkins result field (SUCCESS, FAILURE, UNSTABLE, ABORTED, null if still running)
	Building bool   `json:"building"` // Whether build is currently in progress
	Duration int64  `json:"duration"` // Duration in milliseconds (0 if still running)
	BuildURL string `json:"buildUrl"` // Full URL to the build
}

// NewJenkinsService creates a new Jenkins service
func NewJenkinsService(cfg *config.Config) *JenkinsService {
	// If no config provided, create empty config
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Set default Jenkins base URL if not provided
	if cfg.JenkinsBaseURL == "" {
		cfg.JenkinsBaseURL = "https://{jaasName}.jaas-gcp.cloud.sap.corp"
	}

	// Configure HTTP client with TLS settings
	// InsecureSkipVerify is set to true by default for SAP internal CAs
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.JenkinsInsecureSkipVerify,
		},
	}

	return &JenkinsService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// getJenkinsCredentials retrieves Jenkins credentials from environment variables
// Expected env vars: JENKINS_{JAASNAME_UPPERCASE}_JAAS_TOKEN (instance-specific) and JENKINS_P_USER (global username)
func (s *JenkinsService) getJenkinsCredentials(ctx context.Context, jaasName string) (tokenName, token string, err error) {
	log := logger.WithContext(ctx)
	jaasNameUpper := strings.ToUpper(strings.ReplaceAll(jaasName, "-", "_"))

	tokenEnvVar := fmt.Sprintf("JENKINS_%s_JAAS_TOKEN", jaasNameUpper)
	userEnvVar := "JENKINS_P_USER"

	token = os.Getenv(tokenEnvVar)
	tokenName = os.Getenv(userEnvVar)

	if token == "" {
		log.Errorf("Jenkins token not found: missing %s environment variable", tokenEnvVar)
		return "", "", fmt.Errorf("%w: missing %s environment variable", errors.ErrJenkinsTokenNotFound, tokenEnvVar)
	}

	if tokenName == "" {
		log.Errorf("Jenkins username not found: missing %s environment variable", userEnvVar)
		return "", "", fmt.Errorf("%w: missing %s environment variable", errors.ErrJenkinsUserNotFound, userEnvVar)
	}

	return tokenName, token, nil
}

// buildJenkinsURL constructs the Jenkins URL for a given JAAS instance and job
func (s *JenkinsService) buildJenkinsURL(jaasName, jobName string) string {
	// Replace {jaasName} placeholder in base URL
	baseURL := strings.Replace(s.cfg.JenkinsBaseURL, "{jaasName}", jaasName, -1)
	return fmt.Sprintf("%s/job/%s", baseURL, jobName)
}

// GetJobParameters retrieves the parameters definition for a Jenkins job
func (s *JenkinsService) GetJobParameters(ctx context.Context, jaasName, jobName string) (interface{}, error) {
	log := logger.WithContext(ctx).WithFields(map[string]interface{}{
		"jaasName": jaasName,
		"jobName":  jobName,
	})

	log.Infof("Getting job parameters for Jenkins job")

	tokenName, token, err := s.getJenkinsCredentials(ctx, jaasName)
	if err != nil {
		log.Errorf("Failed to get Jenkins credentials: %v", err)
		return nil, err
	}

	// Build the Jenkins API URL
	baseURL := s.buildJenkinsURL(jaasName, jobName)
	fullURL := fmt.Sprintf("%s/api/json?tree=property[parameterDefinitions[name,type,defaultParameterValue[value],choices,description]]", baseURL)

	log.Infof("Jenkins GET parameters request: url=%s", fullURL)

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	// Set Basic Auth header using token name as username and token as password
	cred := base64.StdEncoding.EncodeToString([]byte(tokenName + ":" + token))
	req.Header.Set("Authorization", "Basic "+cred)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Errorf("Jenkins GET parameters request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Infof("Jenkins GET parameters response: status=%d", resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Jenkins GET parameters failed: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("jenkins request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Decode the full JSON response
	var result map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		log.Errorf("Failed to decode Jenkins response: %v", err)
		return nil, err
	}

	// Filter to only return parameterDefinitions from hudson.model.ParametersDefinitionProperty
	filteredResult := s.filterParameterDefinitions(ctx, result)

	log.Info("Successfully retrieved job parameters")
	return filteredResult, nil
}

// filterParameterDefinitions extracts parameterDefinitions from hudson.model.ParametersDefinitionProperty
func (s *JenkinsService) filterParameterDefinitions(ctx context.Context, response map[string]interface{}) interface{} {
	// Get the property array
	properties, ok := response["property"].([]interface{})
	if !ok {
		return response
	}

	// Find the ParametersDefinitionProperty
	for _, prop := range properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this is the ParametersDefinitionProperty
		class, ok := propMap["_class"].(string)
		if !ok || class != "hudson.model.ParametersDefinitionProperty" {
			continue
		}

		// Extract parameterDefinitions
		paramDefs, ok := propMap["parameterDefinitions"]
		if !ok {
			return map[string]interface{}{
				"parameterDefinitions": []interface{}{},
			}
		}

		return map[string]interface{}{
			"parameterDefinitions": paramDefs,
		}
	}

	// If no ParametersDefinitionProperty found, return empty parameterDefinitions
	return map[string]interface{}{
		"parameterDefinitions": []interface{}{},
	}
}

// TriggerJob triggers a Jenkins job with the provided parameters
func (s *JenkinsService) TriggerJob(ctx context.Context, jaasName, jobName string, parameters map[string]string) (*JenkinsTriggerResult, error) {
	log := logger.WithContext(ctx).WithFields(map[string]interface{}{
		"jaasName":   jaasName,
		"jobName":    jobName,
		"paramCount": len(parameters),
	})

	log.Info("Triggering Jenkins job")

	tokenName, token, err := s.getJenkinsCredentials(ctx, jaasName)
	if err != nil {
		log.Errorf("Failed to get Jenkins credentials: %v", err)
		return nil, err
	}

	// Build the Jenkins trigger URL
	// Always use buildWithParameters endpoint - it works for both parameterized and non-parameterized jobs
	// Jenkins will use default values for any parameters not provided
	baseURL := s.buildJenkinsURL(jaasName, jobName)
	fullURL := fmt.Sprintf("%s/buildWithParameters", baseURL)

	if len(parameters) > 0 {
		log.Infof("Jenkins trigger job with %d parameter(s): url=%s", len(parameters), fullURL)
	} else {
		log.Infof("Jenkins trigger job with default parameters: url=%s", fullURL)
	}

	// Prepare form data (even if empty, Jenkins accepts empty form for defaults)
	formData := url.Values{}
	for key, value := range parameters {
		formData.Set(key, value)
	}

	// Always POST with form-encoded body (empty form is valid)
	req, err := http.NewRequest(http.MethodPost, fullURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set Basic Auth header using token name as username and token as password
	cred := base64.StdEncoding.EncodeToString([]byte(tokenName + ":" + token))
	req.Header.Set("Authorization", "Basic "+cred)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Errorf("Jenkins trigger job request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Infof("Jenkins trigger job response: status=%d", resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Jenkins trigger job failed: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("jenkins trigger failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Extract queue URL from Location header
	queueURL := resp.Header.Get("Location")
	queueItemID := ""

	// Parse queue item ID from URL (e.g., https://jenkins/queue/item/12345/ -> 12345)
	if queueURL != "" {
		parts := strings.Split(strings.TrimSuffix(queueURL, "/"), "/")
		if len(parts) > 0 {
			queueItemID = parts[len(parts)-1]
		}
		log.Infof("Job queued with ID: %s, queue URL: %s", queueItemID, queueURL)
	}

	// Build result
	result := &JenkinsTriggerResult{
		Status:      "queued",
		Message:     "Job successfully queued in Jenkins",
		QueueURL:    queueURL,
		QueueItemID: queueItemID,
		BaseJobURL:  baseURL,
		JobName:     jobName,
		JaasName:    jaasName,
	}

	log.Info("Successfully triggered Jenkins job")
	return result, nil
}

// GetQueueItemStatus retrieves the status of a queued Jenkins job
func (s *JenkinsService) GetQueueItemStatus(ctx context.Context, jaasName, queueItemID string) (*JenkinsQueueStatusResult, error) {
	log := logger.WithContext(ctx).WithFields(map[string]interface{}{
		"jaasName":    jaasName,
		"queueItemID": queueItemID,
	})

	log.Info("Getting Jenkins queue item status")

	// Get credentials
	tokenName, token, err := s.getJenkinsCredentials(ctx, jaasName)
	if err != nil {
		return nil, err
	}

	// Build queue item URL
	baseURL := strings.Replace(s.cfg.JenkinsBaseURL, "{jaasName}", jaasName, -1)
	queueURL := fmt.Sprintf("%s/queue/item/%s/api/json", baseURL, queueItemID)

	// Create request
	req, err := http.NewRequest(http.MethodGet, queueURL, nil)
	if err != nil {
		return nil, err
	}

	// Set Basic Auth header
	cred := base64.StdEncoding.EncodeToString([]byte(tokenName + ":" + token))
	req.Header.Set("Authorization", "Basic "+cred)

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Errorf("Jenkins queue item request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Infof("Jenkins queue item response: status=%d", resp.StatusCode)

	// Handle 404 - queue item not found
	if resp.StatusCode == http.StatusNotFound {
		log.Warnf("Queue item not found: %s", queueItemID)
		return nil, fmt.Errorf("%w: queue item %s", errors.ErrJenkinsQueueItemNotFound, queueItemID)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Jenkins queue item request failed: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("jenkins queue item request failed: status=%d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		return nil, err
	}

	var queueData struct {
		ID           int    `json:"id"`
		Why          string `json:"why"`
		InQueueSince int64  `json:"inQueueSince"`
		Cancelled    bool   `json:"cancelled"`
		Blocked      bool   `json:"blocked"`
		Buildable    bool   `json:"buildable"`
		Executable   *struct {
			Number int    `json:"number"`
			URL    string `json:"url"`
		} `json:"executable"`
	}

	if err := json.Unmarshal(body, &queueData); err != nil {
		log.Errorf("Failed to parse queue item response: %v", err)
		return nil, fmt.Errorf("failed to parse queue item response: %w", err)
	}

	// Calculate wait time
	currentTime := time.Now().Unix() * 1000                        // Convert to milliseconds
	waitTime := int((currentTime - queueData.InQueueSince) / 1000) // Convert to seconds

	// Determine status
	status := "queued"
	var buildNumber *int
	buildURL := ""

	if queueData.Cancelled {
		status = "cancelled"
	} else if queueData.Executable != nil {
		// Job has started
		status = "running"
		buildNumber = &queueData.Executable.Number
		buildURL = queueData.Executable.URL
		log.Infof("Queue item has started: build #%d", queueData.Executable.Number)
	}

	result := &JenkinsQueueStatusResult{
		Status:       status,
		BuildNumber:  buildNumber,
		BuildURL:     buildURL,
		QueuedReason: queueData.Why,
		WaitTime:     waitTime,
	}

	log.Info("Successfully retrieved queue item status")
	return result, nil
}

// GetBuildStatus retrieves the status of a Jenkins build
func (s *JenkinsService) GetBuildStatus(ctx context.Context, jaasName, jobName string, buildNumber int) (*JenkinsBuildStatusResult, error) {
	log := logger.WithContext(ctx).WithFields(map[string]interface{}{
		"jaasName":    jaasName,
		"jobName":     jobName,
		"buildNumber": buildNumber,
	})

	log.Info("Getting Jenkins build status")

	// Get credentials
	tokenName, token, err := s.getJenkinsCredentials(ctx, jaasName)
	if err != nil {
		return nil, err
	}

	// Build URL
	baseURL := strings.Replace(s.cfg.JenkinsBaseURL, "{jaasName}", jaasName, -1)
	buildURL := fmt.Sprintf("%s/job/%s/%d/api/json", baseURL, jobName, buildNumber)

	// Create request
	req, err := http.NewRequest(http.MethodGet, buildURL, nil)
	if err != nil {
		return nil, err
	}

	// Set Basic Auth header
	cred := base64.StdEncoding.EncodeToString([]byte(tokenName + ":" + token))
	req.Header.Set("Authorization", "Basic "+cred)

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Errorf("Jenkins build status request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Infof("Jenkins build status response: status=%d", resp.StatusCode)

	// Handle 404 - build not found
	if resp.StatusCode == http.StatusNotFound {
		log.Warnf("Build not found: %s #%d", jobName, buildNumber)
		return nil, fmt.Errorf("%w: job %s build #%d", errors.ErrJenkinsBuildNotFound, jobName, buildNumber)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Jenkins build status request failed: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("jenkins build status request failed: status=%d", resp.StatusCode)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read response body: %v", err)
		return nil, err
	}

	var buildData struct {
		Number    int    `json:"number"`
		Result    string `json:"result"` // SUCCESS, FAILURE, UNSTABLE, ABORTED, or null
		Building  bool   `json:"building"`
		Duration  int64  `json:"duration"` // milliseconds
		URL       string `json:"url"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(body, &buildData); err != nil {
		log.Errorf("Failed to parse build status response: %v", err)
		return nil, fmt.Errorf("failed to parse build status response: %w", err)
	}

	// Map Jenkins result to user-friendly status
	status := "running"
	if !buildData.Building {
		switch buildData.Result {
		case "SUCCESS":
			status = "success"
		case "FAILURE":
			status = "failure"
		case "ABORTED":
			status = "aborted"
		case "UNSTABLE":
			status = "unstable"
		default:
			status = "unknown"
		}
	}

	result := &JenkinsBuildStatusResult{
		Status:   status,
		Result:   buildData.Result,
		Building: buildData.Building,
		Duration: buildData.Duration,
		BuildURL: buildData.URL,
	}

	log.Infof("Build status: %s (result=%s, building=%v)", status, buildData.Result, buildData.Building)
	return result, nil
}
