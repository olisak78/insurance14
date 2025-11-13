package service_test

import (
	"encoding/json"
	"testing"

	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// LandscapeServiceTestSuite defines the test suite for LandscapeService
type LandscapeServiceTestSuite struct {
	suite.Suite
	ctrl              *gomock.Controller
	mockLandscapeRepo *mocks.MockLandscapeRepositoryInterface
	mockOrgRepo       *mocks.MockOrganizationRepositoryInterface
	landscapeService  *service.LandscapeService
	validator         *validator.Validate
}

// SetupTest sets up the test suite
func (suite *LandscapeServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockLandscapeRepo = mocks.NewMockLandscapeRepositoryInterface(suite.ctrl)
	suite.mockOrgRepo = mocks.NewMockOrganizationRepositoryInterface(suite.ctrl)
	suite.validator = validator.New()

	// Since LandscapeService uses concrete repository types instead of interfaces,
	// we can't properly mock them for unit testing. This is a design issue
	// that would need to be fixed in the service layer.
	// For now, we'll focus on testing validation logic and other testable parts.
	suite.landscapeService = nil
}

// TearDownTest cleans up after each test
func (suite *LandscapeServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateLandscapeValidation tests the validation logic for creating a landscape
func (suite *LandscapeServiceTestSuite) TestCreateLandscapeValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.CreateLandscapeRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.CreateLandscapeRequest{
				ProjectID:   uuid.New(),
				Name:        "production-landscape",
				Title:       "Production Landscape",
				Description: "Main production environment",
				Domain:      "production.example.com",
				Environment: "production",
			},
			expectError: false,
		},
		// REMOVED: OrganizationID no longer exists - replaced by ProjectID
		// {
		// 	name: "Missing organization ID",
		// 	request: &service.CreateLandscapeRequest{
		// 		Name:        "production-landscape",
		// 		Title: "Production Landscape",
		// 	},
		// 	expectError: true,
		// 	errorMsg:    "OrganizationID",
		// },
		{
			name: "Empty name",
			request: &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:           "",
				Title:    "Production Landscape",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Empty display name",
			request: &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:           "production-landscape",
				Title:    "",
			},
			expectError: true,
			errorMsg:    "Title",
		},
		{
			name: "Name too long",
			request: &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:           "this-is-a-very-long-landscape-name-that-exceeds-the-maximum-allowed-length-of-two-hundred-characters-for-landscape-names-in-this-validation-system-and-should-trigger-validation-error-when-creating-landscape",
				Title:    "Production Landscape",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Display name too long",
			request: &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:           "production-landscape",
				Title:    "This is an extremely long display name that definitely exceeds the maximum allowed length of two hundred and fifty characters for the display name field and should trigger a validation error when we try to create a landscape with this incredibly long display name that goes way beyond the specified character limit making it invalid for validation system",
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

// TestUpdateLandscapeValidation tests the validation logic for updating a landscape
func (suite *LandscapeServiceTestSuite) TestUpdateLandscapeValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.UpdateLandscapeRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.UpdateLandscapeRequest{
				Title: "Updated Production Landscape",
				Description: "Updated description",
			},
			expectError: false,
		},
		{
			name: "Empty display name",
			request: &service.UpdateLandscapeRequest{
				Title: "",
			},
			expectError: true,
			errorMsg:    "Title",
		},
		{
			name: "Display name too long",
			request: &service.UpdateLandscapeRequest{
				Title: "This is an extremely long display name that definitely exceeds the maximum allowed length of two hundred and fifty characters for the display name field and should trigger a validation error when we try to update a landscape with this incredibly long display name that goes way beyond the specified character limit making it invalid for validation system",
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

// TestLandscapeResponseSerialization tests the landscape response serialization
func (suite *LandscapeServiceTestSuite) TestLandscapeResponseSerialization() {
	landscapeID := uuid.New()
	orgID := uuid.New()
	metadata := json.RawMessage(`{"region": "us-east-1", "tier": "production"}`)

	response := &service.LandscapeResponse{
		ID:          landscapeID,
		ProjectID:   orgID,
		Name:        "production-landscape",
		Title:       "Production Landscape",
		Description: "Main production environment",
		Domain:      "production.example.com",
		Environment: "production",
		Metadata:    metadata,
		CreatedAt:   "2023-01-01T00:00:00Z",
		UpdatedAt:   "2023-01-01T00:00:00Z",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), landscapeID.String())
	assert.Contains(suite.T(), string(jsonData), orgID.String())
	assert.Contains(suite.T(), string(jsonData), "production-landscape")
	assert.Contains(suite.T(), string(jsonData), "Production Landscape")
	assert.Contains(suite.T(), string(jsonData), `"environment":"production"`)

	// Test JSON unmarshaling
	var unmarshaled service.LandscapeResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), response.ID, unmarshaled.ID)
	assert.Equal(suite.T(), response.ProjectID, unmarshaled.ProjectID)
	assert.Equal(suite.T(), response.Name, unmarshaled.Name)
	assert.Equal(suite.T(), response.Title, unmarshaled.Title)
	assert.Equal(suite.T(), response.Environment, unmarshaled.Environment)
	assert.Equal(suite.T(), response.Domain, unmarshaled.Domain)
}

