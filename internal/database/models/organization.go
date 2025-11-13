package models

// Organization represents the root entity for multi-tenancy
type Organization struct {
	BaseModel
	Owner string `json:"owner" gorm:"not null;size:20" validate:"required,min=5,max=20"` // I/C/D user
	Email string `json:"email" gorm:"not null;size:50" validate:"required,min=5,max=50"` // DL
}

// TableName returns the table name for Organization
func (Organization) TableName() string {
	return "organizations"
}
