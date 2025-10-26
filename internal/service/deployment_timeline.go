package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"developer-portal-backend/internal/database/models"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeploymentTimelineService handles business logic for deployment timeline
type DeploymentTimelineService struct {
	repo          *repository.DeploymentTimelineRepository
	landscapeRepo *repository.LandscapeRepository
	orgRepo       *repository.OrganizationRepository
	validator     *validator.Validate
}

// NewDeploymentTimelineService creates a new deployment timeline service
func NewDeploymentTimelineService(
	repo *repository.DeploymentTimelineRepository,
	landscapeRepo *repository.LandscapeRepository,
	orgRepo *repository.OrganizationRepository,
	validator *validator.Validate,
) *DeploymentTimelineService {
	return &DeploymentTimelineService{
		repo:          repo,
		landscapeRepo: landscapeRepo,
		orgRepo:       orgRepo,
		validator:     validator,
	}
}

// CreateDeploymentTimelineRequest represents the request to create a deployment timeline entry
type CreateDeploymentTimelineRequest struct {
	OrganizationID  uuid.UUID              `json:"organization_id" validate:"required"`
	LandscapeID     uuid.UUID              `json:"landscape_id" validate:"required"`
	TimelineCode    string                 `json:"timeline_code" validate:"required,max=100"`
	TimelineName    string                 `json:"timeline_name" validate:"required,max=200"`
	ScheduledDate   time.Time              `json:"scheduled_date" validate:"required"`
	IsCompleted     bool                   `json:"is_completed"`
	StatusIndicator string                 `json:"status_indicator,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateDeploymentTimelineRequest represents the request to update a deployment timeline entry
type UpdateDeploymentTimelineRequest struct {
	TimelineCode    *string                `json:"timeline_code,omitempty" validate:"omitempty,max=100"`
	TimelineName    *string                `json:"timeline_name,omitempty" validate:"omitempty,max=200"`
	ScheduledDate   *time.Time             `json:"scheduled_date,omitempty"`
	IsCompleted     *bool                  `json:"is_completed,omitempty"`
	StatusIndicator *string                `json:"status_indicator,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DeploymentTimelineResponse represents the response for deployment timeline operations
type DeploymentTimelineResponse struct {
	ID              uuid.UUID              `json:"id"`
	OrganizationID  uuid.UUID              `json:"organization_id"`
	LandscapeID     uuid.UUID              `json:"landscape_id"`
	TimelineCode    string                 `json:"timeline_code"`
	TimelineName    string                 `json:"timeline_name"`
	ScheduledDate   string                 `json:"scheduled_date"`
	IsCompleted     bool                   `json:"is_completed"`
	StatusIndicator string                 `json:"status_indicator"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

// DeploymentTimelineListResponse represents a paginated list of deployment timeline entries
type DeploymentTimelineListResponse struct {
	Timelines []DeploymentTimelineResponse `json:"timelines"`
	Total     int64                        `json:"total"`
	Page      int                          `json:"page"`
	PageSize  int                          `json:"page_size"`
}

// Create creates a new deployment timeline entry
func (s *DeploymentTimelineService) Create(req *CreateDeploymentTimelineRequest) (*DeploymentTimelineResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate organization exists
	_, err := s.orgRepo.GetByID(req.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	// Validate landscape exists
	_, err = s.landscapeRepo.GetByID(req.LandscapeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to verify landscape: %w", err)
	}

	// Convert metadata to JSON
	var metadataJSON json.RawMessage
	if req.Metadata != nil {
		jsonData, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = jsonData
	}

	// Create timeline entry
	timeline := &models.DeploymentTimeline{
		OrganizationID:  req.OrganizationID,
		LandscapeID:     req.LandscapeID,
		TimelineCode:    req.TimelineCode,
		TimelineName:    req.TimelineName,
		ScheduledDate:   req.ScheduledDate,
		IsCompleted:     req.IsCompleted,
		StatusIndicator: req.StatusIndicator,
		Metadata:        metadataJSON,
	}

	if err := s.repo.Create(timeline); err != nil {
		return nil, fmt.Errorf("failed to create deployment timeline entry: %w", err)
	}

	return s.toResponse(timeline), nil
}

// GetByID retrieves a deployment timeline entry by ID
func (s *DeploymentTimelineService) GetByID(id uuid.UUID) (*DeploymentTimelineResponse, error) {
	timeline, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentTimelineNotFound
		}
		return nil, fmt.Errorf("failed to get deployment timeline entry: %w", err)
	}

	return s.toResponse(timeline), nil
}

// GetByLandscape retrieves deployment timeline entries for a landscape
func (s *DeploymentTimelineService) GetByLandscape(landscapeID uuid.UUID, page, pageSize int) (*DeploymentTimelineListResponse, error) {
	// Validate landscape exists
	_, err := s.landscapeRepo.GetByID(landscapeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to verify landscape: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	timelines, total, err := s.repo.GetByLandscapeID(landscapeID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscape deployment timeline: %w", err)
	}

	responses := make([]DeploymentTimelineResponse, len(timelines))
	for i, timeline := range timelines {
		responses[i] = *s.toResponse(&timeline)
	}

	return &DeploymentTimelineListResponse{
		Timelines: responses,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// GetByOrganization retrieves deployment timeline entries for an organization
func (s *DeploymentTimelineService) GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*DeploymentTimelineListResponse, error) {
	// Validate organization exists
	_, err := s.orgRepo.GetByID(organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	timelines, total, err := s.repo.GetByOrganizationID(organizationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization deployment timeline: %w", err)
	}

	responses := make([]DeploymentTimelineResponse, len(timelines))
	for i, timeline := range timelines {
		responses[i] = *s.toResponse(&timeline)
	}

	return &DeploymentTimelineListResponse{
		Timelines: responses,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// GetByDateRange retrieves deployment timeline entries within a date range
func (s *DeploymentTimelineService) GetByDateRange(startDate, endDate time.Time, page, pageSize int) (*DeploymentTimelineListResponse, error) {
	// Validate date range
	if endDate.Before(startDate) {
		return nil, errors.New("end date must be after start date")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	timelines, total, err := s.repo.GetByDateRange(startDate, endDate, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment timeline entries by date range: %w", err)
	}

	responses := make([]DeploymentTimelineResponse, len(timelines))
	for i, timeline := range timelines {
		responses[i] = *s.toResponse(&timeline)
	}

	return &DeploymentTimelineListResponse{
		Timelines: responses,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// GetCompleted retrieves completed deployment timeline entries
func (s *DeploymentTimelineService) GetCompleted(page, pageSize int) (*DeploymentTimelineListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	timelines, total, err := s.repo.GetCompleted(pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed deployment timeline entries: %w", err)
	}

	responses := make([]DeploymentTimelineResponse, len(timelines))
	for i, timeline := range timelines {
		responses[i] = *s.toResponse(&timeline)
	}

	return &DeploymentTimelineListResponse{
		Timelines: responses,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// GetPending retrieves pending deployment timeline entries
func (s *DeploymentTimelineService) GetPending(page, pageSize int) (*DeploymentTimelineListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	timelines, total, err := s.repo.GetPending(pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending deployment timeline entries: %w", err)
	}

	responses := make([]DeploymentTimelineResponse, len(timelines))
	for i, timeline := range timelines {
		responses[i] = *s.toResponse(&timeline)
	}

	return &DeploymentTimelineListResponse{
		Timelines: responses,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// Update updates a deployment timeline entry
func (s *DeploymentTimelineService) Update(id uuid.UUID, req *UpdateDeploymentTimelineRequest) (*DeploymentTimelineResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing timeline
	timeline, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentTimelineNotFound
		}
		return nil, fmt.Errorf("failed to get deployment timeline entry: %w", err)
	}

	// Update fields if provided
	if req.TimelineCode != nil {
		timeline.TimelineCode = *req.TimelineCode
	}
	if req.TimelineName != nil {
		timeline.TimelineName = *req.TimelineName
	}
	if req.ScheduledDate != nil {
		timeline.ScheduledDate = *req.ScheduledDate
	}
	if req.IsCompleted != nil {
		timeline.IsCompleted = *req.IsCompleted
	}
	if req.StatusIndicator != nil {
		timeline.StatusIndicator = *req.StatusIndicator
	}
	if req.Metadata != nil {
		jsonData, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		timeline.Metadata = jsonData
	}

	if err := s.repo.Update(timeline); err != nil {
		return nil, fmt.Errorf("failed to update deployment timeline entry: %w", err)
	}

	return s.toResponse(timeline), nil
}

// MarkCompleted marks a deployment timeline entry as completed
func (s *DeploymentTimelineService) MarkCompleted(id uuid.UUID) (*DeploymentTimelineResponse, error) {
	// Get existing timeline
	timeline, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentTimelineNotFound
		}
		return nil, fmt.Errorf("failed to get deployment timeline entry: %w", err)
	}

	// Mark as completed
	timeline.IsCompleted = true

	if err := s.repo.Update(timeline); err != nil {
		return nil, fmt.Errorf("failed to mark deployment timeline entry as completed: %w", err)
	}

	return s.toResponse(timeline), nil
}

// MarkPending marks a deployment timeline entry as pending
func (s *DeploymentTimelineService) MarkPending(id uuid.UUID) (*DeploymentTimelineResponse, error) {
	// Get existing timeline
	timeline, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDeploymentTimelineNotFound
		}
		return nil, fmt.Errorf("failed to get deployment timeline entry: %w", err)
	}

	// Mark as pending
	timeline.IsCompleted = false

	if err := s.repo.Update(timeline); err != nil {
		return nil, fmt.Errorf("failed to mark deployment timeline entry as pending: %w", err)
	}

	return s.toResponse(timeline), nil
}

// Delete deletes a deployment timeline entry
func (s *DeploymentTimelineService) Delete(id uuid.UUID) error {
	// Check if timeline exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrDeploymentTimelineNotFound
		}
		return fmt.Errorf("failed to get deployment timeline entry: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete deployment timeline entry: %w", err)
	}

	return nil
}

// GetStats retrieves deployment timeline statistics
func (s *DeploymentTimelineService) GetStats() (map[string]interface{}, error) {
	// This would typically use repository methods to get stats
	// For now, return basic structure
	stats := make(map[string]interface{})
	stats["total"] = 0
	stats["completed"] = 0
	stats["pending"] = 0

	return stats, nil
}

// BulkCreate creates multiple deployment timeline entries
func (s *DeploymentTimelineService) BulkCreate(requests []CreateDeploymentTimelineRequest) ([]DeploymentTimelineResponse, []error) {
	responses := make([]DeploymentTimelineResponse, 0, len(requests))
	errors := make([]error, 0)

	for _, req := range requests {
		response, err := s.Create(&req)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if response != nil {
			responses = append(responses, *response)
		}
	}

	return responses, errors
}

// BulkUpdate updates multiple deployment timeline entries
func (s *DeploymentTimelineService) BulkUpdate(updates []struct {
	ID      uuid.UUID                       `json:"id"`
	Request UpdateDeploymentTimelineRequest `json:"request"`
}) ([]DeploymentTimelineResponse, []error) {
	responses := make([]DeploymentTimelineResponse, 0, len(updates))
	errors := make([]error, 0)

	for _, update := range updates {
		response, err := s.Update(update.ID, &update.Request)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if response != nil {
			responses = append(responses, *response)
		}
	}

	return responses, errors
}

// BulkMarkCompleted marks multiple deployment timeline entries as completed
func (s *DeploymentTimelineService) BulkMarkCompleted(ids []uuid.UUID) ([]DeploymentTimelineResponse, []error) {
	responses := make([]DeploymentTimelineResponse, 0, len(ids))
	errors := make([]error, 0)

	for _, id := range ids {
		response, err := s.MarkCompleted(id)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if response != nil {
			responses = append(responses, *response)
		}
	}

	return responses, errors
}

// BulkDelete deletes multiple deployment timeline entries
func (s *DeploymentTimelineService) BulkDelete(ids []uuid.UUID) []error {
	errors := make([]error, 0)

	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// toResponse converts a deployment timeline model to response
func (s *DeploymentTimelineService) toResponse(timeline *models.DeploymentTimeline) *DeploymentTimelineResponse {
	response := &DeploymentTimelineResponse{
		ID:              timeline.ID,
		OrganizationID:  timeline.OrganizationID,
		LandscapeID:     timeline.LandscapeID,
		TimelineCode:    timeline.TimelineCode,
		TimelineName:    timeline.TimelineName,
		ScheduledDate:   timeline.ScheduledDate.Format("2006-01-02"),
		IsCompleted:     timeline.IsCompleted,
		StatusIndicator: timeline.StatusIndicator,
		CreatedAt:       timeline.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       timeline.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Parse metadata from JSON
	if timeline.Metadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal(timeline.Metadata, &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}
