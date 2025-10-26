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

// OutageCallAssigneeService handles business logic for outage call assignees
type OutageCallAssigneeService struct {
	repo           *repository.OutageCallAssigneeRepository
	outageCallRepo *repository.OutageCallRepository
	memberRepo     *repository.MemberRepository
	validator      *validator.Validate
}

// NewOutageCallAssigneeService creates a new outage call assignee service
func NewOutageCallAssigneeService(
	repo *repository.OutageCallAssigneeRepository,
	outageCallRepo *repository.OutageCallRepository,
	memberRepo *repository.MemberRepository,
	validator *validator.Validate,
) *OutageCallAssigneeService {
	return &OutageCallAssigneeService{
		repo:           repo,
		outageCallRepo: outageCallRepo,
		memberRepo:     memberRepo,
		validator:      validator,
	}
}

// CreateOutageCallAssigneeRequest represents the request to create an outage call assignee
type CreateOutageCallAssigneeRequest struct {
	OutageCallID uuid.UUID           `json:"outage_call_id" validate:"required"`
	MemberID     uuid.UUID           `json:"member_id" validate:"required"`
	Role         models.AssigneeRole `json:"role" validate:"required"`
	AssignedAt   *time.Time          `json:"assigned_at,omitempty"`
	IsActive     *bool               `json:"is_active,omitempty"`
}

// UpdateOutageCallAssigneeRequest represents the request to update an outage call assignee
type UpdateOutageCallAssigneeRequest struct {
	Role     *models.AssigneeRole `json:"role,omitempty"`
	IsActive *bool                `json:"is_active,omitempty"`
}

