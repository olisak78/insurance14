package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

// AICoreCredentials represents the credentials for a specific team
type AICoreCredentials struct {
	Team          string `json:"team"`
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	OAuthURL      string `json:"oauthUrl"`
	APIURL        string `json:"apiUrl"`
	ResourceGroup string `json:"resourceGroup"`
}

// AICoreTokenResponse represents the OAuth token response
type AICoreTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// tokenCache represents a cached access token with expiration
type tokenCache struct {
	token     string
	expiresAt time.Time
}

// AICoreDeployment represents a deployment from AI Core
type AICoreDeployment struct {
	ID              string                 `json:"id"`
	ConfigurationID string                 `json:"configurationId"`
	Status          string                 `json:"status"`
	StatusMessage   string                 `json:"statusMessage"`
	DeploymentURL   string                 `json:"deploymentUrl"`
	CreatedAt       string                 `json:"createdAt"`
	ModifiedAt      string                 `json:"modifiedAt"`
	Details         map[string]interface{} `json:"details,omitempty"`
}

// AICoreTeamDeployments represents deployments for a specific team
type AICoreTeamDeployments struct {
	Team        string             `json:"team"`
	Deployments []AICoreDeployment `json:"deployments"`
}

// AICoreDeploymentsResponse represents the response from AI Core deployments API
type AICoreDeploymentsResponse struct {
	Count       int                     `json:"count"`
	Deployments []AICoreTeamDeployments `json:"deployments"`
}

// AICoreModel represents a model from AI Core
type AICoreModel struct {
	Model        string                   `json:"model"`
	ExecutableID string                   `json:"executableId"`
	Description  string                   `json:"description"`
	DisplayName  string                   `json:"displayName,omitempty"`
	AccessType   string                   `json:"accessType,omitempty"`
	Provider     string                   `json:"provider,omitempty"`
	Versions     []AICoreModelVersion     `json:"versions"`
	Scenarios    []map[string]interface{} `json:"allowedScenarios,omitempty"`
}

// AICoreModelVersion represents a model version
type AICoreModelVersion struct {
	Name                      string                   `json:"name"`
	IsLatest                  bool                     `json:"isLatest"`
	Deprecated                bool                     `json:"deprecated"`
	RetirementDate            string                   `json:"retirementDate,omitempty"`
	ContextLength             int                      `json:"contextLength,omitempty"`
	InputTypes                []string                 `json:"inputTypes,omitempty"`
	Capabilities              []string                 `json:"capabilities,omitempty"`
	Metadata                  []map[string]interface{} `json:"metadata,omitempty"`
	Cost                      []map[string]interface{} `json:"cost,omitempty"`
	SuggestedReplacements     []string                 `json:"suggestedReplacements,omitempty"`
	StreamingSupported        bool                     `json:"streamingSupported,omitempty"`
	OrchestrationCapabilities []string                 `json:"orchestrationCapabilities,omitempty"`
}

// AICoreModelsResponse represents the response from AI Core models API
type AICoreModelsResponse struct {
	Count     int           `json:"count"`
	Resources []AICoreModel `json:"resources"`
}

// AICoreConfigurationsResponse represents the response from AI Core configurations API
type AICoreConfigurationsResponse struct {
	Count     int                   `json:"count"`
	Resources []AICoreConfiguration `json:"resources"`
}

// AICoreConfiguration represents a configuration from AI Core
type AICoreConfiguration struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	ExecutableID          string                 `json:"executableId"`
	ScenarioID            string                 `json:"scenarioId"`
	ParameterBindings     []map[string]string    `json:"parameterBindings,omitempty"`
	InputArtifactBindings []map[string]string    `json:"inputArtifactBindings,omitempty"`
	CreatedAt             string                 `json:"createdAt"`
	Scenario              map[string]interface{} `json:"scenario,omitempty"`
}

// AICoreConfigurationRequest represents a request to create a configuration
type AICoreConfigurationRequest struct {
	Name                  string              `json:"name" validate:"required"`
	ExecutableID          string              `json:"executableId" validate:"required"`
	ScenarioID            string              `json:"scenarioId" validate:"required"`
	ParameterBindings     []map[string]string `json:"parameterBindings,omitempty"`
	InputArtifactBindings []map[string]string `json:"inputArtifactBindings,omitempty"`
}

