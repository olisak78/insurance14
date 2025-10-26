package handlers

import (
	"fmt"
	"net/http"
	"testing"

	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// OrganizationHandlerTestSuite defines the test suite for OrganizationHandler
type OrganizationHandlerTestSuite struct {
	suite.Suite
	ctrl                    *gomock.Controller
	mockOrganizationService *mocks.MockOrganizationServiceInterface
	handler                 *OrganizationHandler
	httpSuite               *testutils.HTTPTestSuite
}

// SetupTest sets up the test suite
func (suite *OrganizationHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockOrganizationService = mocks.NewMockOrganizationServiceInterface(suite.ctrl)

	// Create handler with mock service
	suite.handler = NewOrganizationHandler(suite.mockOrganizationService)

	// Setup HTTP test suite
	suite.httpSuite = testutils.SetupHTTPTest()

	// Register routes
	v1 := suite.httpSuite.Router.Group("/api/v1")
	orgs := v1.Group("/organizations")
	{
		orgs.POST("/", suite.handler.CreateOrganization)
		orgs.GET("/:id", suite.handler.GetOrganization)
		orgs.GET("/by-name/:name", suite.handler.GetOrganizationByName)
		orgs.PUT("/:id", suite.handler.UpdateOrganization)
		orgs.DELETE("/:id", suite.handler.DeleteOrganization)
		orgs.GET("/", suite.handler.ListOrganizations)
	}
}

// TearDownTest cleans up after each test
func (suite *OrganizationHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateOrganization tests creating an organization
func (suite *OrganizationHandlerTestSuite) TestCreateOrganization() {
	orgID := uuid.New()
	requestBody := map[string]interface{}{
		"name":         "test-org",
		"display_name": "Test Organization",
		"domain":       "test.com",
		"description":  "Test description",
	}

	expectedResponse := &service.OrganizationResponse{
		ID:          orgID,
		Name:        "test-org",
		DisplayName: "Test Organization",
		Domain:      "test.com",
		Description: "Test description",
		CreatedAt:   "2023-01-01T00:00:00Z",
		UpdatedAt:   "2023-01-01T00:00:00Z",
	}

	suite.mockOrganizationService.EXPECT().
		Create(gomock.Any()).
		Return(expectedResponse, nil).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("POST", "/api/v1/organizations/", requestBody)

	assert.Equal(suite.T(), http.StatusCreated, recorder.Code)

	var response service.OrganizationResponse
	testutils.ParseJSONResponse(suite.T(), recorder, &response)
	assert.Equal(suite.T(), expectedResponse.Name, response.Name)
	assert.Equal(suite.T(), expectedResponse.DisplayName, response.DisplayName)
}

// TestCreateOrganizationBadRequest tests creating an organization with invalid data
func (suite *OrganizationHandlerTestSuite) TestCreateOrganizationBadRequest() {
	requestBody := map[string]interface{}{
		"name": "", // Empty name should fail validation
	}

	// Since validation happens at JSON binding level, we expect service.Create to be called
	// with invalid data, which should return a validation error
	suite.mockOrganizationService.EXPECT().
		Create(gomock.Any()).
		Return(nil, fmt.Errorf("validation error: name is required")).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("POST", "/api/v1/organizations/", requestBody)

	assert.Equal(suite.T(), http.StatusInternalServerError, recorder.Code)
	testutils.AssertErrorResponse(suite.T(), recorder, http.StatusInternalServerError, "Failed to create organization")
}

// TestCreateOrganizationServiceError tests creating an organization with service error
func (suite *OrganizationHandlerTestSuite) TestCreateOrganizationServiceError() {
	requestBody := map[string]interface{}{
		"name":         "test-org",
		"display_name": "Test Organization",
		"domain":       "test.com",
		"description":  "Test description",
	}

	suite.mockOrganizationService.EXPECT().
		Create(gomock.Any()).
		Return(nil, fmt.Errorf("service error")).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("POST", "/api/v1/organizations/", requestBody)

	assert.Equal(suite.T(), http.StatusInternalServerError, recorder.Code)
	testutils.AssertErrorResponse(suite.T(), recorder, http.StatusInternalServerError, "Failed to create organization")
}

// TestGetOrganization tests getting an organization by ID
func (suite *OrganizationHandlerTestSuite) TestGetOrganization() {
	orgID := uuid.New()
	expectedResponse := &service.OrganizationResponse{
		ID:          orgID,
		Name:        "test-org",
		DisplayName: "Test Organization",
		Domain:      "test.com",
		Description: "Test description",
		CreatedAt:   "2023-01-01T00:00:00Z",
		UpdatedAt:   "2023-01-01T00:00:00Z",
	}

	suite.mockOrganizationService.EXPECT().
		GetByID(orgID).
		Return(expectedResponse, nil).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/organizations/%s", orgID), nil)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response service.OrganizationResponse
	testutils.ParseJSONResponse(suite.T(), recorder, &response)
	assert.Equal(suite.T(), expectedResponse.ID, response.ID)
	assert.Equal(suite.T(), expectedResponse.Name, response.Name)
}

// TestGetOrganizationInvalidID tests getting an organization with invalid ID
func (suite *OrganizationHandlerTestSuite) TestGetOrganizationInvalidID() {
	recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/organizations/invalid-uuid", nil)

	assert.Equal(suite.T(), http.StatusBadRequest, recorder.Code)
	testutils.AssertErrorResponse(suite.T(), recorder, http.StatusBadRequest, "invalid UUID")
}

// TestGetOrganizationNotFound tests getting a non-existent organization
func (suite *OrganizationHandlerTestSuite) TestGetOrganizationNotFound() {
	orgID := uuid.New()

	suite.mockOrganizationService.EXPECT().
		GetByID(orgID).
		Return(nil, apperrors.ErrOrganizationNotFound).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/organizations/%s", orgID), nil)

	assert.Equal(suite.T(), http.StatusNotFound, recorder.Code)
	testutils.AssertErrorResponse(suite.T(), recorder, http.StatusNotFound, "organization not found")
}

// TestGetOrganizationByName tests getting an organization by name
func (suite *OrganizationHandlerTestSuite) TestGetOrganizationByName() {
	orgName := "test-org"
	expectedResponse := &service.OrganizationResponse{
		ID:          uuid.New(),
		Name:        "test-org",
		DisplayName: "Test Organization",
		Domain:      "test.com",
		Description: "Test description",
		CreatedAt:   "2023-01-01T00:00:00Z",
		UpdatedAt:   "2023-01-01T00:00:00Z",
	}

	suite.mockOrganizationService.EXPECT().
		GetByName(orgName).
		Return(expectedResponse, nil).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/organizations/by-name/%s", orgName), nil)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response service.OrganizationResponse
	testutils.ParseJSONResponse(suite.T(), recorder, &response)
	assert.Equal(suite.T(), expectedResponse.ID, response.ID)
	assert.Equal(suite.T(), expectedResponse.Name, response.Name)
}

// TestGetOrganizationByNameEmptyName tests getting an organization with empty name
func (suite *OrganizationHandlerTestSuite) TestGetOrganizationByNameEmptyName() {
	// Mock the service call for a space character
	suite.mockOrganizationService.EXPECT().
		GetByName(" ").
		Return(nil, apperrors.ErrOrganizationNotFound).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/organizations/by-name/ ", nil)

	// The name parameter will be a space, which should return not found
	assert.Equal(suite.T(), http.StatusNotFound, recorder.Code)
	testutils.AssertErrorResponse(suite.T(), recorder, http.StatusNotFound, "organization not found")
}

// TestGetOrganizationByNameNotFound tests getting a non-existent organization by name
func (suite *OrganizationHandlerTestSuite) TestGetOrganizationByNameNotFound() {
	orgName := "non-existent-org"

	suite.mockOrganizationService.EXPECT().
		GetByName(orgName).
		Return(nil, apperrors.ErrOrganizationNotFound).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/organizations/by-name/%s", orgName), nil)

	assert.Equal(suite.T(), http.StatusNotFound, recorder.Code)
	testutils.AssertErrorResponse(suite.T(), recorder, http.StatusNotFound, "organization not found")
}

// TestUpdateOrganization tests updating an organization
func (suite *OrganizationHandlerTestSuite) TestUpdateOrganization() {
	orgID := uuid.New()
	requestBody := map[string]interface{}{
		"display_name": "Updated Organization",
		"description":  "Updated description",
	}

	expectedResponse := &service.OrganizationResponse{
		ID:          orgID,
		Name:        "test-org",
		DisplayName: "Updated Organization",
		Domain:      "test.com",
		Description: "Updated description",
		CreatedAt:   "2023-01-01T00:00:00Z",
		UpdatedAt:   "2023-01-01T00:00:00Z",
	}

	suite.mockOrganizationService.EXPECT().
		Update(orgID, gomock.Any()).
		Return(expectedResponse, nil).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("PUT", fmt.Sprintf("/api/v1/organizations/%s", orgID), requestBody)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response service.OrganizationResponse
	testutils.ParseJSONResponse(suite.T(), recorder, &response)
	assert.Equal(suite.T(), expectedResponse.DisplayName, response.DisplayName)
	assert.Equal(suite.T(), expectedResponse.Description, response.Description)
}

// TestDeleteOrganization tests deleting an organization
func (suite *OrganizationHandlerTestSuite) TestDeleteOrganization() {
	orgID := uuid.New()

	suite.mockOrganizationService.EXPECT().
		Delete(orgID).
		Return(nil).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s", orgID), nil)

	assert.Equal(suite.T(), http.StatusNoContent, recorder.Code)
}

// TestDeleteOrganizationNotFound tests deleting a non-existent organization
func (suite *OrganizationHandlerTestSuite) TestDeleteOrganizationNotFound() {
	orgID := uuid.New()

	suite.mockOrganizationService.EXPECT().
		Delete(orgID).
		Return(apperrors.ErrOrganizationNotFound).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s", orgID), nil)

	assert.Equal(suite.T(), http.StatusNotFound, recorder.Code)
	testutils.AssertErrorResponse(suite.T(), recorder, http.StatusNotFound, "organization not found")
}

// TestListOrganizations tests listing organizations
func (suite *OrganizationHandlerTestSuite) TestListOrganizations() {
	expectedResponse := &service.OrganizationListResponse{
		Organizations: []service.OrganizationResponse{
			{
				ID:          uuid.New(),
				Name:        "org-1",
				DisplayName: "Organization 1",
				Domain:      "org1.com",
				Description: "Description 1",
			},
			{
				ID:          uuid.New(),
				Name:        "org-2",
				DisplayName: "Organization 2",
				Domain:      "org2.com",
				Description: "Description 2",
			},
		},
		Total:    2,
		Page:     1,
		PageSize: 20,
	}

	suite.mockOrganizationService.EXPECT().
		GetAll(1, 20).
		Return(expectedResponse, nil).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/organizations/", nil)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response service.OrganizationListResponse
	testutils.ParseJSONResponse(suite.T(), recorder, &response)
	assert.Len(suite.T(), response.Organizations, 2)
	assert.Equal(suite.T(), int64(2), response.Total)
}

// TestListOrganizationsWithPagination tests listing organizations with pagination
func (suite *OrganizationHandlerTestSuite) TestListOrganizationsWithPagination() {
	expectedResponse := &service.OrganizationListResponse{
		Organizations: []service.OrganizationResponse{
			{
				ID:          uuid.New(),
				Name:        "org-3",
				DisplayName: "Organization 3",
				Domain:      "org3.com",
			},
		},
		Total:    3,
		Page:     3,
		PageSize: 1,
	}

	suite.mockOrganizationService.EXPECT().
		GetAll(3, 1).
		Return(expectedResponse, nil).
		Times(1)

	recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/organizations/?page=3&page_size=1", nil)

	assert.Equal(suite.T(), http.StatusOK, recorder.Code)

	var response service.OrganizationListResponse
	testutils.ParseJSONResponse(suite.T(), recorder, &response)
	assert.Len(suite.T(), response.Organizations, 1)
	assert.Equal(suite.T(), int64(3), response.Total)
	assert.Equal(suite.T(), 3, response.Page)
}

// TestOrganizationHandlerTestSuite runs the test suite
func TestOrganizationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationHandlerTestSuite))
}
