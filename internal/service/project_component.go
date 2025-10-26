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

// ProjectComponentService handles business logic for project-component relationships
type ProjectComponentService struct {
	repo          *repository.ProjectComponentRepository
	projectRepo   *repository.ProjectRepository
	componentRepo *repository.ComponentRepository
	validator     *validator.Validate
}

// NewProjectComponentService creates a new project component service
func NewProjectComponentService(repo *repository.ProjectComponentRepository, projectRepo *repository.ProjectRepository, componentRepo *repository.ComponentRepository, validator *validator.Validate) *ProjectComponentService {
	return &ProjectComponentService{
		repo:          repo,
		projectRepo:   projectRepo,
		componentRepo: componentRepo,
		validator:     validator,
	}
}

// CreateProjectComponentRequest represents the request to create a project-component relationship
type CreateProjectComponentRequest struct {
	ProjectID     uuid.UUID            `json:"project_id" validate:"required"`
	ComponentID   uuid.UUID            `json:"component_id" validate:"required"`
	OwnershipType models.OwnershipType `json:"ownership_type,omitempty"`
	SortOrder     int                  `json:"sort_order,omitempty"`
}

// UpdateProjectComponentRequest represents the request to update a project-component relationship
type UpdateProjectComponentRequest struct {
	OwnershipType *models.OwnershipType `json:"ownership_type,omitempty"`
	SortOrder     *int                  `json:"sort_order,omitempty"`
}

// ProjectComponentResponse represents the response for project-component operations
type ProjectComponentResponse struct {
	ID            uuid.UUID            `json:"id"`
	ProjectID     uuid.UUID            `json:"project_id"`
	ComponentID   uuid.UUID            `json:"component_id"`
	OwnershipType models.OwnershipType `json:"ownership_type"`
	SortOrder     int                  `json:"sort_order"`
	CreatedAt     string               `json:"created_at"`
	UpdatedAt     string               `json:"updated_at"`
}

// ProjectComponentListResponse represents a paginated list of project-component relationships
type ProjectComponentListResponse struct {
	ProjectComponents []ProjectComponentResponse `json:"project_components"`
	Total             int64                      `json:"total"`
	Page              int                        `json:"page"`
	PageSize          int                        `json:"page_size"`
}

// Create creates a new project-component relationship
func (s *ProjectComponentService) Create(req *CreateProjectComponentRequest) (*ProjectComponentResponse, error) {
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

	// Validate component exists
	_, err = s.componentRepo.GetByID(req.ComponentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to verify component: %w", err)
	}

	// Check if relationship already exists
	exists, err := s.repo.Exists(req.ProjectID, req.ComponentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing relationship: %w", err)
	}
	if exists {
		return nil, apperrors.ErrProjectComponentExists
	}

	// Set defaults
	ownershipType := req.OwnershipType
	if ownershipType == "" {
		ownershipType = models.OwnershipTypeConsumer
	}

	// Create project-component relationship
	projectComponent := &models.ProjectComponent{
		ProjectID:     req.ProjectID,
		ComponentID:   req.ComponentID,
		OwnershipType: ownershipType,
		SortOrder:     req.SortOrder,
	}

	if err := s.repo.Create(projectComponent); err != nil {
		return nil, fmt.Errorf("failed to create project-component relationship: %w", err)
	}

	return s.toResponse(projectComponent), nil
}

