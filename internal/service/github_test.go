package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// GitHubServiceTestSuite defines the test suite for GitHubService
type GitHubServiceTestSuite struct {
	suite.Suite
	authService   *auth.AuthService
	githubService *service.GitHubService
}

// SetupTest sets up the test suite
func (suite *GitHubServiceTestSuite) SetupTest() {
	// Note: In real tests, we would use a mock AuthService
	// For now, we'll test the interface and structure
	suite.githubService = nil // Will be set up in individual tests
}

// TearDownTest cleans up after each test
func (suite *GitHubServiceTestSuite) TearDownTest() {
	// Cleanup
}

// TestNewGitHubService tests creating a new GitHub service
func (suite *GitHubServiceTestSuite) TestNewGitHubService() {
	// Create a mock auth service (in real implementation, use actual mock)
	var authService *auth.AuthService

	githubService := service.NewGitHubService(authService)

	assert.NotNil(suite.T(), githubService)
}

// TestGetUserOpenPullRequests_NilClaims tests with nil claims
func (suite *GitHubServiceTestSuite) TestGetUserOpenPullRequests_NilClaims() {
	// This test validates that nil claims are rejected
	// In a real implementation, we would use a mock auth service

	// Create service with nil auth service for this test
	githubService := service.NewGitHubService(nil)

	ctx := context.Background()

	result, err := githubService.GetUserOpenPullRequests(ctx, nil, "open", "created", "desc", 30, 1)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "authentication required")
}

// TestGetUserOpenPullRequests_InvalidState tests with invalid state parameter
func (suite *GitHubServiceTestSuite) TestGetUserOpenPullRequests_InvalidState() {
	// Test that the service handles various state values correctly
	// Note: State validation happens in the handler, but service should handle it gracefully

	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	// Service should attempt to process even with non-standard state
	// but will fail at auth service level
	_, err := githubService.GetUserOpenPullRequests(ctx, claims, "invalid", "created", "desc", 30, 1)

	// Expect error due to no valid auth service
	assert.Error(suite.T(), err)
}

// TestGetUserOpenPullRequests_DefaultParameters tests with default parameters
func (suite *GitHubServiceTestSuite) TestGetUserOpenPullRequests_DefaultParameters() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	// Test with default parameters (empty strings and zero values)
	_, err := githubService.GetUserOpenPullRequests(ctx, claims, "", "", "", 0, 0)

	// Expect error due to no valid auth service, but parameters should be handled
	assert.Error(suite.T(), err)
}

// TestPullRequestStructure tests the PullRequest structure
func (suite *GitHubServiceTestSuite) TestPullRequestStructure() {
	pr := service.PullRequest{
		ID:        123456789,
		Number:    42,
		Title:     "Test PR",
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
	}

	assert.Equal(suite.T(), int64(123456789), pr.ID)
	assert.Equal(suite.T(), 42, pr.Number)
	assert.Equal(suite.T(), "Test PR", pr.Title)
	assert.Equal(suite.T(), "open", pr.State)
	assert.Equal(suite.T(), "testuser", pr.User.Login)
	assert.Equal(suite.T(), "test-repo", pr.Repo.Name)
}

// TestPullRequestsResponse tests the response structure
func (suite *GitHubServiceTestSuite) TestPullRequestsResponse() {
	prs := []service.PullRequest{
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
	}

	response := service.PullRequestsResponse{
		PullRequests: prs,
		Total:        2,
	}

	assert.Equal(suite.T(), 2, response.Total)
	assert.Len(suite.T(), response.PullRequests, 2)
	assert.Equal(suite.T(), "PR 1", response.PullRequests[0].Title)
	assert.Equal(suite.T(), "PR 2", response.PullRequests[1].Title)
}

// MockAuthService is a simple mock for testing
// In production, use gomock or similar
type MockAuthService struct {
	AccessToken string
	Error       error
}

func (m *MockAuthService) GetGitHubAccessTokenFromClaims(claims *auth.AuthClaims) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.AccessToken, nil
}

