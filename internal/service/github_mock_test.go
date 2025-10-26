package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestGetUserOpenPullRequests_FullFlow_WithMocks tests the complete flow with mocked auth service
func TestGetUserOpenPullRequests_FullFlow_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock GitHub API server
	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request structure
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/search/issues")
		assert.Contains(t, r.URL.RawQuery, "is%3Apr")
		assert.Contains(t, r.URL.RawQuery, "author%3A%40me")

		// Return mock PR data
		response := map[string]interface{}{
			"total_count": 2,
			"items": []map[string]interface{}{
				{
					"id":         int64(123456789),
					"number":     42,
					"title":      "Add new feature",
					"state":      "open",
					"created_at": "2025-01-01T12:00:00Z",
					"updated_at": "2025-01-02T12:00:00Z",
					"html_url":   mockGitHubServer.URL + "/owner/repo/pull/42",
					"draft":      false,
					"user": map[string]interface{}{
						"login":      "testuser",
						"id":         int64(12345),
						"avatar_url": "https://avatars.githubusercontent.com/u/12345",
					},
					"pull_request": map[string]interface{}{
						"url": mockGitHubServer.URL + "/repos/owner/repo/pulls/42",
					},
					"repository": map[string]interface{}{
						"name":      "test-repo",
						"full_name": "owner/test-repo",
						"private":   false,
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
				{
					"id":         int64(987654321),
					"number":     43,
					"title":      "Fix critical bug",
					"state":      "open",
					"created_at": "2025-01-03T12:00:00Z",
					"updated_at": "2025-01-04T12:00:00Z",
					"html_url":   mockGitHubServer.URL + "/owner/repo/pull/43",
					"draft":      true,
					"user": map[string]interface{}{
						"login":      "testuser",
						"id":         int64(12345),
						"avatar_url": "https://avatars.githubusercontent.com/u/12345",
					},
					"pull_request": map[string]interface{}{
						"url": mockGitHubServer.URL + "/repos/owner/repo/pulls/43",
					},
					"repository": map[string]interface{}{
						"name":      "another-repo",
						"full_name": "owner/another-repo",
						"private":   true,
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockGitHubServer.Close()

	// Create mock auth service
	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)

	// Set expectations
	mockAuthService.EXPECT().
		GetGitHubAccessTokenFromClaims(gomock.Any()).
		Return("test_github_token_123", nil).
		Times(1)

	// Create a GitHub client that points to our mock server
	envConfig := &auth.EnvironmentConfig{
		ClientID:          "test_client_id",
		ClientSecret:      "test_client_secret",
		EnterpriseBaseURL: mockGitHubServer.URL,
	}
	githubClient := auth.NewGitHubClient(envConfig)

	mockAuthService.EXPECT().
		GetGitHubClient("githubtools", "development").
		Return(githubClient, nil).
		Times(1)

	// Create GitHub service with mock auth
	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)

	claims := &auth.AuthClaims{
		UserID:      12345,
		Username:    "testuser",
		Email:       "test@example.com",
		Provider:    "githubtools",
		Environment: "development",
	}

	ctx := context.Background()

	// Execute the test
	result, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.PullRequests, 2)

	// Verify first PR
	pr1 := result.PullRequests[0]
	assert.Equal(t, int64(123456789), pr1.ID)
	assert.Equal(t, 42, pr1.Number)
	assert.Equal(t, "Add new feature", pr1.Title)
	assert.Equal(t, "open", pr1.State)
	assert.False(t, pr1.Draft)
	assert.Equal(t, "testuser", pr1.User.Login)
	assert.Equal(t, "test-repo", pr1.Repo.Name)
	assert.Equal(t, "owner", pr1.Repo.Owner)
	assert.False(t, pr1.Repo.Private)

	// Verify second PR
	pr2 := result.PullRequests[1]
	assert.Equal(t, int64(987654321), pr2.ID)
	assert.Equal(t, 43, pr2.Number)
	assert.Equal(t, "Fix critical bug", pr2.Title)
	assert.True(t, pr2.Draft)
	assert.Equal(t, "another-repo", pr2.Repo.Name)
	assert.True(t, pr2.Repo.Private)
}

// TestGetUserOpenPullRequests_ClosedState tests fetching closed PRs
func TestGetUserOpenPullRequests_ClosedState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "state%3Aclosed")

		response := map[string]interface{}{
			"total_count": 1,
			"items": []map[string]interface{}{
				{
					"id":         int64(111),
					"number":     10,
					"title":      "Closed PR",
					"state":      "closed",
					"created_at": "2025-01-01T12:00:00Z",
					"updated_at": "2025-01-05T12:00:00Z",
					"html_url":   mockGitHubServer.URL + "/owner/repo/pull/10",
					"draft":      false,
					"user": map[string]interface{}{
						"login":      "testuser",
						"id":         int64(12345),
						"avatar_url": "https://avatars.githubusercontent.com/u/12345",
					},
					"pull_request": map[string]interface{}{
						"url": mockGitHubServer.URL + "/repos/owner/repo/pulls/10",
					},
					"repository": map[string]interface{}{
						"name":      "test-repo",
						"full_name": "owner/test-repo",
						"private":   false,
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)

	envConfig := &auth.EnvironmentConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any(), gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{
		UserID:      12345,
		Provider:    "githubtools",
		Environment: "development",
	}

	result, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "closed", "created", "desc", 30, 1)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "closed", result.PullRequests[0].State)
}

// TestGetUserOpenPullRequests_EmptyResults tests when no PRs are found
func TestGetUserOpenPullRequests_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"total_count": 0,
			"items":       []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)

	envConfig := &auth.EnvironmentConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any(), gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 12345, Provider: "githubtools", Environment: "development"}

	result, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "open", "created", "desc", 30, 1)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.PullRequests)
}

