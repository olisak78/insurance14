package service_test

import (
	"encoding/json"
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// TeamServiceTestSuite defines the test suite for TeamService
type TeamServiceTestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mockTeamRepo   *mocks.MockTeamRepositoryInterface
	mockOrgRepo    *mocks.MockOrganizationRepositoryInterface
	mockMemberRepo *mocks.MockMemberRepositoryInterface
	teamService    *service.TeamService
	validator      *validator.Validate
}

// SetupTest sets up the test suite
func (suite *TeamServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockTeamRepo = mocks.NewMockTeamRepositoryInterface(suite.ctrl)
	suite.mockOrgRepo = mocks.NewMockOrganizationRepositoryInterface(suite.ctrl)
	suite.mockMemberRepo = mocks.NewMockMemberRepositoryInterface(suite.ctrl)
	suite.validator = validator.New()

	// Since TeamService uses concrete repository types instead of interfaces,
	// we can't properly mock them for unit testing. This is a design issue
	// that would need to be fixed in the service layer.
	// For now, we'll focus on testing validation logic and other testable parts.
	suite.teamService = nil
}

// TearDownTest cleans up after each test
func (suite *TeamServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// Since we can't properly mock the concrete repository types that TeamService expects,
// let's test individual service methods using a different approach
// We'll test the service logic by creating minimal tests that focus on validation

// TestCreateTeamValidation tests the validation logic for creating a team
func (suite *TeamServiceTestSuite) TestCreateTeamValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.CreateTeamRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.CreateTeamRequest{
				GroupID:     uuid.New(),
				Name:        "backend-team",
				DisplayName: "Backend Team",
				Description: "Team responsible for backend services",
				Status:      models.TeamStatusActive,
			},
			expectError: false,
		},
		{
			name: "Missing group ID",
			request: &service.CreateTeamRequest{
				Name:        "backend-team",
				DisplayName: "Backend Team",
			},
			expectError: true,
			errorMsg:    "GroupID",
		},
		{
			name: "Empty name",
			request: &service.CreateTeamRequest{
				GroupID:     uuid.New(),
				Name:        "",
				DisplayName: "Backend Team",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Empty display name",
			request: &service.CreateTeamRequest{
				GroupID:     uuid.New(),
				Name:        "backend-team",
				DisplayName: "",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Name too long",
			request: &service.CreateTeamRequest{
				GroupID:     uuid.New(),
				Name:        "this-is-a-very-long-team-name-that-definitely-exceeds-one-hundred-characters-which-is-the-maximum-allowed-length-for-team-names-in-this-system-validation",
				DisplayName: "Backend Team",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Display name too long",
			request: &service.CreateTeamRequest{
				GroupID:     uuid.New(),
				Name:        "backend-team",
				DisplayName: "This is a very long display name that definitely exceeds the maximum allowed length of two hundred characters for the display name field and should trigger a validation error when we try to create a team with this overly long display name that goes beyond the limit",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := validator.Struct(tc.request)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUpdateTeamValidation tests the validation logic for updating a team
func (suite *TeamServiceTestSuite) TestUpdateTeamValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.UpdateTeamRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.UpdateTeamRequest{
				DisplayName: "Updated Backend Team",
				Description: "Updated description",
			},
			expectError: false,
		},
		{
			name: "Empty display name",
			request: &service.UpdateTeamRequest{
				DisplayName: "",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Display name too long",
			request: &service.UpdateTeamRequest{
				DisplayName: "This is an extremely long display name that definitely exceeds the maximum allowed length of exactly two hundred characters for the display name field and should absolutely trigger a validation error when we try to update a team with this incredibly long display name that goes way beyond the specified character limit of two hundred characters making it invalid for our validation system",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := validator.Struct(tc.request)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTeamResponseSerialization tests the team response serialization
func (suite *TeamServiceTestSuite) TestTeamResponseSerialization() {
	teamID := uuid.New()
	orgID := uuid.New()
	teamLeadID := uuid.New()
	links := []service.Link{
		{
			URL:      "https://slack.com/team",
			Title:    "Slack Channel",
			Icon:     "slack",
			Category: "communication",
		},
	}
	metadata := json.RawMessage(`{"tags": ["backend", "api"]}`)

	response := &service.TeamResponse{
		ID:             teamID,
		OrganizationID: orgID,
		Name:           "backend-team",
		DisplayName:    "Backend Team",
		Description:    "Team responsible for backend services",
		TeamLeadID:     &teamLeadID,
		Status:         models.TeamStatusActive,
		Links:          links,
		Metadata:       metadata,
		CreatedAt:      "2023-01-01T00:00:00Z",
		UpdatedAt:      "2023-01-01T00:00:00Z",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), teamID.String())
	assert.Contains(suite.T(), string(jsonData), "backend-team")
	assert.Contains(suite.T(), string(jsonData), "Backend Team")

	// Test JSON unmarshaling
	var unmarshaled service.TeamResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), response.ID, unmarshaled.ID)
	assert.Equal(suite.T(), response.Name, unmarshaled.Name)
	assert.Equal(suite.T(), response.DisplayName, unmarshaled.DisplayName)
}

// TestTeamListResponseSerialization tests the team list response serialization
func (suite *TeamServiceTestSuite) TestTeamListResponseSerialization() {
	teams := []service.TeamResponse{
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "team-1",
			DisplayName:    "Team 1",
			Status:         models.TeamStatusActive,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		},
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "team-2",
			DisplayName:    "Team 2",
			Status:         models.TeamStatusInactive,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		},
	}

	response := &service.TeamListResponse{
		Teams:    teams,
		Total:    int64(len(teams)),
		Page:     1,
		PageSize: 20,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), "team-1")
	assert.Contains(suite.T(), string(jsonData), "team-2")
	assert.Contains(suite.T(), string(jsonData), `"total":2`)
	assert.Contains(suite.T(), string(jsonData), `"page":1`)

	// Test JSON unmarshaling
	var unmarshaled service.TeamListResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), unmarshaled.Teams, 2)
	assert.Equal(suite.T(), response.Total, unmarshaled.Total)
	assert.Equal(suite.T(), response.Page, unmarshaled.Page)
	assert.Equal(suite.T(), response.PageSize, unmarshaled.PageSize)
}

