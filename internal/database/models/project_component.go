package models

import (
	"github.com/google/uuid"
)

// OwnershipType represents the type of ownership relationship
type OwnershipType string

const (
	OwnershipTypePrimary     OwnershipType = "primary"
	OwnershipTypeSecondary   OwnershipType = "secondary"
	OwnershipTypeContributor OwnershipType = "contributor"
	OwnershipTypeConsumer    OwnershipType = "consumer"
)

// ProjectComponent represents the many-to-many relationship between projects and components
type ProjectComponent struct {
	BaseModel
	ProjectID     uuid.UUID     `json:"project_id" gorm:"type:uuid;not null;index" validate:"required"`
	ComponentID   uuid.UUID     `json:"component_id" gorm:"type:uuid;not null;index" validate:"required"`
	OwnershipType OwnershipType `json:"ownership_type" gorm:"type:varchar(50);default:'consumer'" validate:"required"`
	SortOrder     int           `json:"sort_order" gorm:"default:0"`

	// Relationships
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Component Component `json:"component,omitempty" gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for ProjectComponent
func (ProjectComponent) TableName() string {
	return "project_components"
}
