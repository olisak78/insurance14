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
	projectRepo      *repository.ProjectRepository
	validator        *validator.Validate
}

// NewLandscapeService creates a new landscape service
func NewLandscapeService(repo *repository.LandscapeRepository, orgRepo *repository.OrganizationRepository, projectRepo *repository.ProjectRepository, validator *validator.Validate) *LandscapeService {
	return &LandscapeService{
		repo:             repo,
		organizationRepo: orgRepo,
		projectRepo:      projectRepo,
		validator:        validator,
	}
}

// CreateLandscapeRequest represents the request to create a landscape (new model)
type CreateLandscapeRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=40"`
	Title       string          `json:"title" validate:"required,min=1,max=100"`
	Description string          `json:"description,omitempty" validate:"max=200"`
	ProjectID   uuid.UUID       `json:"project_id" validate:"required"`
	Domain      string          `json:"domain" validate:"required,max=200"`
	Environment string          `json:"environment" validate:"required,max=20"`
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// UpdateLandscapeRequest represents the request to update a landscape (new model)
type UpdateLandscapeRequest struct {
	Title       string          `json:"title" validate:"required,min=1,max=100"`
	Description string          `json:"description,omitempty" validate:"max=200"`
	ProjectID   *uuid.UUID      `json:"project_id,omitempty"`
	Domain      string          `json:"domain,omitempty" validate:"max=200"`
	Environment string          `json:"environment,omitempty" validate:"max=20"`
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// LandscapeResponse represents the response for landscape operations (new model)
type LandscapeResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	ProjectID   uuid.UUID       `json:"project_id"`
	Domain      string          `json:"domain"`
	Environment string          `json:"environment"`
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt   string          `jsoncreated_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// LandscapeMinimalResponse represents a trimmed landscape projection for list endpoints
type LandscapeMinimalResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Domain      string          `json:"domain"`
	Environment string          `json:"environment"`
	Metadata    json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
}

// LandscapeListResponse represents a paginated list of landscapes
type LandscapeListResponse struct {
	Landscapes []LandscapeResponse `json:"landscapes"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
}

