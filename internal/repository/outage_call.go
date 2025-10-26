package repository

import (
	"developer-portal-backend/internal/database/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OutageCallRepository handles database operations for outage calls
type OutageCallRepository struct {
	db *gorm.DB
}

// NewOutageCallRepository creates a new outage call repository
func NewOutageCallRepository(db *gorm.DB) *OutageCallRepository {
	return &OutageCallRepository{db: db}
}

// Create creates a new outage call
func (r *OutageCallRepository) Create(call *models.OutageCall) error {
	return r.db.Create(call).Error
}

// GetByID retrieves an outage call by ID
func (r *OutageCallRepository) GetByID(id uuid.UUID) (*models.OutageCall, error) {
	var call models.OutageCall
	err := r.db.First(&call, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &call, nil
}

// GetByTeamID retrieves all outage calls for a team
func (r *OutageCallRepository) GetByTeamID(teamID uuid.UUID, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	if err := r.db.Model(&models.OutageCall{}).Where("team_id = ?", teamID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("team_id = ?", teamID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// GetByStatus retrieves outage calls by status
func (r *OutageCallRepository) GetByStatus(status models.OutageCallStatus, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	query := r.db.Model(&models.OutageCall{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// GetActiveCalls retrieves all active outage calls
func (r *OutageCallRepository) GetActiveCalls(limit, offset int) ([]models.OutageCall, int64, error) {
	return r.GetByStatus(models.OutageCallStatusOpen, limit, offset)
}

// GetRecentCalls retrieves recent outage calls within specified days
func (r *OutageCallRepository) GetRecentCalls(days int, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	cutoffDate := time.Now().AddDate(0, 0, -days)
	query := r.db.Model(&models.OutageCall{}).Where("created_at >= ?", cutoffDate)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// Update updates an outage call
func (r *OutageCallRepository) Update(call *models.OutageCall) error {
	return r.db.Save(call).Error
}

// Delete deletes an outage call
func (r *OutageCallRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.OutageCall{}, "id = ?", id).Error
}

// GetWithAssignees retrieves an outage call with assignees
func (r *OutageCallRepository) GetWithAssignees(id uuid.UUID) (*models.OutageCall, error) {
	var call models.OutageCall
	err := r.db.Preload("Assignees").Preload("Assignees.Member").First(&call, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &call, nil
}

// GetWithTeam retrieves an outage call with team details
func (r *OutageCallRepository) GetWithTeam(id uuid.UUID) (*models.OutageCall, error) {
	var call models.OutageCall
	err := r.db.Preload("Team").First(&call, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &call, nil
}

// SetStatus sets the status of an outage call
func (r *OutageCallRepository) SetStatus(callID uuid.UUID, status models.OutageCallStatus) error {
	return r.db.Model(&models.OutageCall{}).Where("id = ?", callID).Update("status", status).Error
}

// GetByOrganizationID retrieves outage calls for a specific organization
func (r *OutageCallRepository) GetByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	if err := r.db.Model(&models.OutageCall{}).Where("organization_id = ?", organizationID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("organization_id = ?", organizationID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// GetByYear retrieves outage calls for a specific year
func (r *OutageCallRepository) GetByYear(year int, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	if err := r.db.Model(&models.OutageCall{}).Where("year = ?", year).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("year = ?", year).Order("created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// GetBySeverity retrieves outage calls by severity
func (r *OutageCallRepository) GetBySeverity(severity models.OutageCallSeverity, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	query := r.db.Model(&models.OutageCall{}).Where("severity = ?", severity)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// GetByResolutionTime retrieves outage calls by resolution time range
func (r *OutageCallRepository) GetByResolutionTime(minMinutes, maxMinutes int, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	query := r.db.Model(&models.OutageCall{}).Where("resolution_time_minutes BETWEEN ? AND ?", minMinutes, maxMinutes)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// GetByCallTimestamp retrieves outage calls within a timestamp range
func (r *OutageCallRepository) GetByCallTimestamp(startTime, endTime time.Time, limit, offset int) ([]models.OutageCall, int64, error) {
	var calls []models.OutageCall
	var total int64

	query := r.db.Model(&models.OutageCall{}).Where("call_timestamp BETWEEN ? AND ?", startTime, endTime)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("call_timestamp DESC").Limit(limit).Offset(offset).Find(&calls).Error
	return calls, total, err
}

// GetOutageStats retrieves outage statistics
func (r *OutageCallRepository) GetOutageStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	var results []struct {
		Status string
		Count  int64
	}

	err := r.db.Model(&models.OutageCall{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	for _, result := range results {
		stats[result.Status] = result.Count
	}

	// Total count
	var total int64
	if err := r.db.Model(&models.OutageCall{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	return stats, nil
}