// TestGetUserOpenPullRequests_TokenRetrievalFailure tests auth service failure
func TestGetUserOpenPullRequests_TokenRetrievalFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().
		GetGitHubAccessTokenFromClaims(gomock.Any()).
		Return("", fmt.Errorf("no valid session found")).
		Times(1)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 12345}

	result, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "open", "created", "desc", 30, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get GitHub access token")
}

// TestGetUserOpenPullRequests_GitHubClientRetrievalFailure tests client retrieval failure
func TestGetUserOpenPullRequests_GitHubClientRetrievalFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)
	mockAuthService.EXPECT().
		GetGitHubClient(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("client not found")).
		Times(1)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 12345, Provider: "invalid", Environment: "dev"}

	result, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "open", "created", "desc", 30, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get GitHub client")
}

// TestGetUserOpenPullRequests_GitHubAPIRateLimit tests rate limit handling
func TestGetUserOpenPullRequests_GitHubAPIRateLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		response := map[string]interface{}{
			"message": "API rate limit exceeded",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)

	envConfig := &auth.EnvironmentConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any(), gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 12345, Provider: "githubtools", Environment: "development"}

	result, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "open", "created", "desc", 30, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

// TestGetUserOpenPullRequests_DefaultParameterNormalization tests parameter defaults
func TestGetUserOpenPullRequests_DefaultParameterNormalization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var capturedQuery string
	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		response := map[string]interface{}{
			"total_count": 0,
			"items":       []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)

	envConfig := &auth.EnvironmentConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any(), gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 12345, Provider: "githubtools", Environment: "development"}

	// Call with empty parameters to test defaults
	_, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "", "", "", 0, 0)

	require.NoError(t, err)
	// Verify defaults were applied: state=open, sort=created, order=desc
	assert.Contains(t, capturedQuery, "state%3Aopen")
	assert.Contains(t, capturedQuery, "sort=created")
	assert.Contains(t, capturedQuery, "order=desc")
}

// TestGetUserOpenPullRequests_PaginationParameters tests pagination
func TestGetUserOpenPullRequests_PaginationParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var capturedQuery string
	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		response := map[string]interface{}{
			"total_count": 100,
			"items":       []interface{}{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)

	envConfig := &auth.EnvironmentConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any(), gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 12345, Provider: "githubtools", Environment: "development"}

	// Test with specific pagination
	_, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "open", "created", "desc", 50, 3)

	require.NoError(t, err)
	assert.Contains(t, capturedQuery, "per_page=50")
	assert.Contains(t, capturedQuery, "page=3")
}

// TestGetUserOpenPullRequests_PRDataParsing tests that all PR fields are correctly parsed
func TestGetUserOpenPullRequests_PRDataParsing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"total_count": 1,
			"items": []map[string]interface{}{
				{
					"id":         int64(999),
					"number":     99,
					"title":      "Comprehensive Test PR",
					"state":      "open",
					"created_at": "2025-01-10T10:00:00Z",
					"updated_at": "2025-01-11T10:00:00Z",
					"html_url":   "https://github.com/test/repo/pull/99",
					"draft":      true,
					"user": map[string]interface{}{
						"login":      "contributor",
						"id":         int64(54321),
						"avatar_url": "https://avatars.githubusercontent.com/u/54321",
					},
					"pull_request": map[string]interface{}{
						"url": "https://api.github.com/repos/test/repo/pulls/99",
					},
					"repository": map[string]interface{}{
						"name":      "comprehensive-repo",
						"full_name": "test/comprehensive-repo",
						"private":   true,
						"owner": map[string]interface{}{
							"login": "test",
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)

	envConfig := &auth.EnvironmentConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any(), gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 12345, Provider: "githubtools", Environment: "development"}

	result, err := githubService.GetUserOpenPullRequests(context.Background(), claims, "open", "created", "desc", 30, 1)

	require.NoError(t, err)
	require.Len(t, result.PullRequests, 1)

	pr := result.PullRequests[0]
	assert.Equal(t, int64(999), pr.ID)
	assert.Equal(t, 99, pr.Number)
	assert.Equal(t, "Comprehensive Test PR", pr.Title)
	assert.Equal(t, "open", pr.State)
	assert.True(t, pr.Draft)
	assert.Equal(t, "contributor", pr.User.Login)
	assert.Equal(t, int64(54321), pr.User.ID)
	assert.Equal(t, "comprehensive-repo", pr.Repo.Name)
	assert.Equal(t, "test/comprehensive-repo", pr.Repo.FullName)
	assert.Equal(t, "test", pr.Repo.Owner)
	assert.True(t, pr.Repo.Private)
}
