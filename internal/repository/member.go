package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MemberRepository handles database operations for members
type MemberRepository struct {
	db *gorm.DB
}

// NewMemberRepository creates a new member repository
func NewMemberRepository(db *gorm.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

// Create creates a new member
func (r *MemberRepository) Create(member *models.Member) error {
	return r.db.Create(member).Error
}

// GetByID retrieves a member by ID
func (r *MemberRepository) GetByID(id uuid.UUID) (*models.Member, error) {
	var member models.Member
	err := r.db.First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetByEmail retrieves a member by email
func (r *MemberRepository) GetByEmail(email string) (*models.Member, error) {
	var member models.Member
	err := r.db.First(&member, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetByOrganizationID retrieves all members for an organization with pagination
func (r *MemberRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	// Get total count
	if err := r.db.Model(&models.Member{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("organization_id = ?", orgID).Limit(limit).Offset(offset).Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetByTeamID retrieves all members for a team with pagination
func (r *MemberRepository) GetByTeamID(teamID uuid.UUID, limit, offset int) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	// Get total count
	if err := r.db.Model(&models.Member{}).Where("team_id = ?", teamID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("team_id = ?", teamID).Limit(limit).Offset(offset).Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetByRole retrieves all members with a specific role in an organization
func (r *MemberRepository) GetByRole(orgID uuid.UUID, role models.MemberRole, limit, offset int) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	query := r.db.Model(&models.Member{}).Where("organization_id = ? AND role = ?", orgID, role)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetActiveMembers retrieves all active members for an organization
func (r *MemberRepository) GetActiveMembers(orgID uuid.UUID, limit, offset int) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	query := r.db.Model(&models.Member{}).Where("organization_id = ? AND is_active = ?", orgID, true)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// Update updates a member
func (r *MemberRepository) Update(member *models.Member) error {
	return r.db.Save(member).Error
}

// Delete deletes a member
func (r *MemberRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Member{}, "id = ?", id).Error
}

// GetWithOrganization retrieves a member with organization details
func (r *MemberRepository) GetWithOrganization(id uuid.UUID) (*models.Member, error) {
	var member models.Member
	err := r.db.Preload("Organization").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetWithTeam retrieves a member with team details
func (r *MemberRepository) GetWithTeam(id uuid.UUID) (*models.Member, error) {
	var member models.Member
	err := r.db.Preload("Team").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetWithLeadingTeams retrieves a member with teams they lead
func (r *MemberRepository) GetWithLeadingTeams(id uuid.UUID) (*models.Member, error) {
	var member models.Member
	err := r.db.Preload("LeadingTeams").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetWithDutySchedules retrieves a member with their duty schedules
func (r *MemberRepository) GetWithDutySchedules(id uuid.UUID) (*models.Member, error) {
	var member models.Member
	err := r.db.Preload("DutySchedules").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// AssignToTeam assigns a member to a team
func (r *MemberRepository) AssignToTeam(memberID, teamID uuid.UUID) error {
	return r.db.Model(&models.Member{}).Where("id = ?", memberID).Update("team_id", teamID).Error
}

// RemoveFromTeam removes a member from their team
func (r *MemberRepository) RemoveFromTeam(memberID uuid.UUID) error {
	return r.db.Model(&models.Member{}).Where("id = ?", memberID).Update("team_id", nil).Error
}

// UpdateRole updates a member's role
func (r *MemberRepository) UpdateRole(memberID uuid.UUID, role models.MemberRole) error {
	return r.db.Model(&models.Member{}).Where("id = ?", memberID).Update("role", role).Error
}

// SetActiveStatus sets the active status of a member
func (r *MemberRepository) SetActiveStatus(memberID uuid.UUID, isActive bool) error {
	return r.db.Model(&models.Member{}).Where("id = ?", memberID).Update("is_active", isActive).Error
}

// Search searches for members by name or email
func (r *MemberRepository) Search(orgID uuid.UUID, query string, limit, offset int) ([]models.Member, int64, error) {
	var members []models.Member
	var total int64

	searchQuery := r.db.Model(&models.Member{}).Where("organization_id = ? AND (full_name ILIKE ? OR email ILIKE ?)", orgID, "%"+query+"%", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := searchQuery.Limit(limit).Offset(offset).Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// SearchByOrganization searches for members by name or email within an organization
func (r *MemberRepository) SearchByOrganization(orgID uuid.UUID, query string, limit, offset int) ([]models.Member, int64, error) {
	return r.Search(orgID, query, limit, offset)
}

// GetActiveByOrganization retrieves all active members for an organization
func (r *MemberRepository) GetActiveByOrganization(orgID uuid.UUID, limit, offset int) ([]models.Member, int64, error) {
	return r.GetActiveMembers(orgID, limit, offset)
}
