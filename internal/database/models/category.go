package models

type Category struct {
	BaseModel
	Icon  string `json:"icon" gorm:"not null;size:50" validate:"required,min=3,max=50"`
	Color string `json:"color" gorm:"not null;size:50" validate:"required,min=3,max=50"`
}

// TableName returns the table name for Category
func (Category) TableName() string {
	return "categories"
}
