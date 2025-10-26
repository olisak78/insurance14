package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectRepository handles database operations for projects
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

// GetByID retrieves a project by ID
func (r *ProjectRepository) GetByID(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByName retrieves a project by name within an organization
func (r *ProjectRepository) GetByName(orgID uuid.UUID, name string) (*models.Project, error) {
	var project models.Project
	err := r.db.First(&project, "organization_id = ? AND name = ?", orgID, name).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByOrganizationID retrieves all projects for an organization with pagination
func (r *ProjectRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	// Get total count
	if err := r.db.Model(&models.Project{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("organization_id = ?", orgID).Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// GetByStatus retrieves all projects with a specific status in an organization
func (r *ProjectRepository) GetByStatus(orgID uuid.UUID, status models.ProjectStatus, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := r.db.Model(&models.Project{}).Where("organization_id = ? AND status = ?", orgID, status)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// GetActiveProjects retrieves all active projects for an organization
func (r *ProjectRepository) GetActiveProjects(orgID uuid.UUID, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := r.db.Model(&models.Project{}).Where("organization_id = ? AND status = ?", orgID, models.ProjectStatusActive)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// Update updates a project
func (r *ProjectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

// Delete deletes a project
func (r *ProjectRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Project{}, "id = ?", id).Error
}

// GetWithOrganization retrieves a project with organization details
func (r *ProjectRepository) GetWithOrganization(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Organization").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetWithComponents retrieves a project with all its components
func (r *ProjectRepository) GetWithComponents(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Components").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetWithLandscapes retrieves a project with all its landscapes
func (r *ProjectRepository) GetWithLandscapes(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Landscapes").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetWithComponentsAndLandscapes retrieves a project with components and landscapes
func (r *ProjectRepository) GetWithComponentsAndLandscapes(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Components").Preload("Landscapes").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetWithDeploymentTimelines retrieves a project with deployment timelines
func (r *ProjectRepository) GetWithDeploymentTimelines(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("DeploymentTimelines").First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetWithFullDetails retrieves a project with all relationships
func (r *ProjectRepository) GetWithFullDetails(id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.db.
		Preload("Organization").
		Preload("Components").
		Preload("Landscapes").
		Preload("DeploymentTimelines").
		First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// SetStatus sets the status of a project
func (r *ProjectRepository) SetStatus(projectID uuid.UUID, status models.ProjectStatus) error {
	return r.db.Model(&models.Project{}).Where("id = ?", projectID).Update("status", status).Error
}

// Search searches for projects by name or description
func (r *ProjectRepository) Search(orgID uuid.UUID, query string, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	searchQuery := r.db.Model(&models.Project{}).Where("organization_id = ? AND (name ILIKE ? OR description ILIKE ?)", orgID, "%"+query+"%", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := searchQuery.Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// GetComponentCount returns the number of components associated with a project
func (r *ProjectRepository) GetComponentCount(projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProjectComponent{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

// GetLandscapeCount returns the number of landscapes associated with a project
func (r *ProjectRepository) GetLandscapeCount(projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ProjectLandscape{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}

// GetProjectsWithCounts retrieves projects with their component and landscape counts
func (r *ProjectRepository) GetProjectsWithCounts(orgID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error) {
	var projects []models.Project
	var total int64
	var results []map[string]interface{}

	// Get total count
	if err := r.db.Model(&models.Project{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get projects with counts
	err := r.db.Raw(`
		SELECT p.*, 
			COUNT(DISTINCT pc.component_id) as component_count,
			COUNT(DISTINCT pl.landscape_id) as landscape_count
		FROM projects p
		LEFT JOIN project_components pc ON p.id = pc.project_id
		LEFT JOIN project_landscapes pl ON p.id = pl.project_id
		WHERE p.organization_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, orgID, limit, offset).Scan(&projects).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to map format for easier JSON handling
	for _, project := range projects {
		projectMap := map[string]interface{}{
			"id":              project.ID,
			"name":            project.Name,
			"display_name":    project.DisplayName,
			"description":     project.Description,
			"status":          project.Status,
			"organization_id": project.OrganizationID,
			"created_at":      project.CreatedAt,
			"updated_at":      project.UpdatedAt,
		}
		results = append(results, projectMap)
	}

	return results, total, nil
}

// CheckProjectExists checks if a project exists by ID
func (r *ProjectRepository) CheckProjectExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Project{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CheckProjectNameExists checks if a project name exists within an organization
func (r *ProjectRepository) CheckProjectNameExists(orgID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&models.Project{}).Where("organization_id = ? AND name = ?", orgID, name)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// GetProjectsByComponentID retrieves all projects that use a specific component
func (r *ProjectRepository) GetProjectsByComponentID(componentID uuid.UUID, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	// Get total count
	subQuery := r.db.Model(&models.ProjectComponent{}).Select("project_id").Where("component_id = ?", componentID)
	if err := r.db.Model(&models.Project{}).Where("id IN (?)", subQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("id IN (?)", subQuery).Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// GetProjectsByLandscapeID retrieves all projects that use a specific landscape
func (r *ProjectRepository) GetProjectsByLandscapeID(landscapeID uuid.UUID, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	// Get total count
	subQuery := r.db.Model(&models.ProjectLandscape{}).Select("project_id").Where("landscape_id = ?", landscapeID)
	if err := r.db.Model(&models.Project{}).Where("id IN (?)", subQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("id IN (?)", subQuery).Limit(limit).Offset(offset).Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// AddComponent adds a component to a project
func (r *ProjectRepository) AddComponent(projectID, componentID uuid.UUID) error {
	projectComponent := &models.ProjectComponent{
		ProjectID:   projectID,
		ComponentID: componentID,
	}
	return r.db.Create(projectComponent).Error
}

// RemoveComponent removes a component from a project
func (r *ProjectRepository) RemoveComponent(projectID, componentID uuid.UUID) error {
	return r.db.Where("project_id = ? AND component_id = ?", projectID, componentID).Delete(&models.ProjectComponent{}).Error
}

// AddLandscape adds a landscape to a project
func (r *ProjectRepository) AddLandscape(projectID, landscapeID uuid.UUID) error {
	projectLandscape := &models.ProjectLandscape{
		ProjectID:   projectID,
		LandscapeID: landscapeID,
	}
	return r.db.Create(projectLandscape).Error
}

// RemoveLandscape removes a landscape from a project
func (r *ProjectRepository) RemoveLandscape(projectID, landscapeID uuid.UUID) error {
	return r.db.Where("project_id = ? AND landscape_id = ?", projectID, landscapeID).Delete(&models.ProjectLandscape{}).Error
}

// CheckComponentInProject checks if a component is associated with a project
func (r *ProjectRepository) CheckComponentInProject(projectID, componentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProjectComponent{}).Where("project_id = ? AND component_id = ?", projectID, componentID).Count(&count).Error
	return count > 0, err
}

// CheckLandscapeInProject checks if a landscape is associated with a project
func (r *ProjectRepository) CheckLandscapeInProject(projectID, landscapeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProjectLandscape{}).Where("project_id = ? AND landscape_id = ?", projectID, landscapeID).Count(&count).Error
	return count > 0, err
}
