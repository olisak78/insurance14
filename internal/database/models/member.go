package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// MemberRole represents the role of a member in an organization
type MemberRole string

const (
	MemberRoleAdmin     MemberRole = "admin"
	MemberRoleDeveloper MemberRole = "developer"
	MemberRoleManager   MemberRole = "manager"
	MemberRoleViewer    MemberRole = "viewer"
)

// ExternalType represents the type of external system integration
type ExternalType string

const (
	ExternalTypeInternal ExternalType = "internal"
	ExternalTypeGitHub   ExternalType = "github"
	ExternalTypeJira     ExternalType = "jira"
	ExternalTypeLDAP     ExternalType = "ldap"
)

// TeamRole represents the role within a team context
type TeamRole string

const (
	TeamRoleMember   TeamRole = "member"
	TeamRoleTeamLead TeamRole = "team_lead"
)

// Member represents a member of an organization (replaces User)
type Member struct {
	BaseModel
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	GroupID        *uuid.UUID      `json:"group_id,omitempty" gorm:"type:uuid;index"`
	TeamID         *uuid.UUID      `json:"team_id,omitempty" gorm:"type:uuid;index"`
	FullName       string          `json:"full_name" gorm:"not null;size:200" validate:"required,max=200"`
	FirstName      string          `json:"first_name" gorm:"not null;size:100" validate:"required,max=100"`
	LastName       string          `json:"last_name" gorm:"not null;size:100" validate:"required,max=100"`
	Email          string          `json:"email" gorm:"uniqueIndex:idx_members_email_active,where:deleted_at IS NULL;not null;size:255" validate:"required,email,max=255"` // Partial unique index excludes soft-deleted records to allow re-adding members
	PhoneNumber    string          `json:"phone_number" gorm:"size:20"`
	IUser          string          `json:"iuser" gorm:"size:50;uniqueIndex:idx_members_iuser_active,where:deleted_at IS NULL" validate:"required,max=50"` // Partial unique index excludes soft-deleted records to allow re-adding members
	Role           MemberRole      `json:"role" gorm:"type:varchar(50);not null;default:'developer'" validate:"required"`
	TeamRole       TeamRole        `json:"team_role" gorm:"type:varchar(50);not null;default:'member'"`
	IsActive       bool            `json:"is_active" gorm:"default:true"`
	ExternalType   ExternalType    `json:"external_type" gorm:"type:varchar(50);default:'internal'"`
	Metadata       json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization        Organization         `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Group               *Group               `json:"group,omitempty" gorm:"foreignKey:GroupID;constraint:OnDelete:SET NULL"`
	Team                *Team                `json:"team,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:SET NULL"`
	DutySchedules       []DutySchedule       `json:"duty_schedules,omitempty" gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE"`
	OutageCallAssignees []OutageCallAssignee `json:"outage_call_assignees,omitempty" gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Member
func (Member) TableName() string {
	return "members"
}
