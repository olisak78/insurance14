package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ComponentType represents the type of a component
type ComponentType string

const (
	ComponentTypeService     ComponentType = "service"
	ComponentTypeLibrary     ComponentType = "library"
	ComponentTypeApplication ComponentType = "application"
	ComponentTypeDatabase    ComponentType = "database"
	ComponentTypeAPI         ComponentType = "api"
)

// ComponentStatus represents the status of a component
type ComponentStatus string

const (
	ComponentStatusActive      ComponentStatus = "active"
	ComponentStatusInactive    ComponentStatus = "inactive"
	ComponentStatusDeprecated  ComponentStatus = "deprecated"
	ComponentStatusMaintenance ComponentStatus = "maintenance"
)

// Component represents a software component/service
type Component struct {
	BaseModel
	OrganizationID   uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	Name             string          `json:"name" gorm:"uniqueIndex:idx_org_component_name,composite:organization_id;not null;size:200" validate:"required,min=1,max=200"`
	DisplayName      string          `json:"display_name" gorm:"not null;size:250" validate:"required,max=250"`
	Description      string          `json:"description" gorm:"type:text"`
	ComponentType    ComponentType   `json:"component_type" gorm:"type:varchar(50);default:'service'" validate:"required"`
	Status           ComponentStatus `json:"status" gorm:"type:varchar(50);default:'active'" validate:"required"`
	GroupName        string          `json:"group_name" gorm:"size:100"`
	ArtifactName     string          `json:"artifact_name" gorm:"size:100"`
	GitRepositoryURL string          `json:"git_repository_url" gorm:"size:500"`
	DocumentationURL string          `json:"documentation_url" gorm:"size:500"`
	Links            json.RawMessage `json:"links" gorm:"type:jsonb"`
	Metadata         json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization            Organization             `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	ComponentDeployments    []ComponentDeployment    `json:"component_deployments,omitempty" gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
	ProjectComponents       []ProjectComponent       `json:"project_components,omitempty" gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
	TeamComponentOwnerships []TeamComponentOwnership `json:"team_component_ownerships,omitempty" gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Component
func (Component) TableName() string {
	return "components"
}
