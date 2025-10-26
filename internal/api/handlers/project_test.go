package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ProjectHandlerTestSuite defines the test suite for ProjectHandler
type ProjectHandlerTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockService *mocks.MockProjectServiceInterface
	handler     *handlers.ProjectHandler
	router      *gin.Engine
}

// SetupTest sets up the test suite
func (suite *ProjectHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockProjectServiceInterface(suite.ctrl)

	// Create a concrete ProjectService instance for the handler
	projectService := &service.ProjectService{}
	suite.handler = handlers.NewProjectHandler(projectService)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownTest cleans up after each test
func (suite *ProjectHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// setupRoutes sets up the routes for testing
func (suite *ProjectHandlerTestSuite) setupRoutes() {
	suite.router.POST("/projects", suite.handler.CreateProject)
	suite.router.GET("/projects/:id", suite.handler.GetProject)
	suite.router.PUT("/projects/:id", suite.handler.UpdateProject)
	suite.router.DELETE("/projects/:id", suite.handler.DeleteProject)
	suite.router.GET("/organizations/:orgId/projects", suite.handler.GetProjectsByOrganization)
}

// TestCreateProject tests the CreateProject handler
func (suite *ProjectHandlerTestSuite) TestCreateProject() {

	// Test validation error - invalid request body
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestGetProject tests the GetProject handler
func (suite *ProjectHandlerTestSuite) TestGetProject() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid project ID")
	})
}

// TestUpdateProject tests the UpdateProject handler
func (suite *ProjectHandlerTestSuite) TestUpdateProject() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		updateRequest := service.UpdateProjectRequest{
			DisplayName: "Updated Customer Portal",
			Description: "Updated description",
		}

		body, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/projects/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid project ID")
	})

	// Test invalid JSON
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		projectID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/projects/%s", projectID), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestDeleteProject tests the DeleteProject handler
func (suite *ProjectHandlerTestSuite) TestDeleteProject() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/projects/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid project ID")
	})
}

// TestGetProjectsByOrganization tests the GetProjectsByOrganization handler
func (suite *ProjectHandlerTestSuite) TestGetProjectsByOrganization() {
	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/projects", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid organization ID")
	})
}

// TestProjectHandlerTestSuite runs the test suite
func TestProjectHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectHandlerTestSuite))
}
