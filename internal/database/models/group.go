package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Group represents a group of teams within an organization
type Group struct {
	BaseModel
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	Name           string          `json:"name" gorm:"uniqueIndex:idx_org_group_name,composite:organization_id;not null;size:100" validate:"required,min=1,max=100"`
	DisplayName    string          `json:"display_name" gorm:"not null;size:200" validate:"required,max=200"`
	Description    string          `json:"description" gorm:"type:text"`
	Metadata       json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Teams        []Team       `json:"teams,omitempty" gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Group
func (Group) TableName() string {
	return "groups"
}
