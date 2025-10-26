package handlers_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/database/models"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// TeamHandlerTestSuite defines the test suite for TeamHandler
type TeamHandlerTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockService *mocks.MockTeamServiceInterface
	handler     *handlers.TeamHandler
	httpSuite   *testutils.HTTPTestSuite
}

// SetupTest sets up the test suite
func (suite *TeamHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockTeamServiceInterface(suite.ctrl)

	// Create handler with mock service
	suite.handler = handlers.NewTeamHandler(suite.mockService)

	// Setup HTTP test suite
	suite.httpSuite = testutils.SetupHTTPTest()

	// Register routes
	v1 := suite.httpSuite.Router.Group("/api/v1")
	teams := v1.Group("/teams")
	{
		teams.POST("/", suite.handler.CreateTeam)
		teams.GET("/:id", suite.handler.GetTeam)
		teams.GET("/", suite.handler.ListTeams)
		teams.PUT("/:id", suite.handler.UpdateTeam)
		teams.DELETE("/:id", suite.handler.DeleteTeam)
		teams.GET("/by-name/:name", suite.handler.GetTeamByName)
		teams.POST("/:id/links", suite.handler.AddLink)
		teams.DELETE("/:id/links", suite.handler.RemoveLink)
	}

	// Organization teams route
	v1.GET("/organizations/:orgId/teams", suite.handler.GetTeamsByOrganization)

	// Team components route
	teams.GET("/:id/components", suite.handler.GetTeamComponents)
}

// TearDownTest cleans up after each test
func (suite *TeamHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// Helper method to make invalid JSON requests
func (suite *TeamHandlerTestSuite) makeInvalidJSONRequest(method, url string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	suite.httpSuite.Router.ServeHTTP(recorder, req)

	return recorder
}

// TestCreateTeam tests the CreateTeam handler
func (suite *TeamHandlerTestSuite) TestCreateTeam() {
	// Test successful team creation
	suite.T().Run("Success", func(t *testing.T) {
		orgID := uuid.New()
		teamLeadID := uuid.New()
		teamID := uuid.New()

		requestBody := map[string]interface{}{
			"organization_id": orgID.String(),
			"name":            "backend-team",
			"display_name":    "Backend Development Team",
			"description":     "Team responsible for backend services",
			"status":          "active",
			"team_lead_id":    teamLeadID.String(),
		}

		expectedResponse := &service.TeamResponse{
			ID:             teamID,
			OrganizationID: orgID,
			Name:           "backend-team",
			DisplayName:    "Backend Development Team",
			Description:    "Team responsible for backend services",
			Status:         models.TeamStatusActive,
			TeamLeadID:     &teamLeadID,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		}

		suite.mockService.EXPECT().
			Create(gomock.Any()).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("POST", "/api/v1/teams/", requestBody)

		assert.Equal(t, http.StatusCreated, recorder.Code)

		var response service.TeamResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Equal(t, expectedResponse.Name, response.Name)
		assert.Equal(t, expectedResponse.DisplayName, response.DisplayName)
	})

	// Test validation error - service returns error
	suite.T().Run("Service Error", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"name": "invalid-team",
		}

		suite.mockService.EXPECT().
			Create(gomock.Any()).
			Return(nil, fmt.Errorf("validation error: organization_id is required")).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("POST", "/api/v1/teams/", requestBody)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusInternalServerError, "validation error: organization_id is required")
	})

	// Test organization not found
	suite.T().Run("Organization Not Found", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"organization_id": uuid.New().String(),
			"name":            "test-team",
			"display_name":    "Test Team",
		}

		suite.mockService.EXPECT().
			Create(gomock.Any()).
			Return(nil, apperrors.ErrOrganizationNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("POST", "/api/v1/teams/", requestBody)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "organization not found")
	})

	// Test invalid JSON
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		recorder := suite.makeInvalidJSONRequest("POST", "/api/v1/teams/")

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

