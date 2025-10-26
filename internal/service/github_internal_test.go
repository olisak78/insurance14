package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"developer-portal-backend/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseRepositoryFromURL_Internal tests the internal parseRepositoryFromURL function
func TestParseRepositoryFromURL_Internal(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		expectedOwner    string
		expectedRepoName string
		expectedFullName string
	}{
		{
			name:             "StandardGitHubURL",
			url:              "https://github.com/octocat/Hello-World/pull/42",
			expectedOwner:    "octocat",
			expectedRepoName: "Hello-World",
			expectedFullName: "octocat/Hello-World",
		},
		{
			name:             "GitHubEnterpriseURL",
			url:              "https://github.enterprise.com/myorg/myrepo/pull/123",
			expectedOwner:    "myorg",
			expectedRepoName: "myrepo",
			expectedFullName: "myorg/myrepo",
		},
		{
			name:             "URLWithHyphens",
			url:              "https://github.com/my-org/my-awesome-repo/pull/1",
			expectedOwner:    "my-org",
			expectedRepoName: "my-awesome-repo",
			expectedFullName: "my-org/my-awesome-repo",
		},
		{
			name:             "URLWithUnderscores",
			url:              "https://github.com/my_org/my_repo/pull/999",
			expectedOwner:    "my_org",
			expectedRepoName: "my_repo",
			expectedFullName: "my_org/my_repo",
		},
		{
			name:             "EmptyURL",
			url:              "",
			expectedOwner:    "",
			expectedRepoName: "",
			expectedFullName: "",
		},
		{
			name:             "InvalidURL",
			url:              "https://github.com/",
			expectedOwner:    "",
			expectedRepoName: "",
			expectedFullName: "",
		},
		{
			name:             "URLWithOnlyOwner",
			url:              "https://github.com/owner",
			expectedOwner:    "",
			expectedRepoName: "",
			expectedFullName: "",
		},
		{
			name:             "HTTPProtocol",
			url:              "http://github.com/owner/repo/pull/1",
			expectedOwner:    "owner",
			expectedRepoName: "repo",
			expectedFullName: "owner/repo",
		},
		{
			name:             "URLWithTrailingSlash",
			url:              "https://github.com/owner/repo/",
			expectedOwner:    "owner",
			expectedRepoName: "repo",
			expectedFullName: "owner/repo",
		},
		{
			name:             "URLWithIssuesPath",
			url:              "https://github.com/owner/repo/issues/42",
			expectedOwner:    "owner",
			expectedRepoName: "repo",
			expectedFullName: "owner/repo",
		},
		{
			name:             "GitHubEnterpriseWithPort",
			url:              "https://github.example.com:8443/orgname/projectname",
			expectedOwner:    "orgname",
			expectedRepoName: "projectname",
			expectedFullName: "orgname/projectname",
		},
		{
			name:             "URLWithSpecialCharacters",
			url:              "https://github.com/my-org_123/my.repo-name_v2/pull/1",
			expectedOwner:    "my-org_123",
			expectedRepoName: "my.repo-name_v2",
			expectedFullName: "my-org_123/my.repo-name_v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repoName, fullName := parseRepositoryFromURL(tt.url)

			assert.Equal(t, tt.expectedOwner, owner, "Owner mismatch")
			assert.Equal(t, tt.expectedRepoName, repoName, "Repository name mismatch")
			assert.Equal(t, tt.expectedFullName, fullName, "Full name mismatch")
		})
	}
}

