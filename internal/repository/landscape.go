package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LandscapeRepository handles database operations for landscapes
type LandscapeRepository struct {
	db *gorm.DB
}

// NewLandscapeRepository creates a new landscape repository
func NewLandscapeRepository(db *gorm.DB) *LandscapeRepository {
	return &LandscapeRepository{db: db}
}

// Create creates a new landscape
func (r *LandscapeRepository) Create(landscape *models.Landscape) error {
	return r.db.Create(landscape).Error
}

// GetByID retrieves a landscape by ID
func (r *LandscapeRepository) GetByID(id uuid.UUID) (*models.Landscape, error) {
	var landscape models.Landscape
	err := r.db.First(&landscape, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &landscape, nil
}

// GetByName retrieves a landscape by name
func (r *LandscapeRepository) GetByName(name string) (*models.Landscape, error) {
	// New model has no organization scope; filter by name only
	var landscape models.Landscape
	err := r.db.First(&landscape, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &landscape, nil
}

// GetByOrganizationID retrieves landscapes with pagination
func (r *LandscapeRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	// New model has no organization scope; return all landscapes paginated
	var landscapes []models.Landscape
	var total int64

	if err := r.db.Model(&models.Landscape{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Model(&models.Landscape{}).Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetByType retrieves all landscapes of a specific environment
func (r *LandscapeRepository) GetByType(orgID uuid.UUID, landscapeType string, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{}).Where("environment = ?", landscapeType)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetByStatus returns all landscapes (status not present in new model)
func (r *LandscapeRepository) GetByStatus(status string, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetActiveLandscapes returns all landscapes (no status column)
func (r *LandscapeRepository) GetActiveLandscapes(limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	if err := r.db.Model(&models.Landscape{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Model(&models.Landscape{}).Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// Update updates a landscape
func (r *LandscapeRepository) Update(landscape *models.Landscape) error {
	return r.db.Save(landscape).Error
}

// Delete deletes a landscape
func (r *LandscapeRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Landscape{}, "id = ?", id).Error
}

/* GetWithOrganization returns the landscape (no organization relation in new model) */
func (r *LandscapeRepository) GetWithOrganization(id uuid.UUID) (*models.Landscape, error) {
	return r.GetByID(id)
}

// GetWithProjects returns the landscape (project relation is via ProjectID field)
func (r *LandscapeRepository) GetWithProjects(id uuid.UUID) (*models.Landscape, error) {
	return r.GetByID(id)
}

// GetWithComponentDeployments returns the landscape (no direct preload on new model struct)
func (r *LandscapeRepository) GetWithComponentDeployments(id uuid.UUID) (*models.Landscape, error) {
	return r.GetByID(id)
}

// GetWithFullDetails returns the landscape (no additional relations on new model struct)
func (r *LandscapeRepository) GetWithFullDetails(id uuid.UUID) (*models.Landscape, error) {
	return r.GetByID(id)
}

// SetStatus is a no-op (status not present in new model)
func (r *LandscapeRepository) SetStatus(landscapeID uuid.UUID, status string) error {
	return nil
}

// Search searches for landscapes by name, title, or description
func (r *LandscapeRepository) Search(orgID uuid.UUID, q string, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	searchQuery := r.db.Model(&models.Landscape{}).
		Where("(name ILIKE ? OR title ILIKE ? OR description ILIKE ?)", "%"+q+"%", "%"+q+"%", "%"+q+"%")

	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := searchQuery.Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetProjectCount returns the number of projects using a landscape (1 if ProjectID set)
func (r *LandscapeRepository) GetProjectCount(landscapeID uuid.UUID) (int64, error) {
	var l models.Landscape
	if err := r.db.Select("project_id").First(&l, "id = ?", landscapeID).Error; err != nil {
		return 0, err
	}
	if l.ProjectID == uuid.Nil {
		return 0, nil
	}
	return 1, nil
}

// CheckLandscapeExists checks if a landscape exists by ID
func (r *LandscapeRepository) CheckLandscapeExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Landscape{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CheckLandscapeNameExists checks if a landscape name exists (ignores organization scope)
func (r *LandscapeRepository) CheckLandscapeNameExists(orgID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&models.Landscape{}).Where("name = ?", name)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// GetLandscapesByProjectID retrieves all landscapes for a specific project
func (r *LandscapeRepository) GetLandscapesByProjectID(projectID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	if err := r.db.Model(&models.Landscape{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Where("project_id = ?", projectID).Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetLandscapesByTypeAndStatus retrieves landscapes by environment (status ignored)
func (r *LandscapeRepository) GetLandscapesByTypeAndStatus(orgID uuid.UUID, landscapeType string, status string, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{}).Where("environment = ?", landscapeType)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetLandscapesByEnvironment retrieves landscapes by environment
func (r *LandscapeRepository) GetLandscapesByEnvironment(orgID uuid.UUID, environment string, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{}).Where("environment = ?", environment)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&landscapes).Error; err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// Convenience helpers by environment

func (r *LandscapeRepository) GetProductionLandscapes(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	return r.GetLandscapesByEnvironment(orgID, "production", limit, offset)
}

func (r *LandscapeRepository) GetDevelopmentLandscapes(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	return r.GetLandscapesByEnvironment(orgID, "development", limit, offset)
}

func (r *LandscapeRepository) GetStagingLandscapes(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	return r.GetLandscapesByEnvironment(orgID, "staging", limit, offset)
}
