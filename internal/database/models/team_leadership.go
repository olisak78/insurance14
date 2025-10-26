package models

import (
	"github.com/google/uuid"
)

// TeamLeadership represents a linking table between teams and members for leadership.
// Enforces a single leader per team via a unique index on TeamID.
type TeamLeadership struct {
	BaseModel
	TeamID   uuid.UUID `json:"team_id" gorm:"type:uuid;not null;uniqueIndex"`
	MemberID uuid.UUID `json:"member_id" gorm:"type:uuid;not null;index"`

	// Relationships
	Team   Team   `json:"team,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	Member Member `json:"member,omitempty" gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for TeamLeadership
func (TeamLeadership) TableName() string {
	return "team_leaderships"
}
