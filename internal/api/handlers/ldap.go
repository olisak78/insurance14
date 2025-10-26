package handlers

import (
	"net/http"

	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// LDAPHandler handles LDAP-related HTTP requests
type LDAPHandler struct {
	service *service.LDAPService
}

// NewLDAPHandler creates a new LDAP handler
func NewLDAPHandler(s *service.LDAPService) *LDAPHandler {
	return &LDAPHandler{service: s}
}

// UserSearch searches LDAP users by CN prefix
// @Summary Search LDAP users by CN prefix
// @Description Searches LDAP directory for users where cn starts with given prefix
// @Tags ldap
// @Produce json
// @Param cn query string true "Common name prefix"
// @Success 200 {object} map[string]interface{} "Search results"
// @Failure 400 {object} map[string]interface{} "Missing or invalid query parameter"
// @Failure 502 {object} map[string]interface{} "LDAP connection or search failed"
// @Security BearerAuth
// @Router /ldap/users/search [get]
func (h *LDAPHandler) UserSearch(c *gin.Context) {
	cn := c.Query("cn")
	if cn == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter: cn"})
		return
	}

	users, err := h.service.SearchUsersByCN(cn)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "ldap search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": users})
}
