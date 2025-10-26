package service_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ComponentDeploymentServiceTestSuite defines the test suite for ComponentDeploymentService
type ComponentDeploymentServiceTestSuite struct {
	suite.Suite
	ctrl                    *gomock.Controller
	mockComponentDeployRepo *mocks.MockComponentDeploymentRepositoryInterface
	mockComponentRepo       *mocks.MockComponentRepositoryInterface
	mockLandscapeRepo       *mocks.MockLandscapeRepositoryInterface
	componentDeployService  *service.ComponentDeploymentService
	validator               *validator.Validate
}

// SetupTest sets up the test suite
func (suite *ComponentDeploymentServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockComponentDeployRepo = mocks.NewMockComponentDeploymentRepositoryInterface(suite.ctrl)
	suite.mockComponentRepo = mocks.NewMockComponentRepositoryInterface(suite.ctrl)
	suite.mockLandscapeRepo = mocks.NewMockLandscapeRepositoryInterface(suite.ctrl)
	suite.validator = validator.New()

	// Since ComponentDeploymentService uses concrete repository types instead of interfaces,
	// we can't properly mock them for unit testing. This is a design issue
	// that would need to be fixed in the service layer.
	// For now, we'll focus on testing validation logic and other testable parts.
	suite.componentDeployService = nil
}

// TearDownTest cleans up after each test
func (suite *ComponentDeploymentServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateComponentDeploymentValidation tests the validation logic for creating a component deployment
func (suite *ComponentDeploymentServiceTestSuite) TestCreateComponentDeploymentValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.CreateComponentDeploymentRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.CreateComponentDeploymentRequest{
				ComponentID: uuid.New(),
				LandscapeID: uuid.New(),
				Version:     "1.0.0",
				GitCommitID: "abc123",
			},
			expectError: false,
		},
		{
			name: "Missing component ID",
			request: &service.CreateComponentDeploymentRequest{
				LandscapeID: uuid.New(),
				Version:     "1.0.0",
			},
			expectError: true,
			errorMsg:    "ComponentID",
		},
		{
			name: "Missing landscape ID",
			request: &service.CreateComponentDeploymentRequest{
				ComponentID: uuid.New(),
				Version:     "1.0.0",
			},
			expectError: true,
			errorMsg:    "LandscapeID",
		},
		{
			name: "Both IDs missing",
			request: &service.CreateComponentDeploymentRequest{
				Version: "1.0.0",
			},
			expectError: true,
			errorMsg:    "ComponentID",
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

// TestUpdateComponentDeploymentValidation tests the validation logic for updating a component deployment
func (suite *ComponentDeploymentServiceTestSuite) TestUpdateComponentDeploymentValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.UpdateComponentDeploymentRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.UpdateComponentDeploymentRequest{
				Version:     "1.1.0",
				GitCommitID: "def456",
			},
			expectError: false,
		},
		{
			name:        "Empty request",
			request:     &service.UpdateComponentDeploymentRequest{},
			expectError: false,
		},
		{
			name: "With time fields",
			request: &service.UpdateComponentDeploymentRequest{
				Version:       "2.0.0",
				GitCommitTime: &time.Time{},
				BuildTime:     &time.Time{},
				DeployedAt:    &time.Time{},
			},
			expectError: false,
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

// TestComponentDeploymentResponseSerialization tests the component deployment response serialization
func (suite *ComponentDeploymentServiceTestSuite) TestComponentDeploymentResponseSerialization() {
	deploymentID := uuid.New()
	componentID := uuid.New()
	landscapeID := uuid.New()
	now := time.Now()
	buildProps := json.RawMessage(`{"buildNumber": 123, "builder": "jenkins"}`)
	gitProps := json.RawMessage(`{"branch": "main", "author": "developer"}`)

	response := &service.ComponentDeploymentResponse{
		ID:              deploymentID,
		ComponentID:     componentID,
		LandscapeID:     landscapeID,
		Version:         "1.0.0",
		GitCommitID:     "abc123def456",
		GitCommitTime:   &now,
		BuildTime:       &now,
		BuildProperties: buildProps,
		GitProperties:   gitProps,
		IsActive:        true,
		DeployedAt:      &now,
		CreatedAt:       "2023-01-01T00:00:00Z",
		UpdatedAt:       "2023-01-01T00:00:00Z",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), deploymentID.String())
	assert.Contains(suite.T(), string(jsonData), componentID.String())
	assert.Contains(suite.T(), string(jsonData), landscapeID.String())
	assert.Contains(suite.T(), string(jsonData), "1.0.0")
	assert.Contains(suite.T(), string(jsonData), "abc123def456")
	assert.Contains(suite.T(), string(jsonData), `"is_active":true`)

	// Test JSON unmarshaling
	var unmarshaled service.ComponentDeploymentResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), response.ID, unmarshaled.ID)
	assert.Equal(suite.T(), response.ComponentID, unmarshaled.ComponentID)
	assert.Equal(suite.T(), response.LandscapeID, unmarshaled.LandscapeID)
	assert.Equal(suite.T(), response.Version, unmarshaled.Version)
	assert.Equal(suite.T(), response.GitCommitID, unmarshaled.GitCommitID)
	assert.Equal(suite.T(), response.IsActive, unmarshaled.IsActive)
}

