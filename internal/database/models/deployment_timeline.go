package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DeploymentTimeline represents deployment timelines for landscapes
type DeploymentTimeline struct {
	BaseModel
	OrganizationID  uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	LandscapeID     uuid.UUID       `json:"landscape_id" gorm:"type:uuid;not null;index" validate:"required"`
	TimelineCode    string          `json:"timeline_code" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	TimelineName    string          `json:"timeline_name" gorm:"type:varchar(200);not null" validate:"required,max=200"`
	ScheduledDate   time.Time       `json:"scheduled_date" gorm:"type:date;not null" validate:"required"`
	IsCompleted     bool            `json:"is_completed" gorm:"default:false"`
	StatusIndicator string          `json:"status_indicator" gorm:"type:text"`
	Metadata        json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Landscape    Landscape    `json:"landscape,omitempty" gorm:"foreignKey:LandscapeID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for DeploymentTimeline
func (DeploymentTimeline) TableName() string {
	return "deployment_timelines"
}
