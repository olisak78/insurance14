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

// GroupService handles business logic for groups
type GroupService struct {
	repo      repository.GroupRepositoryInterface
	orgRepo   repository.OrganizationRepositoryInterface
	validator *validator.Validate
}

// NewGroupService creates a new group service
func NewGroupService(repo repository.GroupRepositoryInterface, orgRepo repository.OrganizationRepositoryInterface, validator *validator.Validate) *GroupService {
	return &GroupService{
		repo:      repo,
		orgRepo:   orgRepo,
		validator: validator,
	}
}

// CreateGroupRequest represents the request to create a group
type CreateGroupRequest struct {
	OrganizationID uuid.UUID       `json:"organization_id" validate:"required"`
	Name           string          `json:"name" validate:"required,min=1,max=100"`
	DisplayName    string          `json:"display_name" validate:"required,max=200"`
	Description    string          `json:"description"`
	Metadata       json.RawMessage `json:"metadata" swaggertype:"object"`
}

// UpdateGroupRequest represents the request to update a group
type UpdateGroupRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=100"`
	DisplayName string          `json:"display_name" validate:"required,max=200"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata" swaggertype:"object"`
}

// GroupResponse represents the response for group operations
type GroupResponse struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	Name           string          `json:"name"`
	DisplayName    string          `json:"display_name"`
	Description    string          `json:"description"`
	Metadata       json.RawMessage `json:"metadata" swaggertype:"object"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// GroupListResponse represents a paginated list of groups
type GroupListResponse struct {
	Groups   []GroupResponse `json:"groups"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// GroupWithTeamsResponse represents a group with its teams
type GroupWithTeamsResponse struct {
	Group    GroupResponse  `json:"group"`
	Teams    []TeamResponse `json:"teams"`
	Total    int64          `json:"total_teams"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// Create creates a new group
func (s *GroupService) Create(req *CreateGroupRequest) (*GroupResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if organization exists
	_, err := s.orgRepo.GetByID(req.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	// Check if group with same name exists in the organization
	existing, err := s.repo.GetByName(req.OrganizationID, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing group: %w", err)
	}
	if existing != nil {
		return nil, apperrors.ErrGroupExists
	}

	// Create group
	group := &models.Group{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		Metadata:       req.Metadata,
	}

	if err := s.repo.Create(group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return s.toResponse(group), nil
}

// GetByID retrieves a group by ID
func (s *GroupService) GetByID(id uuid.UUID) (*GroupResponse, error) {
	group, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return s.toResponse(group), nil
}

// GetAll retrieves all groups for an organization with pagination
func (s *GroupService) GetAll(orgID uuid.UUID, page, pageSize int) (*GroupListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Check if organization exists
	_, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	offset := (page - 1) * pageSize
	groups, total, err := s.repo.GetByOrganizationID(orgID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	responses := make([]GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = *s.toResponse(&group)
	}

	return &GroupListResponse{
		Groups:   responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update updates a group
func (s *GroupService) Update(id uuid.UUID, req *UpdateGroupRequest) (*GroupResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing group
	existingGroup, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Check if another group with same name exists in the organization
	existing, err := s.repo.GetByName(existingGroup.OrganizationID, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing group: %w", err)
	}
	if existing != nil && existing.ID != id {
		return nil, apperrors.ErrGroupExists
	}

	// Prepare updates
	updates := map[string]interface{}{
		"name":         req.Name,
		"display_name": req.DisplayName,
		"description":  req.Description,
		"metadata":     req.Metadata,
	}

	if err := s.repo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	// Get updated group
	updatedGroup, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated group: %w", err)
	}

	return s.toResponse(updatedGroup), nil
}

// Delete deletes a group
func (s *GroupService) Delete(id uuid.UUID) error {
	// Check if group exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrGroupNotFound
		}
		return fmt.Errorf("failed to get group: %w", err)
	}

	// TODO: In a real implementation, you might want to check if there are teams
	// associated with this group and handle them appropriately (reassign or prevent deletion)

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// GetByName retrieves a group by name within an organization
func (s *GroupService) GetByName(organizationID uuid.UUID, name string) (*GroupResponse, error) {
	group, err := s.repo.GetByName(organizationID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return s.toResponse(group), nil
}

// GetByOrganization retrieves groups for an organization with pagination
func (s *GroupService) GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*GroupListResponse, error) {
	return s.GetAll(organizationID, page, pageSize)
}

// GetWithTeams retrieves a group with its teams
func (s *GroupService) GetWithTeams(id uuid.UUID, page, pageSize int) (*GroupWithTeamsResponse, error) {
	// Get the group
	group, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Get teams for this group (would need to implement this in team repo)
	// For now, return empty teams list
	// TODO: Implement team fetching through team repository
	teams := []TeamResponse{}
	total := int64(0)

	return &GroupWithTeamsResponse{
		Group:    *s.toResponse(group),
		Teams:    teams,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Search searches for groups by name or description within an organization
func (s *GroupService) Search(organizationID uuid.UUID, query string, page, pageSize int) (*GroupListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Check if organization exists
	_, err := s.orgRepo.GetByID(organizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	offset := (page - 1) * pageSize
	groups, total, err := s.repo.Search(organizationID, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}

	responses := make([]GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = *s.toResponse(&group)
	}

	return &GroupListResponse{
		Groups:   responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// toResponse converts a group model to response
func (s *GroupService) toResponse(group *models.Group) *GroupResponse {
	return &GroupResponse{
		ID:             group.ID,
		OrganizationID: group.OrganizationID,
		Name:           group.Name,
		DisplayName:    group.DisplayName,
		Description:    group.Description,
		Metadata:       group.Metadata,
		CreatedAt:      group.CreatedAt,
		UpdatedAt:      group.UpdatedAt,
	}
}