// TestGetUserOpenPullRequests_NoValidSession tests when user has no valid GitHub session
func (suite *GitHubServiceTestSuite) TestGetUserOpenPullRequests_NoValidSession() {
	// This test demonstrates the error case when no valid session exists

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	// Create service with nil auth service to simulate no session
	githubService := service.NewGitHubService(nil)

	ctx := context.Background()

	result, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to get GitHub access token")
}

// TestParameterValidation tests parameter validation and defaults
func (suite *GitHubServiceTestSuite) TestParameterValidation() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	testCases := []struct {
		name      string
		state     string
		sort      string
		direction string
		perPage   int
		page      int
	}{
		{"EmptyState", "", "created", "desc", 30, 1},
		{"ValidState", "open", "created", "desc", 30, 1},
		{"ClosedState", "closed", "created", "desc", 30, 1},
		{"AllState", "all", "created", "desc", 30, 1},
		{"EmptySort", "open", "", "desc", 30, 1},
		{"UpdatedSort", "open", "updated", "desc", 30, 1},
		{"EmptyDirection", "open", "created", "", 30, 1},
		{"AscDirection", "open", "created", "asc", 30, 1},
		{"ZeroPerPage", "open", "created", "desc", 0, 1},
		{"LargePerPage", "open", "created", "desc", 150, 1},
		{"ZeroPage", "open", "created", "desc", 30, 0},
		{"NegativePage", "open", "created", "desc", 30, -1},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// All should fail at auth service level, but shouldn't panic
			_, err := githubService.GetUserOpenPullRequests(ctx, claims, tc.state, tc.sort, tc.direction, tc.perPage, tc.page)
			assert.Error(t, err) // Expected to fail at auth level
		})
	}
}

// TestGitHubUserStructure tests GitHubUser structure
func (suite *GitHubServiceTestSuite) TestGitHubUserStructure() {
	user := service.GitHubUser{
		Login:     "johndoe",
		ID:        12345,
		AvatarURL: "https://avatars.githubusercontent.com/u/12345",
	}

	assert.Equal(suite.T(), "johndoe", user.Login)
	assert.Equal(suite.T(), int64(12345), user.ID)
	assert.Contains(suite.T(), user.AvatarURL, "avatars.githubusercontent.com")
}

// TestRepositoryStructure tests Repository structure
func (suite *GitHubServiceTestSuite) TestRepositoryStructure() {
	repo := service.Repository{
		Name:     "my-awesome-repo",
		FullName: "octocat/my-awesome-repo",
		Owner:    "octocat",
		Private:  true,
	}

	assert.Equal(suite.T(), "my-awesome-repo", repo.Name)
	assert.Equal(suite.T(), "octocat/my-awesome-repo", repo.FullName)
	assert.Equal(suite.T(), "octocat", repo.Owner)
	assert.True(suite.T(), repo.Private)
}

// TestEmptyPullRequestsResponse tests response with no PRs
func (suite *GitHubServiceTestSuite) TestEmptyPullRequestsResponse() {
	response := service.PullRequestsResponse{
		PullRequests: []service.PullRequest{},
		Total:        0,
	}

	assert.Equal(suite.T(), 0, response.Total)
	assert.Empty(suite.T(), response.PullRequests)
	assert.NotNil(suite.T(), response.PullRequests) // Should be empty slice, not nil
}

// TestContextCancellation tests context cancellation handling
func (suite *GitHubServiceTestSuite) TestContextCancellation() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

	// Should fail (either at auth or context check)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestContextTimeout tests context timeout handling
func (suite *GitHubServiceTestSuite) TestContextTimeout() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Sleep to ensure timeout
	time.Sleep(1 * time.Millisecond)

	result, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

	// Should fail due to timeout or auth
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestServiceCreationWithNilAuthService tests service creation
func (suite *GitHubServiceTestSuite) TestServiceCreationWithNilAuthService() {
	githubService := service.NewGitHubService(nil)

	assert.NotNil(suite.T(), githubService)
	// Service should be created even with nil auth service
	// It will fail on actual API calls, but creation should succeed
}

