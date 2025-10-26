package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// MockJiraService is a mock implementation of JiraServiceInterface for testing
type MockJiraService struct {
	ctrl                     *gomock.Controller
	getIssuesFunc            func(filters service.JiraIssueFilters) (*service.JiraIssuesResponse, error)
	getIssuesCountFunc       func(filters service.JiraIssueFilters) (int, error)
}

func NewMockJiraService(ctrl *gomock.Controller) *MockJiraService {
	return &MockJiraService{ctrl: ctrl}
}

func (m *MockJiraService) GetIssues(filters service.JiraIssueFilters) (*service.JiraIssuesResponse, error) {
	if m.getIssuesFunc != nil {
		return m.getIssuesFunc(filters)
	}
	return nil, nil
}

func (m *MockJiraService) GetIssuesCount(filters service.JiraIssueFilters) (int, error) {
	if m.getIssuesCountFunc != nil {
		return m.getIssuesCountFunc(filters)
	}
	return 0, nil
}

// JiraHandlerTestSuite defines the test suite for JiraHandler
type JiraHandlerTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockService *MockJiraService
	handler     *handlers.JiraHandler
	router      *gin.Engine
}

// SetupTest sets up the test suite
func (suite *JiraHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = NewMockJiraService(suite.ctrl)

	// Create handler with mock service
	suite.handler = handlers.NewJiraHandler(suite.mockService)
	suite.router = gin.New()
	suite.setupRoutes()
}

