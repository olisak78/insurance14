package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"developer-portal-backend/internal/auth"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// GitHubHandler handles GitHub-related HTTP requests
type GitHubHandler struct {
	service service.GitHubServiceInterface
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler(s service.GitHubServiceInterface) *GitHubHandler {
	return &GitHubHandler{service: s}
}

// GetMyPullRequests returns all pull requests created by the authenticated user
// @Summary Get my pull requests
// @Description Returns all pull requests created by the authenticated user across all repositories they have access to
// @Tags github
// @Produce json
// @Param state query string false "Filter by state: open, closed, all" default(open)
// @Param sort query string false "Sort by: created, updated, popularity, long-running" default(created)
// @Param direction query string false "Sort direction: asc, desc" default(desc)
// @Param per_page query int false "Results per page (1-100)" default(30)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} service.PullRequestsResponse
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 502 {object} ErrorResponse "GitHub API error"
// @Security BearerAuth
// @Router /github/pull-requests [get]
func (h *GitHubHandler) GetMyPullRequests(c *gin.Context) {
	// Get authenticated user claims from context (set by auth middleware)
	claimsInterface, exists := c.Get("auth_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	claims, ok := claimsInterface.(*auth.AuthClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authentication claims"})
		return
	}

	// Get query parameters
	state := c.DefaultQuery("state", "open")
	sort := c.DefaultQuery("sort", "created")
	direction := c.DefaultQuery("direction", "desc")

	perPageStr := c.DefaultQuery("per_page", "30")
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage <= 0 || perPage > 100 {
		perPage = 30
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	// Validate state parameter
	if state != "open" && state != "closed" && state != "all" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter. Must be: open, closed, or all"})
		return
	}

	// Validate sort parameter
	validSorts := map[string]bool{
		"created":      true,
		"updated":      true,
		"popularity":   true,
		"long-running": true,
	}
	if !validSorts[sort] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort parameter. Must be: created, updated, popularity, or long-running"})
		return
	}

	// Validate direction parameter
	if direction != "asc" && direction != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid direction parameter. Must be: asc or desc"})
		return
	}

	// Call service to get pull requests
	response, err := h.service.GetUserOpenPullRequests(c.Request.Context(), claims, state, sort, direction, perPage, page)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, apperrors.ErrGitHubAPIRateLimitExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch pull requests: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserTotalContributions returns the total contributions count for the authenticated user
// @Summary Get user total contributions
// @Description Returns the total number of contributions made by the authenticated user. If no period specified, uses GitHub's default (last year based on user's timezone). Uses GitHub GraphQL API to fetch contribution data.
// @Tags github
// @Produce json
// @Param period query string false "Time period in days (e.g., '30d', '90d', '365d'). If omitted, uses GitHub's default period. Maximum: 365 days"
// @Success 200 {object} service.TotalContributionsResponse
// @Failure 400 {object} ErrorResponse "Invalid period parameter"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 429 {object} ErrorResponse "Rate limit exceeded"
// @Failure 502 {object} ErrorResponse "GitHub API error"
// @Security BearerAuth
// @Router /github/contributions [get]
func (h *GitHubHandler) GetUserTotalContributions(c *gin.Context) {
	// Get authenticated user claims from context (set by auth middleware)
	claimsInterface, exists := c.Get("auth_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	claims, ok := claimsInterface.(*auth.AuthClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authentication claims"})
		return
	}

	// Get query parameter for period (empty = use GitHub's default)
	period := c.Query("period")

	// Call service to get total contributions
	response, err := h.service.GetUserTotalContributions(c.Request.Context(), claims, period)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, apperrors.ErrGitHubAPIRateLimitExceeded) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		// Check if it's a validation error (invalid period format)
		if strings.Contains(err.Error(), "invalid period format") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to fetch contributions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
