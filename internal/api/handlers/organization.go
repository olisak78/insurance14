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

// OrganizationHandler handles HTTP requests for organizations
type OrganizationHandler struct {
	service service.OrganizationServiceInterface
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(service service.OrganizationServiceInterface) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

// CreateOrganization handles POST /api/organizations
// @Summary Create a new organization
// @Description Create a new organization with the provided details
// @Tags organizations
// @Accept json
// @Produce json
// @Param organization body service.CreateOrganizationRequest true "Organization data"
// @Success 201 {object} service.OrganizationResponse "Successfully created organization"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Organization already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /organizations [post]
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req service.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	org, err := h.service.Create(&req)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// GetOrganization handles GET /api/organizations/:id
// @Summary Get organization by ID
// @Description Get a specific organization by its UUID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID (UUID)"
// @Success 200 {object} service.OrganizationResponse "Successfully retrieved organization"
// @Failure 400 {object} map[string]interface{} "Invalid organization ID"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID: invalid UUID format"})
		return
	}

	org, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// GetOrganizationByName handles GET /api/v1/organizations/by-name/:name
// @Summary Get organization by name
// @Description Get a specific organization by its name
// @Tags organizations
// @Accept json
// @Produce json
// @Param name path string true "Organization name"
// @Success 200 {object} service.OrganizationResponse "Successfully retrieved organization"
// @Failure 400 {object} map[string]interface{} "Invalid organization name"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /organizations/by-name/{name} [get]
func (h *OrganizationHandler) GetOrganizationByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	org, err := h.service.GetByName(name)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// GetOrganizationByDomain handles GET /api/organizations/domain/:domain
func (h *OrganizationHandler) GetOrganizationByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization domain is required"})
		return
	}

	org, err := h.service.GetByDomain(domain)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// ListOrganizations handles GET /api/organizations
// @Summary List all organizations
// @Description Get all organizations with pagination support
// @Tags organizations
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} service.OrganizationListResponse "Successfully retrieved organizations"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /organizations [get]
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orgs, err := h.service.GetAll(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organizations", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orgs)
}

// UpdateOrganization handles PUT /api/organizations/:id
// @Summary Update organization
// @Description Update an existing organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID (UUID)"
// @Param organization body service.UpdateOrganizationRequest true "Updated organization data"
// @Success 200 {object} service.OrganizationResponse "Successfully updated organization"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req service.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	org, err := h.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

// DeleteOrganization handles DELETE /api/organizations/:id
// @Summary Delete organization
// @Description Delete an organization by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID (UUID)"
// @Success 204 "Successfully deleted organization"
// @Failure 400 {object} map[string]interface{} "Invalid organization ID"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	err = h.service.Delete(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization", "details": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetOrganizationMembers handles GET /api/organizations/:id/members
func (h *OrganizationHandler) GetOrganizationMembers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := h.service.GetWithMembers(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization members", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": org, "members": org.Members})
}

// GetOrganizationGroups handles GET /api/organizations/:id/groups
func (h *OrganizationHandler) GetOrganizationGroups(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := h.service.GetWithGroups(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization groups", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": org, "groups": org.Groups})
}

// GetOrganizationProjects handles GET /api/organizations/:id/projects
func (h *OrganizationHandler) GetOrganizationProjects(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := h.service.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization projects", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": org, "projects": org.Projects})
}

// GetOrganizationComponents handles GET /api/organizations/:id/components
func (h *OrganizationHandler) GetOrganizationComponents(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := h.service.GetWithComponents(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization components", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": org, "components": org.Components})
}

// GetOrganizationLandscapes handles GET /api/organizations/:id/landscapes
func (h *OrganizationHandler) GetOrganizationLandscapes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	org, err := h.service.GetWithLandscapes(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization landscapes", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organization": org, "landscapes": org.Landscapes})
}
