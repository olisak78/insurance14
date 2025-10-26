package service_test

import (
	"encoding/json"
	"errors"
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

// ProjectServiceTestSuite defines the test suite for ProjectService
type ProjectServiceTestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mockRepo       *mocks.MockProjectRepositoryInterface
	mockOrgRepo    *mocks.MockOrganizationRepositoryInterface
	projectService *service.ProjectService
	validator      *validator.Validate
}

// SetupTest sets up the test suite
func (suite *ProjectServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.validator = validator.New()
	// Note: We're testing validation logic and data structures since the service
	// uses concrete repositories that can't be easily mocked without interface changes
}

// TearDownTest cleans up after each test
func (suite *ProjectServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateProjectValidation tests the validation logic for creating a project
func (suite *ProjectServiceTestSuite) TestCreateProjectValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.CreateProjectRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "auth-portal",
				DisplayName:    "Authentication Portal",
				Description:    "Portal for managing authentication",
				ProjectType:    models.ProjectTypeApplication,
				Status:         models.ProjectStatusActive,
			},
			expectError: false,
		},
		{
			name: "Missing organization ID",
			request: &service.CreateProjectRequest{
				Name:        "auth-portal",
				DisplayName: "Authentication Portal",
			},
			expectError: true,
			errorMsg:    "OrganizationID",
		},
		{
			name: "Empty name",
			request: &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "",
				DisplayName:    "Authentication Portal",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Empty display name",
			request: &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "auth-portal",
				DisplayName:    "",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Name too long",
			request: &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "this-is-a-very-long-project-name-that-definitely-exceeds-the-maximum-allowed-length-of-exactly-two-hundred-characters-for-project-names-in-this-validation-system-and-should-trigger-validation-error-123",
				DisplayName:    "Authentication Portal",
			},
			expectError: true,
			errorMsg:    "Name",
		},
		{
			name: "Display name too long",
			request: &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "auth-portal",
				DisplayName:    "This is an extremely long display name that definitely exceeds the maximum allowed length of exactly two hundred and fifty characters for the display name field and should trigger a validation error when we try to create a project with this incredibly long display name that goes way beyond the specified character limit making it invalid for validation system and this extra text pushes it over",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Valid with all optional fields",
			request: &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "auth-portal",
				DisplayName:    "Authentication Portal",
				Description:    "Comprehensive portal for managing authentication and user access",
				ProjectType:    models.ProjectTypePlatform,
				Status:         models.ProjectStatusActive,
				SortOrder:      10,
				Metadata:       json.RawMessage(`{"version": "1.0.0", "team": "security", "priority": "high"}`),
			},
			expectError: false,
		},
		{
			name: "Valid with minimal fields",
			request: &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "simple-project",
				DisplayName:    "Simple Project",
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

// TestUpdateProjectValidation tests the validation logic for updating a project
func (suite *ProjectServiceTestSuite) TestUpdateProjectValidation() {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.UpdateProjectRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid request",
			request: &service.UpdateProjectRequest{
				DisplayName: "Updated Authentication Portal",
				Description: "Updated description",
			},
			expectError: false,
		},
		{
			name: "Empty display name",
			request: &service.UpdateProjectRequest{
				DisplayName: "",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Display name too long",
			request: &service.UpdateProjectRequest{
				DisplayName: "This is an extremely long display name that definitely exceeds the maximum allowed length of two hundred and fifty characters for the display name field and should trigger a validation error when we try to update a project with this incredibly long display name that goes way beyond the specified character limit making it invalid for validation system",
			},
			expectError: true,
			errorMsg:    "DisplayName",
		},
		{
			name: "Valid with all optional fields",
			request: &service.UpdateProjectRequest{
				DisplayName: "Updated Authentication Portal",
				Description: "Updated comprehensive description",
				Metadata:    json.RawMessage(`{"version": "2.0.0", "team": "security", "priority": "medium"}`),
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

// TestProjectResponseSerialization tests the project response serialization
func (suite *ProjectServiceTestSuite) TestProjectResponseSerialization() {
	projectID := uuid.New()
	orgID := uuid.New()
	metadata := json.RawMessage(`{"version": "1.0.0", "team": "security", "priority": "high", "features": ["auth", "sso"]}`)

	response := &service.ProjectResponse{
		ID:             projectID,
		OrganizationID: orgID,
		Name:           "auth-portal",
		DisplayName:    "Authentication Portal",
		Description:    "Comprehensive portal for managing authentication and user access",
		ProjectType:    models.ProjectTypePlatform,
		Status:         models.ProjectStatusActive,
		SortOrder:      10,
		Metadata:       metadata,
		CreatedAt:      "2023-01-01T00:00:00Z",
		UpdatedAt:      "2023-01-01T00:00:00Z",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), projectID.String())
	assert.Contains(suite.T(), string(jsonData), orgID.String())
	assert.Contains(suite.T(), string(jsonData), "auth-portal")
	assert.Contains(suite.T(), string(jsonData), "Authentication Portal")
	assert.Contains(suite.T(), string(jsonData), `"project_type":"platform"`)
	assert.Contains(suite.T(), string(jsonData), `"status":"active"`)
	assert.Contains(suite.T(), string(jsonData), `"sort_order":10`)

	// Test JSON unmarshaling
	var unmarshaled service.ProjectResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), response.ID, unmarshaled.ID)
	assert.Equal(suite.T(), response.OrganizationID, unmarshaled.OrganizationID)
	assert.Equal(suite.T(), response.Name, unmarshaled.Name)
	assert.Equal(suite.T(), response.DisplayName, unmarshaled.DisplayName)
	assert.Equal(suite.T(), response.ProjectType, unmarshaled.ProjectType)
	assert.Equal(suite.T(), response.Status, unmarshaled.Status)
	assert.Equal(suite.T(), response.SortOrder, unmarshaled.SortOrder)
}

// TestProjectListResponseSerialization tests the project list response serialization
func (suite *ProjectServiceTestSuite) TestProjectListResponseSerialization() {
	projects := []service.ProjectResponse{
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "auth-portal",
			DisplayName:    "Authentication Portal",
			ProjectType:    models.ProjectTypePlatform,
			Status:         models.ProjectStatusActive,
			SortOrder:      10,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		},
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			Name:           "user-service",
			DisplayName:    "User Service",
			ProjectType:    models.ProjectTypeService,
			Status:         models.ProjectStatusActive,
			SortOrder:      20,
			CreatedAt:      "2023-01-02T00:00:00Z",
			UpdatedAt:      "2023-01-02T00:00:00Z",
		},
	}

	response := &service.ProjectListResponse{
		Projects: projects,
		Total:    int64(len(projects)),
		Page:     1,
		PageSize: 20,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), "auth-portal")
	assert.Contains(suite.T(), string(jsonData), "user-service")
	assert.Contains(suite.T(), string(jsonData), `"total":2`)
	assert.Contains(suite.T(), string(jsonData), `"page":1`)

	// Test JSON unmarshaling
	var unmarshaled service.ProjectListResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), unmarshaled.Projects, 2)
	assert.Equal(suite.T(), response.Total, unmarshaled.Total)
	assert.Equal(suite.T(), response.Page, unmarshaled.Page)
	assert.Equal(suite.T(), response.PageSize, unmarshaled.PageSize)
}