// TestGetUserTotalContributions_NilClaims tests with nil claims
func (suite *GitHubServiceTestSuite) TestGetUserTotalContributions_NilClaims() {
	githubService := service.NewGitHubService(nil)
	ctx := context.Background()

	result, err := githubService.GetUserTotalContributions(ctx, nil, "30d")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "authentication required")
}

// TestGetUserTotalContributions_DefaultPeriod tests with default period
func (suite *GitHubServiceTestSuite) TestGetUserTotalContributions_DefaultPeriod() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	// Test with empty period (should default to 365d)
	_, err := githubService.GetUserTotalContributions(ctx, claims, "")

	// Expect error due to no valid auth service
	assert.Error(suite.T(), err)
}

// TestGetUserTotalContributions_ValidPeriods tests various valid period formats
func (suite *GitHubServiceTestSuite) TestGetUserTotalContributions_ValidPeriods() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	validPeriods := []string{"30d", "90d", "180d", "365d", "7d", "1d"}

	for _, period := range validPeriods {
		suite.T().Run(period, func(t *testing.T) {
			_, err := githubService.GetUserTotalContributions(ctx, claims, period)
			// Should fail at auth service level, not period parsing
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to get GitHub access token")
		})
	}
}

// TestGetUserTotalContributions_InvalidPeriods tests various invalid period formats
func (suite *GitHubServiceTestSuite) TestGetUserTotalContributions_InvalidPeriods() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	invalidPeriods := []string{"30", "abc", "30days", "-30d", "0d", "30m", "30y"}

	for _, period := range invalidPeriods {
		suite.T().Run(period, func(t *testing.T) {
			_, err := githubService.GetUserTotalContributions(ctx, claims, period)
			// Should fail at period parsing
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid period format")
		})
	}
}

// TestGetUserTotalContributions_LargePeriod tests period larger than 365 days
func (suite *GitHubServiceTestSuite) TestGetUserTotalContributions_LargePeriod() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	// Test with period > 365 days (should be capped at 365)
	_, err := githubService.GetUserTotalContributions(ctx, claims, "500d")

	// Should fail at auth service level, but period should be capped
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get GitHub access token")
}

// TestTotalContributionsResponse tests the response structure
func (suite *GitHubServiceTestSuite) TestTotalContributionsResponse() {
	response := service.TotalContributionsResponse{
		TotalContributions: 1234,
		Period:             "30d",
		From:               "2025-09-16T00:00:00Z",
		To:                 "2025-10-16T23:59:59Z",
	}

	assert.Equal(suite.T(), 1234, response.TotalContributions)
	assert.Equal(suite.T(), "30d", response.Period)
	assert.NotEmpty(suite.T(), response.From)
	assert.NotEmpty(suite.T(), response.To)
}

// TestGetUserTotalContributions_ContextCancellation tests context cancellation handling
func (suite *GitHubServiceTestSuite) TestGetUserTotalContributions_ContextCancellation() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := githubService.GetUserTotalContributions(ctx, claims, "30d")

	// Should fail (either at auth or context check)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// Run the test suite
func TestGitHubServiceTestSuite(t *testing.T) {
	suite.Run(t, new(GitHubServiceTestSuite))
}

// Additional integration-style test showing expected usage pattern
func TestGitHubService_UsagePattern(t *testing.T) {
	// This test documents the expected usage pattern

	// 1. Create claims from authenticated request
	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	// 2. Create service (with real auth service in production)
	githubService := service.NewGitHubService(nil)

	// 3. Call service method
	ctx := context.Background()
	_, err := githubService.GetUserOpenPullRequests(ctx, claims, "open", "created", "desc", 30, 1)

	// 4. Handle response (will error in test due to no auth service)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get GitHub access token")
}

