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

// ComponentDeploymentHandler handles HTTP requests for component deployment operations
type ComponentDeploymentHandler struct {
	componentDeploymentService *service.ComponentDeploymentService
}

// NewComponentDeploymentHandler creates a new component deployment handler
func NewComponentDeploymentHandler(componentDeploymentService *service.ComponentDeploymentService) *ComponentDeploymentHandler {
	return &ComponentDeploymentHandler{
		componentDeploymentService: componentDeploymentService,
	}
}

// CreateComponentDeployment handles POST /component-deployments
// @Summary Create a new component deployment
// @Description Deploy a component to a specific landscape
// @Tags component-deployments
// @Accept json
// @Produce json
// @Param deployment body service.CreateComponentDeploymentRequest true "Component deployment data"
// @Success 201 {object} service.ComponentDeploymentResponse "Successfully created component deployment"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 404 {object} map[string]interface{} "Component or landscape not found"
// @Failure 409 {object} map[string]interface{} "Component deployment already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /component-deployments [post]
func (h *ComponentDeploymentHandler) CreateComponentDeployment(c *gin.Context) {
	var req service.CreateComponentDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deployment, err := h.componentDeploymentService.Create(&req)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentNotFound) || errors.Is(err, apperrors.ErrLandscapeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, apperrors.ErrComponentDeploymentExists) || errors.Is(err, apperrors.ErrActiveComponentDeploymentExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, deployment)
}

// GetComponentDeployment handles GET /component-deployments/:id
// @Summary Get component deployment by ID
// @Description Get a specific component deployment by its UUID
// @Tags component-deployments
// @Accept json
// @Produce json
// @Param id path string true "Component deployment ID (UUID)"
// @Success 200 {object} service.ComponentDeploymentResponse "Successfully retrieved component deployment"
// @Failure 400 {object} map[string]interface{} "Invalid component deployment ID"
// @Failure 404 {object} map[string]interface{} "Component deployment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /component-deployments/{id} [get]
func (h *ComponentDeploymentHandler) GetComponentDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component deployment ID"})
		return
	}

	deployment, err := h.componentDeploymentService.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentDeploymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

// UpdateComponentDeployment handles PUT /component-deployments/:id
// @Summary Update component deployment
// @Description Update an existing component deployment by ID
// @Tags component-deployments
// @Accept json
// @Produce json
// @Param id path string true "Component deployment ID (UUID)"
// @Param deployment body service.UpdateComponentDeploymentRequest true "Updated component deployment data"
// @Success 200 {object} service.ComponentDeploymentResponse "Successfully updated component deployment"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Component deployment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /component-deployments/{id} [put]
func (h *ComponentDeploymentHandler) UpdateComponentDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component deployment ID"})
		return
	}

	var req service.UpdateComponentDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deployment, err := h.componentDeploymentService.Update(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentDeploymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

// DeleteComponentDeployment handles DELETE /component-deployments/:id
// @Summary Delete component deployment
// @Description Delete a component deployment by ID
// @Tags component-deployments
// @Accept json
// @Produce json
// @Param id path string true "Component deployment ID (UUID)"
// @Success 204 "Successfully deleted component deployment"
// @Failure 400 {object} map[string]interface{} "Invalid component deployment ID"
// @Failure 404 {object} map[string]interface{} "Component deployment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /component-deployments/{id} [delete]
func (h *ComponentDeploymentHandler) DeleteComponentDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component deployment ID"})
		return
	}

	if err := h.componentDeploymentService.Delete(id); err != nil {
		if errors.Is(err, apperrors.ErrComponentDeploymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ListComponentDeployments handles GET /component-deployments
// @Summary List component deployments
// @Description Get component deployments filtered by component or landscape with pagination
// @Tags component-deployments
// @Accept json
// @Produce json
// @Param component_id query string false "Component ID (UUID) to filter deployments"
// @Param landscape_id query string false "Landscape ID (UUID) to filter deployments"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} service.ComponentDeploymentListResponse "Successfully retrieved component deployments"
// @Failure 400 {object} map[string]interface{} "Invalid parameters - component_id or landscape_id required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /component-deployments [get]
func (h *ComponentDeploymentHandler) ListComponentDeployments(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")
	componentIDStr := c.Query("component_id")
	landscapeIDStr := c.Query("landscape_id")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var deployments *service.ComponentDeploymentListResponse

	if componentIDStr != "" {
		componentID, err := uuid.Parse(componentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
			return
		}
		deployments, err = h.componentDeploymentService.GetByComponent(componentID, page, pageSize)
	} else if landscapeIDStr != "" {
		landscapeID, err := uuid.Parse(landscapeIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
			return
		}
		deployments, err = h.componentDeploymentService.GetByLandscape(landscapeID, page, pageSize)
	} else {
		// For now, require either component_id or landscape_id
		c.JSON(http.StatusBadRequest, gin.H{"error": "either component_id or landscape_id parameter is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// GetComponentDeploymentsByComponent handles GET /components/:componentId/deployments
func (h *ComponentDeploymentHandler) GetComponentDeploymentsByComponent(c *gin.Context) {
	componentIDStr := c.Param("componentId")
	componentID, err := uuid.Parse(componentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component ID"})
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

	deployments, err := h.componentDeploymentService.GetByComponent(componentID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// GetComponentDeploymentsByLandscape handles GET /landscapes/:landscapeId/component-deployments
func (h *ComponentDeploymentHandler) GetComponentDeploymentsByLandscape(c *gin.Context) {
	landscapeIDStr := c.Param("landscapeId")
	landscapeID, err := uuid.Parse(landscapeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid landscape ID"})
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

	deployments, err := h.componentDeploymentService.GetByLandscape(landscapeID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// GetComponentDeploymentWithDetails handles GET /component-deployments/:id/details
// @Summary Get component deployment with full details
// @Description Get a component deployment with all its related data
// @Tags component-deployments
// @Accept json
// @Produce json
// @Param id path string true "Component deployment ID (UUID)"
// @Success 200 {object} map[string]interface{} "Successfully retrieved component deployment with details"
// @Failure 400 {object} map[string]interface{} "Invalid component deployment ID"
// @Failure 404 {object} map[string]interface{} "Component deployment not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /component-deployments/{id}/details [get]
func (h *ComponentDeploymentHandler) GetComponentDeploymentWithDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component deployment ID"})
		return
	}

	deployment, err := h.componentDeploymentService.GetWithFullDetails(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrComponentDeploymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deployment)
}