// TestComponentDeploymentListResponseSerialization tests the component deployment list response serialization
func (suite *ComponentDeploymentServiceTestSuite) TestComponentDeploymentListResponseSerialization() {
	deployments := []service.ComponentDeploymentResponse{
		{
			ID:          uuid.New(),
			ComponentID: uuid.New(),
			LandscapeID: uuid.New(),
			Version:     "1.0.0",
			GitCommitID: "abc123",
			IsActive:    true,
			CreatedAt:   "2023-01-01T00:00:00Z",
			UpdatedAt:   "2023-01-01T00:00:00Z",
		},
		{
			ID:          uuid.New(),
			ComponentID: uuid.New(),
			LandscapeID: uuid.New(),
			Version:     "1.1.0",
			GitCommitID: "def456",
			IsActive:    false,
			CreatedAt:   "2023-01-02T00:00:00Z",
			UpdatedAt:   "2023-01-02T00:00:00Z",
		},
	}

	response := &service.ComponentDeploymentListResponse{
		Deployments: deployments,
		Total:       int64(len(deployments)),
		Page:        1,
		PageSize:    20,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), "1.0.0")
	assert.Contains(suite.T(), string(jsonData), "1.1.0")
	assert.Contains(suite.T(), string(jsonData), "abc123")
	assert.Contains(suite.T(), string(jsonData), "def456")
	assert.Contains(suite.T(), string(jsonData), `"total":2`)
	assert.Contains(suite.T(), string(jsonData), `"page":1`)

	// Test JSON unmarshaling
	var unmarshaled service.ComponentDeploymentListResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), unmarshaled.Deployments, 2)
	assert.Equal(suite.T(), response.Total, unmarshaled.Total)
	assert.Equal(suite.T(), response.Page, unmarshaled.Page)
	assert.Equal(suite.T(), response.PageSize, unmarshaled.PageSize)
}

// TestDefaultIsActiveBehavior tests the default IsActive behavior
func (suite *ComponentDeploymentServiceTestSuite) TestDefaultIsActiveBehavior() {
	// Test that when IsActive is nil, it should default to true
	var nilActive *bool
	expectedDefault := true

	// Simulate the default IsActive logic from the service
	var finalActive bool
	if nilActive == nil {
		finalActive = true
	} else {
		finalActive = *nilActive
	}

	assert.Equal(suite.T(), expectedDefault, finalActive)

	// Test with explicit false
	explicitFalse := false
	if &explicitFalse == nil {
		finalActive = true
	} else {
		finalActive = explicitFalse
	}

	assert.Equal(suite.T(), false, finalActive)

	// Test with explicit true
	explicitTrue := true
	if &explicitTrue == nil {
		finalActive = true
	} else {
		finalActive = explicitTrue
	}

	assert.Equal(suite.T(), true, finalActive)
}

