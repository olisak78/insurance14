package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ComponentRepository handles database operations for components
type ComponentRepository struct {
	db *gorm.DB
}

// NewComponentRepository creates a new component repository
func NewComponentRepository(db *gorm.DB) *ComponentRepository {
	return &ComponentRepository{db: db}
}

// Create creates a new component
func (r *ComponentRepository) Create(component *models.Component) error {
	return r.db.Create(component).Error
}

// GetByID retrieves a component by ID
func (r *ComponentRepository) GetByID(id uuid.UUID) (*models.Component, error) {
	var component models.Component
	err := r.db.First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetByName retrieves a component by name within an organization
func (r *ComponentRepository) GetByName(orgID uuid.UUID, name string) (*models.Component, error) {
	var component models.Component
	err := r.db.First(&component, "organization_id = ? AND name = ?", orgID, name).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetByOrganizationID retrieves all components for an organization with pagination
func (r *ComponentRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	// Get total count
	if err := r.db.Model(&models.Component{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("organization_id = ?", orgID).Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetByType retrieves all components of a specific type in an organization
func (r *ComponentRepository) GetByType(orgID uuid.UUID, componentType models.ComponentType, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	query := r.db.Model(&models.Component{}).Where("organization_id = ? AND component_type = ?", orgID, componentType)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetByStatus retrieves all components with a specific status in an organization
func (r *ComponentRepository) GetByStatus(orgID uuid.UUID, status models.ComponentStatus, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	query := r.db.Model(&models.Component{}).Where("organization_id = ? AND status = ?", orgID, status)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetActiveComponents retrieves all active components for an organization
func (r *ComponentRepository) GetActiveComponents(orgID uuid.UUID, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	query := r.db.Model(&models.Component{}).Where("organization_id = ? AND status = ?", orgID, models.ComponentStatusActive)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// Update updates a component
func (r *ComponentRepository) Update(component *models.Component) error {
	return r.db.Save(component).Error
}

// Delete deletes a component
func (r *ComponentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Component{}, "id = ?", id).Error
}

// GetWithOrganization retrieves a component with organization details
func (r *ComponentRepository) GetWithOrganization(id uuid.UUID) (*models.Component, error) {
	var component models.Component
	err := r.db.Preload("Organization").First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetWithProjects retrieves a component with all its projects
func (r *ComponentRepository) GetWithProjects(id uuid.UUID) (*models.Component, error) {
	var component models.Component
	err := r.db.Preload("Projects").First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetWithDeployments retrieves a component with all its deployments
func (r *ComponentRepository) GetWithDeployments(id uuid.UUID) (*models.Component, error) {
	var component models.Component
	err := r.db.Preload("Deployments").First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetWithTeamOwnerships retrieves a component with team ownerships
func (r *ComponentRepository) GetWithTeamOwnerships(id uuid.UUID) (*models.Component, error) {
	var component models.Component
	err := r.db.Preload("TeamOwnerships").Preload("TeamOwnerships.Team").First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetWithFullDetails retrieves a component with all relationships
func (r *ComponentRepository) GetWithFullDetails(id uuid.UUID) (*models.Component, error) {
	var component models.Component
	err := r.db.
		Preload("Organization").
		Preload("Projects").
		Preload("Deployments").
		Preload("TeamOwnerships").
		Preload("TeamOwnerships.Team").
		First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// SetStatus sets the status of a component
func (r *ComponentRepository) SetStatus(componentID uuid.UUID, status models.ComponentStatus) error {
	return r.db.Model(&models.Component{}).Where("id = ?", componentID).Update("status", status).Error
}

// Search searches for components by name or description
func (r *ComponentRepository) Search(orgID uuid.UUID, query string, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	searchQuery := r.db.Model(&models.Component{}).Where("organization_id = ? AND (name ILIKE ? OR description ILIKE ?)", orgID, "%"+query+"%", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := searchQuery.Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetProjectCount returns the number of projects using a component
func (r *ComponentRepository) GetProjectCount(componentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProjectComponent{}).Where("component_id = ?", componentID).Count(&count).Error
	return count, err
}

// GetDeploymentCount returns the number of deployments for a component
func (r *ComponentRepository) GetDeploymentCount(componentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ComponentDeployment{}).Where("component_id = ?", componentID).Count(&count).Error
	return count, err
}

// GetOwnershipCount returns the number of teams that own a component
func (r *ComponentRepository) GetOwnershipCount(componentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.TeamComponentOwnership{}).Where("component_id = ?", componentID).Count(&count).Error
	return count, err
}

// GetComponentsWithCounts retrieves components with their project, deployment, and ownership counts
func (r *ComponentRepository) GetComponentsWithCounts(orgID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error) {
	var components []models.Component
	var total int64
	var results []map[string]interface{}

	// Get total count
	if err := r.db.Model(&models.Component{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get components with counts
	err := r.db.Raw(`
		SELECT c.*, 
			COUNT(DISTINCT pc.project_id) as project_count,
			COUNT(DISTINCT cd.id) as deployment_count,
			COUNT(DISTINCT tco.team_id) as ownership_count
		FROM components c
		LEFT JOIN project_components pc ON c.id = pc.component_id
		LEFT JOIN component_deployments cd ON c.id = cd.component_id
		LEFT JOIN team_component_ownerships tco ON c.id = tco.component_id
		WHERE c.organization_id = ?
		GROUP BY c.id
		ORDER BY c.created_at DESC
		LIMIT ? OFFSET ?
	`, orgID, limit, offset).Scan(&components).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to map format for easier JSON handling
	for _, component := range components {
		componentMap := map[string]interface{}{
			"id":              component.ID,
			"name":            component.Name,
			"display_name":    component.DisplayName,
			"description":     component.Description,
			"component_type":  component.ComponentType,
			"status":          component.Status,
			"organization_id": component.OrganizationID,
			"created_at":      component.CreatedAt,
			"updated_at":      component.UpdatedAt,
		}
		results = append(results, componentMap)
	}

	return results, total, nil
}

// CheckComponentExists checks if a component exists by ID
func (r *ComponentRepository) CheckComponentExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Component{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CheckComponentNameExists checks if a component name exists within an organization
func (r *ComponentRepository) CheckComponentNameExists(orgID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&models.Component{}).Where("organization_id = ? AND name = ?", orgID, name)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// GetComponentsByTeamID retrieves all components owned by a specific team
func (r *ComponentRepository) GetComponentsByTeamID(teamID uuid.UUID, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	// Get total count
	subQuery := r.db.Model(&models.TeamComponentOwnership{}).Select("component_id").Where("team_id = ?", teamID)
	if err := r.db.Model(&models.Component{}).Where("id IN (?)", subQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("id IN (?)", subQuery).Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetComponentsByProjectID retrieves all components used by a specific project
func (r *ComponentRepository) GetComponentsByProjectID(projectID uuid.UUID, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	// Get total count
	subQuery := r.db.Model(&models.ProjectComponent{}).Select("component_id").Where("project_id = ?", projectID)
	if err := r.db.Model(&models.Component{}).Where("id IN (?)", subQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("id IN (?)", subQuery).Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetUnownedComponents retrieves components that have no team ownership
func (r *ComponentRepository) GetUnownedComponents(orgID uuid.UUID, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	// Get total count
	subQuery := r.db.Model(&models.TeamComponentOwnership{}).Select("component_id")
	query := r.db.Model(&models.Component{}).Where("organization_id = ? AND id NOT IN (?)", orgID, subQuery)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetComponentsByTypeAndStatus retrieves components by type and status
func (r *ComponentRepository) GetComponentsByTypeAndStatus(orgID uuid.UUID, componentType models.ComponentType, status models.ComponentStatus, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	query := r.db.Model(&models.Component{}).Where("organization_id = ? AND component_type = ? AND status = ?", orgID, componentType, status)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GetComponentsByMetadata searches components by metadata field
func (r *ComponentRepository) GetComponentsByMetadata(orgID uuid.UUID, metadata string, limit, offset int) ([]models.Component, int64, error) {
	var components []models.Component
	var total int64

	// Search in JSONB field
	query := r.db.Model(&models.Component{}).Where("organization_id = ? AND metadata::text ILIKE ?", orgID, "%"+metadata+"%")

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&components).Error
	if err != nil {
		return nil, 0, err
	}

	return components, total, nil
}
