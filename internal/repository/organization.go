package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationRepository handles database operations for organizations
type OrganizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create creates a new organization
func (r *OrganizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}

// GetByID retrieves an organization by ID
func (r *OrganizationRepository) GetByID(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetByName retrieves an organization by name
func (r *OrganizationRepository) GetByName(name string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.First(&org, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetByDomain retrieves an organization by domain
func (r *OrganizationRepository) GetByDomain(domain string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.First(&org, "domain = ?", domain).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetAll retrieves all organizations with pagination
func (r *OrganizationRepository) GetAll(limit, offset int) ([]models.Organization, int64, error) {
	var orgs []models.Organization
	var total int64

	// Get total count
	if err := r.db.Model(&models.Organization{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Limit(limit).Offset(offset).Find(&orgs).Error
	if err != nil {
		return nil, 0, err
	}

	return orgs, total, nil
}

// Update updates an organization
func (r *OrganizationRepository) Update(org *models.Organization) error {
	return r.db.Save(org).Error
}

// Delete deletes an organization
func (r *OrganizationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Organization{}, "id = ?", id).Error
}

// GetWithMembers retrieves an organization with its members
func (r *OrganizationRepository) GetWithMembers(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Preload("Users").First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetWithGroups retrieves an organization with its groups
func (r *OrganizationRepository) GetWithGroups(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Preload("Groups").First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetWithProjects retrieves an organization with its projects
func (r *OrganizationRepository) GetWithProjects(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Preload("Projects").First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetWithComponents retrieves an organization with its components
func (r *OrganizationRepository) GetWithComponents(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Preload("Components").First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetWithLandscapes retrieves an organization with its landscapes
func (r *OrganizationRepository) GetWithLandscapes(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Preload("Landscapes").First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetWithAllRelations retrieves an organization with all its relations
func (r *OrganizationRepository) GetWithAllRelations(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.
		Preload("Users").
		Preload("Groups").
		Preload("Projects").
		Preload("Components").
		Preload("Landscapes").
		First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}
