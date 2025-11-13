package handlers

import (
	"net/http"
	"strconv"

	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// CategoryHandler handles HTTP requests for category operations
type CategoryHandler struct {
	categoryService service.CategoryServiceInterface
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryService service.CategoryServiceInterface) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// ListCategories handles GET /categories
// @Summary List all categories
// @Description Get all categories with pagination support
// @Tags categories
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(1000)
// @Success 200 {object} service.CategoryListResponse "Successfully retrieved categories"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	// Parse pagination parameters (default to large page size to effectively return all)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "1000"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 1000
	}

	resp, err := h.categoryService.GetAll(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
