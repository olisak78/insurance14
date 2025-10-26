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

// LandscapeHandler handles HTTP requests for landscape operations
type LandscapeHandler struct {
	landscapeService *service.LandscapeService
}

// NewLandscapeHandler creates a new landscape handler
func NewLandscapeHandler(landscapeService *service.LandscapeService) *LandscapeHandler {
	return &LandscapeHandler{
		landscapeService: landscapeService,
	}
}

// CreateLandscape handles POST /landscapes
// @Summary Create a new landscape
// @Description Create a new landscape within an organization
// @Tags landscapes
// @Accept json
// @Produce json
// @Param landscape body service.CreateLandscapeRequest true "Landscape data"
// @Success 201 {object} service.LandscapeResponse "Successfully created landscape"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 409 {object} map[string]interface{} "Landscape already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes [post]
func (h *LandscapeHandler) CreateLandscape(c *gin.Context) {
	var req service.CreateLandscapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	landscape, err := h.landscapeService.Create(&req)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrLandscapeExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, landscape)
}

// GetLandscape handles GET /landscapes/:id
// @Summary Get landscape by ID
// @Description Get a specific landscape by its UUID
// @Tags landscapes
// @Accept json
// @Produce json
// @Param id path string true "Landscape ID (UUID)"
// @Success 200 {object} service.LandscapeResponse "Successfully retrieved landscape"
// @Failure 400 {object} map[string]interface{} "Invalid landscape ID"
// @Failure 404 {object} map[string]interface{} "Landscape not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes/{id} [get]
func (h *LandscapeHandler) GetLandscape(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
		return
	}

	landscape, err := h.landscapeService.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrLandscapeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscape)
}

// UpdateLandscape handles PUT /landscapes/:id
// @Summary Update landscape
// @Description Update an existing landscape by ID
// @Tags landscapes
// @Accept json
// @Produce json
// @Param id path string true "Landscape ID (UUID)"
// @Param landscape body service.UpdateLandscapeRequest true "Updated landscape data"
// @Success 200 {object} service.LandscapeResponse "Successfully updated landscape"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Landscape not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes/{id} [put]
func (h *LandscapeHandler) UpdateLandscape(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
		return
	}

	var req service.UpdateLandscapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	landscape, err := h.landscapeService.Update(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrLandscapeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscape)
}

// DeleteLandscape handles DELETE /landscapes/:id
// @Summary Delete landscape
// @Description Delete a landscape by ID
// @Tags landscapes
// @Accept json
// @Produce json
// @Param id path string true "Landscape ID (UUID)"
// @Success 204 "Successfully deleted landscape"
// @Failure 400 {object} map[string]interface{} "Invalid landscape ID"
// @Failure 404 {object} map[string]interface{} "Landscape not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes/{id} [delete]
func (h *LandscapeHandler) DeleteLandscape(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
		return
	}

	if err := h.landscapeService.Delete(id); err != nil {
		if errors.Is(err, apperrors.ErrLandscapeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListLandscapes handles GET /landscapes (requires organization_id parameter)
// @Summary List landscapes
// @Description Get all landscapes in an organization with optional search and pagination
// @Tags landscapes
// @Accept json
// @Produce json
// @Param organization_id query string true "Organization ID (UUID)"
// @Param search query string false "Search term for landscape name or description"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} service.LandscapeListResponse "Successfully retrieved landscapes"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes [get]
func (h *LandscapeHandler) ListLandscapes(c *gin.Context) {
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

	var landscapes *service.LandscapeListResponse
	if search != "" {
		landscapes, err = h.landscapeService.Search(organizationID, search, page, pageSize)
	} else {
		landscapes, err = h.landscapeService.GetByOrganization(organizationID, page, pageSize)
	}

	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscapes)
}

// GetLandscapesByOrganization handles GET /organizations/:orgId/landscapes
func (h *LandscapeHandler) GetLandscapesByOrganization(c *gin.Context) {
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

	landscapes, err := h.landscapeService.GetByOrganization(orgID, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscapes)
}

// GetLandscapesByEnvironment handles GET /landscapes/environment/:environment
// @Summary Get landscapes by environment type
// @Description Get landscapes filtered by environment type (development, staging, production, testing, preview)
// @Tags landscapes
// @Accept json
// @Produce json
// @Param environment path string true "Environment type" Enums(development, staging, production, testing, preview)
// @Param organization_id query string true "Organization ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} service.LandscapeListResponse "Successfully retrieved landscapes"
// @Failure 400 {object} map[string]interface{} "Invalid parameters or environment type"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes/environment/{environment} [get]
func (h *LandscapeHandler) GetLandscapesByEnvironment(c *gin.Context) {
	environmentStr := c.Param("environment")
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

	// Convert string to LandscapeType
	var landscapeType models.LandscapeType
	switch environmentStr {
	case "development":
		landscapeType = models.LandscapeTypeDevelopment
	case "staging":
		landscapeType = models.LandscapeTypeStaging
	case "production":
		landscapeType = models.LandscapeTypeProduction
	case "testing":
		landscapeType = models.LandscapeTypeTesting
	case "preview":
		landscapeType = models.LandscapeTypePreview
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape type. Valid values: development, staging, production, testing, preview"})
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

	landscapes, err := h.landscapeService.GetByType(organizationID, landscapeType, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscapes)
}

// GetLandscapeWithProjects handles GET /landscapes/:id/projects
// @Summary Get landscape with its projects
// @Description Get a landscape including all its associated projects
// @Tags landscapes
// @Accept json
// @Produce json
// @Param id path string true "Landscape ID (UUID)"
// @Success 200 {object} map[string]interface{} "Successfully retrieved landscape with projects"
// @Failure 400 {object} map[string]interface{} "Invalid landscape ID"
// @Failure 404 {object} map[string]interface{} "Landscape not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes/{id}/projects [get]
func (h *LandscapeHandler) GetLandscapeWithProjects(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
		return
	}

	landscape, err := h.landscapeService.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrLandscapeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscape)
}

// GetLandscapeWithDeployments handles GET /landscapes/:id/deployments
// @Summary Get landscape with its component deployments
// @Description Get a landscape including all its component deployments
// @Tags landscapes
// @Accept json
// @Produce json
// @Param id path string true "Landscape ID (UUID)"
// @Success 200 {object} map[string]interface{} "Successfully retrieved landscape with deployments"
// @Failure 400 {object} map[string]interface{} "Invalid landscape ID"
// @Failure 404 {object} map[string]interface{} "Landscape not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes/{id}/deployments [get]
func (h *LandscapeHandler) GetLandscapeWithDeployments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
		return
	}

	landscape, err := h.landscapeService.GetWithComponentDeployments(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrLandscapeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscape)
}

// GetLandscapeWithFullDetails handles GET /landscapes/:id/details
// @Summary Get landscape with full details
// @Description Get a landscape with all its related data (projects, deployments, etc.)
// @Tags landscapes
// @Accept json
// @Produce json
// @Param id path string true "Landscape ID (UUID)"
// @Success 200 {object} map[string]interface{} "Successfully retrieved landscape with full details"
// @Failure 400 {object} map[string]interface{} "Invalid landscape ID"
// @Failure 404 {object} map[string]interface{} "Landscape not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /landscapes/{id}/details [get]
func (h *LandscapeHandler) GetLandscapeWithFullDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
		return
	}

	landscape, err := h.landscapeService.GetWithFullDetails(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrLandscapeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, landscape)
}
