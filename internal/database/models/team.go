package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// TeamStatus represents the status of a team
type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
	TeamStatusArchived TeamStatus = "archived"
)

// Team represents a team in the developer portal with group context
type Team struct {
	BaseModel
	GroupID     uuid.UUID       `json:"group_id" gorm:"type:uuid;not null;index" validate:"required"`
	Name        string          `json:"name" gorm:"uniqueIndex:idx_team_name;not null;size:100" validate:"required,min=1,max=100"`
	DisplayName string          `json:"display_name" gorm:"not null;size:200" validate:"required,max=200"`
	Description string          `json:"description" gorm:"type:text"`
	Status      TeamStatus      `json:"status" gorm:"type:varchar(50);default:'active'" validate:"required"`
	Links       json.RawMessage `json:"links" gorm:"type:jsonb"`
	Metadata    json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Group                   Group                    `json:"group,omitempty" gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	Members                 []Member                 `json:"members,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:SET NULL"`
	TeamComponentOwnerships []TeamComponentOwnership `json:"team_component_ownerships,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	DutySchedules           []DutySchedule           `json:"duty_schedules,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	OutageCalls             []OutageCall             `json:"outage_calls,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Team
func (Team) TableName() string {
	return "teams"
}
