package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"developer-portal-backend/internal/config"
	apperrors "developer-portal-backend/internal/errors"
)

// JiraService provides methods to interact with Jira
type JiraService struct {
	cfg        *config.Config
	httpClient *http.Client

	// Personal Access Token (PAT) management
	patToken string
	patExpiry time.Time
	tokenMu   sync.Mutex

	// Fixed PAT name including machine identifier
	patName string
}

/**
 * NewJiraService creates a new Jira service
 */
func NewJiraService(cfg *config.Config) *JiraService {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown-host"
	}
	name := fmt.Sprintf("DeveloperPortal-%s", hostname)
	name = strings.ReplaceAll(name, " ", "-")

	return &JiraService{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		patName:    name,
	}
}

// patTokenResponse represents the response from Jira PAT creation endpoint
type patTokenResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"createdAt"`
	ExpiringAt string `json:"expiringAt"`
	RawToken   string `json:"rawToken"`
}

// InitializePATOnStartup checks for an existing PAT with the fixed name and deletes it before creating a new one.
// This should be called once on server startup to avoid hitting PAT limits and to ensure a clean token lifecycle.
func (s *JiraService) InitializePATOnStartup() error {
	// Parse base URL from config (same handling as in searchIssues)
	base := s.cfg.JiraDomain
	if base == "" {
		return fmt.Errorf("jira domain is not configured")
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	baseURL, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return fmt.Errorf("invalid jira domain URL '%s': %w", base, err)
	}

	// Attempt cleanup and log outcome
	found, err := s.cleanupExistingPAT(baseURL)
	if err != nil {
		log.Printf("Jira PAT cleanup error for name=%s: %v", s.patName, err)
	} else if found {
		log.Printf("Jira PAT with name=%s was found and deleted", s.patName)
	} else {
		log.Printf("No existing Jira PAT found with name=%s", s.patName)
	}

	// Create new PAT
	return s.createPAT(baseURL)
}

// cleanupExistingPAT finds a PAT by name and deletes it if present.
// Returns true if a matching token was found and deleted.
func (s *JiraService) cleanupExistingPAT(baseURL *url.URL) (bool, error) {
	// GET all PATs
	listURL := baseURL.String() + "/rest/pat/latest/tokens"
	req, err := http.NewRequest(http.MethodGet, listURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create PAT list request: %w", err)
	}
	cred := base64.StdEncoding.EncodeToString([]byte(s.cfg.JiraUser + ":" + s.cfg.JiraPassword))
	req.Header.Set("Authorization", "Basic "+cred)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("jira PAT list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("jira PAT list failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Decode array of PATs
 var tokens []patTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return false, fmt.Errorf("failed to decode PAT list: %w", err)
	}

	for _, tok := range tokens {
		if tok.Name == s.patName {
			// DELETE the token by id
			delURL := fmt.Sprintf("%s/rest/pat/latest/tokens/%d", baseURL.String(), tok.ID)
			delReq, err := http.NewRequest(http.MethodDelete, delURL, nil)
			if err != nil {
				return true, fmt.Errorf("failed to create PAT delete request: %w", err)
			}
			delReq.Header.Set("Authorization", "Basic "+cred)
			delReq.Header.Set("Accept", "application/json")

			delResp, err := s.httpClient.Do(delReq)
			if err != nil {
				return true, fmt.Errorf("jira PAT delete request failed: %w", err)
			}
			defer delResp.Body.Close()

			if delResp.StatusCode < 200 || delResp.StatusCode >= 300 {
				body, _ := io.ReadAll(delResp.Body)
				return true, fmt.Errorf("jira PAT delete failed: status=%d body=%s", delResp.StatusCode, string(body))
			}
			return true, nil
		}
	}

	return false, nil
}

// ensurePAT makes sure a valid PAT exists and is not expiring within one week.
// If no token exists or it's near expiration, it will create a new PAT using Basic auth.
func (s *JiraService) ensurePAT(baseURL *url.URL) error {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	// Renew if token missing or will expire within 7 days spare window
	if s.patToken != "" && time.Now().Before(s.patExpiry.Add(-7*24*time.Hour)) {
		return nil
	}

	return s.createPAT(baseURL)
}

// createPAT creates a new Personal Access Token via Jira PAT endpoint using Basic auth.
func (s *JiraService) createPAT(baseURL *url.URL) error {
	if s.cfg.JiraUser == "" || s.cfg.JiraPassword == "" {
		return apperrors.ErrJiraConfigMissing
	}

	// Build PAT creation URL
	patURL := baseURL.String() + "/rest/pat/latest/tokens"

	// Prepare request body
	type patCreateRequest struct {
		Name               string `json:"name"`
		ExpirationDuration int    `json:"expirationDuration"`
	}
	reqBody := patCreateRequest{
		Name:               s.patName,
		ExpirationDuration: 90, // 90 days as requested (number, not string)
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode PAT create body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, patURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create PAT HTTP request: %w", err)
	}
	// Basic auth with configured Jira credentials
	cred := base64.StdEncoding.EncodeToString([]byte(s.cfg.JiraUser + ":" + s.cfg.JiraPassword))
	req.Header.Set("Authorization", "Basic "+cred)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("jira PAT request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("jira PAT creation failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var patResp patTokenResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&patResp); err != nil {
		return fmt.Errorf("failed to decode PAT response: %w", err)
	}

	if patResp.RawToken == "" || patResp.ExpiringAt == "" {
		return fmt.Errorf("jira PAT response missing token or expiry")
	}

	// Parse expiry time (RFC3339 with fractional seconds)
	expiry, err := time.Parse(time.RFC3339Nano, patResp.ExpiringAt)
	if err != nil {
		return fmt.Errorf("failed to parse PAT expiringAt: %w", err)
	}

	s.patToken = patResp.RawToken
	s.patExpiry = expiry

	log.Printf("Jira PAT created: name=%s expires=%s", patResp.Name, patResp.ExpiringAt)
	return nil
}

type jiraSearchResponse struct {
	Total  int         `json:"total"`
	Issues []JiraIssue `json:"issues"`
}

// JiraIssue represents a Jira issue in search results
type JiraIssue struct {
	ID      string          `json:"id"`
	Key     string          `json:"key"`
	Fields  JiraIssueFields `json:"fields"`
	Project string          `json:"project,omitempty"`
	Link    string          `json:"link,omitempty"`
}

// JiraIssueFields represents the fields of a Jira issue
type JiraIssueFields struct {
	Summary     string          `json:"summary"`
	Status      JiraStatus      `json:"status"`
	IssueType   JiraIssueType   `json:"issuetype"`
	Priority    JiraPriority    `json:"priority,omitempty"`
	Assignee    *JiraUser       `json:"assignee,omitempty"`
	Reporter    *JiraUser       `json:"reporter,omitempty"`
	Created     string          `json:"created"`
	Updated     string          `json:"updated"`
	Description string          `json:"description,omitempty"`
	Labels      []string        `json:"labels,omitempty"`
	Components  []JiraComponent `json:"components,omitempty"`
}

// JiraStatus represents the status of a Jira issue
type JiraStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraIssueType represents the type of a Jira issue
type JiraIssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraPriority represents the priority of a Jira issue
type JiraPriority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraUser represents a Jira user
type JiraUser struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

// JiraComponent represents a Jira component
type JiraComponent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// JiraIssuesResponse represents the response for issue search endpoints
type JiraIssuesResponse struct {
	Issues []JiraIssue `json:"issues"`
	Total  int         `json:"total"`
	Page   int         `json:"page,omitempty"`
	Limit  int         `json:"limit,omitempty"`
}

// JiraIssueFilters represents the filters for Jira issue search
type JiraIssueFilters struct {
	Project  string // Real Jira project key (e.g., "SAPBTPCFS")
	Status   string // Real Jira status values (e.g., "Open,In Progress,Re Opened")
	Team     string // Team name for filtering
	User     string // Username (for user-specific searches)
	Date     string // Date for date-based filtering (yyyy-MM-dd format)
	Assignee string // Assignee username for filtering
	Type     string // Issue type (e.g., "Bug,Task,Story")
	Summary  string // Free text search in summary
	Key      string // Specific issue key (e.g., "BUG-1234")
	Page     int    // Page number for pagination (1-based)
	Limit    int    // Number of items per page (max 100)
}

// GetIssues returns Jira issues based on the provided filters.
func (s *JiraService) GetIssues(filters JiraIssueFilters) (*JiraIssuesResponse, error) {
	if s.cfg.JiraDomain == "" || s.cfg.JiraUser == "" || s.cfg.JiraPassword == "" {
		return nil, apperrors.ErrJiraConfigMissing
	}

	jql, err := s.buildJQL(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to build JQL query: %w", err)
	}

	return s.searchIssues(jql, filters, false)
}

// GetIssuesCount returns the count of Jira issues based on the provided filters.
func (s *JiraService) GetIssuesCount(filters JiraIssueFilters) (int, error) {
	if s.cfg.JiraDomain == "" || s.cfg.JiraUser == "" || s.cfg.JiraPassword == "" {
		return 0, apperrors.ErrJiraConfigMissing
	}

	jql, err := s.buildJQL(filters)
	if err != nil {
		return 0, fmt.Errorf("failed to build JQL query: %w", err)
	}

	response, err := s.searchIssues(jql, filters, true)
	if err != nil {
		return 0, err
	}
	return response.Total, nil
}

// buildJQL constructs the JQL query based on the provided filters with validation
func (s *JiraService) buildJQL(filters JiraIssueFilters) (string, error) {
	var conditions []string

	// Project filter - use real Jira project key with validation
	if filters.Project != "" {
		if err := s.validateJQLValue(filters.Project); err != nil {
			return "", fmt.Errorf("invalid project value: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf(`project = "%s"`, s.escapeJQLValue(filters.Project)))
	}

	// Status filter - use real Jira status values with validation
	if filters.Status != "" {
		// Handle comma-separated status values
		statusValues := strings.Split(filters.Status, ",")
		var validatedStatuses []string
		for _, status := range statusValues {
			status = strings.TrimSpace(status)
			if err := s.validateJQLValue(status); err != nil {
				return "", fmt.Errorf("invalid status value '%s': %w", status, err)
			}
			validatedStatuses = append(validatedStatuses, fmt.Sprintf(`"%s"`, s.escapeJQLValue(status)))
		}
		if len(validatedStatuses) == 1 {
			conditions = append(conditions, fmt.Sprintf(`status = %s`, validatedStatuses[0]))
		} else {
			conditions = append(conditions, fmt.Sprintf(`status IN (%s)`, strings.Join(validatedStatuses, ", ")))
		}
	}

	// Team filter with validation
	if filters.Team != "" {
		if err := s.validateJQLValue(filters.Team); err != nil {
			return "", fmt.Errorf("invalid team value: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf(`"Team(s)" = "%s"`, s.escapeJQLValue(filters.Team)))
	}

	// User filter with validation
	if filters.User != "" {
		if err := s.validateJQLValue(filters.User); err != nil {
			return "", fmt.Errorf("invalid user value: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf(`assignee = "%s"`, s.escapeJQLValue(filters.User)))
	}

	// Assignee filter with validation (separate from User filter)
	if filters.Assignee != "" {
		if err := s.validateJQLValue(filters.Assignee); err != nil {
			return "", fmt.Errorf("invalid assignee value: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf(`assignee = "%s"`, s.escapeJQLValue(filters.Assignee)))
	}

	// Type filter with validation (issue type)
	if filters.Type != "" {
		// Handle comma-separated type values
		typeValues := strings.Split(filters.Type, ",")
		var validatedTypes []string
		for _, issueType := range typeValues {
			issueType = strings.TrimSpace(issueType)
			if err := s.validateJQLValue(issueType); err != nil {
				return "", fmt.Errorf("invalid type value '%s': %w", issueType, err)
			}
			validatedTypes = append(validatedTypes, fmt.Sprintf(`"%s"`, s.escapeJQLValue(issueType)))
		}
		if len(validatedTypes) == 1 {
			conditions = append(conditions, fmt.Sprintf(`issuetype = %s`, validatedTypes[0]))
		} else {
			conditions = append(conditions, fmt.Sprintf(`issuetype IN (%s)`, strings.Join(validatedTypes, ", ")))
		}
	}

	// Summary filter with validation (text search)
	if filters.Summary != "" {
		if err := s.validateJQLValue(filters.Summary); err != nil {
			return "", fmt.Errorf("invalid summary value: %w", err)
		}
		// Use ~ for text search in summary field
		conditions = append(conditions, fmt.Sprintf(`summary ~ "%s"`, s.escapeJQLValue(filters.Summary)))
	}

	// Key filter with validation (specific issue key)
	if filters.Key != "" {
		if err := s.validateJQLValue(filters.Key); err != nil {
			return "", fmt.Errorf("invalid key value: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf(`key = "%s"`, s.escapeJQLValue(filters.Key)))
	}

	// Date filter for resolved issues with validation
	if filters.Date != "" && filters.User != "" && strings.Contains(strings.ToLower(filters.Status), "resolved") {
		if err := s.validateDateFormat(filters.Date); err != nil {
			return "", fmt.Errorf("invalid date format: %w", err)
		}
		conditions = append(conditions, fmt.Sprintf(`status CHANGED TO "resolved" BY "%s" AFTER %s`, s.escapeJQLValue(filters.User), filters.Date))
	}

	// Check if we have any search criteria at all
	if filters.Project == "" && filters.Status == "" && filters.Team == "" && filters.User == "" && 
	   filters.Assignee == "" && filters.Type == "" && filters.Summary == "" && filters.Key == "" {
		return "", fmt.Errorf("at least one search criterion must be provided (project, status, team, user, assignee, type, summary, or key)")
	}

	if len(conditions) == 0 {
		// This should not happen if validation above is correct, but keeping as safety check
		return "", fmt.Errorf("no valid search conditions generated")
	}

	jql := strings.Join(conditions, " AND ")

	// Add ordering for consistent pagination (part of JQL, not URL parameter)
	jql += " ORDER BY created DESC"

	// Validate final JQL length
	if len(jql) > 8000 { // Jira has a practical limit on JQL length
		return "", fmt.Errorf("generated JQL query is too long (%d characters, max 8000)", len(jql))
	}

	log.Printf("Generated JQL: %s", jql)
	return jql, nil
}

// validateJQLValue validates a value to be used in JQL queries
func (s *JiraService) validateJQLValue(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}

	// Check for potentially dangerous characters
	if strings.ContainsAny(value, "\n\r\t") {
		return fmt.Errorf("value contains invalid characters (newlines or tabs)")
	}

	// Check for excessively long values
	if len(value) > 255 {
		return fmt.Errorf("value is too long (max 255 characters)")
	}

	return nil
}

// escapeJQLValue escapes special characters in JQL values
func (s *JiraService) escapeJQLValue(value string) string {
	// Escape double quotes by doubling them
	return strings.ReplaceAll(value, `"`, `""`)
}