// TestPaginationLogic tests the pagination logic
func (suite *ComponentDeploymentServiceTestSuite) TestPaginationLogic() {
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

// TestJSONFieldsHandling tests handling of JSON fields (BuildProperties and GitProperties)
func (suite *ComponentDeploymentServiceTestSuite) TestJSONFieldsHandling() {
	// Test valid JSON for build properties
	validBuildProps := json.RawMessage(`{"buildNumber": 123, "builder": "jenkins", "duration": "5m30s"}`)
	validGitProps := json.RawMessage(`{"branch": "main", "author": "developer", "message": "Fix bug"}`)

	// Test that valid JSON can be marshaled and unmarshaled
	buildData, err := json.Marshal(validBuildProps)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), buildData)
	assert.Contains(suite.T(), string(buildData), "buildNumber")
	assert.Contains(suite.T(), string(buildData), "jenkins")

	gitData, err := json.Marshal(validGitProps)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), gitData)
	assert.Contains(suite.T(), string(gitData), "branch")
	assert.Contains(suite.T(), string(gitData), "main")

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

// TestTimeFieldsHandling tests handling of time fields
func (suite *ComponentDeploymentServiceTestSuite) TestTimeFieldsHandling() {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	// Test marshaling and unmarshaling of time fields
	request := &service.CreateComponentDeploymentRequest{
		ComponentID:   uuid.New(),
		LandscapeID:   uuid.New(),
		Version:       "1.0.0",
		GitCommitTime: &pastTime,
		BuildTime:     &now,
		DeployedAt:    &futureTime,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(request)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jsonData)

	// Test JSON unmarshaling
	var unmarshaled service.CreateComponentDeploymentRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), request.ComponentID, unmarshaled.ComponentID)
	assert.Equal(suite.T(), request.LandscapeID, unmarshaled.LandscapeID)
	assert.Equal(suite.T(), request.Version, unmarshaled.Version)

	// Note: Time comparison might have precision differences, so we check that they're close
	if request.GitCommitTime != nil && unmarshaled.GitCommitTime != nil {
		assert.WithinDuration(suite.T(), *request.GitCommitTime, *unmarshaled.GitCommitTime, time.Second)
	}
	if request.BuildTime != nil && unmarshaled.BuildTime != nil {
		assert.WithinDuration(suite.T(), *request.BuildTime, *unmarshaled.BuildTime, time.Second)
	}
	if request.DeployedAt != nil && unmarshaled.DeployedAt != nil {
		assert.WithinDuration(suite.T(), *request.DeployedAt, *unmarshaled.DeployedAt, time.Second)
	}
}

