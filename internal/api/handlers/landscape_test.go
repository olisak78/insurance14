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

// LandscapeHandlerTestSuite defines the test suite for LandscapeHandler
type LandscapeHandlerTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockService *mocks.MockLandscapeServiceInterface
	handler     *handlers.LandscapeHandler
	router      *gin.Engine
}

// SetupTest sets up the test suite
func (suite *LandscapeHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockLandscapeServiceInterface(suite.ctrl)

	// Create a concrete LandscapeService instance for the handler
	// Note: Using nil service since we're testing HTTP interface behavior only
	suite.handler = handlers.NewLandscapeHandler(nil)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownTest cleans up after each test
func (suite *LandscapeHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// setupRoutes sets up the routes for testing
func (suite *LandscapeHandlerTestSuite) setupRoutes() {
	suite.router.POST("/landscapes", suite.handler.CreateLandscape)
	suite.router.GET("/landscapes/:id", suite.handler.GetLandscape)
	suite.router.PUT("/landscapes/:id", suite.handler.UpdateLandscape)
	suite.router.DELETE("/landscapes/:id", suite.handler.DeleteLandscape)
	suite.router.GET("/organizations/:orgId/landscapes", suite.handler.GetLandscapesByOrganization)
}

// TestCreateLandscape tests the CreateLandscape handler
func (suite *LandscapeHandlerTestSuite) TestCreateLandscape() {
	// Test validation error - invalid request body
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/landscapes", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestGetLandscape tests the GetLandscape handler
func (suite *LandscapeHandlerTestSuite) TestGetLandscape() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/landscapes/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid landscape ID")
	})
}

// TestUpdateLandscape tests the UpdateLandscape handler
func (suite *LandscapeHandlerTestSuite) TestUpdateLandscape() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		updateRequest := service.UpdateLandscapeRequest{
			DisplayName: "Updated Production Environment",
			Description: "Updated description",
		}

		body, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/landscapes/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid landscape ID")
	})

	// Test invalid JSON
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		landscapeID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/landscapes/%s", landscapeID), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestDeleteLandscape tests the DeleteLandscape handler
func (suite *LandscapeHandlerTestSuite) TestDeleteLandscape() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/landscapes/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid landscape ID")
	})
}

// TestGetLandscapesByOrganization tests the GetLandscapesByOrganization handler
func (suite *LandscapeHandlerTestSuite) TestGetLandscapesByOrganization() {
	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/landscapes", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid organization ID")
	})
}

// TestLandscapeHandlerTestSuite runs the test suite
func TestLandscapeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LandscapeHandlerTestSuite))
}
