package testutils

import (
	"time"

	"developer-portal-backend/internal/database/models"

	"github.com/google/uuid"
)

// OrganizationFactory provides methods to create test Organization data
type OrganizationFactory struct{}

// NewOrganizationFactory creates a new OrganizationFactory
func NewOrganizationFactory() *OrganizationFactory {
	return &OrganizationFactory{}
}

// Create creates a test Organization with default values
func (f *OrganizationFactory) Create() *models.Organization {
	return &models.Organization{
		BaseModel: models.BaseModel{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "test-org",
			Title:       "Test Organization Display Name",
			Description: "A test organization for testing purposes",
		},
		Owner: "I00001",
		Email: "org@test.com",
	}
}

// WithName sets a custom name for the organization
func (f *OrganizationFactory) WithName(name string) *models.Organization {
	org := f.Create()
	org.Name = name
	org.Title = name + " Display Name"
	return org
}

// WithDomain sets a custom domain for the organization
func (f *OrganizationFactory) WithDomain(domain string) *models.Organization {
	org := f.Create()
	// Domain field removed in new model; approximate by setting email for tests
	org.Email = domain
	return org
}

// UserFactory (alias to User) provides methods to create test User data
type UserFactory struct{}

// NewUserFactory creates a new UserFactory
func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

// Create creates a test User with default values
func (f *UserFactory) Create() *models.User {
	id := uuid.New()
	// Generate unique short user id using part of UUID to avoid conflicts
	userID := "I" + id.String()[:6]

	return &models.User{
		TeamID:     nil,
		UserID:     userID,
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@test.com",
		Mobile:     "+1-555-0123",
		TeamDomain: models.TeamDomainDeveloper,
		TeamRole:   models.TeamRoleMember,
	}
}

// WithOrganization is retained for compatibility; returns a default user
func (f *UserFactory) WithOrganization(orgID uuid.UUID) *models.User {
	return f.Create()
}

// WithTeam sets the team ID for the member (user)
func (f *UserFactory) WithTeam(teamID uuid.UUID) *models.User {
	member := f.Create()
	member.TeamID = &teamID
	return member
}

// WithEmail sets a custom email for the member
func (f *UserFactory) WithEmail(email string) *models.User {
	member := f.Create()
	member.Email = email
	return member
}

// WithRole sets a custom role for the member
func (f *UserFactory) WithRole(role models.TeamRole) *models.User {
	member := f.Create()
	member.TeamRole = role
	return member
}

// GroupFactory provides methods to create test Group data
type GroupFactory struct{}

// NewGroupFactory creates a new GroupFactory
func NewGroupFactory() *GroupFactory {
	return &GroupFactory{}
}

// Create creates a test Group with default values
func (f *GroupFactory) Create() *models.Group {
	return &models.Group{
		BaseModel: models.BaseModel{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "test-group",
			Title:       "Test Group",
			Description: "A test group for testing purposes",
		},
		OrgID:      uuid.New(),
		Owner:      "I00001",
		Email:      "group@test.com",
		PictureURL: "https://example.com/picture.png",
	}
}

// WithOrganization sets the organization ID for the group
func (f *GroupFactory) WithOrganization(orgID uuid.UUID) *models.Group {
	group := f.Create()
	group.OrgID = orgID
	return group
}

// WithName sets a custom name for the group
func (f *GroupFactory) WithName(name string) *models.Group {
	group := f.Create()
	group.Name = name
	group.Title = name + " Group"
	return group
}

// TeamFactory provides methods to create test Team data
type TeamFactory struct{}

// NewTeamFactory creates a new TeamFactory
func NewTeamFactory() *TeamFactory {
	return &TeamFactory{}
}

// Create creates a test Team with default values
func (f *TeamFactory) Create() *models.Team {
	return &models.Team{
		BaseModel: models.BaseModel{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "test-team",
			Title:       "Test Team",
			Description: "A test team for testing purposes",
		},
		GroupID:    uuid.New(),
		Owner:      "I00001",
		Email:      "team@test.com",
		PictureURL: "https://example.com/picture.png",
	}
}

// WithGroup sets the group ID for the team
func (f *TeamFactory) WithGroup(groupID uuid.UUID) *models.Team {
	team := f.Create()
	team.GroupID = groupID
	return team
}

// WithOrganization creates a team with a group in the specified organization
func (f *TeamFactory) WithOrganization(orgID uuid.UUID) *models.Team {
	// Create a default group for the organization
	groupFactory := NewGroupFactory()
	group := groupFactory.WithOrganization(orgID)

	team := f.Create()
	team.GroupID = group.ID
	return team
}

// WithName sets a custom name for the team
func (f *TeamFactory) WithName(name string) *models.Team {
	team := f.Create()
	team.Name = name
	team.Title = name + " Team"
	return team
}