// TearDownTest cleans up after each test
func (suite *JiraHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// setupRoutes sets up the routes for testing
func (suite *JiraHandlerTestSuite) setupRoutes() {
	suite.router.GET("/jira/issues", suite.handler.GetIssues)
	suite.router.GET("/jira/issues/me", suite.handler.GetMyIssues)
	suite.router.GET("/jira/issues/me/count", suite.handler.GetMyIssuesCount)
}

// TestGetIssues tests the consolidated GetIssues handler
func (suite *JiraHandlerTestSuite) TestGetIssues() {
	suite.T().Run("Successful request with project and status", func(t *testing.T) {
		expectedResponse := &service.JiraIssuesResponse{
			Total: 2,
			Issues: []service.JiraIssue{
				{
					ID:  "1",
					Key: "SAPBTPCFS-123",
					Fields: service.JiraIssueFields{
						Summary: "Test issue 1",
						Status:  service.JiraStatus{ID: "1", Name: "Open"},
						IssueType: service.JiraIssueType{ID: "1", Name: "Story"},
						Created: "2023-01-01T00:00:00.000Z",
						Updated: "2023-01-02T00:00:00.000Z",
					},
				},
				{
					ID:  "2",
					Key: "SAPBTPCFS-124",
					Fields: service.JiraIssueFields{
						Summary: "Test issue 2",
						Status:  service.JiraStatus{ID: "2", Name: "In Progress"},
						IssueType: service.JiraIssueType{ID: "1", Name: "Story"},
						Created: "2023-01-01T00:00:00.000Z",
						Updated: "2023-01-02T00:00:00.000Z",
					},
				},
			},
		}

		suite.mockService.getIssuesFunc = func(filters service.JiraIssueFilters) (*service.JiraIssuesResponse, error) {
			assert.Equal(t, "SAPBTPCFS", filters.Project)
			assert.Equal(t, "Open,In Progress", filters.Status)
			assert.Equal(t, "TestTeam", filters.Team)
			return expectedResponse, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/jira/issues?project=SAPBTPCFS&status=Open%2CIn+Progress&team=TestTeam", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "SAPBTPCFS-123")
		assert.Contains(t, w.Body.String(), "Test issue 1")
		assert.Contains(t, w.Body.String(), `"total":2`)
	})

	suite.T().Run("Successful request with minimal parameters", func(t *testing.T) {
		expectedResponse := &service.JiraIssuesResponse{
			Total: 1,
			Issues: []service.JiraIssue{
				{
					ID:  "1",
					Key: "SAPBTPCFS-999",
					Fields: service.JiraIssueFields{
						Summary: "Minimal test issue",
						Status:  service.JiraStatus{ID: "1", Name: "Open"},
						IssueType: service.JiraIssueType{ID: "1", Name: "Bug"},
						Created: "2023-01-01T00:00:00.000Z",
						Updated: "2023-01-02T00:00:00.000Z",
					},
				},
			},
		}

		suite.mockService.getIssuesFunc = func(filters service.JiraIssueFilters) (*service.JiraIssuesResponse, error) {
			assert.Equal(t, "", filters.Project)
			assert.Equal(t, "", filters.Status)
			assert.Equal(t, "", filters.Team)
			return expectedResponse, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/jira/issues", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "SAPBTPCFS-999")
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	suite.T().Run("Service error", func(t *testing.T) {
		suite.mockService.getIssuesFunc = func(filters service.JiraIssueFilters) (*service.JiraIssuesResponse, error) {
			return nil, errors.New("jira connection failed")
		}

		req := httptest.NewRequest(http.MethodGet, "/jira/issues?project=SAPBTPCFS", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
		assert.Contains(t, w.Body.String(), "jira search failed")
		assert.Contains(t, w.Body.String(), "jira connection failed")
	})
}

// TestGetMyIssues tests the consolidated GetMyIssues handler
func (suite *JiraHandlerTestSuite) TestGetMyIssues() {
	suite.T().Run("Missing authentication", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authentication required")
	})

	suite.T().Run("Invalid authentication claims", func(t *testing.T) {
		router := gin.New()
		router.GET("/jira/issues/me", func(c *gin.Context) {
			c.Set("auth_claims", "invalid_claims")
			suite.handler.GetMyIssues(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authentication claims")
	})

	suite.T().Run("Missing username in claims", func(t *testing.T) {
		router := gin.New()
		router.GET("/jira/issues/me", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID: 12345,
				// Username is empty
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssues(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Username not available in authentication claims")
	})

	suite.T().Run("Successful request with full issues", func(t *testing.T) {
		expectedResponse := &service.JiraIssuesResponse{
			Total: 2,
			Issues: []service.JiraIssue{
				{
					ID:  "1",
					Key: "SAPBTPCFS-999",
					Fields: service.JiraIssueFields{
						Summary: "My issue 1",
						Status:  service.JiraStatus{ID: "1", Name: "Open"},
						IssueType: service.JiraIssueType{ID: "1", Name: "Story"},
						Assignee: &service.JiraUser{
							AccountID:   "12345",
							DisplayName: "Test User",
						},
						Created: "2023-01-01T00:00:00.000Z",
						Updated: "2023-01-02T00:00:00.000Z",
					},
				},
			},
		}

		suite.mockService.getIssuesFunc = func(filters service.JiraIssueFilters) (*service.JiraIssuesResponse, error) {
			assert.Equal(t, "testuser", filters.User)
			assert.Equal(t, "Open", filters.Status)
			assert.Equal(t, "SAPBTPCFS", filters.Project)
			return expectedResponse, nil
		}

		router := gin.New()
		router.GET("/jira/issues/me", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssues(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me?status=Open&project=SAPBTPCFS", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "SAPBTPCFS-999")
		assert.Contains(t, w.Body.String(), "My issue 1")
		assert.Contains(t, w.Body.String(), `"total":2`)
	})

	suite.T().Run("Service error", func(t *testing.T) {
		suite.mockService.getIssuesFunc = func(filters service.JiraIssueFilters) (*service.JiraIssuesResponse, error) {
			return nil, errors.New("jira service unavailable")
		}

		router := gin.New()
		router.GET("/jira/issues/me", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssues(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
		assert.Contains(t, w.Body.String(), "jira search failed")
		assert.Contains(t, w.Body.String(), "jira service unavailable")
	})
}

// TestGetMyIssuesCount tests the consolidated GetMyIssuesCount handler
func (suite *JiraHandlerTestSuite) TestGetMyIssuesCount() {
	suite.T().Run("Missing authentication", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me/count?status=Resolved", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authentication required")
	})

	suite.T().Run("Missing status parameter", func(t *testing.T) {
		router := gin.New()
		router.GET("/jira/issues/me/count", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssuesCount(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me/count", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "missing query parameter: status")
	})

	suite.T().Run("Invalid date format", func(t *testing.T) {
		router := gin.New()
		router.GET("/jira/issues/me/count", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssuesCount(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me/count?status=Resolved&date=invalid-date", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid date format: must be yyyy-MM-dd")
	})

	suite.T().Run("Successful request with resolved status and default date", func(t *testing.T) {
		suite.mockService.getIssuesCountFunc = func(filters service.JiraIssueFilters) (int, error) {
			assert.Equal(t, "testuser", filters.User)
			assert.Equal(t, "Resolved", filters.Status)
			assert.NotEmpty(t, filters.Date) // Should have default date
			return 7, nil
		}

		router := gin.New()
		router.GET("/jira/issues/me/count", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssuesCount(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me/count?status=Resolved", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"count":7`)
	})

	suite.T().Run("Successful request with custom date", func(t *testing.T) {
		suite.mockService.getIssuesCountFunc = func(filters service.JiraIssueFilters) (int, error) {
			assert.Equal(t, "testuser", filters.User)
			assert.Equal(t, "Resolved", filters.Status)
			assert.Equal(t, "2023-06-01", filters.Date)
			return 4, nil
		}

		router := gin.New()
		router.GET("/jira/issues/me/count", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssuesCount(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me/count?status=Resolved&date=2023-06-01", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"count":4`)
	})

	suite.T().Run("Service error", func(t *testing.T) {
		suite.mockService.getIssuesCountFunc = func(filters service.JiraIssueFilters) (int, error) {
			return 0, errors.New("jira configuration missing")
		}

		router := gin.New()
		router.GET("/jira/issues/me/count", func(c *gin.Context) {
			claims := &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
			}
			c.Set("auth_claims", claims)
			suite.handler.GetMyIssuesCount(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/jira/issues/me/count?status=Resolved", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadGateway, w.Code)
		assert.Contains(t, w.Body.String(), "jira search failed")
		assert.Contains(t, w.Body.String(), "jira configuration missing")
	})
}

// TestJiraHandlerTestSuite runs the test suite
func TestJiraHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(JiraHandlerTestSuite))
}
