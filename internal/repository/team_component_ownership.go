package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamComponentOwnershipRepository handles database operations for team component ownership
type TeamComponentOwnershipRepository struct {
	db *gorm.DB
}

// NewTeamComponentOwnershipRepository creates a new team component ownership repository
func NewTeamComponentOwnershipRepository(db *gorm.DB) *TeamComponentOwnershipRepository {
	return &TeamComponentOwnershipRepository{db: db}
}

// Create creates a new team component ownership
func (r *TeamComponentOwnershipRepository) Create(ownership *models.TeamComponentOwnership) error {
	return r.db.Create(ownership).Error
}

// GetByTeamID retrieves all component ownerships for a team
func (r *TeamComponentOwnershipRepository) GetByTeamID(teamID uuid.UUID) ([]models.TeamComponentOwnership, error) {
	var ownerships []models.TeamComponentOwnership
	err := r.db.Where("team_id = ?", teamID).Find(&ownerships).Error
	return ownerships, err
}

// GetByComponentID retrieves all team ownerships for a component
func (r *TeamComponentOwnershipRepository) GetByComponentID(componentID uuid.UUID) ([]models.TeamComponentOwnership, error) {
	var ownerships []models.TeamComponentOwnership
	err := r.db.Where("component_id = ?", componentID).Find(&ownerships).Error
	return ownerships, err
}

// GetByOwnershipType retrieves ownerships by type
func (r *TeamComponentOwnershipRepository) GetByOwnershipType(ownershipType models.OwnershipType) ([]models.TeamComponentOwnership, error) {
	var ownerships []models.TeamComponentOwnership
	err := r.db.Where("ownership_type = ?", ownershipType).Find(&ownerships).Error
	return ownerships, err
}

// Delete removes a team component ownership
func (r *TeamComponentOwnershipRepository) Delete(teamID, componentID uuid.UUID) error {
	return r.db.Where("team_id = ? AND component_id = ?", teamID, componentID).Delete(&models.TeamComponentOwnership{}).Error
}

// Update updates ownership details
func (r *TeamComponentOwnershipRepository) Update(ownership *models.TeamComponentOwnership) error {
	return r.db.Save(ownership).Error
}

// Exists checks if a team component ownership exists
func (r *TeamComponentOwnershipRepository) Exists(teamID, componentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.TeamComponentOwnership{}).Where("team_id = ? AND component_id = ?", teamID, componentID).Count(&count).Error
	return count > 0, err
}