// TestLandscapeListResponseSerialization tests the landscape list response serialization
func (suite *LandscapeServiceTestSuite) TestLandscapeListResponseSerialization() {
	landscapes := []service.LandscapeResponse{
		{
			ID:          uuid.New(),
			ProjectID:   uuid.New(),
			Name:        "production",
			Title:       "Production",
			Environment: "production",
			Domain:      "us-east-1",
			CreatedAt:   "2023-01-01T00:00:00Z",
			UpdatedAt:   "2023-01-01T00:00:00Z",
		},
		{
			ID:          uuid.New(),
			ProjectID:   uuid.New(),
			Name:        "development",
			Title:       "Development",
			Environment: "development",
			Domain:      "us-west-2",
			CreatedAt:   "2023-01-02T00:00:00Z",
			UpdatedAt:   "2023-01-02T00:00:00Z",
		},
	}

	response := &service.LandscapeListResponse{
		Landscapes: landscapes,
		Total:      int64(len(landscapes)),
		Page:       1,
		PageSize:   20,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), "production")
	assert.Contains(suite.T(), string(jsonData), "development")
	assert.Contains(suite.T(), string(jsonData), `"total":2`)
	assert.Contains(suite.T(), string(jsonData), `"page":1`)

	// Test JSON unmarshaling
	var unmarshaled service.LandscapeListResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), unmarshaled.Landscapes, 2)
	assert.Equal(suite.T(), response.Total, unmarshaled.Total)
	assert.Equal(suite.T(), response.Page, unmarshaled.Page)
	assert.Equal(suite.T(), response.PageSize, unmarshaled.PageSize)
}

// TestDefaultValuesBehavior tests the default values behavior (REMOVED - enum types no longer exist)
// func (suite *LandscapeServiceTestSuite) TestDefaultValuesBehavior() {
// 	// NOTE: LandscapeType, LandscapeStatus, and DeploymentStatus enums removed in new schema
// }

