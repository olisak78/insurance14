package handlers

import (
	"developer-portal-backend/internal/service"
	"errors"
	"net/http"
	"strconv"

	apperrors "developer-portal-backend/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MemberHandler handles HTTP requests for members
type MemberHandler struct {
	memberService *service.MemberService
}

// NewMemberHandler creates a new member handler
func NewMemberHandler(memberService *service.MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

// CreateMember creates a new member
// @Summary Create a new member
// @Description Create a new member in the system with optional default values.
// @Description
// @Description Optional Fields with Defaults:
// @Description - role: Defaults to 'developer' (valid values: admin, developer, manager, viewer)
// @Description - team_role: Defaults to 'member' (valid values: member, team_lead)
// @Description - external_type: Defaults to 'internal'
// @Description - is_active: Defaults to true
// @Tags members
// @Accept json
// @Produce json
// @Param member body service.CreateMemberRequest true "Member data"
// @Success 201 {object} service.MemberResponse "Successfully created member"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Security BearerAuth
// @Router /members [post]
func (h *MemberHandler) CreateMember(c *gin.Context) {
	var req service.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.memberService.CreateMember(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// GetMember retrieves a member by ID
// @Summary Get member by ID
// @Description Get a specific member by their UUID
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Member ID (UUID)"
// @Success 200 {object} service.MemberResponse "Successfully retrieved member"
// @Failure 400 {object} map[string]interface{} "Invalid member ID"
// @Failure 404 {object} map[string]interface{} "Member not found"
// @Security BearerAuth
// @Router /members/{id} [get]
func (h *MemberHandler) GetMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	member, err := h.memberService.GetMemberByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// GetMembersByOrganization retrieves members for an organization
// @Summary List members by organization
// @Description Get all members belonging to an organization with pagination. Can be accessed via /members?organization_id=xxx or /organizations/:id/members
// @Tags members
// @Accept json
// @Produce json
// @Param organization_id query string false "Organization ID (UUID) - used when accessing via /members endpoint"
// @Param id path string false "Organization ID (UUID) - used when accessing via /organizations/:id/members endpoint"
// @Param limit query int false "Number of items to return" default(20)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} map[string]interface{} "Successfully retrieved members list"
// @Failure 400 {object} map[string]interface{} "Invalid organization ID or parameters"
// @Security BearerAuth
// @Router /members [get]
func (h *MemberHandler) GetMembersByOrganization(c *gin.Context) {
	// Try to get organization ID from path parameter first (for /organizations/:id/members)
	// Then fall back to query parameter (for /members?organization_id=...)
	orgIDStr := c.Param("id")
	if orgIDStr == "" {
		orgIDStr = c.Query("organization_id")
	}

	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID is required (provide as query parameter 'organization_id' or path parameter 'id')"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	members, total, err := h.memberService.GetMembersByOrganization(orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// SearchMembers searches for members
func (h *MemberHandler) SearchMembers(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	members, total, err := h.memberService.SearchMembers(orgID, query, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"query":   query,
	})
}

// UpdateMember updates an existing member
// @Summary Update member
// @Description Update an existing member by ID
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Member ID (UUID)"
// @Param member body service.UpdateMemberRequest true "Updated member data"
// @Success 200 {object} service.MemberResponse "Successfully updated member"
// @Failure 400 {object} map[string]interface{} "Invalid request body or member ID"
// @Failure 404 {object} map[string]interface{} "Member not found"
// @Security BearerAuth
// @Router /members/{id} [put]
func (h *MemberHandler) UpdateMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	var req service.UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.memberService.UpdateMember(id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

// DeleteMember deletes a member
// @Summary Delete member
// @Description Delete a member by ID
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Member ID (UUID)"
// @Success 204 "Successfully deleted member"
// @Failure 400 {object} map[string]interface{} "Invalid member ID"
// @Failure 404 {object} map[string]interface{} "Member not found"
// @Security BearerAuth
// @Router /members/{id} [delete]
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
		return
	}

	if err := h.memberService.DeleteMember(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetActiveMembers retrieves active members for an organization
func (h *MemberHandler) GetActiveMembers(c *gin.Context) {
	orgIDStr := c.Param("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	members, total, err := h.memberService.GetActiveMembers(orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// AddQuickLink handles POST /members/:id/quick-links
// @Summary Add a quick link to a member
// @Description Add a new quick link to a member's metadata.quick_links array. Links support url, title, icon, and category fields.
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Member ID (UUID)"
// @Param link body service.AddQuickLinkRequest true "Quick link data"
// @Success 200 {object} service.MemberResponse "Successfully added quick link"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Member not found"
// @Failure 409 {object} map[string]interface{} "Link with URL already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /members/{id}/quick-links [post]
func (h *MemberHandler) AddQuickLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
		return
	}

	var req service.AddQuickLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.memberService.AddQuickLink(id, &req)
	if err != nil {
		if errors.Is(err, apperrors.ErrMemberNotFound) {
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

	c.JSON(http.StatusOK, member)
}

// GetQuickLinks handles GET /members/:id/quick-links
// @Summary Get quick links for a member
// @Description Retrieve all quick links from a member's metadata.quick_links array
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Member ID (UUID)"
// @Success 200 {object} service.QuickLinksResponse "Successfully retrieved quick links"
// @Failure 400 {object} map[string]interface{} "Invalid member ID"
// @Failure 404 {object} map[string]interface{} "Member not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /members/{id}/quick-links [get]
func (h *MemberHandler) GetQuickLinks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
		return
	}

	quickLinks, err := h.memberService.GetQuickLinks(id)
	if err != nil {
		if errors.Is(err, apperrors.ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quickLinks)
}

// RemoveQuickLink handles DELETE /members/:id/quick-links
// @Summary Remove a quick link from a member
// @Description Remove a quick link from a member's metadata.quick_links array by URL
// @Tags members
// @Accept json
// @Produce json
// @Param id path string true "Member ID (UUID)"
// @Param url query string true "Link URL to remove"
// @Success 200 {object} service.MemberResponse "Successfully removed quick link"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Member or link not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /members/{id}/quick-links [delete]
func (h *MemberHandler) RemoveQuickLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
		return
	}

	linkURL := c.Query("url")
	if linkURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}

	member, err := h.memberService.RemoveQuickLink(id, linkURL)
	if err != nil {
		if errors.Is(err, apperrors.ErrMemberNotFound) || errors.Is(err, apperrors.ErrLinkNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}