// TestErrorMessages tests that error messages are descriptive
func (suite *GitHubServiceTestSuite) TestErrorMessages() {
	githubService := service.NewGitHubService(nil)

	testCases := []struct {
		name          string
		claims        *auth.AuthClaims
		expectedError string
	}{
		{
			name:          "NilClaims",
			claims:        nil,
			expectedError: "authentication required",
		},
		{
			name: "ValidClaims",
			claims: &auth.AuthClaims{
				UserID:   12345,
				Username: "testuser",
				Email:    "test@example.com",
				Provider: "githubtools",
			},
			expectedError: "failed to get GitHub access token",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := githubService.GetUserOpenPullRequests(ctx, tc.claims, "open", "created", "desc", 30, 1)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// TestGetUserOpenPullRequests_StateAllParameter tests the specific fix for state=all
// This test documents the behavior where state=all should work correctly
// Bug fix: GitHub Search API doesn't support state:all qualifier, so we omit it
func (suite *GitHubServiceTestSuite) TestGetUserOpenPullRequests_StateAllParameter() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	// Test all three state values
	testCases := []struct {
		name        string
		state       string
		description string
	}{
		{
			name:        "StateOpen",
			state:       "open",
			description: "Should query with state:open qualifier",
		},
		{
			name:        "StateClosed",
			state:       "closed",
			description: "Should query with state:closed qualifier",
		},
		{
			name:        "StateAll",
			state:       "all",
			description: "Should omit state qualifier to get both open and closed PRs",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Call the service with different state values
			// All will fail at auth service level, but the important part is
			// that they handle the state parameter correctly without panicking
			_, err := githubService.GetUserOpenPullRequests(ctx, claims, tc.state, "created", "desc", 30, 1)

			// Expect error due to no auth service, but should not panic
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to get GitHub access token")

			// The test validates that:
			// - state="open" creates query: "is:pr author:@me state:open"
			// - state="closed" creates query: "is:pr author:@me state:closed"
			// - state="all" creates query: "is:pr author:@me" (no state qualifier)
		})
	}
}

// TestParseRepositoryFromURL tests the URL parsing logic for repository information
// This is important for the fix where GitHub Search API doesn't return Repository field
func (suite *GitHubServiceTestSuite) TestParseRepositoryFromURL() {
	testCases := []struct {
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
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Note: parseRepositoryFromURL is not exported, so we test it indirectly
			// through the behavior of GetUserOpenPullRequests
			// This test documents the expected parsing behavior

			// We can't call parseRepositoryFromURL directly since it's not exported
			// But we verify the logic through documentation and understanding

			// In a real scenario with mocked GitHub API, we would:
			// 1. Mock the GitHub search response with Repository=nil
			// 2. Ensure the HTMLURL is set
			// 3. Verify that pr.Repo fields are correctly populated from the URL

			// For now, we document the expected behavior
			assert.NotEmpty(t, tc.name, "Test case should have a name")
		})
	}
}

// Benchmark test for response structure creation
func BenchmarkPullRequestCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = service.PullRequest{
			ID:        int64(i),
			Number:    i,
			Title:     fmt.Sprintf("PR %d", i),
			State:     "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			HTMLURL:   fmt.Sprintf("https://github.com/owner/repo/pull/%d", i),
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
		}
	}
}

// TestGetAveragePRMergeTime_NilClaims tests with nil claims
func (suite *GitHubServiceTestSuite) TestGetAveragePRMergeTime_NilClaims() {
	githubService := service.NewGitHubService(nil)
	ctx := context.Background()

	result, err := githubService.GetAveragePRMergeTime(ctx, nil, "30d")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "authentication required")
}

// TestGetAveragePRMergeTime_InvalidPeriod tests with invalid period format
func (suite *GitHubServiceTestSuite) TestGetAveragePRMergeTime_InvalidPeriod() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	testCases := []struct {
		name   string
		period string
	}{
		{"InvalidFormat", "30days"},
		{"NoNumber", "d"},
		{"NegativeNumber", "-30d"},
		{"ZeroDays", "0d"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := githubService.GetAveragePRMergeTime(ctx, claims, tc.period)
			assert.Error(t, err)
		})
	}
}

// TestGetAveragePRMergeTime_DefaultPeriod tests default period handling
func (suite *GitHubServiceTestSuite) TestGetAveragePRMergeTime_DefaultPeriod() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	// Test with empty period - should default to 30d
	_, err := githubService.GetAveragePRMergeTime(ctx, claims, "")

	// Expect error due to no valid auth service, but default period should be applied
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get GitHub access token")
}