// ProjectFactory provides methods to create test Project data
type ProjectFactory struct{}

// NewProjectFactory creates a new ProjectFactory
func NewProjectFactory() *ProjectFactory {
	return &ProjectFactory{}
}

// Create creates a test Project with default values
func (f *ProjectFactory) Create() *models.Project {
	return &models.Project{
		BaseModel: models.BaseModel{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "test-project",
			Title:       "Test Project",
			Description: "A test project for testing purposes",
		},
	}
}

// WithName sets a custom name for the project
func (f *ProjectFactory) WithName(name string) *models.Project {
	project := f.Create()
	project.Name = name
	project.Title = name + " Project"
	return project
}

// ComponentFactory provides methods to create test Component data
type ComponentFactory struct{}

// NewComponentFactory creates a new ComponentFactory
func NewComponentFactory() *ComponentFactory {
	return &ComponentFactory{}
}

// Create creates a test Component with default values
func (f *ComponentFactory) Create() *models.Component {
	return &models.Component{
		BaseModel: models.BaseModel{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "test-component",
			Title:       "Test Component",
			Description: "A test component for testing purposes",
		},
		ProjectID: uuid.New(),
		OwnerID:   uuid.New(),
	}
}

// WithName sets a custom name for the component
func (f *ComponentFactory) WithName(name string) *models.Component {
	component := f.Create()
	component.Name = name
	component.Title = name + " Component"
	return component
}

// LandscapeFactory provides methods to create test Landscape data
type LandscapeFactory struct{}

// NewLandscapeFactory creates a new LandscapeFactory
func NewLandscapeFactory() *LandscapeFactory {
	return &LandscapeFactory{}
}

// Create creates a test Landscape with default values
func (f *LandscapeFactory) Create() *models.Landscape {
	return &models.Landscape{
		BaseModel: models.BaseModel{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "test-landscape",
			Title:       "Test Landscape",
			Description: "A test landscape for testing purposes",
		},
		ProjectID:   uuid.New(),
		Domain:      "example.com",
		Environment: "development",
	}
}

// WithOrganization removed; no organization on Landscape in new model, keep for compatibility
func (f *LandscapeFactory) WithOrganization(orgID uuid.UUID) *models.Landscape {
	return f.Create()
}

// WithName sets a custom name for the landscape
func (f *LandscapeFactory) WithName(name string) *models.Landscape {
	landscape := f.Create()
	// Truncate name to 40 chars (BaseModel.Name has max length 40)
	if len(name) > 40 {
		name = name[:40]
	}
	landscape.Name = name
	landscape.Title = name + " Landscape"
	return landscape
}

// ComponentDeploymentFactory provides methods to create test ComponentDeployment data
type ComponentDeploymentFactory struct{}

// NewComponentDeploymentFactory creates a new ComponentDeploymentFactory
func NewComponentDeploymentFactory() *ComponentDeploymentFactory {
	return &ComponentDeploymentFactory{}
}

// FactorySet provides access to all factories
type FactorySet struct {
	Organization        *OrganizationFactory
	User                *UserFactory
	Group               *GroupFactory
	Team                *TeamFactory
	Project             *ProjectFactory
	Component           *ComponentFactory
	Landscape           *LandscapeFactory
	ComponentDeployment *ComponentDeploymentFactory
}

// NewFactorySet creates a new FactorySet with all factories initialized
func NewFactorySet() *FactorySet {
	return &FactorySet{
		Organization:        NewOrganizationFactory(),
		User:                NewUserFactory(),
		Group:               NewGroupFactory(),
		Team:                NewTeamFactory(),
		Project:             NewProjectFactory(),
		Component:           NewComponentFactory(),
		Landscape:           NewLandscapeFactory(),
		ComponentDeployment: NewComponentDeploymentFactory(),
	}
}

// CreateFullOrganizationHierarchy creates a complete organization with team, user, project, component, and landscape
func (fs *FactorySet) CreateFullOrganizationHierarchy() (*models.Organization, *models.Team, *models.User, *models.Project, *models.Component, *models.Landscape) {
	// Create organization
	org := fs.Organization.Create()

	// Create team under org
	team := fs.Team.WithOrganization(org.ID)

	// Create member and assign to team
	member := fs.User.WithOrganization(org.ID)
	member.TeamID = &team.ID

	// Create project
	project := fs.Project.Create()

	// Create component owned by team and linked to project
	component := fs.Component.Create()
	component.OwnerID = team.ID
	component.ProjectID = project.ID

	// Create landscape linked to project
	landscape := fs.Landscape.Create()
	landscape.ProjectID = project.ID

	return org, team, member, project, component, landscape
}