// OutageCallAssigneeResponse represents the response for outage call assignee operations
type OutageCallAssigneeResponse struct {
	ID           uuid.UUID           `json:"id"`
	OutageCallID uuid.UUID           `json:"outage_call_id"`
	MemberID     uuid.UUID           `json:"member_id"`
	Role         models.AssigneeRole `json:"role"`
	AssignedAt   string              `json:"assigned_at"`
	IsActive     bool                `json:"is_active"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
}

// AssignMemberRequest represents the request to assign a member to an outage call
type AssignMemberRequest struct {
	MemberID uuid.UUID           `json:"member_id" validate:"required"`
	Role     models.AssigneeRole `json:"role" validate:"required"`
}

// BulkAssignRequest represents the request to assign multiple members to an outage call
type BulkAssignRequest struct {
	OutageCallID uuid.UUID             `json:"outage_call_id" validate:"required"`
	Members      []AssignMemberRequest `json:"members" validate:"required,min=1"`
}

// Create creates a new outage call assignee
func (s *OutageCallAssigneeService) Create(req *CreateOutageCallAssigneeRequest) (*OutageCallAssigneeResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate outage call exists
	_, err := s.outageCallRepo.GetByID(req.OutageCallID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOutageCallNotFound
		}
		return nil, fmt.Errorf("failed to verify outage call: %w", err)
	}

	// Validate member exists
	_, err = s.memberRepo.GetByID(req.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to verify member: %w", err)
	}

	// Check if assignment already exists
	exists, err := s.repo.Exists(req.OutageCallID, req.MemberID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing assignment: %w", err)
	}
	if exists {
		return nil, apperrors.ErrMemberAlreadyAssigned
	}

	// Set default values
	assignedAt := time.Now()
	if req.AssignedAt != nil {
		assignedAt = *req.AssignedAt
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create assignee
	assignee := &models.OutageCallAssignee{
		OutageCallID: req.OutageCallID,
		MemberID:     req.MemberID,
		Role:         req.Role,
		AssignedAt:   assignedAt,
		IsActive:     isActive,
	}

	if err := s.repo.Create(assignee); err != nil {
		return nil, fmt.Errorf("failed to create outage call assignee: %w", err)
	}

	return s.toResponse(assignee), nil
}

// GetByOutageCall retrieves all assignees for an outage call
func (s *OutageCallAssigneeService) GetByOutageCall(outageCallID uuid.UUID) ([]OutageCallAssigneeResponse, error) {
	// Validate outage call exists
	_, err := s.outageCallRepo.GetByID(outageCallID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOutageCallNotFound
		}
		return nil, fmt.Errorf("failed to verify outage call: %w", err)
	}

	assignees, err := s.repo.GetByOutageCallID(outageCallID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outage call assignees: %w", err)
	}

	responses := make([]OutageCallAssigneeResponse, len(assignees))
	for i, assignee := range assignees {
		responses[i] = *s.toResponse(&assignee)
	}

	return responses, nil
}

// GetByMember retrieves all outage call assignments for a member
func (s *OutageCallAssigneeService) GetByMember(memberID uuid.UUID) ([]OutageCallAssigneeResponse, error) {
	// Validate member exists
	_, err := s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrMemberNotFound
		}
		return nil, fmt.Errorf("failed to verify member: %w", err)
	}

	assignees, err := s.repo.GetByMemberID(memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member outage call assignments: %w", err)
	}

	responses := make([]OutageCallAssigneeResponse, len(assignees))
	for i, assignee := range assignees {
		responses[i] = *s.toResponse(&assignee)
	}

	return responses, nil
}

// GetActiveByMember retrieves active outage call assignments for a member
func (s *OutageCallAssigneeService) GetActiveByMember(memberID uuid.UUID) ([]OutageCallAssigneeResponse, error) {
	assignees, err := s.GetByMember(memberID)
	if err != nil {
		return nil, err
	}

	// Filter active assignments
	activeAssignees := make([]OutageCallAssigneeResponse, 0)
	for _, assignee := range assignees {
		if assignee.IsActive {
			activeAssignees = append(activeAssignees, assignee)
		}
	}

	return activeAssignees, nil
}

// GetByRole retrieves assignees by role for an outage call
func (s *OutageCallAssigneeService) GetByRole(outageCallID uuid.UUID, role models.AssigneeRole) ([]OutageCallAssigneeResponse, error) {
	assignees, err := s.GetByOutageCall(outageCallID)
	if err != nil {
		return nil, err
	}

	// Filter by role
	roleAssignees := make([]OutageCallAssigneeResponse, 0)
	for _, assignee := range assignees {
		if assignee.Role == role {
			roleAssignees = append(roleAssignees, assignee)
		}
	}

	return roleAssignees, nil
}

// GetPrimaryAssignees retrieves primary assignees for an outage call
func (s *OutageCallAssigneeService) GetPrimaryAssignees(outageCallID uuid.UUID) ([]OutageCallAssigneeResponse, error) {
	return s.GetByRole(outageCallID, models.AssigneeRolePrimary)
}

// GetSecondaryAssignees retrieves secondary assignees for an outage call
func (s *OutageCallAssigneeService) GetSecondaryAssignees(outageCallID uuid.UUID) ([]OutageCallAssigneeResponse, error) {
	return s.GetByRole(outageCallID, models.AssigneeRoleSecondary)
}

// GetObservers retrieves observers for an outage call
func (s *OutageCallAssigneeService) GetObservers(outageCallID uuid.UUID) ([]OutageCallAssigneeResponse, error) {
	return s.GetByRole(outageCallID, models.AssigneeRoleObserver)
}

// AssignMember assigns a member to an outage call
func (s *OutageCallAssigneeService) AssignMember(outageCallID uuid.UUID, req *AssignMemberRequest) (*OutageCallAssigneeResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	createReq := &CreateOutageCallAssigneeRequest{
		OutageCallID: outageCallID,
		MemberID:     req.MemberID,
		Role:         req.Role,
	}

	return s.Create(createReq)
}

// UnassignMember removes a member from an outage call
func (s *OutageCallAssigneeService) UnassignMember(outageCallID, memberID uuid.UUID) error {
	// Validate outage call exists
	_, err := s.outageCallRepo.GetByID(outageCallID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrOutageCallNotFound
		}
		return fmt.Errorf("failed to verify outage call: %w", err)
	}

	// Validate member exists
	_, err = s.memberRepo.GetByID(memberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrMemberNotFound
		}
		return fmt.Errorf("failed to verify member: %w", err)
	}

	// Check if assignment exists
	exists, err := s.repo.Exists(outageCallID, memberID)
	if err != nil {
		return fmt.Errorf("failed to check assignment existence: %w", err)
	}
	if !exists {
		return apperrors.ErrMemberNotAssigned
	}

	if err := s.repo.Delete(outageCallID, memberID); err != nil {
		return fmt.Errorf("failed to unassign member: %w", err)
	}

	return nil
}

// BulkAssign assigns multiple members to an outage call
func (s *OutageCallAssigneeService) BulkAssign(req *BulkAssignRequest) ([]OutageCallAssigneeResponse, []error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, []error{fmt.Errorf("validation failed: %w", err)}
	}

	responses := make([]OutageCallAssigneeResponse, 0, len(req.Members))
	errors := make([]error, 0)

	for _, member := range req.Members {
		response, err := s.AssignMember(req.OutageCallID, &member)
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

// BulkUnassign removes multiple members from an outage call
func (s *OutageCallAssigneeService) BulkUnassign(outageCallID uuid.UUID, memberIDs []uuid.UUID) []error {
	errors := make([]error, 0)

	for _, memberID := range memberIDs {
		if err := s.UnassignMember(outageCallID, memberID); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// UpdateRole updates the role of an assignee
func (s *OutageCallAssigneeService) UpdateRole(outageCallID, memberID uuid.UUID, role models.AssigneeRole) (*OutageCallAssigneeResponse, error) {
	// Check if assignment exists
	exists, err := s.repo.Exists(outageCallID, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to check assignment existence: %w", err)
	}
	if !exists {
		return nil, apperrors.ErrMemberNotAssigned
	}

	// Since the repository doesn't have an update method, we need to delete and recreate
	// This is a limitation of the current repository design
	if err := s.repo.Delete(outageCallID, memberID); err != nil {
		return nil, fmt.Errorf("failed to remove existing assignment: %w", err)
	}

	// Create new assignment with updated role
	createReq := &CreateOutageCallAssigneeRequest{
		OutageCallID: outageCallID,
		MemberID:     memberID,
		Role:         role,
	}

	return s.Create(createReq)
}

// SetActive sets the active status of an assignee
func (s *OutageCallAssigneeService) SetActive(outageCallID, memberID uuid.UUID, isActive bool) (*OutageCallAssigneeResponse, error) {
	// Since the repository doesn't have update functionality, we'll need to implement this
	// using the same delete/recreate pattern. This is a limitation of the current design.

	// Check if assignment exists
	exists, err := s.repo.Exists(outageCallID, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to check assignment existence: %w", err)
	}
	if !exists {
		return nil, apperrors.ErrMemberNotAssigned
	}

	// For now, return an error indicating this operation needs repository enhancement
	return nil, errors.New("updating assignee status requires repository enhancement")
}

// IsAssigned checks if a member is assigned to an outage call
func (s *OutageCallAssigneeService) IsAssigned(outageCallID, memberID uuid.UUID) (bool, error) {
	return s.repo.Exists(outageCallID, memberID)
}

// GetAssignmentCount returns the number of assignees for an outage call
func (s *OutageCallAssigneeService) GetAssignmentCount(outageCallID uuid.UUID) (int, error) {
	assignees, err := s.repo.GetByOutageCallID(outageCallID)
	if err != nil {
		return 0, fmt.Errorf("failed to get assignees: %w", err)
	}

	return len(assignees), nil
}

// GetAssignmentCountByRole returns the number of assignees by role for an outage call
func (s *OutageCallAssigneeService) GetAssignmentCountByRole(outageCallID uuid.UUID, role models.AssigneeRole) (int, error) {
	assignees, err := s.GetByRole(outageCallID, role)
	if err != nil {
		return 0, err
	}

	return len(assignees), nil
}

// GetMemberWorkload returns the number of active outage calls a member is assigned to
func (s *OutageCallAssigneeService) GetMemberWorkload(memberID uuid.UUID) (int, error) {
	assignees, err := s.GetActiveByMember(memberID)
	if err != nil {
		return 0, err
	}

	return len(assignees), nil
}

// toResponse converts an outage call assignee model to response
func (s *OutageCallAssigneeService) toResponse(assignee *models.OutageCallAssignee) *OutageCallAssigneeResponse {
	return &OutageCallAssigneeResponse{
		ID:           assignee.ID,
		OutageCallID: assignee.OutageCallID,
		MemberID:     assignee.MemberID,
		Role:         assignee.Role,
		AssignedAt:   assignee.AssignedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsActive:     assignee.IsActive,
		CreatedAt:    assignee.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    assignee.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
