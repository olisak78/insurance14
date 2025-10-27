package handlers

import (
	"net/http"

	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// LinkHandler handles link-related HTTP requests
type LinkHandler struct {
	service *service.LinkService
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(service *service.LinkService) *LinkHandler {
	return &LinkHandler{
		service: service,
	}
}

// GetLinksByMemberID handles GET /api/v1/links/:id
func (h *LinkHandler) GetLinksByMemberID(c *gin.Context) {
	memberID := c.Param("id")

	response := h.service.GetLinksByMemberID(memberID)
	c.JSON(http.StatusOK, response)
}