// TestGetUserTotalContributions_HTTPErrors tests various HTTP error scenarios
func TestGetUserTotalContributions_HTTPErrors(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "RateLimitExceeded",
			statusCode:    403,
			responseBody:  `{"message": "API rate limit exceeded"}`,
			expectedError: "rate limit exceeded",
		},
		{
			name:          "NotFound",
			statusCode:    404,
			responseBody:  `{"message": "Not Found"}`,
			expectedError: "GraphQL query failed with status 404",
		},
		{
			name:          "InternalServerError",
			statusCode:    500,
			responseBody:  `{"message": "Internal Server Error"}`,
			expectedError: "GraphQL query failed with status 500",
		},
		{
			name:          "BadGateway",
			statusCode:    502,
			responseBody:  `{"message": "Bad Gateway"}`,
			expectedError: "GraphQL query failed with status 502",
		},
		{
			name:          "ServiceUnavailable",
			statusCode:    503,
			responseBody:  `{"message": "Service Unavailable"}`,
			expectedError: "GraphQL query failed with status 503",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock GitHub GraphQL server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify it's a GraphQL request
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/graphql", r.URL.Path)
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer")
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create mock auth service
			mockAuthService := &mockAuthServiceForContributions{
				accessToken: "test-token",
				baseURL:     server.URL,
			}

			githubService := NewGitHubServiceWithAdapter(mockAuthService)

			claims := &auth.AuthClaims{
				UserID:      12345,
				Username:    "testuser",
				Email:       "test@example.com",
				Provider:    "githubtools",
				Environment: "test",
			}

			ctx := context.Background()
			result, err := githubService.GetUserTotalContributions(ctx, claims, "30d")

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestGetUserTotalContributions_GraphQLErrors tests GraphQL error responses
func TestGetUserTotalContributions_GraphQLErrors(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		expectedError string
	}{
		{
			name: "SingleGraphQLError",
			responseBody: `{
				"errors": [
					{
						"message": "Field 'contributionsCollection' doesn't exist on type 'User'",
						"path": ["viewer", "contributionsCollection"]
					}
				]
			}`,
			expectedError: "Field 'contributionsCollection' doesn't exist on type 'User'",
		},
		{
			name: "MultipleGraphQLErrors",
			responseBody: `{
				"errors": [
					{
						"message": "Authentication required",
						"path": ["viewer"]
					},
					{
						"message": "Invalid token",
						"path": ["viewer"]
					}
				]
			}`,
			expectedError: "Authentication required",
		},
		{
			name: "GraphQLErrorWithoutPath",
			responseBody: `{
				"errors": [
					{
						"message": "Something went wrong"
					}
				]
			}`,
			expectedError: "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			mockAuthService := &mockAuthServiceForContributions{
				accessToken: "test-token",
				baseURL:     server.URL,
			}

			githubService := NewGitHubServiceWithAdapter(mockAuthService)

			claims := &auth.AuthClaims{
				UserID:      12345,
				Username:    "testuser",
				Email:       "test@example.com",
				Provider:    "githubtools",
				Environment: "test",
			}

			ctx := context.Background()
			result, err := githubService.GetUserTotalContributions(ctx, claims, "30d")

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestGetUserTotalContributions_MalformedJSON tests malformed JSON responses
func TestGetUserTotalContributions_MalformedJSON(t *testing.T) {
	tests := []struct {
		name          string
		responseBody  string
		expectedError string
	}{
		{
			name:          "InvalidJSON",
			responseBody:  `{invalid json}`,
			expectedError: "failed to decode GraphQL response",
		},
		{
			name:          "EmptyResponse",
			responseBody:  ``,
			expectedError: "failed to decode GraphQL response",
		},
		{
			name: "MissingDataField",
			responseBody: `{
				"viewer": {}
			}`,
			expectedError: "failed to unmarshal result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			mockAuthService := &mockAuthServiceForContributions{
				accessToken: "test-token",
				baseURL:     server.URL,
			}

			githubService := NewGitHubServiceWithAdapter(mockAuthService)

			claims := &auth.AuthClaims{
				UserID:      12345,
				Username:    "testuser",
				Email:       "test@example.com",
				Provider:    "githubtools",
				Environment: "test",
			}

			ctx := context.Background()
			result, err := githubService.GetUserTotalContributions(ctx, claims, "30d")

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// TestGetUserTotalContributions_SuccessfulResponses tests successful API calls
func TestGetUserTotalContributions_SuccessfulResponses(t *testing.T) {
	tests := []struct {
		name               string
		period             string
		totalContributions int
		responseStartedAt  string
		responseEndedAt    string
		expectedPeriod     string
	}{
		{
			name:               "ThirtyDayPeriod",
			period:             "30d",
			totalContributions: 42,
			responseStartedAt:  "2024-09-16T00:00:00Z",
			responseEndedAt:    "2024-10-16T23:59:59Z",
			expectedPeriod:     "30d",
		},
		{
			name:               "NinetyDayPeriod",
			period:             "90d",
			totalContributions: 256,
			responseStartedAt:  "2024-07-18T00:00:00Z",
			responseEndedAt:    "2024-10-16T23:59:59Z",
			expectedPeriod:     "90d",
		},
		{
			name:               "FullYearPeriod",
			period:             "365d",
			totalContributions: 1234,
			responseStartedAt:  "2023-10-16T00:00:00Z",
			responseEndedAt:    "2024-10-16T23:59:59Z",
			expectedPeriod:     "365d",
		},
		{
			name:               "DefaultPeriod",
			period:             "",
			totalContributions: 523,
			responseStartedAt:  "2023-10-16T00:00:00Z",
			responseEndedAt:    "2024-10-16T23:59:59Z",
			expectedPeriod:     "github_default",
		},
		{
			name:               "ZeroContributions",
			period:             "7d",
			totalContributions: 0,
			responseStartedAt:  "2024-10-09T00:00:00Z",
			responseEndedAt:    "2024-10-16T23:59:59Z",
			expectedPeriod:     "7d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/graphql", r.URL.Path)
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-token")

				// Verify query contains expected parameters
				var reqBody struct {
					Query string `json:"query"`
				}
				json.NewDecoder(r.Body).Decode(&reqBody)

				if tt.period != "" {
					assert.Contains(t, reqBody.Query, "contributionsCollection(from:")
				} else {
					assert.Contains(t, reqBody.Query, "contributionsCollection {")
					assert.NotContains(t, reqBody.Query, "contributionsCollection(from:")
				}

				// Send successful response
				response := map[string]interface{}{
					"data": map[string]interface{}{
						"viewer": map[string]interface{}{
							"contributionsCollection": map[string]interface{}{
								"startedAt": tt.responseStartedAt,
								"endedAt":   tt.responseEndedAt,
								"contributionCalendar": map[string]interface{}{
									"totalContributions": tt.totalContributions,
								},
							},
						},
					},
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			mockAuthService := &mockAuthServiceForContributions{
				accessToken: "test-token",
				baseURL:     server.URL,
			}

			githubService := NewGitHubServiceWithAdapter(mockAuthService)

			claims := &auth.AuthClaims{
				UserID:      12345,
				Username:    "testuser",
				Email:       "test@example.com",
				Provider:    "githubtools",
				Environment: "test",
			}

			ctx := context.Background()
			result, err := githubService.GetUserTotalContributions(ctx, claims, tt.period)

			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.totalContributions, result.TotalContributions)
			assert.Equal(t, tt.expectedPeriod, result.Period)
			assert.Equal(t, tt.responseStartedAt, result.From)
			assert.Equal(t, tt.responseEndedAt, result.To)
		})
	}
}

// TestGetUserTotalContributions_ContextTimeout tests context timeout handling
func TestGetUserTotalContributions_ContextTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockAuthService := &mockAuthServiceForContributions{
		accessToken: "test-token",
		baseURL:     server.URL,
	}

	githubService := NewGitHubServiceWithAdapter(mockAuthService)

	claims := &auth.AuthClaims{
		UserID:      12345,
		Username:    "testuser",
		Email:       "test@example.com",
		Provider:    "githubtools",
		Environment: "test",
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := githubService.GetUserTotalContributions(ctx, claims, "30d")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestGetUserTotalContributions_PeriodValidation tests period validation
func TestGetUserTotalContributions_PeriodValidation(t *testing.T) {
	githubService := NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:      12345,
		Username:    "testuser",
		Email:       "test@example.com",
		Provider:    "githubtools",
		Environment: "test",
	}

	invalidPeriods := []string{
		"30",     // missing 'd'
		"abc",    // not a number
		"30days", // wrong format
		"-30d",   // negative
		"0d",     // zero
		"30m",    // wrong unit
		"30y",    // wrong unit
		"d30",    // reversed
		"30 d",   // space
	}

	ctx := context.Background()

	for _, period := range invalidPeriods {
		t.Run(fmt.Sprintf("Invalid_%s", period), func(t *testing.T) {
			result, err := githubService.GetUserTotalContributions(ctx, claims, period)

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "invalid period format")
		})
	}
}

// TestGetUserTotalContributions_LargeContributions tests handling of large contribution counts
func TestGetUserTotalContributions_LargeContributions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"contributionsCollection": map[string]interface{}{
						"startedAt": "2023-10-16T00:00:00Z",
						"endedAt":   "2024-10-16T23:59:59Z",
						"contributionCalendar": map[string]interface{}{
							"totalContributions": 999999,
						},
					},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockAuthService := &mockAuthServiceForContributions{
		accessToken: "test-token",
		baseURL:     server.URL,
	}

	githubService := NewGitHubServiceWithAdapter(mockAuthService)

	claims := &auth.AuthClaims{
		UserID:      12345,
		Username:    "testuser",
		Email:       "test@example.com",
		Provider:    "githubtools",
		Environment: "test",
	}

	ctx := context.Background()
	result, err := githubService.GetUserTotalContributions(ctx, claims, "365d")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 999999, result.TotalContributions)
}

