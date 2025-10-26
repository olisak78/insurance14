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

// ProjectLandscapeService handles business logic for project-landscape relationships
type ProjectLandscapeService struct {
	repo          *repository.ProjectLandscapeRepository
	projectRepo   *repository.ProjectRepository
	landscapeRepo *repository.LandscapeRepository
	validator     *validator.Validate
}

// NewProjectLandscapeService creates a new project landscape service
func NewProjectLandscapeService(repo *repository.ProjectLandscapeRepository, projectRepo *repository.ProjectRepository, landscapeRepo *repository.LandscapeRepository, validator *validator.Validate) *ProjectLandscapeService {
	return &ProjectLandscapeService{
		repo:          repo,
		projectRepo:   projectRepo,
		landscapeRepo: landscapeRepo,
		validator:     validator,
	}
}

// CreateProjectLandscapeRequest represents the request to create a project-landscape relationship
type CreateProjectLandscapeRequest struct {
	ProjectID      uuid.UUID `json:"project_id" validate:"required"`
	LandscapeID    uuid.UUID `json:"landscape_id" validate:"required"`
	LandscapeGroup string    `json:"landscape_group,omitempty"`
	SortOrder      int       `json:"sort_order,omitempty"`
}

// UpdateProjectLandscapeRequest represents the request to update a project-landscape relationship
type UpdateProjectLandscapeRequest struct {
	LandscapeGroup *string `json:"landscape_group,omitempty"`
	SortOrder      *int    `json:"sort_order,omitempty"`
}

