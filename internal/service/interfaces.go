package service

import (
	"context"

	"developer-portal-backend/internal/auth"
	"developer-portal-backend/internal/database/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=../mocks/service_mocks.go -package=mocks

// OrganizationServiceInterface defines the interface for organization service
type OrganizationServiceInterface interface {
	Create(req *CreateOrganizationRequest) (*OrganizationResponse, error)
	GetByID(id uuid.UUID) (*OrganizationResponse, error)
	GetByName(name string) (*OrganizationResponse, error)
	GetByDomain(domain string) (*OrganizationResponse, error)
	Update(id uuid.UUID, req *UpdateOrganizationRequest) (*OrganizationResponse, error)
	Delete(id uuid.UUID) error
	GetAll(page, pageSize int) (*OrganizationListResponse, error)
	GetWithMembers(id uuid.UUID) (*models.Organization, error)
	GetWithGroups(id uuid.UUID) (*models.Organization, error)
	GetWithProjects(id uuid.UUID) (*models.Organization, error)
	GetWithComponents(id uuid.UUID) (*models.Organization, error)
	GetWithLandscapes(id uuid.UUID) (*models.Organization, error)
}

// MemberServiceInterface defines the interface for member service
type MemberServiceInterface interface {
	CreateMember(req *CreateMemberRequest) (*MemberResponse, error)
	GetMemberByID(id uuid.UUID) (*MemberResponse, error)
	GetMembersByOrganization(organizationID uuid.UUID, limit, offset int) ([]MemberResponse, int64, error)
	UpdateMember(id uuid.UUID, req *UpdateMemberRequest) (*MemberResponse, error)
	DeleteMember(id uuid.UUID) error
	SearchMembers(organizationID uuid.UUID, query string, limit, offset int) ([]MemberResponse, int64, error)
	GetActiveMembers(organizationID uuid.UUID, limit, offset int) ([]MemberResponse, int64, error)
	GetQuickLinks(id uuid.UUID) (*QuickLinksResponse, error)
	AddQuickLink(id uuid.UUID, req *AddQuickLinkRequest) (*MemberResponse, error)
	RemoveQuickLink(id uuid.UUID, linkURL string) (*MemberResponse, error)
}

// GroupServiceInterface defines the interface for group service
type GroupServiceInterface interface {
	Create(req *CreateGroupRequest) (*GroupResponse, error)
	GetByID(id uuid.UUID) (*GroupResponse, error)
	GetByName(organizationID uuid.UUID, name string) (*GroupResponse, error)
	GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*GroupListResponse, error)
	GetAll(orgID uuid.UUID, page, pageSize int) (*GroupListResponse, error)
	GetWithTeams(id uuid.UUID, page, pageSize int) (*GroupWithTeamsResponse, error)
	Search(organizationID uuid.UUID, query string, page, pageSize int) (*GroupListResponse, error)
	Update(id uuid.UUID, req *UpdateGroupRequest) (*GroupResponse, error)
	Delete(id uuid.UUID) error
}

// TeamServiceInterface defines the interface for team service
type TeamServiceInterface interface {
	Create(req *CreateTeamRequest) (*TeamResponse, error)
	GetByID(id uuid.UUID) (*TeamResponse, error)
	GetByName(organization uuid.UUID, name string) (*TeamResponse, error)
	GetByOrganization(organizationID uuid.UUID, page, pageSize int) (*TeamListResponse, error)
	Search(organizationID uuid.UUID, query string, page, pageSize int) (*TeamListResponse, error)
	Update(id uuid.UUID, req *UpdateTeamRequest) (*TeamResponse, error)
	Delete(id uuid.UUID) error
	GetWithMembers(id uuid.UUID) (*models.Team, error)
	GetWithProjects(id uuid.UUID) (*models.Team, error)
	GetWithComponentOwnerships(id uuid.UUID) (*models.Team, error)
	GetWithDutySchedules(id uuid.UUID) (*models.Team, error)
	GetTeamLead(id uuid.UUID) (*models.Team, error)
	GetAllTeams(organizationID *uuid.UUID, page, pageSize int) (*TeamListResponse, error)
	GetMembersOnly(id uuid.UUID) ([]MemberResponse, error)
	GetTeamMembersByName(organizationID uuid.UUID, teamName string, page, pageSize int) ([]models.Member, int64, error)
	GetTeamComponentsByName(organizationID uuid.UUID, teamName string, page, pageSize int) ([]models.Component, int64, error)
	GetTeamComponentsByID(id uuid.UUID, page, pageSize int) ([]models.Component, int64, error)
	AddLink(id uuid.UUID, req *AddLinkRequest) (*TeamResponse, error)
	RemoveLink(id uuid.UUID, linkURL string) (*TeamResponse, error)
	UpdateLinks(id uuid.UUID, req *UpdateLinksRequest) (*TeamResponse, error)
}

