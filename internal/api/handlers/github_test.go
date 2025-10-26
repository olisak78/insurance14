package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/auth"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// GitHubHandlerTestSuite defines the test suite for GitHubHandler
type GitHubHandlerTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupTest sets up the test suite
func (suite *GitHubHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
}

// TearDownTest cleans up after each test
func (suite *GitHubHandlerTestSuite) TearDownTest() {
	// Cleanup
}

// MockGitHubService is a mock implementation for testing
type MockGitHubService struct {
	Response *service.PullRequestsResponse
	Error    error
}

func (m *MockGitHubService) GetUserOpenPullRequests(ctx context.Context, claims *auth.AuthClaims, state, sort, direction string, perPage, page int) (*service.PullRequestsResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Response, nil
}

func (m *MockGitHubService) GetUserTotalContributions(ctx context.Context, claims *auth.AuthClaims, period string) (*service.TotalContributionsResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	// Return a mock response
	return &service.TotalContributionsResponse{
		TotalContributions: 1234,
		Period:             period,
		From:               "2024-10-16T00:00:00Z",
		To:                 "2025-10-16T23:59:59Z",
	}, nil
}

// TestGetMyPullRequests_Success tests successful PR retrieval
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_Success() {
	// Create mock service with successful response
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{
				{
					ID:        123456789,
					Number:    42,
					Title:     "Add new feature",
					State:     "open",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					HTMLURL:   "https://github.com/owner/repo/pull/42",
					Draft:     false,
					User: service.GitHubUser{
						Login:     "testuser",
						ID:        12345,
						AvatarURL: "https://avatars.githubusercontent.com/u/12345",
					},
					Repo: service.Repository{
						Name:     "test-repo",
						FullName: "owner/test-repo",
						Owner:    "owner",
						Private:  false,
					},
				},
			},
			Total: 1,
		},
	}

	// Create handler with mock service
	handler := handlers.NewGitHubHandler(mockService)

	// Setup route
	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		// Set mock auth claims in context
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	// Make request
	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.PullRequestsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, response.Total)
	assert.Len(suite.T(), response.PullRequests, 1)
	assert.Equal(suite.T(), 42, response.PullRequests[0].Number)
	assert.Equal(suite.T(), "Add new feature", response.PullRequests[0].Title)
}

// TestGetMyPullRequests_Unauthorized tests missing authentication
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_Unauthorized() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	// Setup route without auth claims
	suite.router.GET("/github/pull-requests", handler.GetMyPullRequests)

	// Make request
	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Authentication required")
}

// TestGetMyPullRequests_ServiceError tests service error handling
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_ServiceError() {
	mockService := &MockGitHubService{
		Error: fmt.Errorf("failed to fetch pull requests"),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Failed to fetch pull requests")
}

// TestGetMyPullRequests_RateLimitError tests rate limit error handling
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_RateLimitError() {
	mockService := &MockGitHubService{
		Error: apperrors.ErrGitHubAPIRateLimitExceeded,
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "rate limit exceeded")
}

// TestGetMyPullRequests_WithQueryParameters tests query parameter handling
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_WithQueryParameters() {
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{},
			Total:        0,
		},
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	testCases := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{"ValidOpenState", "?state=open", http.StatusOK},
		{"ValidClosedState", "?state=closed", http.StatusOK},
		{"ValidAllState", "?state=all", http.StatusOK},
		{"InvalidState", "?state=invalid", http.StatusBadRequest},
		{"ValidSort", "?sort=created", http.StatusOK},
		{"ValidUpdatedSort", "?sort=updated", http.StatusOK},
		{"InvalidSort", "?sort=invalid", http.StatusBadRequest},
		{"ValidDirection", "?direction=asc", http.StatusOK},
		{"InvalidDirection", "?direction=invalid", http.StatusBadRequest},
		{"ValidPerPage", "?per_page=50", http.StatusOK},
		{"ValidPage", "?page=2", http.StatusOK},
		{"MultipleParams", "?state=closed&sort=updated&direction=asc&per_page=50&page=2", http.StatusOK},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests"+tc.queryParams, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

// TestGetMyPullRequests_InvalidClaims tests invalid claims type
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_InvalidClaims() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		// Set invalid claims type
		c.Set("auth_claims", "invalid")
		handler.GetMyPullRequests(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Invalid authentication claims")
}

// TestGetMyPullRequests_EmptyResponse tests empty PR list
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_EmptyResponse() {
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{},
			Total:        0,
		},
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.PullRequestsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, response.Total)
	assert.Empty(suite.T(), response.PullRequests)
}

// TestGetMyPullRequests_MultiplePRs tests response with multiple PRs
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_MultiplePRs() {
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{
				{
					ID:     1,
					Number: 1,
					Title:  "PR 1",
					State:  "open",
				},
				{
					ID:     2,
					Number: 2,
					Title:  "PR 2",
					State:  "open",
				},
				{
					ID:     3,
					Number: 3,
					Title:  "PR 3",
					State:  "open",
				},
			},
			Total: 3,
		},
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.PullRequestsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, response.Total)
	assert.Len(suite.T(), response.PullRequests, 3)
}

// TestGetMyPullRequests_DefaultParameters tests default parameter values
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_DefaultParameters() {
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{},
			Total:        0,
		},
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	// Request without any query parameters
	req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestGetMyPullRequests_InvalidPerPage tests invalid per_page values
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_InvalidPerPage() {
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{},
			Total:        0,
		},
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	testCases := []string{
		"?per_page=abc", // Non-numeric
		"?per_page=-1",  // Negative
		"?per_page=0",   // Zero
		"?per_page=200", // Too large
		"?per_page=",    // Empty
	}

	for _, queryParam := range testCases {
		req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests"+queryParam, nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Should still succeed with defaults
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	}
}

