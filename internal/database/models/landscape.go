package models

import (
	"github.com/google/uuid"
)

type Landscape struct {
	BaseModel
	ProjectID   uuid.UUID `json:"project_id" gorm:"type:uuid;not null;index" validate:"required"`
	Domain      string    `json:"domain" gorm:"not null;size:200" validate:"required,max=200"`
	Environment string    `json:"environment" gorm:"not null;size:20" validate:"required,max=20"`
}

func (Landscape) TableName() string {
	return "landscapes"
}
