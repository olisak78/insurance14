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

// TeamService handles business logic for teams
type TeamService struct {
	repo             *repository.TeamRepository
	groupRepo        repository.GroupRepositoryInterface
	organizationRepo *repository.OrganizationRepository
	memberRepo       *repository.MemberRepository
	validator        *validator.Validate
}

// NewTeamService creates a new team service
func NewTeamService(repo *repository.TeamRepository, groupRepo repository.GroupRepositoryInterface, orgRepo *repository.OrganizationRepository, memberRepo *repository.MemberRepository, validator *validator.Validate) *TeamService {
	return &TeamService{
		repo:             repo,
		groupRepo:        groupRepo,
		organizationRepo: orgRepo,
		memberRepo:       memberRepo,
		validator:        validator,
	}
}

// Link represents a link with URL, title, icon, and category
type Link struct {
	URL      string `json:"url" validate:"required,url"`
	Title    string `json:"title" validate:"required"`
	Icon     string `json:"icon,omitempty"`
	Category string `json:"category,omitempty"`
}

// CreateTeamRequest represents the request to create a team
type CreateTeamRequest struct {
	GroupID     uuid.UUID         `json:"group_id" validate:"required"`
	Name        string            `json:"name" validate:"required,min=1,max=100"`
	DisplayName string            `json:"display_name" validate:"required,max=200"`
	Description string            `json:"description"`
	TeamLeadID  *uuid.UUID        `json:"team_lead_id,omitempty"`
	Status      models.TeamStatus `json:"status" validate:"required,oneof=active inactive archived"`
	Links       []Link            `json:"links,omitempty" validate:"dive"`
	Metadata    json.RawMessage   `json:"metadata" swaggertype:"object"`
}

// UpdateTeamRequest represents the request to update a team
type UpdateTeamRequest struct {
	DisplayName string             `json:"display_name" validate:"required,max=200"`
	Description string             `json:"description,omitempty"`
	TeamLeadID  *uuid.UUID         `json:"team_lead_id,omitempty"`
	Status      *models.TeamStatus `json:"status,omitempty"`
	Links       []Link             `json:"links,omitempty" validate:"dive"`
	Metadata    json.RawMessage    `json:"metadata,omitempty" swaggertype:"object"`
}

