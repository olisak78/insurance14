package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel provides common fields for all models with UUID primary keys
type BaseModel struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	CreatedBy    string         `json:"created_by" gorm:"size:40" validate:"max=40"`
	UpdatedAt    time.Time      `json:"updated_at"`
	UpdatedBy    string         `json:"updated_by" gorm:"size:40" validate:"max=40"`
	Name         string         `json:"name" gorm:"size:40;not null" validate:"required,min=1,max=40"` // readable 'id'
	Title        string         `json:"title" gorm:"size:100;not null" validate:"required,min=1,max=100"` // AKA display name
	Description  string         `json:"description" gorm:"size:200" validate:"max=200"`
	Metadata     json.RawMessage `json:"metadata" gorm:"type:jsonb"`
}

// BeforeCreate sets the UUID if not already set
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}

// TODO: delete OldBaseModel
type OldBaseModel struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (base *OldBaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}
