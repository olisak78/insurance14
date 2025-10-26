package repository

import (
	"developer-portal-backend/internal/database/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DutyScheduleRepository handles database operations for duty schedules
type DutyScheduleRepository struct {
	db *gorm.DB
}

// NewDutyScheduleRepository creates a new duty schedule repository
func NewDutyScheduleRepository(db *gorm.DB) *DutyScheduleRepository {
	return &DutyScheduleRepository{db: db}
}

// Create creates a new duty schedule
func (r *DutyScheduleRepository) Create(schedule *models.DutySchedule) error {
	return r.db.Create(schedule).Error
}

// GetByID retrieves a duty schedule by ID
func (r *DutyScheduleRepository) GetByID(id uuid.UUID) (*models.DutySchedule, error) {
	var schedule models.DutySchedule
	err := r.db.First(&schedule, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// GetByTeamID retrieves all duty schedules for a team
func (r *DutyScheduleRepository) GetByTeamID(teamID uuid.UUID, limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	if err := r.db.Model(&models.DutySchedule{}).Where("team_id = ?", teamID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("team_id = ?", teamID).Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// GetByMemberID retrieves all duty schedules for a member
func (r *DutyScheduleRepository) GetByMemberID(memberID uuid.UUID, limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	if err := r.db.Model(&models.DutySchedule{}).Where("member_id = ?", memberID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("member_id = ?", memberID).Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// GetCurrentDuty retrieves current duty schedules
func (r *DutyScheduleRepository) GetCurrentDuty(teamID uuid.UUID) ([]models.DutySchedule, error) {
	var schedules []models.DutySchedule
	now := time.Now()
	err := r.db.Where("team_id = ? AND start_date <= ? AND end_date >= ?", teamID, now, now).Find(&schedules).Error
	return schedules, err
}

// GetUpcomingDuty retrieves upcoming duty schedules
func (r *DutyScheduleRepository) GetUpcomingDuty(teamID uuid.UUID, days int, limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	futureDate := time.Now().AddDate(0, 0, days)
	query := r.db.Model(&models.DutySchedule{}).Where("team_id = ? AND start_date <= ? AND start_date >= ?", teamID, futureDate, time.Now())

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("start_date ASC").Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// Update updates a duty schedule
func (r *DutyScheduleRepository) Update(schedule *models.DutySchedule) error {
	return r.db.Save(schedule).Error
}

// Delete deletes a duty schedule
func (r *DutyScheduleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.DutySchedule{}, "id = ?", id).Error
}

// GetByOrganizationID retrieves duty schedules for a specific organization
func (r *DutyScheduleRepository) GetByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	if err := r.db.Model(&models.DutySchedule{}).Where("organization_id = ?", organizationID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("organization_id = ?", organizationID).Order("start_date DESC").Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// GetByYear retrieves duty schedules for a specific year
func (r *DutyScheduleRepository) GetByYear(year int, limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	if err := r.db.Model(&models.DutySchedule{}).Where("year = ?", year).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("year = ?", year).Order("start_date DESC").Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// GetByScheduleType retrieves duty schedules by schedule type
func (r *DutyScheduleRepository) GetByScheduleType(scheduleType models.ScheduleType, limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	if err := r.db.Model(&models.DutySchedule{}).Where("schedule_type = ?", scheduleType).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("schedule_type = ?", scheduleType).Order("start_date DESC").Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// GetByShiftType retrieves duty schedules by shift type
func (r *DutyScheduleRepository) GetByShiftType(shiftType models.ShiftType, limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	if err := r.db.Model(&models.DutySchedule{}).Where("shift_type = ?", shiftType).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("shift_type = ?", shiftType).Order("start_date DESC").Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// GetCalled retrieves duty schedules where was_called is true
func (r *DutyScheduleRepository) GetCalled(limit, offset int) ([]models.DutySchedule, int64, error) {
	var schedules []models.DutySchedule
	var total int64

	if err := r.db.Model(&models.DutySchedule{}).Where("was_called = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("was_called = ?", true).Order("start_date DESC").Limit(limit).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// CheckConflict checks for scheduling conflicts
func (r *DutyScheduleRepository) CheckConflict(memberID uuid.UUID, startDate, endDate time.Time, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&models.DutySchedule{}).Where(
		"member_id = ? AND ((start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?))",
		memberID, startDate, startDate, endDate, endDate,
	)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
