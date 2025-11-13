package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ComponentHandler handles HTTP requests for component operations
type ComponentHandler struct {
	componentService *service.ComponentService
	teamService      service.TeamServiceInterface
}

// NewComponentHandler creates a new component handler
func NewComponentHandler(componentService *service.ComponentService, teamService service.TeamServiceInterface) *ComponentHandler {
	return &ComponentHandler{
		componentService: componentService,
		teamService:      teamService,
	}
}

// ListComponents handles GET /components
// @Summary List components
// @Description List components filtered by either team-id or project-name. Returns an array of minimal component views. One of team-id or project-name is required.
// @Tags components
// @Accept json
// @Produce json
// @Param team-id query string false "Team ID (UUID) to filter by owner_id"
// @Param project-name query string false "Project name"
// @Success 200 {array} object "Successfully retrieved components"
// @Failure 400 {object} map[string]interface{} "team-id or project-name parameter is required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components [get]
func (h *ComponentHandler) ListComponents(c *gin.Context) {
	projectName := c.Query("project-name")

	// If team-id is provided, return components owned by that team (uses pagination)
	teamIDStr := c.Query("team-id")
	if teamIDStr != "" {
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
			return
		}
		components, _, err := h.teamService.GetTeamComponentsByID(teamID, 1, 1000000)
		if err != nil {
			if errors.Is(err, apperrors.ErrTeamNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Build minimal view items with project info (same fields as project-name view plus project_id and project_title)
		items := make([]gin.H, len(components))
		for i, c := range components {
			// Extract qos, sonar, github from metadata if present
			var qos, sonar, github string
			if len(c.Metadata) > 0 {
				var meta map[string]interface{}
				if err := json.Unmarshal(c.Metadata, &meta); err == nil {
					// qos from metadata.ci.qos
					if ciRaw, ok := meta["ci"]; ok {
						if ciMap, ok := ciRaw.(map[string]interface{}); ok {
							if qosRaw, ok := ciMap["qos"]; ok {
								if qosStr, ok := qosRaw.(string); ok {
									qos = qosStr
								}
							}
						}
					}
					// sonar from metadata.sonar.project_id
					if sonarRaw, ok := meta["sonar"]; ok {
						if sonarMap, ok := sonarRaw.(map[string]interface{}); ok {
							if pidRaw, ok := sonarMap["project_id"]; ok {
								if pidStr, ok := pidRaw.(string); ok {
									sonar = "https://sonar.tools.sap/dashboard?id=" + pidStr
								}
							}
						}
					}
					// github from metadata.github.url
					if ghRaw, ok := meta["github"]; ok {
						if ghMap, ok := ghRaw.(map[string]interface{}); ok {
							if urlRaw, ok := ghMap["url"]; ok {
								if urlStr, ok := urlRaw.(string); ok {
									github = urlStr
								}
							}
						}
					}
				}
			}
			// Fetch project title (non-fatal if not found)
			projectTitle := ""
			if title, err := h.componentService.GetProjectTitleByID(c.ProjectID); err == nil {
				projectTitle = title
			}

			items[i] = gin.H{
				"id":            c.ID,
				"owner_id":      c.OwnerID,
				"name":          c.Name,
				"title":         c.Title,
				"description":   c.Description,
				"qos":           qos,
				"sonar":         sonar,
				"github":        github,
				"project_id":    c.ProjectID,
				"project_title": projectTitle,
			}
		}
		c.JSON(http.StatusOK, items)
		return
	}

	// If project-name is provided, return ALL components for the project (unpaginated minimal view) and ignore organization_id requirement
	if projectName != "" {
		views, err := h.componentService.GetByProjectNameAllView(projectName)
		if err != nil {
			if errors.Is(err, apperrors.ErrProjectNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, views)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "team-id or project-name parameter is required"})
	return
}
