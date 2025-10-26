package models

import (
	"github.com/google/uuid"
)

// TeamComponentOwnership represents the ownership relationship between teams and components
type TeamComponentOwnership struct {
	BaseModel
	TeamID        uuid.UUID     `json:"team_id" gorm:"type:uuid;not null;index" validate:"required"`
	ComponentID   uuid.UUID     `json:"component_id" gorm:"type:uuid;not null;index" validate:"required"`
	OwnershipType OwnershipType `json:"ownership_type" gorm:"type:varchar(50);default:'primary'" validate:"required"`

	// Relationships
	Team      Team      `json:"team,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	Component Component `json:"component,omitempty" gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for TeamComponentOwnership
func (TeamComponentOwnership) TableName() string {
	return "team_component_ownership"
}
