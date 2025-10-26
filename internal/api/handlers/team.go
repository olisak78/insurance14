package handlers

import (
	"errors"
	"net/http"
	"strconv"

	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TeamHandler handles HTTP requests for team operations
type TeamHandler struct {
	teamService service.TeamServiceInterface
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamService service.TeamServiceInterface) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// CreateTeam handles POST /teams
// @Summary Create a new team
// @Description Create a new team with the provided details
// @Tags teams
// @Accept json
// @Produce json
// @Param team body service.CreateTeamRequest true "Team data"
// @Success 201 {object} service.TeamResponse "Successfully created team"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Organization or leader not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req service.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamService.Create(&req)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrLeaderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, team)
}

// GetTeam handles GET /teams/:id
// @Summary Get team by ID
// @Description Get a specific team by its UUID
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Success 200 {object} service.TeamResponse "Successfully retrieved team"
// @Failure 400 {object} map[string]interface{} "Invalid team ID"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id} [get]
func (h *TeamHandler) GetTeam(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	team, err := h.teamService.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// GetAllTeams handles GET /teams (optional organization_id parameter)
// @Summary List all teams
// @Description Get all teams with optional organization filtering and pagination
// @Tags teams
// @Accept json
// @Produce json
// @Param organization_id query string false "Organization ID (UUID) to filter teams"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} service.TeamListResponse "Successfully retrieved teams"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams [get]
func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	organizationIDStr := c.Query("organization_id")
	var organizationID *uuid.UUID

	if organizationIDStr != "" {
		id, err := uuid.Parse(organizationIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
			return
		}
		organizationID = &id
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	teams, err := h.teamService.GetAllTeams(organizationID, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// GetAllTeamsDeprecated handles GET /teams/all (deprecated, kept for backward compatibility)
// This endpoint is hidden from Swagger documentation. Use GET /teams instead.
func (h *TeamHandler) GetAllTeamsDeprecated(c *gin.Context) {
	// Add deprecation headers following RFC 8594
	c.Header("Deprecation", "true")
	c.Header("X-API-Warn", "This endpoint is deprecated. Please use GET /api/v1/teams instead.")

	// Reuse the same logic as GetAllTeams
	h.GetAllTeams(c)
}

// GetTeamByName handles GET /teams/by-name/:name (requires organization_id parameter)
func (h *TeamHandler) GetTeamByName(c *gin.Context) {
	teamName := c.Param("name")
	organizationIDStr := c.Query("organization_id")

	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id parameter is required"})
		return
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	team, err := h.teamService.GetByName(organizationID, teamName)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// GetTeamMembersByName handles GET /teams/by-name/:name/members (requires organization_id parameter)
func (h *TeamHandler) GetTeamMembersByName(c *gin.Context) {
	teamName := c.Param("name")
	organizationIDStr := c.Query("organization_id")

	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id parameter is required"})
		return
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	members, total, err := h.teamService.GetTeamMembersByName(organizationID, teamName, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"members":   members,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetTeamComponentsByName handles GET /teams/by-name/:name/components (requires organization_id parameter)
// DEPRECATED: Use GET /teams/:id/components instead
func (h *TeamHandler) GetTeamComponentsByName(c *gin.Context) {
	teamName := c.Param("name")
	organizationIDStr := c.Query("organization_id")

	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id parameter is required"})
		return
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	components, total, err := h.teamService.GetTeamComponentsByName(organizationID, teamName, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"components": components,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetTeamComponents handles GET /teams/:id/components
// @Summary Get components owned by team
// @Description Get all components owned by a team identified by ID. This is a cleaner endpoint that uses team ID instead of name
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} map[string]interface{} "Successfully retrieved team components"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id}/components [get]
func (h *TeamHandler) GetTeamComponents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	components, total, err := h.teamService.GetTeamComponentsByID(id, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		// Log the full error for debugging
		c.Error(err) // This adds the error to Gin's error context for logging middleware
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"components": components,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTeam handles PUT /teams/:id
// @Summary Update team
// @Description Update an existing team by ID
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Param team body service.UpdateTeamRequest true "Updated team data"
// @Success 200 {object} service.TeamResponse "Successfully updated team"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Team or leader not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id} [put]
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var req service.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamService.Update(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrLeaderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// DeleteTeam handles DELETE /teams/:id
// @Summary Delete team
// @Description Delete a team by ID
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Success 204 "Successfully deleted team"
// @Failure 400 {object} map[string]interface{} "Invalid team ID"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id} [delete]
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	if err := h.teamService.Delete(id); err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListTeams handles GET /teams (requires organization_id parameter)
func (h *TeamHandler) ListTeams(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	search := c.Query("search")
	organizationIDStr := c.Query("organization_id")

	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id parameter is required"})
		return
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var teams *service.TeamListResponse
	if search != "" {
		teams, err = h.teamService.Search(organizationID, search, page, pageSize)
	} else {
		teams, err = h.teamService.GetByOrganization(organizationID, page, pageSize)
	}

	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// GetTeamsByOrganization handles GET /organizations/:orgId/teams
func (h *TeamHandler) GetTeamsByOrganization(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	teams, err := h.teamService.GetByOrganization(orgID, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, teams)
}

// GetTeamWithMembers handles GET /teams/:id/members
// @Summary Get team members
// @Description Get all members of a specific team by team ID
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Success 200 {array} service.MemberResponse "Successfully retrieved team members"
// @Failure 400 {object} map[string]interface{} "Invalid team ID"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id}/members [get]
func (h *TeamHandler) GetTeamWithMembers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	members, err := h.teamService.GetMembersOnly(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}

// GetTeamWithProjects handles GET /teams/:id/projects
func (h *TeamHandler) GetTeamWithProjects(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	team, err := h.teamService.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// GetTeamWithComponents handles GET /teams/:id/components
func (h *TeamHandler) GetTeamWithComponents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	team, err := h.teamService.GetWithComponentOwnerships(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// GetTeamWithDutySchedules handles GET /teams/:id/duty-schedules
func (h *TeamHandler) GetTeamWithDutySchedules(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	team, err := h.teamService.GetWithDutySchedules(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// GetTeamWithLeader handles GET /teams/:id/leader
func (h *TeamHandler) GetTeamWithLeader(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	team, err := h.teamService.GetTeamLead(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// AddLink handles POST /teams/:id/links
// @Summary Add a link to a team
// @Description Add a new link to a team's links array. Links support url, title, icon, and category fields.
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Param link body service.AddLinkRequest true "Link data"
// @Success 200 {object} service.TeamResponse "Successfully added link"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 409 {object} map[string]interface{} "Link with URL already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id}/links [post]
func (h *TeamHandler) AddLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var req service.AddLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamService.AddLink(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrLinkExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// RemoveLink handles DELETE /teams/:id/links
// @Summary Remove a link from a team
// @Description Remove a link from a team's links array by URL
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Param url query string true "Link URL to remove"
// @Success 200 {object} service.TeamResponse "Successfully removed link"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Team or link not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id}/links [delete]
func (h *TeamHandler) RemoveLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	linkURL := c.Query("url")
	if linkURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}

	team, err := h.teamService.RemoveLink(id, linkURL)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrLinkNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

// UpdateLinks handles PUT /teams/:id/links
// @Summary Update all links for a team
// @Description Replace all links for a team with the provided list
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Param links body service.UpdateLinksRequest true "Links data"
// @Success 200 {object} service.TeamResponse "Successfully updated links"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id}/links [put]
func (h *TeamHandler) UpdateLinks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var req service.UpdateLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamService.UpdateLinks(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}
