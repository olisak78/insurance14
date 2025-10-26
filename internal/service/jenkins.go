package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"developer-portal-backend/internal/logger"
)

// JenkinsService provides methods to interact with Jenkins JAAS instances
type JenkinsService struct {
	httpClient *http.Client
}

// NewJenkinsService creates a new Jenkins service
func NewJenkinsService() *JenkinsService {
	return &JenkinsService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// getJenkinsCredentials retrieves Jenkins credentials from environment variables
// Expected env vars: JENKINS_{JAASNAME_UPPERCASE}_JAAS_TOKEN (instance-specific) and JENKINS_P_USER (global username)
func (s *JenkinsService) getJenkinsCredentials(ctx context.Context, jaasName string) (tokenName, token string, err error) {
	log := logger.WithContext(ctx)
	jaasNameUpper := strings.ToUpper(strings.ReplaceAll(jaasName, "-", "_"))

	tokenEnvVar := fmt.Sprintf("JENKINS_%s_JAAS_TOKEN", jaasNameUpper)
	userEnvVar := "JENKINS_P_USER"

	log.Debugf("Looking for Jenkins credentials: %s and %s", tokenEnvVar, userEnvVar)

	token = os.Getenv(tokenEnvVar)
	tokenName = os.Getenv(userEnvVar)

	if token == "" {
		log.Errorf("Jenkins token not found: missing %s environment variable", tokenEnvVar)
		return "", "", fmt.Errorf("jenkins token not found: missing %s environment variable", tokenEnvVar)
	}

	if tokenName == "" {
		log.Errorf("Jenkins username not found: missing %s environment variable", userEnvVar)
		return "", "", fmt.Errorf("jenkins username not found: missing %s environment variable", userEnvVar)
	}

	log.Debugf("Found Jenkins credentials for %s (username: %s)", jaasName, tokenName)
	return tokenName, token, nil
}

// buildJenkinsURL constructs the Jenkins URL for a given JAAS instance and job
func (s *JenkinsService) buildJenkinsURL(jaasName, jobName string) string {
	return fmt.Sprintf("https://%s.jaas-gcp.cloud.sap.corp/job/%s", jaasName, jobName)
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

	// Log auth info (without exposing full token)
	tokenPreview := "***"
	if len(token) >= 12 {
		// Only show preview for reasonably long tokens (12+ chars)
		// to avoid revealing too much of short tokens
		tokenPreview = token[:4] + "..." + token[len(token)-4:]
	}
	log.Debugf("Using Basic Auth: username=%s, token=%s", tokenName, tokenPreview)

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
	log := logger.WithContext(ctx)

	// Get the property array
	properties, ok := response["property"].([]interface{})
	if !ok {
		log.Debug("No property array found in Jenkins response")
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
			log.Debug("No parameterDefinitions found in ParametersDefinitionProperty")
			return map[string]interface{}{
				"parameterDefinitions": []interface{}{},
			}
		}

		log.Debug("Found ParametersDefinitionProperty with parameter definitions")
		return map[string]interface{}{
			"parameterDefinitions": paramDefs,
		}
	}

	// If no ParametersDefinitionProperty found, return empty parameterDefinitions
	log.Debug("No ParametersDefinitionProperty found in Jenkins response")
	return map[string]interface{}{
		"parameterDefinitions": []interface{}{},
	}
}

// TriggerJob triggers a Jenkins job with the provided parameters
func (s *JenkinsService) TriggerJob(ctx context.Context, jaasName, jobName string, parameters map[string]string) error {
	log := logger.WithContext(ctx).WithFields(map[string]interface{}{
		"jaasName":   jaasName,
		"jobName":    jobName,
		"paramCount": len(parameters),
	})

	log.Info("Triggering Jenkins job")

	tokenName, token, err := s.getJenkinsCredentials(ctx, jaasName)
	if err != nil {
		log.Errorf("Failed to get Jenkins credentials: %v", err)
		return err
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
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set Basic Auth header using token name as username and token as password
	cred := base64.StdEncoding.EncodeToString([]byte(tokenName + ":" + token))
	req.Header.Set("Authorization", "Basic "+cred)

	// Log auth info (without exposing full token)
	tokenPreview := "***"
	if len(token) >= 12 {
		// Only show preview for reasonably long tokens (12+ chars)
		// to avoid revealing too much of short tokens
		tokenPreview = token[:4] + "..." + token[len(token)-4:]
	}
	log.Debugf("Using Basic Auth: username=%s, token=%s", tokenName, tokenPreview)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Errorf("Jenkins trigger job request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	log.Infof("Jenkins trigger job response: status=%d", resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Jenkins trigger job failed: status=%d, body=%s", resp.StatusCode, string(body))
		return fmt.Errorf("jenkins trigger failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	log.Info("Successfully triggered Jenkins job")
	return nil
}
