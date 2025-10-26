package main

import (
	"developer-portal-backend/internal/config"
	"developer-portal-backend/internal/database"
	"developer-portal-backend/internal/database/models"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Simple structures that directly match DB schema
type OrganizationData struct {
	Name        string                 `yaml:"name"`
	DisplayName string                 `yaml:"display_name"`
	Domain      string                 `yaml:"domain"`
	Description string                 `yaml:"description"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

type GroupData struct {
	Name             string                 `yaml:"name"`
	OrganizationName string                 `yaml:"organization_name"`
	DisplayName      string                 `yaml:"display_name"`
	Description      string                 `yaml:"description"`
	Metadata         map[string]interface{} `yaml:"metadata,omitempty"`
}

type TeamData struct {
	Name        string                 `yaml:"name"`
	GroupName   string                 `yaml:"group_name"`
	DisplayName string                 `yaml:"display_name"`
	Description string                 `yaml:"description"`
	Status      string                 `yaml:"status"`
	Links       []Link                 `yaml:"links,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

type MemberData struct {
	Name             string                 `yaml:"name"`
	OrganizationName string                 `yaml:"organization_name"`
	GroupName        string                 `yaml:"group_name,omitempty"`
	TeamName         string                 `yaml:"team_name,omitempty"`
	FullName         string                 `yaml:"full_name"`
	FirstName        string                 `yaml:"first_name"`
	LastName         string                 `yaml:"last_name"`
	Email            string                 `yaml:"email"`
	PhoneNumber      string                 `yaml:"phone_number,omitempty"`
	IUser            string                 `yaml:"iuser"`
	Role             string                 `yaml:"role"`
	TeamRole         string                 `yaml:"team_role"`
	IsActive         bool                   `yaml:"is_active"`
	ExternalType     string                 `yaml:"external_type"`
	Metadata         map[string]interface{} `yaml:"metadata,omitempty"`
}

type ComponentData struct {
	Name             string                 `yaml:"name"`
	OrganizationName string                 `yaml:"organization_name"`
	TeamName         string                 `yaml:"team_name,omitempty"`
	DisplayName      string                 `yaml:"display_name"`
	Description      string                 `yaml:"description"`
	ComponentType    string                 `yaml:"component_type"`
	Status           string                 `yaml:"status"`
	GroupName        string                 `yaml:"group_name,omitempty"`
	ArtifactName     string                 `yaml:"artifact_name,omitempty"`
	GitRepositoryURL string                 `yaml:"git_repository_url,omitempty"`
	DocumentationURL string                 `yaml:"documentation_url,omitempty"`
	Links            []Link                 `yaml:"links,omitempty"`
	Metadata         map[string]interface{} `yaml:"metadata,omitempty"`
}

type LandscapeData struct {
	Name             string                 `yaml:"name"`
	OrganizationName string                 `yaml:"organization_name"`
	DisplayName      string                 `yaml:"display_name"`
	Description      string                 `yaml:"description"`
	LandscapeType    string                 `yaml:"landscape_type"`
	EnvironmentGroup string                 `yaml:"environment_group,omitempty"`
	Status           string                 `yaml:"status"`
	DeploymentStatus string                 `yaml:"deployment_status"`
	GitHubConfigURL  string                 `yaml:"github_config_url,omitempty"`
	AWSAccountID     string                 `yaml:"aws_account_id,omitempty"`
	CAMProfileURL    string                 `yaml:"cam_profile_url,omitempty"`
	SortOrder        int                    `yaml:"sort_order"`
	Metadata         map[string]interface{} `yaml:"metadata,omitempty"`
}

type ProjectData struct {
	Name             string                 `yaml:"name"`
	OrganizationName string                 `yaml:"organization_name"`
	DisplayName      string                 `yaml:"display_name"`
	Description      string                 `yaml:"description"`
	ProjectType      string                 `yaml:"project_type"`
	Status           string                 `yaml:"status"`
	SortOrder        int                    `yaml:"sort_order"`
	Metadata         map[string]interface{} `yaml:"metadata,omitempty"`
}

type Link struct {
	URL   string `yaml:"url"`
	Title string `yaml:"title"`
	Icon  string `yaml:"icon"`
}

// File structures
type OrganizationsFile struct {
	Organizations []OrganizationData `yaml:"organizations"`
}

type GroupsFile struct {
	Groups []GroupData `yaml:"groups"`
}

type TeamsFile struct {
	Teams []TeamData `yaml:"teams"`
}

type MembersFile struct {
	Members []MemberData `yaml:"members"`
}

type ComponentsFile struct {
	Components []ComponentData `yaml:"components"`
}

type LandscapesFile struct {
	Landscapes []LandscapeData `yaml:"landscapes"`
}

type ProjectsFile struct {
	Projects []ProjectData `yaml:"projects"`
}

func main() {
	log.Println("🚀 Loading initial data from YAML files...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database with retry (for dockerized Postgres startup)
	db, err := connectWithRetry(cfg.DatabaseURL, 60, time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Load data from YAML files
	if err := loadDataFromYAMLFiles(db, "scripts/data"); err != nil {
		log.Fatalf("Failed to load data from YAML files: %v", err)
	}

	log.Println("✅ Initial data loaded successfully!")
}

// connectWithRetry attempts to initialize the DB with retries to wait for Postgres readiness.
func connectWithRetry(dsn string, maxAttempts int, delay time.Duration) (*gorm.DB, error) {
	// Configure database options to suppress verbose logging during data loading
	opts := &database.Options{
		LogLevel: logger.Silent, // Suppress all GORM logs including SQL queries and "record not found"
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err := database.Initialize(dsn, opts)
		if err == nil {
			return db, nil
		}
		// Only log every 10 attempts to reduce noise
		if attempt%10 == 0 || attempt == maxAttempts {
			log.Printf("Database not ready (%d/%d): %v", attempt, maxAttempts, err)
		}
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("database not ready after %d attempts", maxAttempts)
}

func loadDataFromYAMLFiles(db *gorm.DB, dataDir string) error {
	// Load all data from YAML files
	organizations, err := loadOrganizations(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load organizations: %w", err)
	}

	groups, err := loadGroups(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load groups: %w", err)
	}

	teams, err := loadTeams(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load teams: %w", err)
	}

	members, err := loadMembers(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load members: %w", err)
	}

	components, err := loadComponents(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	landscapes, err := loadLandscapes(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load landscapes: %w", err)
	}

	projects, err := loadProjects(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	// Create organizations first
	orgMap := make(map[string]*models.Organization)
	orgCreated := 0
	for _, orgData := range organizations {
		org, created, err := createOrganization(db, orgData)
		if err != nil {
			return fmt.Errorf("failed to create organization %s: %w", orgData.Name, err)
		}
		orgMap[orgData.Name] = org
		if created {
			orgCreated++
		}
	}
	log.Printf("📋 Organizations: %d created, %d total", orgCreated, len(organizations))

	// Create groups
	groupMap := make(map[string]*models.Group)
	groupCreated := 0
	for _, groupData := range groups {
		group, created, err := createGroup(db, groupData, orgMap)
		if err != nil {
			return fmt.Errorf("failed to create group %s: %w", groupData.Name, err)
		}
		groupMap[groupData.Name] = group
		if created {
			groupCreated++
		}
	}
	log.Printf("📋 Groups: %d created, %d total", groupCreated, len(groups))

	// Create teams
	teamMap := make(map[string]*models.Team)
	teamCreated := 0
	for _, teamData := range teams {
		team, created, err := createTeam(db, teamData, groupMap)
		if err != nil {
			return fmt.Errorf("failed to create team %s: %w", teamData.Name, err)
		}
		teamMap[teamData.Name] = team
		if created {
			teamCreated++
		}
	}
	log.Printf("📋 Teams: %d created, %d total", teamCreated, len(teams))

	// Create members
	memberCreated := 0
	for _, memberData := range members {
		_, created, err := createMember(db, memberData, orgMap, groupMap, teamMap)
		if err != nil {
			return fmt.Errorf("failed to create member %s: %w", memberData.Name, err)
		}
		if created {
			memberCreated++
		}
	}
	log.Printf("📋 Members: %d created, %d total", memberCreated, len(members))

	// Create components
	componentCreated := 0
	for _, componentData := range components {
		_, created, err := createComponent(db, componentData, orgMap, teamMap)
		if err != nil {
			log.Printf("⚠️  Warning: failed to create component %s: %v", componentData.Name, err)
			continue // Continue with other components
		}
		if created {
			componentCreated++
		}
	}
	log.Printf("📋 Components: %d created, %d total", componentCreated, len(components))

	// Create landscapes
	landscapeCreated := 0
	for _, landscapeData := range landscapes {
		_, created, err := createLandscape(db, landscapeData, orgMap)
		if err != nil {
			log.Printf("⚠️  Warning: failed to create landscape %s: %v", landscapeData.Name, err)
			continue // Continue with other landscapes
		}
		if created {
			landscapeCreated++
		}
	}
	log.Printf("📋 Landscapes: %d created, %d total", landscapeCreated, len(landscapes))

	// Create projects (must be done before creating project relationships)
	projectMap := make(map[string]*models.Project)
	projectCreated := 0
	for _, projectData := range projects {
		project, created, err := createProject(db, projectData, orgMap)
		if err != nil {
			log.Printf("⚠️  Warning: failed to create project %s: %v", projectData.Name, err)
			continue // Continue with other projects
		}
		projectMap[projectData.Name] = project
		if created {
			projectCreated++
		}
	}
	log.Printf("📋 Projects: %d created, %d total", projectCreated, len(projects))

	// Create project-component relationships
	componentMap := make(map[string]*models.Component)
	// Build component map for relationship creation
	var allComponents []models.Component
	if err := db.Find(&allComponents).Error; err == nil {
		for i := range allComponents {
			componentMap[allComponents[i].Name] = &allComponents[i]
		}
	}

	landscapeMap := make(map[string]*models.Landscape)
	// Build landscape map for relationship creation
	var allLandscapes []models.Landscape
	if err := db.Find(&allLandscapes).Error; err == nil {
		for i := range allLandscapes {
			landscapeMap[allLandscapes[i].Name] = &allLandscapes[i]
		}
	}

	// Create project-component and project-landscape relationships
	projectComponentsCreated, projectLandscapesCreated := createProjectRelationships(db, projectMap, componentMap, landscapeMap)
	log.Printf("📋 Project-Component relationships: %d created", projectComponentsCreated)
	log.Printf("📋 Project-Landscape relationships: %d created", projectLandscapesCreated)

	return nil
}

func loadOrganizations(dataDir string) ([]OrganizationData, error) {
	var allOrgs []OrganizationData

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && strings.Contains(path, "organizations") {
			var file OrganizationsFile
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &file); err != nil {
				return err
			}

			allOrgs = append(allOrgs, file.Organizations...)
		}
		return nil
	})

	return allOrgs, err
}

func loadGroups(dataDir string) ([]GroupData, error) {
	var allGroups []GroupData

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && strings.Contains(path, "groups") {
			var file GroupsFile
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &file); err != nil {
				return err
			}

			allGroups = append(allGroups, file.Groups...)
		}
		return nil
	})

	return allGroups, err
}

func loadTeams(dataDir string) ([]TeamData, error) {
	var allTeams []TeamData

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && strings.Contains(path, "teams") {
			var file TeamsFile
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &file); err != nil {
				return err
			}

			allTeams = append(allTeams, file.Teams...)
		}
		return nil
	})

	return allTeams, err
}

