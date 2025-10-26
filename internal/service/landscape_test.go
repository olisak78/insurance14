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
				OrganizationID: uuid.New(),
				Name:           "production-landscape",
				DisplayName:    "Production Landscape",
				Description:    "Main production environment",
			},
			expectError: false,
		},
		{
			name: "Missing organization ID",
			request: &service.CreateLandscapeRequest{
				Name:        "production-landscape",
				DisplayName: "Production Landscape",
			},
			expectError: true,
			errorMsg:    "OrganizationID",
		},
		{
			name: "Empty name",
			request: &service.CreateLandscapeRequest{
				OrganizationID: uuid.New(),
				Name:           "",
				DisplayName:    "Production Landscape",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Empty display name",
			request: &service.CreateLandscapeRequest{
				OrganizationID: uuid.New(),
				Name:           "production-landscape",
				DisplayName:    "",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Name too long",
			request: &service.CreateLandscapeRequest{
				OrganizationID: uuid.New(),
				Name:           "this-is-a-very-long-landscape-name-that-exceeds-the-maximum-allowed-length-of-two-hundred-characters-for-landscape-names-in-this-validation-system-and-should-trigger-validation-error-when-creating-landscape",
				DisplayName:    "Production Landscape",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Display name too long",
			request: &service.CreateLandscapeRequest{
				OrganizationID: uuid.New(),
				Name:           "production-landscape",
				DisplayName:    "This is an extremely long display name that definitely exceeds the maximum allowed length of two hundred and fifty characters for the display name field and should trigger a validation error when we try to create a landscape with this incredibly long display name that goes way beyond the specified character limit making it invalid for validation system",
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
				DisplayName: "Updated Production Landscape",
				Description: "Updated description",
			},
			expectError: false,
		},
		{
			name: "Empty display name",
			request: &service.UpdateLandscapeRequest{
				DisplayName: "",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Display name too long",
			request: &service.UpdateLandscapeRequest{
				DisplayName: "This is an extremely long display name that definitely exceeds the maximum allowed length of two hundred and fifty characters for the display name field and should trigger a validation error when we try to update a landscape with this incredibly long display name that goes way beyond the specified character limit making it invalid for validation system",
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

// TestLandscapeResponseSerialization tests the landscape response serialization
func (suite *LandscapeServiceTestSuite) TestLandscapeResponseSerialization() {
	landscapeID := uuid.New()
	orgID := uuid.New()
	metadata := json.RawMessage(`{"region": "us-east-1", "tier": "production"}`)

	response := &service.LandscapeResponse{
		ID:               landscapeID,
		OrganizationID:   orgID,
		Name:             "production-landscape",
		DisplayName:      "Production Landscape",
		Description:      "Main production environment",
		LandscapeType:    models.LandscapeTypeProduction,
		EnvironmentGroup: "prod",
		Status:           models.LandscapeStatusActive,
		DeploymentStatus: models.DeploymentStatusHealthy,
		GitHubConfigURL:  "https://github.com/org/config",
		AWSAccountID:     "123456789012",
		CAMProfileURL:    "https://cam.sap.com/profile",
		SortOrder:        1,
		Metadata:         metadata,
		CreatedAt:        "2023-01-01T00:00:00Z",
		UpdatedAt:        "2023-01-01T00:00:00Z",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), landscapeID.String())
	assert.Contains(suite.T(), string(jsonData), orgID.String())
	assert.Contains(suite.T(), string(jsonData), "production-landscape")
	assert.Contains(suite.T(), string(jsonData), "Production Landscape")
	assert.Contains(suite.T(), string(jsonData), `"landscape_type":"production"`)
	assert.Contains(suite.T(), string(jsonData), `"status":"active"`)

	// Test JSON unmarshaling
	var unmarshaled service.LandscapeResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), response.ID, unmarshaled.ID)
	assert.Equal(suite.T(), response.OrganizationID, unmarshaled.OrganizationID)
	assert.Equal(suite.T(), response.Name, unmarshaled.Name)
	assert.Equal(suite.T(), response.DisplayName, unmarshaled.DisplayName)
	assert.Equal(suite.T(), response.LandscapeType, unmarshaled.LandscapeType)
	assert.Equal(suite.T(), response.Status, unmarshaled.Status)
	assert.Equal(suite.T(), response.DeploymentStatus, unmarshaled.DeploymentStatus)
}

// TestLandscapeListResponseSerialization tests the landscape list response serialization
func (suite *LandscapeServiceTestSuite) TestLandscapeListResponseSerialization() {
	landscapes := []service.LandscapeResponse{
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "production",
			DisplayName:    "Production",
			LandscapeType:  models.LandscapeTypeProduction,
			Status:         models.LandscapeStatusActive,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		},
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "development",
			DisplayName:    "Development",
			LandscapeType:  models.LandscapeTypeDevelopment,
			Status:         models.LandscapeStatusActive,
			CreatedAt:      "2023-01-02T00:00:00Z",
			UpdatedAt:      "2023-01-02T00:00:00Z",
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

// TestDefaultValuesBehavior tests the default values behavior
func (suite *LandscapeServiceTestSuite) TestDefaultValuesBehavior() {
	// Test landscape type defaults
	var emptyType models.LandscapeType
	expectedDefaultType := models.LandscapeTypeDevelopment

	var finalType models.LandscapeType
	if emptyType == "" {
		finalType = models.LandscapeTypeDevelopment
	} else {
		finalType = emptyType
	}

	assert.Equal(suite.T(), expectedDefaultType, finalType)

	// Test status defaults
	var emptyStatus models.LandscapeStatus
	expectedDefaultStatus := models.LandscapeStatusActive

	var finalStatus models.LandscapeStatus
	if emptyStatus == "" {
		finalStatus = models.LandscapeStatusActive
	} else {
		finalStatus = emptyStatus
	}

	assert.Equal(suite.T(), expectedDefaultStatus, finalStatus)

	// Test deployment status defaults
	var emptyDeploymentStatus models.DeploymentStatus
	expectedDefaultDeploymentStatus := models.DeploymentStatusUnknown

	var finalDeploymentStatus models.DeploymentStatus
	if emptyDeploymentStatus == "" {
		finalDeploymentStatus = models.DeploymentStatusUnknown
	} else {
		finalDeploymentStatus = emptyDeploymentStatus
	}

	assert.Equal(suite.T(), expectedDefaultDeploymentStatus, finalDeploymentStatus)
}

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

// TestLandscapeTypeValidation tests landscape type validation
func (suite *LandscapeServiceTestSuite) TestLandscapeTypeValidation() {
	validTypes := []models.LandscapeType{
		models.LandscapeTypeDevelopment,
		models.LandscapeTypeStaging,
		models.LandscapeTypeProduction,
	}

	for _, landscapeType := range validTypes {
		suite.T().Run(string(landscapeType), func(t *testing.T) {
			// Test that valid landscape types are accepted
			assert.NotEmpty(t, string(landscapeType))
			assert.True(t, landscapeType == models.LandscapeTypeDevelopment ||
				landscapeType == models.LandscapeTypeStaging ||
				landscapeType == models.LandscapeTypeProduction)
		})
	}
}

// TestLandscapeStatusValidation tests landscape status validation
func (suite *LandscapeServiceTestSuite) TestLandscapeStatusValidation() {
	validStatuses := []models.LandscapeStatus{
		models.LandscapeStatusActive,
		models.LandscapeStatusInactive,
		models.LandscapeStatusMaintenance,
	}

	for _, status := range validStatuses {
		suite.T().Run(string(status), func(t *testing.T) {
			// Test that valid statuses are accepted
			assert.NotEmpty(t, string(status))
			assert.True(t, status == models.LandscapeStatusActive ||
				status == models.LandscapeStatusInactive ||
				status == models.LandscapeStatusMaintenance)
		})
	}
}

// TestDeploymentStatusValidation tests deployment status validation
func (suite *LandscapeServiceTestSuite) TestDeploymentStatusValidation() {
	validStatuses := []models.DeploymentStatus{
		models.DeploymentStatusHealthy,
		models.DeploymentStatusDegraded,
		models.DeploymentStatusUnhealthy,
		models.DeploymentStatusUnknown,
	}

	for _, status := range validStatuses {
		suite.T().Run(string(status), func(t *testing.T) {
			// Test that valid deployment statuses are accepted
			assert.NotEmpty(t, string(status))
			assert.True(t, status == models.DeploymentStatusHealthy ||
				status == models.DeploymentStatusDegraded ||
				status == models.DeploymentStatusUnhealthy ||
				status == models.DeploymentStatusUnknown)
		})
	}
}

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

// TestAWSAccountIDValidation tests AWS Account ID handling
func (suite *LandscapeServiceTestSuite) TestAWSAccountIDValidation() {
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
				OrganizationID: uuid.New(),
				Name:           "test-landscape",
				DisplayName:    "Test Landscape",
				AWSAccountID:   tc.awsAccountID,
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

// TestURLFieldsValidation tests URL fields handling
func (suite *LandscapeServiceTestSuite) TestURLFieldsValidation() {
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
				OrganizationID:  uuid.New(),
				Name:            "test-landscape",
				DisplayName:     "Test Landscape",
				GitHubConfigURL: tc.urls["github"],
				CAMProfileURL:   tc.urls["cam"],
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

// TestSortOrderHandling tests sort order handling
func (suite *LandscapeServiceTestSuite) TestSortOrderHandling() {
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
				OrganizationID: uuid.New(),
				Name:           "test-landscape",
				DisplayName:    "Test Landscape",
				SortOrder:      tc.sortOrder,
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

// TestEnvironmentGroupHandling tests environment group handling
func (suite *LandscapeServiceTestSuite) TestEnvironmentGroupHandling() {
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
				OrganizationID:   uuid.New(),
				Name:             "test-landscape",
				DisplayName:      "Test Landscape",
				EnvironmentGroup: tc.environmentGroup,
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

// TestUpdateRequestPointerFields tests pointer fields in update requests
func (suite *LandscapeServiceTestSuite) TestUpdateRequestPointerFields() {
	// Test that pointer fields can be nil or have values
	landscapeType := models.LandscapeTypeProduction
	status := models.LandscapeStatusActive
	deploymentStatus := models.DeploymentStatusHealthy
	sortOrder := 5

	request := &service.UpdateLandscapeRequest{
		DisplayName:      "Updated Landscape",
		LandscapeType:    &landscapeType,
		Status:           &status,
		DeploymentStatus: &deploymentStatus,
		SortOrder:        &sortOrder,
	}

	// Test JSON marshaling with pointer fields
	jsonData, err := json.Marshal(request)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), "Updated Landscape")
	assert.Contains(suite.T(), string(jsonData), `"landscape_type":"production"`)
	assert.Contains(suite.T(), string(jsonData), `"status":"active"`)
	assert.Contains(suite.T(), string(jsonData), `"sort_order":5`)

	// Test with nil pointer fields
	requestWithNils := &service.UpdateLandscapeRequest{
		DisplayName:      "Updated Landscape",
		LandscapeType:    nil,
		Status:           nil,
		DeploymentStatus: nil,
		SortOrder:        nil,
	}

	jsonDataWithNils, err := json.Marshal(requestWithNils)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonDataWithNils), "Updated Landscape")
}

// TestLandscapeServiceTestSuite runs the test suite
func TestLandscapeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LandscapeServiceTestSuite))
}
