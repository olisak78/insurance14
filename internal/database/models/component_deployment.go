package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ComponentDeployment represents a component deployment instance
type ComponentDeployment struct {
	BaseModel
	ComponentID     uuid.UUID       `json:"component_id" gorm:"type:uuid;not null;index" validate:"required"`
	LandscapeID     uuid.UUID       `json:"landscape_id" gorm:"type:uuid;not null;index" validate:"required"`
	Version         string          `json:"version" gorm:"size:100"`
	GitCommitID     string          `json:"git_commit_id" gorm:"size:100"`
	GitCommitTime   *time.Time      `json:"git_commit_time"`
	BuildTime       *time.Time      `json:"build_time"`
	BuildProperties json.RawMessage `json:"build_properties" gorm:"type:jsonb"`
	GitProperties   json.RawMessage `json:"git_properties" gorm:"type:jsonb"`
	IsActive        bool            `json:"is_active" gorm:"not null"`

	DeployedAt *time.Time `json:"deployed_at"`

	// Relationships
	Component Component `json:"component,omitempty" gorm:"foreignKey:ComponentID;constraint:OnDelete:CASCADE"`
	Landscape Landscape `json:"landscape,omitempty" gorm:"foreignKey:LandscapeID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for ComponentDeployment
func (ComponentDeployment) TableName() string {
	return "component_deployments"
}