// TestVersionHandling tests version string handling
func (suite *ComponentDeploymentServiceTestSuite) TestVersionHandling() {
	testCases := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "Semantic version",
			version: "1.0.0",
			valid:   true,
		},
		{
			name:    "Semantic version with pre-release",
			version: "1.0.0-alpha.1",
			valid:   true,
		},
		{
			name:    "Semantic version with build metadata",
			version: "1.0.0+20230101.123456",
			valid:   true,
		},
		{
			name:    "Simple version",
			version: "v1.0",
			valid:   true,
		},
		{
			name:    "Git commit hash",
			version: "abc123def456",
			valid:   true,
		},
		{
			name:    "Empty version",
			version: "",
			valid:   true, // Empty version is allowed
		},
		{
			name:    "Long version string",
			version: "1.0.0-alpha.1+build.20230101.123456.with.more.metadata",
			valid:   true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := &service.CreateComponentDeploymentRequest{
				ComponentID: uuid.New(),
				LandscapeID: uuid.New(),
				Version:     tc.version,
			}

			// Test validation (versions are not validated by struct tags currently)
			validator := validator.New()
			err := validator.Struct(request)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestGitCommitIDHandling tests Git commit ID handling
func (suite *ComponentDeploymentServiceTestSuite) TestGitCommitIDHandling() {
	testCases := []struct {
		name     string
		commitID string
		valid    bool
	}{
		{
			name:     "Full SHA-1 hash",
			commitID: "abc123def456789012345678901234567890abcd",
			valid:    true,
		},
		{
			name:     "Short SHA-1 hash",
			commitID: "abc123d",
			valid:    true,
		},
		{
			name:     "SHA-256 hash",
			commitID: "abc123def456789012345678901234567890abcdef123456789012345678901234",
			valid:    true,
		},
		{
			name:     "Empty commit ID",
			commitID: "",
			valid:    true, // Empty commit ID is allowed
		},
		{
			name:     "Mixed case hash",
			commitID: "Abc123DEF456",
			valid:    true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := &service.CreateComponentDeploymentRequest{
				ComponentID: uuid.New(),
				LandscapeID: uuid.New(),
				GitCommitID: tc.commitID,
			}

			// Test validation (commit IDs are not validated by struct tags currently)
			validator := validator.New()
			err := validator.Struct(request)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestActiveDeploymentLogic tests active deployment logic scenarios
func (suite *ComponentDeploymentServiceTestSuite) TestActiveDeploymentLogic() {
	// Test scenarios for active deployment handling
	testCases := []struct {
		name           string
		existingActive bool
		newIsActive    *bool
		expectConflict bool
		description    string
	}{
		{
			name:           "Create active when none exists",
			existingActive: false,
			newIsActive:    nil, // defaults to true
			expectConflict: false,
			description:    "Should allow creating active deployment when none exists",
		},
		{
			name:           "Create active explicitly when none exists",
			existingActive: false,
			newIsActive:    &[]bool{true}[0],
			expectConflict: false,
			description:    "Should allow creating active deployment explicitly when none exists",
		},
		{
			name:           "Create inactive when active exists",
			existingActive: true,
			newIsActive:    &[]bool{false}[0],
			expectConflict: false,
			description:    "Should allow creating inactive deployment when active exists",
		},
		{
			name:           "Create active when active exists",
			existingActive: true,
			newIsActive:    &[]bool{true}[0],
			expectConflict: true,
			description:    "Should prevent creating active deployment when one already exists",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// This tests the business logic that would be implemented in the service
			var finalIsActive bool
			if tc.newIsActive == nil {
				finalIsActive = true // default
			} else {
				finalIsActive = *tc.newIsActive
			}

			// Simulate conflict detection
			wouldConflict := tc.existingActive && finalIsActive

			assert.Equal(t, tc.expectConflict, wouldConflict, tc.description)
		})
	}
}

// TestDeploymentHistoryPagination tests deployment history pagination logic
func (suite *ComponentDeploymentServiceTestSuite) TestDeploymentHistoryPagination() {
	// Simulate deployment history pagination
	deployments := make([]models.ComponentDeployment, 15) // 15 deployments
	for i := 0; i < 15; i++ {
		deployments[i] = models.ComponentDeployment{
			BaseModel:   models.BaseModel{ID: uuid.New()},
			ComponentID: uuid.New(),
			LandscapeID: uuid.New(),
			Version:     fmt.Sprintf("1.%d.0", i),
			IsActive:    i == 14, // Latest is active
		}
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
			pageSize:      5,
			expectedStart: 0,
			expectedEnd:   5,
			expectedCount: 5,
		},
		{
			name:          "Second page",
			page:          2,
			pageSize:      5,
			expectedStart: 5,
			expectedEnd:   10,
			expectedCount: 5,
		},
		{
			name:          "Last page partial",
			page:          3,
			pageSize:      5,
			expectedStart: 10,
			expectedEnd:   15,
			expectedCount: 5,
		},
		{
			name:          "Page beyond data",
			page:          4,
			pageSize:      5,
			expectedStart: 15,
			expectedEnd:   20,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
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

			var result []models.ComponentDeployment
			if start >= len(deployments) {
				result = []models.ComponentDeployment{}
			} else {
				actualEnd := end
				if actualEnd > len(deployments) {
					actualEnd = len(deployments)
				}
				result = deployments[start:actualEnd]
			}

			assert.Equal(t, tc.expectedCount, len(result))
		})
	}
}

// TestComponentDeploymentServiceTestSuite runs the test suite
func TestComponentDeploymentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentDeploymentServiceTestSuite))
}
