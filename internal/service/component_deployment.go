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

// ComponentDeploymentService handles business logic for component deployments
type ComponentDeploymentService struct {
	repo          *repository.ComponentDeploymentRepository
	componentRepo *repository.ComponentRepository
	landscapeRepo *repository.LandscapeRepository
	validator     *validator.Validate
}

// NewComponentDeploymentService creates a new component deployment service
func NewComponentDeploymentService(repo *repository.ComponentDeploymentRepository, componentRepo *repository.ComponentRepository, landscapeRepo *repository.LandscapeRepository, validator *validator.Validate) *ComponentDeploymentService {
	return &ComponentDeploymentService{
		repo:          repo,
		componentRepo: componentRepo,
		landscapeRepo: landscapeRepo,
		validator:     validator,
	}
}

// CreateComponentDeploymentRequest represents the request to create a component deployment
type CreateComponentDeploymentRequest struct {
	ComponentID     uuid.UUID       `json:"component_id" validate:"required"`
	LandscapeID     uuid.UUID       `json:"landscape_id" validate:"required"`
	Version         string          `json:"version,omitempty"`
	GitCommitID     string          `json:"git_commit_id,omitempty"`
	GitCommitTime   *time.Time      `json:"git_commit_time,omitempty"`
	BuildTime       *time.Time      `json:"build_time,omitempty"`
	BuildProperties json.RawMessage `json:"build_properties,omitempty" swaggertype:"object"`
	GitProperties   json.RawMessage `json:"git_properties,omitempty" swaggertype:"object"`
	IsActive        *bool           `json:"is_active,omitempty"`
	DeployedAt      *time.Time      `json:"deployed_at,omitempty"`
}

// UpdateComponentDeploymentRequest represents the request to update a component deployment
type UpdateComponentDeploymentRequest struct {
	Version         string          `json:"version,omitempty"`
	GitCommitID     string          `json:"git_commit_id,omitempty"`
	GitCommitTime   *time.Time      `json:"git_commit_time,omitempty"`
	BuildTime       *time.Time      `json:"build_time,omitempty"`
	BuildProperties json.RawMessage `json:"build_properties,omitempty" swaggertype:"object"`
	GitProperties   json.RawMessage `json:"git_properties,omitempty" swaggertype:"object"`
	IsActive        *bool           `json:"is_active,omitempty"`
	DeployedAt      *time.Time      `json:"deployed_at,omitempty"`
}