// AICoreConfigurationResponse represents the response from creating a configuration
type AICoreConfigurationResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// AICoreDeploymentRequest represents a request to create a deployment
type AICoreDeploymentRequest struct {
	ConfigurationID string `json:"configurationId" validate:"required"`
	TTL             string `json:"ttl,omitempty"`
}

// AICoreDeploymentResponse represents the response from creating a deployment
type AICoreDeploymentResponse struct {
	ID            string `json:"id"`
	Message       string `json:"message"`
	DeploymentURL string `json:"deploymentUrl,omitempty"`
	Status        string `json:"status,omitempty"`
	TTL           string `json:"ttl,omitempty"`
}

// AICoreDeploymentModificationRequest represents a request to modify a deployment
type AICoreDeploymentModificationRequest struct {
	TargetStatus    string `json:"targetStatus,omitempty"`
	ConfigurationID string `json:"configurationId,omitempty"`
}

// AICoreDeploymentModificationResponse represents the response from modifying a deployment
type AICoreDeploymentModificationResponse struct {
	ID            string `json:"id"`
	Message       string `json:"message"`
	DeploymentURL string `json:"deploymentUrl,omitempty"`
	Status        string `json:"status,omitempty"`
	TargetStatus  string `json:"targetStatus,omitempty"`
}

// AICoreDeploymentDeletionResponse represents the response from deleting a deployment
type AICoreDeploymentDeletionResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// AICoreDeploymentDetailsResponse represents the detailed response for a specific deployment
type AICoreDeploymentDetailsResponse struct {
	ID                           string                 `json:"id"`
	DeploymentURL                string                 `json:"deploymentUrl"`
	ConfigurationID              string                 `json:"configurationId"`
	ConfigurationName            string                 `json:"configurationName"`
	ExecutableID                 string                 `json:"executableId"`
	ScenarioID                   string                 `json:"scenarioId"`
	Status                       string                 `json:"status"`
	StatusMessage                string                 `json:"statusMessage"`
	TargetStatus                 string                 `json:"targetStatus"`
	LastOperation                string                 `json:"lastOperation"`
	LatestRunningConfigurationID string                 `json:"latestRunningConfigurationId"`
	TTL                          string                 `json:"ttl"`
	Details                      map[string]interface{} `json:"details"`
	CreatedAt                    string                 `json:"createdAt"`
	ModifiedAt                   string                 `json:"modifiedAt"`
	SubmissionTime               string                 `json:"submissionTime"`
	StartTime                    string                 `json:"startTime"`
	CompletionTime               string                 `json:"completionTime"`
	StatusDetails                map[string]interface{} `json:"statusDetails"`
}

// AICoreService handles AI Core operations
type AICoreService struct {
	memberRepo      repository.MemberRepositoryInterface
	teamRepo        repository.TeamRepositoryInterface
	groupRepo       repository.GroupRepositoryInterface
	httpClient      *http.Client
	credentials     map[string]*AICoreCredentials // Cached credentials by team name
	credentialsMux  sync.RWMutex                  // Protects credentials cache
	tokenCache      map[string]*tokenCache        // Cached tokens by team name
	tokenCacheMux   sync.RWMutex                  // Protects token cache
	credentialsOnce sync.Once                     // Ensures credentials are loaded only once
}

// NewAICoreService creates a new AI Core service
func NewAICoreService(memberRepo repository.MemberRepositoryInterface, teamRepo repository.TeamRepositoryInterface, groupRepo repository.GroupRepositoryInterface) AICoreServiceInterface {
	return &AICoreService{
		memberRepo:  memberRepo,
		teamRepo:    teamRepo,
		groupRepo:   groupRepo,
		credentials: make(map[string]*AICoreCredentials),
		tokenCache:  make(map[string]*tokenCache),
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Reduced from 30s for better UX
		},
	}
}

// getTeamLimit returns the configurable team limit from environment variable or default
func (s *AICoreService) getTeamLimit() int {
	limitStr := os.Getenv("AI_CORE_TEAM_LIMIT")
	if limitStr == "" {
		return 1000 // Default limit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 1000 // Default limit if invalid
	}

	return limit
}