func loadMembers(dataDir string) ([]MemberData, error) {
	var allMembers []MemberData

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && strings.Contains(path, "members") {
			var file MembersFile
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &file); err != nil {
				return err
			}

			allMembers = append(allMembers, file.Members...)
		}
		return nil
	})

	return allMembers, err
}

func loadComponents(dataDir string) ([]ComponentData, error) {
	var allComponents []ComponentData

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && strings.Contains(path, "components") {
			var file ComponentsFile
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &file); err != nil {
				return err
			}

			allComponents = append(allComponents, file.Components...)
		}
		return nil
	})

	return allComponents, err
}

func loadLandscapes(dataDir string) ([]LandscapeData, error) {
	var allLandscapes []LandscapeData

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && strings.Contains(path, "landscapes") {
			var file LandscapesFile
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &file); err != nil {
				return err
			}

			allLandscapes = append(allLandscapes, file.Landscapes...)
		}
		return nil
	})

	return allLandscapes, err
}

func loadProjects(dataDir string) ([]ProjectData, error) {
	var allProjects []ProjectData

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".yaml") && strings.Contains(path, "projects") {
			var file ProjectsFile
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := yaml.Unmarshal(data, &file); err != nil {
				return err
			}

			allProjects = append(allProjects, file.Projects...)
		}
		return nil
	})

	return allProjects, err
}