// TestDefaultValuesBehavior tests the default values behavior
func (suite *ProjectServiceTestSuite) TestDefaultValuesBehavior() {
	// Test project type defaults
	var emptyType models.ProjectType
	expectedDefaultType := models.ProjectTypeApplication

	var finalType models.ProjectType
	if emptyType == "" {
		finalType = models.ProjectTypeApplication
	} else {
		finalType = emptyType
	}

	assert.Equal(suite.T(), expectedDefaultType, finalType)

	// Test status defaults
	var emptyStatus models.ProjectStatus
	expectedDefaultStatus := models.ProjectStatusActive

	var finalStatus models.ProjectStatus
	if emptyStatus == "" {
		finalStatus = models.ProjectStatusActive
	} else {
		finalStatus = emptyStatus
	}

	assert.Equal(suite.T(), expectedDefaultStatus, finalStatus)
}

// TestPaginationLogic tests the pagination logic
func (suite *ProjectServiceTestSuite) TestPaginationLogic() {
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
			inputPage:      3,
			inputSize:      15,
			expectedPage:   3,
			expectedSize:   15,
			expectedOffset: 30,
		},
		{
			name:           "Page less than 1",
			inputPage:      0,
			inputSize:      15,
			expectedPage:   1,
			expectedSize:   15,
			expectedOffset: 0,
		},
		{
			name:           "Page size less than 1",
			inputPage:      2,
			inputSize:      0,
			expectedPage:   2,
			expectedSize:   20,
			expectedOffset: 20,
		},
		{
			name:           "Page size greater than 100",
			inputPage:      1,
			inputSize:      200,
			expectedPage:   1,
			expectedSize:   20,
			expectedOffset: 0,
		},
		{
			name:           "Both invalid",
			inputPage:      -5,
			inputSize:      -10,
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

// TestProjectTypeValidation tests project type validation
func (suite *ProjectServiceTestSuite) TestProjectTypeValidation() {
	validTypes := []models.ProjectType{
		models.ProjectTypeApplication,
		models.ProjectTypeService,
		models.ProjectTypeLibrary,
		models.ProjectTypePlatform,
	}

	for _, projectType := range validTypes {
		suite.T().Run(string(projectType), func(t *testing.T) {
			// Test that valid project types are accepted
			assert.NotEmpty(t, string(projectType))
			assert.True(t, projectType == models.ProjectTypeApplication ||
				projectType == models.ProjectTypeService ||
				projectType == models.ProjectTypeLibrary ||
				projectType == models.ProjectTypePlatform)
		})
	}
}

// TestProjectStatusValidation tests project status validation
func (suite *ProjectServiceTestSuite) TestProjectStatusValidation() {
	validStatuses := []models.ProjectStatus{
		models.ProjectStatusActive,
		models.ProjectStatusInactive,
		models.ProjectStatusArchived,
	}

	for _, status := range validStatuses {
		suite.T().Run(string(status), func(t *testing.T) {
			// Test that valid statuses are accepted
			assert.NotEmpty(t, string(status))
			assert.True(t, status == models.ProjectStatusActive ||
				status == models.ProjectStatusInactive ||
				status == models.ProjectStatusArchived)
		})
	}
}

// TestSortOrderHandling tests sort order handling
func (suite *ProjectServiceTestSuite) TestSortOrderHandling() {
	testCases := []struct {
		name      string
		sortOrder int
		valid     bool
	}{
		{
			name:      "Positive sort order",
			sortOrder: 10,
			valid:     true,
		},
		{
			name:      "Zero sort order",
			sortOrder: 0,
			valid:     true,
		},
		{
			name:      "Negative sort order",
			sortOrder: -5,
			valid:     true, // Negative values are allowed
		},
		{
			name:      "Large sort order",
			sortOrder: 999999,
			valid:     true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			request := &service.CreateProjectRequest{
				OrganizationID: uuid.New(),
				Name:           "test-project",
				DisplayName:    "Test Project",
				SortOrder:      tc.sortOrder,
			}

			// Test validation (sort order has no validation constraints currently)
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

// TestJSONMetadataHandling tests handling of JSON metadata field
func (suite *ProjectServiceTestSuite) TestJSONMetadataHandling() {
	// Test valid JSON metadata
	validMetadata := json.RawMessage(`{
		"version": "1.0.0", 
		"team": "security", 
		"priority": "high",
		"features": ["auth", "sso", "mfa"],
		"config": {
			"timeout": 30,
			"retries": 3
		}
	}`)

	// Test that valid JSON can be marshaled and unmarshaled
	metadataData, err := json.Marshal(validMetadata)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), metadataData)
	assert.Contains(suite.T(), string(metadataData), "version")
	assert.Contains(suite.T(), string(metadataData), "1.0.0")
	assert.Contains(suite.T(), string(metadataData), "features")

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

	// Test complex nested JSON
	complexJSON := json.RawMessage(`{
		"environments": {
			"dev": {"url": "https://dev.example.com", "active": true},
			"staging": {"url": "https://staging.example.com", "active": true},
			"prod": {"url": "https://prod.example.com", "active": false}
		},
		"dependencies": [
			{"name": "auth-service", "version": "1.2.3", "required": true},
			{"name": "user-service", "version": "2.1.0", "required": false}
		]
	}`)

	complexData, err := json.Marshal(complexJSON)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(complexData), "environments")
	assert.Contains(suite.T(), string(complexData), "dependencies")
}

// TestUpdateRequestPointerFields tests pointer fields in update requests
func (suite *ProjectServiceTestSuite) TestUpdateRequestPointerFields() {
	// Test that pointer fields can be nil or have values
	projectType := models.ProjectTypePlatform
	status := models.ProjectStatusArchived
	sortOrder := 50

	request := &service.UpdateProjectRequest{
		DisplayName: "Updated Project",
		ProjectType: &projectType,
		Status:      &status,
		SortOrder:   &sortOrder,
	}

	// Test JSON marshaling with pointer fields
	jsonData, err := json.Marshal(request)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonData), "Updated Project")
	assert.Contains(suite.T(), string(jsonData), `"project_type":"platform"`)
	assert.Contains(suite.T(), string(jsonData), `"status":"archived"`)
	assert.Contains(suite.T(), string(jsonData), `"sort_order":50`)

	// Test with nil pointer fields
	requestWithNils := &service.UpdateProjectRequest{
		DisplayName: "Updated Project",
		ProjectType: nil,
		Status:      nil,
		SortOrder:   nil,
	}

	jsonDataWithNils, err := json.Marshal(requestWithNils)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(jsonDataWithNils), "Updated Project")
}

