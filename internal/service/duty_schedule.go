package service

import (
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

// DutyScheduleService handles business logic for duty schedules
type DutyScheduleService struct {
	repo       *repository.DutyScheduleRepository
	teamRepo   *repository.TeamRepository
	memberRepo *repository.MemberRepository
	orgRepo    *repository.OrganizationRepository
	validator  *validator.Validate
}

// NewDutyScheduleService creates a new duty schedule service
func NewDutyScheduleService(repo *repository.DutyScheduleRepository, teamRepo *repository.TeamRepository, memberRepo *repository.MemberRepository, orgRepo *repository.OrganizationRepository, validator *validator.Validate) *DutyScheduleService {
	return &DutyScheduleService{
		repo:       repo,
		teamRepo:   teamRepo,
		memberRepo: memberRepo,
		orgRepo:    orgRepo,
		validator:  validator,
	}
}

// CreateDutyScheduleRequest represents the request to create a duty schedule
type CreateDutyScheduleRequest struct {
	OrganizationID uuid.UUID              `json:"organization_id" validate:"required"`
	TeamID         uuid.UUID              `json:"team_id" validate:"required"`
	MemberID       uuid.UUID              `json:"member_id" validate:"required"`
	ScheduleType   models.ScheduleType    `json:"schedule_type" validate:"required"`
	Year           int                    `json:"year" validate:"required,min=2020,max=2100"`
	StartDate      time.Time              `json:"start_date" validate:"required"`
	EndDate        time.Time              `json:"end_date" validate:"required"`
	ShiftType      *models.ShiftType      `json:"shift_type,omitempty"`
	WasCalled      *bool                  `json:"was_called,omitempty"`
	Notes          string                 `json:"notes,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateDutyScheduleRequest represents the request to update a duty schedule
type UpdateDutyScheduleRequest struct {
	ScheduleType *models.ScheduleType   `json:"schedule_type,omitempty"`
	Year         *int                   `json:"year,omitempty"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	EndDate      *time.Time             `json:"end_date,omitempty"`
	ShiftType    *models.ShiftType      `json:"shift_type,omitempty"`
	WasCalled    *bool                  `json:"was_called,omitempty"`
	Notes        *string                `json:"notes,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DutyScheduleResponse represents the response for duty schedule operations
type DutyScheduleResponse struct {
	ID             uuid.UUID           `json:"id"`
	OrganizationID uuid.UUID           `json:"organization_id"`
	TeamID         uuid.UUID           `json:"team_id"`
	MemberID       uuid.UUID           `json:"member_id"`
	ScheduleType   models.ScheduleType `json:"schedule_type"`
	Year           int                 `json:"year"`
	StartDate      string              `json:"start_date"`
	EndDate        string              `json:"end_date"`
	ShiftType      *models.ShiftType   `json:"shift_type,omitempty"`
	WasCalled      bool                `json:"was_called"`
	Notes          string              `json:"notes"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

// DutyScheduleListResponse represents a paginated list of duty schedules
type DutyScheduleListResponse struct {
	DutySchedules []DutyScheduleResponse `json:"duty_schedules"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
}

// Create creates a new duty schedule
func (s *DutyScheduleService) Create(req *CreateDutyScheduleRequest) (*DutyScheduleResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate schedule type
	if !req.ScheduleType.IsValid() {
		return nil, errors.New("invalid schedule type")
	}

	// Validate shift type if provided
	if req.ShiftType != nil && !req.ShiftType.IsValid() {
		return nil, errors.New("invalid shift type")
	}

	// Validate date range
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	// Validate organization exists
	_, err := s.orgRepo.GetByID(req.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	// Validate team exists
	_, err = s.teamRepo.GetByID(req.TeamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to verify team: %w", err)
	}

	// Validate member exists
	_, err = s.memberRepo.GetByID(req.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to verify member: %w", err)
	}

	// Set default values
	wasCalled := false
	if req.WasCalled != nil {
		wasCalled = *req.WasCalled
	}

	// Create duty schedule
	schedule := &models.DutySchedule{
		OrganizationID: req.OrganizationID,
		TeamID:         req.TeamID,
		MemberID:       req.MemberID,
		ScheduleType:   req.ScheduleType,
		Year:           req.Year,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		WasCalled:      wasCalled,
		Notes:          req.Notes,
	}

	if req.ShiftType != nil {
		schedule.ShiftType = *req.ShiftType
	}

	if err := s.repo.Create(schedule); err != nil {
		return nil, fmt.Errorf("failed to create duty schedule: %w", err)
	}

	return s.toResponse(schedule), nil
}

// GetByID retrieves a duty schedule by ID
func (s *DutyScheduleService) GetByID(id uuid.UUID) (*DutyScheduleResponse, error) {
	schedule, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDutyScheduleNotFound
		}
		return nil, fmt.Errorf("failed to get duty schedule: %w", err)
	}

	return s.toResponse(schedule), nil
}

// GetByTeam retrieves duty schedules for a team
func (s *DutyScheduleService) GetByTeam(teamID uuid.UUID, page, pageSize int) (*DutyScheduleListResponse, error) {
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
	schedules, total, err := s.repo.GetByTeamID(teamID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get team duty schedules: %w", err)
	}

	responses := make([]DutyScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		responses[i] = *s.toResponse(&schedule)
	}

	return &DutyScheduleListResponse{
		DutySchedules: responses,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// Update updates a duty schedule
func (s *DutyScheduleService) Update(id uuid.UUID, req *UpdateDutyScheduleRequest) (*DutyScheduleResponse, error) {
	// Get existing schedule
	schedule, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrDutyScheduleNotFound
		}
		return nil, fmt.Errorf("failed to get duty schedule: %w", err)
	}

	// Update fields if provided
	if req.ScheduleType != nil {
		if !req.ScheduleType.IsValid() {
			return nil, errors.New("invalid schedule type")
		}
		schedule.ScheduleType = *req.ScheduleType
	}
	if req.Year != nil {
		if *req.Year < 2020 || *req.Year > 2100 {
			return nil, errors.New("invalid year")
		}
		schedule.Year = *req.Year
	}
	if req.StartDate != nil {
		schedule.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		schedule.EndDate = *req.EndDate
	}
	if req.ShiftType != nil {
		if !req.ShiftType.IsValid() {
			return nil, errors.New("invalid shift type")
		}
		schedule.ShiftType = *req.ShiftType
	}
	if req.WasCalled != nil {
		schedule.WasCalled = *req.WasCalled
	}
	if req.Notes != nil {
		schedule.Notes = *req.Notes
	}

	// Validate date range
	if schedule.EndDate.Before(schedule.StartDate) {
		return nil, errors.New("end date must be after start date")
	}

	if err := s.repo.Update(schedule); err != nil {
		return nil, fmt.Errorf("failed to update duty schedule: %w", err)
	}

	return s.toResponse(schedule), nil
}

// Delete deletes a duty schedule
func (s *DutyScheduleService) Delete(id uuid.UUID) error {
	// Check if schedule exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrDutyScheduleNotFound
		}
		return fmt.Errorf("failed to get duty schedule: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete duty schedule: %w", err)
	}

	return nil
}

// toResponse converts a duty schedule model to response
func (s *DutyScheduleService) toResponse(schedule *models.DutySchedule) *DutyScheduleResponse {
	response := &DutyScheduleResponse{
		ID:             schedule.ID,
		OrganizationID: schedule.OrganizationID,
		TeamID:         schedule.TeamID,
		MemberID:       schedule.MemberID,
		ScheduleType:   schedule.ScheduleType,
		Year:           schedule.Year,
		StartDate:      schedule.StartDate.Format("2006-01-02"),
		EndDate:        schedule.EndDate.Format("2006-01-02"),
		WasCalled:      schedule.WasCalled,
		Notes:          schedule.Notes,
		CreatedAt:      schedule.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      schedule.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Handle optional shift type
	if schedule.ShiftType != "" {
		response.ShiftType = &schedule.ShiftType
	}

	return response
}
