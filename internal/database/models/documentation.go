package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Documentation represents a GitHub documentation endpoint for a team
// Example URL: https://github.tools.sap/cfs-platform-engineering/cfs-platform-docs/tree/main/docs/coe
// Split into: Owner (org), Repo, Branch, DocsPath
type Documentation struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy string         `json:"created_by" gorm:"size:40" validate:"max=40"`
	UpdatedAt time.Time      `json:"updated_at"`
	UpdatedBy string         `json:"updated_by" gorm:"size:40" validate:"max=40"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Foreign key to Team (one-to-many: team has many documentations)
	TeamID uuid.UUID `json:"team_id" gorm:"type:uuid;not null;index" validate:"required"`
	Team   *Team     `json:"team,omitempty" gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE"`

	// GitHub documentation fields
	Owner    string `json:"owner" gorm:"size:100;not null" validate:"required,min=1,max=100"`       // Organization/user (e.g., "cfs-platform-engineering")
	Repo     string `json:"repo" gorm:"size:100;not null" validate:"required,min=1,max=100"`        // Repository name (e.g., "cfs-platform-docs")
	Branch   string `json:"branch" gorm:"size:100;not null;default:main" validate:"required,min=1,max=100"` // Branch name (e.g., "main")
	DocsPath string `json:"docs_path" gorm:"size:500;not null" validate:"required,min=1,max=500"`  // Path within repo (e.g., "docs/coe")

	// Display information
	Title       string `json:"title" gorm:"size:100;not null" validate:"required,min=1,max=100"`     // Display name (e.g., "COE Documentation")
	Description string `json:"description" gorm:"size:200" validate:"max=200"`                       // Optional description
}

// TableName returns the table name for Documentation
func (Documentation) TableName() string {
	return "documentations"
}

// BeforeCreate sets the UUID if not already set
func (doc *Documentation) BeforeCreate(tx *gorm.DB) error {
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	return nil
}

// GetFullURL constructs the full GitHub URL from components
// Returns: https://github.tools.sap/{owner}/{repo}/tree/{branch}/{docs_path}
func (doc *Documentation) GetFullURL(baseURL string) string {
	return baseURL + "/" + doc.Owner + "/" + doc.Repo + "/tree/" + doc.Branch + "/" + doc.DocsPath
}