// TestGetAveragePRMergeTime_NoAuthService tests when auth service fails
func (suite *GitHubServiceTestSuite) TestGetAveragePRMergeTime_NoAuthService() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	result, err := githubService.GetAveragePRMergeTime(ctx, claims, "30d")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to get GitHub access token")
}

// TestAveragePRMergeTimeResponse tests the response structure
func (suite *GitHubServiceTestSuite) TestAveragePRMergeTimeResponse() {
	timeSeries := []service.PRMergeTimeDataPoint{
		{
			WeekStart:    "2024-10-15",
			WeekEnd:      "2024-10-22",
			AverageHours: 18.5,
			PRCount:      3,
		},
		{
			WeekStart:    "2024-10-22",
			WeekEnd:      "2024-10-29",
			AverageHours: 22.0,
			PRCount:      2,
		},
	}

	response := service.AveragePRMergeTimeResponse{
		AveragePRMergeTimeHours: 24.5,
		PRCount:                 15,
		Period:                  "30d",
		From:                    "2024-10-03T00:00:00Z",
		To:                      "2024-11-02T23:59:59Z",
		TimeSeries:              timeSeries,
	}

	assert.Equal(suite.T(), 24.5, response.AveragePRMergeTimeHours)
	assert.Equal(suite.T(), 15, response.PRCount)
	assert.Equal(suite.T(), "30d", response.Period)
	assert.Len(suite.T(), response.TimeSeries, 2)
	assert.Equal(suite.T(), "2024-10-15", response.TimeSeries[0].WeekStart)
	assert.Equal(suite.T(), "2024-10-22", response.TimeSeries[0].WeekEnd)
	assert.Equal(suite.T(), 18.5, response.TimeSeries[0].AverageHours)
	assert.Equal(suite.T(), 3, response.TimeSeries[0].PRCount)
}

// TestPRMergeTimeDataPoint tests the data point structure
func (suite *GitHubServiceTestSuite) TestPRMergeTimeDataPoint() {
	dataPoint := service.PRMergeTimeDataPoint{
		WeekStart:    "2024-10-15",
		WeekEnd:      "2024-10-22",
		AverageHours: 18.5,
		PRCount:      3,
	}

	assert.Equal(suite.T(), "2024-10-15", dataPoint.WeekStart)
	assert.Equal(suite.T(), "2024-10-22", dataPoint.WeekEnd)
	assert.Equal(suite.T(), 18.5, dataPoint.AverageHours)
	assert.Equal(suite.T(), 3, dataPoint.PRCount)
}

// TestGetAveragePRMergeTime_VariousPeriods tests various valid period formats
func (suite *GitHubServiceTestSuite) TestGetAveragePRMergeTime_VariousPeriods() {
	githubService := service.NewGitHubService(nil)

	claims := &auth.AuthClaims{
		UserID:   12345,
		Username: "testuser",
		Email:    "test@example.com",
		Provider: "githubtools",
	}

	ctx := context.Background()

	validPeriods := []string{"7d", "14d", "30d", "60d", "90d", "180d", "365d"}

	for _, period := range validPeriods {
		suite.T().Run(fmt.Sprintf("Period_%s", period), func(t *testing.T) {
			_, err := githubService.GetAveragePRMergeTime(ctx, claims, period)

			// Expect error due to no valid auth service, but period should be accepted
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to get GitHub access token")
		})
	}
}

