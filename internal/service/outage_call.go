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

// OutageCallService handles business logic for outage calls
type OutageCallService struct {
	repo      *repository.OutageCallRepository
	teamRepo  *repository.TeamRepository
	orgRepo   *repository.OrganizationRepository
	validator *validator.Validate
}

// NewOutageCallService creates a new outage call service
func NewOutageCallService(repo *repository.OutageCallRepository, teamRepo *repository.TeamRepository, orgRepo *repository.OrganizationRepository, validator *validator.Validate) *OutageCallService {
	return &OutageCallService{
		repo:      repo,
		teamRepo:  teamRepo,
		orgRepo:   orgRepo,
		validator: validator,
	}
}

// CreateOutageCallRequest represents the request to create an outage call
type CreateOutageCallRequest struct {
	OrganizationID        uuid.UUID                 `json:"organization_id" validate:"required"`
	TeamID                uuid.UUID                 `json:"team_id" validate:"required"`
	Title                 string                    `json:"title" validate:"required,max=200"`
	Description           string                    `json:"description,omitempty"`
	Severity              models.OutageCallSeverity `json:"severity" validate:"required"`
	Year                  int                       `json:"year" validate:"required,min=2020,max=2100"`
	CallTime              time.Time                 `json:"call_time" validate:"required"`
	ResolutionTimeMinutes *int                      `json:"resolution_time_minutes,omitempty"`
	ExternalTicketID      string                    `json:"external_ticket_id,omitempty"`
	Metadata              map[string]interface{}    `json:"metadata,omitempty"`
}

// UpdateOutageCallRequest represents the request to update an outage call
type UpdateOutageCallRequest struct {
	Title            *string                    `json:"title,omitempty" validate:"omitempty,max=200"`
	Description      *string                    `json:"description,omitempty"`
	Severity         *models.OutageCallSeverity `json:"severity,omitempty"`
	Status           *models.OutageCallStatus   `json:"status,omitempty"`
	ResolvedAt       *time.Time                 `json:"resolved_at,omitempty"`
	ExternalTicketID *string                    `json:"external_ticket_id,omitempty"`
	Metadata         map[string]interface{}     `json:"metadata,omitempty"`
}

// OutageCallResponse represents the response for outage call operations
type OutageCallResponse struct {
	ID               uuid.UUID                 `json:"id"`
	TeamID           uuid.UUID                 `json:"team_id"`
	Title            string                    `json:"title"`
	Description      string                    `json:"description"`
	Severity         models.OutageCallSeverity `json:"severity"`
	Status           models.OutageCallStatus   `json:"status"`
	StartedAt        string                    `json:"started_at"`
	ResolvedAt       *string                   `json:"resolved_at"`
	ExternalTicketID string                    `json:"external_ticket_id"`
	Metadata         map[string]interface{}    `json:"metadata"`
	CreatedAt        string                    `json:"created_at"`
	UpdatedAt        string                    `json:"updated_at"`
}

