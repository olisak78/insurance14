package repository

import (
	"developer-portal-backend/internal/database/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles database operations for members
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new member repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new member
func (r *UserRepository) Create(member *models.User) error {
	return r.db.Create(member).Error
}

// GetByID retrieves a member by ID
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var member models.User
	err := r.db.First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetByEmail retrieves a member by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var member models.User
	err := r.db.First(&member, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetByName retrieves a member by BaseModel.Name column
func (r *UserRepository) GetByName(name string) (*models.User, error) {
	var member models.User
	err := r.db.First(&member, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

 // GetByUserID retrieves a member by their string UserID (e.g., I123456)
func (r *UserRepository) GetByUserID(userID string) (*models.User, error) {
	var member models.User
	err := r.db.First(&member, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetAll retrieves all users with pagination
func (r *UserRepository) GetAll(limit, offset int) ([]models.User, int64, error) {
	var members []models.User
	var total int64

	// Get total count
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := r.db.Model(&models.User{}).Limit(limit).Offset(offset).Find(&members).Error; err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetByOrganizationID retrieves all members for an organization with pagination
func (r *UserRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	var members []models.User
	var total int64

	// Users no longer carry organization_id directly; filter by organization through team -> group
	countQuery := r.db.Model(&models.User{}).
		Joins("JOIN teams ON members.team_id = teams.id").
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.org_id = ?", orgID)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Model(&models.User{}).
		Joins("JOIN teams ON members.team_id = teams.id").
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.org_id = ?", orgID).
		Limit(limit).Offset(offset).
		Find(&members).Error
	if err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

// GetByTeamID retrieves all members for a team with pagination
func (r *UserRepository) GetByTeamID(teamID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	var members []models.User
	var total int64

	// Get total count
	if err := r.db.Model(&models.User{}).Where("team_id = ?", teamID).Count(&total).Error; err != nil {
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
func (r *UserRepository) GetByRole(orgID uuid.UUID, role models.TeamDomain, limit, offset int) ([]models.User, int64, error) {
	var members []models.User
	var total int64

	query := r.db.Model(&models.User{}).
		Joins("JOIN teams ON members.team_id = teams.id").
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.org_id = ? AND members.team_domain = ?", orgID, role)

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
func (r *UserRepository) GetActiveMembers(orgID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	// Model no longer has an 'is_active' field; return all members for org via team->group join
	return r.GetByOrganizationID(orgID, limit, offset)
}

// Update updates a member
func (r *UserRepository) Update(member *models.User) error {
	return r.db.Save(member).Error
}

// Delete deletes a member
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// GetWithOrganization retrieves a member with organization details
func (r *UserRepository) GetWithOrganization(id uuid.UUID) (*models.User, error) {
	var member models.User
	err := r.db.Preload("Organization").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetWithTeam retrieves a member with team details
func (r *UserRepository) GetWithTeam(id uuid.UUID) (*models.User, error) {
	var member models.User
	err := r.db.Preload("Team").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetWithLeadingTeams retrieves a member with teams they lead
func (r *UserRepository) GetWithLeadingTeams(id uuid.UUID) (*models.User, error) {
	var member models.User
	err := r.db.Preload("LeadingTeams").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetWithDutySchedules retrieves a member with their duty schedules
func (r *UserRepository) GetWithDutySchedules(id uuid.UUID) (*models.User, error) {
	var member models.User
	err := r.db.Preload("DutySchedules").First(&member, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// AssignToTeam assigns a member to a team
func (r *UserRepository) AssignToTeam(memberID, teamID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", memberID).Update("team_id", teamID).Error
}

// RemoveFromTeam removes a member from their team
func (r *UserRepository) RemoveFromTeam(memberID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", memberID).Update("team_id", nil).Error
}

// UpdateRole updates a member's role
func (r *UserRepository) UpdateRole(memberID uuid.UUID, role models.TeamDomain) error {
	return r.db.Model(&models.User{}).Where("id = ?", memberID).Update("team_domain", role).Error
}

// SetActiveStatus sets the active status of a member
func (r *UserRepository) SetActiveStatus(memberID uuid.UUID, isActive bool) error {
	// Model no longer has 'is_active'; treat as no-op for backward compatibility
	return nil
}

// Search searches for members by name or email
func (r *UserRepository) Search(orgID uuid.UUID, query string, limit, offset int) ([]models.User, int64, error) {
	var members []models.User
	var total int64

	searchQuery := r.db.Model(&models.User{}).
		Joins("JOIN teams ON members.team_id = teams.id").
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.org_id = ? AND (members.first_name ILIKE ? OR members.last_name ILIKE ? OR members.email ILIKE ?)", orgID, "%"+query+"%", "%"+query+"%", "%"+query+"%")

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
func (r *UserRepository) SearchByOrganization(orgID uuid.UUID, query string, limit, offset int) ([]models.User, int64, error) {
	return r.Search(orgID, query, limit, offset)
}

 // GetActiveByOrganization retrieves all active members for an organization
func (r *UserRepository) GetActiveByOrganization(orgID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	return r.GetByOrganizationID(orgID, limit, offset)
}

 // GetUserIDsByPrefix returns user_ids with the given prefix (case-insensitive)
func (r *UserRepository) GetUserIDsByPrefix(prefix string) ([]string, error) {
	var ids []string
	if err := r.db.Model(&models.User{}).
		Where("user_id ILIKE ?", prefix+"%").
		Pluck("user_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// GetExistingUserIDs returns the subset of provided user_ids that already exist
func (r *UserRepository) GetExistingUserIDs(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	var existing []string
	if err := r.db.Model(&models.User{}).
		Where("user_id IN ?", ids).
		Pluck("user_id", &existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

// SearchByNameOrTitleGlobal performs a case-insensitive search across users by BaseModel.Name or BaseModel.Title
func (r *UserRepository) SearchByNameOrTitleGlobal(query string, limit, offset int) ([]models.User, int64, error) {
	var members []models.User
	var total int64

	q := strings.TrimSpace(query)
	if q == "" {
		// When query is empty, behave like GetAll
		if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
			return nil, 0, err
		}
		if err := r.db.Model(&models.User{}).Limit(limit).Offset(offset).Find(&members).Error; err != nil {
			return nil, 0, err
		}
		return members, total, nil
	}

	searchQuery := r.db.Model(&models.User{}).
		Where("name ILIKE ? OR title ILIKE ?", "%"+q+"%", "%"+q+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := searchQuery.Limit(limit).Offset(offset).Find(&members).Error; err != nil {
		return nil, 0, err
	}

	return members, total, nil
}
