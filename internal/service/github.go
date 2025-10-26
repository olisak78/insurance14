package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"developer-portal-backend/internal/auth"
	apperrors "developer-portal-backend/internal/errors"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// GitHubService provides methods to interact with GitHub API
type GitHubService struct {
	authService GitHubAuthService
}

// NewGitHubService creates a new GitHub service
func NewGitHubService(authService *auth.AuthService) *GitHubService {
	return &GitHubService{
		authService: NewAuthServiceAdapter(authService),
	}
}

// NewGitHubServiceWithAdapter creates a new GitHub service with a custom auth service adapter
// This constructor is primarily for testing with mock auth services
func NewGitHubServiceWithAdapter(authService GitHubAuthService) *GitHubService {
	return &GitHubService{
		authService: authService,
	}
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID        int64      `json:"id" example:"1234567890"`
	Number    int        `json:"number" example:"42"`
	Title     string     `json:"title" example:"Add new feature"`
	State     string     `json:"state" example:"open"`
	CreatedAt time.Time  `json:"created_at" example:"2025-01-01T12:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2025-01-02T12:00:00Z"`
	HTMLURL   string     `json:"html_url" example:"https://github.com/owner/repo/pull/42"`
	User      GitHubUser `json:"user"`
	Repo      Repository `json:"repository"`
	Draft     bool       `json:"draft" example:"false"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login     string `json:"login" example:"johndoe"`
	ID        int64  `json:"id" example:"12345"`
	AvatarURL string `json:"avatar_url" example:"https://avatars.githubusercontent.com/u/12345"`
}

// Repository represents a GitHub repository
type Repository struct {
	Name     string `json:"name" example:"my-repo"`
	FullName string `json:"full_name" example:"owner/my-repo"`
	Owner    string `json:"owner" example:"owner"`
	Private  bool   `json:"private" example:"false"`
}

// PullRequestsResponse represents the response for pull requests
type PullRequestsResponse struct {
	PullRequests []PullRequest `json:"pull_requests"`
	Total        int           `json:"total"`
}

// TotalContributionsResponse represents the response for user contributions
type TotalContributionsResponse struct {
	TotalContributions int    `json:"total_contributions" example:"1234"`
	Period             string `json:"period" example:"365d"`
	From               string `json:"from" example:"2024-10-16T00:00:00Z"`
	To                 string `json:"to" example:"2025-10-16T23:59:59Z"`
}

// parseRepositoryFromURL extracts repository information from a GitHub URL
// Handles URLs like: https://github.com/owner/repo/pull/123
// or https://github.enterprise.com/owner/repo/pull/123
func parseRepositoryFromURL(urlStr string) (owner, repoName, fullName string) {
	if urlStr == "" {
		return "", "", ""
	}

	// Remove protocol and split by /
	urlStr = strings.TrimPrefix(urlStr, "https://")
	urlStr = strings.TrimPrefix(urlStr, "http://")
	parts := strings.Split(urlStr, "/")

	// We need at least: domain/owner/repo/...
	if len(parts) < 3 {
		return "", "", ""
	}

	// parts[0] is the domain (github.com or github.enterprise.com)
	// parts[1] is the owner
	// parts[2] is the repo name
	owner = parts[1]
	repoName = parts[2]
	fullName = owner + "/" + repoName

	return owner, repoName, fullName
}

// GetUserOpenPullRequests retrieves all open pull requests for the authenticated user
func (s *GitHubService) GetUserOpenPullRequests(ctx context.Context, claims *auth.AuthClaims, state, sort, direction string, perPage, page int) (*PullRequestsResponse, error) {
	if claims == nil {
		return nil, fmt.Errorf("authentication required")
	}

	// Get GitHub access token using validated JWT claims
	accessToken, err := s.authService.GetGitHubAccessTokenFromClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub access token: %w", err)
	}

	// Get GitHub client configuration for the user's provider
	githubClientConfig, err := s.authService.GetGitHubClient(claims.Provider, claims.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	// Create OAuth2 client with access token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create authenticated GitHub client
	var client *github.Client
	if githubClientConfig != nil && githubClientConfig.GetEnterpriseBaseURL() != "" {
		client, err = github.NewEnterpriseClient(githubClientConfig.GetEnterpriseBaseURL(), githubClientConfig.GetEnterpriseBaseURL(), tc)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub Enterprise client: %w", err)
		}
	} else {
		client = github.NewClient(tc)
	}

	// Set default values
	if state == "" {
		state = "open"
	}
	if sort == "" {
		sort = "created"
	}
	if direction == "" {
		direction = "desc"
	}
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}
	if page <= 0 {
		page = 1
	}

	// Search for pull requests created by the authenticated user
	// Using search API for better filtering capabilities
	// Note: GitHub Search API doesn't support state:all - omit state qualifier to get all PRs
	var query string
	if state == "all" {
		query = "is:pr author:@me"
	} else {
		query = fmt.Sprintf("is:pr author:@me state:%s", state)
	}

	searchOpts := &github.SearchOptions{
		Sort:  sort,
		Order: direction,
		ListOptions: github.ListOptions{
			PerPage: perPage,
			Page:    page,
		},
	}

	result, resp, err := client.Search.Issues(ctx, query, searchOpts)
	if err != nil {
		// Check if it's a rate limit error
		if resp != nil && resp.StatusCode == 403 {
			return nil, apperrors.ErrGitHubAPIRateLimitExceeded
		}
		return nil, fmt.Errorf("failed to search pull requests: %w", err)
	}

	// Convert GitHub issues (PRs are issues in GitHub API) to our PR structure
	pullRequests := make([]PullRequest, 0, len(result.Issues))
	for _, issue := range result.Issues {
		if issue.PullRequestLinks == nil {
			continue // Skip if it's not actually a PR
		}

		pr := PullRequest{
			ID:        issue.GetID(),
			Number:    issue.GetNumber(),
			Title:     issue.GetTitle(),
			State:     issue.GetState(),
			CreatedAt: issue.GetCreatedAt().Time,
			UpdatedAt: issue.GetUpdatedAt().Time,
			HTMLURL:   issue.GetHTMLURL(),
			Draft:     issue.GetDraft(),
			User: GitHubUser{
				Login:     issue.GetUser().GetLogin(),
				ID:        issue.GetUser().GetID(),
				AvatarURL: issue.GetUser().GetAvatarURL(),
			},
		}

		// Parse repository info from the issue
		if issue.Repository != nil {
			pr.Repo = Repository{
				Name:     issue.Repository.GetName(),
				FullName: issue.Repository.GetFullName(),
				Private:  issue.Repository.GetPrivate(),
			}
			if issue.Repository.Owner != nil {
				pr.Repo.Owner = issue.Repository.Owner.GetLogin()
			}
		} else {
			// Fallback: parse repository info from the HTML URL
			// GitHub Search API often doesn't include the Repository field
			owner, repoName, fullName := parseRepositoryFromURL(issue.GetHTMLURL())
			if owner != "" && repoName != "" {
				pr.Repo = Repository{
					Name:     repoName,
					FullName: fullName,
					Owner:    owner,
					// Note: We can't determine if the repo is private from the URL alone
					// Default to false, but this could be enhanced with an additional API call if needed
					Private: false,
				}
			}
		}

		pullRequests = append(pullRequests, pr)
	}

	response := &PullRequestsResponse{
		PullRequests: pullRequests,
		Total:        result.GetTotal(),
	}

	return response, nil
}

// GetUserTotalContributions retrieves the total contributions for the authenticated user over a specified period
func (s *GitHubService) GetUserTotalContributions(ctx context.Context, claims *auth.AuthClaims, period string) (*TotalContributionsResponse, error) {
	if claims == nil {
		return nil, fmt.Errorf("authentication required")
	}

	// Validate period format early (before making any API calls)
	var from, to time.Time
	var parsedPeriod string
	var query string

	if period == "" {
		// No period specified - use GitHub's default behavior
		parsedPeriod = "github_default"
		query = `{
			viewer {
				contributionsCollection {
					startedAt
					endedAt
					contributionCalendar {
						totalContributions
					}
				}
			}
		}`
	} else {
		// Validate period format before parsing
		if len(period) < 2 || period[len(period)-1] != 'd' {
			return nil, fmt.Errorf("invalid period format: period must be in format '<number>d' (e.g., '30d', '90d', '365d')")
		}

		// Parse custom period and calculate date range
		var err error
		from, to, parsedPeriod, err = parsePeriod(period)
		if err != nil {
			return nil, fmt.Errorf("invalid period format: %w", err)
		}

		query = fmt.Sprintf(`{
			viewer {
				contributionsCollection(from: "%s", to: "%s") {
					startedAt
					endedAt
					contributionCalendar {
						totalContributions
					}
				}
			}
		}`, from.Format(time.RFC3339), to.Format(time.RFC3339))
	}

	// Get GitHub access token using validated JWT claims
	accessToken, err := s.authService.GetGitHubAccessTokenFromClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub access token: %w", err)
	}

	// Get GitHub client configuration for the user's provider
	githubClientConfig, err := s.authService.GetGitHubClient(claims.Provider, claims.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	// Execute GraphQL query
	var result struct {
		Viewer struct {
			ContributionsCollection struct {
				StartedAt            string `json:"startedAt"`
				EndedAt              string `json:"endedAt"`
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"viewer"`
	}

	reqBody := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	// Determine the correct GraphQL endpoint
	var graphqlURL string
	if githubClientConfig != nil && githubClientConfig.GetEnterpriseBaseURL() != "" {
		// GitHub Enterprise: Use /api/graphql (NOT /api/v3/graphql)
		graphqlURL = strings.TrimSuffix(githubClientConfig.GetEnterpriseBaseURL(), "/") + "/api/graphql"
	} else {
		// GitHub.com: Use standard GraphQL endpoint
		graphqlURL = "https://api.github.com/graphql"
	}

	// Create HTTP request manually for GraphQL
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL query: %w", err)
	}

	ghReq, err := http.NewRequestWithContext(ctx, "POST", graphqlURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	// Set headers
	ghReq.Header.Set("Authorization", "Bearer "+accessToken)
	ghReq.Header.Set("Content-Type", "application/json")
	ghReq.Header.Set("Accept", "application/json")

	// Execute request - respect context deadline if available
	httpClient := &http.Client{}
	if deadline, ok := ctx.Deadline(); ok {
		httpClient.Timeout = time.Until(deadline)
	} else {
		httpClient.Timeout = 30 * time.Second
	}
	resp, err := httpClient.Do(ghReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL query: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limit
	if resp.StatusCode == 403 {
		return nil, apperrors.ErrGitHubAPIRateLimitExceeded
	}

	// Check for other HTTP errors
	if resp.StatusCode != 200 {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("GraphQL query failed with status %d and failed to read response body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("GraphQL query failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response - GitHub GraphQL returns data in a wrapper
	var graphQLResponse struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string   `json:"message"`
			Path    []string `json:"path,omitempty"`
		} `json:"errors,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&graphQLResponse); err != nil {
		return nil, fmt.Errorf("failed to decode GraphQL response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphQLResponse.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphQLResponse.Errors[0].Message)
	}

	// Parse the actual data
	if err := json.Unmarshal(graphQLResponse.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Use GitHub's actual date range from the response
	// This ensures we return what GitHub actually used, not just what we calculated
	fromStr := result.Viewer.ContributionsCollection.StartedAt
	toStr := result.Viewer.ContributionsCollection.EndedAt

	// If GitHub didn't return dates (shouldn't happen), fall back to calculated dates
	if fromStr == "" && !from.IsZero() {
		fromStr = from.Format(time.RFC3339)
	}
	if toStr == "" && !to.IsZero() {
		toStr = to.Format(time.RFC3339)
	}

	response := &TotalContributionsResponse{
		TotalContributions: result.Viewer.ContributionsCollection.ContributionCalendar.TotalContributions,
		Period:             parsedPeriod,
		From:               fromStr,
		To:                 toStr,
	}

	return response, nil
}

// parsePeriod parses a period string (e.g., "30d", "90d", "365d") and returns the from/to dates
// Default period is 365 days if not specified or invalid
func parsePeriod(period string) (from, to time.Time, parsedPeriod string, err error) {
	to = time.Now().UTC()

	// Default to 365 days if empty or invalid
	days := 365
	parsedPeriod = "365d"

	if period != "" {
		// Parse period format like "30d", "90d", "365d"
		if len(period) < 2 || period[len(period)-1] != 'd' {
			return time.Time{}, time.Time{}, "", fmt.Errorf("period must be in format '<number>d' (e.g., '30d', '90d', '365d')")
		}

		var parseErr error
		days, parseErr = strconv.Atoi(period[:len(period)-1])
		if parseErr != nil || days <= 0 {
			return time.Time{}, time.Time{}, "", fmt.Errorf("period must contain a positive number of days")
		}

		// GitHub API supports max 1 year of contributions
		if days > 365 {
			days = 365
		}

		parsedPeriod = period
	}

	from = to.AddDate(0, 0, -days)

	// Set from to start of day and to to end of day
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, time.UTC)

	return from, to, parsedPeriod, nil
}
