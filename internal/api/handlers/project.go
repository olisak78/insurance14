package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"developer-portal-backend/internal/database/models"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectHandler handles HTTP requests for project operations
type ProjectHandler struct {
	projectService *service.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProject handles POST /projects
// @Summary Create a new project
// @Description Create a new project within an organization
// @Tags projects
// @Accept json
// @Produce json
// @Param project body service.CreateProjectRequest true "Project data"
// @Success 201 {object} service.ProjectResponse "Successfully created project"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Organization or team not found"
// @Failure 409 {object} map[string]interface{} "Project already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req service.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.Create(&req)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) || errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrProjectExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProject handles GET /projects/:id
// @Summary Get project by ID
// @Description Get a specific project by its UUID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Success 200 {object} service.ProjectResponse "Successfully retrieved project"
// @Failure 400 {object} map[string]interface{} "Invalid project ID"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.projectService.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject handles PUT /projects/:id
// @Summary Update project
// @Description Update an existing project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Param project body service.UpdateProjectRequest true "Updated project data"
// @Success 200 {object} service.ProjectResponse "Successfully updated project"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Project or team not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var req service.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.Update(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectNotFound) || errors.Is(err, apperrors.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject handles DELETE /projects/:id
// @Summary Delete project
// @Description Delete a project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID (UUID)"
// @Success 204 "Successfully deleted project"
// @Failure 400 {object} map[string]interface{} "Invalid project ID"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	if err := h.projectService.Delete(id); err != nil {
		if errors.Is(err, apperrors.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListProjects handles GET /projects (requires organization_id parameter)
func (h *ProjectHandler) ListProjects(c *gin.Context) {
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

	var projects *service.ProjectListResponse
	if search != "" {
		projects, err = h.projectService.Search(organizationID, search, page, pageSize)
	} else {
		projects, err = h.projectService.GetByOrganization(organizationID, page, pageSize)
	}

	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProjectsByOrganization handles GET /organizations/:orgId/projects
func (h *ProjectHandler) GetProjectsByOrganization(c *gin.Context) {
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

	projects, err := h.projectService.GetByOrganization(orgID, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProjectsByStatus handles GET /projects/status/:status
func (h *ProjectHandler) GetProjectsByStatus(c *gin.Context) {
	statusStr := c.Param("status")
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

	// Convert string to ProjectStatus
	var status models.ProjectStatus
	switch statusStr {
	case "active":
		status = models.ProjectStatusActive
	case "inactive":
		status = models.ProjectStatusInactive
	case "archived":
		status = models.ProjectStatusArchived
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status. Valid values: active, inactive, archived"})
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

	projects, err := h.projectService.GetByStatus(organizationID, status, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProjectWithOrganization handles GET /projects/:id/organization
func (h *ProjectHandler) GetProjectWithOrganization(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.projectService.GetWithOrganization(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// GetProjectWithComponents handles GET /projects/:id/components
func (h *ProjectHandler) GetProjectWithComponents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.projectService.GetWithComponents(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// GetProjectWithLandscapes handles GET /projects/:id/landscapes
func (h *ProjectHandler) GetProjectWithLandscapes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.projectService.GetWithLandscapes(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}
