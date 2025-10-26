package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"developer-portal-backend/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJiraService_GetIssues_TeamFilter(t *testing.T) {
	tests := []struct {
		name           string
		filters        JiraIssueFilters
		mockResponse   jiraSearchResponse
		mockStatusCode int
		expectError    bool
		expectedTotal  int
	}{
		{
			name: "successful team issues search",
			filters: JiraIssueFilters{
				Team:    "TestTeam",
				Project: "SAPBTPCFS",
				Status:  "Open,In Progress,To Do",
			},
			mockResponse: jiraSearchResponse{
				Total: 2,
				Issues: []JiraIssue{
					{
						ID:  "1",
						Key: "SAPBTPCFS-123",
						Fields: JiraIssueFields{
							Summary: "Test issue 1",
							Status:  JiraStatus{ID: "1", Name: "In Progress"},
							IssueType: JiraIssueType{ID: "1", Name: "Story"},
							Created: "2023-01-01T00:00:00.000Z",
							Updated: "2023-01-02T00:00:00.000Z",
						},
					},
					{
						ID:  "2",
						Key: "SAPBTPCFS-124",
						Fields: JiraIssueFields{
							Summary: "Test issue 2",
							Status:  JiraStatus{ID: "2", Name: "To Do"},
							IssueType: JiraIssueType{ID: "1", Name: "Story"},
							Created: "2023-01-01T00:00:00.000Z",
							Updated: "2023-01-02T00:00:00.000Z",
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  2,
		},
		{
			name: "empty team issues search",
			filters: JiraIssueFilters{
				Team:    "EmptyTeam",
				Project: "SAPBTPCFS",
				Status:  "Open,In Progress,To Do",
			},
			mockResponse:   jiraSearchResponse{Total: 0, Issues: []JiraIssue{}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  0,
		},
		{
			name: "jira server error",
			filters: JiraIssueFilters{
				Team:    "TestTeam",
				Project: "SAPBTPCFS",
				Status:  "Open,In Progress,To Do",
			},
			mockResponse:   jiraSearchResponse{},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Handle PAT creation via Basic auth
	if r.Method == http.MethodPost && r.URL.Path == "/rest/pat/latest/tokens" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         123,
			"name":       "test-token",
			"createdAt":  time.Now().Format(time.RFC3339Nano),
			"expiringAt": time.Now().Add(90 * 24 * time.Hour).Format(time.RFC3339Nano),
			"rawToken":   "test-pat-token",
		})
		return
	}

	// Verify the JQL query contains the expected filters
	jql := r.URL.Query().Get("jql")
	if tt.filters.Project != "" {
		assert.Contains(t, jql, `project = "`+tt.filters.Project+`"`)
	}
	if tt.filters.Team != "" {
		assert.Contains(t, jql, `"Team(s)" = "`+tt.filters.Team+`"`)
	}
	if tt.filters.Status != "" {
		assert.Contains(t, jql, `status IN`)
	}

	w.WriteHeader(tt.mockStatusCode)
	if tt.mockStatusCode == http.StatusOK {
		json.NewEncoder(w).Encode(tt.mockResponse)
	}
}))
			defer server.Close()

			// Create service with mock server URL
			cfg := &config.Config{
				JiraDomain:   server.URL,
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			// Execute test
			result, err := service.GetIssues(tt.filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedTotal, result.Total)
				assert.Len(t, result.Issues, tt.expectedTotal)
				
				// Verify project and link fields are populated
				for _, issue := range result.Issues {
					if issue.Key != "" {
						assert.NotEmpty(t, issue.Project, "Project should be populated")
						assert.NotEmpty(t, issue.Link, "Link should be populated")
						assert.Contains(t, issue.Link, "/browse/"+issue.Key, "Link should contain browse URL with issue key")
					}
				}
			}
		})
	}
}

