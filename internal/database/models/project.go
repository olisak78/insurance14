package models

type Project struct {
	BaseModel
}

// TableName returns the table name for Project
func (Project) TableName() string {
	return "projects"
}
