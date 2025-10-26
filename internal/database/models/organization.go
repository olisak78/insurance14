package models

import (
	"encoding/json"
)

// Organization represents the root entity for multi-tenancy
type Organization struct {
	BaseModel
	Name        string          `json:"name" gorm:"uniqueIndex;not null;size:100" validate:"required,min=1,max=100"`
	DisplayName string          `json:"display_name" gorm:"not null;size:200" validate:"required,max=200"`
	Domain      string          `json:"domain" gorm:"uniqueIndex;not null;size:100" validate:"required,max=100"`
	Description string          `json:"description" gorm:"type:text"`
	Metadata    json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Members    []Member    `json:"members,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Groups     []Group     `json:"groups,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Projects   []Project   `json:"projects,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Components []Component `json:"components,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Landscapes []Landscape `json:"landscapes,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}
