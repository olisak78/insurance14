package service

import (
	"fmt"
	"net/url"
	"strings"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// DocumentationService provides documentation-related business logic
type DocumentationService struct {
	docRepo   repository.DocumentationRepositoryInterface
	teamRepo  repository.TeamRepositoryInterface
	validator *validator.Validate
}

// Ensure DocumentationService implements DocumentationServiceInterface
var _ DocumentationServiceInterface = (*DocumentationService)(nil)

// NewDocumentationService creates a new DocumentationService
func NewDocumentationService(
	docRepo repository.DocumentationRepositoryInterface,
	teamRepo repository.TeamRepositoryInterface,
	validator *validator.Validate,
) *DocumentationService {
	return &DocumentationService{
		docRepo:   docRepo,
		teamRepo:  teamRepo,
		validator: validator,
	}
}

// DocumentationResponse represents a documentation in API responses
type DocumentationResponse struct {
	ID          string `json:"id"`
	TeamID      string `json:"team_id"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	Branch      string `json:"branch"`
	DocsPath    string `json:"docs_path"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	CreatedBy   string `json:"created_by"`
	UpdatedAt   string `json:"updated_at"`
	UpdatedBy   string `json:"updated_by"`
}

// CreateDocumentationRequest represents the payload for creating a documentation
type CreateDocumentationRequest struct {
	TeamID      string `json:"team_id" validate:"required,uuid4"`
	URL         string `json:"url" validate:"required,url,max=1000"` // Full GitHub URL
	Title       string `json:"title" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=200"`
	CreatedBy   string `json:"-"` // derived from bearer token
}

// UpdateDocumentationRequest represents the payload for updating a documentation
type UpdateDocumentationRequest struct {
	URL         *string `json:"url" validate:"omitempty,url,max=1000"`
	Title       *string `json:"title" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=200"`
	UpdatedBy   string  `json:"-"` // derived from bearer token
}

// CreateDocumentation validates and creates a new documentation
func (s *DocumentationService) CreateDocumentation(req *CreateDocumentationRequest) (*DocumentationResponse, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if strings.TrimSpace(req.CreatedBy) == "" {
		return nil, fmt.Errorf("created_by is required")
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("invalid team_id UUID: %w", err)
	}

	// Validate team exists
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	// Parse and validate GitHub URL
	owner, repo, branch, docsPath, err := parseGitHubURL(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub URL: %w", err)
	}

	doc := &models.Documentation{
		TeamID:      team.ID,
		Owner:       owner,
		Repo:        repo,
		Branch:      branch,
		DocsPath:    docsPath,
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
	}

	if err := s.docRepo.Create(doc); err != nil {
		return nil, fmt.Errorf("failed to create documentation: %w", err)
	}

	return toDocumentationResponse(doc), nil
}

// GetDocumentationByID retrieves a documentation by ID
func (s *DocumentationService) GetDocumentationByID(id uuid.UUID) (*DocumentationResponse, error) {
	doc, err := s.docRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("documentation not found: %w", err)
	}
	return toDocumentationResponse(doc), nil
}

// GetDocumentationsByTeamID retrieves all documentations for a team
func (s *DocumentationService) GetDocumentationsByTeamID(teamID uuid.UUID) ([]DocumentationResponse, error) {
	// Validate team exists
	if _, err := s.teamRepo.GetByID(teamID); err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	docs, err := s.docRepo.GetByTeamID(teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documentations: %w", err)
	}

	responses := make([]DocumentationResponse, 0, len(docs))
	for i := range docs {
		responses = append(responses, *toDocumentationResponse(&docs[i]))
	}

	return responses, nil
}