func TestJiraService_GetIssues_UserFilter(t *testing.T) {
	tests := []struct {
		name           string
		filters        JiraIssueFilters
		mockResponse   jiraSearchResponse
		mockStatusCode int
		expectError    bool
		expectedTotal  int
	}{
		{
			name: "successful user issues search",
			filters: JiraIssueFilters{
				User:   "testuser",
				Status: "Open,In Progress,To Do",
			},
			mockResponse: jiraSearchResponse{
				Total: 3,
				Issues: []JiraIssue{
					{
						ID:  "1",
						Key: "SAPBTPCFS-123",
						Fields: JiraIssueFields{
							Summary: "User issue 1",
							Status:  JiraStatus{ID: "1", Name: "In Progress"},
							IssueType: JiraIssueType{ID: "1", Name: "Story"},
							Assignee: &JiraUser{
								AccountID:   "123",
								DisplayName: "Test User",
							},
							Created: "2023-01-01T00:00:00.000Z",
							Updated: "2023-01-02T00:00:00.000Z",
						},
					},
					{
						ID:  "2",
						Key: "SAPBTPCFSBUGS-456",
						Fields: JiraIssueFields{
							Summary: "User bug 1",
							Status:  JiraStatus{ID: "2", Name: "To Do"},
							IssueType: JiraIssueType{ID: "2", Name: "Bug"},
							Assignee: &JiraUser{
								AccountID:   "123",
								DisplayName: "Test User",
							},
							Created: "2023-01-01T00:00:00.000Z",
							Updated: "2023-01-02T00:00:00.000Z",
						},
					},
					{
						ID:  "3",
						Key: "OTHER-789",
						Fields: JiraIssueFields{
							Summary: "User task 1",
							Status:  JiraStatus{ID: "3", Name: "Reopened"},
							IssueType: JiraIssueType{ID: "3", Name: "Task"},
							Assignee: &JiraUser{
								AccountID:   "123",
								DisplayName: "Test User",
							},
							Created: "2023-01-01T00:00:00.000Z",
							Updated: "2023-01-02T00:00:00.000Z",
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  3,
		},
		{
			name: "empty user issues search",
			filters: JiraIssueFilters{
				User:   "emptyuser",
				Status: "Open,In Progress,To Do",
			},
			mockResponse:   jiraSearchResponse{Total: 0, Issues: []JiraIssue{}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  0,
		},
		{
			name: "jira server error",
			filters: JiraIssueFilters{
				User:   "testuser",
				Status: "Open,In Progress,To Do",
			},
			mockResponse:   jiraSearchResponse{},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Handle PAT creation via Basic auth
	if r.Method == http.MethodPost && r.URL.Path == "/rest/pat/latest/tokens" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         123,
			"name":       "test-token",
			"createdAt":  time.Now().Format(time.RFC3339Nano),
			"expiringAt": time.Now().Add(90 *24 * time.Hour).Format(time.RFC3339Nano),
			"rawToken":   "test-pat-token",
		})
		return
	}

	// Verify the JQL query contains the expected user filter
	jql := r.URL.Query().Get("jql")
	if tt.filters.User != "" {
		assert.Contains(t, jql, `assignee = "`+tt.filters.User+`"`)
	}
	if tt.filters.Status != "" {
		assert.Contains(t, jql, `status IN`)
	}

	w.WriteHeader(tt.mockStatusCode)
	if tt.mockStatusCode == http.StatusOK {
		json.NewEncoder(w).Encode(tt.mockResponse)
	}
}))
			defer server.Close()

			// Create service with mock server URL
			cfg := &config.Config{
				JiraDomain:   server.URL,
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			// Execute test
			result, err := service.GetIssues(tt.filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedTotal, result.Total)
				assert.Len(t, result.Issues, tt.expectedTotal)
			}
		})
	}
}