// CreateLandscape creates a new landscape
func (s *LandscapeService) CreateLandscape(req *CreateLandscapeRequest) (*LandscapeResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if landscape with same name exists (global scope in new model)
	existingByName, err := s.repo.GetByName(req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing landscape by name: %w", err)
	}
	if existingByName != nil {
		return nil, apperrors.ErrLandscapeExists
	}

	// Create landscape (new model)
	landscape := &models.Landscape{
		BaseModel: models.BaseModel{
			Name:        req.Name,
			Title:       req.Title,
			Description: req.Description,
			Metadata:    req.Metadata,
		},
		ProjectID:   req.ProjectID,
		Domain:      req.Domain,
		Environment: req.Environment,
	}

	if err := s.repo.Create(landscape); err != nil {
		return nil, fmt.Errorf("failed to create landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// GetLandscapeByID retrieves a landscape by ID
func (s *LandscapeService) GetLandscapeByID(id uuid.UUID) (*LandscapeResponse, error) {
	landscape, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// GetByName retrieves a landscape by name (organization scope not applicable in new model)
func (s *LandscapeService) GetByName(_ uuid.UUID, name string) (*LandscapeResponse, error) {
	landscape, err := s.repo.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrLandscapeNotFound
		}
		return nil, fmt.Errorf("failed to get landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// GetLandscapesByOrganization retrieves landscapes for an organization with pagination
// Note: organization scope is not present in the new model; returns all landscapes paginated.
func (s *LandscapeService) GetLandscapesByOrganization(_ uuid.UUID, limit, offset int) ([]LandscapeResponse, int64, error) {
	// Note: OrganizationID is no longer a direct relationship, but keeping interface for backward compatibility
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Pass through nil orgID for compatibility; repository ignores it internally.
	landscapes, total, err := s.repo.GetByOrganizationID(uuid.Nil, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get landscapes: %w", err)
	}

	responses := make([]LandscapeResponse, len(landscapes))
	for i, landscape := range landscapes {
		responses[i] = *s.toResponse(&landscape)
	}

	return responses, total, nil
}

// GetByType retrieves landscapes by environment within an organization (org ignored)
func (s *LandscapeService) GetByType(_ uuid.UUID, environment string, page, pageSize int) (*LandscapeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	landscapes, total, err := s.repo.GetByType(uuid.Nil, environment, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes by environment: %w", err)
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

// GetByStatus retrieves landscapes by status within an organization (status not present; returns all)
func (s *LandscapeService) GetByStatus(_ uuid.UUID, status string, page, pageSize int) (*LandscapeListResponse, error) {
	_ = status // status is not a column in the new model
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	landscapes, total, err := s.repo.GetByStatus("", pageSize, offset)
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

// GetActiveLandscapes retrieves "active" landscapes (no status column; returns all)
func (s *LandscapeService) GetActiveLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByStatus(organizationID, "active", page, pageSize)
}

// GetByTypeAndStatus retrieves landscapes by environment and status (status ignored)
func (s *LandscapeService) GetByTypeAndStatus(_ uuid.UUID, environment string, status string, page, pageSize int) (*LandscapeListResponse, error) {
	_ = status
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	landscapes, total, err := s.repo.GetLandscapesByTypeAndStatus(uuid.Nil, environment, status, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes by environment and status: %w", err)
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

// GetProductionLandscapes retrieves production landscapes
func (s *LandscapeService) GetProductionLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByType(organizationID, "production", page, pageSize)
}

// GetDevelopmentLandscapes retrieves development landscapes
func (s *LandscapeService) GetDevelopmentLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByType(organizationID, "development", page, pageSize)
}

// GetStagingLandscapes retrieves staging landscapes
func (s *LandscapeService) GetStagingLandscapes(organizationID uuid.UUID, page, pageSize int) (*LandscapeListResponse, error) {
	return s.GetByType(organizationID, "staging", page, pageSize)
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

// GetByProjectName resolves a project by name and returns all its landscapes (unpaginated cap)
func (s *LandscapeService) GetByProjectName(projectName string) (*LandscapeListResponse, error) {
	if projectName == "" {
		return &LandscapeListResponse{
			Landscapes: []LandscapeResponse{},
			Total:      0,
			Page:       1,
			PageSize:   0,
		}, nil
	}
	project, err := s.projectRepo.GetByName(projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to resolve project by name: %w", err)
	}
	if project == nil {
		return nil, apperrors.ErrProjectNotFound
	}
	// Return all landscapes for this project (using large page size)
	return s.GetByProject(project.ID, 1, 1000)
}

// GetByProjectNameAll returns all landscapes for a project name without pagination and with minimal fields
func (s *LandscapeService) GetByProjectNameAll(projectName string) ([]LandscapeMinimalResponse, error) {
	if projectName == "" {
		return []LandscapeMinimalResponse{}, nil
	}
	project, err := s.projectRepo.GetByName(projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to resolve project by name: %w", err)
	}
	if project == nil {
		return nil, apperrors.ErrProjectNotFound
	}
	landscapes, _, err := s.repo.GetLandscapesByProjectID(project.ID, 1000000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get landscapes by project: %w", err)
	}
	responses := make([]LandscapeMinimalResponse, len(landscapes))
	for i, l := range landscapes {
		responses[i] = LandscapeMinimalResponse{
			ID:          l.ID,
			Name:        l.Name,
			Title:       l.Title,
			Description: l.Description,
			Domain:      l.Domain,
			Environment: l.Environment,
			Metadata:    l.Metadata,
		}
	}
	return responses, nil
}

// ListByQuery searches landscapes with filters
func (s *LandscapeService) ListByQuery(q string, domains []string, environments []string, limit int, offset int) (*LandscapeListResponse, error) {
	// Convert limit/offset to page/pageSize
	page := (offset / limit) + 1
	pageSize := limit

	// For now, use basic search (filters can be enhanced later)
	return s.Search(uuid.Nil, q, page, pageSize)
}

// Search searches landscapes by name, title, or description (org ignored)
func (s *LandscapeService) Search(_ uuid.UUID, query string, page, pageSize int) (*LandscapeListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	landscapes, total, err := s.repo.Search(uuid.Nil, query, pageSize, offset)
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

// UpdateLandscape updates a landscape
func (s *LandscapeService) UpdateLandscape(id uuid.UUID, req *UpdateLandscapeRequest) (*LandscapeResponse, error) {
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

	// Update fields aligned with new model
	landscape.Title = req.Title
	landscape.Description = req.Description
	if req.ProjectID != nil {
		landscape.ProjectID = *req.ProjectID
	}
	if req.Domain != "" {
		landscape.Domain = req.Domain
	}
	if req.Environment != "" {
		landscape.Environment = req.Environment
	}
	if req.Metadata != nil {
		landscape.Metadata = req.Metadata
	}

	if err := s.repo.Update(landscape); err != nil {
		return nil, fmt.Errorf("failed to update landscape: %w", err)
	}

	return s.toResponse(landscape), nil
}

// DeleteLandscape deletes a landscape
func (s *LandscapeService) DeleteLandscape(id uuid.UUID) error {
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

// SetStatus sets the status of a landscape (no-op in new model; kept for API compatibility)
func (s *LandscapeService) SetStatus(id uuid.UUID, status string) error {
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

// GetWithOrganization retrieves a landscape with organization details (no org relation in new model)
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

// GetWithProjects retrieves a landscape with its projects (no relations in new model; returns entity)
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
func (s *LandscapeService) GetComponentDeployments(id uuid.UUID) (*models.Landscape, error) {
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

// toResponse converts a landscape model to response (new model)
func (s *LandscapeService) toResponse(landscape *models.Landscape) *LandscapeResponse {
	return &LandscapeResponse{
		ID:          landscape.ID,
		Name:        landscape.Name,
		Title:       landscape.Title,
		Description: landscape.Description,
		ProjectID:   landscape.ProjectID,
		Domain:      landscape.Domain,
		Environment: landscape.Environment,
		Metadata:    landscape.Metadata,
		CreatedAt:   landscape.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   landscape.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
