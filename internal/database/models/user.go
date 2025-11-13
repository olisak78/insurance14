package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type TeamDomain string

const (
	TeamDomainDeveloper TeamDomain = "developer"
	TeamDomainDevOps    TeamDomain = "devops"
	TeamDomainPO        TeamDomain = "po"
	TeamDomainArchitect TeamDomain = "architect"
)

// TeamRole represents the role within a team context
type TeamRole string

const (
	TeamRoleMember  TeamRole = "member"
	TeamRoleScM     TeamRole = "scm"
	TeamRoleManager TeamRole = "manager"
	TeamRoleMMM     TeamRole = "mmm"
)

// Member represents a member of an organization (replaces User)
type User struct {
	BaseModel
	TeamID      *uuid.UUID      `json:"team_id,omitempty" gorm:"type:uuid;index"`
	UserID      string          `gorm:"not null;size:20" validate:"required,min=5,max=20"` // I/C/D user
	FirstName   string          `json:"first_name" gorm:"not null;size:100" validate:"required,max=100"`
	LastName    string          `json:"last_name" gorm:"not null;size:100" validate:"required,max=100"`
	Email       string          `json:"email" gorm:"uniqueIndex:idx_members_email_active;not null;size:255" validate:"required,email,max=255"`
	Mobile      string          `json:"mobile" gorm:"size:20"`
	TeamDomain  TeamDomain      `json:"role" gorm:"type:varchar(50);not null;default:'developer'" validate:"required"`
	TeamRole    TeamRole        `json:"team_role" gorm:"type:varchar(50);not null;default:'member'"`
	Metadata    json.RawMessage `json:"metadata" gorm:"type:jsonb"`
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}