func TestJiraService_GetIssuesCount(t *testing.T) {
	tests := []struct {
		name           string
		filters        JiraIssueFilters
		mockResponse   jiraSearchResponse
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful count query",
			filters: JiraIssueFilters{
				User:   "testuser",
				Status: "Resolved",
				Date:   "2023-01-01",
			},
			mockResponse:   jiraSearchResponse{Total: 5, Issues: []JiraIssue{}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  5,
		},
		{
			name: "zero count query",
			filters: JiraIssueFilters{
				User:   "emptyuser",
				Status: "Resolved",
				Date:   "2023-01-01",
			},
			mockResponse:   jiraSearchResponse{Total: 0, Issues: []JiraIssue{}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name: "jira server error",
			filters: JiraIssueFilters{
				User:   "testuser",
				Status: "Resolved",
				Date:   "2023-01-01",
			},
			mockResponse:   jiraSearchResponse{},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Handle PAT creation via Basic auth
	if r.Method == http.MethodPost && r.URL.Path == "/rest/pat/latest/tokens" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         123,
			"name":       "test-token",
			"createdAt":  time.Now().Format(time.RFC3339Nano),
			"expiringAt": time.Now().Add(90 * 24 * time.Hour).Format(time.RFC3339Nano),
			"rawToken":   "test-pat-token",
		})
		return
	}

	// Verify the JQL query and maxResults parameter
	jql := r.URL.Query().Get("jql")
	maxResults := r.URL.Query().Get("maxResults")
	
	if tt.filters.User != "" {
		assert.Contains(t, jql, `assignee = "`+tt.filters.User+`"`)
	}
	if tt.filters.Status != "" && tt.filters.Date != "" {
		assert.Contains(t, jql, `status CHANGED TO`)
	}
	assert.Equal(t, "0", maxResults) // Count queries should have maxResults=0

	w.WriteHeader(tt.mockStatusCode)
	if tt.mockStatusCode == http.StatusOK {
		json.NewEncoder(w).Encode(tt.mockResponse)
	}
}))
			defer server.Close()

			// Create service with mock server URL
			cfg := &config.Config{
				JiraDomain:   server.URL,
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			// Execute test
			count, err := service.GetIssuesCount(tt.filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0, count)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestJiraService_GetIssues_Pagination(t *testing.T) {
	tests := []struct {
		name           string
		filters        JiraIssueFilters
		mockResponse   jiraSearchResponse
		mockStatusCode int
		expectError    bool
		expectedPage   int
		expectedLimit  int
	}{
		{
			name: "pagination with custom page and limit",
			filters: JiraIssueFilters{
				Project: "SAPBTPCFS",
				Status:  "Open",
				Page:    2,
				Limit:   25,
			},
			mockResponse: jiraSearchResponse{
				Total:  100,
				Issues: make([]JiraIssue, 25), // 25 issues for page 2
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedPage:   2,
			expectedLimit:  25,
		},
		{
			name: "pagination with default values",
			filters: JiraIssueFilters{
				Project: "SAPBTPCFS",
				Status:  "Open",
				// Page and Limit not set, should use defaults
			},
			mockResponse: jiraSearchResponse{
				Total:  100,
				Issues: make([]JiraIssue, 50), // Default limit of 50
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedPage:   1,
			expectedLimit:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Handle PAT creation via Basic auth
	if r.Method == http.MethodPost && r.URL.Path == "/rest/pat/latest/tokens" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         123,
			"name":       "test-token",
			"createdAt":  time.Now().Format(time.RFC3339Nano),
			"expiringAt": time.Now().Add(90 * 24 * time.Hour).Format(time.RFC3339Nano),
			"rawToken":   "test-pat-token",
		})
		return
	}

	// Verify pagination parameters
	startAt := r.URL.Query().Get("startAt")
	maxResults := r.URL.Query().Get("maxResults")
	
	expectedStartAt := (tt.expectedPage - 1) * tt.expectedLimit
	assert.Equal(t, fmt.Sprintf("%d", expectedStartAt), startAt)
	assert.Equal(t, fmt.Sprintf("%d", tt.expectedLimit), maxResults)

	w.WriteHeader(tt.mockStatusCode)
	if tt.mockStatusCode == http.StatusOK {
		json.NewEncoder(w).Encode(tt.mockResponse)
	}
}))
			defer server.Close()

			// Create service with mock server URL
			cfg := &config.Config{
				JiraDomain:   server.URL,
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			// Execute test
			result, err := service.GetIssues(tt.filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedPage, result.Page)
				assert.Equal(t, tt.expectedLimit, result.Limit)
			}
		})
	}
}

func TestJiraService_ConfigurationValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
	}{
		{
			name: "missing jira domain",
			config: &config.Config{
				JiraDomain:   "",
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			},
		},
		{
			name: "missing jira user",
			config: &config.Config{
				JiraDomain:   "https://test.atlassian.net",
				JiraUser:     "",
				JiraPassword: "testpass",
			},
		},
		{
			name: "missing jira password",
			config: &config.Config{
				JiraDomain:   "https://test.atlassian.net",
				JiraUser:     "testuser",
				JiraPassword: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewJiraService(tt.config)

			// Test GetIssues returns configuration error
			filters := JiraIssueFilters{Project: "TEST", Status: "Open"}
			_, err := service.GetIssues(filters)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "jira configuration missing")

			// Test GetIssuesCount returns configuration error
			_, err = service.GetIssuesCount(filters)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "jira configuration missing")
		})
	}
}