func createOrganization(db *gorm.DB, orgData OrganizationData) (*models.Organization, bool, error) {
	var org models.Organization
	if err := db.Where("name = ?", orgData.Name).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metadataJSON, _ := json.Marshal(orgData.Metadata)

			org = models.Organization{
				Name:        orgData.Name,
				DisplayName: orgData.DisplayName,
				Domain:      orgData.Domain,
				Description: orgData.Description,
				Metadata:    metadataJSON,
			}

			if err := db.Create(&org).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create organization: %w", err)
			}
			return &org, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query organization: %w", err)
		}
	}

	return &org, false, nil // created = false (existing)
}

func createGroup(db *gorm.DB, groupData GroupData, orgMap map[string]*models.Organization) (*models.Group, bool, error) {
	org := orgMap[groupData.OrganizationName]
	if org == nil {
		return nil, false, fmt.Errorf("organization %s not found for group %s", groupData.OrganizationName, groupData.Name)
	}

	var group models.Group
	if err := db.Where("name = ? AND organization_id = ?", groupData.Name, org.ID).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metadataJSON, _ := json.Marshal(groupData.Metadata)

			group = models.Group{
				OrganizationID: org.ID,
				Name:           groupData.Name,
				DisplayName:    groupData.DisplayName,
				Description:    groupData.Description,
				Metadata:       metadataJSON,
			}

			if err := db.Create(&group).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create group: %w", err)
			}
			return &group, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query group: %w", err)
		}
	}

	return &group, false, nil // created = false (existing)
}

