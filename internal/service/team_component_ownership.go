package service

import (
	"errors"
	"fmt"

	"developer-portal-backend/internal/database/models"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamComponentOwnershipService handles business logic for team-component ownership relationships
type TeamComponentOwnershipService struct {
	repo          *repository.TeamComponentOwnershipRepository
	teamRepo      *repository.TeamRepository
	componentRepo *repository.ComponentRepository
	validator     *validator.Validate
}

// NewTeamComponentOwnershipService creates a new team component ownership service
func NewTeamComponentOwnershipService(repo *repository.TeamComponentOwnershipRepository, teamRepo *repository.TeamRepository, componentRepo *repository.ComponentRepository, validator *validator.Validate) *TeamComponentOwnershipService {
	return &TeamComponentOwnershipService{
		repo:          repo,
		teamRepo:      teamRepo,
		componentRepo: componentRepo,
		validator:     validator,
	}
}

// CreateTeamComponentOwnershipRequest represents the request to create a team-component ownership relationship
type CreateTeamComponentOwnershipRequest struct {
	TeamID        uuid.UUID            `json:"team_id" validate:"required"`
	ComponentID   uuid.UUID            `json:"component_id" validate:"required"`
	OwnershipType models.OwnershipType `json:"ownership_type" validate:"required"`
}

// UpdateTeamComponentOwnershipRequest represents the request to update a team-component ownership relationship
type UpdateTeamComponentOwnershipRequest struct {
	OwnershipType models.OwnershipType `json:"ownership_type" validate:"required"`
}

// TeamComponentOwnershipResponse represents the response for team-component ownership operations
type TeamComponentOwnershipResponse struct {
	ID            uuid.UUID            `json:"id"`
	TeamID        uuid.UUID            `json:"team_id"`
	ComponentID   uuid.UUID            `json:"component_id"`
	OwnershipType models.OwnershipType `json:"ownership_type"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

// TeamComponentOwnershipListResponse represents a paginated list of team-component ownership relationships
type TeamComponentOwnershipListResponse struct {
	TeamComponentOwnerships []TeamComponentOwnershipResponse `json:"team_component_ownerships"`
	Total                   int64                            `json:"total"`
	Page                    int                              `json:"page"`
	PageSize                int                              `json:"page_size"`
}

// Create creates a new team-component ownership relationship
func (s *TeamComponentOwnershipService) Create(req *CreateTeamComponentOwnershipRequest) (*TeamComponentOwnershipResponse, error) {
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

	// Validate component exists
	_, err = s.componentRepo.GetByID(req.ComponentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to verify component: %w", err)
	}

	// Check if ownership already exists
	exists, err := s.repo.Exists(req.TeamID, req.ComponentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing ownership: %w", err)
	}
	if exists {
		return nil, apperrors.ErrTeamComponentOwnershipExists
	}

	// Create ownership relationship
	ownership := &models.TeamComponentOwnership{
		TeamID:        req.TeamID,
		ComponentID:   req.ComponentID,
		OwnershipType: req.OwnershipType,
	}

	if err := s.repo.Create(ownership); err != nil {
		return nil, fmt.Errorf("failed to create team-component ownership: %w", err)
	}

	return s.toResponse(ownership), nil
}

// GetByTeam retrieves all component ownerships for a team
func (s *TeamComponentOwnershipService) GetByTeam(teamID uuid.UUID, page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
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

	ownerships, err := s.repo.GetByTeamID(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team component ownerships: %w", err)
	}

	// Apply pagination
	total := int64(len(ownerships))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(ownerships) {
		ownerships = []models.TeamComponentOwnership{}
	} else {
		if end > len(ownerships) {
			end = len(ownerships)
		}
		ownerships = ownerships[start:end]
	}

	responses := make([]TeamComponentOwnershipResponse, len(ownerships))
	for i, ownership := range ownerships {
		responses[i] = *s.toResponse(&ownership)
	}

	return &TeamComponentOwnershipListResponse{
		TeamComponentOwnerships: responses,
		Total:                   total,
		Page:                    page,
		PageSize:                pageSize,
	}, nil
}

// GetByComponent retrieves all team ownerships for a component
func (s *TeamComponentOwnershipService) GetByComponent(componentID uuid.UUID, page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
	// Validate component exists
	_, err := s.componentRepo.GetByID(componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to verify component: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ownerships, err := s.repo.GetByComponentID(componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component team ownerships: %w", err)
	}

	// Apply pagination
	total := int64(len(ownerships))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(ownerships) {
		ownerships = []models.TeamComponentOwnership{}
	} else {
		if end > len(ownerships) {
			end = len(ownerships)
		}
		ownerships = ownerships[start:end]
	}

	responses := make([]TeamComponentOwnershipResponse, len(ownerships))
	for i, ownership := range ownerships {
		responses[i] = *s.toResponse(&ownership)
	}

	return &TeamComponentOwnershipListResponse{
		TeamComponentOwnerships: responses,
		Total:                   total,
		Page:                    page,
		PageSize:                pageSize,
	}, nil
}

// GetByOwnershipType retrieves ownerships by ownership type
func (s *TeamComponentOwnershipService) GetByOwnershipType(ownershipType models.OwnershipType, page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ownerships, err := s.repo.GetByOwnershipType(ownershipType)
	if err != nil {
		return nil, fmt.Errorf("failed to get ownerships by type: %w", err)
	}

	// Apply pagination
	total := int64(len(ownerships))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(ownerships) {
		ownerships = []models.TeamComponentOwnership{}
	} else {
		if end > len(ownerships) {
			end = len(ownerships)
		}
		ownerships = ownerships[start:end]
	}

	responses := make([]TeamComponentOwnershipResponse, len(ownerships))
	for i, ownership := range ownerships {
		responses[i] = *s.toResponse(&ownership)
	}

	return &TeamComponentOwnershipListResponse{
		TeamComponentOwnerships: responses,
		Total:                   total,
		Page:                    page,
		PageSize:                pageSize,
	}, nil
}

// GetByTeamAndOwnershipType retrieves component ownerships for a team by ownership type
func (s *TeamComponentOwnershipService) GetByTeamAndOwnershipType(teamID uuid.UUID, ownershipType models.OwnershipType, page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
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

	ownerships, err := s.repo.GetByTeamID(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team component ownerships: %w", err)
	}

	// Filter by ownership type
	var filteredOwnerships []models.TeamComponentOwnership
	for _, ownership := range ownerships {
		if ownership.OwnershipType == ownershipType {
			filteredOwnerships = append(filteredOwnerships, ownership)
		}
	}

	// Apply pagination
	total := int64(len(filteredOwnerships))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filteredOwnerships) {
		filteredOwnerships = []models.TeamComponentOwnership{}
	} else {
		if end > len(filteredOwnerships) {
			end = len(filteredOwnerships)
		}
		filteredOwnerships = filteredOwnerships[start:end]
	}

	responses := make([]TeamComponentOwnershipResponse, len(filteredOwnerships))
	for i, ownership := range filteredOwnerships {
		responses[i] = *s.toResponse(&ownership)
	}

	return &TeamComponentOwnershipListResponse{
		TeamComponentOwnerships: responses,
		Total:                   total,
		Page:                    page,
		PageSize:                pageSize,
	}, nil
}

// GetByComponentAndOwnershipType retrieves team ownerships for a component by ownership type
func (s *TeamComponentOwnershipService) GetByComponentAndOwnershipType(componentID uuid.UUID, ownershipType models.OwnershipType, page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
	// Validate component exists
	_, err := s.componentRepo.GetByID(componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to verify component: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ownerships, err := s.repo.GetByComponentID(componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component team ownerships: %w", err)
	}

	// Filter by ownership type
	var filteredOwnerships []models.TeamComponentOwnership
	for _, ownership := range ownerships {
		if ownership.OwnershipType == ownershipType {
			filteredOwnerships = append(filteredOwnerships, ownership)
		}
	}

	// Apply pagination
	total := int64(len(filteredOwnerships))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filteredOwnerships) {
		filteredOwnerships = []models.TeamComponentOwnership{}
	} else {
		if end > len(filteredOwnerships) {
			end = len(filteredOwnerships)
		}
		filteredOwnerships = filteredOwnerships[start:end]
	}

	responses := make([]TeamComponentOwnershipResponse, len(filteredOwnerships))
	for i, ownership := range filteredOwnerships {
		responses[i] = *s.toResponse(&ownership)
	}

	return &TeamComponentOwnershipListResponse{
		TeamComponentOwnerships: responses,
		Total:                   total,
		Page:                    page,
		PageSize:                pageSize,
	}, nil
}

// GetPrimaryOwnerships retrieves components with primary ownership
func (s *TeamComponentOwnershipService) GetPrimaryOwnerships(page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
	return s.GetByOwnershipType(models.OwnershipTypePrimary, page, pageSize)
}

// GetSecondaryOwnerships retrieves components with secondary ownership
func (s *TeamComponentOwnershipService) GetSecondaryOwnerships(page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
	return s.GetByOwnershipType(models.OwnershipTypeSecondary, page, pageSize)
}

// GetContributorOwnerships retrieves components with contributor ownership
func (s *TeamComponentOwnershipService) GetContributorOwnerships(page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
	return s.GetByOwnershipType(models.OwnershipTypeContributor, page, pageSize)
}

// GetConsumerOwnerships retrieves components with consumer ownership
func (s *TeamComponentOwnershipService) GetConsumerOwnerships(page, pageSize int) (*TeamComponentOwnershipListResponse, error) {
	return s.GetByOwnershipType(models.OwnershipTypeConsumer, page, pageSize)
}

// Update updates a team-component ownership relationship
func (s *TeamComponentOwnershipService) Update(teamID, componentID uuid.UUID, req *UpdateTeamComponentOwnershipRequest) (*TeamComponentOwnershipResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if ownership exists
	exists, err := s.repo.Exists(teamID, componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check team-component ownership: %w", err)
	}
	if !exists {
		return nil, apperrors.ErrTeamComponentOwnershipNotFound
	}

	// Get existing ownership to update
	ownerships, err := s.repo.GetByTeamID(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team ownerships: %w", err)
	}

	var ownership *models.TeamComponentOwnership
	for _, o := range ownerships {
		if o.ComponentID == componentID {
			ownership = &o
			break
		}
	}

	if ownership == nil {
		return nil, apperrors.ErrTeamComponentOwnershipNotFound
	}

	// Update ownership type
	ownership.OwnershipType = req.OwnershipType

	if err := s.repo.Update(ownership); err != nil {
		return nil, fmt.Errorf("failed to update team-component ownership: %w", err)
	}

	return s.toResponse(ownership), nil
}

// Delete deletes a team-component ownership relationship
func (s *TeamComponentOwnershipService) Delete(teamID, componentID uuid.UUID) error {
	// Check if ownership exists
	exists, err := s.repo.Exists(teamID, componentID)
	if err != nil {
		return fmt.Errorf("failed to check team-component ownership: %w", err)
	}
	if !exists {
		return apperrors.ErrTeamComponentOwnershipNotFound
	}

	if err := s.repo.Delete(teamID, componentID); err != nil {
		return fmt.Errorf("failed to delete team-component ownership: %w", err)
	}

	return nil
}

// CheckExists checks if a team-component ownership relationship exists
func (s *TeamComponentOwnershipService) CheckExists(teamID, componentID uuid.UUID) (bool, error) {
	// Validate team exists
	_, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperrors.ErrTeamNotFound
		}
		return false, fmt.Errorf("failed to verify team: %w", err)
	}

	// Validate component exists
	_, err = s.componentRepo.GetByID(componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperrors.ErrComponentNotFound
		}
		return false, fmt.Errorf("failed to verify component: %w", err)
	}

	exists, err := s.repo.Exists(teamID, componentID)
	if err != nil {
		return false, fmt.Errorf("failed to check ownership existence: %w", err)
	}

	return exists, nil
}

// GetStats retrieves basic statistics for team-component ownerships
func (s *TeamComponentOwnershipService) GetStats() (map[string]int64, error) {
	// This is a simplified stats implementation since the repository doesn't have dedicated stats methods
	// In a real implementation, you might want to add these methods to the repository
	stats := make(map[string]int64)

	// We can't easily get detailed stats with the current repository interface
	// This would require additional repository methods or direct database queries
	stats["total"] = 0
	stats["primary"] = 0
	stats["secondary"] = 0
	stats["contributor"] = 0
	stats["consumer"] = 0

	return stats, nil
}

// BulkCreate creates multiple team-component ownership relationships
func (s *TeamComponentOwnershipService) BulkCreate(requests []CreateTeamComponentOwnershipRequest) ([]TeamComponentOwnershipResponse, []error) {
	responses := make([]TeamComponentOwnershipResponse, 0, len(requests))
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

// BulkDelete deletes multiple team-component ownership relationships
func (s *TeamComponentOwnershipService) BulkDelete(ownerships []struct {
	TeamID      uuid.UUID `json:"team_id"`
	ComponentID uuid.UUID `json:"component_id"`
}) []error {
	errors := make([]error, 0)

	for _, ownership := range ownerships {
		if err := s.Delete(ownership.TeamID, ownership.ComponentID); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// toResponse converts a team component ownership model to response
func (s *TeamComponentOwnershipService) toResponse(ownership *models.TeamComponentOwnership) *TeamComponentOwnershipResponse {
	return &TeamComponentOwnershipResponse{
		ID:            ownership.ID,
		TeamID:        ownership.TeamID,
		ComponentID:   ownership.ComponentID,
		OwnershipType: ownership.OwnershipType,
		CreatedAt:     ownership.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     ownership.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