func TestJiraService_JQLValidation(t *testing.T) {
	tests := []struct {
		name        string
		filters     JiraIssueFilters
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid filters",
			filters: JiraIssueFilters{
				Project: "SAPBTPCFS",
				Status:  "Open,In Progress",
				Team:    "TestTeam",
			},
			expectError: false,
		},
		{
			name: "invalid project with newlines",
			filters: JiraIssueFilters{
				Project: "SAPBTPCFS\nmalicious",
				Status:  "Open",
			},
			expectError: true,
			errorMsg:    "invalid project value",
		},
		{
			name: "invalid status with tabs",
			filters: JiraIssueFilters{
				Project: "SAPBTPCFS",
				Status:  "Open\tmalicious",
			},
			expectError: true,
			errorMsg:    "invalid status value",
		},
		{
			name: "no search criteria",
			filters: JiraIssueFilters{
				// All fields empty
			},
			expectError: true,
			errorMsg:    "at least one search criterion must be provided",
		},
		{
			name: "invalid date format",
			filters: JiraIssueFilters{
				User:   "testuser",
				Status: "Resolved",
				Date:   "invalid-date",
			},
			expectError: true,
			errorMsg:    "invalid date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				JiraDomain:   "https://test.atlassian.net",
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			// Execute test
			_, err := service.GetIssues(tt.filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// For valid filters, we expect a network error since we're not mocking the server
				// but we should not get a JQL validation error
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid")
					assert.NotContains(t, err.Error(), "no valid search criteria")
				}
			}
		})
	}
}

func TestJiraService_URLValidation(t *testing.T) {
	tests := []struct {
		name           string
		jiraDomain     string
		expectError    bool
		errorMsg       string
	}{
		{
			name:        "valid domain without protocol",
			jiraDomain:  "test.atlassian.net",
			expectError: false,
		},
		{
			name:        "valid domain with https",
			jiraDomain:  "https://test.atlassian.net",
			expectError: false,
		},
		{
			name:        "valid domain with http",
			jiraDomain:  "http://test.atlassian.net",
			expectError: false,
		},
		{
			name:        "invalid URL with spaces",
			jiraDomain:  "invalid url with spaces",
			expectError: true,
			errorMsg:    "invalid jira domain URL",
		},
		{
			name:        "empty domain",
			jiraDomain:  "",
			expectError: true,
			errorMsg:    "jira configuration missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				JiraDomain:   tt.jiraDomain,
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			filters := JiraIssueFilters{
				Project: "SAPBTPCFS",
				Status:  "Open",
			}

			// Execute test
			_, err := service.GetIssues(filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// For valid URLs, we expect a network error since we're not mocking the server
				// but we should not get a URL validation error
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid jira domain URL")
					assert.NotContains(t, err.Error(), "jira domain is not configured")
				}
			}
		})
	}
}

