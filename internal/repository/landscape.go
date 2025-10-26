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

// GetByName retrieves a landscape by name within an organization
func (r *LandscapeRepository) GetByName(orgID uuid.UUID, name string) (*models.Landscape, error) {
	var landscape models.Landscape
	err := r.db.First(&landscape, "organization_id = ? AND name = ?", orgID, name).Error
	if err != nil {
		return nil, err
	}
	return &landscape, nil
}

// GetByOrganizationID retrieves all landscapes for an organization with pagination
func (r *LandscapeRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	// Get total count
	if err := r.db.Model(&models.Landscape{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("organization_id = ?", orgID).Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetByType retrieves all landscapes of a specific type in an organization
func (r *LandscapeRepository) GetByType(orgID uuid.UUID, landscapeType models.LandscapeType, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{}).Where("organization_id = ? AND landscape_type = ?", orgID, landscapeType)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetByStatus retrieves all landscapes with a specific status in an organization
func (r *LandscapeRepository) GetByStatus(orgID uuid.UUID, status models.LandscapeStatus, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{}).Where("organization_id = ? AND status = ?", orgID, status)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetActiveLandscapes retrieves all active landscapes for an organization
func (r *LandscapeRepository) GetActiveLandscapes(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{}).Where("organization_id = ? AND status = ?", orgID, models.LandscapeStatusActive)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
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

// GetWithOrganization retrieves a landscape with organization details
func (r *LandscapeRepository) GetWithOrganization(id uuid.UUID) (*models.Landscape, error) {
	var landscape models.Landscape
	err := r.db.Preload("Organization").First(&landscape, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &landscape, nil
}

// GetWithProjects retrieves a landscape with all its projects
func (r *LandscapeRepository) GetWithProjects(id uuid.UUID) (*models.Landscape, error) {
	var landscape models.Landscape
	err := r.db.Preload("Projects").First(&landscape, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &landscape, nil
}

// GetWithComponentDeployments retrieves a landscape with all component deployments
func (r *LandscapeRepository) GetWithComponentDeployments(id uuid.UUID) (*models.Landscape, error) {
	var landscape models.Landscape
	err := r.db.Preload("ComponentDeployments").Preload("ComponentDeployments.Component").First(&landscape, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &landscape, nil
}

// GetWithFullDetails retrieves a landscape with all relationships
func (r *LandscapeRepository) GetWithFullDetails(id uuid.UUID) (*models.Landscape, error) {
	var landscape models.Landscape
	err := r.db.
		Preload("Organization").
		Preload("Projects").
		Preload("ComponentDeployments").
		Preload("ComponentDeployments.Component").
		First(&landscape, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &landscape, nil
}

// SetStatus sets the status of a landscape
func (r *LandscapeRepository) SetStatus(landscapeID uuid.UUID, status models.LandscapeStatus) error {
	return r.db.Model(&models.Landscape{}).Where("id = ?", landscapeID).Update("status", status).Error
}

// Search searches for landscapes by name or description
func (r *LandscapeRepository) Search(orgID uuid.UUID, query string, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	searchQuery := r.db.Model(&models.Landscape{}).Where("organization_id = ? AND (name ILIKE ? OR description ILIKE ?)", orgID, "%"+query+"%", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := searchQuery.Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetProjectCount returns the number of projects using a landscape
func (r *LandscapeRepository) GetProjectCount(landscapeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProjectLandscape{}).Where("landscape_id = ?", landscapeID).Count(&count).Error
	return count, err
}

// GetDeploymentCount returns the number of component deployments in a landscape
func (r *LandscapeRepository) GetDeploymentCount(landscapeID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ComponentDeployment{}).Where("landscape_id = ?", landscapeID).Count(&count).Error
	return count, err
}

// GetLandscapesWithCounts retrieves landscapes with their project and deployment counts
func (r *LandscapeRepository) GetLandscapesWithCounts(orgID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error) {
	var landscapes []models.Landscape
	var total int64
	var results []map[string]interface{}

	// Get total count
	if err := r.db.Model(&models.Landscape{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get landscapes with counts
	err := r.db.Raw(`
		SELECT l.*, 
			COUNT(DISTINCT pl.project_id) as project_count,
			COUNT(DISTINCT cd.id) as deployment_count
		FROM landscapes l
		LEFT JOIN project_landscapes pl ON l.id = pl.landscape_id
		LEFT JOIN component_deployments cd ON l.id = cd.landscape_id
		WHERE l.organization_id = ?
		GROUP BY l.id
		ORDER BY l.created_at DESC
		LIMIT ? OFFSET ?
	`, orgID, limit, offset).Scan(&landscapes).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to map format for easier JSON handling
	for _, landscape := range landscapes {
		landscapeMap := map[string]interface{}{
			"id":              landscape.ID,
			"name":            landscape.Name,
			"display_name":    landscape.DisplayName,
			"description":     landscape.Description,
			"landscape_type":  landscape.LandscapeType,
			"status":          landscape.Status,
			"organization_id": landscape.OrganizationID,
			"created_at":      landscape.CreatedAt,
			"updated_at":      landscape.UpdatedAt,
		}
		results = append(results, landscapeMap)
	}

	return results, total, nil
}

// CheckLandscapeExists checks if a landscape exists by ID
func (r *LandscapeRepository) CheckLandscapeExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Landscape{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CheckLandscapeNameExists checks if a landscape name exists within an organization
func (r *LandscapeRepository) CheckLandscapeNameExists(orgID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&models.Landscape{}).Where("organization_id = ? AND name = ?", orgID, name)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// GetLandscapesByProjectID retrieves all landscapes used by a specific project
func (r *LandscapeRepository) GetLandscapesByProjectID(projectID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	// Get total count
	subQuery := r.db.Model(&models.ProjectLandscape{}).Select("landscape_id").Where("project_id = ?", projectID)
	if err := r.db.Model(&models.Landscape{}).Where("id IN (?)", subQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("id IN (?)", subQuery).Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetLandscapesByTypeAndStatus retrieves landscapes by type and status
func (r *LandscapeRepository) GetLandscapesByTypeAndStatus(orgID uuid.UUID, landscapeType models.LandscapeType, status models.LandscapeStatus, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	query := r.db.Model(&models.Landscape{}).Where("organization_id = ? AND landscape_type = ? AND status = ?", orgID, landscapeType, status)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetLandscapesByEnvironment retrieves landscapes by environment (from metadata)
func (r *LandscapeRepository) GetLandscapesByEnvironment(orgID uuid.UUID, environment string, limit, offset int) ([]models.Landscape, int64, error) {
	var landscapes []models.Landscape
	var total int64

	// Search in JSONB field for environment
	query := r.db.Model(&models.Landscape{}).Where("organization_id = ? AND metadata->'environment' = ?", orgID, `"`+environment+`"`)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&landscapes).Error
	if err != nil {
		return nil, 0, err
	}

	return landscapes, total, nil
}

// GetProductionLandscapes retrieves all production landscapes
func (r *LandscapeRepository) GetProductionLandscapes(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	return r.GetLandscapesByEnvironment(orgID, "production", limit, offset)
}

// GetDevelopmentLandscapes retrieves all development landscapes
func (r *LandscapeRepository) GetDevelopmentLandscapes(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	return r.GetLandscapesByEnvironment(orgID, "development", limit, offset)
}

// GetStagingLandscapes retrieves all staging landscapes
func (r *LandscapeRepository) GetStagingLandscapes(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error) {
	return r.GetLandscapesByEnvironment(orgID, "staging", limit, offset)
}
