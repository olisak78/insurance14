package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock implementations for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

// Additional methods to satisfy repository.UserRepositoryInterface
func (m *MockUserRepository) GetByName(name string) (*models.User, error) {
	args := m.Called(name)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) GetByUserID(userID string) (*models.User, error) {
	args := m.Called(userID)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) GetAll(limit, offset int) ([]models.User, int64, error) {
	args := m.Called(limit, offset)
	users, _ := args.Get(0).([]models.User)
	total, _ := args.Get(1).(int64)
	return users, total, args.Error(2)
}

func (m *MockUserRepository) GetByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(organizationID, limit, offset)
	users, _ := args.Get(0).([]models.User)
	total, _ := args.Get(1).(int64)
	return users, total, args.Error(2)
}

func (m *MockUserRepository) GetWithOrganization(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) SearchByOrganization(orgID uuid.UUID, query string, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(orgID, query, limit, offset)
	users, _ := args.Get(0).([]models.User)
	total, _ := args.Get(1).(int64)
	return users, total, args.Error(2)
}

func (m *MockUserRepository) SearchByNameOrTitleGlobal(query string, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(query, limit, offset)
	users, _ := args.Get(0).([]models.User)
	total, _ := args.Get(1).(int64)
	return users, total, args.Error(2)
}

func (m *MockUserRepository) GetActiveByOrganization(orgID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(orgID, limit, offset)
	users, _ := args.Get(0).([]models.User)
	total, _ := args.Get(1).(int64)
	return users, total, args.Error(2)
}

func (m *MockUserRepository) GetUserIDsByPrefix(prefix string) ([]string, error) {
	args := m.Called(prefix)
	ids, _ := args.Get(0).([]string)
	return ids, args.Error(1)
}

func (m *MockUserRepository) GetExistingUserIDs(ids []string) ([]string, error) {
	args := m.Called(ids)
	existing, _ := args.Get(0).([]string)
	return existing, args.Error(1)
}