func TestJiraService_NewParameterFilters(t *testing.T) {
	tests := []struct {
		name           string
		filters        JiraIssueFilters
		mockResponse   jiraSearchResponse
		mockStatusCode int
		expectError    bool
		expectedTotal  int
		expectedJQL    string
	}{
		{
			name: "assignee filter",
			filters: JiraIssueFilters{
				Assignee: "john.doe",
				Status:   "Open",
			},
			mockResponse: jiraSearchResponse{
				Total: 5,
				Issues: make([]JiraIssue, 5),
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  5,
			expectedJQL:    `status = "Open" AND assignee = "john.doe" ORDER BY created DESC`,
		},
		{
			name: "type filter single value",
			filters: JiraIssueFilters{
				Type:    "Bug",
				Project: "SAPBTPCFS",
			},
			mockResponse: jiraSearchResponse{
				Total: 3,
				Issues: make([]JiraIssue, 3),
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  3,
			expectedJQL:    `project = "SAPBTPCFS" AND issuetype = "Bug" ORDER BY created DESC`,
		},
		{
			name: "type filter multiple values",
			filters: JiraIssueFilters{
				Type:    "Bug,Task,Story",
				Project: "SAPBTPCFS",
			},
			mockResponse: jiraSearchResponse{
				Total: 10,
				Issues: make([]JiraIssue, 10),
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  10,
			expectedJQL:    `project = "SAPBTPCFS" AND issuetype IN ("Bug", "Task", "Story") ORDER BY created DESC`,
		},
		{
			name: "summary text search",
			filters: JiraIssueFilters{
				Summary: "authentication",
				Project: "SAPBTPCFS",
			},
			mockResponse: jiraSearchResponse{
				Total: 2,
				Issues: make([]JiraIssue, 2),
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  2,
			expectedJQL:    `project = "SAPBTPCFS" AND summary ~ "authentication" ORDER BY created DESC`,
		},
		{
			name: "key filter",
			filters: JiraIssueFilters{
				Key: "BUG-1234",
			},
			mockResponse: jiraSearchResponse{
				Total: 1,
				Issues: []JiraIssue{
					{
						ID:  "1",
						Key: "BUG-1234",
						Fields: JiraIssueFields{
							Summary: "Test bug",
							Status:  JiraStatus{ID: "1", Name: "Open"},
							IssueType: JiraIssueType{ID: "1", Name: "Bug"},
							Created: "2023-01-01T00:00:00.000Z",
							Updated: "2023-01-02T00:00:00.000Z",
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  1,
			expectedJQL:    `key = "BUG-1234" ORDER BY created DESC`,
		},
		{
			name: "complex multi-parameter filter",
			filters: JiraIssueFilters{
				Project:  "SAPBTPCFS",
				Status:   "Open,In Progress",
				Assignee: "john.doe",
				Type:     "Bug",
				Summary:  "login",
			},
			mockResponse: jiraSearchResponse{
				Total: 1,
				Issues: make([]JiraIssue, 1),
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedTotal:  1,
			expectedJQL:    `project = "SAPBTPCFS" AND status IN ("Open", "In Progress") AND assignee = "john.doe" AND issuetype = "Bug" AND summary ~ "login" ORDER BY created DESC`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Handle PAT creation via Basic auth
	if r.Method == http.MethodPost && r.URL.Path == "/rest/pat/latest/tokens" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         123,
			"name":       "test-token",
			"createdAt":  time.Now().Format(time.RFC3339Nano),
			"expiringAt": time.Now().Add(90 * 24 * time.Hour).Format(time.RFC3339Nano),
			"rawToken":   "test-pat-token",
		})
		return
	}

	// Verify the JQL query matches expected
	jql := r.URL.Query().Get("jql")
	assert.Equal(t, tt.expectedJQL, jql, "Generated JQL should match expected")

	w.WriteHeader(tt.mockStatusCode)
	if tt.mockStatusCode == http.StatusOK {
		json.NewEncoder(w).Encode(tt.mockResponse)
	}
}))
			defer server.Close()

			// Create service with mock server URL
			cfg := &config.Config{
				JiraDomain:   server.URL,
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			// Execute test
			result, err := service.GetIssues(tt.filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedTotal, result.Total)
				assert.Len(t, result.Issues, tt.expectedTotal)
				
				// Verify project and link fields are populated for issues with keys
				for _, issue := range result.Issues {
					if issue.Key != "" {
						assert.NotEmpty(t, issue.Project, "Project should be populated")
						assert.NotEmpty(t, issue.Link, "Link should be populated")
						assert.Contains(t, issue.Link, "/browse/"+issue.Key, "Link should contain browse URL with issue key")
					}
				}
			}
		})
	}
}

func TestJiraService_NewParameterValidation(t *testing.T) {
	tests := []struct {
		name        string
		filters     JiraIssueFilters
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid assignee",
			filters: JiraIssueFilters{
				Assignee: "john.doe",
			},
			expectError: false,
		},
		{
			name: "invalid assignee with newlines",
			filters: JiraIssueFilters{
				Assignee: "john.doe\nmalicious",
			},
			expectError: true,
			errorMsg:    "invalid assignee value",
		},
		{
			name: "valid type",
			filters: JiraIssueFilters{
				Type: "Bug,Task",
			},
			expectError: false,
		},
		{
			name: "invalid type with tabs",
			filters: JiraIssueFilters{
				Type: "Bug\tmalicious",
			},
			expectError: true,
			errorMsg:    "invalid type value",
		},
		{
			name: "valid summary",
			filters: JiraIssueFilters{
				Summary: "authentication issue",
			},
			expectError: false,
		},
		{
			name: "invalid summary with newlines",
			filters: JiraIssueFilters{
				Summary: "auth\nmalicious",
			},
			expectError: true,
			errorMsg:    "invalid summary value",
		},
		{
			name: "valid key",
			filters: JiraIssueFilters{
				Key: "BUG-1234",
			},
			expectError: false,
		},
		{
			name: "invalid key with tabs",
			filters: JiraIssueFilters{
				Key: "BUG-1234\tmalicious",
			},
			expectError: true,
			errorMsg:    "invalid key value",
		},
		{
			name: "too long assignee",
			filters: JiraIssueFilters{
				Assignee: strings.Repeat("a", 256),
			},
			expectError: true,
			errorMsg:    "invalid assignee value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				JiraDomain:   "https://test.atlassian.net",
				JiraUser:     "testuser",
				JiraPassword: "testpass",
			}
			service := NewJiraService(cfg)

			// Execute test
			_, err := service.GetIssues(tt.filters)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// For valid filters, we expect a network error since we're not mocking the server
				// but we should not get a validation error
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid")
				}
			}
		})
	}
}
