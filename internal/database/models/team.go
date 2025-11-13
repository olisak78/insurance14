package models

import (
	"github.com/google/uuid"
)

// Team represents a team in the developer portal with group context
type Team struct {
	BaseModel
	GroupID    uuid.UUID       `json:"group_id" gorm:"type:uuid;not null;index" validate:"required"`
	Owner      string          `json:"owner" gorm:"not null;size:20" validate:"required,min=5,max=20"` // I/C/D user
	Email      string          `json:"email" gorm:"not null;size:50" validate:"required,min=5,max=50"` // DL
	PictureURL string          `json:"picture_url" gorm:"not null;size:200" validate:"required,min=5,max=200"`

	// Relationships
	Documentations []Documentation `json:"documentations,omitempty" gorm:"foreignKey:TeamID"`
}

// TableName returns the table name for Team
func (Team) TableName() string {
	return "teams"
}