func (m *MockUserRepository) Update(member *models.User) error {
	args := m.Called(member)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(team *models.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByName(groupID uuid.UUID, name string) (*models.Team, error) {
	args := m.Called(groupID, name)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]models.Team, int64, error) {
	args := m.Called(organizationID, limit, offset)
	return args.Get(0).([]models.Team), args.Get(1).(int64), args.Error(2)
}

func (m *MockTeamRepository) GetByGroupID(groupID uuid.UUID, limit, offset int) ([]models.Team, int64, error) {
	args := m.Called(groupID, limit, offset)
	return args.Get(0).([]models.Team), args.Get(1).(int64), args.Error(2)
}

func (m *MockTeamRepository) GetAll() ([]models.Team, error) {
	args := m.Called()
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByNameGlobal(name string) (*models.Team, error) {
	args := m.Called(name)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetWithMembers(id uuid.UUID) (*models.Team, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetWithComponentOwnerships(id uuid.UUID) (*models.Team, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetWithProjects(id uuid.UUID) (*models.Team, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetWithDutySchedules(id uuid.UUID) (*models.Team, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) Update(team *models.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamMembersByName(organizationID uuid.UUID, teamName string, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(organizationID, teamName, limit, offset)
	users, _ := args.Get(0).([]models.User)
	total, _ := args.Get(1).(int64)
	return users, total, args.Error(2)
}

func (m *MockTeamRepository) GetTeamComponentsByID(id uuid.UUID, limit, offset int) ([]models.Component, int64, error) {
	args := m.Called(id, limit, offset)
	return args.Get(0).([]models.Component), args.Get(1).(int64), args.Error(2)
}

func (m *MockTeamRepository) GetTeamComponentsByName(organizationID uuid.UUID, teamName string, limit, offset int) ([]models.Component, int64, error) {
	args := m.Called(organizationID, teamName, limit, offset)
	return args.Get(0).([]models.Component), args.Get(1).(int64), args.Error(2)
}

type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Create(group *models.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByID(id uuid.UUID) (*models.Group, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByName(organizationID uuid.UUID, name string) (*models.Group, error) {
	args := m.Called(organizationID, name)
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]models.Group, int64, error) {
	args := m.Called(organizationID, limit, offset)
	return args.Get(0).([]models.Group), args.Get(1).(int64), args.Error(2)
}

func (m *MockGroupRepository) Search(organizationID uuid.UUID, query string, limit, offset int) ([]models.Group, int64, error) {
	args := m.Called(organizationID, query, limit, offset)
	return args.Get(0).([]models.Group), args.Get(1).(int64), args.Error(2)
}

func (m *MockGroupRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGroupRepository) GetWithTeams(id uuid.UUID) (*models.Group, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetWithOrganization(id uuid.UUID) (*models.Group, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Group), args.Error(1)
}

type AICoreServiceTestSuite struct {
	suite.Suite
	service   *AICoreService
	userRepo  *MockUserRepository
	teamRepo  *MockTeamRepository
	groupRepo *MockGroupRepository
	server    *httptest.Server
}

func (suite *AICoreServiceTestSuite) SetupTest() {
	suite.userRepo = new(MockUserRepository)
	suite.teamRepo = new(MockTeamRepository)
	suite.groupRepo = new(MockGroupRepository)

	suite.service = &AICoreService{
		userRepo:    suite.userRepo,
		teamRepo:    suite.teamRepo,
		credentials: make(map[string]*AICoreCredentials),
		tokenCache:  make(map[string]*tokenCache),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (suite *AICoreServiceTestSuite) TearDownTest() {
	if suite.server != nil {
		suite.server.Close()
	}
	// Clear environment variables
	os.Unsetenv("AI_CORE_CREDENTIALS")
}

func (suite *AICoreServiceTestSuite) setupMockServer(responses map[string]mockResponse) {
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
		if response, exists := responses[key]; exists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(response.StatusCode)
			_, _ = w.Write([]byte(response.Body))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

type mockResponse struct {
	StatusCode int
	Body       string
}

func (suite *AICoreServiceTestSuite) setupCredentials(teams []string) {
	credentials := make([]AICoreCredentials, 0)
	serverURL := "http://localhost:8080" // Default URL
	if suite.server != nil {
		serverURL = suite.server.URL
	}

	for _, team := range teams {
		credentials = append(credentials, AICoreCredentials{
			Team:          team,
			ClientID:      fmt.Sprintf("client-%s", team),
			ClientSecret:  fmt.Sprintf("secret-%s", team),
			OAuthURL:      fmt.Sprintf("%s/oauth/token", serverURL),
			APIURL:        serverURL,
			ResourceGroup: "default",
		})
	}

	credentialsJSON, _ := json.Marshal(credentials)
	_ = os.Setenv("AI_CORE_CREDENTIALS", string(credentialsJSON))

	// Reset the service's credentials cache and once flag
	suite.service.credentials = make(map[string]*AICoreCredentials)
	suite.service.credentialsOnce = sync.Once{}
}

func (suite *AICoreServiceTestSuite) createGinContext(email string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Mock the auth context
	c.Set("email", email)

	return c
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_TeamMember_Success() {
	// Setup
	email := "team.member@example.com"
	teamID := uuid.New()

	member := &models.User{
		TeamID:   &teamID,
		TeamRole: models.TeamRoleMember,
	}

	team := &models.Team{
		BaseModel: models.BaseModel{ID: teamID, Name: "team-alpha"},
		Owner:     "team-alpha",
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"GET:/v2/lm/deployments": {
			StatusCode: 200,
			Body: `{
				"count": 2,
				"resources": [
					{
						"id": "deployment-1",
						"configurationId": "config-1",
						"status": "RUNNING",
						"statusMessage": "Deployment is running",
						"deploymentUrl": "https://api.example.com/v1/deployments/deployment-1",
						"createdAt": "2023-01-01T00:00:00Z",
						"modifiedAt": "2023-01-01T01:00:00Z"
					},
					{
						"id": "deployment-2",
						"configurationId": "config-2",
						"status": "STOPPED",
						"statusMessage": "Deployment is stopped",
						"deploymentUrl": "https://api.example.com/v1/deployments/deployment-2",
						"createdAt": "2023-01-01T00:00:00Z",
						"modifiedAt": "2023-01-01T02:00:00Z"
					}
				]
			}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-alpha"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByID", teamID).Return(team, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(2, result.Count)
	suite.Len(result.Deployments, 1)
	suite.Equal("team-alpha", result.Deployments[0].Team)
	suite.Len(result.Deployments[0].Deployments, 2)
	suite.Equal("deployment-1", result.Deployments[0].Deployments[0].ID)
	suite.Equal("RUNNING", result.Deployments[0].Deployments[0].Status)
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_GroupManager_Success() {
	// Setup
	email := "group.manager@example.com"
	groupID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()

	metadata := map[string]interface{}{
		"ai_core_member_of": []string{"team-alpha", "team-beta"},
	}
	metadataJSON, _ := json.Marshal(metadata)

	member := &models.User{
		TeamID:   nil,
		TeamRole: models.TeamRoleManager,
		Metadata: metadataJSON,
	}

	teams := []models.Team{
		{BaseModel: models.BaseModel{ID: team1ID, Name: "team-alpha"}, Owner: "team-alpha"},
		{BaseModel: models.BaseModel{ID: team2ID, Name: "team-beta"}, Owner: "team-beta"},
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"GET:/v2/lm/deployments": {
			StatusCode: 200,
			Body: `{
				"count": 1,
				"resources": [
					{
						"id": "deployment-1",
						"configurationId": "config-1",
						"status": "RUNNING",
						"statusMessage": "Deployment is running",
						"deploymentUrl": "https://api.example.com/v1/deployments/deployment-1",
						"createdAt": "2023-01-01T00:00:00Z",
						"modifiedAt": "2023-01-01T01:00:00Z"
					}
				]
			}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-alpha", "team-beta"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByGroupID", groupID, 1000, 0).Return(teams, int64(2), nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(2, result.Count) // 1 deployment from each team
	suite.Len(result.Deployments, 2)
	suite.Equal("team-alpha", result.Deployments[0].Team)
	suite.Equal("team-beta", result.Deployments[1].Team)
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_OrganizationManager_Success() {
	// Setup
	email := "org.manager@example.com"
	orgID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()

	metadata := map[string]interface{}{
		"ai_core_member_of": []string{"team-alpha", "team-beta"},
	}
	metadataJSON, _ := json.Marshal(metadata)

	member := &models.User{
		TeamID:   nil,
		TeamRole: models.TeamRoleManager,
		Metadata: metadataJSON,
	}

	teams := []models.Team{
		{BaseModel: models.BaseModel{ID: team1ID, Name: "team-alpha"}, Owner: "team-alpha"},
		{BaseModel: models.BaseModel{ID: team2ID, Name: "team-beta"}, Owner: "team-beta"},
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"GET:/v2/lm/deployments": {
			StatusCode: 200,
			Body: `{
				"count": 1,
				"resources": [
					{
						"id": "deployment-1",
						"configurationId": "config-1",
						"status": "RUNNING",
						"statusMessage": "Deployment is running",
						"deploymentUrl": "https://api.example.com/v1/deployments/deployment-1",
						"createdAt": "2023-01-01T00:00:00Z",
						"modifiedAt": "2023-01-01T01:00:00Z"
					}
				]
			}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-alpha", "team-beta"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByOrganizationID", orgID, 1000, 0).Return(teams, int64(2), nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(2, result.Count)
	suite.Len(result.Deployments, 2)
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_PartialCredentials_Success() {
	// Setup - Group manager with 3 teams, but only 2 have AI Core credentials
	email := "group.manager@example.com"
	groupID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()
	team3ID := uuid.New()

	metadata := map[string]interface{}{
		"ai_core_member_of": []string{"team-alpha", "team-beta"},
	}
	metadataJSON, _ := json.Marshal(metadata)

	member := &models.User{
		TeamID:   nil,
		TeamRole: models.TeamRoleManager,
		Metadata: metadataJSON,
	}

	teams := []models.Team{
		{BaseModel: models.BaseModel{ID: team1ID, Name: "team-alpha"}, Owner: "team-alpha"}, // Has credentials
		{BaseModel: models.BaseModel{ID: team2ID, Name: "team-beta"}, Owner: "team-beta"},   // Has credentials
		{BaseModel: models.BaseModel{ID: team3ID, Name: "team-gamma"}, Owner: "team-gamma"}, // No credentials
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"GET:/v2/lm/deployments": {
			StatusCode: 200,
			Body: `{
				"count": 1,
				"resources": [
					{
						"id": "deployment-1",
						"configurationId": "config-1",
						"status": "RUNNING",
						"statusMessage": "Deployment is running",
						"deploymentUrl": "https://api.example.com/v1/deployments/deployment-1",
						"createdAt": "2023-01-01T00:00:00Z",
						"modifiedAt": "2023-01-01T01:00:00Z"
					}
				]
			}`,
		},
	}
	suite.setupMockServer(responses)
	// Only setup credentials for team-alpha and team-beta, not team-gamma
	suite.setupCredentials([]string{"team-alpha", "team-beta"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByGroupID", groupID, 1000, 0).Return(teams, int64(3), nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(2, result.Count)     // Only 2 deployments from teams with credentials
	suite.Len(result.Deployments, 2) // Only 2 teams returned (team-gamma skipped)

	teamNames := make([]string, len(result.Deployments))
	for i, td := range result.Deployments {
		teamNames[i] = td.Team
	}
	suite.Contains(teamNames, "team-alpha")
	suite.Contains(teamNames, "team-beta")
	suite.NotContains(teamNames, "team-gamma") // Should be skipped due to missing credentials
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_NoCredentials_Error() {
	// Setup
	email := "team.member@example.com"
	teamID := uuid.New()

	member := &models.User{
		TeamID:   &teamID,
		TeamRole: models.TeamRoleMember,
	}

	team := &models.Team{
		BaseModel: models.BaseModel{ID: teamID, Name: "team-alpha"},
		Owner:     "team-alpha",
	}

	// Don't setup any credentials
	os.Unsetenv("AI_CORE_CREDENTIALS")

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByID", teamID).Return(team, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err) // Should not error, just return empty result
	suite.NotNil(result)
	suite.Equal(0, result.Count)
	suite.Len(result.Deployments, 0)
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_UserNotFound_Error() {
	// Setup
	email := "nonexistent@example.com"

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return((*models.User)(nil), errors.ErrUserNotFound)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Equal(errors.ErrUserNotFoundInDB, err)
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_UserNotAssignedToTeam_Error() {
	// Setup
	email := "unassigned@example.com"

	member := &models.User{
		TeamID:   nil,
		TeamRole: models.TeamRoleMember, // Not a manager
	}

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Equal(errors.ErrUserNotAssignedToTeam, err)
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_APIError_SkipsTeam() {
	// Setup
	email := "group.manager@example.com"
	groupID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()

	metadata := map[string]interface{}{
		"ai_core_member_of": []string{"team-alpha", "team-beta"},
	}
	metadataJSON, _ := json.Marshal(metadata)

	member := &models.User{
		TeamID:   nil,
		TeamRole: models.TeamRoleManager,
		Metadata: metadataJSON,
	}

	teams := []models.Team{
		{BaseModel: models.BaseModel{ID: team1ID, Name: "team-alpha"}, Owner: "team-alpha"},
		{BaseModel: models.BaseModel{ID: team2ID, Name: "team-beta"}, Owner: "team-beta"},
	}

	// Setup mock server responses - team-alpha returns error, team-beta succeeds
	callCount := 0
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`))
			return
		}

		if r.URL.Path == "/v2/lm/deployments" {
			callCount++
			if callCount == 1 {
				// First call (team-alpha) returns error
				w.WriteHeader(500)
				_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
			} else {
				// Second call (team-beta) succeeds
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
				_, _ = w.Write([]byte(`{
					"count": 1,
					"resources": [
						{
							"id": "deployment-1",
							"configurationId": "config-1",
							"status": "RUNNING",
							"statusMessage": "Deployment is running",
							"deploymentUrl": "https://api.example.com/v1/deployments/deployment-1",
							"createdAt": "2023-01-01T00:00:00Z",
							"modifiedAt": "2023-01-01T01:00:00Z"
						}
					]
				}`))
			}
		}
	}))

	suite.setupCredentials([]string{"team-alpha", "team-beta"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByGroupID", groupID, 1000, 0).Return(teams, int64(2), nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err) // Should not error, just skip the failing team
	suite.NotNil(result)
	suite.Equal(1, result.Count)     // Only 1 deployment from team-beta
	suite.Len(result.Deployments, 1) // Only team-beta returned
	suite.Equal("team-beta", result.Deployments[0].Team)
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_EmptyResponse_Success() {
	// Setup
	email := "team.member@example.com"
	teamID := uuid.New()

	member := &models.User{
		TeamID:   &teamID,
		TeamRole: models.TeamRoleMember,
	}

	team := &models.Team{
		BaseModel: models.BaseModel{ID: teamID, Name: "team-alpha"},
		Owner:     "team-alpha",
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"GET:/v2/lm/deployments": {
			StatusCode: 200,
			Body: `{
				"count": 0,
				"resources": []
			}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-alpha"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByID", teamID).Return(team, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(0, result.Count)
	suite.Len(result.Deployments, 1)
	suite.Equal("team-alpha", result.Deployments[0].Team)
	suite.Len(result.Deployments[0].Deployments, 0)
}
func (suite *AICoreServiceTestSuite) TestLoadCredentials_InvalidJSON_Error() {
	// Setup invalid JSON
	_ = os.Setenv("AI_CORE_CREDENTIALS", `{"invalid": json}`)

	// Execute
	err := suite.service.loadCredentials()

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "invalid character")
}

func (suite *AICoreServiceTestSuite) TestLoadCredentials_MissingEnvVar_Error() {
	// Setup
	_ = os.Unsetenv("AI_CORE_CREDENTIALS")

	// Execute
	err := suite.service.loadCredentials()

	// Assert
	suite.Error(err)
	suite.Equal(errors.ErrAICoreCredentialsNotSet, err)
}

func (suite *AICoreServiceTestSuite) TestGetCredentialsForTeam_TeamNotFound_Error() {
	// Setup
	suite.setupCredentials([]string{"team-alpha"})

	// Execute
	_, err := suite.service.getCredentialsForTeam("team-nonexistent")

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "team-nonexistent")
}

func (suite *AICoreServiceTestSuite) TestTokenCaching() {
	// Setup mock server first
	tokenCallCount := 0
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			tokenCallCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`))
		}
	}))

	// Now setup credentials with the server URL
	suite.setupCredentials([]string{"team-alpha"})

	credentials := &AICoreCredentials{
		Team:          "team-alpha",
		ClientID:      "client-alpha",
		ClientSecret:  "secret-alpha",
		OAuthURL:      fmt.Sprintf("%s/oauth/token", suite.server.URL),
		APIURL:        suite.server.URL,
		ResourceGroup: "default",
	}

	// Execute - call getAccessToken twice
	token1, err1 := suite.service.getAccessToken(credentials)
	token2, err2 := suite.service.getAccessToken(credentials)

	// Assert
	suite.NoError(err1)
	suite.NoError(err2)
	suite.Equal("test-token", token1)
	suite.Equal("test-token", token2)
	suite.Equal(1, tokenCallCount) // Should only call the token endpoint once due to caching
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_MetadataTeams_Success() {
	// Setup - User with team assignment AND metadata teams
	email := "user.with.metadata@example.com"
	teamID := uuid.New()

	// Create metadata with ai_core_member_of field
	metadata := map[string]interface{}{
		"ai_core_member_of": []string{"team-gamma", "team-delta"},
	}
	metadataJSON, _ := json.Marshal(metadata)

	member := &models.User{
		TeamID:   &teamID,
		TeamRole: models.TeamRoleMember,
		Metadata: metadataJSON,
	}

	team := &models.Team{
		BaseModel: models.BaseModel{ID: teamID, Name: "team-alpha"},
		Owner:     "team-alpha", // User's assigned team
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"GET:/v2/lm/deployments": {
			StatusCode: 200,
			Body: `{
				"count": 1,
				"resources": [
					{
						"id": "deployment-1",
						"configurationId": "config-1",
						"status": "RUNNING",
						"statusMessage": "Deployment is running",
						"deploymentUrl": "https://api.example.com/v1/deployments/deployment-1",
						"createdAt": "2023-01-01T00:00:00Z",
						"modifiedAt": "2023-01-01T01:00:00Z"
					}
				]
			}`,
		},
	}
	suite.setupMockServer(responses)
	// Setup credentials for assigned team + metadata teams
	suite.setupCredentials([]string{"team-alpha", "team-gamma", "team-delta"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByID", teamID).Return(team, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(3, result.Count)     // 1 deployment from each of 3 teams
	suite.Len(result.Deployments, 3) // Should have 3 teams: assigned team + 2 metadata teams

	teamNames := make([]string, len(result.Deployments))
	for i, td := range result.Deployments {
		teamNames[i] = td.Team
	}
	suite.Contains(teamNames, "team-alpha") // Assigned team
	suite.Contains(teamNames, "team-gamma") // Metadata team
	suite.Contains(teamNames, "team-delta") // Metadata team
}

func (suite *AICoreServiceTestSuite) TestGetDeployments_MetadataOnly_Success() {
	// Setup - User with NO team assignment but WITH metadata teams
	email := "metadata.only@example.com"

	// Create metadata with ai_core_member_of field
	metadata := map[string]interface{}{
		"ai_core_member_of": []string{"team-gamma", "team-delta"},
	}
	metadataJSON, _ := json.Marshal(metadata)

	member := &models.User{
		TeamID:   nil, // No team assignment
		TeamRole: models.TeamRoleMember,
		Metadata: metadataJSON,
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"GET:/v2/lm/deployments": {
			StatusCode: 200,
			Body: `{
				"count": 1,
				"resources": [
					{
						"id": "deployment-1",
						"configurationId": "config-1",
						"status": "RUNNING",
						"statusMessage": "Deployment is running",
						"deploymentUrl": "https://api.example.com/v1/deployments/deployment-1",
						"createdAt": "2023-01-01T00:00:00Z",
						"modifiedAt": "2023-01-01T01:00:00Z"
					}
				]
			}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-gamma", "team-delta"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.GetDeployments(c)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(2, result.Count)     // 1 deployment from each of 2 metadata teams
	suite.Len(result.Deployments, 2) // Should have 2 teams from metadata

	teamNames := make([]string, len(result.Deployments))
	for i, td := range result.Deployments {
		teamNames[i] = td.Team
	}
	suite.Contains(teamNames, "team-gamma")
	suite.Contains(teamNames, "team-delta")
}

func (suite *AICoreServiceTestSuite) TestCreateDeployment_WithConfigurationID_Success() {
	// Setup
	email := "team.member@example.com"
	teamID := uuid.New()

	member := &models.User{
		TeamID:   &teamID,
		TeamRole: models.TeamRoleMember,
	}

	team := &models.Team{
		BaseModel: models.BaseModel{ID: teamID, Name: "team-alpha"},
		Owner:     "team-alpha",
	}

	configID := "config-123"
	deploymentRequest := &AICoreDeploymentRequest{
		ConfigurationID: &configID,
		TTL:             "1h",
	}

	// Setup mock server responses
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"POST:/v2/lm/deployments": {
			StatusCode: 202,
			Body: `{
				"id": "deployment-123",
				"message": "Deployment created successfully",
				"deploymentUrl": "https://api.example.com/v1/deployments/deployment-123",
				"status": "PENDING",
				"ttl": "1h"
			}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-alpha"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByID", teamID).Return(team, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.CreateDeployment(c, deploymentRequest)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("deployment-123", result.ID)
	suite.Equal("Deployment created successfully", result.Message)
	suite.Equal("PENDING", result.Status)
}

func (suite *AICoreServiceTestSuite) TestCreateDeployment_WithConfigurationRequest_Success() {
	// Setup
	email := "team.member@example.com"
	teamID := uuid.New()

	member := &models.User{
		TeamID:   &teamID,
		TeamRole: models.TeamRoleMember,
	}

	team := &models.Team{
		BaseModel: models.BaseModel{ID: teamID, Name: "team-alpha"},
		Owner:     "team-alpha",
	}

	deploymentRequest := &AICoreDeploymentRequest{
		ConfigurationRequest: &AICoreConfigurationRequest{
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

	// Setup mock server - first create config, then create deployment
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"POST:/v2/lm/configurations": {
			StatusCode: 201,
			Body: `{
				"id": "config-456",
				"message": "Configuration created successfully"
			}`,
		},
		"POST:/v2/lm/deployments": {
			StatusCode: 202,
			Body: `{
				"id": "deployment-456",
				"message": "Deployment created successfully",
				"deploymentUrl": "https://api.example.com/v1/deployments/deployment-456",
				"status": "PENDING",
				"ttl": "2h"
			}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-alpha"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByID", teamID).Return(team, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.CreateDeployment(c, deploymentRequest)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal("deployment-456", result.ID)
	suite.Equal("Deployment created successfully", result.Message)
	suite.Equal("PENDING", result.Status)
}

func (suite *AICoreServiceTestSuite) TestCreateDeployment_BothFieldsProvided_Error() {
	// Setup
	email := "team.member@example.com"

	member := &models.User{
		TeamID:   nil, // not needed for validation test
		TeamRole: models.TeamRoleMember,
	}

	configID := "config-123"
	deploymentRequest := &AICoreDeploymentRequest{
		ConfigurationID: &configID,
		ConfigurationRequest: &AICoreConfigurationRequest{
			Name:         "my-llm-config",
			ExecutableID: "aicore-llm",
			ScenarioID:   "foundation-models",
		},
		TTL: "1h",
	}

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.CreateDeployment(c, deploymentRequest)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "configurationId and configurationRequest cannot both be provided")
}

func (suite *AICoreServiceTestSuite) TestCreateDeployment_NeitherFieldProvided_Error() {
	// Setup
	email := "team.member@example.com"

	member := &models.User{
		TeamID:   nil, // not needed for validation test
		TeamRole: models.TeamRoleMember,
	}

	deploymentRequest := &AICoreDeploymentRequest{
		TTL: "1h",
	}

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.CreateDeployment(c, deploymentRequest)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "either configurationId or configurationRequest must be provided")
}

func (suite *AICoreServiceTestSuite) TestCreateDeployment_ConfigurationCreationFails_Error() {
	// Setup
	email := "team.member@example.com"
	teamID := uuid.New()

	member := &models.User{
		TeamID:   &teamID,
		TeamRole: models.TeamRoleMember,
	}

	team := &models.Team{
		BaseModel: models.BaseModel{ID: teamID, Name: "team-alpha"},
		Owner:     "team-alpha",
	}

	deploymentRequest := &AICoreDeploymentRequest{
		ConfigurationRequest: &AICoreConfigurationRequest{
			Name:         "my-llm-config",
			ExecutableID: "aicore-llm",
			ScenarioID:   "foundation-models",
		},
		TTL: "1h",
	}

	// Setup mock server responses - configuration creation fails
	responses := map[string]mockResponse{
		"POST:/oauth/token": {
			StatusCode: 200,
			Body:       `{"access_token": "test-token", "token_type": "Bearer", "expires_in": 3600}`,
		},
		"POST:/v2/lm/configurations": {
			StatusCode: 400,
			Body:       `{"error": "Invalid configuration request"}`,
		},
	}
	suite.setupMockServer(responses)
	suite.setupCredentials([]string{"team-alpha"})

	// Setup mocks
	suite.userRepo.On("GetByEmail", email).Return(member, nil)
	suite.teamRepo.On("GetByID", teamID).Return(team, nil)

	// Execute
	c := suite.createGinContext(email)
	result, err := suite.service.CreateDeployment(c, deploymentRequest)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "failed to create configuration")
}

func TestAICoreServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AICoreServiceTestSuite))
}