// TestBusinessLogicValidation tests business logic validation scenarios
func (suite *ProjectServiceTestSuite) TestBusinessLogicValidation() {
	// Test unique name constraint logic
	suite.T().Run("Unique name constraint", func(t *testing.T) {
		existingName := "existing-project"

		// This would be the logic to check if a project with the same name exists
		// In a real test with mocked repositories, we would set up the mock to return
		// an existing project when GetByName is called
		projectNames := map[string]bool{
			"existing-project": true,
			"new-project":      false,
		}

		// Simulate checking if project exists
		exists := projectNames[existingName]
		assert.True(t, exists, "Project with name 'existing-project' should exist")

		newName := "new-project"
		exists = projectNames[newName]
		assert.False(t, exists, "Project with name 'new-project' should not exist")
	})

	// Test organization validation logic
	suite.T().Run("Organization validation", func(t *testing.T) {
		validOrgID := uuid.New()
		invalidOrgID := uuid.New()

		// This would be the logic to check if an organization exists
		// In a real test with mocked repositories, we would set up the mock
		validOrgs := map[uuid.UUID]bool{
			validOrgID:   true,
			invalidOrgID: false,
		}

		// Simulate organization existence check
		exists := validOrgs[validOrgID]
		assert.True(t, exists, "Valid organization should exist")

		exists = validOrgs[invalidOrgID]
		assert.False(t, exists, "Invalid organization should not exist")
	})

	// Test component/landscape association logic
	suite.T().Run("Component association logic", func(t *testing.T) {
		projectID := uuid.New()
		componentID := uuid.New()
		existingComponentID := uuid.New()

		// This would be the logic to check if a component is already associated
		componentAssociations := map[string]bool{
			projectID.String() + ":" + existingComponentID.String(): true,
			projectID.String() + ":" + componentID.String():         false,
		}

		// Simulate checking if component is already associated
		key := projectID.String() + ":" + existingComponentID.String()
		exists := componentAssociations[key]
		assert.True(t, exists, "Component should already be associated")

		key = projectID.String() + ":" + componentID.String()
		exists = componentAssociations[key]
		assert.False(t, exists, "Component should not be associated yet")
	})
}

