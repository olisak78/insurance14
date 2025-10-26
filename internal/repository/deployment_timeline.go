package repository

import (
	"developer-portal-backend/internal/database/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeploymentTimelineRepository handles database operations for deployment timeline
type DeploymentTimelineRepository struct {
	db *gorm.DB
}

// NewDeploymentTimelineRepository creates a new deployment timeline repository
func NewDeploymentTimelineRepository(db *gorm.DB) *DeploymentTimelineRepository {
	return &DeploymentTimelineRepository{db: db}
}

// Create creates a new deployment timeline entry
func (r *DeploymentTimelineRepository) Create(timeline *models.DeploymentTimeline) error {
	return r.db.Create(timeline).Error
}

// GetByID retrieves a deployment timeline by ID
func (r *DeploymentTimelineRepository) GetByID(id uuid.UUID) (*models.DeploymentTimeline, error) {
	var timeline models.DeploymentTimeline
	err := r.db.First(&timeline, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &timeline, nil
}

// GetByComponentDeploymentID retrieves timeline entries for a component deployment
func (r *DeploymentTimelineRepository) GetByComponentDeploymentID(deploymentID uuid.UUID, limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	var timelines []models.DeploymentTimeline
	var total int64

	if err := r.db.Model(&models.DeploymentTimeline{}).Where("component_deployment_id = ?", deploymentID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("component_deployment_id = ?", deploymentID).Order("event_time DESC").Limit(limit).Offset(offset).Find(&timelines).Error
	return timelines, total, err
}

// GetByDateRange retrieves timeline entries within a date range
func (r *DeploymentTimelineRepository) GetByDateRange(startDate, endDate time.Time, limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	var timelines []models.DeploymentTimeline
	var total int64

	query := r.db.Model(&models.DeploymentTimeline{}).Where("event_time BETWEEN ? AND ?", startDate, endDate)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("event_time DESC").Limit(limit).Offset(offset).Find(&timelines).Error
	return timelines, total, err
}

// GetRecentDeployments retrieves recent deployment events
func (r *DeploymentTimelineRepository) GetRecentDeployments(days int, limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.GetByDateRange(cutoffDate, time.Now(), limit, offset)
}

// Update updates a deployment timeline entry
func (r *DeploymentTimelineRepository) Update(timeline *models.DeploymentTimeline) error {
	return r.db.Save(timeline).Error
}

// Delete deletes a deployment timeline entry
func (r *DeploymentTimelineRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.DeploymentTimeline{}, "id = ?", id).Error
}

// GetByLandscapeID retrieves timeline entries for a specific landscape
func (r *DeploymentTimelineRepository) GetByLandscapeID(landscapeID uuid.UUID, limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	var timelines []models.DeploymentTimeline
	var total int64

	if err := r.db.Model(&models.DeploymentTimeline{}).Where("landscape_id = ?", landscapeID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("landscape_id = ?", landscapeID).Order("scheduled_date DESC").Limit(limit).Offset(offset).Find(&timelines).Error
	return timelines, total, err
}

// GetByOrganizationID retrieves timeline entries for a specific organization
func (r *DeploymentTimelineRepository) GetByOrganizationID(organizationID uuid.UUID, limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	var timelines []models.DeploymentTimeline
	var total int64

	if err := r.db.Model(&models.DeploymentTimeline{}).Where("organization_id = ?", organizationID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("organization_id = ?", organizationID).Order("scheduled_date DESC").Limit(limit).Offset(offset).Find(&timelines).Error
	return timelines, total, err
}

// GetCompleted retrieves completed timeline entries
func (r *DeploymentTimelineRepository) GetCompleted(limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	var timelines []models.DeploymentTimeline
	var total int64

	if err := r.db.Model(&models.DeploymentTimeline{}).Where("is_completed = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("is_completed = ?", true).Order("scheduled_date DESC").Limit(limit).Offset(offset).Find(&timelines).Error
	return timelines, total, err
}

// GetPending retrieves pending timeline entries
func (r *DeploymentTimelineRepository) GetPending(limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	var timelines []models.DeploymentTimeline
	var total int64

	if err := r.db.Model(&models.DeploymentTimeline{}).Where("is_completed = ?", false).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("is_completed = ?", false).Order("scheduled_date ASC").Limit(limit).Offset(offset).Find(&timelines).Error
	return timelines, total, err
}

// GetDeploymentHistory retrieves deployment history for a component
func (r *DeploymentTimelineRepository) GetDeploymentHistory(componentID uuid.UUID, limit, offset int) ([]models.DeploymentTimeline, int64, error) {
	var timelines []models.DeploymentTimeline
	var total int64

	query := r.db.Table("deployment_timelines dt").
		Joins("JOIN component_deployments cd ON dt.component_deployment_id = cd.id").
		Where("cd.component_id = ?", componentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("dt.event_time DESC").Limit(limit).Offset(offset).Find(&timelines).Error
	return timelines, total, err
}
