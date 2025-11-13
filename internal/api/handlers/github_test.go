package handlers_test

import (
	"bytes"
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

func (m *MockGitHubService) GetContributionsHeatmap(ctx context.Context, claims *auth.AuthClaims, period string) (*service.ContributionsHeatmapResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	// Return a mock response
	return &service.ContributionsHeatmapResponse{
		TotalContributions: 1234,
		Weeks: []service.ContributionWeek{
			{
				FirstDay: "2024-10-27",
				ContributionDays: []service.ContributionDay{
					{
						Date:              "2024-10-27",
						ContributionCount: 5,
						ContributionLevel: "SECOND_QUARTILE",
						Color:             "#40c463",
					},
					{
						Date:              "2024-10-28",
						ContributionCount: 10,
						ContributionLevel: "THIRD_QUARTILE",
						Color:             "#30a14e",
					},
				},
			},
		},
		From: "2024-10-30T00:00:00Z",
		To:   "2025-10-30T23:59:59Z",
	}, nil
}

func (m *MockGitHubService) GetAveragePRMergeTime(ctx context.Context, claims *auth.AuthClaims, period string) (*service.AveragePRMergeTimeResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	// Return a mock response
	return &service.AveragePRMergeTimeResponse{
		AveragePRMergeTimeHours: 24.5,
		PRCount:                 15,
		Period:                  period,
		From:                    "2024-10-03T00:00:00Z",
		To:                      "2024-11-02T23:59:59Z",
		TimeSeries: []service.PRMergeTimeDataPoint{
			{
				WeekStart:    "2024-10-26",
				WeekEnd:      "2024-11-02",
				AverageHours: 18.5,
				PRCount:      3,
			},
			{
				WeekStart:    "2024-10-19",
				WeekEnd:      "2024-10-26",
				AverageHours: 22.0,
				PRCount:      2,
			},
			{
				WeekStart:    "2024-10-12",
				WeekEnd:      "2024-10-19",
				AverageHours: 30.0,
				PRCount:      5,
			},
			{
				WeekStart:    "2024-10-05",
				WeekEnd:      "2024-10-12",
				AverageHours: 25.5,
				PRCount:      5,
			},
		},
	}, nil
}

func (m *MockGitHubService) GetUserPRReviewComments(ctx context.Context, claims *auth.AuthClaims, period string) (*service.PRReviewCommentsResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	// Return a mock response
	return &service.PRReviewCommentsResponse{
		TotalComments: 42,
		Period:        period,
		From:          "2024-10-03T00:00:00Z",
		To:            "2024-11-02T23:59:59Z",
	}, nil
}

func (m *MockGitHubService) GetGitHubAsset(ctx context.Context, claims *auth.AuthClaims, assetURL string) ([]byte, string, error) {
	if m.Error != nil {
		return nil, "", m.Error
	}
	// Return a mock response
	return []byte("mock asset content"), "application/octet-stream", nil
}

func (m *MockGitHubService) GetRepositoryContent(ctx context.Context, claims *auth.AuthClaims, owner, repo, path, ref string) (interface{}, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	// Return a mock response
	return map[string]interface{}{
		"name":    "example.md",
		"path":    path,
		"content": "bW9jayBjb250ZW50", // base64 encoded "mock content"
		"sha":     "abc123",
	}, nil
}

func (m *MockGitHubService) UpdateRepositoryFile(ctx context.Context, claims *auth.AuthClaims, owner, repo, path, message, content, sha, branch string) (interface{}, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	// Return a mock response
	return map[string]interface{}{
		"commit": map[string]interface{}{
			"sha":     "def456",
			"message": message,
		},
		"content": map[string]interface{}{
			"name": path,
			"sha":  "ghi789",
		},
	}, nil
}

// Implement UpdatePullRequestState to satisfy service.GitHubService
func (m *MockGitHubService) UpdatePullRequestState(ctx context.Context, claims *auth.AuthClaims, owner, repo string, prNumber int, state string) (*service.PullRequest, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	if state != "open" && state != "closed" {
		return nil, apperrors.ErrInvalidStatus
	}
	return &service.PullRequest{
		ID:     1,
		Number: prNumber,
		Title:  "Mock PR",
	State:  state,
		User:   service.GitHubUser{Login: "testuser", ID: 12345, AvatarURL: "https://avatars.githubusercontent.com/u/12345"},
		Repo:   service.Repository{Name: repo, FullName: owner + "/" + repo, Owner: owner, Private: false},
	}, nil
}

// Implement ClosePullRequest to satisfy service.GitHubServiceInterface
func (m *MockGitHubService) ClosePullRequest(ctx context.Context, claims *auth.AuthClaims, owner, repo string, prNumber int, deleteBranch bool) (*service.PullRequest, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return &service.PullRequest{
		ID:     1,
		Number: prNumber,
		Title:  "Closed PR",
		State:  "closed",
		User:   service.GitHubUser{Login: "testuser", ID: 12345, AvatarURL: "https://avatars.githubusercontent.com/u/12345"},
		Repo:   service.Repository{Name: repo, FullName: owner + "/" + repo, Owner: owner, Private: false},
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
					UserID:   12345,
					Username: "testuser",
					Email:    "test@example.com",
					Provider: provider,
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
		Error: fmt.Errorf("%w: period must be in format '<number>d' (e.g., '30d', '90d', '365d')", apperrors.ErrInvalidPeriodFormat),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/contributions", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
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
					UserID:   12345,
					Username: "testuser",
					Email:    "test@example.com",
					Provider: provider,
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

// TestGetContributionsHeatmap_Success tests successful heatmap retrieval
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_Success() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetContributionsHeatmap(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/githubtools/heatmap", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.ContributionsHeatmapResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1234, response.TotalContributions)
	assert.Equal(suite.T(), "2024-10-30T00:00:00Z", response.From)
	assert.Equal(suite.T(), "2025-10-30T23:59:59Z", response.To)
	assert.Len(suite.T(), response.Weeks, 1)
	assert.Equal(suite.T(), "2024-10-27", response.Weeks[0].FirstDay)
	assert.Len(suite.T(), response.Weeks[0].ContributionDays, 2)
}

// TestGetContributionsHeatmap_WithPeriod tests heatmap with period parameter
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_WithPeriod() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetContributionsHeatmap(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/githubtools/heatmap?period=90d", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.ContributionsHeatmapResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
}

// TestGetContributionsHeatmap_NoAuthClaims tests missing auth claims
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_NoAuthClaims() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", handler.GetContributionsHeatmap)

	req, _ := http.NewRequest(http.MethodGet, "/github/githubtools/heatmap", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Authentication required")
}

// TestGetContributionsHeatmap_ProviderMismatch tests provider mismatch
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_ProviderMismatch() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools", // User is authenticated with githubtools
		}
		c.Set("auth_claims", claims)
		handler.GetContributionsHeatmap(c)
	})

	// But requesting data for githubwdf
	req, _ := http.NewRequest(http.MethodGet, "/github/githubwdf/heatmap", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "does not match")
}