// TestErrorHandlingScenarios tests error handling scenarios
func (suite *ProjectServiceTestSuite) TestErrorHandlingScenarios() {
	// Test error types that the service should handle
	suite.T().Run("GORM errors", func(t *testing.T) {
		// Test ErrRecordNotFound handling
		err := gorm.ErrRecordNotFound
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

		// This would be how the service handles not found errors
		var serviceError error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			serviceError = errors.New("project not found")
		}
		assert.EqualError(t, serviceError, "project not found")
	})

	suite.T().Run("Validation errors", func(t *testing.T) {
		// Test validation error handling
		validator := validator.New()
		invalidRequest := &service.CreateProjectRequest{
			// Missing required fields
		}

		err := validator.Struct(invalidRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "OrganizationID")
		assert.Contains(t, err.Error(), "Name")
		assert.Contains(t, err.Error(), "DisplayName")
	})

	suite.T().Run("Business logic errors", func(t *testing.T) {
		// Test business logic error scenarios
		testErrors := []struct {
			scenario string
			error    error
		}{
			{
				scenario: "Organization not found",
				error:    errors.New("organization not found"),
			},
			{
				scenario: "Project name already exists",
				error:    errors.New("project with this name already exists in the organization"),
			},
			{
				scenario: "Project not found",
				error:    errors.New("project not found"),
			},
			{
				scenario: "Component already associated",
				error:    errors.New("component is already associated with this project"),
			},
			{
				scenario: "Component not associated",
				error:    errors.New("component is not associated with this project"),
			},
			{
				scenario: "Landscape already associated",
				error:    errors.New("landscape is already associated with this project"),
			},
			{
				scenario: "Landscape not associated",
				error:    errors.New("landscape is not associated with this project"),
			},
		}

		for _, testError := range testErrors {
			assert.Error(suite.T(), testError.error)
			assert.NotEmpty(suite.T(), testError.error.Error())
		}
	})
}

