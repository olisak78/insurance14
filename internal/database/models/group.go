package models

import (
	"github.com/google/uuid"
)

// Group represents a group of teams within an organization
type Group struct {
	BaseModel
	OrgID      uuid.UUID `json:"org_id" gorm:"type:uuid;not null;index" validate:"required"`
	Owner      string    `json:"owner" gorm:"not null;size:20" validate:"required,min=5,max=20"` // I/C/D user
	Email      string    `json:"email" gorm:"not null;size:50" validate:"required,min=5,max=50"` // DL
	PictureURL string    `json:"picture_url" gorm:"not null;size:200" validate:"required,min=5,max=200"`
}

// TableName returns the table name for Group
func (Group) TableName() string {
	return "groups"
}