// OutageCallListResponse represents a paginated list of outage calls
type OutageCallListResponse struct {
	OutageCalls []OutageCallResponse `json:"outage_calls"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
}

// Create creates a new outage call
func (s *OutageCallService) Create(req *CreateOutageCallRequest) (*OutageCallResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate team exists
	_, err := s.teamRepo.GetByID(req.TeamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to verify team: %w", err)
	}

	// Validate organization exists
	_, err = s.orgRepo.GetByID(req.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	// Validate call time is not in the future
	if req.CallTime.After(time.Now()) {
		return nil, apperrors.ErrCallTimeInFuture
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

	// Create outage call
	call := &models.OutageCall{
		OrganizationID:   req.OrganizationID,
		TeamID:           req.TeamID,
		Title:            req.Title,
		Description:      req.Description,
		Severity:         req.Severity,
		Year:             req.Year,
		Status:           models.OutageCallStatusOpen,
		CallTimestamp:    req.CallTime,
		ExternalTicketID: req.ExternalTicketID,
		Metadata:         metadataJSON,
	}

	if req.ResolutionTimeMinutes != nil {
		call.ResolutionTimeMinutes = *req.ResolutionTimeMinutes
	}

	if err := s.repo.Create(call); err != nil {
		return nil, fmt.Errorf("failed to create outage call: %w", err)
	}

	return s.toResponse(call), nil
}

// GetByID retrieves an outage call by ID
func (s *OutageCallService) GetByID(id uuid.UUID) (*OutageCallResponse, error) {
	call, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOutageCallNotFound
		}
		return nil, fmt.Errorf("failed to get outage call: %w", err)
	}

	return s.toResponse(call), nil
}

// GetByTeam retrieves outage calls for a team
func (s *OutageCallService) GetByTeam(teamID uuid.UUID, page, pageSize int) (*OutageCallListResponse, error) {
	// Validate team exists
	_, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to verify team: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	calls, total, err := s.repo.GetByTeamID(teamID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get team outage calls: %w", err)
	}

	responses := make([]OutageCallResponse, len(calls))
	for i, call := range calls {
		responses[i] = *s.toResponse(&call)
	}

	return &OutageCallListResponse{
		OutageCalls: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetByStatus retrieves outage calls by status
func (s *OutageCallService) GetByStatus(status models.OutageCallStatus, page, pageSize int) (*OutageCallListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	calls, total, err := s.repo.GetByStatus(status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get outage calls by status: %w", err)
	}

	responses := make([]OutageCallResponse, len(calls))
	for i, call := range calls {
		responses[i] = *s.toResponse(&call)
	}

	return &OutageCallListResponse{
		OutageCalls: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetActiveCalls retrieves all active outage calls
func (s *OutageCallService) GetActiveCalls(page, pageSize int) (*OutageCallListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	calls, total, err := s.repo.GetActiveCalls(pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active outage calls: %w", err)
	}

	responses := make([]OutageCallResponse, len(calls))
	for i, call := range calls {
		responses[i] = *s.toResponse(&call)
	}

	return &OutageCallListResponse{
		OutageCalls: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetRecentCalls retrieves recent outage calls within specified days
func (s *OutageCallService) GetRecentCalls(days int, page, pageSize int) (*OutageCallListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	if days < 1 {
		days = 7 // Default to 7 days
	}

	offset := (page - 1) * pageSize
	calls, total, err := s.repo.GetRecentCalls(days, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent outage calls: %w", err)
	}

	responses := make([]OutageCallResponse, len(calls))
	for i, call := range calls {
		responses[i] = *s.toResponse(&call)
	}

	return &OutageCallListResponse{
		OutageCalls: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetBySeverity retrieves outage calls by severity
func (s *OutageCallService) GetBySeverity(severity models.OutageCallSeverity, page, pageSize int) (*OutageCallListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	calls, total, err := s.repo.GetBySeverity(severity, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get outage calls by severity: %w", err)
	}

	responses := make([]OutageCallResponse, len(calls))
	for i, call := range calls {
		responses[i] = *s.toResponse(&call)
	}

	return &OutageCallListResponse{
		OutageCalls: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetOpenCalls retrieves open outage calls
func (s *OutageCallService) GetOpenCalls(page, pageSize int) (*OutageCallListResponse, error) {
	return s.GetByStatus(models.OutageCallStatusOpen, page, pageSize)
}

// GetInProgressCalls retrieves in-progress outage calls
func (s *OutageCallService) GetInProgressCalls(page, pageSize int) (*OutageCallListResponse, error) {
	return s.GetByStatus(models.OutageCallStatusInProgress, page, pageSize)
}

// GetResolvedCalls retrieves resolved outage calls
func (s *OutageCallService) GetResolvedCalls(page, pageSize int) (*OutageCallListResponse, error) {
	return s.GetByStatus(models.OutageCallStatusResolved, page, pageSize)
}

// Update updates an outage call
func (s *OutageCallService) Update(id uuid.UUID, req *UpdateOutageCallRequest) (*OutageCallResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing call
	call, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOutageCallNotFound
		}
		return nil, fmt.Errorf("failed to get outage call: %w", err)
	}

	// Update fields if provided
	if req.Title != nil {
		call.Title = *req.Title
	}
	if req.Description != nil {
		call.Description = *req.Description
	}
	if req.Severity != nil {
		call.Severity = *req.Severity
	}
	if req.Status != nil {
		call.Status = *req.Status
		// Auto-set resolved time if status is resolved
		if *req.Status == models.OutageCallStatusResolved && call.ResolvedAt == nil {
			now := time.Now()
			call.ResolvedAt = &now
		}
	}
	if req.ResolvedAt != nil {
		call.ResolvedAt = req.ResolvedAt
	}
	if req.ExternalTicketID != nil {
		call.ExternalTicketID = *req.ExternalTicketID
	}
	if req.Metadata != nil {
		jsonData, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		call.Metadata = jsonData
	}

	if err := s.repo.Update(call); err != nil {
		return nil, fmt.Errorf("failed to update outage call: %w", err)
	}

	return s.toResponse(call), nil
}

// SetStatus sets the status of an outage call
func (s *OutageCallService) SetStatus(id uuid.UUID, status models.OutageCallStatus) (*OutageCallResponse, error) {
	// Check if call exists
	call, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOutageCallNotFound
		}
		return nil, fmt.Errorf("failed to get outage call: %w", err)
	}

	// Update status
	call.Status = status

	// Auto-set resolved time if status is resolved
	if status == models.OutageCallStatusResolved && call.ResolvedAt == nil {
		now := time.Now()
		call.ResolvedAt = &now
	}

	if err := s.repo.Update(call); err != nil {
		return nil, fmt.Errorf("failed to update outage call status: %w", err)
	}

	return s.toResponse(call), nil
}

// Resolve resolves an outage call
func (s *OutageCallService) Resolve(id uuid.UUID) (*OutageCallResponse, error) {
	return s.SetStatus(id, models.OutageCallStatusResolved)
}

// Cancel cancels an outage call
func (s *OutageCallService) Cancel(id uuid.UUID) (*OutageCallResponse, error) {
	return s.SetStatus(id, models.OutageCallStatusCancelled)
}

// Delete deletes an outage call
func (s *OutageCallService) Delete(id uuid.UUID) error {
	// Check if call exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrOutageCallNotFound
		}
		return fmt.Errorf("failed to get outage call: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete outage call: %w", err)
	}

	return nil
}

// GetStats retrieves outage call statistics
func (s *OutageCallService) GetStats() (map[string]int64, error) {
	stats, err := s.repo.GetOutageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get outage stats: %w", err)
	}

	return stats, nil
}

// GetWithAssignees retrieves an outage call with assignees
func (s *OutageCallService) GetWithAssignees(id uuid.UUID) (*OutageCallResponse, error) {
	call, err := s.repo.GetWithAssignees(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOutageCallNotFound
		}
		return nil, fmt.Errorf("failed to get outage call with assignees: %w", err)
	}

	return s.toResponse(call), nil
}

// GetWithTeam retrieves an outage call with team details
func (s *OutageCallService) GetWithTeam(id uuid.UUID) (*OutageCallResponse, error) {
	call, err := s.repo.GetWithTeam(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOutageCallNotFound
		}
		return nil, fmt.Errorf("failed to get outage call with team: %w", err)
	}

	return s.toResponse(call), nil
}

// BulkCreate creates multiple outage calls
func (s *OutageCallService) BulkCreate(requests []CreateOutageCallRequest) ([]OutageCallResponse, []error) {
	responses := make([]OutageCallResponse, 0, len(requests))
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

// BulkUpdate updates multiple outage calls
func (s *OutageCallService) BulkUpdate(updates []struct {
	ID      uuid.UUID               `json:"id"`
	Request UpdateOutageCallRequest `json:"request"`
}) ([]OutageCallResponse, []error) {
	responses := make([]OutageCallResponse, 0, len(updates))
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

// BulkSetStatus sets status for multiple outage calls
func (s *OutageCallService) BulkSetStatus(ids []uuid.UUID, status models.OutageCallStatus) ([]OutageCallResponse, []error) {
	responses := make([]OutageCallResponse, 0, len(ids))
	errors := make([]error, 0)

	for _, id := range ids {
		response, err := s.SetStatus(id, status)
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

// BulkResolve resolves multiple outage calls
func (s *OutageCallService) BulkResolve(ids []uuid.UUID) ([]OutageCallResponse, []error) {
	return s.BulkSetStatus(ids, models.OutageCallStatusResolved)
}

// BulkDelete deletes multiple outage calls
func (s *OutageCallService) BulkDelete(ids []uuid.UUID) []error {
	errors := make([]error, 0)

	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// toResponse converts an outage call model to response
func (s *OutageCallService) toResponse(call *models.OutageCall) *OutageCallResponse {
	response := &OutageCallResponse{
		ID:               call.ID,
		TeamID:           call.TeamID,
		Title:            call.Title,
		Description:      call.Description,
		Severity:         call.Severity,
		Status:           call.Status,
		StartedAt:        call.CallTimestamp.Format("2006-01-02T15:04:05Z07:00"),
		ExternalTicketID: call.ExternalTicketID,
		CreatedAt:        call.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        call.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if call.ResolvedAt != nil {
		resolvedAt := call.ResolvedAt.Format("2006-01-02T15:04:05Z07:00")
		response.ResolvedAt = &resolvedAt
	}

	// Parse metadata from JSON
	if call.Metadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal(call.Metadata, &metadata); err == nil {
			response.Metadata = metadata
		}
	}

	return response
}