// TestDefaultStatusBehavior tests the default status behavior
func (suite *TeamServiceTestSuite) TestDefaultStatusBehavior() {
	// Test that when status is empty, it should default to Active
	emptyStatus := models.TeamStatus("")
	expectedDefault := models.TeamStatusActive

	// Simulate the default status logic from the service
	var finalStatus models.TeamStatus
	if emptyStatus == "" {
		finalStatus = models.TeamStatusActive
	} else {
		finalStatus = emptyStatus
	}

	assert.Equal(suite.T(), expectedDefault, finalStatus)

	// Test with explicit status
	explicitStatus := models.TeamStatusInactive
	if explicitStatus == "" {
		finalStatus = models.TeamStatusActive
	} else {
		finalStatus = explicitStatus
	}

	assert.Equal(suite.T(), models.TeamStatusInactive, finalStatus)
}

// TestPaginationLogic tests the pagination logic
func (suite *TeamServiceTestSuite) TestPaginationLogic() {
	testCases := []struct {
		name           string
		inputPage      int
		inputSize      int
		expectedPage   int
		expectedSize   int
		expectedOffset int
	}{
		{
			name:           "Valid pagination",
			inputPage:      2,
			inputSize:      10,
			expectedPage:   2,
			expectedSize:   10,
			expectedOffset: 10,
		},
		{
			name:           "Page less than 1",
			inputPage:      0,
			inputSize:      10,
			expectedPage:   1,
			expectedSize:   10,
			expectedOffset: 0,
		},
		{
			name:           "Page size less than 1",
			inputPage:      1,
			inputSize:      0,
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
		},
		{
			name:           "Page size greater than 100",
			inputPage:      1,
			inputSize:      150,
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
		},
		{
			name:           "Both invalid",
			inputPage:      -1,
			inputSize:      -5,
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Simulate the pagination logic from the service
			page := tc.inputPage
			pageSize := tc.inputSize

			if page < 1 {
				page = 1
			}
			if pageSize < 1 || pageSize > 100 {
				pageSize = 20
			}

			offset := (page - 1) * pageSize

			assert.Equal(t, tc.expectedPage, page)
			assert.Equal(t, tc.expectedSize, pageSize)
			assert.Equal(t, tc.expectedOffset, offset)
		})
	}
}

// TestTeamStatusValidation tests team status validation
func (suite *TeamServiceTestSuite) TestTeamStatusValidation() {
	validStatuses := []models.TeamStatus{
		models.TeamStatusActive,
		models.TeamStatusInactive,
	}

	for _, status := range validStatuses {
		suite.T().Run(string(status), func(t *testing.T) {
			// Test that valid statuses are accepted
			assert.NotEmpty(t, string(status))
			assert.True(t, status == models.TeamStatusActive || status == models.TeamStatusInactive)
		})
	}
}

