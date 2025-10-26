package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// LandscapeType represents the type of a landscape
type LandscapeType string

const (
	LandscapeTypeDevelopment LandscapeType = "development"
	LandscapeTypeStaging     LandscapeType = "staging"
	LandscapeTypeProduction  LandscapeType = "production"
	LandscapeTypeTesting     LandscapeType = "testing"
	LandscapeTypePreview     LandscapeType = "preview"
)

// LandscapeStatus represents the status of a landscape
type LandscapeStatus string

const (
	LandscapeStatusActive      LandscapeStatus = "active"
	LandscapeStatusInactive    LandscapeStatus = "inactive"
	LandscapeStatusMaintenance LandscapeStatus = "maintenance"
	LandscapeStatusRetired     LandscapeStatus = "retired"
)

// DeploymentStatus represents the deployment status of a landscape
type DeploymentStatus string

const (
	DeploymentStatusHealthy   DeploymentStatus = "healthy"
	DeploymentStatusDegraded  DeploymentStatus = "degraded"
	DeploymentStatusUnhealthy DeploymentStatus = "unhealthy"
	DeploymentStatusUnknown   DeploymentStatus = "unknown"
)

// Landscape represents deployment environments
type Landscape struct {
	BaseModel
	OrganizationID   uuid.UUID        `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	Name             string           `json:"name" gorm:"uniqueIndex:idx_org_landscape_name,composite:organization_id;not null;size:200" validate:"required,min=1,max=200"`
	DisplayName      string           `json:"display_name" gorm:"not null;size:250" validate:"required,max=250"`
	Description      string           `json:"description" gorm:"type:text"`
	LandscapeType    LandscapeType    `json:"landscape_type" gorm:"type:varchar(50);default:'development'" validate:"required"`
	EnvironmentGroup string           `json:"environment_group" gorm:"size:100"`
	Status           LandscapeStatus  `json:"status" gorm:"type:varchar(50);default:'active'" validate:"required"`
	DeploymentStatus DeploymentStatus `json:"deployment_status" gorm:"type:varchar(50);default:'unknown'"`
	GitHubConfigURL  string           `json:"github_config_url" gorm:"size:500"`
	AWSAccountID     string           `json:"aws_account_id" gorm:"size:50"`
	CAMProfileURL    string           `json:"cam_profile_url" gorm:"size:500"`
	SortOrder        int              `json:"sort_order" gorm:"default:0"`
	Metadata         json.RawMessage  `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization         Organization          `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	ComponentDeployments []ComponentDeployment `json:"component_deployments,omitempty" gorm:"foreignKey:LandscapeID;constraint:OnDelete:CASCADE"`
	ProjectLandscapes    []ProjectLandscape    `json:"project_landscapes,omitempty" gorm:"foreignKey:LandscapeID;constraint:OnDelete:CASCADE"`
	DeploymentTimelines  []DeploymentTimeline  `json:"deployment_timelines,omitempty" gorm:"foreignKey:LandscapeID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for Landscape
func (Landscape) TableName() string {
	return "landscapes"
}
