package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DutySchedule represents duty schedules for teams and members
type DutySchedule struct {
	BaseModel
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index" validate:"required"`
	TeamID         uuid.UUID       `json:"team_id" gorm:"type:uuid;not null;index" validate:"required"`
	MemberID       uuid.UUID       `json:"member_id" gorm:"type:uuid;not null;index" validate:"required"`
	ScheduleType   ScheduleType    `json:"schedule_type" gorm:"type:varchar(50);not null" validate:"required"`
	Year           int             `json:"year" gorm:"not null" validate:"required,min=2020,max=2100"`
	StartDate      time.Time       `json:"start_date" gorm:"type:date;not null" validate:"required"`
	EndDate        time.Time       `json:"end_date" gorm:"type:date;not null" validate:"required"`
	ShiftType      ShiftType       `json:"shift_type" gorm:"type:varchar(50)"`
	WasCalled      bool            `json:"was_called" gorm:"default:false"`
	Notes          string          `json:"notes" gorm:"type:text"`
	Metadata       json.RawMessage `json:"metadata" gorm:"type:jsonb"`

	// Relationships
	Organization Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Team         Team         `json:"team,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`
	Member       Member       `json:"member,omitempty" gorm:"foreignKey:MemberID;constraint:OnDelete:CASCADE"`
}

// TableName returns the table name for DutySchedule
func (DutySchedule) TableName() string {
	return "duty_schedules"
}
