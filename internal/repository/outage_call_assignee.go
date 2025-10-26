package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OutageCallAssigneeRepository handles database operations for outage call assignees
type OutageCallAssigneeRepository struct {
	db *gorm.DB
}

// NewOutageCallAssigneeRepository creates a new outage call assignee repository
func NewOutageCallAssigneeRepository(db *gorm.DB) *OutageCallAssigneeRepository {
	return &OutageCallAssigneeRepository{db: db}
}

// Create creates a new outage call assignee
func (r *OutageCallAssigneeRepository) Create(assignee *models.OutageCallAssignee) error {
	return r.db.Create(assignee).Error
}

// GetByOutageCallID retrieves all assignees for an outage call
func (r *OutageCallAssigneeRepository) GetByOutageCallID(outageCallID uuid.UUID) ([]models.OutageCallAssignee, error) {
	var assignees []models.OutageCallAssignee
	err := r.db.Where("outage_call_id = ?", outageCallID).Find(&assignees).Error
	return assignees, err
}

// GetByMemberID retrieves all outage call assignments for a member
func (r *OutageCallAssigneeRepository) GetByMemberID(memberID uuid.UUID) ([]models.OutageCallAssignee, error) {
	var assignees []models.OutageCallAssignee
	err := r.db.Where("member_id = ?", memberID).Find(&assignees).Error
	return assignees, err
}

// Delete removes an outage call assignee
func (r *OutageCallAssigneeRepository) Delete(outageCallID, memberID uuid.UUID) error {
	return r.db.Where("outage_call_id = ? AND member_id = ?", outageCallID, memberID).Delete(&models.OutageCallAssignee{}).Error
}

// Exists checks if an outage call assignee exists
func (r *OutageCallAssigneeRepository) Exists(outageCallID, memberID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.OutageCallAssignee{}).Where("outage_call_id = ? AND member_id = ?", outageCallID, memberID).Count(&count).Error
	return count > 0, err
}
