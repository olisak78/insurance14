package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectLandscapeRepository handles database operations for project-landscape associations
type ProjectLandscapeRepository struct {
	db *gorm.DB
}

// NewProjectLandscapeRepository creates a new project landscape repository
func NewProjectLandscapeRepository(db *gorm.DB) *ProjectLandscapeRepository {
	return &ProjectLandscapeRepository{db: db}
}

// Create creates a new project-landscape association
func (r *ProjectLandscapeRepository) Create(projectLandscape *models.ProjectLandscape) error {
	return r.db.Create(projectLandscape).Error
}

// GetByProjectID retrieves all landscapes for a project
func (r *ProjectLandscapeRepository) GetByProjectID(projectID uuid.UUID) ([]models.ProjectLandscape, error) {
	var projectLandscapes []models.ProjectLandscape
	err := r.db.Where("project_id = ?", projectID).Find(&projectLandscapes).Error
	return projectLandscapes, err
}

// GetByLandscapeID retrieves all projects for a landscape
func (r *ProjectLandscapeRepository) GetByLandscapeID(landscapeID uuid.UUID) ([]models.ProjectLandscape, error) {
	var projectLandscapes []models.ProjectLandscape
	err := r.db.Where("landscape_id = ?", landscapeID).Find(&projectLandscapes).Error
	return projectLandscapes, err
}

// Delete removes a project-landscape association
func (r *ProjectLandscapeRepository) Delete(projectID, landscapeID uuid.UUID) error {
	return r.db.Where("project_id = ? AND landscape_id = ?", projectID, landscapeID).Delete(&models.ProjectLandscape{}).Error
}

// Exists checks if a project-landscape association exists
func (r *ProjectLandscapeRepository) Exists(projectID, landscapeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProjectLandscape{}).Where("project_id = ? AND landscape_id = ?", projectID, landscapeID).Count(&count).Error
	return count > 0, err
}