func createTeam(db *gorm.DB, teamData TeamData, groupMap map[string]*models.Group) (*models.Team, bool, error) {
	group := groupMap[teamData.GroupName]
	if group == nil {
		return nil, false, fmt.Errorf("group %s not found for team %s", teamData.GroupName, teamData.Name)
	}

	var team models.Team
	if err := db.Where("name = ? AND group_id = ?", teamData.Name, group.ID).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			linksJSON, _ := json.Marshal(teamData.Links)
			metadataJSON, _ := json.Marshal(teamData.Metadata)

			status := models.TeamStatusActive
			if teamData.Status != "" {
				status = models.TeamStatus(teamData.Status)
			}

			team = models.Team{
				GroupID:     group.ID,
				Name:        teamData.Name,
				DisplayName: teamData.DisplayName,
				Description: teamData.Description,
				Status:      status,
				Links:       linksJSON,
				Metadata:    metadataJSON,
			}

			if err := db.Create(&team).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create team: %w", err)
			}
			return &team, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query team: %w", err)
		}
	}

	return &team, false, nil // created = false (existing)
}

func createMember(db *gorm.DB, memberData MemberData, orgMap map[string]*models.Organization, groupMap map[string]*models.Group, teamMap map[string]*models.Team) (*models.Member, bool, error) {
	org := orgMap[memberData.OrganizationName]
	if org == nil {
		return nil, false, fmt.Errorf("organization %s not found for member %s", memberData.OrganizationName, memberData.Name)
	}

	var groupID *uuid.UUID
	if memberData.GroupName != "" {
		if group := groupMap[memberData.GroupName]; group != nil {
			groupID = &group.ID
		}
	}

	var teamID *uuid.UUID
	if memberData.TeamName != "" {
		if team := teamMap[memberData.TeamName]; team != nil {
			teamID = &team.ID
		}
	}

	var member models.Member
	if err := db.Where("email = ?", memberData.Email).First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metadataJSON, _ := json.Marshal(memberData.Metadata)

			member = models.Member{
				OrganizationID: org.ID,
				GroupID:        groupID,
				TeamID:         teamID,
				FullName:       memberData.FullName,
				FirstName:      memberData.FirstName,
				LastName:       memberData.LastName,
				Email:          memberData.Email,
				PhoneNumber:    memberData.PhoneNumber,
				IUser:          memberData.IUser,
				Role:           models.MemberRole(memberData.Role),
				TeamRole:       models.TeamRole(memberData.TeamRole),
				IsActive:       memberData.IsActive,
				ExternalType:   models.ExternalType(memberData.ExternalType),
				Metadata:       metadataJSON,
			}

			if err := db.Create(&member).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create member: %w", err)
			}

			// If this is a team lead, create the team leadership record
			if memberData.TeamRole == "team_lead" && teamID != nil {
				leadership := models.TeamLeadership{
					TeamID:   *teamID,
					MemberID: member.ID,
				}
				if err := db.Where("team_id = ?", *teamID).FirstOrCreate(&leadership, leadership).Error; err != nil {
					log.Printf("⚠️  Warning: failed to create team leadership: %v", err)
				}
			}
			return &member, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query member: %w", err)
		}
	}

	return &member, false, nil // created = false (existing)
}