// TestGetTeam tests the GetTeam handler
func (suite *TeamHandlerTestSuite) TestGetTeam() {
	// Test successful retrieval
	suite.T().Run("Success", func(t *testing.T) {
		teamID := uuid.New()
		orgID := uuid.New()

		expectedResponse := &service.TeamResponse{
			ID:             teamID,
			OrganizationID: orgID,
			Name:           "backend-team",
			DisplayName:    "Backend Development Team",
			Description:    "Team responsible for backend services",
			Status:         models.TeamStatusActive,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		}

		suite.mockService.EXPECT().
			GetByID(teamID).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/%s", teamID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response service.TeamResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Equal(t, expectedResponse.ID, response.ID)
		assert.Equal(t, expectedResponse.Name, response.Name)
	})

	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/teams/invalid-uuid", nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid team ID")
	})

	// Test team not found
	suite.T().Run("Not Found", func(t *testing.T) {
		teamID := uuid.New()

		suite.mockService.EXPECT().
			GetByID(teamID).
			Return(nil, apperrors.ErrTeamNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/%s", teamID), nil)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "team not found")
	})
}

// TestUpdateTeam tests the UpdateTeam handler
func (suite *TeamHandlerTestSuite) TestUpdateTeam() {
	// Test successful update
	suite.T().Run("Success", func(t *testing.T) {
		teamID := uuid.New()
		orgID := uuid.New()

		requestBody := map[string]interface{}{
			"display_name": "Updated Backend Team",
			"description":  "Updated description",
		}

		expectedResponse := &service.TeamResponse{
			ID:             teamID,
			OrganizationID: orgID,
			Name:           "backend-team",
			DisplayName:    "Updated Backend Team",
			Description:    "Updated description",
			Status:         models.TeamStatusActive,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		}

		suite.mockService.EXPECT().
			Update(teamID, gomock.Any()).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("PUT", fmt.Sprintf("/api/v1/teams/%s", teamID), requestBody)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response service.TeamResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Equal(t, expectedResponse.DisplayName, response.DisplayName)
		assert.Equal(t, expectedResponse.Description, response.Description)
	})

	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"display_name": "Updated Team",
		}

		recorder := suite.httpSuite.MakeRequest("PUT", "/api/v1/teams/invalid-uuid", requestBody)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid team ID")
	})

	// Test team not found
	suite.T().Run("Not Found", func(t *testing.T) {
		teamID := uuid.New()
		requestBody := map[string]interface{}{
			"display_name": "Updated Team",
		}

		suite.mockService.EXPECT().
			Update(teamID, gomock.Any()).
			Return(nil, apperrors.ErrTeamNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("PUT", fmt.Sprintf("/api/v1/teams/%s", teamID), requestBody)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "team not found")
	})

	// Test invalid JSON
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		teamID := uuid.New()
		recorder := suite.makeInvalidJSONRequest("PUT", fmt.Sprintf("/api/v1/teams/%s", teamID))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

// TestDeleteTeam tests the DeleteTeam handler
func (suite *TeamHandlerTestSuite) TestDeleteTeam() {
	// Test successful deletion
	suite.T().Run("Success", func(t *testing.T) {
		teamID := uuid.New()

		suite.mockService.EXPECT().
			Delete(teamID).
			Return(nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/%s", teamID), nil)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})

	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		recorder := suite.httpSuite.MakeRequest("DELETE", "/api/v1/teams/invalid-uuid", nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid team ID")
	})

	// Test team not found
	suite.T().Run("Not Found", func(t *testing.T) {
		teamID := uuid.New()

		suite.mockService.EXPECT().
			Delete(teamID).
			Return(apperrors.ErrTeamNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/%s", teamID), nil)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "team not found")
	})
}

// TestListTeams tests the ListTeams handler
func (suite *TeamHandlerTestSuite) TestListTeams() {
	// Test successful listing
	suite.T().Run("Success", func(t *testing.T) {
		orgID := uuid.New()
		expectedResponse := &service.TeamListResponse{
			Teams: []service.TeamResponse{
				{
					ID:             uuid.New(),
					OrganizationID: orgID,
					Name:           "backend-team",
					DisplayName:    "Backend Team",
					Status:         models.TeamStatusActive,
				},
				{
					ID:             uuid.New(),
					OrganizationID: orgID,
					Name:           "frontend-team",
					DisplayName:    "Frontend Team",
					Status:         models.TeamStatusActive,
				},
			},
			Total:    2,
			Page:     1,
			PageSize: 20,
		}

		suite.mockService.EXPECT().
			GetByOrganization(orgID, 1, 20).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/?organization_id=%s", orgID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response service.TeamListResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Len(t, response.Teams, 2)
		assert.Equal(t, int64(2), response.Total)
	})

	// Test missing organization_id
	suite.T().Run("Missing Organization ID", func(t *testing.T) {
		recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/teams/", nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "organization_id parameter is required")
	})

	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/teams/?organization_id=invalid-uuid", nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid organization ID")
	})

	// Test with search parameter
	suite.T().Run("Search Teams", func(t *testing.T) {
		orgID := uuid.New()
		expectedResponse := &service.TeamListResponse{
			Teams: []service.TeamResponse{
				{
					ID:             uuid.New(),
					OrganizationID: orgID,
					Name:           "backend-team",
					DisplayName:    "Backend Team",
					Status:         models.TeamStatusActive,
				},
			},
			Total:    1,
			Page:     1,
			PageSize: 20,
		}

		suite.mockService.EXPECT().
			Search(orgID, "backend", 1, 20).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/?organization_id=%s&search=backend", orgID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response service.TeamListResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Len(t, response.Teams, 1)
	})
}

