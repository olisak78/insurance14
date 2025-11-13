package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentationRepository handles database operations for documentations
type DocumentationRepository struct {
	db *gorm.DB
}

// Ensure DocumentationRepository implements DocumentationRepositoryInterface
var _ DocumentationRepositoryInterface = (*DocumentationRepository)(nil)

// NewDocumentationRepository creates a new documentation repository
func NewDocumentationRepository(db *gorm.DB) *DocumentationRepository {
	return &DocumentationRepository{db: db}
}

// Create inserts a new documentation
func (r *DocumentationRepository) Create(doc *models.Documentation) error {
	return r.db.Create(doc).Error
}

// GetByID retrieves a documentation by its ID
func (r *DocumentationRepository) GetByID(id uuid.UUID) (*models.Documentation, error) {
	var doc models.Documentation
	if err := r.db.First(&doc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetByTeamID retrieves all documentations for a specific team
func (r *DocumentationRepository) GetByTeamID(teamID uuid.UUID) ([]models.Documentation, error) {
	var docs []models.Documentation
	if err := r.db.Where("team_id = ?", teamID).Order("title ASC").Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// Update updates an existing documentation
func (r *DocumentationRepository) Update(doc *models.Documentation) error {
	return r.db.Save(doc).Error
}

// Delete removes a documentation by ID (soft delete)
func (r *DocumentationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Documentation{}, "id = ?", id).Error
}

// GetAll retrieves all documentations (with optional pagination)
func (r *DocumentationRepository) GetAll(limit, offset int) ([]models.Documentation, int64, error) {
	var docs []models.Documentation
	var total int64

	// Count total
	if err := r.db.Model(&models.Documentation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := r.db.Order("title ASC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}
