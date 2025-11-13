package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ComponentHandlerTestSuite defines the test suite for ComponentHandler (aligned to current /components API)
type ComponentHandlerTestSuite struct {
	suite.Suite
	handler *handlers.ComponentHandler
	router  *gin.Engine
}

// SetupTest sets up the test suite
func (suite *ComponentHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	// ComponentHandler: only /components GET is supported with team-id or project-name
	// No component service or team service required for the negative-path tests we keep
	suite.handler = handlers.NewComponentHandler(nil, nil)
	suite.router = gin.New()
	suite.setupRoutes()
}

// setupRoutes sets up only the relevant route for testing
func (suite *ComponentHandlerTestSuite) setupRoutes() {
	// Only the supported endpoint remains
	suite.router.GET("/components", suite.handler.ListComponents)
}

// TestListComponents tests the ListComponents handler negative-paths
func (suite *ComponentHandlerTestSuite) TestListComponents() {
	// Missing both team-id and project-name -> 400
	suite.T().Run("Missing Parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "team-id or project-name parameter is required")
	})

	// Invalid team-id format -> 400
	suite.T().Run("Invalid Team ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/components?team-id=invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid team ID")
	})
}

// TestComponentHandlerTestSuite runs the test suite
func TestComponentHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentHandlerTestSuite))
}