// UpdateDocumentation updates an existing documentation
func (s *DocumentationService) UpdateDocumentation(id uuid.UUID, req *UpdateDocumentationRequest) (*DocumentationResponse, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if strings.TrimSpace(req.UpdatedBy) == "" {
		return nil, fmt.Errorf("updated_by is required")
	}

	// Get existing documentation
	doc, err := s.docRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("documentation not found: %w", err)
	}

	// Update fields if provided
	if req.URL != nil && *req.URL != "" {
		owner, repo, branch, docsPath, err := parseGitHubURL(*req.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid GitHub URL: %w", err)
		}
		doc.Owner = owner
		doc.Repo = repo
		doc.Branch = branch
		doc.DocsPath = docsPath
	}

	if req.Title != nil {
		doc.Title = *req.Title
	}

	if req.Description != nil {
		doc.Description = *req.Description
	}

	doc.UpdatedBy = req.UpdatedBy

	if err := s.docRepo.Update(doc); err != nil {
		return nil, fmt.Errorf("failed to update documentation: %w", err)
	}

	return toDocumentationResponse(doc), nil
}

// DeleteDocumentation deletes a documentation by ID
func (s *DocumentationService) DeleteDocumentation(id uuid.UUID) error {
	// Check if documentation exists
	if _, err := s.docRepo.GetByID(id); err != nil {
		return fmt.Errorf("documentation not found: %w", err)
	}

	if err := s.docRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete documentation: %w", err)
	}

	return nil
}

// parseGitHubURL parses a GitHub URL and extracts owner, repo, branch, and docs path
// Supports two formats:
// 1. Full path: https://github.tools.sap/{owner}/{repo}/tree/{branch}/{path}
// 2. Repository root: https://github.tools.sap/{owner}/{repo} (defaults to main branch and root path)
// Example: https://github.tools.sap/cfs-platform-engineering/cfs-platform-docs/tree/main/docs/coe
// Example: https://github.tools.sap/cfs-platform-engineering/developer-portal-frontend
func parseGitHubURL(urlStr string) (owner, repo, branch, docsPath string, err error) {
	// Parse URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate host (support both github.tools.sap and github.com)
	if u.Host != "github.tools.sap" && u.Host != "github.com" {
		return "", "", "", "", fmt.Errorf("invalid GitHub host: %s (expected github.tools.sap or github.com)", u.Host)
	}

	// Parse path
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("invalid GitHub URL format: expected at least /{owner}/{repo}")
	}

	owner = parts[0]
	repo = parts[1]

	// Check if this is a repository root URL (just /{owner}/{repo})
	if len(parts) == 2 {
		// Default to main branch and root path for repository root URLs
		branch = "main"
		docsPath = "/"
	} else if len(parts) >= 5 {
		// Full path format: /{owner}/{repo}/tree/{branch}/{docs_path}
		if parts[2] != "tree" && parts[2] != "blob" {
			return "", "", "", "", fmt.Errorf("invalid GitHub URL format: expected /tree/ or /blob/ in URL")
		}
		branch = parts[3]
		docsPath = strings.Join(parts[4:], "/")
	} else {
		return "", "", "", "", fmt.Errorf("invalid GitHub URL format: expected /{owner}/{repo} or /{owner}/{repo}/tree/{branch}/{path}")
	}

	if owner == "" || repo == "" || branch == "" {
		return "", "", "", "", fmt.Errorf("incomplete GitHub URL: missing owner, repo, or branch")
	}

	return owner, repo, branch, docsPath, nil
}

// toDocumentationResponse converts a Documentation model to DocumentationResponse
func toDocumentationResponse(doc *models.Documentation) *DocumentationResponse {
	return &DocumentationResponse{
		ID:          doc.ID.String(),
		TeamID:      doc.TeamID.String(),
		Owner:       doc.Owner,
		Repo:        doc.Repo,
		Branch:      doc.Branch,
		DocsPath:    doc.DocsPath,
		Title:       doc.Title,
		Description: doc.Description,
		CreatedAt:   doc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy:   doc.CreatedBy,
		UpdatedAt:   doc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedBy:   doc.UpdatedBy,
	}
}
