package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusInactive ProjectStatus = "inactive"
	ProjectStatusArchived ProjectStatus = "archived"
)

// ProjectType represents the type of a project
type ProjectType string

const (
	ProjectTypeApplication ProjectType = "application"
	ProjectTypeService     ProjectType = "service"
	ProjectTypeLibrary     ProjectType = "library"
	ProjectTypePlatform    ProjectType = "platform"
)

// Project represents a project in the developer portal with organization context
type Project struct {
	BaseModel
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	Name           string          `json:"name" gorm:"not null;size:200" validate:"required,min=1,max=200"`
	DisplayName    string          `json:"display_name" gorm:"not null;size:250" validate:"required,max=250"`
	Description    string          `json:"description" gorm:"type:text"`
	ProjectType    ProjectType     `json:"project_type" gorm:"type:varchar(50);default:'application'" validate:"required"`
	Status         ProjectStatus   `json:"status" gorm:"type:varchar(50);default:'active'" validate:"required"`
	SortOrder      int             `json:"sort_order" gorm:"default:0"`
	Metadata       json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization      Organization       `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	ProjectComponents []ProjectComponent `json:"project_components,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	ProjectLandscapes []ProjectLandscape `json:"project_landscapes,omitempty" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Project
func (Project) TableName() string {
	return "projects"
}
