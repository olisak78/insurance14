package models

import (
	"github.com/google/uuid"
)

// Component represents a software component/service
type Component struct {
	BaseModel
	ProjectID uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index" validate:"required"`
	OwnerID   uuid.UUID `json:"owner_id" gorm:"type:uuid;not null;index" validate:"required"`
}

// TableName returns the table name for Component
func (Component) TableName() string {
	return "components"
}
