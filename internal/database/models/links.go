package models

import (
	"github.com/google/uuid"
)

type Link struct {
	BaseModel
	Owner    uuid.UUID `json:"owner_id" gorm:"type:uuid;not null;index" validate:"required"`
	URL      string    `json:"url" gorm:"not null;size:1000" validate:"required,max=1000"`
	CategoryID uuid.UUID `json:"category_id" gorm:"type:uuid;not null;index" validate:"required"`
	Tags     string    `json:"tags" gorm:"size:200" validate:"max=200"` // comma seperated values
}

// TableName returns the table name for Link
func (Link) TableName() string {
	return "links"
}
