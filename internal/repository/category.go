package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRepository handles database operations for categories
type CategoryRepository struct {
	db *gorm.DB
}

// Ensure CategoryRepository implements CategoryRepositoryInterface
var _ CategoryRepositoryInterface = (*CategoryRepository)(nil)

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// GetAll retrieves all categories with pagination
func (r *CategoryRepository) GetAll(limit, offset int) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64

	// Count total
	if err := r.db.Model(&models.Category{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch page
	if err := r.db.Limit(limit).Offset(offset).Order("title ASC").Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

// GetByID retrieves a category by its UUID
func (r *CategoryRepository) GetByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	if err := r.db.First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
