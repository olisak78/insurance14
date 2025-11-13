package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
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
	suite.handler = handlers.NewLandscapeHandler(suite.mockService)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownTest cleans up after each test
func (suite *LandscapeHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// setupRoutes sets up the routes for testing
func (suite *LandscapeHandlerTestSuite) setupRoutes() {
	suite.router.GET("/landscapes", suite.handler.ListLandscapesByQuery)
}


// TestListLandscapesByQuery tests the ListLandscapesByQuery handler
func (suite *LandscapeHandlerTestSuite) TestListLandscapesByQuery() {
	// Test without query parameters
	suite.T().Run("Without Query Parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/landscapes", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		// Should return 400 since project-name is required
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test with project-name query parameter
	suite.T().Run("With Project Name Query", func(t *testing.T) {
		// Setup mock expectation for project name lookup
		suite.mockService.EXPECT().
			GetByProjectNameAll("test-project").
			Return([]service.LandscapeMinimalResponse{}, nil)
		
		req := httptest.NewRequest(http.MethodGet, "/landscapes?project-name=test-project", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Run the test suite
func TestLandscapeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LandscapeHandlerTestSuite))
}