func createComponent(db *gorm.DB, componentData ComponentData, orgMap map[string]*models.Organization, teamMap map[string]*models.Team) (*models.Component, bool, error) {
	org := orgMap[componentData.OrganizationName]
	if org == nil {
		return nil, false, fmt.Errorf("organization %s not found for component %s", componentData.OrganizationName, componentData.Name)
	}

	// Try to find team if specified
	var ownerTeamID *uuid.UUID
	if componentData.TeamName != "" {
		if team := teamMap[componentData.TeamName]; team != nil {
			ownerTeamID = &team.ID
		} else {
			// Team not found, log warning but continue
			log.Printf("⚠️  Warning: team %s not found for component %s", componentData.TeamName, componentData.Name)
		}
	}

	var component models.Component
	if err := db.Where("name = ? AND organization_id = ?", componentData.Name, org.ID).First(&component).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			linksJSON, _ := json.Marshal(componentData.Links)
			metadataJSON, _ := json.Marshal(componentData.Metadata)

			componentType := models.ComponentTypeService
			if componentData.ComponentType != "" {
				componentType = models.ComponentType(componentData.ComponentType)
			}

			status := models.ComponentStatusActive
			if componentData.Status != "" {
				status = models.ComponentStatus(componentData.Status)
			}

			component = models.Component{
				OrganizationID:   org.ID,
				Name:             componentData.Name,
				DisplayName:      componentData.DisplayName,
				Description:      componentData.Description,
				ComponentType:    componentType,
				Status:           status,
				GroupName:        componentData.GroupName,
				ArtifactName:     componentData.ArtifactName,
				GitRepositoryURL: componentData.GitRepositoryURL,
				DocumentationURL: componentData.DocumentationURL,
				Links:            linksJSON,
				Metadata:         metadataJSON,
			}

			if err := db.Create(&component).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create component: %w", err)
			}

			// Create team ownership if team is specified
			if ownerTeamID != nil {
				ownership := models.TeamComponentOwnership{
					TeamID:        *ownerTeamID,
					ComponentID:   component.ID,
					OwnershipType: models.OwnershipTypePrimary,
				}
				if err := db.Create(&ownership).Error; err != nil {
					log.Printf("⚠️  Warning: failed to create team ownership for component %s: %v", componentData.Name, err)
				}
			}

			return &component, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query component: %w", err)
		}
	}

	return &component, false, nil // created = false (existing)
}

func createLandscape(db *gorm.DB, landscapeData LandscapeData, orgMap map[string]*models.Organization) (*models.Landscape, bool, error) {
	org := orgMap[landscapeData.OrganizationName]
	if org == nil {
		return nil, false, fmt.Errorf("organization %s not found for landscape %s", landscapeData.OrganizationName, landscapeData.Name)
	}

	var landscape models.Landscape
	if err := db.Where("name = ? AND organization_id = ?", landscapeData.Name, org.ID).First(&landscape).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metadataJSON, _ := json.Marshal(landscapeData.Metadata)
			// Note: Links are stored in metadata, not as a separate field in Landscape model

			landscapeType := models.LandscapeTypeDevelopment
			if landscapeData.LandscapeType != "" {
				landscapeType = models.LandscapeType(landscapeData.LandscapeType)
			}

			status := models.LandscapeStatusActive
			if landscapeData.Status != "" {
				status = models.LandscapeStatus(landscapeData.Status)
			}

			deploymentStatus := models.DeploymentStatusUnknown
			if landscapeData.DeploymentStatus != "" {
				deploymentStatus = models.DeploymentStatus(landscapeData.DeploymentStatus)
			}

			landscape = models.Landscape{
				OrganizationID:   org.ID,
				Name:             landscapeData.Name,
				DisplayName:      landscapeData.DisplayName,
				Description:      landscapeData.Description,
				LandscapeType:    landscapeType,
				EnvironmentGroup: landscapeData.EnvironmentGroup,
				Status:           status,
				DeploymentStatus: deploymentStatus,
				GitHubConfigURL:  landscapeData.GitHubConfigURL,
				AWSAccountID:     landscapeData.AWSAccountID,
				CAMProfileURL:    landscapeData.CAMProfileURL,
				SortOrder:        landscapeData.SortOrder,
				Metadata:         metadataJSON,
			}

			if err := db.Create(&landscape).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create landscape: %w", err)
			}

			return &landscape, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query landscape: %w", err)
		}
	}

	return &landscape, false, nil // created = false (existing)
}