// TestGetContributionsHeatmap_ServiceError tests service error handling
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_ServiceError() {
	mockService := &MockGitHubService{
		Error: fmt.Errorf("service error"),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetContributionsHeatmap(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/githubtools/heatmap", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)
}

// TestGetContributionsHeatmap_RateLimitExceeded tests rate limit error
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_RateLimitExceeded() {
	mockService := &MockGitHubService{
		Error: apperrors.ErrGitHubAPIRateLimitExceeded,
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetContributionsHeatmap(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/githubtools/heatmap", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
}

// TestGetContributionsHeatmap_InvalidPeriod tests invalid period format
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_InvalidPeriod() {
	mockService := &MockGitHubService{
		Error: fmt.Errorf("%w: period must be in format '<number>d'", apperrors.ErrInvalidPeriodFormat),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetContributionsHeatmap(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/githubtools/heatmap?period=invalid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestGetContributionsHeatmap_ProviderNotConfigured tests provider not configured error
func (suite *GitHubHandlerTestSuite) TestGetContributionsHeatmap_ProviderNotConfigured() {
	mockService := &MockGitHubService{
		Error: fmt.Errorf("%w: provider 'invalid'. Please check available providers in auth.yaml", apperrors.ErrProviderNotConfigured),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/:provider/heatmap", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "invalid",
		}
		c.Set("auth_claims", claims)
		handler.GetContributionsHeatmap(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/invalid/heatmap", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "is not configured")
}

// TestGetAveragePRMergeTime_Success tests successful average PR merge time retrieval
func (suite *GitHubHandlerTestSuite) TestGetAveragePRMergeTime_Success() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/average-pr-time", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetAveragePRMergeTime(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/average-pr-time", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.AveragePRMergeTimeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 24.5, response.AveragePRMergeTimeHours)
	assert.Equal(suite.T(), 15, response.PRCount)
	assert.Equal(suite.T(), "2024-10-03T00:00:00Z", response.From)
	assert.Equal(suite.T(), "2024-11-02T23:59:59Z", response.To)
	assert.Len(suite.T(), response.TimeSeries, 4)
}

// TestGetAveragePRMergeTime_WithPeriod tests average PR time with period parameter
func (suite *GitHubHandlerTestSuite) TestGetAveragePRMergeTime_WithPeriod() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/average-pr-time", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetAveragePRMergeTime(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/average-pr-time?period=90d", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.AveragePRMergeTimeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "90d", response.Period)
}

// TestGetAveragePRMergeTime_NoAuthClaims tests missing auth claims
func (suite *GitHubHandlerTestSuite) TestGetAveragePRMergeTime_NoAuthClaims() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/average-pr-time", handler.GetAveragePRMergeTime)

	req, _ := http.NewRequest(http.MethodGet, "/github/average-pr-time", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Authentication required", response["error"])
}

// TestGetAveragePRMergeTime_ServiceError tests service error handling
func (suite *GitHubHandlerTestSuite) TestGetAveragePRMergeTime_ServiceError() {
	mockService := &MockGitHubService{
		Error: fmt.Errorf("GitHub API error"),
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/average-pr-time", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetAveragePRMergeTime(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/average-pr-time", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadGateway, w.Code)
}

// TestGetAveragePRMergeTime_RateLimitExceeded tests rate limit error
func (suite *GitHubHandlerTestSuite) TestGetAveragePRMergeTime_RateLimitExceeded() {
	mockService := &MockGitHubService{
		Error: apperrors.ErrGitHubAPIRateLimitExceeded,
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/average-pr-time", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetAveragePRMergeTime(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/average-pr-time", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
}

// TestGetAveragePRMergeTime_InvalidPeriod tests invalid period format
func (suite *GitHubHandlerTestSuite) TestGetAveragePRMergeTime_InvalidPeriod() {
	mockService := &MockGitHubService{
		Error: apperrors.ErrInvalidPeriodFormat,
	}
	handler := handlers.NewGitHubHandler(mockService)

	suite.router.GET("/github/average-pr-time", func(c *gin.Context) {
		claims := &auth.AuthClaims{
			UserID:   12345,
			Username: "testuser",
			Email:    "test@example.com",
			Provider: "githubtools",
		}
		c.Set("auth_claims", claims)
		handler.GetAveragePRMergeTime(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/github/average-pr-time?period=invalid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

 // ClosePullRequest handler tests

func (suite *GitHubHandlerTestSuite) TestClosePR_Success() {
	mockService := &MockGitHubService{}
	handler := handlers.NewGitHubHandler(mockService)

	// Route for ClosePullRequest
	suite.router.PATCH("/github/pull-requests/:pr_number", func(c *gin.Context) {
		claims := &auth.AuthClaims{UserID: 12345, Provider: "githubtools"}
		c.Set("auth_claims", claims)
		handler.ClosePullRequest(c)
	})

	// Prepare request
	body := map[string]interface{}{
		"owner":         "owner",
		"repo":          "repo",
		"delete_branch": false,
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPatch, "/github/pull-requests/42", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var pr service.PullRequest
	err := json.Unmarshal(w.Body.Bytes(), &pr)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "closed", pr.State)
	assert.Equal(suite.T(), 42, pr.Number)
	assert.Equal(suite.T(), "owner/repo", pr.Repo.FullName)
}

func (suite *GitHubHandlerTestSuite) TestClosePR_NotFound() {
	mockService := &MockGitHubService{Error: apperrors.NewNotFoundError("pull request")}
	handler := handlers.NewGitHubHandler(mockService)

	// Route for ClosePullRequest
	suite.router.PATCH("/github/pull-requests/:pr_number", func(c *gin.Context) {
		claims := &auth.AuthClaims{UserID: 12345, Provider: "githubtools"}
		c.Set("auth_claims", claims)
		handler.ClosePullRequest(c)
	})

	// Prepare request
	body := map[string]interface{}{
		"owner":         "owner",
		"repo":          "repo",
		"delete_branch": false,
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPatch, "/github/pull-requests/99", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// Run the test suite
func TestGitHubHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(GitHubHandlerTestSuite))
}