// validateDateFormat validates date format for JQL queries
func (s *JiraService) validateDateFormat(date string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("date must be in yyyy-MM-dd format")
	}
	return nil
}

// searchIssues is a helper method to perform Jira issue searches with pagination support
func (s *JiraService) searchIssues(jql string, filters JiraIssueFilters, countOnly bool) (*JiraIssuesResponse, error) {
	// Validate and parse base URL
	base := s.cfg.JiraDomain
	if base == "" {
		return nil, fmt.Errorf("jira domain is not configured")
	}

	// Ensure proper URL scheme
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}

	// Parse and validate the base URL
	baseURL, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid jira domain URL '%s': %w", base, err)
	}

	// Build request URL with proper URL encoding
	values := url.Values{}
	values.Set("jql", jql)

	if countOnly {
		values.Set("maxResults", "0") // Only get count
	} else {
		// Set pagination parameters
		limit := filters.Limit
		if limit <= 0 {
			limit = 50 // Default limit
		}
		if limit > 100 {
			limit = 100 // Max limit
		}

		page := filters.Page
		if page <= 0 {
			page = 1 // Default page
		}

		// Calculate startAt for Jira API (0-based)
		startAt := (page - 1) * limit

		values.Set("maxResults", fmt.Sprintf("%d", limit))
		values.Set("startAt", fmt.Sprintf("%d", startAt))

		// Optimize field selection for better performance
		values.Set("fields", "key,summary,status,issuetype,priority,assignee,created,updated")
	}

	// Construct the full URL safely
	searchPath := "/rest/api/2/search"
	fullURL := baseURL.String() + searchPath + "?" + values.Encode()

	// Validate final URL length (browsers and servers have limits)
	if len(fullURL) > 2048 {
		return nil, fmt.Errorf("constructed URL is too long (%d characters, max 2048)", len(fullURL))
	}

	log.Printf("Jira search URL: %s", fullURL)

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Ensure PAT exists and is valid (renew if expiring within 7 days)
	if err := s.ensurePAT(baseURL); err != nil {
		return nil, fmt.Errorf("failed to ensure Jira PAT: %w", err)
	}

	// Use PAT as Bearer token for Jira interactions
	req.Header.Set("Authorization", "Bearer "+s.patToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira search failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var parsed jiraSearchResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}

	// Enhance issues with project and link information
	enhancedIssues := make([]JiraIssue, len(parsed.Issues))
	for i, issue := range parsed.Issues {
		enhancedIssues[i] = issue

		// Extract project key from issue key (e.g., "SAPBTPCFS-123" -> "SAPBTPCFS")
		if issue.Key != "" {
			parts := strings.Split(issue.Key, "-")
			if len(parts) > 0 {
				enhancedIssues[i].Project = parts[0]
			}
		}

		// Generate issue link
		if issue.Key != "" {
			enhancedIssues[i].Link = fmt.Sprintf("%s/browse/%s", baseURL.String(), issue.Key)
		}
	}

	response := &JiraIssuesResponse{
		Issues: enhancedIssues,
		Total:  parsed.Total,
	}

	// Add pagination metadata for non-count queries
	if !countOnly {
		limit := filters.Limit
		if limit <= 0 {
			limit = 50
		}
		if limit > 100 {
			limit = 100
		}

		page := filters.Page
		if page <= 0 {
			page = 1
		}

		response.Page = page
		response.Limit = limit
	}

	return response, nil
}