func createProject(db *gorm.DB, projectData ProjectData, orgMap map[string]*models.Organization) (*models.Project, bool, error) {
	org := orgMap[projectData.OrganizationName]
	if org == nil {
		return nil, false, fmt.Errorf("organization %s not found for project %s", projectData.OrganizationName, projectData.Name)
	}

	var project models.Project
	if err := db.Where("name = ? AND organization_id = ?", projectData.Name, org.ID).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metadataJSON, _ := json.Marshal(projectData.Metadata)

			projectType := models.ProjectTypePlatform
			if projectData.ProjectType != "" {
				projectType = models.ProjectType(projectData.ProjectType)
			}

			status := models.ProjectStatusActive
			if projectData.Status != "" {
				status = models.ProjectStatus(projectData.Status)
			}

			project = models.Project{
				OrganizationID: org.ID,
				Name:           projectData.Name,
				DisplayName:    projectData.DisplayName,
				Description:    projectData.Description,
				ProjectType:    projectType,
				Status:         status,
				SortOrder:      projectData.SortOrder,
				Metadata:       metadataJSON,
			}

			if err := db.Create(&project).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create project: %w", err)
			}

			return &project, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query project: %w", err)
		}
	}

	return &project, false, nil // created = false (existing)
}

func createProjectRelationships(db *gorm.DB, projectMap map[string]*models.Project, componentMap map[string]*models.Component, landscapeMap map[string]*models.Landscape) (int, int) {
	projectComponentsCreated := 0
	projectLandscapesCreated := 0

	// Map project names to their associated namespaces/systems/groups
	projectMappings := map[string][]string{
		"atom": {"unified-services", "atom", "resource-manager", "platform-engineering"},
		"btp":  {"cloud-foundry", "btp", "cis-cf", "cis-neo", "service-manager", "cloud-automation"},
	}

	// Create project-component relationships based on namespace/system
	for _, component := range componentMap {
		// Parse metadata to find namespace and system
		var metadata map[string]interface{}
		if err := json.Unmarshal(component.Metadata, &metadata); err == nil {
			namespace, _ := metadata["namespace"].(string)
			system, _ := metadata["system"].(string)
			domain, _ := metadata["domain"].(string)

			// Determine which project(s) this component belongs to
			for projectName, keywords := range projectMappings {
				project := projectMap[projectName]
				if project == nil {
					continue
				}

				// Check if namespace, system, or domain matches any keyword
				matched := false
				for _, keyword := range keywords {
					if namespace == keyword || system == keyword || domain == keyword || component.GroupName == keyword {
						matched = true
						break
					}
				}

				if matched {
					// Create project-component relationship
					var existing models.ProjectComponent
					err := db.Where("project_id = ? AND component_id = ?", project.ID, component.ID).First(&existing).Error
					if err == gorm.ErrRecordNotFound {
						projectComponent := models.ProjectComponent{
							ProjectID:     project.ID,
							ComponentID:   component.ID,
							OwnershipType: models.OwnershipTypePrimary,
							SortOrder:     0,
						}
						if err := db.Create(&projectComponent).Error; err != nil {
							log.Printf("⚠️  Warning: failed to create project-component relationship: %v", err)
						} else {
							projectComponentsCreated++
						}
					}
				}
			}
		}
	}

	// Create project-landscape relationships based on environment_group
	for _, landscape := range landscapeMap {
		// Find project based on environment group
		for projectName, keywords := range projectMappings {
			project := projectMap[projectName]
			if project == nil {
				continue
			}

			// Check if environment group matches any keyword
			matched := false
			for _, keyword := range keywords {
				if landscape.EnvironmentGroup == keyword {
					matched = true
					break
				}
			}

			if matched {
				// Create project-landscape relationship
				var existing models.ProjectLandscape
				err := db.Where("project_id = ? AND landscape_id = ?", project.ID, landscape.ID).First(&existing).Error
				if err == gorm.ErrRecordNotFound {
					projectLandscape := models.ProjectLandscape{
						ProjectID:      project.ID,
						LandscapeID:    landscape.ID,
						LandscapeGroup: landscape.EnvironmentGroup,
						SortOrder:      landscape.SortOrder,
					}
					if err := db.Create(&projectLandscape).Error; err != nil {
						log.Printf("⚠️  Warning: failed to create project-landscape relationship: %v", err)
					} else {
						projectLandscapesCreated++
					}
				}
			}
		}
	}

	return projectComponentsCreated, projectLandscapesCreated
}
