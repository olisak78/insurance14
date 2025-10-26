package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"developer-portal-backend/internal/database/models"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectService handles business logic for projects
type ProjectService struct {
	repo             *repository.ProjectRepository
	organizationRepo *repository.OrganizationRepository
	validator        *validator.Validate
}

// NewProjectService creates a new project service
func NewProjectService(repo *repository.ProjectRepository, orgRepo *repository.OrganizationRepository, validator *validator.Validate) *ProjectService {
	return &ProjectService{
		repo:             repo,
		organizationRepo: orgRepo,
		validator:        validator,
	}
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	OrganizationID uuid.UUID            `json:"organization_id" validate:"required"`
	Name           string               `json:"name" validate:"required,min=1,max=200"`
	DisplayName    string               `json:"display_name" validate:"required,max=250"`
	Description    string               `json:"description,omitempty"`
	ProjectType    models.ProjectType   `json:"project_type,omitempty"`
	Status         models.ProjectStatus `json:"status,omitempty"`
	SortOrder      int                  `json:"sort_order,omitempty"`
	Metadata       json.RawMessage      `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	DisplayName string                `json:"display_name" validate:"required,max=250"`
	Description string                `json:"description,omitempty"`
	ProjectType *models.ProjectType   `json:"project_type,omitempty"`
	Status      *models.ProjectStatus `json:"status,omitempty"`
	SortOrder   *int                  `json:"sort_order,omitempty"`
	Metadata    json.RawMessage       `json:"metadata,omitempty" swaggertype:"object"`
}

// ProjectResponse represents the response for project operations
type ProjectResponse struct {
	ID             uuid.UUID            `json:"id"`
	OrganizationID uuid.UUID            `json:"organization_id"`
	Name           string               `json:"name"`
	DisplayName    string               `json:"display_name"`
	Description    string               `json:"description"`
	ProjectType    models.ProjectType   `json:"project_type"`
	Status         models.ProjectStatus `json:"status"`
	SortOrder      int                  `json:"sort_order"`
	Metadata       json.RawMessage      `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt      string               `json:"created_at"`
	UpdatedAt      string               `json:"updated_at"`
}

