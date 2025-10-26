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

// LandscapeService handles business logic for landscapes
type LandscapeService struct {
	repo             *repository.LandscapeRepository
	organizationRepo *repository.OrganizationRepository
	validator        *validator.Validate
}

// NewLandscapeService creates a new landscape service
func NewLandscapeService(repo *repository.LandscapeRepository, orgRepo *repository.OrganizationRepository, validator *validator.Validate) *LandscapeService {
	return &LandscapeService{
		repo:             repo,
		organizationRepo: orgRepo,
		validator:        validator,
	}
}

// CreateLandscapeRequest represents the request to create a landscape
type CreateLandscapeRequest struct {
	OrganizationID   uuid.UUID               `json:"organization_id" validate:"required"`
	Name             string                  `json:"name" validate:"required,min=1,max=200"`
	DisplayName      string                  `json:"display_name" validate:"required,max=250"`
	Description      string                  `json:"description,omitempty"`
	LandscapeType    models.LandscapeType    `json:"landscape_type,omitempty"`
	EnvironmentGroup string                  `json:"environment_group,omitempty"`
	Status           models.LandscapeStatus  `json:"status,omitempty"`
	DeploymentStatus models.DeploymentStatus `json:"deployment_status,omitempty"`
	GitHubConfigURL  string                  `json:"github_config_url,omitempty"`
	AWSAccountID     string                  `json:"aws_account_id,omitempty"`
	CAMProfileURL    string                  `json:"cam_profile_url,omitempty"`
	SortOrder        int                     `json:"sort_order,omitempty"`
	Metadata         json.RawMessage         `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateLandscapeRequest represents the request to update a landscape
type UpdateLandscapeRequest struct {
	DisplayName      string                   `json:"display_name" validate:"required,max=250"`
	Description      string                   `json:"description,omitempty"`
	LandscapeType    *models.LandscapeType    `json:"landscape_type,omitempty"`
	EnvironmentGroup string                   `json:"environment_group,omitempty"`
	Status           *models.LandscapeStatus  `json:"status,omitempty"`
	DeploymentStatus *models.DeploymentStatus `json:"deployment_status,omitempty"`
	GitHubConfigURL  string                   `json:"github_config_url,omitempty"`
	AWSAccountID     string                   `json:"aws_account_id,omitempty"`
	CAMProfileURL    string                   `json:"cam_profile_url,omitempty"`
	SortOrder        *int                     `json:"sort_order,omitempty"`
	Metadata         json.RawMessage          `json:"metadata,omitempty" swaggertype:"object"`
}

// LandscapeResponse represents the response for landscape operations
type LandscapeResponse struct {
	ID               uuid.UUID               `json:"id"`
	OrganizationID   uuid.UUID               `json:"organization_id"`
	Name             string                  `json:"name"`
	DisplayName      string                  `json:"display_name"`
	Description      string                  `json:"description"`
	LandscapeType    models.LandscapeType    `json:"landscape_type"`
	EnvironmentGroup string                  `json:"environment_group,omitempty"`
	Status           models.LandscapeStatus  `json:"status"`
	DeploymentStatus models.DeploymentStatus `json:"deployment_status"`
	GitHubConfigURL  string                  `json:"github_config_url,omitempty"`
	AWSAccountID     string                  `json:"aws_account_id,omitempty"`
	CAMProfileURL    string                  `json:"cam_profile_url,omitempty"`
	SortOrder        int                     `json:"sort_order"`
	Metadata         json.RawMessage         `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt        string                  `json:"created_at"`
	UpdatedAt        string                  `json:"updated_at"`
}