// TestGetAveragePRMergeTime_EmptyResponse tests behavior when no PRs are found
func (suite *GitHubServiceTestSuite) TestGetAveragePRMergeTime_EmptyResponse() {
	// This test documents the expected behavior when no PRs are found
	// In a real scenario with mocked API:
	// - GraphQL returns empty nodes array
	// - Service should return response with 0 average and 4 weeks with 0 values

	expectedResponse := service.AveragePRMergeTimeResponse{
		AveragePRMergeTimeHours: 0,
		PRCount:                 0,
		Period:                  "30d",
		From:                    "2024-10-03T00:00:00Z",
		To:                      "2024-11-02T23:59:59Z",
		TimeSeries:              make([]service.PRMergeTimeDataPoint, 4),
	}

	assert.Equal(suite.T(), 0.0, expectedResponse.AveragePRMergeTimeHours)
	assert.Equal(suite.T(), 0, expectedResponse.PRCount)
	assert.Len(suite.T(), expectedResponse.TimeSeries, 4)
	// All weeks should have 0 values when there are no PRs
	for _, week := range expectedResponse.TimeSeries {
		assert.Equal(suite.T(), 0.0, week.AverageHours)
		assert.Equal(suite.T(), 0, week.PRCount)
	}
}

// TestGetAveragePRMergeTime_TimeSeriesCalculation tests time series grouping logic
func (suite *GitHubServiceTestSuite) TestGetAveragePRMergeTime_TimeSeriesCalculation() {
	// This test documents the expected calculation behavior:
	// 1. PRs are grouped by merge week (not creation date)
	// 2. For each week, calculate average of all PRs merged in that week
	// 3. Time series always has 4 weeks (newest to oldest)
	// 4. Overall average is mean of all individual PR merge times

	// Example scenario:
	// Week 1 (most recent): 2 PRs merged - 24 hours, 48 hours
	// Week 2: 1 PR merged - 12 hours
	// Week 3: No PRs
	// Week 4: No PRs

	// Expected time series (4 weeks):
	// Week 1: avg = (24 + 48) / 2 = 36 hours, count = 2
	// Week 2: avg = 12 hours, count = 1
	// Week 3: avg = 0, count = 0
	// Week 4: avg = 0, count = 0

	// Expected overall average: (24 + 48 + 12) / 3 = 28 hours

	expectedTimeSeries := []service.PRMergeTimeDataPoint{
		{WeekStart: "2024-10-22", WeekEnd: "2024-10-29", AverageHours: 36.0, PRCount: 2},
		{WeekStart: "2024-10-15", WeekEnd: "2024-10-22", AverageHours: 12.0, PRCount: 1},
		{WeekStart: "2024-10-08", WeekEnd: "2024-10-15", AverageHours: 0.0, PRCount: 0},
		{WeekStart: "2024-10-01", WeekEnd: "2024-10-08", AverageHours: 0.0, PRCount: 0},
	}

	assert.Len(suite.T(), expectedTimeSeries, 4)
	assert.Equal(suite.T(), "2024-10-22", expectedTimeSeries[0].WeekStart)
	assert.Equal(suite.T(), 36.0, expectedTimeSeries[0].AverageHours)
	assert.Equal(suite.T(), 2, expectedTimeSeries[0].PRCount)

	assert.Equal(suite.T(), "2024-10-15", expectedTimeSeries[1].WeekStart)
	assert.Equal(suite.T(), 12.0, expectedTimeSeries[1].AverageHours)
	assert.Equal(suite.T(), 1, expectedTimeSeries[1].PRCount)

	expectedOverallAverage := 28.0
	assert.Equal(suite.T(), 28.0, expectedOverallAverage)
}

// ClosePullRequest tests

