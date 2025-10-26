package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ComponentDeploymentRepository handles database operations for component deployments
type ComponentDeploymentRepository struct {
	db *gorm.DB
}

// NewComponentDeploymentRepository creates a new component deployment repository
func NewComponentDeploymentRepository(db *gorm.DB) *ComponentDeploymentRepository {
	return &ComponentDeploymentRepository{db: db}
}

// Create creates a new component deployment
func (r *ComponentDeploymentRepository) Create(deployment *models.ComponentDeployment) error {
	return r.db.Create(deployment).Error
}

// GetByID retrieves a component deployment by ID
func (r *ComponentDeploymentRepository) GetByID(id uuid.UUID) (*models.ComponentDeployment, error) {
	var deployment models.ComponentDeployment
	err := r.db.First(&deployment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetByComponentID retrieves all deployments for a component with pagination
func (r *ComponentDeploymentRepository) GetByComponentID(componentID uuid.UUID, limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	// Get total count
	if err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ?", componentID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("component_id = ?", componentID).Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetByLandscapeID retrieves all deployments in a landscape with pagination
func (r *ComponentDeploymentRepository) GetByLandscapeID(landscapeID uuid.UUID, limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	// Get total count
	if err := r.db.Model(&models.ComponentDeployment{}).Where("landscape_id = ?", landscapeID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("landscape_id = ?", landscapeID).Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetByActiveStatus retrieves all deployments with a specific active status
func (r *ComponentDeploymentRepository) GetByActiveStatus(isActive bool, limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	query := r.db.Model(&models.ComponentDeployment{}).Where("is_active = ?", isActive)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetActiveDeployments retrieves all active deployments
func (r *ComponentDeploymentRepository) GetActiveDeployments(limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	query := r.db.Model(&models.ComponentDeployment{}).Where("is_active = ?", true)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetInactiveDeployments retrieves all inactive deployments
func (r *ComponentDeploymentRepository) GetInactiveDeployments(limit, offset int) ([]models.ComponentDeployment, int64, error) {
	return r.GetByActiveStatus(false, limit, offset)
}

// Update updates a component deployment
func (r *ComponentDeploymentRepository) Update(deployment *models.ComponentDeployment) error {
	return r.db.Save(deployment).Error
}

// Delete deletes a component deployment
func (r *ComponentDeploymentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.ComponentDeployment{}, "id = ?", id).Error
}

// GetWithComponent retrieves a deployment with component details
func (r *ComponentDeploymentRepository) GetWithComponent(id uuid.UUID) (*models.ComponentDeployment, error) {
	var deployment models.ComponentDeployment
	err := r.db.Preload("Component").First(&deployment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetWithLandscape retrieves a deployment with landscape details
func (r *ComponentDeploymentRepository) GetWithLandscape(id uuid.UUID) (*models.ComponentDeployment, error) {
	var deployment models.ComponentDeployment
	err := r.db.Preload("Landscape").First(&deployment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetWithFullDetails retrieves a deployment with all relationships
func (r *ComponentDeploymentRepository) GetWithFullDetails(id uuid.UUID) (*models.ComponentDeployment, error) {
	var deployment models.ComponentDeployment
	err := r.db.
		Preload("Component").
		Preload("Landscape").
		First(&deployment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// SetActiveStatus sets the active status of a deployment
func (r *ComponentDeploymentRepository) SetActiveStatus(deploymentID uuid.UUID, isActive bool) error {
	return r.db.Model(&models.ComponentDeployment{}).Where("id = ?", deploymentID).Update("is_active", isActive).Error
}

// GetByComponentAndLandscape retrieves a deployment by component and landscape
func (r *ComponentDeploymentRepository) GetByComponentAndLandscape(componentID, landscapeID uuid.UUID) (*models.ComponentDeployment, error) {
	var deployment models.ComponentDeployment
	err := r.db.First(&deployment, "component_id = ? AND landscape_id = ?", componentID, landscapeID).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetByVersion retrieves deployments by version
func (r *ComponentDeploymentRepository) GetByVersion(componentID uuid.UUID, version string, limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	query := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND version = ?", componentID, version)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetLatestVersion retrieves the latest deployment for a component in a landscape
func (r *ComponentDeploymentRepository) GetLatestVersion(componentID, landscapeID uuid.UUID) (*models.ComponentDeployment, error) {
	var deployment models.ComponentDeployment
	err := r.db.Where("component_id = ? AND landscape_id = ?", componentID, landscapeID).
		Order("deployed_at DESC").
		First(&deployment).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

// GetDeploymentHistory retrieves deployment history for a component in a landscape
func (r *ComponentDeploymentRepository) GetDeploymentHistory(componentID, landscapeID uuid.UUID, limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	query := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND landscape_id = ?", componentID, landscapeID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results ordered by deployment date
	err := query.Order("deployed_at DESC").Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// CheckDeploymentExists checks if a deployment exists by ID
func (r *ComponentDeploymentRepository) CheckDeploymentExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ComponentDeployment{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CheckDeploymentExistsByComponentAndLandscape checks if a deployment exists for component and landscape
func (r *ComponentDeploymentRepository) CheckDeploymentExistsByComponentAndLandscape(componentID, landscapeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND landscape_id = ?", componentID, landscapeID).Count(&count).Error
	return count > 0, err
}

// GetUndeployedComponents retrieves deployments that haven't been deployed yet
func (r *ComponentDeploymentRepository) GetUndeployedComponents(limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	query := r.db.Model(&models.ComponentDeployment{}).Where("deployed_at IS NULL")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("created_at ASC").Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetRecentlyDeployed retrieves recently deployed components
func (r *ComponentDeploymentRepository) GetRecentlyDeployed(days int, limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	query := r.db.Model(&models.ComponentDeployment{}).Where("deployed_at IS NOT NULL AND deployed_at >= NOW() - INTERVAL ? DAY", days)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("deployed_at DESC").Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetDeploymentsByDateRange retrieves deployments within a date range
func (r *ComponentDeploymentRepository) GetDeploymentsByDateRange(componentID uuid.UUID, startDate, endDate string, limit, offset int) ([]models.ComponentDeployment, int64, error) {
	var deployments []models.ComponentDeployment
	var total int64

	query := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND deployed_at BETWEEN ? AND ?", componentID, startDate, endDate)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("deployed_at DESC").Limit(limit).Offset(offset).Find(&deployments).Error
	if err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// GetDeploymentStats retrieves deployment statistics
func (r *ComponentDeploymentRepository) GetDeploymentStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count active/inactive deployments
	var activeCount, inactiveCount, deployedCount, undeployedCount int64

	// Count active deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("is_active = ?", true).Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active"] = activeCount

	// Count inactive deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("is_active = ?", false).Count(&inactiveCount).Error; err != nil {
		return nil, err
	}
	stats["inactive"] = inactiveCount

	// Count deployed deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("deployed_at IS NOT NULL").Count(&deployedCount).Error; err != nil {
		return nil, err
	}
	stats["deployed"] = deployedCount

	// Count undeployed deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("deployed_at IS NULL").Count(&undeployedCount).Error; err != nil {
		return nil, err
	}
	stats["undeployed"] = undeployedCount

	// Total count
	var total int64
	if err := r.db.Model(&models.ComponentDeployment{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}

// GetComponentDeploymentStats retrieves deployment statistics for a specific component
func (r *ComponentDeploymentRepository) GetComponentDeploymentStats(componentID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count active/inactive deployments for the component
	var activeCount, inactiveCount, deployedCount, undeployedCount int64

	// Count active deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND is_active = ?", componentID, true).Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active"] = activeCount

	// Count inactive deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND is_active = ?", componentID, false).Count(&inactiveCount).Error; err != nil {
		return nil, err
	}
	stats["inactive"] = inactiveCount

	// Count deployed deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND deployed_at IS NOT NULL", componentID).Count(&deployedCount).Error; err != nil {
		return nil, err
	}
	stats["deployed"] = deployedCount

	// Count undeployed deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ? AND deployed_at IS NULL", componentID).Count(&undeployedCount).Error; err != nil {
		return nil, err
	}
	stats["undeployed"] = undeployedCount

	// Total count for the component
	var total int64
	if err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ?", componentID).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}

// GetLandscapeDeploymentStats retrieves deployment statistics for a specific landscape
func (r *ComponentDeploymentRepository) GetLandscapeDeploymentStats(landscapeID uuid.UUID) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count active/inactive deployments for the landscape
	var activeCount, inactiveCount, deployedCount, undeployedCount int64

	// Count active deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("landscape_id = ? AND is_active = ?", landscapeID, true).Count(&activeCount).Error; err != nil {
		return nil, err
	}
	stats["active"] = activeCount

	// Count inactive deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("landscape_id = ? AND is_active = ?", landscapeID, false).Count(&inactiveCount).Error; err != nil {
		return nil, err
	}
	stats["inactive"] = inactiveCount

	// Count deployed deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("landscape_id = ? AND deployed_at IS NOT NULL", landscapeID).Count(&deployedCount).Error; err != nil {
		return nil, err
	}
	stats["deployed"] = deployedCount

	// Count undeployed deployments
	if err := r.db.Model(&models.ComponentDeployment{}).Where("landscape_id = ? AND deployed_at IS NULL", landscapeID).Count(&undeployedCount).Error; err != nil {
		return nil, err
	}
	stats["undeployed"] = undeployedCount

	// Total count for the landscape
	var total int64
	if err := r.db.Model(&models.ComponentDeployment{}).Where("landscape_id = ?", landscapeID).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}
