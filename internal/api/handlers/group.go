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

// GroupHandler handles HTTP requests for groups
type GroupHandler struct {
	service service.GroupServiceInterface
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(service service.GroupServiceInterface) *GroupHandler {
	return &GroupHandler{service: service}
}

// CreateGroup creates a new group
// @Summary Create a new group
// @Description Create a new group within an organization
// @Tags groups
// @Accept json
// @Produce json
// @Param group body service.CreateGroupRequest true "Group data"
// @Success 201 {object} service.GroupResponse "Successfully created group"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req service.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.service.Create(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// GetGroup retrieves a group by ID
// @Summary Get group by ID
// @Description Get a specific group by its UUID
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Success 200 {object} service.GroupResponse "Successfully retrieved group"
// @Failure 400 {object} map[string]interface{} "Invalid group ID"
// @Failure 404 {object} map[string]interface{} "Group not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /groups/{id} [get]
func (h *GroupHandler) GetGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	group, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// UpdateGroup updates a group
// @Summary Update group
// @Description Update an existing group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Param group body service.UpdateGroupRequest true "Updated group data"
// @Success 200 {object} service.GroupResponse "Successfully updated group"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Group not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /groups/{id} [put]
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var req service.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.service.Update(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a group
// @Summary Delete group
// @Description Delete a group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID (UUID)"
// @Success 204 "Successfully deleted group"
// @Failure 400 {object} map[string]interface{} "Invalid group ID"
// @Failure 404 {object} map[string]interface{} "Group not found"
// @Security BearerAuth
// @Router /groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	err = h.service.Delete(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetGroupsByOrganization retrieves all groups for an organization
// @Summary List groups by organization
// @Description Get all groups belonging to an organization with pagination
// @Tags groups
// @Accept json
// @Produce json
// @Param organization_id query string true "Organization ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(20)
// @Success 200 {object} service.GroupListResponse "Successfully retrieved groups"
// @Failure 400 {object} map[string]interface{} "Invalid organization ID"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /groups [get]
func (h *GroupHandler) GetGroupsByOrganization(c *gin.Context) {
	organizationIDStr := c.Param("id")
	if organizationIDStr == "" {
		organizationIDStr = c.Query("organization_id")
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	groups, err := h.service.GetByOrganization(organizationID, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// GetGroupWithTeams retrieves a group with its teams
func (h *GroupHandler) GetGroupWithTeams(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	// Parse pagination parameters for teams
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	group, err := h.service.GetWithTeams(id, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// GetGroupByName retrieves a group by name within an organization
func (h *GroupHandler) GetGroupByName(c *gin.Context) {
	name := c.Param("name")
	organizationIDStr := c.Query("organization_id")

	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id parameter is required"})
		return
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	group, err := h.service.GetByName(organizationID, name)
	if err != nil {
		if errors.Is(err, apperrors.ErrGroupNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// SearchGroups searches for groups by name or description within an organization
func (h *GroupHandler) SearchGroups(c *gin.Context) {
	organizationIDStr := c.Query("organization_id")
	query := c.Query("q")

	if organizationIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization_id parameter is required"})
		return
	}

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter is required"})
		return
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	groups, err := h.service.Search(organizationID, query, page, pageSize)
	if err != nil {
		if errors.Is(err, apperrors.ErrOrganizationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}