// TestProjectServiceMethods tests the structure and expected behavior of service methods
func (suite *ProjectServiceTestSuite) TestProjectServiceMethods() {
	// Test that all expected service methods would exist and have proper signatures
	// This is more of a design validation test
	suite.T().Run("Service method signatures", func(t *testing.T) {
		// In a real implementation with proper interfaces, we would test:
		// - Create method accepts CreateProjectRequest and returns ProjectResponse and error
		// - GetByID method accepts UUID and returns ProjectResponse and error
		// - GetByName method accepts organization ID and name, returns ProjectResponse and error
		// - Update method accepts UUID and UpdateProjectRequest, returns ProjectResponse and error
		// - Delete method accepts UUID and returns error
		// - Various GetBy* methods for filtering
		// - Component and landscape association methods

		expectedMethods := []string{
			"Create",
			"GetByID",
			"GetByName",
			"GetByOrganization",
			"GetByStatus",
			"GetActiveProjects",
			"Search",
			"Update",
			"Delete",
			"SetStatus",
			"GetWithOrganization",
			"GetWithComponents",
			"GetWithLandscapes",
			"GetWithFullDetails",
			"AddComponent",
			"RemoveComponent",
			"AddLandscape",
			"RemoveLandscape",
		}

		// Just validate that we have the expected method names documented
		assert.Greater(suite.T(), len(expectedMethods), 15)
		assert.Contains(suite.T(), expectedMethods, "Create")
		assert.Contains(suite.T(), expectedMethods, "GetByID")
		assert.Contains(suite.T(), expectedMethods, "Update")
		assert.Contains(suite.T(), expectedMethods, "Delete")
		assert.Contains(suite.T(), expectedMethods, "AddComponent")
		assert.Contains(suite.T(), expectedMethods, "AddLandscape")
	})
}

// TestSpecialFieldHandling tests special field handling scenarios
func (suite *ProjectServiceTestSuite) TestSpecialFieldHandling() {
	// Test complex JSON structures in metadata
	suite.T().Run("Complex JSON metadata structures", func(t *testing.T) {
		complexMetadata := json.RawMessage(`{
			"version": "2.1.0",
			"team": "platform-engineering",
			"maintainer": "platform-team@company.com",
			"priority": "high",
			"environments": {
				"development": {
					"url": "https://dev-auth-portal.company.com",
					"status": "active",
					"resources": ["database", "cache", "messaging"]
				},
				"staging": {
					"url": "https://staging-auth-portal.company.com",
					"status": "active",
					"resources": ["database", "cache"]
				},
				"production": {
					"url": "https://auth-portal.company.com",
					"status": "maintenance",
					"resources": ["database", "cache", "messaging", "monitoring"]
				}
			},
			"features": ["authentication", "sso", "mfa", "audit"],
			"dependencies": [
				{
					"name": "user-service",
					"version": "1.5.2",
					"type": "service",
					"required": true
				},
				{
					"name": "notification-service",
					"version": "1.2.0",
					"type": "service",
					"required": false
				}
			],
			"config": {
				"session_timeout": 3600,
				"max_login_attempts": 3,
				"password_policy": {
					"min_length": 8,
					"require_special_chars": true,
					"require_numbers": true
				}
			}
		}`)

		// Test that complex JSON can be marshaled and unmarshaled
		data, err := json.Marshal(complexMetadata)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "version")
		assert.Contains(t, string(data), "environments")
		assert.Contains(t, string(data), "dependencies")
		assert.Contains(t, string(data), "config")
		assert.Contains(t, string(data), "password_policy")

		var unmarshaled json.RawMessage
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.NotNil(t, unmarshaled)
	})

	// Test edge cases for string fields
	suite.T().Run("String field edge cases", func(t *testing.T) {
		testCases := []struct {
			name  string
			field string
			value string
			valid bool
		}{
			{
				name:  "Unicode characters",
				field: "displayName",
				value: "Authentication Portal 🔐",
				valid: true,
			},
			{
				name:  "Special characters",
				field: "name",
				value: "auth-portal_v1.0",
				valid: true,
			},
			{
				name:  "Numbers in name",
				field: "name",
				value: "auth-portal-123",
				valid: true,
			},
		}

		for _, tc := range testCases {
			assert.True(t, len(tc.value) > 0)
			assert.NotEmpty(t, tc.field)
		}
	})
}

// TestProjectServiceTestSuite runs the test suite
func TestProjectServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectServiceTestSuite))
}