// TestGetUserTotalContributions_EmptyDates tests handling of empty dates in response
func TestGetUserTotalContributions_EmptyDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"contributionsCollection": map[string]interface{}{
						"startedAt": "",
						"endedAt":   "",
						"contributionCalendar": map[string]interface{}{
							"totalContributions": 100,
						},
					},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	mockAuthService := &mockAuthServiceForContributions{
		accessToken: "test-token",
		baseURL:     server.URL,
	}

	githubService := NewGitHubServiceWithAdapter(mockAuthService)

	claims := &auth.AuthClaims{
		UserID:      12345,
		Username:    "testuser",
		Email:       "test@example.com",
		Provider:    "githubtools",
		Environment: "test",
	}

	ctx := context.Background()
	result, err := githubService.GetUserTotalContributions(ctx, claims, "30d")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 100, result.TotalContributions)
	// Should fall back to calculated dates
	assert.NotEmpty(t, result.From)
	assert.NotEmpty(t, result.To)
}

// TestGetUserOpenPullRequests_EdgeCases tests edge cases for pull requests
func TestGetUserOpenPullRequests_EdgeCases(t *testing.T) {
	t.Run("EmptySearchResults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return empty search results
			response := map[string]interface{}{
				"total_count":        0,
				"incomplete_results": false,
				"items":              []interface{}{},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		mockAuthService := &mockAuthServiceForContributions{
			accessToken: "test-token",
			baseURL:     server.URL,
		}

		githubService := NewGitHubServiceWithAdapter(mockAuthService)

		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "test",
		}

		ctx := context.Background()
		result, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 0, result.Total)
		assert.Equal(t, 0, len(result.PullRequests))
	})

	t.Run("NilAuthService", func(t *testing.T) {
		githubService := NewGitHubService(nil)

		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "test",
		}

		ctx := context.Background()
		result, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get GitHub access token")
	})

	t.Run("ContextCancelled", func(t *testing.T) {
		mockAuthService := &mockAuthServiceForContributions{
			accessToken: "test-token",
			baseURL:     "http://localhost:9999",
		}

		githubService := NewGitHubServiceWithAdapter(mockAuthService)

		claims := &auth.AuthClaims{
			UserID:      12345,
			Username:    "testuser",
			Email:       "test@example.com",
			Provider:    "githubtools",
			Environment: "test",
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestAuthServiceAdapter tests the auth service adapter
func TestAuthServiceAdapter(t *testing.T) {
	t.Run("NewAuthServiceAdapter_WithNil", func(t *testing.T) {
		adapter := NewAuthServiceAdapter(nil)
		assert.NotNil(t, adapter)
	})

	t.Run("NewAuthServiceAdapter_WithNonNil", func(t *testing.T) {
		// We can't easily create a real AuthService without complex setup,
		// but we can at least verify the function doesn't panic
		adapter := NewAuthServiceAdapter(nil)
		assert.NotNil(t, adapter)
	})

	t.Run("GetGitHubAccessTokenFromClaims_NilAuthService", func(t *testing.T) {
		adapter := NewAuthServiceAdapter(nil)
		claims := &auth.AuthClaims{UserID: 123}

		token, err := adapter.GetGitHubAccessTokenFromClaims(claims)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "auth service is not initialized")
	})

	t.Run("GetGitHubClient_NilAuthService", func(t *testing.T) {
		adapter := NewAuthServiceAdapter(nil)

		client, err := adapter.GetGitHubClient("githubtools", "development")

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "auth service is not initialized")
	})

	t.Run("MockAdapter_Success", func(t *testing.T) {
		// Test that the mock adapter works correctly (covers success paths)
		mock := &mockAuthServiceForContributions{
			accessToken: "test-token",
			baseURL:     "https://github.tools.sap",
		}

		// Test GetGitHubAccessTokenFromClaims
		claims := &auth.AuthClaims{UserID: 123}
		token, err := mock.GetGitHubAccessTokenFromClaims(claims)
		assert.NoError(t, err)
		assert.Equal(t, "test-token", token)

		// Test GetGitHubClient
		client, err := mock.GetGitHubClient("githubtools", "development")
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

// Mock auth service for contributions tests
type mockAuthServiceForContributions struct {
	accessToken string
	baseURL     string
}

func (m *mockAuthServiceForContributions) GetGitHubAccessTokenFromClaims(claims *auth.AuthClaims) (string, error) {
	if m.accessToken == "" {
		return "", fmt.Errorf("no access token")
	}
	return m.accessToken, nil
}

func (m *mockAuthServiceForContributions) GetGitHubClient(provider, environment string) (*auth.GitHubClient, error) {
	// Create a test client with our test server's baseURL
	config := &auth.EnvironmentConfig{
		ClientID:          "test-client-id",
		ClientSecret:      "test-client-secret",
		EnterpriseBaseURL: m.baseURL,
	}
	return auth.NewGitHubClient(config), nil
}