// TeamResponse represents the response for team operations
type TeamResponse struct {
	ID             uuid.UUID         `json:"id"`
	GroupID        uuid.UUID         `json:"group_id"`
	OrganizationID uuid.UUID         `json:"organization_id"` // Include org ID for backwards compatibility
	Name           string            `json:"name"`
	DisplayName    string            `json:"display_name"`
	Description    string            `json:"description"`
	TeamLeadID     *uuid.UUID        `json:"team_lead_id,omitempty"`
	Status         models.TeamStatus `json:"status"`
	Links          []Link            `json:"links,omitempty"`
	Metadata       json.RawMessage   `json:"metadata,omitempty" swaggertype:"object"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

// TeamListResponse represents a paginated list of teams
type TeamListResponse struct {
	Teams    []TeamResponse `json:"teams"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// Create creates a new team
func (s *TeamService) Create(req *CreateTeamRequest) (*TeamResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate group exists
	if _, err := s.groupRepo.GetByID(req.GroupID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to verify group: %w", err)
	}

	// Check if team with same name exists in group
	existingByName, err := s.repo.GetByName(req.GroupID, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing team by name: %w", err)
	}
	if existingByName != nil {
		return nil, apperrors.ErrTeamExists
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = models.TeamStatusActive
	}

	// Marshal links to JSON
	var linksJSON json.RawMessage
	if len(req.Links) > 0 {
		var err error
		linksJSON, err = json.Marshal(req.Links)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal links: %w", err)
		}
	}

	// Create team
	team := &models.Team{
		GroupID:     req.GroupID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      status,
		Links:       linksJSON,
		Metadata:    req.Metadata,
	}

	if err := s.repo.Create(team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return s.toResponse(team), nil
}

// GetByID retrieves a team by ID
func (s *TeamService) GetByID(id uuid.UUID) (*TeamResponse, error) {
	team, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return s.toResponse(team), nil
}

// GetByName retrieves a team by name within an organization
func (s *TeamService) GetByName(organizationID uuid.UUID, name string) (*TeamResponse, error) {
	team, err := s.repo.GetByName(organizationID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return s.toResponse(team), nil
}

// GetByOrganization retrieves teams for an organization with pagination
func (s *TeamService) GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*TeamListResponse, error) {
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
	teams, total, err := s.repo.GetByOrganizationID(organizationID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	responses := make([]TeamResponse, len(teams))
	for i, team := range teams {
		responses[i] = *s.toResponse(&team)
	}

	return &TeamListResponse{
		Teams:    responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Search searches teams by name or display name within an organization
func (s *TeamService) Search(organizationID uuid.UUID, query string, page, pageSize int) (*TeamListResponse, error) {
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
	teams, total, err := s.repo.Search(organizationID, query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search teams: %w", err)
	}

	responses := make([]TeamResponse, len(teams))
	for i, team := range teams {
		responses[i] = *s.toResponse(&team)
	}

	return &TeamListResponse{
		Teams:    responses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update updates a team
func (s *TeamService) Update(id uuid.UUID, req *UpdateTeamRequest) (*TeamResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing team
	team, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Update fields
	team.DisplayName = req.DisplayName
	team.Description = req.Description
	if req.Status != nil {
		team.Status = *req.Status
	}
	if req.Links != nil {
		linksJSON, err := json.Marshal(req.Links)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal links: %w", err)
		}
		team.Links = linksJSON
	}
	if req.Metadata != nil {
		team.Metadata = req.Metadata
	}

	if err := s.repo.Update(team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return s.toResponse(team), nil
}

// Delete deletes a team
func (s *TeamService) Delete(id uuid.UUID) error {
	// Check if team exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrTeamNotFound
		}
		return fmt.Errorf("failed to get team: %w", err)
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

// GetWithMembers retrieves a team with its members
func (s *TeamService) GetWithMembers(id uuid.UUID) (*models.Team, error) {
	team, err := s.repo.GetWithMembers(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team with members: %w", err)
	}

	return team, nil
}

// GetWithProjects retrieves a team with its projects
func (s *TeamService) GetWithProjects(id uuid.UUID) (*models.Team, error) {
	team, err := s.repo.GetWithProjects(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team with projects: %w", err)
	}

	return team, nil
}

// GetWithComponentOwnerships retrieves a team with its component ownerships
func (s *TeamService) GetWithComponentOwnerships(id uuid.UUID) (*models.Team, error) {
	team, err := s.repo.GetWithComponentOwnerships(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team with component ownerships: %w", err)
	}

	return team, nil
}

// GetWithDutySchedules retrieves a team with its duty schedules
func (s *TeamService) GetWithDutySchedules(id uuid.UUID) (*models.Team, error) {
	team, err := s.repo.GetWithDutySchedules(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team with duty schedules: %w", err)
	}

	return team, nil
}

// GetWithTeamLead retrieves a team with its team lead details
func (s *TeamService) GetTeamLead(id uuid.UUID) (*models.Team, error) {
	team, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

// GetAllTeams retrieves teams for a specific organization or all teams if organizationID is nil
func (s *TeamService) GetAllTeams(organizationID *uuid.UUID, page, pageSize int) (*TeamListResponse, error) {
	if organizationID != nil {
		// Get teams for specific organization
		return s.GetByOrganization(*organizationID, page, pageSize)
	}

	// Get all teams across all organizations (no pagination since user mentioned <100 teams)
	teams, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all teams: %w", err)
	}

	responses := make([]TeamResponse, len(teams))
	for i, team := range teams {
		responses[i] = *s.toResponse(&team)
	}

	return &TeamListResponse{
		Teams:    responses,
		Total:    int64(len(teams)),
		Page:     1,
		PageSize: len(teams),
	}, nil
}

// GetTeamMembersByName retrieves members of a team by team name within an organization
func (s *TeamService) GetTeamMembersByName(organizationID uuid.UUID, teamName string, page, pageSize int) ([]models.Member, int64, error) {
	// First get the team by name
	team, err := s.repo.GetByName(organizationID, teamName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperrors.ErrTeamNotFound
		}
		return nil, 0, fmt.Errorf("failed to get team: %w", err)
	}

	// Set pagination defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	members, total, err := s.memberRepo.GetByTeamID(team.ID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get team members: %w", err)
	}

	return members, total, nil
}

// GetTeamComponentsByName retrieves components owned by a team by team name within an organization
func (s *TeamService) GetTeamComponentsByName(organizationID uuid.UUID, teamName string, page, pageSize int) ([]models.Component, int64, error) {
	// First get the team by name
	team, err := s.repo.GetByName(organizationID, teamName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperrors.ErrTeamNotFound
		}
		return nil, 0, fmt.Errorf("failed to get team: %w", err)
	}

	// Get team with component ownerships to access components
	teamWithComponents, err := s.repo.GetWithComponentOwnerships(team.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get team components: %w", err)
	}

	// Extract components from team component ownerships
	components := make([]models.Component, 0)
	for _, ownership := range teamWithComponents.TeamComponentOwnerships {
		components = append(components, ownership.Component)
	}

	// Apply pagination manually since we're dealing with a preloaded slice
	total := int64(len(components))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(components) {
		return []models.Component{}, total, nil
	}
	if end > len(components) {
		end = len(components)
	}

	paginatedComponents := components[start:end]
	return paginatedComponents, total, nil
}

// GetTeamComponentsByID retrieves components owned by a team by team ID
func (s *TeamService) GetTeamComponentsByID(id uuid.UUID, page, pageSize int) ([]models.Component, int64, error) {
	// Get team with component ownerships to access components
	teamWithComponents, err := s.repo.GetWithComponentOwnerships(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperrors.ErrTeamNotFound
		}
		// Log the detailed error for debugging
		return nil, 0, fmt.Errorf("failed to get team components for team_id=%s: %w", id, err)
	}

	// Extract components from team component ownerships
	components := make([]models.Component, 0)
	for _, ownership := range teamWithComponents.TeamComponentOwnerships {
		components = append(components, ownership.Component)
	}

	// Apply pagination manually since we're dealing with a preloaded slice
	total := int64(len(components))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(components) {
		return []models.Component{}, total, nil
	}
	if end > len(components) {
		end = len(components)
	}

	paginatedComponents := components[start:end]
	return paginatedComponents, total, nil
}

// GetMembersOnly retrieves all members of a team without pagination
func (s *TeamService) GetMembersOnly(id uuid.UUID) ([]MemberResponse, error) {
	// First verify the team exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Get all members for the team (no pagination)
	members, _, err := s.memberRepo.GetByTeamID(id, 1000, 0) // Using large limit to get all
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	// Convert to member responses
	memberResponses := make([]MemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = MemberResponse{
			ID:             member.ID,
			OrganizationID: member.OrganizationID,
			TeamID:         member.TeamID,
			FullName:       member.FullName,
			FirstName:      member.FirstName,
			LastName:       member.LastName,
			Email:          member.Email,
			PhoneNumber:    member.PhoneNumber,
			IUser:          member.IUser,
			Role:           string(member.Role),
			IsActive:       member.IsActive,
			CreatedAt:      member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:      member.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return memberResponses, nil
}

// AddLinkRequest represents the request to add a link to a team
type AddLinkRequest struct {
	URL      string `json:"url" validate:"required,url"`
	Title    string `json:"title" validate:"required"`
	Icon     string `json:"icon,omitempty"`
	Category string `json:"category,omitempty"`
}

// AddLink adds a link to a team's links array
func (s *TeamService) AddLink(id uuid.UUID, req *AddLinkRequest) (*TeamResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing team
	team, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Parse existing links
	var links []Link
	if len(team.Links) > 0 {
		if err := json.Unmarshal(team.Links, &links); err != nil {
			return nil, fmt.Errorf("failed to parse existing links: %w", err)
		}
	}

	// Check if link with same URL already exists
	for _, link := range links {
		if link.URL == req.URL {
			return nil, apperrors.ErrLinkExists
		}
	}

	// Add new link
	newLink := Link{
		URL:      req.URL,
		Title:    req.Title,
		Icon:     req.Icon,
		Category: req.Category,
	}
	links = append(links, newLink)

	// Marshal back to JSON
	linksJSON, err := json.Marshal(links)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal links: %w", err)
	}

	team.Links = linksJSON

	if err := s.repo.Update(team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return s.toResponse(team), nil
}

// RemoveLink removes a link from a team's links array by URL
func (s *TeamService) RemoveLink(id uuid.UUID, linkURL string) (*TeamResponse, error) {
	if linkURL == "" {
		return nil, apperrors.NewValidationError("linkURL", "link URL is required")
	}

	// Get existing team
	team, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Parse existing links
	var links []Link
	if len(team.Links) > 0 {
		if err := json.Unmarshal(team.Links, &links); err != nil {
			return nil, fmt.Errorf("failed to parse existing links: %w", err)
		}
	}

	// Find and remove the link
	found := false
	newLinks := make([]Link, 0, len(links))
	for _, link := range links {
		if link.URL != linkURL {
			newLinks = append(newLinks, link)
		} else {
			found = true
		}
	}

	if !found {
		return nil, apperrors.ErrLinkNotFound
	}

	// Marshal back to JSON
	var linksJSON json.RawMessage
	if len(newLinks) > 0 {
		var err error
		linksJSON, err = json.Marshal(newLinks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal links: %w", err)
		}
	} else {
		linksJSON = json.RawMessage("[]")
	}

	team.Links = linksJSON

	if err := s.repo.Update(team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return s.toResponse(team), nil
}

// UpdateLinksRequest represents the request to update all links for a team
type UpdateLinksRequest struct {
	Links []Link `json:"links" validate:"required,dive"`
}

// UpdateLinks replaces all links for a team
func (s *TeamService) UpdateLinks(id uuid.UUID, req *UpdateLinksRequest) (*TeamResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing team
	team, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Marshal links to JSON
	linksJSON, err := json.Marshal(req.Links)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal links: %w", err)
	}

	team.Links = linksJSON

	if err := s.repo.Update(team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return s.toResponse(team), nil
}

// toResponse converts a team model to response
func (s *TeamService) toResponse(team *models.Team) *TeamResponse {
	// Get organization ID through group (for backwards compatibility)
	var organizationID uuid.UUID
	if group, err := s.groupRepo.GetByID(team.GroupID); err == nil {
		organizationID = group.OrganizationID
	}

	// Unmarshal links from JSON
	var links []Link
	if len(team.Links) > 0 {
		json.Unmarshal(team.Links, &links) // Ignore error, will return empty array
	}

	return &TeamResponse{
		ID:             team.ID,
		GroupID:        team.GroupID,
		OrganizationID: organizationID,
		Name:           team.Name,
		DisplayName:    team.DisplayName,
		Description:    team.Description,
		TeamLeadID:     nil,
		Status:         team.Status,
		Links:          links,
		Metadata:       team.Metadata,
		CreatedAt:      team.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      team.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