// ProjectLandscapeResponse represents the response for project-landscape operations
type ProjectLandscapeResponse struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	LandscapeID    uuid.UUID `json:"landscape_id"`
	LandscapeGroup string    `json:"landscape_group"`
	SortOrder      int       `json:"sort_order"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

// ProjectLandscapeListResponse represents a paginated list of project-landscape relationships
type ProjectLandscapeListResponse struct {
	ProjectLandscapes []ProjectLandscapeResponse `json:"project_landscapes"`
	Total             int64                      `json:"total"`
	Page              int                        `json:"page"`
	PageSize          int                        `json:"page_size"`
}

// Create creates a new project-landscape relationship
func (s *ProjectLandscapeService) Create(req *CreateProjectLandscapeRequest) (*ProjectLandscapeResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate project exists
	_, err := s.projectRepo.GetByID(req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	// Validate landscape exists
	_, err = s.landscapeRepo.GetByID(req.LandscapeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to verify landscape: %w", err)
	}

	// Check if relationship already exists
	exists, err := s.repo.Exists(req.ProjectID, req.LandscapeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing relationship: %w", err)
	}
	if exists {
		return nil, apperrors.ErrProjectLandscapeExists
	}

	// Create project-landscape relationship
	projectLandscape := &models.ProjectLandscape{
		ProjectID:      req.ProjectID,
		LandscapeID:    req.LandscapeID,
		LandscapeGroup: req.LandscapeGroup,
		SortOrder:      req.SortOrder,
	}

	if err := s.repo.Create(projectLandscape); err != nil {
		return nil, fmt.Errorf("failed to create project-landscape relationship: %w", err)
	}

	return s.toResponse(projectLandscape), nil
}

// GetByProject retrieves all landscape relationships for a project
func (s *ProjectLandscapeService) GetByProject(projectID uuid.UUID, page, pageSize int) (*ProjectLandscapeListResponse, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	projectLandscapes, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project landscapes: %w", err)
	}

	// Apply pagination
	total := int64(len(projectLandscapes))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(projectLandscapes) {
		projectLandscapes = []models.ProjectLandscape{}
	} else {
		if end > len(projectLandscapes) {
			end = len(projectLandscapes)
		}
		projectLandscapes = projectLandscapes[start:end]
	}

	responses := make([]ProjectLandscapeResponse, len(projectLandscapes))
	for i, pl := range projectLandscapes {
		responses[i] = *s.toResponse(&pl)
	}

	return &ProjectLandscapeListResponse{
		ProjectLandscapes: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetByLandscape retrieves all project relationships for a landscape
func (s *ProjectLandscapeService) GetByLandscape(landscapeID uuid.UUID, page, pageSize int) (*ProjectLandscapeListResponse, error) {
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

	projectLandscapes, err := s.repo.GetByLandscapeID(landscapeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscape projects: %w", err)
	}

	// Apply pagination
	total := int64(len(projectLandscapes))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(projectLandscapes) {
		projectLandscapes = []models.ProjectLandscape{}
	} else {
		if end > len(projectLandscapes) {
			end = len(projectLandscapes)
		}
		projectLandscapes = projectLandscapes[start:end]
	}

	responses := make([]ProjectLandscapeResponse, len(projectLandscapes))
	for i, pl := range projectLandscapes {
		responses[i] = *s.toResponse(&pl)
	}

	return &ProjectLandscapeListResponse{
		ProjectLandscapes: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetByProjectAndLandscapeGroup retrieves landscapes for a project by landscape group
func (s *ProjectLandscapeService) GetByProjectAndLandscapeGroup(projectID uuid.UUID, landscapeGroup string, page, pageSize int) (*ProjectLandscapeListResponse, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	projectLandscapes, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project landscapes: %w", err)
	}

	// Filter by landscape group
	var filteredLandscapes []models.ProjectLandscape
	for _, pl := range projectLandscapes {
		if pl.LandscapeGroup == landscapeGroup {
			filteredLandscapes = append(filteredLandscapes, pl)
		}
	}

	// Apply pagination
	total := int64(len(filteredLandscapes))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filteredLandscapes) {
		filteredLandscapes = []models.ProjectLandscape{}
	} else {
		if end > len(filteredLandscapes) {
			end = len(filteredLandscapes)
		}
		filteredLandscapes = filteredLandscapes[start:end]
	}

	responses := make([]ProjectLandscapeResponse, len(filteredLandscapes))
	for i, pl := range filteredLandscapes {
		responses[i] = *s.toResponse(&pl)
	}

	return &ProjectLandscapeListResponse{
		ProjectLandscapes: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetByLandscapeAndGroup retrieves projects for a landscape by landscape group
func (s *ProjectLandscapeService) GetByLandscapeAndGroup(landscapeID uuid.UUID, landscapeGroup string, page, pageSize int) (*ProjectLandscapeListResponse, error) {
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

	projectLandscapes, err := s.repo.GetByLandscapeID(landscapeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscape projects: %w", err)
	}

	// Filter by landscape group
	var filteredLandscapes []models.ProjectLandscape
	for _, pl := range projectLandscapes {
		if pl.LandscapeGroup == landscapeGroup {
			filteredLandscapes = append(filteredLandscapes, pl)
		}
	}

	// Apply pagination
	total := int64(len(filteredLandscapes))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filteredLandscapes) {
		filteredLandscapes = []models.ProjectLandscape{}
	} else {
		if end > len(filteredLandscapes) {
			end = len(filteredLandscapes)
		}
		filteredLandscapes = filteredLandscapes[start:end]
	}

	responses := make([]ProjectLandscapeResponse, len(filteredLandscapes))
	for i, pl := range filteredLandscapes {
		responses[i] = *s.toResponse(&pl)
	}

	return &ProjectLandscapeListResponse{
		ProjectLandscapes: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetLandscapeGroups retrieves all unique landscape groups for a project
func (s *ProjectLandscapeService) GetLandscapeGroups(projectID uuid.UUID) ([]string, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to verify project: %w", err)
	}

	projectLandscapes, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project landscapes: %w", err)
	}

	// Extract unique landscape groups
	groupSet := make(map[string]bool)
	for _, pl := range projectLandscapes {
		if pl.LandscapeGroup != "" {
			groupSet[pl.LandscapeGroup] = true
		}
	}

	groups := make([]string, 0, len(groupSet))
	for group := range groupSet {
		groups = append(groups, group)
	}

	return groups, nil
}

// Delete deletes a project-landscape relationship
func (s *ProjectLandscapeService) Delete(projectID, landscapeID uuid.UUID) error {
	// Check if relationship exists
	exists, err := s.repo.Exists(projectID, landscapeID)
	if err != nil {
		return fmt.Errorf("failed to check project-landscape relationship: %w", err)
	}
	if !exists {
		return apperrors.ErrProjectLandscapeNotFound
	}

	if err := s.repo.Delete(projectID, landscapeID); err != nil {
		return fmt.Errorf("failed to delete project-landscape relationship: %w", err)
	}

	return nil
}

// CheckExists checks if a project-landscape relationship exists
func (s *ProjectLandscapeService) CheckExists(projectID, landscapeID uuid.UUID) (bool, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperrors.ErrProjectNotFound
		}
		return false, fmt.Errorf("failed to verify project: %w", err)
	}

	// Validate landscape exists
	_, err = s.landscapeRepo.GetByID(landscapeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperrors.ErrLandscapeNotFound
		}
		return false, fmt.Errorf("failed to verify landscape: %w", err)
	}

	exists, err := s.repo.Exists(projectID, landscapeID)
	if err != nil {
		return false, fmt.Errorf("failed to check relationship existence: %w", err)
	}

	return exists, nil
}

// GetStats retrieves basic statistics for project-landscape relationships
func (s *ProjectLandscapeService) GetStats() (map[string]int64, error) {
	// This is a simplified stats implementation since the repository doesn't have dedicated stats methods
	// In a real implementation, you might want to add these methods to the repository
	stats := make(map[string]int64)

	// We can't easily get detailed stats with the current repository interface
	// This would require additional repository methods or direct database queries
	stats["total"] = 0
	stats["grouped"] = 0
	stats["ungrouped"] = 0

	return stats, nil
}

// BulkCreate creates multiple project-landscape relationships
func (s *ProjectLandscapeService) BulkCreate(requests []CreateProjectLandscapeRequest) ([]ProjectLandscapeResponse, []error) {
	responses := make([]ProjectLandscapeResponse, 0, len(requests))
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

// BulkDelete deletes multiple project-landscape relationships
func (s *ProjectLandscapeService) BulkDelete(relationships []struct {
	ProjectID   uuid.UUID `json:"project_id"`
	LandscapeID uuid.UUID `json:"landscape_id"`
}) []error {
	errors := make([]error, 0)

	for _, rel := range relationships {
		if err := s.Delete(rel.ProjectID, rel.LandscapeID); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// toResponse converts a project landscape model to response
func (s *ProjectLandscapeService) toResponse(projectLandscape *models.ProjectLandscape) *ProjectLandscapeResponse {
	return &ProjectLandscapeResponse{
		ID:             projectLandscape.ID,
		ProjectID:      projectLandscape.ProjectID,
		LandscapeID:    projectLandscape.LandscapeID,
		LandscapeGroup: projectLandscape.LandscapeGroup,
		SortOrder:      projectLandscape.SortOrder,
		CreatedAt:      projectLandscape.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      projectLandscape.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
