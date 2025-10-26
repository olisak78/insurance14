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

// OrganizationService handles business logic for organizations
type OrganizationService struct {
	repo      repository.OrganizationRepositoryInterface
	validator *validator.Validate
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(repo repository.OrganizationRepositoryInterface, validator *validator.Validate) *OrganizationService {
	return &OrganizationService{
		repo:      repo,
		validator: validator,
	}
}

// CreateOrganizationRequest represents the request to create an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	DisplayName string `json:"display_name" validate:"required,max=200"`
	Domain      string `json:"domain" validate:"required,max=100"`
	Description string `json:"description,omitempty"`
}

// UpdateOrganizationRequest represents the request to update an organization
type UpdateOrganizationRequest struct {
	DisplayName string `json:"display_name" validate:"required,max=200"`
	Description string `json:"description,omitempty"`
}

// OrganizationResponse represents the response for organization operations
type OrganizationResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Domain      string    `json:"domain"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// OrganizationListResponse represents a paginated list of organizations
type OrganizationListResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
}

// Create creates a new organization
func (s *OrganizationService) Create(req *CreateOrganizationRequest) (*OrganizationResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if organization with same name exists
	existingByName, err := s.repo.GetByName(req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing organization by name: %w", err)
	}
	if existingByName != nil {
		return nil, apperrors.ErrOrganizationExists
	}

	// Check if organization with same domain exists
	existingByDomain, err := s.repo.GetByDomain(req.Domain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing organization by domain: %w", err)
	}
	if existingByDomain != nil {
		return nil, apperrors.ErrOrganizationExists
	}

	// Create organization
	org := &models.Organization{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Domain:      req.Domain,
		Description: req.Description,
	}

	if err := s.repo.Create(org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return s.toResponse(org), nil
}

// GetByID retrieves an organization by ID
func (s *OrganizationService) GetByID(id uuid.UUID) (*OrganizationResponse, error) {
	org, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return s.toResponse(org), nil
}

// GetByName retrieves an organization by name
func (s *OrganizationService) GetByName(name string) (*OrganizationResponse, error) {
	org, err := s.repo.GetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return s.toResponse(org), nil
}

// GetByDomain retrieves an organization by domain
func (s *OrganizationService) GetByDomain(domain string) (*OrganizationResponse, error) {
	org, err := s.repo.GetByDomain(domain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return s.toResponse(org), nil
}

// GetAll retrieves all organizations with pagination
func (s *OrganizationService) GetAll(page, pageSize int) (*OrganizationListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	orgs, total, err := s.repo.GetAll(pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}

	responses := make([]OrganizationResponse, len(orgs))
	for i, org := range orgs {
		responses[i] = *s.toResponse(&org)
	}

	return &OrganizationListResponse{
		Organizations: responses,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// Update updates an organization
func (s *OrganizationService) Update(id uuid.UUID, req *UpdateOrganizationRequest) (*OrganizationResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing organization
	org, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Update fields
	org.DisplayName = req.DisplayName
	org.Description = req.Description

	if err := s.repo.Update(org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return s.toResponse(org), nil
}

// Delete deletes an organization
func (s *OrganizationService) Delete(id uuid.UUID) error {
	// Check if organization exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrOrganizationNotFound
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return nil
}

// GetWithMembers retrieves an organization with its members
func (s *OrganizationService) GetWithMembers(id uuid.UUID) (*models.Organization, error) {
	org, err := s.repo.GetWithMembers(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization with members: %w", err)
	}

	return org, nil
}

// GetWithGroups retrieves an organization with its groups
func (s *OrganizationService) GetWithGroups(id uuid.UUID) (*models.Organization, error) {
	org, err := s.repo.GetWithGroups(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization with groups: %w", err)
	}

	return org, nil
}

// GetWithProjects retrieves an organization with its projects
func (s *OrganizationService) GetWithProjects(id uuid.UUID) (*models.Organization, error) {
	org, err := s.repo.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization with projects: %w", err)
	}

	return org, nil
}

// GetWithComponents retrieves an organization with its components
func (s *OrganizationService) GetWithComponents(id uuid.UUID) (*models.Organization, error) {
	org, err := s.repo.GetWithComponents(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization with components: %w", err)
	}

	return org, nil
}

// GetWithLandscapes retrieves an organization with its landscapes
func (s *OrganizationService) GetWithLandscapes(id uuid.UUID) (*models.Organization, error) {
	org, err := s.repo.GetWithLandscapes(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization with landscapes: %w", err)
	}

	return org, nil
}

// toResponse converts an organization model to response
func (s *OrganizationService) toResponse(org *models.Organization) *OrganizationResponse {
	return &OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		DisplayName: org.DisplayName,
		Domain:      org.Domain,
		Description: org.Description,
		CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
