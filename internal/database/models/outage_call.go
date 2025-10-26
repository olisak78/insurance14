package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OutageCallSeverity represents the severity of an outage call
type OutageCallSeverity string

const (
	OutageCallSeverityCritical OutageCallSeverity = "critical"
	OutageCallSeverityHigh     OutageCallSeverity = "high"
	OutageCallSeverityMedium   OutageCallSeverity = "medium"
	OutageCallSeverityLow      OutageCallSeverity = "low"
)

// OutageCallStatus represents the status of an outage call
type OutageCallStatus string

const (
	OutageCallStatusOpen       OutageCallStatus = "open"
	OutageCallStatusInProgress OutageCallStatus = "in_progress"
	OutageCallStatusResolved   OutageCallStatus = "resolved"
	OutageCallStatusCancelled  OutageCallStatus = "cancelled"
)

// OutageCall represents an outage or incident call
type OutageCall struct {
	BaseModel
	OrganizationID        uuid.UUID          `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	TeamID                uuid.UUID          `json:"team_id" gorm:"type:uuid;not null;index" validate:"required"`
	CallTimestamp         time.Time          `json:"call_timestamp" gorm:"not null" validate:"required"`
	Year                  int                `json:"year" gorm:"not null" validate:"required,min=2020,max=2100"`
	Title                 string             `json:"title" gorm:"not null;size:200" validate:"required,max=200"`
	Description           string             `json:"description" gorm:"type:text"`
	Severity              OutageCallSeverity `json:"severity" gorm:"type:varchar(50);default:'medium'" validate:"required"`
	Status                OutageCallStatus   `json:"status" gorm:"type:varchar(50);default:'open'" validate:"required"`
	ResolutionTimeMinutes int                `json:"resolution_time_minutes" gorm:""`
	ResolvedAt            *time.Time         `json:"resolved_at"`
	ExternalTicketID      string             `json:"external_ticket_id" gorm:"size:100"`
	Metadata              json.RawMessage    `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization        Organization         `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Team                Team                 `json:"team,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	OutageCallAssignees []OutageCallAssignee `json:"outage_call_assignees,omitempty" gorm:"foreignKey:OutageCallID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for OutageCall
func (OutageCall) TableName() string {
	return "outage_calls"
}