// TestGetMyPullRequests_InvalidPage tests invalid page values
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_InvalidPage() {
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{},
			Total:        0,
		},
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/pull-requests", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetMyPullRequests(c)
	})

	testCases := []string{
		"?page=abc", // Non-numeric
		"?page=-1",  // Negative
		"?page=0",   // Zero
		"?page=",    // Empty
	}

	for _, queryParam := range testCases {
		req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests"+queryParam, nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Should still succeed with defaults
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	}
}

// TestNewGitHubHandler tests handler creation
func (suite *GitHubHandlerTestSuite) TestNewGitHubHandler() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	assert.NotNil(suite.T(), handler)
}

// TestGetMyPullRequests_DifferentProviders tests different provider values
func (suite *GitHubHandlerTestSuite) TestGetMyPullRequests_DifferentProviders() {
	mockService := &MockGitHubService{
		Response: &service.PullRequestsResponse{
			PullRequests: []service.PullRequest{},
			Total:        0,
		},
	}
	handler := handlers.NewGitHubHandler(mockService)

	providers := []string{"githubtools", "githubwdf"}

	for _, provider := range providers {
		suite.T().Run(provider, func(t *testing.T) {
			router := gin.New()
			router.GET("/github/pull-requests", func(c *gin.Context) {
				claims := &auth.AuthClaims{
					UserID:      12345,
					Username:    "testuser",
					Email:       "test@example.com",
					Provider:    provider,
					Environment: "development",
				}
				c.Set("auth_claims", claims)
				handler.GetMyPullRequests(c)
			})

			req, _ := http.NewRequest(http.MethodGet, "/github/pull-requests", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestGetUserTotalContributions_Success tests successful contribution retrieval
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_Success() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetUserTotalContributions(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/contributions?period=30d", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.TotalContributionsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1234, response.TotalContributions)
	assert.Equal(suite.T(), "30d", response.Period)
	assert.NotEmpty(suite.T(), response.From)
	assert.NotEmpty(suite.T(), response.To)
}

// TestGetUserTotalContributions_DefaultPeriod tests default period handling
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_DefaultPeriod() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetUserTotalContributions(c)
	})

	// Request without period parameter (should default to 365d)
	req, _ := http.NewRequest(http.MethodGet, "/github/contributions", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.TotalContributionsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1234, response.TotalContributions)
}

// TestGetUserTotalContributions_Unauthorized tests missing authentication
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_Unauthorized() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	// Setup route without auth claims
	suite.router.GET("/github/contributions", handler.GetUserTotalContributions)

	req, _ := http.NewRequest(http.MethodGet, "/github/contributions", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Authentication required")
}

// TestGetUserTotalContributions_InvalidClaims tests invalid claims type
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_InvalidClaims() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		// Set invalid claims type
		c.Set("auth_claims", "invalid")
		handler.GetUserTotalContributions(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/contributions", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Invalid authentication claims")
}

// TestGetUserTotalContributions_InvalidPeriod tests invalid period format
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_InvalidPeriod() {
	mockService := &MockGitHubService{
		Error: fmt.Errorf("invalid period format: period must be in format '<number>d' (e.g., '30d', '90d', '365d')"),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetUserTotalContributions(c)
	})

	testCases := []string{"30", "abc", "30days", "-30d", "0d"}

	for _, period := range testCases {
		suite.T().Run(period, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/github/contributions?period="+period, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["error"], "invalid period format")
		})
	}
}

// TestGetUserTotalContributions_RateLimitError tests rate limit error handling
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_RateLimitError() {
	mockService := &MockGitHubService{
		Error: apperrors.ErrGitHubAPIRateLimitExceeded,
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetUserTotalContributions(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/contributions?period=30d", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "rate limit exceeded")
}

// TestGetUserTotalContributions_ServiceError tests service error handling
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_ServiceError() {
	mockService := &MockGitHubService{
		Error: fmt.Errorf("failed to fetch contributions"),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetUserTotalContributions(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/contributions?period=30d", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Failed to fetch contributions")
}

// TestGetUserTotalContributions_ValidPeriods tests various valid period values
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_ValidPeriods() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "development",
		}
		c.Set("auth_claims", claims)
		handler.GetUserTotalContributions(c)
	})

	validPeriods := []string{"7d", "30d", "90d", "180d", "365d"}

	for _, period := range validPeriods {
		suite.T().Run(period, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/github/contributions?period="+period, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response service.TotalContributionsResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, 1234, response.TotalContributions)
			assert.Equal(t, period, response.Period)
		})
	}
}

// TestGetUserTotalContributions_DifferentProviders tests different provider values
func (suite *GitHubHandlerTestSuite) TestGetUserTotalContributions_DifferentProviders() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	providers := []string{"githubtools", "githubwdf"}

	for _, provider := range providers {
		suite.T().Run(provider, func(t *testing.T) {
			router := gin.New()
			router.GET("/github/contributions", func(c *gin.Context) {
				claims := &auth.AuthClaims{
					UserID:      12345,
					Username:    "testuser",
					Email:       "test@example.com",
					Provider:    provider,
					Environment: "development",
				}
				c.Set("auth_claims", claims)
				handler.GetUserTotalContributions(c)
			})

			req, _ := http.NewRequest(http.MethodGet, "/github/contributions?period=30d", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// Run the test suite
func TestGitHubHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(GitHubHandlerTestSuite))
}