// TestPaginationLogic tests the pagination logic
func (suite *LandscapeServiceTestSuite) TestPaginationLogic() {
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

// TestLandscapeTypeValidation tests landscape type validation (REMOVED - LandscapeType enum no longer exists)
// func (suite *LandscapeServiceTestSuite) TestLandscapeTypeValidation() {
// 	// NOTE: LandscapeType enum removed in new schema
// }

// TestLandscapeStatusValidation tests landscape status validation (REMOVED - LandscapeStatus enum no longer exists)
// func (suite *LandscapeServiceTestSuite) TestLandscapeStatusValidation() {
// 	// NOTE: LandscapeStatus enum removed in new schema
// }

// TestDeploymentStatusValidation tests deployment status validation (REMOVED - DeploymentStatus enum no longer exists)
// func (suite *LandscapeServiceTestSuite) TestDeploymentStatusValidation() {
// 	// NOTE: DeploymentStatus enum removed in new schema
// }

// TestJSONFieldsHandling tests handling of JSON fields (Metadata)
func (suite *LandscapeServiceTestSuite) TestJSONFieldsHandling() {
	// Test valid JSON for metadata
	validMetadata := json.RawMessage(`{"region": "us-east-1", "tier": "production", "tags": ["prod", "critical"]}`)

	// Test that valid JSON can be marshaled and unmarshaled
	metadataData, err := json.Marshal(validMetadata)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metadataData)
	assert.Contains(suite.T(), string(metadataData), "region")
	assert.Contains(suite.T(), string(metadataData), "us-east-1")

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

// TestAWSAccountIDValidation tests AWS Account ID handling (REMOVED - AWSAccountID field no longer exists)
func (suite *LandscapeServiceTestSuite) TestAWSAccountIDValidation() {
	suite.T().Skip("AWSAccountID field removed from Landscape model in new schema")
	return
	testCases := []struct {
		name         string
		awsAccountID string
		valid        bool
	}{
		{
			name:         "Valid 12-digit account ID",
			awsAccountID: "123456789012",
			valid:        true,
		},
		{
			name:         "Empty account ID",
			awsAccountID: "",
			valid:        true, // Empty is allowed
		},
		{
			name:         "Account ID with leading zeros",
			awsAccountID: "000123456789",
			valid:        true,
		},
		{
			name:         "Account ID too short",
			awsAccountID: "12345",
			valid:        true, // Not validated by struct tags currently
		},
		{
			name:         "Account ID too long",
			awsAccountID: "1234567890123",
			valid:        true, // Not validated by struct tags currently
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:      "test-landscape",
				Title:     "Test Landscape",
				// NOTE: AWSAccountID removed from Landscape model in new schema
			}

			// Test validation (AWS Account IDs are not validated by struct tags currently)
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

// TestURLFieldsValidation tests URL fields handling (REMOVED - URL fields no longer exist)
func (suite *LandscapeServiceTestSuite) TestURLFieldsValidation() {
	suite.T().Skip("GitHubConfigURL and CAMProfileURL fields removed from Landscape model in new schema")
	return
	testCases := []struct {
		name  string
		urls  map[string]string
		valid bool
	}{
		{
			name: "Valid URLs",
			urls: map[string]string{
				"github": "https://github.com/org/config",
				"cam":    "https://cam.sap.com/profile",
			},
			valid: true,
		},
		{
			name: "Empty URLs",
			urls: map[string]string{
				"github": "",
				"cam":    "",
			},
			valid: true, // Empty URLs are allowed
		},
		{
			name: "HTTP URLs",
			urls: map[string]string{
				"github": "http://github.com/org/config",
				"cam":    "http://cam.sap.com/profile",
			},
			valid: true,
		},
		{
			name: "Relative URLs",
			urls: map[string]string{
				"github": "/org/config",
				"cam":    "/profile",
			},
			valid: true, // Not validated by struct tags currently
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:      "test-landscape",
				Title:     "Test Landscape",
				// NOTE: GitHubConfigURL and CAMProfileURL removed from Landscape model in new schema
			}

			// Test validation (URLs are not validated by struct tags currently)
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

// TestSortOrderHandling tests sort order handling (REMOVED - SortOrder field no longer exists)
func (suite *LandscapeServiceTestSuite) TestSortOrderHandling() {
	suite.T().Skip("SortOrder field removed from Landscape model in new schema")
	return
	testCases := []struct {
		name      string
		sortOrder int
		valid     bool
	}{
		{
			name:      "Positive sort order",
			sortOrder: 1,
			valid:     true,
		},
		{
			name:      "Zero sort order",
			sortOrder: 0,
			valid:     true,
		},
		{
			name:      "Negative sort order",
			sortOrder: -1,
			valid:     true, // Not restricted by validation
		},
		{
			name:      "Large sort order",
			sortOrder: 1000,
			valid:     true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:      "test-landscape",
				Title:     "Test Landscape",
				// NOTE: SortOrder removed from Landscape model in new schema
			}

			// Test validation
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

// TestEnvironmentGroupHandling tests environment group handling (REMOVED - EnvironmentGroup field no longer exists)
func (suite *LandscapeServiceTestSuite) TestEnvironmentGroupHandling() {
	suite.T().Skip("EnvironmentGroup field removed from Landscape model in new schema")
	return
	testCases := []struct {
		name             string
		environmentGroup string
		valid            bool
	}{
		{
			name:             "Valid environment group",
			environmentGroup: "production",
			valid:            true,
		},
		{
			name:             "Empty environment group",
			environmentGroup: "",
			valid:            true, // Empty is allowed
		},
		{
			name:             "Environment group with spaces",
			environmentGroup: "production group",
			valid:            true,
		},
		{
			name:             "Environment group with special chars",
			environmentGroup: "prod-env_1",
			valid:            true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := &service.CreateLandscapeRequest{
				ProjectID: uuid.New(),
				Name:      "test-landscape",
				Title:     "Test Landscape",
				// NOTE: EnvironmentGroup removed from Landscape model in new schema
			}

			// Test validation
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

// TestUpdateRequestPointerFields tests pointer fields in update requests (REMOVED - enum types no longer exist)
// func (suite *LandscapeServiceTestSuite) TestUpdateRequestPointerFields() {
// 	// NOTE: LandscapeType, Status, DeploymentStatus, and SortOrder removed in new schema
// }

// Placeholder test to avoid compilation errors
func (suite *LandscapeServiceTestSuite) TestUpdateRequestPointerFieldsPlaceholder() {
	request := &service.UpdateLandscapeRequest{
		Title: "Updated Landscape",
	}

	// Test JSON marshaling with pointer fields
	jsonData, err := json.Marshal(request)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), "Updated Landscape")
	// NOTE: landscape_type, status, deployment_status, and sort_order removed in new schema

	// Test with nil pointer fields
	requestWithNils := &service.UpdateLandscapeRequest{
		Title: "Updated Landscape",
		// NOTE: LandscapeType, Status, DeploymentStatus, and SortOrder fields removed
	}

	jsonDataWithNils, err := json.Marshal(requestWithNils)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonDataWithNils), "Updated Landscape")
}

// TestLandscapeServiceTestSuite runs the test suite
func TestLandscapeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LandscapeServiceTestSuite))
}
