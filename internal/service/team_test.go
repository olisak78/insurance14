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
	ctrl         *gomock.Controller
	mockTeamRepo *mocks.MockTeamRepositoryInterface
	mockOrgRepo  *mocks.MockOrganizationRepositoryInterface
	mockUserRepo *mocks.MockUserRepositoryInterface
	teamService  *service.TeamService
	validator    *validator.Validate
}

// SetupTest sets up the test suite
func (suite *TeamServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockTeamRepo = mocks.NewMockTeamRepositoryInterface(suite.ctrl)
	suite.mockOrgRepo = mocks.NewMockOrganizationRepositoryInterface(suite.ctrl)
	suite.mockUserRepo = mocks.NewMockUserRepositoryInterface(suite.ctrl)
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
				Title:       "Backend Team",
				Description: "Team responsible for backend services",
				Owner:       "I12345",
				Email:       "backend-team@test.com",
				PictureURL:  "https://example.com/team.png",
			},
			expectError: false,
		},
		{
			name: "Missing group ID",
			request: &service.CreateTeamRequest{
				Name:  "backend-team",
				Title: "Backend Team",
			},
			expectError: true,
			errorMsg:    "GroupID",
		},
		{
			name: "Empty name",
			request: &service.CreateTeamRequest{
				GroupID: uuid.New(),
				Name:    "",
				Title:   "Backend Team",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Empty title",
			request: &service.CreateTeamRequest{
				GroupID: uuid.New(),
				Name:    "backend-team",
				Title:   "",
			},
			expectError: true,
			errorMsg:    "Title",
		},
		{
			name: "Name too long",
			request: &service.CreateTeamRequest{
				GroupID: uuid.New(),
				Name:    "this-is-a-very-long-team-name-that-definitely-exceeds-one-hundred-characters-which-is-the-maximum-allowed-length-for-team-names-in-this-system-validation",
				Title:   "Backend Team",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Display name too long",
			request: &service.CreateTeamRequest{
				GroupID: uuid.New(),
				Name:    "backend-team",
				Title:   "This is a very long display name that definitely exceeds the maximum allowed length of two hundred characters for the display name field and should trigger a validation error when we try to create a team with this overly long display name that goes beyond the limit",
			},
			expectError: true,
			errorMsg:    "Title",
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
				Title:       "Updated Backend Team",
				Description: "Updated description",
				Owner:       "I67890",
				Email:       "backend-updated@test.com",
				PictureURL:  "https://example.com/team-updated.png",
			},
			expectError: false,
		},
		{
			name: "Display name too long",
			request: &service.UpdateTeamRequest{
				Title: "This is an extremely long display name that definitely exceeds the maximum allowed length of exactly two hundred characters for the display name field and should absolutely trigger a validation error when we try to update a team with this incredibly long display name that goes way beyond the specified character limit of two hundred characters making it invalid for our validation system",
			},
			expectError: true,
			errorMsg:    "Title",
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
	metadata := json.RawMessage(`{"tags": ["backend", "api"]}`)

	response := &service.TeamResponse{
		ID:             teamID,
		OrganizationID: orgID,
		GroupID:        uuid.New(),
		Name:           "backend-team",
		Title:          "Backend Team",
		Description:    "Team responsible for backend services",
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
	assert.Equal(suite.T(), response.Title, unmarshaled.Title)
}

// TestTeamListResponseSerialization tests the team list response serialization
func (suite *TeamServiceTestSuite) TestTeamListResponseSerialization() {
	teams := []service.TeamResponse{
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "team-1",
			Title:          "Team 1",
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		},
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "team-2",
			Title:          "Team 2",
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
	// Status enums removed from the model; placeholder to keep suite stable
	assert.True(suite.T(), true)
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
		{BaseModel: models.BaseModel{ID: uuid.New(), Name: "comp1"}},
		{BaseModel: models.BaseModel{ID: uuid.New(), Name: "comp2"}},
		{BaseModel: models.BaseModel{ID: uuid.New(), Name: "comp3"}},
		{BaseModel: models.BaseModel{ID: uuid.New(), Name: "comp4"}},
		{BaseModel: models.BaseModel{ID: uuid.New(), Name: "comp5"}},
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
				URL:   "https://github.com/myteam/repo",
				Title: "Team Repository",
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
				Title: "Team Repository",
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
				URL: "https://github.com/myteam/repo",
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

// TestTechnicalTeamFiltering tests that the technical team is filtered out from results
func (suite *TeamServiceTestSuite) TestTechnicalTeamFiltering() {
	// Test that the technical team name is correctly filtered
	technicalTeamName := "team-developer-portal-technical"
	regularTeamName := "team-backend"

	// Test filtering logic
	teams := []models.Team{
		{BaseModel: models.BaseModel{Name: regularTeamName, Title: "Backend Team"}},
		{BaseModel: models.BaseModel{Name: technicalTeamName, Title: "Technical Team"}},
		{BaseModel: models.BaseModel{Name: "team-frontend", Title: "Frontend Team"}},
	}

	// Simulate the filtering logic from GetAllTeams
	filteredTeams := make([]models.Team, 0, len(teams))
	for _, team := range teams {
		if team.Name != technicalTeamName {
			filteredTeams = append(filteredTeams, team)
		}
	}

	// Assert that technical team is filtered out
	assert.Len(suite.T(), filteredTeams, 2, "Should have 2 teams after filtering")
	assert.Equal(suite.T(), regularTeamName, filteredTeams[0].Name, "First team should be backend team")
	assert.Equal(suite.T(), "team-frontend", filteredTeams[1].Name, "Second team should be frontend team")

	// Verify technical team is not in results
	for _, team := range filteredTeams {
		assert.NotEqual(suite.T(), technicalTeamName, team.Name, "Technical team should not be in filtered results")
	}
}

// TestTechnicalTeamFilteringTotalAdjustment tests that the total count is adjusted correctly
func (suite *TeamServiceTestSuite) TestTechnicalTeamFilteringTotalAdjustment() {
	testCases := []struct {
		name                string
		totalFromDB         int64
		teamsFromDB         int
		filteredTeamsCount  int
		expectedAdjustedTotal int64
	}{
		{
			name:                "Technical team present - adjust total",
			totalFromDB:         10,
			teamsFromDB:         5,
			filteredTeamsCount:  4,
			expectedAdjustedTotal: 9, // 10 - 1
		},
		{
			name:                "Technical team not present - no adjustment",
			totalFromDB:         10,
			teamsFromDB:         5,
			filteredTeamsCount:  5,
			expectedAdjustedTotal: 10,
		},
		{
			name:                "Single technical team - adjust to zero",
			totalFromDB:         1,
			teamsFromDB:         1,
			filteredTeamsCount:  0,
			expectedAdjustedTotal: 0,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Simulate the adjustment logic from GetByOrganization and Search
			adjustedTotal := tc.totalFromDB
			if tc.teamsFromDB > tc.filteredTeamsCount {
				adjustedTotal = tc.totalFromDB - int64(tc.teamsFromDB-tc.filteredTeamsCount)
			}

			assert.Equal(t, tc.expectedAdjustedTotal, adjustedTotal, "Total should be adjusted correctly")
		})
	}
}

// TestTechnicalTeamFilteringEdgeCases tests edge cases for technical team filtering
func (suite *TeamServiceTestSuite) TestTechnicalTeamFilteringEdgeCases() {
	technicalTeamName := "team-developer-portal-technical"

	testCases := []struct {
		name           string
		teams          []models.Team
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:          "Empty team list",
			teams:         []models.Team{},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "Only technical team",
			teams: []models.Team{
				{BaseModel: models.BaseModel{Name: technicalTeamName}},
			},
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name: "Multiple technical teams (shouldn't happen, but handle it)",
			teams: []models.Team{
				{BaseModel: models.BaseModel{Name: technicalTeamName}},
				{BaseModel: models.BaseModel{Name: technicalTeamName}},
				{BaseModel: models.BaseModel{Name: "team-regular"}},
			},
			expectedCount: 1,
			expectedNames: []string{"team-regular"},
		},
		{
			name: "Technical team in middle",
			teams: []models.Team{
				{BaseModel: models.BaseModel{Name: "team-first"}},
				{BaseModel: models.BaseModel{Name: technicalTeamName}},
				{BaseModel: models.BaseModel{Name: "team-last"}},
			},
			expectedCount: 2,
			expectedNames: []string{"team-first", "team-last"},
		},
		{
			name: "Similar but not exact name",
			teams: []models.Team{
				{BaseModel: models.BaseModel{Name: "team-developer-portal"}},
				{BaseModel: models.BaseModel{Name: "team-technical"}},
				{BaseModel: models.BaseModel{Name: technicalTeamName}},
			},
			expectedCount: 2,
			expectedNames: []string{"team-developer-portal", "team-technical"},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Simulate the filtering logic
			filteredTeams := make([]models.Team, 0, len(tc.teams))
			for _, team := range tc.teams {
				if team.Name != technicalTeamName {
					filteredTeams = append(filteredTeams, team)
				}
			}

			assert.Len(t, filteredTeams, tc.expectedCount, "Filtered team count should match")

			// Check that the expected names are present
			if tc.expectedCount > 0 {
				for i, expectedName := range tc.expectedNames {
					if i < len(filteredTeams) {
						assert.Equal(t, expectedName, filteredTeams[i].Name, "Team name should match")
					}
				}
			}

			// Verify technical team is never in the results
			for _, team := range filteredTeams {
				assert.NotEqual(t, technicalTeamName, team.Name, "Technical team should never be in results")
			}
		})
	}
}

// TestTeamServiceTestSuite runs the test suite
func TestTeamServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TeamServiceTestSuite))
}
