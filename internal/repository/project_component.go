package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectComponentRepository handles database operations for project-component associations
type ProjectComponentRepository struct {
	db *gorm.DB
}

// NewProjectComponentRepository creates a new project component repository
func NewProjectComponentRepository(db *gorm.DB) *ProjectComponentRepository {
	return &ProjectComponentRepository{db: db}
}

// Create creates a new project-component association
func (r *ProjectComponentRepository) Create(projectComponent *models.ProjectComponent) error {
	return r.db.Create(projectComponent).Error
}

// GetByProjectID retrieves all components for a project
func (r *ProjectComponentRepository) GetByProjectID(projectID uuid.UUID) ([]models.ProjectComponent, error) {
	var projectComponents []models.ProjectComponent
	err := r.db.Where("project_id = ?", projectID).Find(&projectComponents).Error
	return projectComponents, err
}

// GetByComponentID retrieves all projects for a component
func (r *ProjectComponentRepository) GetByComponentID(componentID uuid.UUID) ([]models.ProjectComponent, error) {
	var projectComponents []models.ProjectComponent
	err := r.db.Where("component_id = ?", componentID).Find(&projectComponents).Error
	return projectComponents, err
}

// Delete removes a project-component association
func (r *ProjectComponentRepository) Delete(projectID, componentID uuid.UUID) error {
	return r.db.Where("project_id = ? AND component_id = ?", projectID, componentID).Delete(&models.ProjectComponent{}).Error
}

// Exists checks if a project-component association exists
func (r *ProjectComponentRepository) Exists(projectID, componentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProjectComponent{}).Where("project_id = ? AND component_id = ?", projectID, componentID).Count(&count).Error
	return count > 0, err
}
