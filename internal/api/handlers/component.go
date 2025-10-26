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

// CreateComponent handles POST /components
// @Summary Create a new component
// @Description Create a new component within an organization
// @Tags components
// @Accept json
// @Produce json
// @Param component body service.CreateComponentRequest true "Component data"
// @Success 201 {object} service.ComponentResponse "Successfully created component"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 409 {object} map[string]interface{} "Component already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components [post]
func (h *ComponentHandler) CreateComponent(c *gin.Context) {
	var req service.CreateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	component, err := h.componentService.Create(&req)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrComponentExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, component)
}

// GetComponent handles GET /components/:id
// @Summary Get component by ID
// @Description Get a specific component by its UUID
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID (UUID)"
// @Success 200 {object} service.ComponentResponse "Successfully retrieved component"
// @Failure 400 {object} map[string]interface{} "Invalid component ID"
// @Failure 404 {object} map[string]interface{} "Component not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components/{id} [get]
func (h *ComponentHandler) GetComponent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
		return
	}

	component, err := h.componentService.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, component)
}

// GetComponentByName handles GET /components/by-name/:name (requires organization_id parameter)
// @Summary Get component by name
// @Description Get a component by its name within a specific organization
// @Tags components
// @Accept json
// @Produce json
// @Param name path string true "Component name"
// @Param organization_id query string true "Organization ID (UUID)"
// @Success 200 {object} service.ComponentResponse "Successfully retrieved component"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Component not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components/by-name/{name} [get]
func (h *ComponentHandler) GetComponentByName(c *gin.Context) {
	componentName := c.Param("name")
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

	component, err := h.componentService.GetByName(organizationID, componentName)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, component)
}

// UpdateComponent handles PUT /components/:id
// @Summary Update component
// @Description Update an existing component by ID
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID (UUID)"
// @Param component body service.UpdateComponentRequest true "Updated component data"
// @Success 200 {object} service.ComponentResponse "Successfully updated component"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Component not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components/{id} [put]
func (h *ComponentHandler) UpdateComponent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
		return
	}

	var req service.UpdateComponentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	component, err := h.componentService.Update(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, component)
}

// DeleteComponent handles DELETE /components/:id
// @Summary Delete component
// @Description Delete a component by ID
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID (UUID)"
// @Success 204 "Successfully deleted component"
// @Failure 400 {object} map[string]interface{} "Invalid component ID"
// @Failure 404 {object} map[string]interface{} "Component not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components/{id} [delete]
func (h *ComponentHandler) DeleteComponent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
		return
	}

	if err := h.componentService.Delete(id); err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListComponents handles GET /components (requires organization_id parameter)
// @Summary List components
// @Description Get all components in an organization with optional search and pagination
// @Tags components
// @Accept json
// @Produce json
// @Param organization_id query string true "Organization ID (UUID)"
// @Param search query string false "Search term for component name or description"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} service.ComponentListResponse "Successfully retrieved components"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components [get]
func (h *ComponentHandler) ListComponents(c *gin.Context) {
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

	var components *service.ComponentListResponse
	if search != "" {
		components, err = h.componentService.Search(organizationID, search, page, pageSize)
	} else {
		components, err = h.componentService.GetByOrganization(organizationID, page, pageSize)
	}

	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, components)
}

// GetComponentsByOrganization handles GET /organizations/:orgId/components
func (h *ComponentHandler) GetComponentsByOrganization(c *gin.Context) {
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

	components, err := h.componentService.GetByOrganization(orgID, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, components)
}

// GetComponentWithOwnerships handles GET /components/:id/ownerships
func (h *ComponentHandler) GetComponentWithOwnerships(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
		return
	}

	component, err := h.componentService.GetWithTeamOwnerships(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, component)
}

// GetComponentWithDeployments handles GET /components/:id/deployments
func (h *ComponentHandler) GetComponentWithDeployments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
		return
	}

	component, err := h.componentService.GetWithDeployments(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, component)
}

// GetComponentWithProjects handles GET /components/:id/projects
func (h *ComponentHandler) GetComponentWithProjects(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
		return
	}

	component, err := h.componentService.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, component)
}

// GetComponentWithFullDetails handles GET /components/:id/details
func (h *ComponentHandler) GetComponentWithFullDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
		return
	}

	component, err := h.componentService.GetWithFullDetails(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, component)
}

// GetComponentsByTeamName handles GET /components/by-team/:teamName (requires organization_id parameter)
// DEPRECATED: Use GET /components/by-team/:id instead
func (h *ComponentHandler) GetComponentsByTeamName(c *gin.Context) {
	teamName := c.Param("teamName")
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

// GetComponentsByTeamID handles GET /components/by-team/:id
// @Summary Get components by team ID
// @Description Get all components owned by a team identified by ID. This is a cleaner endpoint that uses team ID instead of name
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Team ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} map[string]interface{} "Successfully retrieved components"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /components/by-team/{id} [get]
func (h *ComponentHandler) GetComponentsByTeamID(c *gin.Context) {
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