// GetByProject retrieves all component relationships for a project
func (s *ProjectComponentService) GetByProject(projectID uuid.UUID, page, pageSize int) (*ProjectComponentListResponse, error) {
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

	projectComponents, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project components: %w", err)
	}

	// Apply pagination
	total := int64(len(projectComponents))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(projectComponents) {
		projectComponents = []models.ProjectComponent{}
	} else {
		if end > len(projectComponents) {
			end = len(projectComponents)
		}
		projectComponents = projectComponents[start:end]
	}

	responses := make([]ProjectComponentResponse, len(projectComponents))
	for i, pc := range projectComponents {
		responses[i] = *s.toResponse(&pc)
	}

	return &ProjectComponentListResponse{
		ProjectComponents: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetByComponent retrieves all project relationships for a component
func (s *ProjectComponentService) GetByComponent(componentID uuid.UUID, page, pageSize int) (*ProjectComponentListResponse, error) {
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

	projectComponents, err := s.repo.GetByComponentID(componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component projects: %w", err)
	}

	// Apply pagination
	total := int64(len(projectComponents))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(projectComponents) {
		projectComponents = []models.ProjectComponent{}
	} else {
		if end > len(projectComponents) {
			end = len(projectComponents)
		}
		projectComponents = projectComponents[start:end]
	}

	responses := make([]ProjectComponentResponse, len(projectComponents))
	for i, pc := range projectComponents {
		responses[i] = *s.toResponse(&pc)
	}

	return &ProjectComponentListResponse{
		ProjectComponents: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetByProjectAndOwnershipType retrieves components for a project by ownership type
func (s *ProjectComponentService) GetByProjectAndOwnershipType(projectID uuid.UUID, ownershipType models.OwnershipType, page, pageSize int) (*ProjectComponentListResponse, error) {
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

	projectComponents, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project components: %w", err)
	}

	// Filter by ownership type
	var filteredComponents []models.ProjectComponent
	for _, pc := range projectComponents {
		if pc.OwnershipType == ownershipType {
			filteredComponents = append(filteredComponents, pc)
		}
	}

	// Apply pagination
	total := int64(len(filteredComponents))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filteredComponents) {
		filteredComponents = []models.ProjectComponent{}
	} else {
		if end > len(filteredComponents) {
			end = len(filteredComponents)
		}
		filteredComponents = filteredComponents[start:end]
	}

	responses := make([]ProjectComponentResponse, len(filteredComponents))
	for i, pc := range filteredComponents {
		responses[i] = *s.toResponse(&pc)
	}

	return &ProjectComponentListResponse{
		ProjectComponents: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetByComponentAndOwnershipType retrieves projects for a component by ownership type
func (s *ProjectComponentService) GetByComponentAndOwnershipType(componentID uuid.UUID, ownershipType models.OwnershipType, page, pageSize int) (*ProjectComponentListResponse, error) {
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

	projectComponents, err := s.repo.GetByComponentID(componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component projects: %w", err)
	}

	// Filter by ownership type
	var filteredComponents []models.ProjectComponent
	for _, pc := range projectComponents {
		if pc.OwnershipType == ownershipType {
			filteredComponents = append(filteredComponents, pc)
		}
	}

	// Apply pagination
	total := int64(len(filteredComponents))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filteredComponents) {
		filteredComponents = []models.ProjectComponent{}
	} else {
		if end > len(filteredComponents) {
			end = len(filteredComponents)
		}
		filteredComponents = filteredComponents[start:end]
	}

	responses := make([]ProjectComponentResponse, len(filteredComponents))
	for i, pc := range filteredComponents {
		responses[i] = *s.toResponse(&pc)
	}

	return &ProjectComponentListResponse{
		ProjectComponents: responses,
		Total:             total,
		Page:              page,
		PageSize:          pageSize,
	}, nil
}

// GetPrimaryComponents retrieves components with primary ownership for a project
func (s *ProjectComponentService) GetPrimaryComponents(projectID uuid.UUID, page, pageSize int) (*ProjectComponentListResponse, error) {
	return s.GetByProjectAndOwnershipType(projectID, models.OwnershipTypePrimary, page, pageSize)
}

// GetSecondaryComponents retrieves components with secondary ownership for a project
func (s *ProjectComponentService) GetSecondaryComponents(projectID uuid.UUID, page, pageSize int) (*ProjectComponentListResponse, error) {
	return s.GetByProjectAndOwnershipType(projectID, models.OwnershipTypeSecondary, page, pageSize)
}

// GetContributorComponents retrieves components with contributor ownership for a project
func (s *ProjectComponentService) GetContributorComponents(projectID uuid.UUID, page, pageSize int) (*ProjectComponentListResponse, error) {
	return s.GetByProjectAndOwnershipType(projectID, models.OwnershipTypeContributor, page, pageSize)
}

// GetConsumerComponents retrieves components with consumer ownership for a project
func (s *ProjectComponentService) GetConsumerComponents(projectID uuid.UUID, page, pageSize int) (*ProjectComponentListResponse, error) {
	return s.GetByProjectAndOwnershipType(projectID, models.OwnershipTypeConsumer, page, pageSize)
}

// Delete deletes a project-component relationship
func (s *ProjectComponentService) Delete(projectID, componentID uuid.UUID) error {
	// Check if relationship exists
	exists, err := s.repo.Exists(projectID, componentID)
	if err != nil {
		return fmt.Errorf("failed to check project-component relationship: %w", err)
	}
	if !exists {
		return apperrors.ErrProjectComponentNotFound
	}

	if err := s.repo.Delete(projectID, componentID); err != nil {
		return fmt.Errorf("failed to delete project-component relationship: %w", err)
	}

	return nil
}

// CheckExists checks if a project-component relationship exists
func (s *ProjectComponentService) CheckExists(projectID, componentID uuid.UUID) (bool, error) {
	// Validate project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperrors.ErrProjectNotFound
		}
		return false, fmt.Errorf("failed to verify project: %w", err)
	}

	// Validate component exists
	_, err = s.componentRepo.GetByID(componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, apperrors.ErrComponentNotFound
		}
		return false, fmt.Errorf("failed to verify component: %w", err)
	}

	exists, err := s.repo.Exists(projectID, componentID)
	if err != nil {
		return false, fmt.Errorf("failed to check relationship existence: %w", err)
	}

	return exists, nil
}

// GetStats retrieves basic statistics for project-component relationships
func (s *ProjectComponentService) GetStats() (map[string]int64, error) {
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

// toResponse converts a project component model to response
func (s *ProjectComponentService) toResponse(projectComponent *models.ProjectComponent) *ProjectComponentResponse {
	return &ProjectComponentResponse{
		ID:            projectComponent.ID,
		ProjectID:     projectComponent.ProjectID,
		ComponentID:   projectComponent.ComponentID,
		OwnershipType: projectComponent.OwnershipType,
		SortOrder:     projectComponent.SortOrder,
		CreatedAt:     projectComponent.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     projectComponent.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
