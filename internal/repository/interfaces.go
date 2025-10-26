package repository

import (
	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=../mocks/repository_mocks.go -package=mocks

// OrganizationRepositoryInterface defines the interface for organization repository operations
type OrganizationRepositoryInterface interface {
	Create(org *models.Organization) error
	GetByID(id uuid.UUID) (*models.Organization, error)
	GetByName(name string) (*models.Organization, error)
	GetByDomain(domain string) (*models.Organization, error)
	GetAll(limit, offset int) ([]models.Organization, int64, error)
	Update(org *models.Organization) error
	Delete(id uuid.UUID) error
	GetWithMembers(id uuid.UUID) (*models.Organization, error)
	GetWithGroups(id uuid.UUID) (*models.Organization, error)
	GetWithProjects(id uuid.UUID) (*models.Organization, error)
	GetWithComponents(id uuid.UUID) (*models.Organization, error)
	GetWithLandscapes(id uuid.UUID) (*models.Organization, error)
	GetWithAllRelations(id uuid.UUID) (*models.Organization, error)
}

// MemberRepositoryInterface defines the interface for member repository operations
type MemberRepositoryInterface interface {
	Create(member *models.Member) error
	GetByID(id uuid.UUID) (*models.Member, error)
	GetByEmail(email string) (*models.Member, error)
	GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Member, int64, error)
	GetWithOrganization(id uuid.UUID) (*models.Member, error)
	SearchByOrganization(orgID uuid.UUID, query string, limit, offset int) ([]models.Member, int64, error)
	GetActiveByOrganization(orgID uuid.UUID, limit, offset int) ([]models.Member, int64, error)
	Update(member *models.Member) error
	Delete(id uuid.UUID) error
}

// GroupRepositoryInterface defines the interface for group repository operations
type GroupRepositoryInterface interface {
	Create(group *models.Group) error
	GetByID(id uuid.UUID) (*models.Group, error)
	GetByName(orgID uuid.UUID, name string) (*models.Group, error)
	GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Group, int64, error)
	Search(organizationID uuid.UUID, query string, limit, offset int) ([]models.Group, int64, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	GetWithTeams(id uuid.UUID) (*models.Group, error)
	GetWithOrganization(id uuid.UUID) (*models.Group, error)
}

// TeamRepositoryInterface defines the interface for team repository operations
type TeamRepositoryInterface interface {
	Create(team *models.Team) error
	GetByID(id uuid.UUID) (*models.Team, error)
	GetByName(groupID uuid.UUID, name string) (*models.Team, error)
	GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Team, int64, error)
	GetByGroupID(groupID uuid.UUID, limit, offset int) ([]models.Team, int64, error)
	GetAll() ([]models.Team, error)
	GetWithMembers(id uuid.UUID) (*models.Team, error)
	Update(team *models.Team) error
	Delete(id uuid.UUID) error
}

// ProjectRepositoryInterface defines the interface for project repository operations
type ProjectRepositoryInterface interface {
	Create(project *models.Project) error
	GetByID(id uuid.UUID) (*models.Project, error)
	GetByName(name string, orgID uuid.UUID) (*models.Project, error)
	GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Project, int64, error)
	GetByStatus(status models.ProjectStatus, limit, offset int) ([]models.Project, int64, error)
	GetActiveProjects(limit, offset int) ([]models.Project, int64, error)
	Update(project *models.Project) error
	Delete(id uuid.UUID) error
}

// ComponentRepositoryInterface defines the interface for component repository operations
type ComponentRepositoryInterface interface {
	Create(component *models.Component) error
	GetByID(id uuid.UUID) (*models.Component, error)
	GetByName(name string, orgID uuid.UUID) (*models.Component, error)
	GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Component, int64, error)
	GetByStatus(status models.ComponentStatus, limit, offset int) ([]models.Component, int64, error)
	GetByType(componentType models.ComponentType, limit, offset int) ([]models.Component, int64, error)
	GetActiveComponents(limit, offset int) ([]models.Component, int64, error)
	Update(component *models.Component) error
	Delete(id uuid.UUID) error
}

// LandscapeRepositoryInterface defines the interface for landscape repository operations
type LandscapeRepositoryInterface interface {
	Create(landscape *models.Landscape) error
	GetByID(id uuid.UUID) (*models.Landscape, error)
	GetByName(name string, orgID uuid.UUID) (*models.Landscape, error)
	GetByOrganizationID(orgID uuid.UUID, limit, offset int) ([]models.Landscape, int64, error)
	GetByStatus(status models.LandscapeStatus, limit, offset int) ([]models.Landscape, int64, error)
	GetActiveLandscapes(limit, offset int) ([]models.Landscape, int64, error)
	Update(landscape *models.Landscape) error
	Delete(id uuid.UUID) error
}

// ComponentDeploymentRepositoryInterface defines the interface for component deployment repository operations
type ComponentDeploymentRepositoryInterface interface {
	Create(deployment *models.ComponentDeployment) error
	GetByID(id uuid.UUID) (*models.ComponentDeployment, error)
	GetByComponentID(componentID uuid.UUID, limit, offset int) ([]models.ComponentDeployment, int64, error)
	GetByLandscapeID(landscapeID uuid.UUID, limit, offset int) ([]models.ComponentDeployment, int64, error)
	GetByComponentAndLandscape(componentID, landscapeID uuid.UUID) (*models.ComponentDeployment, error)
	GetByActiveStatus(isActive bool, limit, offset int) ([]models.ComponentDeployment, int64, error)
	Update(deployment *models.ComponentDeployment) error
	Delete(id uuid.UUID) error
}