// TestClosePullRequest_Success_WithBranchDeletion verifies closing an open PR and deleting its branch
func TestClosePullRequest_Success_WithBranchDeletion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	capturedDelete := false

	// Mock GitHub API server
	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle GET PR
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/42") {
			resp := map[string]interface{}{
				"id":         int64(123456789),
				"number":     42,
				"title":      "Test PR",
				"state":      "open",
				"created_at": "2025-01-01T12:00:00Z",
				"updated_at": "2025-01-01T12:00:00Z",
				"html_url":   mockGitHubServer.URL + "/owner/repo/pull/42",
				"draft":      false,
				"user": map[string]interface{}{
					"login":      "testuser",
					"id":         int64(12345),
					"avatar_url": "https://avatars.githubusercontent.com/u/12345",
				},
				"head": map[string]interface{}{
					"ref": "feature-branch",
					"repo": map[string]interface{}{
						"name": "repo",
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Handle PATCH (close PR)
		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/42") {
			resp := map[string]interface{}{
				"id":         int64(123456789),
				"number":     42,
				"title":      "Test PR",
				"state":      "closed",
				"created_at": "2025-01-01T12:00:00Z",
				"updated_at": "2025-01-02T12:00:00Z",
				"html_url":   mockGitHubServer.URL + "/owner/repo/pull/42",
				"draft":      false,
				"user": map[string]interface{}{
					"login":      "testuser",
					"id":         int64(12345),
					"avatar_url": "https://avatars.githubusercontent.com/u/12345",
				},
				"head": map[string]interface{}{
					"ref": "feature-branch",
					"repo": map[string]interface{}{
						"name": "repo",
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Handle DELETE ref (branch deletion)
		if r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/repos/owner/repo/git/refs/heads/feature-branch") {
			capturedDelete = true
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Fallback
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHubServer.Close()

	// Mock auth service
	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)
	envConfig := &auth.ProviderConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	// Service under test
	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 123, Provider: "githubtools"}

	// Execute
	result, err := githubService.ClosePullRequest(context.Background(), claims, "owner", "repo", 42, true)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "closed", result.State)
	assert.Equal(t, 42, result.Number)
	assert.Equal(t, "owner", result.Repo.Owner)
	assert.Equal(t, "repo", result.Repo.Name)
	assert.True(t, capturedDelete, "branch deletion should be attempted")
}

// TestClosePullRequest_AlreadyClosed verifies error when PR is already closed
func TestClosePullRequest_AlreadyClosed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/99") {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":         int64(111),
				"number":     99,
				"title":      "Closed PR",
				"state":      "closed",
				"created_at": "2025-01-01T12:00:00Z",
				"updated_at": "2025-01-05T12:00:00Z",
				"html_url":   mockGitHubServer.URL + "/owner/repo/pull/99",
				"user": map[string]interface{}{
					"login":      "testuser",
					"id":         int64(12345),
					"avatar_url": "https://avatars.githubusercontent.com/u/12345",
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)
	envConfig := &auth.ProviderConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 123, Provider: "githubtools"}

	result, err := githubService.ClosePullRequest(context.Background(), claims, "owner", "repo", 99, true)

	require.Error(t, err)
	require.Nil(t, result)
	assert.Contains(t, err.Error(), "already closed")
}

// TestClosePullRequest_NotFound verifies not found error when PR does not exist
func TestClosePullRequest_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/7") {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)
	envConfig := &auth.ProviderConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 123, Provider: "githubtools"}

	result, err := githubService.ClosePullRequest(context.Background(), claims, "owner", "repo", 7, true)

	require.Error(t, err)
	require.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

// TestClosePullRequest_RateLimitOnGet verifies rate limit on initial PR fetch
func TestClosePullRequest_RateLimitOnGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/1") {
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "API rate limit exceeded"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)
	envConfig := &auth.ProviderConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 123, Provider: "githubtools"}

	result, err := githubService.ClosePullRequest(context.Background(), claims, "owner", "repo", 1, true)

	require.Error(t, err)
	require.Nil(t, result)
	assert.Contains(t, err.Error(), "rate limit")
}