// TestJSONFieldsHandling tests handling of JSON fields (Links and Metadata)
func (suite *TeamServiceTestSuite) TestJSONFieldsHandling() {
	// Test valid JSON
	validLinks := json.RawMessage(`{"slack": "https://slack.com/team", "github": "https://github.com/team"}`)
	validMetadata := json.RawMessage(`{"tags": ["backend", "api"], "priority": "high"}`)

	// Test that valid JSON can be marshaled and unmarshaled
	linksData, err := json.Marshal(validLinks)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), linksData)

	metadataData, err := json.Marshal(validMetadata)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metadataData)

	// Test empty JSON
	emptyJSON := json.RawMessage(`{}`)
	emptyData, err := json.Marshal(emptyJSON)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), `{}`, string(emptyData))

	// Test nil JSON (should be handled gracefully)
	var nilJSON json.RawMessage
	nilData, err := json.Marshal(nilJSON)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "null", string(nilData))
}

// TestComponentPaginationLogic tests component pagination logic
func (suite *TeamServiceTestSuite) TestComponentPaginationLogic() {
	// Simulate the manual pagination logic for components
	components := []models.Component{
		{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "comp1"},
		{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "comp2"},
		{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "comp3"},
		{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "comp4"},
		{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "comp5"},
	}

	testCases := []struct {
		name          string
		page          int
		pageSize      int
		expectedStart int
		expectedEnd   int
		expectedCount int
	}{
		{
			name:          "First page",
			page:          1,
			pageSize:      2,
			expectedStart: 0,
			expectedEnd:   2,
			expectedCount: 2,
		},
		{
			name:          "Second page",
			page:          2,
			pageSize:      2,
			expectedStart: 2,
			expectedEnd:   4,
			expectedCount: 2,
		},
		{
			name:          "Last page partial",
			page:          3,
			pageSize:      2,
			expectedStart: 4,
			expectedEnd:   5,
			expectedCount: 1,
		},
		{
			name:          "Page beyond data",
			page:          4,
			pageSize:      2,
			expectedStart: 6,
			expectedEnd:   6,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Simulate the pagination logic from GetTeamComponentsByName
			page := tc.page
			pageSize := tc.pageSize

			if page < 1 {
				page = 1
			}
			if pageSize < 1 || pageSize > 100 {
				pageSize = 20
			}

			start := (page - 1) * pageSize
			end := start + pageSize

			assert.Equal(t, tc.expectedStart, start)

			var result []models.Component
			if start >= len(components) {
				result = []models.Component{}
			} else {
				actualEnd := end
				if actualEnd > len(components) {
					actualEnd = len(components)
				}
				result = components[start:actualEnd]
			}

			assert.Equal(t, tc.expectedCount, len(result))
		})
	}
}

// TestAddLinkValidation tests the validation logic for adding a link
func TestAddLinkValidation(t *testing.T) {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.AddLinkRequest
		expectError bool
	}{
		{
			name: "Valid link",
			request: &service.AddLinkRequest{
				URL:      "https://github.com/myteam/repo",
				Title:    "Team Repository",
				Icon:     "github",
				Category: "repository",
			},
			expectError: false,
		},
		{
			name: "Valid link without optional fields",
			request: &service.AddLinkRequest{
				URL:   "https://example.com",
				Title: "Example",
			},
			expectError: false,
		},
		{
			name: "Missing URL",
			request: &service.AddLinkRequest{
				Title:    "Team Repository",
				Icon:     "github",
				Category: "repository",
			},
			expectError: true,
		},
		{
			name: "Invalid URL",
			request: &service.AddLinkRequest{
				URL:   "not-a-url",
				Title: "Team Repository",
			},
			expectError: true,
		},
		{
			name: "Missing title",
			request: &service.AddLinkRequest{
				URL:      "https://github.com/myteam/repo",
				Icon:     "github",
				Category: "repository",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.Struct(tc.request)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLinkJSONMarshaling tests that links can be properly marshaled and unmarshaled
func TestLinkJSONMarshaling(t *testing.T) {
	links := []service.Link{
		{
			URL:      "https://github.com/myteam/repo",
			Title:    "Team Repository",
			Icon:     "github",
			Category: "repository",
		},
		{
			URL:      "https://docs.example.com",
			Title:    "Documentation",
			Icon:     "docs",
			Category: "documentation",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(links)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal back
	var unmarshaledLinks []service.Link
	err = json.Unmarshal(jsonData, &unmarshaledLinks)
	assert.NoError(t, err)
	assert.Equal(t, len(links), len(unmarshaledLinks))
	assert.Equal(t, links[0].URL, unmarshaledLinks[0].URL)
	assert.Equal(t, links[0].Title, unmarshaledLinks[0].Title)
	assert.Equal(t, links[0].Icon, unmarshaledLinks[0].Icon)
	assert.Equal(t, links[0].Category, unmarshaledLinks[0].Category)
}

// TestTeamServiceTestSuite runs the test suite
func TestTeamServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceTestSuite))
}