// ComponentDeploymentResponse represents the response for component deployment operations
type ComponentDeploymentResponse struct {
	ID              uuid.UUID       `json:"id"`
	ComponentID     uuid.UUID       `json:"component_id"`
	LandscapeID     uuid.UUID       `json:"landscape_id"`
	Version         string          `json:"version"`
	GitCommitID     string          `json:"git_commit_id"`
	GitCommitTime   *time.Time      `json:"git_commit_time,omitempty"`
	BuildTime       *time.Time      `json:"build_time,omitempty"`
	BuildProperties json.RawMessage `json:"build_properties,omitempty" swaggertype:"object"`
	GitProperties   json.RawMessage `json:"git_properties,omitempty" swaggertype:"object"`
	IsActive        bool            `json:"is_active"`
	DeployedAt      *time.Time      `json:"deployed_at,omitempty"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

// ComponentDeploymentListResponse represents a paginated list of component deployments
type ComponentDeploymentListResponse struct {
	Deployments []ComponentDeploymentResponse `json:"deployments"`
	Total       int64                         `json:"total"`
	Page        int                           `json:"page"`
	PageSize    int                           `json:"page_size"`
}

// Create creates a new component deployment
func (s *ComponentDeploymentService) Create(req *CreateComponentDeploymentRequest) (*ComponentDeploymentResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate component exists
	_, err := s.componentRepo.GetByID(req.ComponentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to verify component: %w", err)
	}

	// Validate landscape exists
	_, err = s.landscapeRepo.GetByID(req.LandscapeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to verify landscape: %w", err)
	}

	// Check if active deployment already exists for this component-landscape combination
	if req.IsActive == nil || *req.IsActive {
		// Check if there's an existing deployment for this component-landscape combination
		exists, err := s.repo.CheckDeploymentExistsByComponentAndLandscape(req.ComponentID, req.LandscapeID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing deployment: %w", err)
		}
		if exists {
			// Get the existing deployment to check if it's active
			existingDeployment, err := s.repo.GetByComponentAndLandscape(req.ComponentID, req.LandscapeID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("failed to get existing deployment: %w", err)
			}
			if existingDeployment != nil && existingDeployment.IsActive {
				return nil, apperrors.ErrActiveComponentDeploymentExists
			}
		}
	}

	// Set defaults
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create component deployment
	deployment := &models.ComponentDeployment{
		ComponentID:     req.ComponentID,
		LandscapeID:     req.LandscapeID,
		Version:         req.Version,
		GitCommitID:     req.GitCommitID,
		GitCommitTime:   req.GitCommitTime,
		BuildTime:       req.BuildTime,
		BuildProperties: req.BuildProperties,
		GitProperties:   req.GitProperties,
		IsActive:        isActive,
		DeployedAt:      req.DeployedAt,
	}

	if err := s.repo.Create(deployment); err != nil {
		return nil, fmt.Errorf("failed to create component deployment: %w", err)
	}

	return s.toResponse(deployment), nil
}

// GetByID retrieves a component deployment by ID
func (s *ComponentDeploymentService) GetByID(id uuid.UUID) (*ComponentDeploymentResponse, error) {
	deployment, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentDeploymentNotFound
		}
		return nil, fmt.Errorf("failed to get component deployment: %w", err)
	}

	return s.toResponse(deployment), nil
}

// GetByComponent retrieves deployments for a component with pagination
func (s *ComponentDeploymentService) GetByComponent(componentID uuid.UUID, page, pageSize int) (*ComponentDeploymentListResponse, error) {
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

	offset := (page - 1) * pageSize
	deployments, total, err := s.repo.GetByComponentID(componentID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments by component: %w", err)
	}

	responses := make([]ComponentDeploymentResponse, len(deployments))
	for i, deployment := range deployments {
		responses[i] = *s.toResponse(&deployment)
	}

	return &ComponentDeploymentListResponse{
		Deployments: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetByLandscape retrieves deployments for a landscape with pagination
func (s *ComponentDeploymentService) GetByLandscape(landscapeID uuid.UUID, page, pageSize int) (*ComponentDeploymentListResponse, error) {
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

	offset := (page - 1) * pageSize
	deployments, total, err := s.repo.GetByLandscapeID(landscapeID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments by landscape: %w", err)
	}

	responses := make([]ComponentDeploymentResponse, len(deployments))
	for i, deployment := range deployments {
		responses[i] = *s.toResponse(&deployment)
	}

	return &ComponentDeploymentListResponse{
		Deployments: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetByComponentAndLandscape retrieves deployments for a specific component-landscape combination
func (s *ComponentDeploymentService) GetByComponentAndLandscape(componentID, landscapeID uuid.UUID, page, pageSize int) (*ComponentDeploymentListResponse, error) {
	// Validate component and landscape exist
	_, err := s.componentRepo.GetByID(componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to verify component: %w", err)
	}

	_, err = s.landscapeRepo.GetByID(landscapeID)
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

	offset := (page - 1) * pageSize
	deployments, total, err := s.repo.GetDeploymentHistory(componentID, landscapeID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments by component and landscape: %w", err)
	}

	responses := make([]ComponentDeploymentResponse, len(deployments))
	for i, deployment := range deployments {
		responses[i] = *s.toResponse(&deployment)
	}

	return &ComponentDeploymentListResponse{
		Deployments: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetActiveByComponent retrieves active deployments for a component
func (s *ComponentDeploymentService) GetActiveByComponent(componentID uuid.UUID, page, pageSize int) (*ComponentDeploymentListResponse, error) {
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

	offset := (page - 1) * pageSize
	deployments, _, err := s.repo.GetByActiveStatus(true, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active deployments by component: %w", err)
	}

	// Filter by component ID
	var filteredDeployments []models.ComponentDeployment
	for _, deployment := range deployments {
		if deployment.ComponentID == componentID {
			filteredDeployments = append(filteredDeployments, deployment)
		}
	}

	// Get component deployment stats for the total count
	stats, err := s.repo.GetComponentDeploymentStats(componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component deployment stats: %w", err)
	}
	componentTotal := stats["active"]

	responses := make([]ComponentDeploymentResponse, len(filteredDeployments))
	for i, deployment := range filteredDeployments {
		responses[i] = *s.toResponse(&deployment)
	}

	return &ComponentDeploymentListResponse{
		Deployments: responses,
		Total:       componentTotal,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetActiveByLandscape retrieves active deployments for a landscape
func (s *ComponentDeploymentService) GetActiveByLandscape(landscapeID uuid.UUID, page, pageSize int) (*ComponentDeploymentListResponse, error) {
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

	offset := (page - 1) * pageSize
	deployments, _, err := s.repo.GetByActiveStatus(true, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active deployments by landscape: %w", err)
	}

	// Filter by landscape ID
	var filteredDeployments []models.ComponentDeployment
	for _, deployment := range deployments {
		if deployment.LandscapeID == landscapeID {
			filteredDeployments = append(filteredDeployments, deployment)
		}
	}

	// Get landscape deployment stats for the total count
	stats, err := s.repo.GetLandscapeDeploymentStats(landscapeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscape deployment stats: %w", err)
	}
	landscapeTotal := stats["active"]

	responses := make([]ComponentDeploymentResponse, len(filteredDeployments))
	for i, deployment := range filteredDeployments {
		responses[i] = *s.toResponse(&deployment)
	}

	return &ComponentDeploymentListResponse{
		Deployments: responses,
		Total:       landscapeTotal,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// GetActiveByComponentAndLandscape retrieves the active deployment for a specific component-landscape combination
func (s *ComponentDeploymentService) GetActiveByComponentAndLandscape(componentID, landscapeID uuid.UUID) (*ComponentDeploymentResponse, error) {
	// Validate component and landscape exist
	_, err := s.componentRepo.GetByID(componentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentNotFound
		}
		return nil, fmt.Errorf("failed to verify component: %w", err)
	}

	_, err = s.landscapeRepo.GetByID(landscapeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to verify landscape: %w", err)
	}

	deployment, err := s.repo.GetByComponentAndLandscape(componentID, landscapeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentDeploymentNotFound
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Check if it's active
	if !deployment.IsActive {
		return nil, apperrors.ErrActiveDeploymentNotFound
	}

	return s.toResponse(deployment), nil
}

// GetByVersion retrieves deployments by version with pagination
func (s *ComponentDeploymentService) GetByVersion(componentID uuid.UUID, version string, page, pageSize int) (*ComponentDeploymentListResponse, error) {
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

	offset := (page - 1) * pageSize
	deployments, total, err := s.repo.GetByVersion(componentID, version, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments by version: %w", err)
	}

	responses := make([]ComponentDeploymentResponse, len(deployments))
	for i, deployment := range deployments {
		responses[i] = *s.toResponse(&deployment)
	}

	return &ComponentDeploymentListResponse{
		Deployments: responses,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

// Update updates a component deployment
func (s *ComponentDeploymentService) Update(id uuid.UUID, req *UpdateComponentDeploymentRequest) (*ComponentDeploymentResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing deployment
	deployment, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentDeploymentNotFound
		}
		return nil, fmt.Errorf("failed to get component deployment: %w", err)
	}

	// Check if trying to activate and there's already an active deployment
	if req.IsActive != nil && *req.IsActive && !deployment.IsActive {
		existingDeployment, err := s.repo.GetByComponentAndLandscape(deployment.ComponentID, deployment.LandscapeID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing deployment: %w", err)
		}
		if existingDeployment != nil && existingDeployment.IsActive && existingDeployment.ID != deployment.ID {
			return nil, apperrors.ErrActiveComponentDeploymentExists
		}
	}

	// Update fields
	if req.Version != "" {
		deployment.Version = req.Version
	}
	if req.GitCommitID != "" {
		deployment.GitCommitID = req.GitCommitID
	}
	if req.GitCommitTime != nil {
		deployment.GitCommitTime = req.GitCommitTime
	}
	if req.BuildTime != nil {
		deployment.BuildTime = req.BuildTime
	}
	if req.BuildProperties != nil {
		deployment.BuildProperties = req.BuildProperties
	}
	if req.GitProperties != nil {
		deployment.GitProperties = req.GitProperties
	}
	if req.IsActive != nil {
		deployment.IsActive = *req.IsActive
	}
	if req.DeployedAt != nil {
		deployment.DeployedAt = req.DeployedAt
	}

	if err := s.repo.Update(deployment); err != nil {
		return nil, fmt.Errorf("failed to update component deployment: %w", err)
	}

	return s.toResponse(deployment), nil
}

// Delete deletes a component deployment
func (s *ComponentDeploymentService) Delete(id uuid.UUID) error {
	// Check if deployment exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrComponentDeploymentNotFound
		}
		return fmt.Errorf("failed to get component deployment: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete component deployment: %w", err)
	}

	return nil
}

// SetActive activates a deployment (and deactivates others for the same component-landscape combination)
func (s *ComponentDeploymentService) SetActive(id uuid.UUID) error {
	// Get deployment
	deployment, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrComponentDeploymentNotFound
		}
		return fmt.Errorf("failed to get component deployment: %w", err)
	}

	// First get the deployment history to deactivate other deployments
	history, _, err := s.repo.GetDeploymentHistory(deployment.ComponentID, deployment.LandscapeID, 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get deployment history: %w", err)
	}

	// Deactivate all other deployments for this component-landscape combination
	for _, dep := range history {
		if dep.ID != id && dep.IsActive {
			if err := s.repo.SetActiveStatus(dep.ID, false); err != nil {
				return fmt.Errorf("failed to deactivate deployment %s: %w", dep.ID, err)
			}
		}
	}

	// Activate this deployment
	if err := s.repo.SetActiveStatus(id, true); err != nil {
		return fmt.Errorf("failed to activate deployment: %w", err)
	}

	return nil
}

// SetInactive deactivates a deployment
func (s *ComponentDeploymentService) SetInactive(id uuid.UUID) error {
	// Check if deployment exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrComponentDeploymentNotFound
		}
		return fmt.Errorf("failed to get component deployment: %w", err)
	}

	if err := s.repo.SetActiveStatus(id, false); err != nil {
		return fmt.Errorf("failed to deactivate deployment: %w", err)
	}

	return nil
}

// GetWithComponent retrieves a deployment with component details
func (s *ComponentDeploymentService) GetWithComponent(id uuid.UUID) (*models.ComponentDeployment, error) {
	deployment, err := s.repo.GetWithComponent(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentDeploymentNotFound
		}
		return nil, fmt.Errorf("failed to get deployment with component: %w", err)
	}

	return deployment, nil
}

// GetWithLandscape retrieves a deployment with landscape details
func (s *ComponentDeploymentService) GetWithLandscape(id uuid.UUID) (*models.ComponentDeployment, error) {
	deployment, err := s.repo.GetWithLandscape(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentDeploymentNotFound
		}
		return nil, fmt.Errorf("failed to get deployment with landscape: %w", err)
	}

	return deployment, nil
}

// GetWithFullDetails retrieves a deployment with all relationships
func (s *ComponentDeploymentService) GetWithFullDetails(id uuid.UUID) (*models.ComponentDeployment, error) {
	deployment, err := s.repo.GetWithFullDetails(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrComponentDeploymentNotFound
		}
		return nil, fmt.Errorf("failed to get deployment with full details: %w", err)
	}

	return deployment, nil
}

// toResponse converts a component deployment model to response
func (s *ComponentDeploymentService) toResponse(deployment *models.ComponentDeployment) *ComponentDeploymentResponse {
	return &ComponentDeploymentResponse{
		ID:              deployment.ID,
		ComponentID:     deployment.ComponentID,
		LandscapeID:     deployment.LandscapeID,
		Version:         deployment.Version,
		GitCommitID:     deployment.GitCommitID,
		GitCommitTime:   deployment.GitCommitTime,
		BuildTime:       deployment.BuildTime,
		BuildProperties: deployment.BuildProperties,
		GitProperties:   deployment.GitProperties,
		IsActive:        deployment.IsActive,
		DeployedAt:      deployment.DeployedAt,
		CreatedAt:       deployment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       deployment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
