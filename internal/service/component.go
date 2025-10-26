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

// ComponentService handles business logic for components
type ComponentService struct {
	repo             *repository.ComponentRepository
	organizationRepo *repository.OrganizationRepository
	validator        *validator.Validate
}

// NewComponentService creates a new component service
func NewComponentService(repo *repository.ComponentRepository, orgRepo *repository.OrganizationRepository, validator *validator.Validate) *ComponentService {
	return &ComponentService{
		repo:             repo,
		organizationRepo: orgRepo,
		validator:        validator,
	}
}

// CreateComponentRequest represents the request to create a component
type CreateComponentRequest struct {
	OrganizationID   uuid.UUID              `json:"organization_id" validate:"required"`
	Name             string                 `json:"name" validate:"required,min=1,max=200"`
	DisplayName      string                 `json:"display_name" validate:"required,max=250"`
	Description      string                 `json:"description,omitempty"`
	ComponentType    models.ComponentType   `json:"component_type,omitempty"`
	Status           models.ComponentStatus `json:"status,omitempty"`
	GroupName        string                 `json:"group_name,omitempty"`
	ArtifactName     string                 `json:"artifact_name,omitempty"`
	GitRepositoryURL string                 `json:"git_repository_url,omitempty"`
	DocumentationURL string                 `json:"documentation_url,omitempty"`
	Links            json.RawMessage        `json:"links,omitempty" swaggertype:"object"`
	Metadata         json.RawMessage        `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateComponentRequest represents the request to update a component
type UpdateComponentRequest struct {
	DisplayName      string                  `json:"display_name" validate:"required,max=250"`
	Description      string                  `json:"description,omitempty"`
	ComponentType    *models.ComponentType   `json:"component_type,omitempty"`
	Status           *models.ComponentStatus `json:"status,omitempty"`
	GroupName        string                  `json:"group_name,omitempty"`
	ArtifactName     string                  `json:"artifact_name,omitempty"`
	GitRepositoryURL string                  `json:"git_repository_url,omitempty"`
	DocumentationURL string                  `json:"documentation_url,omitempty"`
	Links            json.RawMessage         `json:"links,omitempty" swaggertype:"object"`
	Metadata         json.RawMessage         `json:"metadata,omitempty" swaggertype:"object"`
}

// ComponentResponse represents the response for component operations
type ComponentResponse struct {
	ID               uuid.UUID              `json:"id"`
	OrganizationID   uuid.UUID              `json:"organization_id"`
	Name             string                 `json:"name"`
	DisplayName      string                 `json:"display_name"`
	Description      string                 `json:"description"`
	ComponentType    models.ComponentType   `json:"component_type"`
	Status           models.ComponentStatus `json:"status"`
	GroupName        string                 `json:"group_name,omitempty"`
	ArtifactName     string                 `json:"artifact_name,omitempty"`
	GitRepositoryURL string                 `json:"git_repository_url,omitempty"`
	DocumentationURL string                 `json:"documentation_url,omitempty"`
	Links            json.RawMessage        `json:"links,omitempty" swaggertype:"object"`
	Metadata         json.RawMessage        `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

// ComponentListResponse represents a paginated list of components
type ComponentListResponse struct {
	Components []ComponentResponse `json:"components"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// Create creates a new component
func (s *ComponentService) Create(req *CreateComponentRequest) (*ComponentResponse, error) {
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

	// Check if component with same name exists in organization
	existingByName, err := s.repo.GetByName(req.OrganizationID, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing component by name: %w", err)
	}
	if existingByName != nil {
		return nil, apperrors.ErrComponentExists
	}

	// Set defaults
	componentType := req.ComponentType
	if componentType == "" {
		componentType = models.ComponentTypeService
	}

	status := req.Status
	if status == "" {
		status = models.ComponentStatusActive
	}

	// Create component
	component := &models.Component{
		OrganizationID:   req.OrganizationID,
		Name:             req.Name,
		DisplayName:      req.DisplayName,
		Description:      req.Description,
		ComponentType:    componentType,
		Status:           status,
		GroupName:        req.GroupName,
		ArtifactName:     req.ArtifactName,
		GitRepositoryURL: req.GitRepositoryURL,
		DocumentationURL: req.DocumentationURL,
		Links:            req.Links,
		Metadata:         req.Metadata,
	}

	if err := s.repo.Create(component); err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	return s.toResponse(component), nil
}

// GetByID retrieves a component by ID
func (s *ComponentService) GetByID(id uuid.UUID) (*ComponentResponse, error) {
	component, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	return s.toResponse(component), nil
}

// GetByName retrieves a component by name within an organization
func (s *ComponentService) GetByName(organizationID uuid.UUID, name string) (*ComponentResponse, error) {
	component, err := s.repo.GetByName(organizationID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	return s.toResponse(component), nil
}

// GetByOrganization retrieves components for an organization with pagination
func (s *ComponentService) GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*ComponentListResponse, error) {
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
	components, total, err := s.repo.GetByOrganizationID(organizationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get components: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetByType retrieves components by type within an organization
func (s *ComponentService) GetByType(organizationID uuid.UUID, componentType models.ComponentType, page, pageSize int) (*ComponentListResponse, error) {
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
	components, total, err := s.repo.GetByType(organizationID, componentType, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get components by type: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetByStatus retrieves components by status within an organization
func (s *ComponentService) GetByStatus(organizationID uuid.UUID, status models.ComponentStatus, page, pageSize int) (*ComponentListResponse, error) {
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
	components, total, err := s.repo.GetByStatus(organizationID, status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get components by status: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetActiveComponents retrieves active components for an organization
func (s *ComponentService) GetActiveComponents(organizationID uuid.UUID, page, pageSize int) (*ComponentListResponse, error) {
	return s.GetByStatus(organizationID, models.ComponentStatusActive, page, pageSize)
}

// GetByTypeAndStatus retrieves components by type and status within an organization
func (s *ComponentService) GetByTypeAndStatus(organizationID uuid.UUID, componentType models.ComponentType, status models.ComponentStatus, page, pageSize int) (*ComponentListResponse, error) {
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
	components, total, err := s.repo.GetComponentsByTypeAndStatus(organizationID, componentType, status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get components by type and status: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// Search searches components by name or description within an organization
func (s *ComponentService) Search(organizationID uuid.UUID, query string, page, pageSize int) (*ComponentListResponse, error) {
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
	components, total, err := s.repo.Search(organizationID, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search components: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// SearchByMetadata searches components by metadata within an organization
func (s *ComponentService) SearchByMetadata(organizationID uuid.UUID, metadata string, page, pageSize int) (*ComponentListResponse, error) {
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
	components, total, err := s.repo.GetComponentsByMetadata(organizationID, metadata, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search components by metadata: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetByTeam retrieves components owned by a specific team
func (s *ComponentService) GetByTeam(teamID uuid.UUID, page, pageSize int) (*ComponentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	components, total, err := s.repo.GetComponentsByTeamID(teamID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get components by team: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetByProject retrieves components used by a specific project
func (s *ComponentService) GetByProject(projectID uuid.UUID, page, pageSize int) (*ComponentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	components, total, err := s.repo.GetComponentsByProjectID(projectID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get components by project: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetUnowned retrieves components that have no team ownership
func (s *ComponentService) GetUnowned(organizationID uuid.UUID, page, pageSize int) (*ComponentListResponse, error) {
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
	components, total, err := s.repo.GetUnownedComponents(organizationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get unowned components: %w", err)
	}

	responses := make([]ComponentResponse, len(components))
	for i, component := range components {
		responses[i] = *s.toResponse(&component)
	}

	return &ComponentListResponse{
		Components: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// Update updates a component
func (s *ComponentService) Update(id uuid.UUID, req *UpdateComponentRequest) (*ComponentResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing component
	component, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	// Update fields
	component.DisplayName = req.DisplayName
	component.Description = req.Description
	component.GroupName = req.GroupName
	component.ArtifactName = req.ArtifactName
	component.GitRepositoryURL = req.GitRepositoryURL
	component.DocumentationURL = req.DocumentationURL
	if req.ComponentType != nil {
		component.ComponentType = *req.ComponentType
	}
	if req.Status != nil {
		component.Status = *req.Status
	}
	if req.Links != nil {
		component.Links = req.Links
	}
	if req.Metadata != nil {
		component.Metadata = req.Metadata
	}

	if err := s.repo.Update(component); err != nil {
		return nil, fmt.Errorf("failed to update component: %w", err)
	}

	return s.toResponse(component), nil
}

// Delete deletes a component
func (s *ComponentService) Delete(id uuid.UUID) error {
	// Check if component exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrComponentNotFound
		}
		return fmt.Errorf("failed to get component: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete component: %w", err)
	}

	return nil
}

// SetStatus sets the status of a component
func (s *ComponentService) SetStatus(id uuid.UUID, status models.ComponentStatus) error {
	// Check if component exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrComponentNotFound
		}
		return fmt.Errorf("failed to get component: %w", err)
	}

	if err := s.repo.SetStatus(id, status); err != nil {
		return fmt.Errorf("failed to set component status: %w", err)
	}

	return nil
}

// GetWithOrganization retrieves a component with organization details
func (s *ComponentService) GetWithOrganization(id uuid.UUID) (*models.Component, error) {
	component, err := s.repo.GetWithOrganization(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component with organization: %w", err)
	}

	return component, nil
}

// GetWithProjects retrieves a component with its projects
func (s *ComponentService) GetWithProjects(id uuid.UUID) (*models.Component, error) {
	component, err := s.repo.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component with projects: %w", err)
	}

	return component, nil
}

// GetWithDeployments retrieves a component with its deployments
func (s *ComponentService) GetWithDeployments(id uuid.UUID) (*models.Component, error) {
	component, err := s.repo.GetWithDeployments(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component with deployments: %w", err)
	}

	return component, nil
}

// GetWithTeamOwnerships retrieves a component with team ownerships
func (s *ComponentService) GetWithTeamOwnerships(id uuid.UUID) (*models.Component, error) {
	component, err := s.repo.GetWithTeamOwnerships(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component with team ownerships: %w", err)
	}

	return component, nil
}

// GetWithFullDetails retrieves a component with all relationships
func (s *ComponentService) GetWithFullDetails(id uuid.UUID) (*models.Component, error) {
	component, err := s.repo.GetWithFullDetails(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to get component with full details: %w", err)
	}

	return component, nil
}

// toResponse converts a component model to response
func (s *ComponentService) toResponse(component *models.Component) *ComponentResponse {
	return &ComponentResponse{
		ID:               component.ID,
		OrganizationID:   component.OrganizationID,
		Name:             component.Name,
		DisplayName:      component.DisplayName,
		Description:      component.Description,
		ComponentType:    component.ComponentType,
		Status:           component.Status,
		GroupName:        component.GroupName,
		ArtifactName:     component.ArtifactName,
		GitRepositoryURL: component.GitRepositoryURL,
		DocumentationURL: component.DocumentationURL,
		Links:            component.Links,
		Metadata:         component.Metadata,
		CreatedAt:        component.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        component.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