// TestGetTeamByName tests the GetTeamByName handler
func (suite *TeamHandlerTestSuite) TestGetTeamByName() {
	// Test successful retrieval
	suite.T().Run("Success", func(t *testing.T) {
		orgID := uuid.New()
		teamName := "backend-team"
		teamID := uuid.New()

		expectedResponse := &service.TeamResponse{
			ID:             teamID,
			OrganizationID: orgID,
			Name:           teamName,
			DisplayName:    "Backend Development Team",
			Description:    "Team responsible for backend services",
			Status:         models.TeamStatusActive,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		}

		suite.mockService.EXPECT().
			GetByName(orgID, teamName).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/by-name/%s?organization_id=%s", teamName, orgID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response service.TeamResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Equal(t, expectedResponse.ID, response.ID)
		assert.Equal(t, expectedResponse.Name, response.Name)
		assert.Equal(t, expectedResponse.DisplayName, response.DisplayName)
	})

	// Test missing organization_id parameter
	suite.T().Run("Missing Organization ID", func(t *testing.T) {
		teamName := "backend-team"
		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/by-name/%s", teamName), nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "organization_id parameter is required")
	})

	// Test invalid organization_id parameter
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		teamName := "backend-team"
		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/by-name/%s?organization_id=invalid-uuid", teamName), nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid organization ID")
	})

	// Test team not found
	suite.T().Run("Team Not Found", func(t *testing.T) {
		orgID := uuid.New()
		teamName := "nonexistent-team"

		suite.mockService.EXPECT().
			GetByName(orgID, teamName).
			Return(nil, apperrors.ErrTeamNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/by-name/%s?organization_id=%s", teamName, orgID), nil)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "team not found")
	})

	// Test service error
	suite.T().Run("Service Error", func(t *testing.T) {
		orgID := uuid.New()
		teamName := "backend-team"

		suite.mockService.EXPECT().
			GetByName(orgID, teamName).
			Return(nil, fmt.Errorf("database connection error")).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/by-name/%s?organization_id=%s", teamName, orgID), nil)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusInternalServerError, "database connection error")
	})

	// Test with special characters in team name
	suite.T().Run("Team Name With Special Characters", func(t *testing.T) {
		orgID := uuid.New()
		teamName := "backend-team-2024"
		teamID := uuid.New()

		expectedResponse := &service.TeamResponse{
			ID:             teamID,
			OrganizationID: orgID,
			Name:           teamName,
			DisplayName:    "Backend Team 2024",
			Status:         models.TeamStatusActive,
			CreatedAt:      "2023-01-01T00:00:00Z",
			UpdatedAt:      "2023-01-01T00:00:00Z",
		}

		suite.mockService.EXPECT().
			GetByName(orgID, teamName).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/by-name/%s?organization_id=%s", teamName, orgID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response service.TeamResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Equal(t, expectedResponse.Name, response.Name)
	})
}

// TestGetTeamsByOrganization tests the GetTeamsByOrganization handler
func (suite *TeamHandlerTestSuite) TestGetTeamsByOrganization() {
	// Test successful retrieval
	suite.T().Run("Success", func(t *testing.T) {
		orgID := uuid.New()
		expectedResponse := &service.TeamListResponse{
			Teams: []service.TeamResponse{
				{
					ID:             uuid.New(),
					OrganizationID: orgID,
					Name:           "platform-team",
					DisplayName:    "Platform Team",
					Status:         models.TeamStatusActive,
				},
			},
			Total:    1,
			Page:     1,
			PageSize: 20,
		}

		suite.mockService.EXPECT().
			GetByOrganization(orgID, 1, 20).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/teams", orgID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response service.TeamListResponse
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Len(t, response.Teams, 1)
	})

	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/organizations/invalid-uuid/teams", nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid organization ID")
	})

	// Test organization not found
	suite.T().Run("Organization Not Found", func(t *testing.T) {
		orgID := uuid.New()

		suite.mockService.EXPECT().
			GetByOrganization(orgID, 1, 20).
			Return(nil, apperrors.ErrOrganizationNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/teams", orgID), nil)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "organization not found")
	})
}