// ProjectServiceInterface defines the interface for project service
type ProjectServiceInterface interface {
	CreateProject(req *CreateProjectRequest) (*ProjectResponse, error)
	GetProjectByID(id uuid.UUID) (*ProjectResponse, error)
	GetProjectsByOrganization(organizationID uuid.UUID, limit, offset int) ([]ProjectResponse, int64, error)
	UpdateProject(id uuid.UUID, req *UpdateProjectRequest) (*ProjectResponse, error)
	DeleteProject(id uuid.UUID) error
}

// ComponentServiceInterface defines the interface for component service
type ComponentServiceInterface interface {
	CreateComponent(req *CreateComponentRequest) (*ComponentResponse, error)
	GetComponentByID(id uuid.UUID) (*ComponentResponse, error)
	GetComponentsByProject(projectID uuid.UUID, limit, offset int) ([]ComponentResponse, int64, error)
	UpdateComponent(id uuid.UUID, req *UpdateComponentRequest) (*ComponentResponse, error)
	DeleteComponent(id uuid.UUID) error
}

// LandscapeServiceInterface defines the interface for landscape service
type LandscapeServiceInterface interface {
	CreateLandscape(req *CreateLandscapeRequest) (*LandscapeResponse, error)
	GetLandscapeByID(id uuid.UUID) (*LandscapeResponse, error)
	GetLandscapesByOrganization(organizationID uuid.UUID, limit, offset int) ([]LandscapeResponse, int64, error)
	UpdateLandscape(id uuid.UUID, req *UpdateLandscapeRequest) (*LandscapeResponse, error)
	DeleteLandscape(id uuid.UUID) error
}

// ComponentDeploymentServiceInterface defines the interface for component deployment service
type ComponentDeploymentServiceInterface interface {
	CreateComponentDeployment(req *CreateComponentDeploymentRequest) (*ComponentDeploymentResponse, error)
	GetComponentDeploymentByID(id uuid.UUID) (*ComponentDeploymentResponse, error)
	GetComponentDeploymentsByComponent(componentID uuid.UUID, limit, offset int) ([]ComponentDeploymentResponse, int64, error)
	UpdateComponentDeployment(id uuid.UUID, req *UpdateComponentDeploymentRequest) (*ComponentDeploymentResponse, error)
	DeleteComponentDeployment(id uuid.UUID) error
}

// GitHubServiceInterface defines the interface for GitHub service
type GitHubServiceInterface interface {
	GetUserOpenPullRequests(ctx context.Context, claims *auth.AuthClaims, state, sort, direction string, perPage, page int) (*PullRequestsResponse, error)
	GetUserTotalContributions(ctx context.Context, claims *auth.AuthClaims, period string) (*TotalContributionsResponse, error)
}

// JenkinsServiceInterface defines the interface for Jenkins service
type JenkinsServiceInterface interface {
	GetJobParameters(ctx context.Context, jaasName, jobName string) (interface{}, error)
	TriggerJob(ctx context.Context, jaasName, jobName string, parameters map[string]string) error
}

// SonarServiceInterface defines the interface for Sonar service
type SonarServiceInterface interface {
	GetComponentMeasures(projectKey string) (*SonarCombinedResponse, error)
}

// JiraServiceInterface defines the interface for Jira service
type JiraServiceInterface interface {
	GetIssues(filters JiraIssueFilters) (*JiraIssuesResponse, error)
	GetIssuesCount(filters JiraIssueFilters) (int, error)
}

// AICoreServiceInterface defines the interface for AI Core service
type AICoreServiceInterface interface {
	GetDeployments(c *gin.Context) (*AICoreDeploymentsResponse, error)
	GetDeploymentDetails(c *gin.Context, deploymentID string) (*AICoreDeploymentDetailsResponse, error)
	GetModels(c *gin.Context, scenarioID string) (*AICoreModelsResponse, error)
	GetConfigurations(c *gin.Context) (*AICoreConfigurationsResponse, error)
	CreateConfiguration(c *gin.Context, req *AICoreConfigurationRequest) (*AICoreConfigurationResponse, error)
	CreateDeployment(c *gin.Context, req *AICoreDeploymentRequest) (*AICoreDeploymentResponse, error)
	UpdateDeployment(c *gin.Context, deploymentID string, req *AICoreDeploymentModificationRequest) (*AICoreDeploymentModificationResponse, error)
	DeleteDeployment(c *gin.Context, deploymentID string) (*AICoreDeploymentDeletionResponse, error)
}
