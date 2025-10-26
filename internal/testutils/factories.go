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
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        "Test Organization",
		DisplayName: "Test Organization Display Name",
		Description: "A test organization for testing purposes",
		Domain:      "test.com",
		Metadata:    nil,
	}
}

// WithName sets a custom name for the organization
func (f *OrganizationFactory) WithName(name string) *models.Organization {
	org := f.Create()
	org.Name = name
	org.DisplayName = name + " Display Name"
	return org
}

// WithDomain sets a custom domain for the organization
func (f *OrganizationFactory) WithDomain(domain string) *models.Organization {
	org := f.Create()
	org.Domain = domain
	return org
}

// MemberFactory provides methods to create test Member data
type MemberFactory struct{}

// NewMemberFactory creates a new MemberFactory
func NewMemberFactory() *MemberFactory {
	return &MemberFactory{}
}

// Create creates a test Member with default values
func (f *MemberFactory) Create() *models.Member {
	id := uuid.New()
	// Generate unique IUser using part of UUID to avoid conflicts
	iUser := "I" + id.String()[:6]

	return &models.Member{
		BaseModel: models.BaseModel{
			ID:        id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OrganizationID: uuid.New(),
		TeamID:         nil,
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john.doe@test.com",
		PhoneNumber:    "+1-555-0123",
		IUser:          iUser,
		Role:           models.MemberRoleDeveloper,
		IsActive:       true,
		ExternalType:   models.ExternalTypeInternal,
		Metadata:       nil,
	}
}

// WithOrganization sets the organization ID for the member
func (f *MemberFactory) WithOrganization(orgID uuid.UUID) *models.Member {
	member := f.Create()
	member.OrganizationID = orgID
	return member
}

// WithTeam sets the team ID for the member
func (f *MemberFactory) WithTeam(teamID uuid.UUID) *models.Member {
	member := f.Create()
	member.TeamID = &teamID
	return member
}

// WithEmail sets a custom email for the member
func (f *MemberFactory) WithEmail(email string) *models.Member {
	member := f.Create()
	member.Email = email
	return member
}

// WithRole sets a custom role for the member
func (f *MemberFactory) WithRole(role models.MemberRole) *models.Member {
	member := f.Create()
	member.Role = role
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
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OrganizationID: uuid.New(),
		Name:           "test-group",
		DisplayName:    "Test Group",
		Description:    "A test group for testing purposes",
		Metadata:       nil,
	}
}

// WithOrganization sets the organization ID for the group
func (f *GroupFactory) WithOrganization(orgID uuid.UUID) *models.Group {
	group := f.Create()
	group.OrganizationID = orgID
	return group
}

// WithName sets a custom name for the group
func (f *GroupFactory) WithName(name string) *models.Group {
	group := f.Create()
	group.Name = name
	group.DisplayName = name + " Group"
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
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		GroupID:     uuid.New(),
		Name:        "test-team",
		DisplayName: "Test Team",
		Description: "A test team for testing purposes",
		Status:      models.TeamStatusActive,
		Links:       nil,
		Metadata:    nil,
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
	team.DisplayName = name + " Team"
	return team
}

// WithTeamLead is deprecated; leadership is handled via the linking table (TeamLeadership).
// This is now a no-op to preserve test compatibility without touching removed fields.
func (f *TeamFactory) WithTeamLead(teamLeadID uuid.UUID) *models.Team {
	team := f.Create()
	// deprecated: no direct leader assignment on team
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
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OrganizationID: uuid.New(),
		Name:           "test-project",
		DisplayName:    "Test Project",
		Description:    "A test project for testing purposes",
		ProjectType:    models.ProjectTypeApplication,
		Status:         models.ProjectStatusActive,
		SortOrder:      0,
		Metadata:       nil,
	}
}

// WithOrganization sets the organization ID for the project
func (f *ProjectFactory) WithOrganization(orgID uuid.UUID) *models.Project {
	project := f.Create()
	project.OrganizationID = orgID
	return project
}

// WithName sets a custom name for the project
func (f *ProjectFactory) WithName(name string) *models.Project {
	project := f.Create()
	project.Name = name
	project.DisplayName = name + " Project"
	return project
}

// WithStatus sets a custom status for the project
func (f *ProjectFactory) WithStatus(status models.ProjectStatus) *models.Project {
	project := f.Create()
	project.Status = status
	return project
}

// WithType sets a custom type for the project
func (f *ProjectFactory) WithType(projectType models.ProjectType) *models.Project {
	project := f.Create()
	project.ProjectType = projectType
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
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OrganizationID:   uuid.New(),
		Name:             "test-component",
		DisplayName:      "Test Component",
		Description:      "A test component for testing purposes",
		ComponentType:    models.ComponentTypeService,
		Status:           models.ComponentStatusActive,
		GroupName:        "test-group",
		ArtifactName:     "test-artifact",
		GitRepositoryURL: "https://github.com/test/test-component",
		DocumentationURL: "https://docs.test.com/test-component",
		Links:            nil,
		Metadata:         nil,
	}
}