// TestAddLink tests the AddLink handler
func (suite *TeamHandlerTestSuite) TestAddLink() {
	// Test successful link addition
	suite.T().Run("Success", func(t *testing.T) {
		teamID := uuid.New()

		requestBody := map[string]interface{}{
			"url":      "https://github.com/myteam/repo",
			"title":    "Team Repository",
			"icon":     "github",
			"category": "repository",
		}

		expectedResponse := &service.TeamResponse{
			ID:          teamID,
			Name:        "backend-team",
			DisplayName: "Backend Team",
			Links: []service.Link{
				{
					URL:      "https://github.com/myteam/repo",
					Title:    "Team Repository",
					Icon:     "github",
					Category: "repository",
				},
			},
		}

		suite.mockService.EXPECT().
			AddLink(teamID, gomock.Any()).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("POST", fmt.Sprintf("/api/v1/teams/%s/links", teamID), requestBody)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	// Test invalid team ID
	suite.T().Run("Invalid Team ID", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"url":   "https://github.com/myteam/repo",
			"title": "Team Repository",
		}

		recorder := suite.httpSuite.MakeRequest("POST", "/api/v1/teams/invalid-uuid/links", requestBody)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid team ID")
	})

	// Test invalid request body
	suite.T().Run("Invalid Request Body", func(t *testing.T) {
		teamID := uuid.New()

		recorder := suite.makeInvalidJSONRequest("POST", fmt.Sprintf("/api/v1/teams/%s/links", teamID))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	// Test team not found
	suite.T().Run("Team Not Found", func(t *testing.T) {
		teamID := uuid.New()

		requestBody := map[string]interface{}{
			"url":   "https://github.com/myteam/repo",
			"title": "Team Repository",
		}

		suite.mockService.EXPECT().
			AddLink(teamID, gomock.Any()).
			Return(nil, apperrors.ErrTeamNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("POST", fmt.Sprintf("/api/v1/teams/%s/links", teamID), requestBody)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "team not found")
	})

	// Test duplicate link
	suite.T().Run("Duplicate Link", func(t *testing.T) {
		teamID := uuid.New()

		requestBody := map[string]interface{}{
			"url":   "https://github.com/myteam/repo",
			"title": "Team Repository",
		}

		suite.mockService.EXPECT().
			AddLink(teamID, gomock.Any()).
			Return(nil, apperrors.ErrLinkExists).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("POST", fmt.Sprintf("/api/v1/teams/%s/links", teamID), requestBody)

		assert.Equal(t, http.StatusConflict, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusConflict, "link already exists")
	})

	// Test internal server error
	suite.T().Run("Internal Server Error", func(t *testing.T) {
		teamID := uuid.New()

		requestBody := map[string]interface{}{
			"url":   "https://github.com/myteam/repo",
			"title": "Team Repository",
		}

		suite.mockService.EXPECT().
			AddLink(teamID, gomock.Any()).
			Return(nil, fmt.Errorf("database error")).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("POST", fmt.Sprintf("/api/v1/teams/%s/links", teamID), requestBody)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusInternalServerError, "database error")
	})
}

// TestRemoveLink tests the RemoveLink handler
func (suite *TeamHandlerTestSuite) TestRemoveLink() {
	// Test successful link removal
	suite.T().Run("Success", func(t *testing.T) {
		teamID := uuid.New()
		linkURL := "https://github.com/myteam/repo"

		expectedResponse := &service.TeamResponse{
			ID:          teamID,
			Name:        "backend-team",
			DisplayName: "Backend Team",
			Links:       []service.Link{},
		}

		suite.mockService.EXPECT().
			RemoveLink(teamID, linkURL).
			Return(expectedResponse, nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/%s/links?url=%s", teamID, linkURL), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	// Test invalid team ID
	suite.T().Run("Invalid Team ID", func(t *testing.T) {
		linkURL := "https://github.com/myteam/repo"

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/invalid-uuid/links?url=%s", linkURL), nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid team ID")
	})

	// Test missing URL parameter
	suite.T().Run("Missing URL Parameter", func(t *testing.T) {
		teamID := uuid.New()

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/%s/links", teamID), nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "url query parameter is required")
	})

	// Test team not found
	suite.T().Run("Team Not Found", func(t *testing.T) {
		teamID := uuid.New()
		linkURL := "https://github.com/myteam/repo"

		suite.mockService.EXPECT().
			RemoveLink(teamID, linkURL).
			Return(nil, apperrors.ErrTeamNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/%s/links?url=%s", teamID, linkURL), nil)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "team not found")
	})

	// Test link not found
	suite.T().Run("Link Not Found", func(t *testing.T) {
		teamID := uuid.New()
		linkURL := "https://github.com/myteam/nonexistent"

		suite.mockService.EXPECT().
			RemoveLink(teamID, linkURL).
			Return(nil, apperrors.ErrLinkNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/%s/links?url=%s", teamID, linkURL), nil)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "link not found")
	})

	// Test internal server error
	suite.T().Run("Internal Server Error", func(t *testing.T) {
		teamID := uuid.New()
		linkURL := "https://github.com/myteam/repo"

		suite.mockService.EXPECT().
			RemoveLink(teamID, linkURL).
			Return(nil, fmt.Errorf("database error")).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("DELETE", fmt.Sprintf("/api/v1/teams/%s/links?url=%s", teamID, linkURL), nil)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusInternalServerError, "database error")
	})
}

// TestGetTeamComponents tests the GetTeamComponents handler
func (suite *TeamHandlerTestSuite) TestGetTeamComponents() {
	// Test successful retrieval
	suite.T().Run("Success", func(t *testing.T) {
		teamID := uuid.New()

		components := []models.Component{
			{
				BaseModel: models.BaseModel{
					ID: uuid.New(),
				},
				Name:        "user-service",
				DisplayName: "User Service",
			},
			{
				BaseModel: models.BaseModel{
					ID: uuid.New(),
				},
				Name:        "auth-service",
				DisplayName: "Auth Service",
			},
		}

		suite.mockService.EXPECT().
			GetTeamComponentsByID(teamID, 1, 20).
			Return(components, int64(2), nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/%s/components", teamID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Equal(t, float64(2), response["total"])
		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(20), response["page_size"])
	})

	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		recorder := suite.httpSuite.MakeRequest("GET", "/api/v1/teams/invalid-uuid/components", nil)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusBadRequest, "invalid team ID")
	})

	// Test team not found
	suite.T().Run("Team Not Found", func(t *testing.T) {
		teamID := uuid.New()

		suite.mockService.EXPECT().
			GetTeamComponentsByID(teamID, 1, 20).
			Return(nil, int64(0), apperrors.ErrTeamNotFound).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/%s/components", teamID), nil)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusNotFound, "team not found")
	})

	// Test with pagination
	suite.T().Run("With Pagination", func(t *testing.T) {
		teamID := uuid.New()

		components := []models.Component{
			{
				BaseModel: models.BaseModel{
					ID: uuid.New(),
				},
				Name:        "user-service",
				DisplayName: "User Service",
			},
		}

		suite.mockService.EXPECT().
			GetTeamComponentsByID(teamID, 2, 10).
			Return(components, int64(15), nil).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/%s/components?page=2&page_size=10", teamID), nil)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		testutils.ParseJSONResponse(t, recorder, &response)
		assert.Equal(t, float64(15), response["total"])
		assert.Equal(t, float64(2), response["page"])
		assert.Equal(t, float64(10), response["page_size"])
	})

	// Test internal server error
	suite.T().Run("Internal Server Error", func(t *testing.T) {
		teamID := uuid.New()

		suite.mockService.EXPECT().
			GetTeamComponentsByID(teamID, 1, 20).
			Return(nil, int64(0), fmt.Errorf("database error")).
			Times(1)

		recorder := suite.httpSuite.MakeRequest("GET", fmt.Sprintf("/api/v1/teams/%s/components", teamID), nil)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		testutils.AssertErrorResponse(t, recorder, http.StatusInternalServerError, "database error")
	})
}

// TestTeamHandlerTestSuite runs the test suite
func TestTeamHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TeamHandlerTestSuite))
}
