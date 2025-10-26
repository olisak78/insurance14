package service

import (
	"developer-portal-backend/internal/database/models"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/repository"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// MemberService handles business logic for members
type MemberService struct {
	repo      repository.MemberRepositoryInterface
	validator *validator.Validate
}

// NewMemberService creates a new member service
func NewMemberService(repo repository.MemberRepositoryInterface, validator *validator.Validate) *MemberService {
	return &MemberService{
		repo:      repo,
		validator: validator,
	}
}

// CreateMemberRequest represents the data needed to create a member
type CreateMemberRequest struct {
	OrganizationID uuid.UUID       `json:"organization_id" validate:"required"`
	TeamID         *uuid.UUID      `json:"team_id"`
	FullName       string          `json:"full_name" validate:"required,max=200"`
	FirstName      string          `json:"first_name" validate:"required,max=100"`
	LastName       string          `json:"last_name" validate:"required,max=100"`
	Email          string          `json:"email" validate:"required,email,max=255"`
	PhoneNumber    string          `json:"phone_number" validate:"max=20"`
	IUser          string          `json:"iuser" validate:"required,max=50"`
	Role           *string         `json:"role" example:"developer" default:"developer"`        // Optional: defaults to "developer" if not provided. Valid values: admin, developer, manager, viewer
	TeamRole       *string         `json:"team_role" example:"member" default:"member"`         // Optional: defaults to "member" if not provided. Valid values: member, team_lead
	ExternalType   *string         `json:"external_type" example:"internal" default:"internal"` // Optional: defaults to "internal" if not provided
	Metadata       json.RawMessage `json:"metadata" swaggertype:"object"`
	IsActive       *bool           `json:"is_active" example:"true" default:"true"` // Optional: defaults to true if not provided
}

// UpdateMemberRequest represents the data needed to update a member
type UpdateMemberRequest struct {
	TeamID       *uuid.UUID      `json:"team_id"`
	FullName     *string         `json:"full_name" validate:"omitempty,max=200"`
	FirstName    *string         `json:"first_name" validate:"omitempty,max=100"`
	LastName     *string         `json:"last_name" validate:"omitempty,max=100"`
	Email        *string         `json:"email" validate:"omitempty,email,max=255"`
	PhoneNumber  *string         `json:"phone_number" validate:"omitempty,max=20"`
	IUser        *string         `json:"iuser" validate:"omitempty,max=50"`
	Role         *string         `json:"role"`
	TeamRole     *string         `json:"team_role"`
	ExternalType *string         `json:"external_type"`
	Metadata     json.RawMessage `json:"metadata" swaggertype:"object"`
	IsActive     *bool           `json:"is_active"`
}

// MemberResponse represents the response data for a member
type MemberResponse struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	TeamID         *uuid.UUID      `json:"team_id,omitempty"`
	FullName       string          `json:"full_name"`
	FirstName      string          `json:"first_name"`
	LastName       string          `json:"last_name"`
	Email          string          `json:"email"`
	PhoneNumber    string          `json:"phone_number"`
	IUser          string          `json:"iuser"`
	Role           string          `json:"role"`
	TeamRole       string          `json:"team_role"`
	ExternalType   string          `json:"external_type"`
	Metadata       json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	IsActive       bool            `json:"is_active"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

// CreateMember creates a new member
func (s *MemberService) CreateMember(req *CreateMemberRequest) (*MemberResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if email already exists
	existingMember, err := s.repo.GetByEmail(req.Email)
	if err == nil && existingMember.OrganizationID == req.OrganizationID {
		return nil, apperrors.ErrMemberExists
	}

	// Set default active status
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Set default role
	role := models.MemberRoleDeveloper
	if req.Role != nil {
		role = models.MemberRole(*req.Role)
	}

	// Set default team role
	teamRole := models.TeamRoleMember
	if req.TeamRole != nil {
		teamRole = models.TeamRole(*req.TeamRole)
	}

	// Set default external type
	externalType := models.ExternalTypeInternal
	if req.ExternalType != nil {
		externalType = models.ExternalType(*req.ExternalType)
	}

	member := &models.Member{
		OrganizationID: req.OrganizationID,
		TeamID:         req.TeamID,
		FullName:       req.FullName,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		PhoneNumber:    req.PhoneNumber,
		IUser:          req.IUser,
		Role:           role,
		TeamRole:       teamRole,
		ExternalType:   externalType,
		Metadata:       req.Metadata,
		IsActive:       isActive,
	}

	if err := s.repo.Create(member); err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	return s.convertToResponse(member), nil
}

// GetMemberByID retrieves a member by ID
func (s *MemberService) GetMemberByID(id uuid.UUID) (*MemberResponse, error) {
	member, err := s.repo.GetByID(id)
	if err != nil {
		return nil, apperrors.ErrMemberNotFound
	}

	return s.convertToResponse(member), nil
}

// GetMembersByOrganization retrieves members for an organization
func (s *MemberService) GetMembersByOrganization(organizationID uuid.UUID, limit, offset int) ([]MemberResponse, int64, error) {
	members, total, err := s.repo.GetByOrganizationID(organizationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get members: %w", err)
	}

	responses := make([]MemberResponse, len(members))
	for i, member := range members {
		responses[i] = *s.convertToResponse(&member)
	}

	return responses, total, nil
}

// UpdateMember updates an existing member
func (s *MemberService) UpdateMember(id uuid.UUID, req *UpdateMemberRequest) (*MemberResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	member, err := s.repo.GetByID(id)
	if err != nil {
		return nil, apperrors.ErrMemberNotFound
	}

	// Check email uniqueness if email is being updated
	if req.Email != nil && *req.Email != member.Email {
		existingMember, err := s.repo.GetByEmail(*req.Email)
		if err == nil && existingMember.OrganizationID == member.OrganizationID && existingMember.ID != id {
			return nil, apperrors.ErrMemberExists
		}
	}

	// Update fields
	if req.TeamID != nil {
		member.TeamID = req.TeamID
	}
	if req.FullName != nil {
		member.FullName = *req.FullName
	}
	if req.FirstName != nil {
		member.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		member.LastName = *req.LastName
	}
	if req.Email != nil {
		member.Email = *req.Email
	}
	if req.PhoneNumber != nil {
		member.PhoneNumber = *req.PhoneNumber
	}
	if req.IUser != nil {
		member.IUser = *req.IUser
	}
	if req.Role != nil {
		member.Role = models.MemberRole(*req.Role)
	}
	if req.TeamRole != nil {
		member.TeamRole = models.TeamRole(*req.TeamRole)
	}
	if req.ExternalType != nil {
		member.ExternalType = models.ExternalType(*req.ExternalType)
	}
	if req.Metadata != nil {
		member.Metadata = req.Metadata
	}
	if req.IsActive != nil {
		member.IsActive = *req.IsActive
	}

	if err := s.repo.Update(member); err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	return s.convertToResponse(member), nil
}

// DeleteMember deletes a member
func (s *MemberService) DeleteMember(id uuid.UUID) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return apperrors.ErrMemberNotFound
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}

	return nil
}

// SearchMembers searches for members by name or email
func (s *MemberService) SearchMembers(organizationID uuid.UUID, query string, limit, offset int) ([]MemberResponse, int64, error) {
	members, total, err := s.repo.SearchByOrganization(organizationID, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search members: %w", err)
	}

	responses := make([]MemberResponse, len(members))
	for i, member := range members {
		responses[i] = *s.convertToResponse(&member)
	}

	return responses, total, nil
}

// GetActiveMembers retrieves active members for an organization
func (s *MemberService) GetActiveMembers(organizationID uuid.UUID, limit, offset int) ([]MemberResponse, int64, error) {
	members, total, err := s.repo.GetActiveByOrganization(organizationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get active members: %w", err)
	}

	responses := make([]MemberResponse, len(members))
	for i, member := range members {
		responses[i] = *s.convertToResponse(&member)
	}

	return responses, total, nil
}

// convertToResponse converts a member model to response
func (s *MemberService) convertToResponse(member *models.Member) *MemberResponse {
	return &MemberResponse{
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
		TeamRole:       string(member.TeamRole),
		ExternalType:   string(member.ExternalType),
		Metadata:       member.Metadata,
		IsActive:       member.IsActive,
		CreatedAt:      member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      member.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// AddQuickLinkRequest represents the request to add a quick link to a member
type AddQuickLinkRequest struct {
	URL      string `json:"url" validate:"required,url"`
	Title    string `json:"title" validate:"required"`
	Icon     string `json:"icon,omitempty"`
	Category string `json:"category,omitempty"`
}

// QuickLink represents a quick link in the response
type QuickLink struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Icon     string `json:"icon,omitempty"`
	Category string `json:"category,omitempty"`
}

// QuickLinksResponse represents the response for getting quick links
type QuickLinksResponse struct {
	QuickLinks []QuickLink `json:"quick_links"`
}

// GetQuickLinks retrieves quick links from a member's metadata
func (s *MemberService) GetQuickLinks(id uuid.UUID) (*QuickLinksResponse, error) {
	// Get existing member
	member, err := s.repo.GetByID(id)
	if err != nil {
		return nil, apperrors.ErrMemberNotFound
	}

	// Parse existing metadata
	var metadata map[string]interface{}
	if len(member.Metadata) > 0 {
		if err := json.Unmarshal(member.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	} else {
		// Return empty quick links if no metadata
		return &QuickLinksResponse{QuickLinks: []QuickLink{}}, nil
	}

	// Get quick_links array
	var quickLinks []QuickLink
	if links, exists := metadata["quick_links"]; exists {
		if linksArray, ok := links.([]interface{}); ok {
			for _, link := range linksArray {
				if linkMap, ok := link.(map[string]interface{}); ok {
					quickLink := QuickLink{}
					if url, ok := linkMap["url"].(string); ok {
						quickLink.URL = url
					}
					if title, ok := linkMap["title"].(string); ok {
						quickLink.Title = title
					}
					if icon, ok := linkMap["icon"].(string); ok {
						quickLink.Icon = icon
					}
					if category, ok := linkMap["category"].(string); ok {
						quickLink.Category = category
					}
					quickLinks = append(quickLinks, quickLink)
				}
			}
		}
	}

	// Return empty array if no quick links found
	if quickLinks == nil {
		quickLinks = []QuickLink{}
	}

	return &QuickLinksResponse{QuickLinks: quickLinks}, nil
}

// AddQuickLink adds a quick link to a member's metadata.quick_links array
func (s *MemberService) AddQuickLink(id uuid.UUID, req *AddQuickLinkRequest) (*MemberResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing member
	member, err := s.repo.GetByID(id)
	if err != nil {
		return nil, apperrors.ErrMemberNotFound
	}

	// Parse existing metadata
	var metadata map[string]interface{}
	if len(member.Metadata) > 0 {
		if err := json.Unmarshal(member.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse existing metadata: %w", err)
		}
	} else {
		metadata = make(map[string]interface{})
	}

	// Get quick_links array
	var quickLinks []map[string]interface{}
	if links, exists := metadata["quick_links"]; exists {
		if linksArray, ok := links.([]interface{}); ok {
			for _, link := range linksArray {
				if linkMap, ok := link.(map[string]interface{}); ok {
					quickLinks = append(quickLinks, linkMap)
				}
			}
		}
	}

	// Check if link with same URL already exists
	for _, link := range quickLinks {
		if url, ok := link["url"].(string); ok && url == req.URL {
			return nil, apperrors.ErrLinkExists
		}
	}

	// Add new link
	newLink := map[string]interface{}{
		"url":   req.URL,
		"title": req.Title,
	}
	if req.Icon != "" {
		newLink["icon"] = req.Icon
	}
	if req.Category != "" {
		newLink["category"] = req.Category
	}
	quickLinks = append(quickLinks, newLink)

	// Update metadata
	metadata["quick_links"] = quickLinks

	// Marshal back to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	member.Metadata = metadataJSON

	if err := s.repo.Update(member); err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	return s.convertToResponse(member), nil
}

// RemoveQuickLink removes a quick link from a member's metadata.quick_links array by URL
func (s *MemberService) RemoveQuickLink(id uuid.UUID, linkURL string) (*MemberResponse, error) {
	if linkURL == "" {
		return nil, apperrors.NewValidationError("url", "link URL is required")
	}

	// Get existing member
	member, err := s.repo.GetByID(id)
	if err != nil {
		return nil, apperrors.ErrMemberNotFound
	}

	// Parse existing metadata
	var metadata map[string]interface{}
	if len(member.Metadata) > 0 {
		if err := json.Unmarshal(member.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse existing metadata: %w", err)
		}
	} else {
		return nil, apperrors.ErrLinkNotFound
	}

	// Get quick_links array
	var quickLinks []map[string]interface{}
	if links, exists := metadata["quick_links"]; exists {
		if linksArray, ok := links.([]interface{}); ok {
			for _, link := range linksArray {
				if linkMap, ok := link.(map[string]interface{}); ok {
					quickLinks = append(quickLinks, linkMap)
				}
			}
		}
	}

	if len(quickLinks) == 0 {
		return nil, apperrors.ErrLinkNotFound
	}

	// Find and remove the link
	found := false
	newLinks := make([]map[string]interface{}, 0, len(quickLinks))
	for _, link := range quickLinks {
		if url, ok := link["url"].(string); ok && url == linkURL {
			found = true
		} else {
			newLinks = append(newLinks, link)
		}
	}

	if !found {
		return nil, apperrors.ErrLinkNotFound
	}

	// Update metadata
	if len(newLinks) > 0 {
		metadata["quick_links"] = newLinks
	} else {
		metadata["quick_links"] = []interface{}{}
	}

	// Marshal back to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	member.Metadata = metadataJSON

	if err := s.repo.Update(member); err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	return s.convertToResponse(member), nil
}
