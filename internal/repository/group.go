package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GroupRepository handles database operations for groups
type GroupRepository struct {
	db *gorm.DB
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

// Create creates a new group
func (r *GroupRepository) Create(group *models.Group) error {
	return r.db.Create(group).Error
}

// GetByID retrieves a group by ID
func (r *GroupRepository) GetByID(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	err := r.db.First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetByName retrieves a group by name within an organization
func (r *GroupRepository) GetByName(orgID uuid.UUID, name string) (*models.Group, error) {
	var group models.Group
	err := r.db.First(&group, "org_id = ? AND name = ?", orgID, name).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetByOrganizationID retrieves all groups for an organization with pagination
func (r *GroupRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64

	// Get total count
	if err := r.db.Model(&models.Group{}).Where("org_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("org_id = ?", orgID).Limit(limit).Offset(offset).Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// Update updates a group using a map of updates
func (r *GroupRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&models.Group{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a group
func (r *GroupRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Group{}, "id = ?", id).Error
}

// GetWithTeams retrieves a group with its teams
func (r *GroupRepository) GetWithTeams(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	err := r.db.Preload("Teams").First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetWithOrganization retrieves a group with its organization
func (r *GroupRepository) GetWithOrganization(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	err := r.db.Preload("Organization").First(&group, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// Search searches for groups by name or description within an organization
func (r *GroupRepository) Search(organizationID uuid.UUID, query string, limit, offset int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64

	// Build search query
	searchQuery := "%" + query + "%"
	whereClause := "org_id = ? AND (name ILIKE ? OR title ILIKE ? OR description ILIKE ?)"

	// Get total count
	if err := r.db.Model(&models.Group{}).Where(whereClause, organizationID, searchQuery, searchQuery, searchQuery).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where(whereClause, organizationID, searchQuery, searchQuery, searchQuery).
		Limit(limit).Offset(offset).Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}