// WithOrganization sets the organization ID for the component
func (f *ComponentFactory) WithOrganization(orgID uuid.UUID) *models.Component {
	component := f.Create()
	component.OrganizationID = orgID
	return component
}

// WithName sets a custom name for the component
func (f *ComponentFactory) WithName(name string) *models.Component {
	component := f.Create()
	component.Name = name
	component.DisplayName = name + " Component"
	return component
}

// WithType sets a custom type for the component
func (f *ComponentFactory) WithType(componentType models.ComponentType) *models.Component {
	component := f.Create()
	component.ComponentType = componentType
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
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OrganizationID:   uuid.New(),
		Name:             "test-landscape",
		DisplayName:      "Test Landscape",
		Description:      "A test landscape for testing purposes",
		LandscapeType:    models.LandscapeTypeDevelopment,
		EnvironmentGroup: "test-group",
		Status:           models.LandscapeStatusActive,
		DeploymentStatus: models.DeploymentStatusHealthy,
		GitHubConfigURL:  "https://github.com/test/config",
		AWSAccountID:     "123456789012",
		CAMProfileURL:    "https://cam.test.com/profile",
		SortOrder:        0,
		Metadata:         nil,
	}
}

// WithOrganization sets the organization ID for the landscape
func (f *LandscapeFactory) WithOrganization(orgID uuid.UUID) *models.Landscape {
	landscape := f.Create()
	landscape.OrganizationID = orgID
	return landscape
}

// WithName sets a custom name for the landscape
func (f *LandscapeFactory) WithName(name string) *models.Landscape {
	landscape := f.Create()
	landscape.Name = name
	landscape.DisplayName = name + " Landscape"
	return landscape
}

// WithLandscapeType sets a custom landscape type for the landscape
func (f *LandscapeFactory) WithLandscapeType(landscapeType models.LandscapeType) *models.Landscape {
	landscape := f.Create()
	landscape.LandscapeType = landscapeType
	return landscape
}

// ComponentDeploymentFactory provides methods to create test ComponentDeployment data
type ComponentDeploymentFactory struct{}

// NewComponentDeploymentFactory creates a new ComponentDeploymentFactory
func NewComponentDeploymentFactory() *ComponentDeploymentFactory {
	return &ComponentDeploymentFactory{}
}

// Create creates a test ComponentDeployment with default values
func (f *ComponentDeploymentFactory) Create() *models.ComponentDeployment {
	now := time.Now()
	return &models.ComponentDeployment{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ComponentID:     uuid.New(),
		LandscapeID:     uuid.New(),
		Version:         "1.0.0",
		GitCommitID:     "abc123def456",
		GitCommitTime:   &now,
		BuildTime:       &now,
		BuildProperties: nil,
		GitProperties:   nil,
		IsActive:        true,
		DeployedAt:      &now,
	}
}

// WithComponent sets the component ID for the deployment
func (f *ComponentDeploymentFactory) WithComponent(componentID uuid.UUID) *models.ComponentDeployment {
	deployment := f.Create()
	deployment.ComponentID = componentID
	return deployment
}

// WithLandscape sets the landscape ID for the deployment
func (f *ComponentDeploymentFactory) WithLandscape(landscapeID uuid.UUID) *models.ComponentDeployment {
	deployment := f.Create()
	deployment.LandscapeID = landscapeID
	return deployment
}

// WithVersion sets a custom version for the deployment
func (f *ComponentDeploymentFactory) WithVersion(version string) *models.ComponentDeployment {
	deployment := f.Create()
	deployment.Version = version
	return deployment
}

// WithActive sets the active status for the deployment
func (f *ComponentDeploymentFactory) WithActive(isActive bool) *models.ComponentDeployment {
	deployment := f.Create()
	deployment.IsActive = isActive
	return deployment
}

// FactorySet provides access to all factories
type FactorySet struct {
	Organization        *OrganizationFactory
	Member              *MemberFactory
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
		Member:              NewMemberFactory(),
		Group:               NewGroupFactory(),
		Team:                NewTeamFactory(),
		Project:             NewProjectFactory(),
		Component:           NewComponentFactory(),
		Landscape:           NewLandscapeFactory(),
		ComponentDeployment: NewComponentDeploymentFactory(),
	}
}

// CreateFullOrganizationHierarchy creates a complete organization with teams, members, projects, and components
func (fs *FactorySet) CreateFullOrganizationHierarchy() (*models.Organization, *models.Team, *models.Member, *models.Project, *models.Component, *models.Landscape) {
	// Create organization
	org := fs.Organization.Create()

	// Create team
	team := fs.Team.WithOrganization(org.ID)

	// Create member as team lead
	member := fs.Member.WithOrganization(org.ID)

	// deprecated: TeamLeadID removed; leadership is handled via linking table

	// Update member with team
	member.TeamID = &team.ID

	// Create project
	project := fs.Project.WithOrganization(org.ID)

	// Create component
	component := fs.Component.WithOrganization(org.ID)

	// Create landscape
	landscape := fs.Landscape.WithOrganization(org.ID)

	return org, team, member, project, component, landscape
}
