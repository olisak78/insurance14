package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/database/models"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ComponentHandlerTestSuite defines the test suite for ComponentHandler
type ComponentHandlerTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	mockService     *mocks.MockComponentServiceInterface
	mockTeamService *mocks.MockTeamServiceInterface
	handler         *handlers.ComponentHandler
	router          *gin.Engine
}

// SetupTest sets up the test suite
func (suite *ComponentHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockComponentServiceInterface(suite.ctrl)
	suite.mockTeamService = mocks.NewMockTeamServiceInterface(suite.ctrl)

	// Create a concrete ComponentService instance for the handler
	// Note: Using nil service since we're testing HTTP interface behavior only
	suite.handler = handlers.NewComponentHandler(nil, suite.mockTeamService)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownTest cleans up after each test
func (suite *ComponentHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// setupRoutes sets up the routes for testing
func (suite *ComponentHandlerTestSuite) setupRoutes() {
	suite.router.POST("/components", suite.handler.CreateComponent)
	suite.router.GET("/components/:id", suite.handler.GetComponent)
	suite.router.GET("/components/by-name/:name", suite.handler.GetComponentByName)
	suite.router.GET("/components/by-team/:id", suite.handler.GetComponentsByTeamID)
	suite.router.PUT("/components/:id", suite.handler.UpdateComponent)
	suite.router.DELETE("/components/:id", suite.handler.DeleteComponent)
	suite.router.GET("/components", suite.handler.ListComponents)
	suite.router.GET("/organizations/:orgId/components", suite.handler.GetComponentsByOrganization)
	suite.router.GET("/components/:id/ownerships", suite.handler.GetComponentWithOwnerships)
	suite.router.GET("/components/:id/deployments", suite.handler.GetComponentWithDeployments)
	suite.router.GET("/components/:id/projects", suite.handler.GetComponentWithProjects)
	suite.router.GET("/components/:id/details", suite.handler.GetComponentWithFullDetails)
}

// TestCreateComponent tests the CreateComponent handler
func (suite *ComponentHandlerTestSuite) TestCreateComponent() {
	// Test validation error - invalid request body
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/components", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestGetComponent tests the GetComponent handler
func (suite *ComponentHandlerTestSuite) TestGetComponent() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})
}

// TestUpdateComponent tests the UpdateComponent handler
func (suite *ComponentHandlerTestSuite) TestUpdateComponent() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		updateRequest := service.UpdateComponentRequest{
			DisplayName: "Updated User Service",
			Description: "Updated description",
		}

		body, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/components/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})

	// Test invalid JSON
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		componentID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/components/%s", componentID), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestDeleteComponent tests the DeleteComponent handler
func (suite *ComponentHandlerTestSuite) TestDeleteComponent() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/components/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})
}

// TestGetComponentsByOrganization tests the GetComponentsByOrganization handler
func (suite *ComponentHandlerTestSuite) TestGetComponentsByOrganization() {
	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/components", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid organization ID")
	})
}

// TestGetComponentByName tests the GetComponentByName handler
func (suite *ComponentHandlerTestSuite) TestGetComponentByName() {
	// Test missing organization_id parameter
	suite.T().Run("Missing Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/by-name/user-service", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "organization_id parameter is required")
	})

	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/by-name/user-service?organization_id=invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid organization ID")
	})
}

// TestListComponents tests the ListComponents handler
func (suite *ComponentHandlerTestSuite) TestListComponents() {
	// Test missing organization_id parameter
	suite.T().Run("Missing Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "organization_id parameter is required")
	})

	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components?organization_id=invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid organization ID")
	})
}

// TestGetComponentWithOwnerships tests the GetComponentWithOwnerships handler
func (suite *ComponentHandlerTestSuite) TestGetComponentWithOwnerships() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/invalid-uuid/ownerships", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})
}

// TestGetComponentWithDeployments tests the GetComponentWithDeployments handler
func (suite *ComponentHandlerTestSuite) TestGetComponentWithDeployments() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/invalid-uuid/deployments", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})
}

// TestGetComponentWithProjects tests the GetComponentWithProjects handler
func (suite *ComponentHandlerTestSuite) TestGetComponentWithProjects() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/invalid-uuid/projects", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})
}

// TestGetComponentWithFullDetails tests the GetComponentWithFullDetails handler
func (suite *ComponentHandlerTestSuite) TestGetComponentWithFullDetails() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/invalid-uuid/details", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})
}

// TestGetComponentsByTeamID tests the GetComponentsByTeamID handler
func (suite *ComponentHandlerTestSuite) TestGetComponentsByTeamID() {
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

		suite.mockTeamService.EXPECT().
			GetTeamComponentsByID(teamID, 1, 20).
			Return(components, int64(2), nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/components/by-team/%s", teamID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["total"])
		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(20), response["page_size"])
	})

	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/by-team/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid team ID")
	})

	// Test team not found
	suite.T().Run("Team Not Found", func(t *testing.T) {
		teamID := uuid.New()

		suite.mockTeamService.EXPECT().
			GetTeamComponentsByID(teamID, 1, 20).
			Return(nil, int64(0), apperrors.ErrTeamNotFound).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/components/by-team/%s", teamID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "team not found")
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

		suite.mockTeamService.EXPECT().
			GetTeamComponentsByID(teamID, 2, 10).
			Return(components, int64(15), nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/components/by-team/%s?page=2&page_size=10", teamID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(15), response["total"])
		assert.Equal(t, float64(2), response["page"])
		assert.Equal(t, float64(10), response["page_size"])
	})

	// Test internal server error
	suite.T().Run("Internal Server Error", func(t *testing.T) {
		teamID := uuid.New()

		suite.mockTeamService.EXPECT().
			GetTeamComponentsByID(teamID, 1, 20).
			Return(nil, int64(0), fmt.Errorf("database error")).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/components/by-team/%s", teamID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "database error")
	})
}

// TestComponentHandlerTestSuite runs the test suite
func TestComponentHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentHandlerTestSuite))
}