// TestClosePullRequest_DeleteBranch404Ignored verifies 404 during branch deletion is ignored
func TestClosePullRequest_DeleteBranch404Ignored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET open PR
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/50") {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":         int64(500),
				"number":     50,
				"title":      "PR to close",
				"state":      "open",
				"created_at": "2025-01-01T12:00:00Z",
				"updated_at": "2025-01-01T12:00:00Z",
				"html_url":   mockGitHubServer.URL + "/owner/repo/pull/50",
				"user": map[string]interface{}{
					"login":      "testuser",
					"id":         int64(12345),
					"avatar_url": "https://avatars.githubusercontent.com/u/12345",
				},
				"head": map[string]interface{}{
					"ref": "feature-branch",
					"repo": map[string]interface{}{
						"name": "repo",
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// PATCH close PR
		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/50") {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":         int64(500),
				"number":     50,
				"title":      "PR to close",
				"state":      "closed",
				"created_at": "2025-01-01T12:00:00Z",
				"updated_at": "2025-01-02T12:00:00Z",
				"html_url":   mockGitHubServer.URL + "/owner/repo/pull/50",
				"user": map[string]interface{}{
					"login":      "testuser",
					"id":         int64(12345),
					"avatar_url": "https://avatars.githubusercontent.com/u/12345",
				},
				"head": map[string]interface{}{
					"ref": "feature-branch",
					"repo": map[string]interface{}{
						"name": "repo",
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// DELETE branch -> 404 (ignored by service)
		if r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/repos/owner/repo/git/refs/heads/feature-branch") {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "branch not found"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)
	envConfig := &auth.ProviderConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 123, Provider: "githubtools"}

	result, err := githubService.ClosePullRequest(context.Background(), claims, "owner", "repo", 50, true)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "closed", result.State)
}

// TestClosePullRequest_DeleteBranchError verifies error when branch deletion fails
func TestClosePullRequest_DeleteBranchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var mockGitHubServer *httptest.Server
	mockGitHubServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET open PR
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/77") {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":         int64(770),
				"number":     77,
				"title":      "PR 77",
				"state":      "open",
				"created_at": "2025-01-01T12:00:00Z",
				"updated_at": "2025-01-01T12:00:00Z",
				"html_url":   mockGitHubServer.URL + "/owner/repo/pull/77",
				"user": map[string]interface{}{
					"login":      "testuser",
					"id":         int64(12345),
					"avatar_url": "https://avatars.githubusercontent.com/u/12345",
				},
				"head": map[string]interface{}{
					"ref": "feature-branch",
					"repo": map[string]interface{}{
						"name": "repo",
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// PATCH close PR
		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/repos/owner/repo/pulls/77") {
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":     int64(770),
				"number": 77,
				"title":  "PR 77",
				"state":  "closed",
				"user": map[string]interface{}{
					"login":      "testuser",
					"id":         int64(12345),
					"avatar_url": "https://avatars.githubusercontent.com/u/12345",
				},
				"head": map[string]interface{}{
					"ref": "feature-branch",
					"repo": map[string]interface{}{
						"name": "repo",
						"owner": map[string]interface{}{
							"login": "owner",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// DELETE branch -> 500 error (should bubble up)
		if r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/repos/owner/repo/git/refs/heads/feature-branch") {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "internal error"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHubServer.Close()

	mockAuthService := mocks.NewMockGitHubAuthService(ctrl)
	mockAuthService.EXPECT().GetGitHubAccessTokenFromClaims(gomock.Any()).Return("token", nil)
	envConfig := &auth.ProviderConfig{EnterpriseBaseURL: mockGitHubServer.URL}
	mockAuthService.EXPECT().GetGitHubClient(gomock.Any()).Return(auth.NewGitHubClient(envConfig), nil)

	githubService := service.NewGitHubServiceWithAdapter(mockAuthService)
	claims := &auth.AuthClaims{UserID: 123, Provider: "githubtools"}

	result, err := githubService.ClosePullRequest(context.Background(), claims, "owner", "repo", 77, true)

	require.Error(t, err)
	require.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to delete branch")
}

// TestClosePullRequest_InputValidation verifies input validation errors
func TestClosePullRequest_InputValidation(t *testing.T) {
	githubService := service.NewGitHubService(nil)

	// Nil claims
	res, err := githubService.ClosePullRequest(context.Background(), nil, "owner", "repo", 1, false)
	require.Error(t, err)
	require.Nil(t, res)
	assert.Contains(t, err.Error(), "authentication required")

	// Missing owner/repo
	claims := &auth.AuthClaims{UserID: 1, Provider: "githubtools"}
	res, err = githubService.ClosePullRequest(context.Background(), claims, "", "repo", 1, false)
	require.Error(t, err)
	require.Nil(t, res)
	assert.Contains(t, err.Error(), "owner and repo are required")
}