// LandscapeListResponse represents a paginated list of landscapes
type LandscapeListResponse struct {
	Landscapes []LandscapeResponse `json:"landscapes"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// Create creates a new landscape
func (s *LandscapeService) Create(req *CreateLandscapeRequest) (*LandscapeResponse, error) {
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

	// Check if landscape with same name exists in organization
	existingByName, err := s.repo.GetByName(req.OrganizationID, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing landscape by name: %w", err)
	}
	if existingByName != nil {
		return nil, apperrors.ErrLandscapeExists
	}

	// Set defaults
	landscapeType := req.LandscapeType
	if landscapeType == "" {
		landscapeType = models.LandscapeTypeDevelopment
	}

	status := req.Status
	if status == "" {
		status = models.LandscapeStatusActive
	}

	deploymentStatus := req.DeploymentStatus
	if deploymentStatus == "" {
		deploymentStatus = models.DeploymentStatusUnknown
	}

	// Create landscape
	landscape := &models.Landscape{
		OrganizationID:   req.OrganizationID,
		Name:             req.Name,
		DisplayName:      req.DisplayName,
		Description:      req.Description,
		LandscapeType:    landscapeType,
		EnvironmentGroup: req.EnvironmentGroup,
		Status:           status,
		DeploymentStatus: deploymentStatus,
		GitHubConfigURL:  req.GitHubConfigURL,
		AWSAccountID:     req.AWSAccountID,
		CAMProfileURL:    req.CAMProfileURL,
		SortOrder:        req.SortOrder,
		Metadata:         req.Metadata,
	}

	if err := s.repo.Create(landscape); err != nil {
		return nil, fmt.Errorf("failed to create landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// GetByID retrieves a landscape by ID
func (s *LandscapeService) GetByID(id uuid.UUID) (*LandscapeResponse, error) {
	landscape, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// GetByName retrieves a landscape by name within an organization
func (s *LandscapeService) GetByName(organizationID uuid.UUID, name string) (*LandscapeResponse, error) {
	landscape, err := s.repo.GetByName(organizationID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// GetByOrganization retrieves landscapes for an organization with pagination
func (s *LandscapeService) GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
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
	landscapes, total, err := s.repo.GetByOrganizationID(organizationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes: %w", err)
	}

	responses := make([]LandscapeResponse, len(landscapes))
	for i, landscape := range landscapes {
		responses[i] = *s.toResponse(&landscape)
	}

	return &LandscapeListResponse{
		Landscapes: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetByType retrieves landscapes by type within an organization
func (s *LandscapeService) GetByType(organizationID uuid.UUID, landscapeType models.LandscapeType, page, pageSize int) (*LandscapeListResponse, error) {
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
	landscapes, total, err := s.repo.GetByType(organizationID, landscapeType, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes by type: %w", err)
	}

	responses := make([]LandscapeResponse, len(landscapes))
	for i, landscape := range landscapes {
		responses[i] = *s.toResponse(&landscape)
	}

	return &LandscapeListResponse{
		Landscapes: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetByStatus retrieves landscapes by status within an organization
func (s *LandscapeService) GetByStatus(organizationID uuid.UUID, status models.LandscapeStatus, page, pageSize int) (*LandscapeListResponse, error) {
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
	landscapes, total, err := s.repo.GetByStatus(organizationID, status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes by status: %w", err)
	}

	responses := make([]LandscapeResponse, len(landscapes))
	for i, landscape := range landscapes {
		responses[i] = *s.toResponse(&landscape)
	}

	return &LandscapeListResponse{
		Landscapes: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetActiveLandscapes retrieves active landscapes for an organization
func (s *LandscapeService) GetActiveLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByStatus(organizationID, models.LandscapeStatusActive, page, pageSize)
}

// GetByTypeAndStatus retrieves landscapes by type and status within an organization
func (s *LandscapeService) GetByTypeAndStatus(organizationID uuid.UUID, landscapeType models.LandscapeType, status models.LandscapeStatus, page, pageSize int) (*LandscapeListResponse, error) {
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
	landscapes, total, err := s.repo.GetLandscapesByTypeAndStatus(organizationID, landscapeType, status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes by type and status: %w", err)
	}

	responses := make([]LandscapeResponse, len(landscapes))
	for i, landscape := range landscapes {
		responses[i] = *s.toResponse(&landscape)
	}

	return &LandscapeListResponse{
		Landscapes: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// GetProductionLandscapes retrieves production landscapes for an organization
func (s *LandscapeService) GetProductionLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByType(organizationID, models.LandscapeTypeProduction, page, pageSize)
}

// GetDevelopmentLandscapes retrieves development landscapes for an organization
func (s *LandscapeService) GetDevelopmentLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByType(organizationID, models.LandscapeTypeDevelopment, page, pageSize)
}

// GetStagingLandscapes retrieves staging landscapes for an organization
func (s *LandscapeService) GetStagingLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByType(organizationID, models.LandscapeTypeStaging, page, pageSize)
}

// GetByProject retrieves landscapes used by a specific project
func (s *LandscapeService) GetByProject(projectID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	landscapes, total, err := s.repo.GetLandscapesByProjectID(projectID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes by project: %w", err)
	}

	responses := make([]LandscapeResponse, len(landscapes))
	for i, landscape := range landscapes {
		responses[i] = *s.toResponse(&landscape)
	}

	return &LandscapeListResponse{
		Landscapes: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// Search searches landscapes by name or description within an organization
func (s *LandscapeService) Search(organizationID uuid.UUID, query string, page, pageSize int) (*LandscapeListResponse, error) {
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
	landscapes, total, err := s.repo.Search(organizationID, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search landscapes: %w", err)
	}

	responses := make([]LandscapeResponse, len(landscapes))
	for i, landscape := range landscapes {
		responses[i] = *s.toResponse(&landscape)
	}

	return &LandscapeListResponse{
		Landscapes: responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// Update updates a landscape
func (s *LandscapeService) Update(id uuid.UUID, req *UpdateLandscapeRequest) (*LandscapeResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing landscape
	landscape, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape: %w", err)
	}

	// Update fields
	landscape.DisplayName = req.DisplayName
	landscape.Description = req.Description
	landscape.EnvironmentGroup = req.EnvironmentGroup
	landscape.GitHubConfigURL = req.GitHubConfigURL
	landscape.AWSAccountID = req.AWSAccountID
	landscape.CAMProfileURL = req.CAMProfileURL
	if req.LandscapeType != nil {
		landscape.LandscapeType = *req.LandscapeType
	}
	if req.Status != nil {
		landscape.Status = *req.Status
	}
	if req.DeploymentStatus != nil {
		landscape.DeploymentStatus = *req.DeploymentStatus
	}
	if req.SortOrder != nil {
		landscape.SortOrder = *req.SortOrder
	}
	if req.Metadata != nil {
		landscape.Metadata = req.Metadata
	}

	if err := s.repo.Update(landscape); err != nil {
		return nil, fmt.Errorf("failed to update landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// Delete deletes a landscape
func (s *LandscapeService) Delete(id uuid.UUID) error {
	// Check if landscape exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrLandscapeNotFound
		}
		return fmt.Errorf("failed to get landscape: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete landscape: %w", err)
	}

	return nil
}

// SetStatus sets the status of a landscape
func (s *LandscapeService) SetStatus(id uuid.UUID, status models.LandscapeStatus) error {
	// Check if landscape exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrLandscapeNotFound
		}
		return fmt.Errorf("failed to get landscape: %w", err)
	}

	if err := s.repo.SetStatus(id, status); err != nil {
		return fmt.Errorf("failed to set landscape status: %w", err)
	}

	return nil
}

// GetWithOrganization retrieves a landscape with organization details
func (s *LandscapeService) GetWithOrganization(id uuid.UUID) (*models.Landscape, error) {
	landscape, err := s.repo.GetWithOrganization(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape with organization: %w", err)
	}

	return landscape, nil
}

// GetWithProjects retrieves a landscape with its projects
func (s *LandscapeService) GetWithProjects(id uuid.UUID) (*models.Landscape, error) {
	landscape, err := s.repo.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape with projects: %w", err)
	}

	return landscape, nil
}

// GetWithComponentDeployments retrieves a landscape with component deployments
func (s *LandscapeService) GetWithComponentDeployments(id uuid.UUID) (*models.Landscape, error) {
	landscape, err := s.repo.GetWithComponentDeployments(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape with component deployments: %w", err)
	}

	return landscape, nil
}

// GetWithFullDetails retrieves a landscape with all relationships
func (s *LandscapeService) GetWithFullDetails(id uuid.UUID) (*models.Landscape, error) {
	landscape, err := s.repo.GetWithFullDetails(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape with full details: %w", err)
	}

	return landscape, nil
}

// toResponse converts a landscape model to response
func (s *LandscapeService) toResponse(landscape *models.Landscape) *LandscapeResponse {
	return &LandscapeResponse{
		ID:               landscape.ID,
		OrganizationID:   landscape.OrganizationID,
		Name:             landscape.Name,
		DisplayName:      landscape.DisplayName,
		Description:      landscape.Description,
		LandscapeType:    landscape.LandscapeType,
		EnvironmentGroup: landscape.EnvironmentGroup,
		Status:           landscape.Status,
		DeploymentStatus: landscape.DeploymentStatus,
		GitHubConfigURL:  landscape.GitHubConfigURL,
		AWSAccountID:     landscape.AWSAccountID,
		CAMProfileURL:    landscape.CAMProfileURL,
		SortOrder:        landscape.SortOrder,
		Metadata:         landscape.Metadata,
		CreatedAt:        landscape.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        landscape.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
