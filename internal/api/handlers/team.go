package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"developer-portal-backend/internal/auth"
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

// GetAllTeams handles GET /teams with optional team-name or team-id filters
// @Summary List teams or fetch a specific team by name/ID
// @Description
// - Without query: returns a list of teams (id, group_id, name, title, description, picture_url)
// - With team-id (UUID): returns a single team enriched with members and links (and jira_team extracted from metadata)
// - With team-name (string): same as team-id variant, by simple name (global)
// @Tags teams
// @Accept json
// @Produce json
// @Param team-name query string false "Team simple name; when provided returns team with members and links"
// @Param team-id query string false "Team ID (UUID); when provided returns team with members and links"
// @Success 200 {object} map[string]interface{} "Teams list or single team response"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams [get]
func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	// 1) team-id branch: resolve ID -> name, then return enriched team by name
	if teamIDStr := c.Query("team-id"); teamIDStr != "" {
		id, err := uuid.Parse(teamIDStr)
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

		// Use name-based retrieval to enrich with members and links, marking favorites for viewer when available
		viewerName, _ := auth.GetUsername(c)
		var teamWithMembers *service.TeamWithMembersResponse
		if viewerName != "" {
			teamWithMembers, err = h.teamService.GetBySimpleNameWithViewer(team.Name, viewerName)
		} else {
			teamWithMembers, err = h.teamService.GetBySimpleName(team.Name)
		}
		if err != nil {
			if errors.Is(err, apperrors.ErrTeamNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform response: extract metadata.jira.team into jira_team and parse metadata for response
		var jiraTeam string
		var metadata map[string]interface{}
		if len(teamWithMembers.TeamResponse.Metadata) > 0 {
			if err := json.Unmarshal(teamWithMembers.TeamResponse.Metadata, &metadata); err == nil && metadata != nil {
				if jiraVal, ok := metadata["jira"].(map[string]interface{}); ok && jiraVal != nil {
					if jt, ok := jiraVal["team"].(string); ok {
						jiraTeam = jt
					}
				}
			}
		}
		resp := gin.H{
			"id":              teamWithMembers.TeamResponse.ID,
			"group_id":        teamWithMembers.TeamResponse.GroupID,
			"organization_id": teamWithMembers.TeamResponse.OrganizationID,
			"name":            teamWithMembers.TeamResponse.Name,
			"title":           teamWithMembers.TeamResponse.Title,
			"description":     teamWithMembers.TeamResponse.Description,
			"owner":           teamWithMembers.TeamResponse.Owner,
			"email":           teamWithMembers.TeamResponse.Email,
			"picture_url":     teamWithMembers.TeamResponse.PictureURL,
			"created_at":      teamWithMembers.TeamResponse.CreatedAt,
			"updated_at":      teamWithMembers.TeamResponse.UpdatedAt,
			"jira_team":       jiraTeam,
			"metadata":        metadata,
			"members":         teamWithMembers.Members,
			"links":           teamWithMembers.Links,
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	// 2) team-name branch: return enriched team by simple name (global)
	if teamName := c.Query("team-name"); teamName != "" {
		viewerName, _ := auth.GetUsername(c)
		var teamWithMembers *service.TeamWithMembersResponse
		var err error
		if viewerName != "" {
			teamWithMembers, err = h.teamService.GetBySimpleNameWithViewer(teamName, viewerName)
		} else {
			teamWithMembers, err = h.teamService.GetBySimpleName(teamName)
		}
		if err != nil {
			if errors.Is(err, apperrors.ErrTeamNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform: extract metadata.jira.team into jira_team and parse metadata for response
		var jiraTeam string
		var metadata map[string]interface{}
		if len(teamWithMembers.TeamResponse.Metadata) > 0 {
			if err := json.Unmarshal(teamWithMembers.TeamResponse.Metadata, &metadata); err == nil && metadata != nil {
				if jiraVal, ok := metadata["jira"].(map[string]interface{}); ok && jiraVal != nil {
					if jt, ok := jiraVal["team"].(string); ok {
						jiraTeam = jt
					}
				}
			}
		}
		resp := gin.H{
			"id":              teamWithMembers.TeamResponse.ID,
			"group_id":        teamWithMembers.TeamResponse.GroupID,
			"organization_id": teamWithMembers.TeamResponse.OrganizationID,
			"name":            teamWithMembers.TeamResponse.Name,
			"title":           teamWithMembers.TeamResponse.Title,
			"description":     teamWithMembers.TeamResponse.Description,
			"owner":           teamWithMembers.TeamResponse.Owner,
			"email":           teamWithMembers.TeamResponse.Email,
			"picture_url":     teamWithMembers.TeamResponse.PictureURL,
			"created_at":      teamWithMembers.TeamResponse.CreatedAt,
			"updated_at":      teamWithMembers.TeamResponse.UpdatedAt,
			"jira_team":       jiraTeam,
			"metadata":        metadata,
			"members":         teamWithMembers.Members,
			"links":           teamWithMembers.Links,
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	// 3) Default: list teams (global)
	teams, err := h.teamService.GetAllTeams(nil, 1, 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]gin.H, 0, len(teams.Teams))
	for _, t := range teams.Teams {
		// Parse metadata to include in response
		var metadata map[string]interface{}
		if len(t.Metadata) > 0 {
			_ = json.Unmarshal(t.Metadata, &metadata)
		}

		items = append(items, gin.H{
			"id":          t.ID,
			"group_id":    t.GroupID,
			"name":        t.Name,
			"title":       t.Title,
			"description": t.Description,
			"picture_url": t.PictureURL,
			"metadata":    metadata,
		})
	}
	response := gin.H{
		"teams":     items,
		"total":     teams.Total,
		"page":      teams.Page,
		"page_size": teams.PageSize,
	}
	c.JSON(http.StatusOK, response)
}

// UpdateTeamMetadataRequest represents a request to update team metadata
type UpdateTeamMetadataRequest struct {
	Metadata map[string]interface{} `json:"metadata" binding:"required"`
}

// UpdateTeamMetadata handles PATCH /teams/:id/metadata
// @Summary Update team metadata (merge, not replace)
// @Description Merges the provided metadata fields with existing team metadata. Only updates/adds the fields provided, preserves other existing fields (e.g., updating color preserves jira config)
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Param request body UpdateTeamMetadataRequest true "Metadata fields to update/add"
// @Success 200 {object} map[string]interface{} "Updated team with merged metadata"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /teams/{id}/metadata [patch]
func (h *TeamHandler) UpdateTeamMetadata(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}

	var req UpdateTeamMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert metadata map to JSON
	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata format"})
		return
	}

	// Update team metadata via service
	updatedTeam, err := h.teamService.UpdateTeamMetadata(teamID, metadataJSON)
	if err != nil {
		if errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse metadata back for response
	var metadata map[string]interface{}
	if len(updatedTeam.Metadata) > 0 {
		_ = json.Unmarshal(updatedTeam.Metadata, &metadata)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          updatedTeam.ID,
		"name":        updatedTeam.Name,
		"title":       updatedTeam.Title,
		"description": updatedTeam.Description,
		"metadata":    metadata,
		"updated_at":  updatedTeam.UpdatedAt,
	})
}
