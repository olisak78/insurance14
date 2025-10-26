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

// ComponentDeploymentHandlerTestSuite defines the test suite for ComponentDeploymentHandler
type ComponentDeploymentHandlerTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockService *mocks.MockComponentDeploymentServiceInterface
	handler     *handlers.ComponentDeploymentHandler
	router      *gin.Engine
}

// SetupTest sets up the test suite
func (suite *ComponentDeploymentHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockComponentDeploymentServiceInterface(suite.ctrl)

	// Create a concrete ComponentDeploymentService instance for the handler
	// Note: Using nil service since we're testing HTTP interface behavior only
	suite.handler = handlers.NewComponentDeploymentHandler(nil)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownTest cleans up after each test
func (suite *ComponentDeploymentHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// setupRoutes sets up the routes for testing
func (suite *ComponentDeploymentHandlerTestSuite) setupRoutes() {
	suite.router.POST("/component-deployments", suite.handler.CreateComponentDeployment)
	suite.router.GET("/component-deployments/:id", suite.handler.GetComponentDeployment)
	suite.router.PUT("/component-deployments/:id", suite.handler.UpdateComponentDeployment)
	suite.router.DELETE("/component-deployments/:id", suite.handler.DeleteComponentDeployment)
	suite.router.GET("/components/:componentId/deployments", suite.handler.GetComponentDeploymentsByComponent)
	suite.router.GET("/landscapes/:landscapeId/deployments", suite.handler.GetComponentDeploymentsByLandscape)
}

// TestCreateComponentDeployment tests the CreateComponentDeployment handler
func (suite *ComponentDeploymentHandlerTestSuite) TestCreateComponentDeployment() {
	// Test validation error - invalid request body
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/component-deployments", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestGetComponentDeployment tests the GetComponentDeployment handler
func (suite *ComponentDeploymentHandlerTestSuite) TestGetComponentDeployment() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/component-deployments/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component deployment ID")
	})
}

// TestUpdateComponentDeployment tests the UpdateComponentDeployment handler
func (suite *ComponentDeploymentHandlerTestSuite) TestUpdateComponentDeployment() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		updateRequest := service.UpdateComponentDeploymentRequest{
			Version: "v1.1.0",
		}

		body, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/component-deployments/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component deployment ID")
	})

	// Test invalid JSON
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		deploymentID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/component-deployments/%s", deploymentID), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestDeleteComponentDeployment tests the DeleteComponentDeployment handler
func (suite *ComponentDeploymentHandlerTestSuite) TestDeleteComponentDeployment() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/component-deployments/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component deployment ID")
	})
}

// TestGetDeploymentsByComponent tests the GetComponentDeploymentsByComponent handler
func (suite *ComponentDeploymentHandlerTestSuite) TestGetDeploymentsByComponent() {
	// Test invalid component_id
	suite.T().Run("Invalid Component ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components/invalid-uuid/deployments", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid component ID")
	})
}

// TestGetDeploymentsByLandscape tests the GetComponentDeploymentsByLandscape handler
func (suite *ComponentDeploymentHandlerTestSuite) TestGetDeploymentsByLandscape() {
	// Test invalid landscape_id
	suite.T().Run("Invalid Landscape ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/landscapes/invalid-uuid/deployments", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid landscape ID")
	})
}

// TestComponentDeploymentHandlerTestSuite runs the test suite
func TestComponentDeploymentHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentDeploymentHandlerTestSuite))
}
