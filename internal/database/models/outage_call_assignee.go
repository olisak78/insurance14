package models

import (
	"time"

	"github.com/google/uuid"
)

// AssigneeRole represents the role of an assignee in an outage call
type AssigneeRole string

const (
	AssigneeRolePrimary   AssigneeRole = "primary"
	AssigneeRoleSecondary AssigneeRole = "secondary"
	AssigneeRoleObserver  AssigneeRole = "observer"
)

// OutageCallAssignee represents the assignment of members to outage calls
type OutageCallAssignee struct {
	BaseModel
	OutageCallID uuid.UUID    `json:"outage_call_id" gorm:"type:uuid;not null;index" validate:"required"`
	MemberID     uuid.UUID    `json:"member_id" gorm:"type:uuid;not null;index" validate:"required"`
	Role         AssigneeRole `json:"role" gorm:"type:varchar(50);default:'primary'" validate:"required"`
	AssignedAt   time.Time    `json:"assigned_at" gorm:"not null" validate:"required"`
	IsActive     bool         `json:"is_active" gorm:"default:true"`

	// Relationships
	OutageCall OutageCall `json:"outage_call,omitempty" gorm:"foreignKey:OutageCallID;constraint:OnDelete:CASCADE"`
	Member     Member     `json:"member,omitempty" gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for OutageCallAssignee
func (OutageCallAssignee) TableName() string {
	return "outage_call_assignees"
}
