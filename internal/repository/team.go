package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamRepository handles database operations for teams
type TeamRepository struct {
	db *gorm.DB
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create creates a new team
func (r *TeamRepository) Create(team *models.Team) error {
	return r.db.Create(team).Error
}

// GetByID retrieves a team by ID
func (r *TeamRepository) GetByID(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetByName retrieves a team by name within a group
func (r *TeamRepository) GetByName(groupID uuid.UUID, name string) (*models.Team, error) {
	var team models.Team
	err := r.db.First(&team, "group_id = ? AND name = ?", groupID, name).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetByGroupID retrieves all teams for a group with pagination
func (r *TeamRepository) GetByGroupID(groupID uuid.UUID, limit, offset int) ([]models.Team, int64, error) {
	var teams []models.Team
	var total int64

	// Get total count
	if err := r.db.Model(&models.Team{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Where("group_id = ?", groupID).Limit(limit).Offset(offset).Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// GetByOrganizationID retrieves all teams for an organization (through groups) with pagination
func (r *TeamRepository) GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Team, int64, error) {
	var teams []models.Team
	var total int64

	// Get total count - join with groups to filter by organization
	if err := r.db.Model(&models.Team{}).
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.organization_id = ?", orgID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.organization_id = ?", orgID).
		Limit(limit).Offset(offset).
		Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// GetActiveTeams retrieves all active teams for an organization (through groups)
func (r *TeamRepository) GetActiveTeams(orgID uuid.UUID, limit, offset int) ([]models.Team, int64, error) {
	var teams []models.Team
	var total int64

	query := r.db.Model(&models.Team{}).
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.organization_id = ? AND teams.status = ?", orgID, models.TeamStatusActive)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// Update updates a team
func (r *TeamRepository) Update(team *models.Team) error {
	return r.db.Save(team).Error
}

// Delete deletes a team
func (r *TeamRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Team{}, "id = ?", id).Error
}

// GetWithGroup retrieves a team with group details (and organization through group)
func (r *TeamRepository) GetWithGroup(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("Group").Preload("Group.Organization").First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetWithOrganization retrieves a team with organization details (legacy method)
func (r *TeamRepository) GetWithOrganization(id uuid.UUID) (*models.Team, error) {
	return r.GetWithGroup(id)
}

// GetWithMembers retrieves a team with all its members
func (r *TeamRepository) GetWithMembers(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("Members").First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetWithComponentOwnerships retrieves a team with component ownerships
func (r *TeamRepository) GetWithComponentOwnerships(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("TeamComponentOwnerships").Preload("TeamComponentOwnerships.Component").First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetWithDutySchedules retrieves a team with duty schedules
func (r *TeamRepository) GetWithDutySchedules(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("DutySchedules").First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetWithProjects retrieves a team with their projects through component ownership
func (r *TeamRepository) GetWithProjects(id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.Preload("TeamComponentOwnerships").
		Preload("TeamComponentOwnerships.Component").
		Preload("TeamComponentOwnerships.Component.ProjectComponents").
		First(&team, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// SetStatus sets the status of a team
func (r *TeamRepository) SetStatus(teamID uuid.UUID, status models.TeamStatus) error {
	return r.db.Model(&models.Team{}).Where("id = ?", teamID).Update("status", status).Error
}

// Search searches for teams by name or description within an organization (through groups)
func (r *TeamRepository) Search(orgID uuid.UUID, query string, limit, offset int) ([]models.Team, int64, error) {
	var teams []models.Team
	var total int64

	searchQuery := r.db.Model(&models.Team{}).
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.organization_id = ? AND (teams.name ILIKE ? OR teams.description ILIKE ?)", orgID, "%"+query+"%", "%"+query+"%")

	// Get total count
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := searchQuery.Limit(limit).Offset(offset).Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// GetMemberCount returns the number of members in a team
func (r *TeamRepository) GetMemberCount(teamID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Member{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}

// GetComponentOwnershipCount returns the number of components owned by a team
func (r *TeamRepository) GetComponentOwnershipCount(teamID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.TeamComponentOwnership{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}

// GetAll retrieves all teams across all organizations
func (r *TeamRepository) GetAll() ([]models.Team, error) {
	var teams []models.Team
	err := r.db.Find(&teams).Error
	return teams, err
}

// GetTeamsWithMemberCount retrieves teams with their member counts for an organization (through groups)
func (r *TeamRepository) GetTeamsWithMemberCount(orgID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error) {
	var teams []models.Team
	var total int64
	var results []map[string]interface{}

	// Get total count
	if err := r.db.Model(&models.Team{}).
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.organization_id = ?", orgID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get teams with member counts
	err := r.db.Raw(`
		SELECT t.*, COUNT(m.id) as member_count
		FROM teams t
		JOIN groups g ON t.group_id = g.id
		LEFT JOIN members m ON t.id = m.team_id
		WHERE g.organization_id = ?
		GROUP BY t.id, g.id
		LIMIT ? OFFSET ?
	`, orgID, limit, offset).Scan(&teams).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert to map format for easier JSON handling
	for _, team := range teams {
		teamMap := map[string]interface{}{
			"id":           team.ID,
			"name":         team.Name,
			"display_name": team.DisplayName,
			"description":  team.Description,
			"group_id":     team.GroupID,
			"status":       team.Status,
			"created_at":   team.CreatedAt,
			"updated_at":   team.UpdatedAt,
		}
		results = append(results, teamMap)
	}

	return results, total, nil
}

// CheckTeamExists checks if a team exists by ID
func (r *TeamRepository) CheckTeamExists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Team{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// CheckTeamNameExists checks if a team name exists within an organization (through groups)
func (r *TeamRepository) CheckTeamNameExists(orgID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&models.Team{}).
		Joins("JOIN groups ON teams.group_id = groups.id").
		Where("groups.organization_id = ? AND teams.name = ?", orgID, name)
	if excludeID != nil {
		query = query.Where("teams.id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