// loadCredentials loads and caches all AI Core credentials from environment variable
func (s *AICoreService) loadCredentials() error {
	credentialsJSON := os.Getenv("AI_CORE_CREDENTIALS")
	if credentialsJSON == "" {
		return errors.ErrAICoreCredentialsNotSet
	}

	var credentialsList []AICoreCredentials
	if err := json.Unmarshal([]byte(credentialsJSON), &credentialsList); err != nil {
		return fmt.Errorf("%w: %v", errors.ErrAICoreCredentialsInvalid, err)
	}

	s.credentialsMux.Lock()
	defer s.credentialsMux.Unlock()

	// Clear existing credentials and rebuild cache
	s.credentials = make(map[string]*AICoreCredentials)
	for i := range credentialsList {
		cred := &credentialsList[i]
		s.credentials[cred.Team] = cred
	}

	return nil
}

// getCredentialsForTeam retrieves AI Core credentials for a specific team from cache
func (s *AICoreService) getCredentialsForTeam(teamName string) (*AICoreCredentials, error) {
	// Load credentials once
	var loadErr error
	s.credentialsOnce.Do(func() {
		loadErr = s.loadCredentials()
	})
	if loadErr != nil {
		return nil, loadErr
	}

	s.credentialsMux.RLock()
	defer s.credentialsMux.RUnlock()

	cred, exists := s.credentials[teamName]
	if !exists {
		return nil, errors.NewAICoreCredentialsNotFoundError(teamName)
	}

	return cred, nil
}

// getUserTeam retrieves the team name for the authenticated user
func (s *AICoreService) getUserTeam(c *gin.Context) (string, error) {
	// Get user email from auth context
	email, exists := auth.GetUserEmail(c)
	if !exists {
		return "", errors.ErrUserEmailNotFound
	}

	// Get user from database
	member, err := s.memberRepo.GetByEmail(email)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", errors.ErrUserNotFoundInDB
		}
		return "", fmt.Errorf("failed to get user from database: %w", err)
	}

	// Get user's team
	if member.TeamID == nil {
		return "", errors.ErrUserNotAssignedToTeam
	}

	team, err := s.teamRepo.GetByID(*member.TeamID)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", errors.ErrTeamNotFoundInDB
		}
		return "", fmt.Errorf("failed to get team from database: %w", err)
	}

	return team.Name, nil
}

// getAccessToken retrieves an access token for AI Core API with caching
func (s *AICoreService) getAccessToken(credentials *AICoreCredentials) (string, error) {
	teamName := credentials.Team

	// Check cache first
	s.tokenCacheMux.RLock()
	if cached, exists := s.tokenCache[teamName]; exists {
		if time.Now().Before(cached.expiresAt) {
			s.tokenCacheMux.RUnlock()
			return cached.token, nil
		}
	}
	s.tokenCacheMux.RUnlock()

	// Token not cached or expired, get new token
	token, expiresIn, err := s.requestNewToken(credentials)
	if err != nil {
		return "", err
	}

	// Cache the token with a buffer (expire 5 minutes early to be safe)
	expiresAt := time.Now().Add(time.Duration(expiresIn-300) * time.Second)

	s.tokenCacheMux.Lock()
	s.tokenCache[teamName] = &tokenCache{
		token:     token,
		expiresAt: expiresAt,
	}
	s.tokenCacheMux.Unlock()

	return token, nil
}

