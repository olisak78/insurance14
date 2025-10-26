package models

import (
	"github.com/google/uuid"
)

// ProjectLandscape represents the many-to-many relationship between projects and landscapes
type ProjectLandscape struct {
	BaseModel
	ProjectID      uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index" validate:"required"`
	LandscapeID    uuid.UUID `json:"landscape_id" gorm:"type:uuid;not null;index" validate:"required"`
	LandscapeGroup string    `json:"landscape_group" gorm:"size:100"`
	SortOrder      int       `json:"sort_order" gorm:"default:0"`

	// Relationships
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Landscape Landscape `json:"landscape,omitempty" gorm:"foreignKey:LandscapeID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for ProjectLandscape
func (ProjectLandscape) TableName() string {
	return "project_landscapes"
}