// ProjectListResponse represents a paginated list of projects
type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// Create creates a new project
func (s *ProjectService) Create(req *CreateProjectRequest) (*ProjectResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate organization exists
	_, err := s.organizationRepo.GetByID(req.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to verify organization: %w", err)
	}

	// Check if project with same name exists in organization
	existingByName, err := s.repo.GetByName(req.OrganizationID, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing project by name: %w", err)
	}
	if existingByName != nil {
		return nil, apperrors.ErrProjectExists
	}

	// Set defaults
	projectType := req.ProjectType
	if projectType == "" {
		projectType = models.ProjectTypeApplication
	}

	status := req.Status
	if status == "" {
		status = models.ProjectStatusActive
	}

	// Create project
	project := &models.Project{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		ProjectType:    projectType,
		Status:         status,
		SortOrder:      req.SortOrder,
		Metadata:       req.Metadata,
	}

	if err := s.repo.Create(project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return s.toResponse(project), nil
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(id uuid.UUID) (*ProjectResponse, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return s.toResponse(project), nil
}

// GetByName retrieves a project by name within an organization
func (s *ProjectService) GetByName(organizationID uuid.UUID, name string) (*ProjectResponse, error) {
	project, err := s.repo.GetByName(organizationID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return s.toResponse(project), nil
}

// GetByOrganization retrieves projects for an organization with pagination
func (s *ProjectService) GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*ProjectListResponse, error) {
	// Validate organization exists
	_, err := s.organizationRepo.GetByID(organizationID)
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
	projects, total, err := s.repo.GetByOrganizationID(organizationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = *s.toResponse(&project)
	}

	return &ProjectListResponse{
		Projects: responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetByStatus retrieves projects by status within an organization
func (s *ProjectService) GetByStatus(organizationID uuid.UUID, status models.ProjectStatus, page, pageSize int) (*ProjectListResponse, error) {
	// Validate organization exists
	_, err := s.organizationRepo.GetByID(organizationID)
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
	projects, total, err := s.repo.GetByStatus(organizationID, status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects by status: %w", err)
	}

	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = *s.toResponse(&project)
	}

	return &ProjectListResponse{
		Projects: responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetActiveProjects retrieves active projects for an organization
func (s *ProjectService) GetActiveProjects(organizationID uuid.UUID, page, pageSize int) (*ProjectListResponse, error) {
	return s.GetByStatus(organizationID, models.ProjectStatusActive, page, pageSize)
}

// Search searches projects by name or description within an organization
func (s *ProjectService) Search(organizationID uuid.UUID, query string, page, pageSize int) (*ProjectListResponse, error) {
	// Validate organization exists
	_, err := s.organizationRepo.GetByID(organizationID)
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
	projects, total, err := s.repo.Search(organizationID, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = *s.toResponse(&project)
	}

	return &ProjectListResponse{
		Projects: responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update updates a project
func (s *ProjectService) Update(id uuid.UUID, req *UpdateProjectRequest) (*ProjectResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing project
	project, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Update fields
	project.DisplayName = req.DisplayName
	project.Description = req.Description
	if req.ProjectType != nil {
		project.ProjectType = *req.ProjectType
	}
	if req.Status != nil {
		project.Status = *req.Status
	}
	if req.SortOrder != nil {
		project.SortOrder = *req.SortOrder
	}
	if req.Metadata != nil {
		project.Metadata = req.Metadata
	}

	if err := s.repo.Update(project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return s.toResponse(project), nil
}

// Delete deletes a project
func (s *ProjectService) Delete(id uuid.UUID) error {
	// Check if project exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// SetStatus sets the status of a project
func (s *ProjectService) SetStatus(id uuid.UUID, status models.ProjectStatus) error {
	// Check if project exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.repo.SetStatus(id, status); err != nil {
		return fmt.Errorf("failed to set project status: %w", err)
	}

	return nil
}

// GetWithOrganization retrieves a project with organization details
func (s *ProjectService) GetWithOrganization(id uuid.UUID) (*models.Project, error) {
	project, err := s.repo.GetWithOrganization(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with organization: %w", err)
	}

	return project, nil
}

// GetWithComponents retrieves a project with its components
func (s *ProjectService) GetWithComponents(id uuid.UUID) (*models.Project, error) {
	project, err := s.repo.GetWithComponents(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with components: %w", err)
	}

	return project, nil
}

// GetWithLandscapes retrieves a project with its landscapes
func (s *ProjectService) GetWithLandscapes(id uuid.UUID) (*models.Project, error) {
	project, err := s.repo.GetWithLandscapes(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with landscapes: %w", err)
	}

	return project, nil
}

// GetWithFullDetails retrieves a project with all relationships
func (s *ProjectService) GetWithFullDetails(id uuid.UUID) (*models.Project, error) {
	project, err := s.repo.GetWithFullDetails(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with full details: %w", err)
	}

	return project, nil
}

// AddComponent adds a component to a project
func (s *ProjectService) AddComponent(projectID, componentID uuid.UUID) error {
	// Check if project exists
	_, err := s.repo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Check if component is already in project
	exists, err := s.repo.CheckComponentInProject(projectID, componentID)
	if err != nil {
		return fmt.Errorf("failed to check component in project: %w", err)
	}
	if exists {
		return apperrors.ErrComponentAlreadyAssociated
	}

	if err := s.repo.AddComponent(projectID, componentID); err != nil {
		return fmt.Errorf("failed to add component to project: %w", err)
	}

	return nil
}

// RemoveComponent removes a component from a project
func (s *ProjectService) RemoveComponent(projectID, componentID uuid.UUID) error {
	// Check if project exists
	_, err := s.repo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Check if component is in project
	exists, err := s.repo.CheckComponentInProject(projectID, componentID)
	if err != nil {
		return fmt.Errorf("failed to check component in project: %w", err)
	}
	if !exists {
		return apperrors.ErrComponentNotAssociated
	}

	if err := s.repo.RemoveComponent(projectID, componentID); err != nil {
		return fmt.Errorf("failed to remove component from project: %w", err)
	}

	return nil
}

// AddLandscape adds a landscape to a project
func (s *ProjectService) AddLandscape(projectID, landscapeID uuid.UUID) error {
	// Check if project exists
	_, err := s.repo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Check if landscape is already in project
	exists, err := s.repo.CheckLandscapeInProject(projectID, landscapeID)
	if err != nil {
		return fmt.Errorf("failed to check landscape in project: %w", err)
	}
	if exists {
		return apperrors.ErrLandscapeAlreadyAssociated
	}

	if err := s.repo.AddLandscape(projectID, landscapeID); err != nil {
		return fmt.Errorf("failed to add landscape to project: %w", err)
	}

	return nil
}

// RemoveLandscape removes a landscape from a project
func (s *ProjectService) RemoveLandscape(projectID, landscapeID uuid.UUID) error {
	// Check if project exists
	_, err := s.repo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Check if landscape is in project
	exists, err := s.repo.CheckLandscapeInProject(projectID, landscapeID)
	if err != nil {
		return fmt.Errorf("failed to check landscape in project: %w", err)
	}
	if !exists {
		return apperrors.ErrLandscapeNotAssociated
	}

	if err := s.repo.RemoveLandscape(projectID, landscapeID); err != nil {
		return fmt.Errorf("failed to remove landscape from project: %w", err)
	}

	return nil
}

// toResponse converts a project model to response
func (s *ProjectService) toResponse(project *models.Project) *ProjectResponse {
	return &ProjectResponse{
		ID:             project.ID,
		OrganizationID: project.OrganizationID,
		Name:           project.Name,
		DisplayName:    project.DisplayName,
		Description:    project.Description,
		ProjectType:    project.ProjectType,
		Status:         project.Status,
		SortOrder:      project.SortOrder,
		Metadata:       project.Metadata,
		CreatedAt:      project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