// requestNewToken requests a new access token from the OAuth endpoint
func (s *AICoreService) requestNewToken(credentials *AICoreCredentials) (string, int, error) {
	// Use proper form encoding instead of string concatenation for security
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", credentials.ClientID)
	data.Set("client_secret", credentials.ClientSecret)

	req, err := http.NewRequest("POST", credentials.OAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp AICoreTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", 0, fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

// makeAICoreRequest makes an authenticated request to AI Core API
func (s *AICoreService) makeAICoreRequest(method, url, accessToken, resourceGroup string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AI-Resource-Group", resourceGroup)

	return s.httpClient.Do(req)
}

// GetDeployments retrieves deployments from AI Core based on user's role
func (s *AICoreService) GetDeployments(c *gin.Context) (*AICoreDeploymentsResponse, error) {
	// Get user email from auth context
	email, exists := auth.GetUserEmail(c)
	if !exists {
		return nil, errors.ErrUserEmailNotFound
	}

	// Get user from database
	member, err := s.memberRepo.GetByEmail(email)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.ErrUserNotFoundInDB
		}
		return nil, fmt.Errorf("failed to get user from database: %w", err)
	}

	// Determine user role and get appropriate teams (including metadata-based teams)
	teamNames, err := s.getAllTeamsForUser(member)
	if err != nil {
		return nil, err
	}

	// Aggregate deployments from all teams, grouped by team
	teamDeployments := make([]AICoreTeamDeployments, 0)
	totalCount := 0

	for _, teamName := range teamNames {
		// Get credentials for the team
		credentials, err := s.getCredentialsForTeam(teamName)
		if err != nil {
			// Skip teams without credentials instead of failing
			continue
		}

		// Get access token
		accessToken, err := s.getAccessToken(credentials)
		if err != nil {
			// Skip teams with token issues instead of failing
			continue
		}

		// Make request to AI Core
		url := fmt.Sprintf("%s/v2/lm/deployments", credentials.APIURL)
		resp, err := s.makeAICoreRequest("GET", url, accessToken, credentials.ResourceGroup, nil)
		if err != nil {
			// Skip teams with API issues instead of failing
			continue
		}

		// Ensure response body is always closed
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// We need to decode into a temporary structure that matches AI Core's actual response
			var tempResp struct {
				Count     int                `json:"count"`
				Resources []AICoreDeployment `json:"resources"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&tempResp); err == nil {
				// Create team deployment entry
				teamDeployment := AICoreTeamDeployments{
					Team:        teamName,
					Deployments: tempResp.Resources,
				}
				teamDeployments = append(teamDeployments, teamDeployment)
				totalCount += tempResp.Count
			}
		}
	}

	// Return aggregated response with new structure
	return &AICoreDeploymentsResponse{
		Count:       totalCount,
		Deployments: teamDeployments,
	}, nil
}

// getTeamsForUser determines which teams a user should see deployments for based on their role
func (s *AICoreService) getTeamsForUser(member *models.Member) ([]string, error) {
	var teamNames []string
	teamLimit := s.getTeamLimit()

	// Check if user is an organization manager (has role manager and no team/group assignment)
	if member.Role == models.MemberRoleManager && member.TeamID == nil && member.GroupID == nil {
		// Organization manager - get all teams in the organization
		teams, _, err := s.teamRepo.GetByOrganizationID(member.OrganizationID, teamLimit, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get teams for organization: %w", err)
		}
		for _, team := range teams {
			teamNames = append(teamNames, team.Name)
		}
		return teamNames, nil
	}

	// Check if user is a group manager (has role manager, has groupID but no teamID)
	if member.Role == models.MemberRoleManager && member.TeamID == nil && member.GroupID != nil {
		// Group manager - get all teams in the group
		teams, _, err := s.teamRepo.GetByGroupID(*member.GroupID, teamLimit, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get teams for group: %w", err)
		}
		for _, team := range teams {
			teamNames = append(teamNames, team.Name)
		}
		return teamNames, nil
	}

	// Team member or team manager - get only their team
	if member.TeamID != nil {
		team, err := s.teamRepo.GetByID(*member.TeamID)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, errors.ErrTeamNotFoundInDB
			}
			return nil, fmt.Errorf("failed to get team from database: %w", err)
		}
		teamNames = append(teamNames, team.Name)
		return teamNames, nil
	}

	// User not assigned to any team and not a manager
	return nil, errors.ErrUserNotAssignedToTeam
}

// getAICoreTeamsFromMetadata extracts team names from the member's metadata.ai_core_member_of field
func (s *AICoreService) getAICoreTeamsFromMetadata(member *models.Member) []string {
	var teamNames []string

	if member.Metadata == nil {
		return teamNames
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(member.Metadata, &metadata); err != nil {
		return teamNames
	}

	aiCoreMemberOf, exists := metadata["ai_core_member_of"]
	if !exists {
		return teamNames
	}

	// Handle different possible types for ai_core_member_of
	switch v := aiCoreMemberOf.(type) {
	case []interface{}:
		// Array of team names
		for _, teamNameInterface := range v {
			if teamName, ok := teamNameInterface.(string); ok && teamName != "" {
				teamNames = append(teamNames, teamName)
			}
		}
	case []string:
		// Array of strings (direct type)
		for _, teamName := range v {
			if teamName != "" {
				teamNames = append(teamNames, teamName)
			}
		}
	case string:
		// Single team name as string
		if v != "" {
			teamNames = append(teamNames, v)
		}
	}

	return teamNames
}

// getAllTeamsForUser combines teams from role-based access and metadata-based access
func (s *AICoreService) getAllTeamsForUser(member *models.Member) ([]string, error) {
	teamNamesSet := make(map[string]bool) // Use a set to avoid duplicates
	var allTeamNames []string

	// Get teams based on role and team assignment
	roleBasedTeams, err := s.getTeamsForUser(member)
	if err != nil && err != errors.ErrUserNotAssignedToTeam {
		return nil, err
	}

	// Add role-based teams to the set
	for _, teamName := range roleBasedTeams {
		if !teamNamesSet[teamName] {
			allTeamNames = append(allTeamNames, teamName)
			teamNamesSet[teamName] = true
		}
	}

	// Get additional teams from metadata
	metadataTeams := s.getAICoreTeamsFromMetadata(member)
	for _, teamName := range metadataTeams {
		if !teamNamesSet[teamName] {
			allTeamNames = append(allTeamNames, teamName)
			teamNamesSet[teamName] = true
		}
	}

	// If no teams found at all, return error
	if len(allTeamNames) == 0 {
		return nil, errors.ErrUserNotAssignedToTeam
	}

	return allTeamNames, nil
}

// GetModels retrieves models from AI Core for the user's team
func (s *AICoreService) GetModels(c *gin.Context, scenarioID string) (*AICoreModelsResponse, error) {
	// Get user's team
	teamName, err := s.getUserTeam(c)
	if err != nil {
		return nil, err
	}

	// Get credentials for the team
	credentials, err := s.getCredentialsForTeam(teamName)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return nil, err
	}

	// Make request to AI Core
	url := fmt.Sprintf("%s/v2/lm/scenarios/%s/models", credentials.APIURL, scenarioID)
	resp, err := s.makeAICoreRequest("GET", url, accessToken, credentials.ResourceGroup, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w with status %d: %s", errors.ErrAICoreAPIRequestFailed, resp.StatusCode, string(body))
	}

	var modelsResp AICoreModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	return &modelsResp, nil
}

// GetConfigurations retrieves configurations from AI Core for the user's team
func (s *AICoreService) GetConfigurations(c *gin.Context) (*AICoreConfigurationsResponse, error) {
	// Get user's team
	teamName, err := s.getUserTeam(c)
	if err != nil {
		return nil, err
	}

	// Get credentials for the team
	credentials, err := s.getCredentialsForTeam(teamName)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return nil, err
	}

	// Make request to AI Core
	url := fmt.Sprintf("%s/v2/lm/configurations", credentials.APIURL)
	resp, err := s.makeAICoreRequest("GET", url, accessToken, credentials.ResourceGroup, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w with status %d: %s", errors.ErrAICoreAPIRequestFailed, resp.StatusCode, string(body))
	}

	var configurationsResp AICoreConfigurationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&configurationsResp); err != nil {
		return nil, fmt.Errorf("failed to decode configurations response: %w", err)
	}

	return &configurationsResp, nil
}

// CreateConfiguration creates a new configuration in AI Core
func (s *AICoreService) CreateConfiguration(c *gin.Context, req *AICoreConfigurationRequest) (*AICoreConfigurationResponse, error) {
	// Get user's team
	teamName, err := s.getUserTeam(c)
	if err != nil {
		return nil, err
	}

	// Get credentials for the team
	credentials, err := s.getCredentialsForTeam(teamName)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return nil, err
	}

	// Make request to AI Core
	url := fmt.Sprintf("%s/v2/lm/configurations", credentials.APIURL)
	resp, err := s.makeAICoreRequest("POST", url, accessToken, credentials.ResourceGroup, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w with status %d: %s", errors.ErrAICoreAPIRequestFailed, resp.StatusCode, string(body))
	}

	var configResp AICoreConfigurationResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return nil, fmt.Errorf("failed to decode configuration response: %w", err)
	}

	return &configResp, nil
}

// CreateDeployment creates a new deployment in AI Core
func (s *AICoreService) CreateDeployment(c *gin.Context, req *AICoreDeploymentRequest) (*AICoreDeploymentResponse, error) {
	// Get user's team
	teamName, err := s.getUserTeam(c)
	if err != nil {
		return nil, err
	}

	// Get credentials for the team
	credentials, err := s.getCredentialsForTeam(teamName)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return nil, err
	}

	// Make request to AI Core
	url := fmt.Sprintf("%s/v2/lm/deployments", credentials.APIURL)
	resp, err := s.makeAICoreRequest("POST", url, accessToken, credentials.ResourceGroup, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w with status %d: %s", errors.ErrAICoreAPIRequestFailed, resp.StatusCode, string(body))
	}

	var deploymentResp AICoreDeploymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&deploymentResp); err != nil {
		return nil, fmt.Errorf("failed to decode deployment response: %w", err)
	}

	return &deploymentResp, nil
}

// UpdateDeployment updates a deployment in AI Core
func (s *AICoreService) UpdateDeployment(c *gin.Context, deploymentID string, req *AICoreDeploymentModificationRequest) (*AICoreDeploymentModificationResponse, error) {
	// Get user's team
	teamName, err := s.getUserTeam(c)
	if err != nil {
		return nil, err
	}

	// Get credentials for the team
	credentials, err := s.getCredentialsForTeam(teamName)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return nil, err
	}

	// Make request to AI Core
	url := fmt.Sprintf("%s/v2/lm/deployments/%s", credentials.APIURL, deploymentID)
	resp, err := s.makeAICoreRequest("PATCH", url, accessToken, credentials.ResourceGroup, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.ErrAICoreDeploymentNotFound
	}

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w with status %d: %s", errors.ErrAICoreAPIRequestFailed, resp.StatusCode, string(body))
	}

	var modificationResp AICoreDeploymentModificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&modificationResp); err != nil {
		return nil, fmt.Errorf("failed to decode deployment modification response: %w", err)
	}

	return &modificationResp, nil
}

// DeleteDeployment deletes a deployment in AI Core
func (s *AICoreService) DeleteDeployment(c *gin.Context, deploymentID string) (*AICoreDeploymentDeletionResponse, error) {
	// Get user's team
	teamName, err := s.getUserTeam(c)
	if err != nil {
		return nil, err
	}

	// Get credentials for the team
	credentials, err := s.getCredentialsForTeam(teamName)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return nil, err
	}

	// Make request to AI Core
	url := fmt.Sprintf("%s/v2/lm/deployments/%s", credentials.APIURL, deploymentID)
	resp, err := s.makeAICoreRequest("DELETE", url, accessToken, credentials.ResourceGroup, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.ErrAICoreDeploymentNotFound
	}

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w with status %d: %s", errors.ErrAICoreAPIRequestFailed, resp.StatusCode, string(body))
	}

	var deletionResp AICoreDeploymentDeletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&deletionResp); err != nil {
		return nil, fmt.Errorf("failed to decode deployment deletion response: %w", err)
	}

	return &deletionResp, nil
}

// GetDeploymentDetails retrieves detailed information about a specific deployment from AI Core
func (s *AICoreService) GetDeploymentDetails(c *gin.Context, deploymentID string) (*AICoreDeploymentDetailsResponse, error) {
	// Get user's team
	teamName, err := s.getUserTeam(c)
	if err != nil {
		return nil, err
	}

	// Get credentials for the team
	credentials, err := s.getCredentialsForTeam(teamName)
	if err != nil {
		return nil, err
	}

	// Get access token
	accessToken, err := s.getAccessToken(credentials)
	if err != nil {
		return nil, err
	}

	// Make request to AI Core
	url := fmt.Sprintf("%s/v2/lm/deployments/%s", credentials.APIURL, deploymentID)
	resp, err := s.makeAICoreRequest("GET", url, accessToken, credentials.ResourceGroup, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.ErrAICoreDeploymentNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w with status %d: %s", errors.ErrAICoreAPIRequestFailed, resp.StatusCode, string(body))
	}

	var deploymentDetails AICoreDeploymentDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&deploymentDetails); err != nil {
		return nil, fmt.Errorf("failed to decode deployment details response: %w", err)
	}

	return &deploymentDetails, nil
}
