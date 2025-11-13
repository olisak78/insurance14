package main

import (
	"developer-portal-backend/internal/config"
	"developer-portal-backend/internal/database"
	"developer-portal-backend/internal/database/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Simple structures that directly match DB schema
type OrganizationData struct {
	Name        string `yaml:"name"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Owner       string `yaml:"owner"`
	Email       string `yaml:"email"`
}

type GroupData struct {
	Name        string `yaml:"name"`
	OrgName     string `yaml:"org"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Owner       string `yaml:"owner"`
	Email       string `yaml:"email"`
	Picture     string `yaml:"picture_url"`
}

type TeamData struct {
	Name        string                 `yaml:"name"`
	GroupName   string                 `yaml:"group_name"`
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	Owner       string                 `yaml:"owner"`
	Email       string                 `yaml:"email"`
	Picture     string                 `yaml:"picture_url"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

type UserData struct {
	UserID      string                 `yaml:"name"`
	TeamName    string                 `yaml:"team_name"`
	FirstName   string                 `yaml:"first_name"`
	LastName    string                 `yaml:"last_name"`
	Email       string                 `yaml:"email"`
	PhoneNumber string                 `yaml:"phone_number"`
	TeamDomain  string                 `yaml:"team_domain"`
	TeamRole    string                 `yaml:"team_role"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

type ComponentData struct {
	Name        string                 `yaml:"name"`
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	Owner       string                 `yaml:"owner,omitempty"`
	Project     string                 `yaml:"project"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

type ProjectData struct {
	Name        string                 `yaml:"name"`
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

// Link for component JSON field
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

type UsersFile struct {
	Users []UserData `yaml:"users"`
}

type ComponentsFile struct {
	Components []ComponentData `yaml:"components"`
}

type LandscapeData struct {
	Name        string                 `yaml:"name"`
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	Project     string                 `yaml:"project"`
	Environment string                 `yaml:"environment"`
	Domain      string                 `yaml:"domain,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

type LandscapesFile struct {
	Landscapes []LandscapeData `yaml:"landscapes"`
}

type ProjectsFile struct {
	Projects []ProjectData `yaml:"projects"`
}

type CategoriesFile struct {
	Categories []CategoryData `yaml:"categories"`
}

type CategoryData struct {
	Name  string `yaml:"name"`
	Title string `yaml:"title"`
	Icon  string `yaml:"icon"`
	Color string `yaml:"color"`
}

// Links initial data (new)
type LinksFile struct {
	Links []InitialLinkData `yaml:"links"`
}

type InitialLinkData struct {
	Title       string      `yaml:"title"`
	Description string      `yaml:"description"`
	URL         string      `yaml:"url"`
	Category    string      `yaml:"category"`
	TeamName    string      `yaml:"team_name,omitempty"`
	Tags        []string    `yaml:"-"`
	TagsRaw     interface{} `yaml:"tags"` // supports "a, b, c" or ["a","b"]
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

	users, err := loadUsers(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}

	components, err := loadComponents(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	// Note: merged landscapes will be loaded later after projects to allow linking by project name.

	projects, err := loadProjects(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	categories, err := loadCategories(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}

	// Create categories
	catCreated := 0
	for _, categoryData := range categories {
		_, created, err := createCategory(db, categoryData)
		if err != nil {
			log.Printf("⚠️  Warning: failed to create category %s: %v", categoryData.Name, err)
			continue
		}
		if created {
			catCreated++
		}
	}
	log.Printf("📋 Categories: %d created, %d total", catCreated, len(categories))

	links, err := loadLinks(dataDir) // NEW
	if err != nil {
		return fmt.Errorf("failed to load links: %w", err)
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

	// Create users
	userCreated := 0
	for _, userData := range users {
		_, created, err := createUser(db, userData, teamMap)
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", userData.UserID, err)
		}
		if created {
			userCreated++
		}
	}
	log.Printf("📋 Users: %d created, %d total", userCreated, len(users))

	// Create links (NEW)
	linkCreated := 0
	linkUpdated := 0
	linkFailed := 0
	failedDetails := make([]string, 0)
	duplicateDetails := make([]string, 0)
	for _, linkData := range links {
		_, created, err := createLink(db, linkData)
		if err != nil {
			linkFailed++
			failedDetails = append(failedDetails, fmt.Sprintf("%s (%s): %v", linkData.Title, linkData.URL, err))
			log.Printf("❌ Error creating link %q: %v", linkData.Title, err)
			continue
		}
		if created {
			linkCreated++
		} else {
			// Existing link (duplicate in YAML relative to DB)
			linkUpdated++
			duplicateDetails = append(duplicateDetails, fmt.Sprintf("%s (%s)", linkData.Title, linkData.URL))
		}
	}
	log.Printf("📋 Links: %d created, %d updated/existing, %d failed, %d total", linkCreated, linkUpdated, linkFailed, len(links))
	if linkFailed > 0 {
		log.Printf("❗ Failed links: %s", strings.Join(failedDetails, "; "))
	}
	if len(duplicateDetails) > 0 {
		log.Printf("ℹ️ Duplicates (already existed): %s", strings.Join(duplicateDetails, "; "))
	}

	// Landscapes will be created after Projects from merged landscapes.yaml

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

	// Create components (after projects to ensure projects are defined only via projects.yaml)
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

	// Load landscapes and create them, then link to projects
	landscapes, err := loadLandscapes(dataDir)
	if err != nil {
		return fmt.Errorf("failed to load landscapes: %w", err)
	}

	landscapeMap := make(map[string]*models.Landscape)
	landscapeCreated := 0
	landscapeUpdated := 0
	landscapeFailed := 0
	landscapeDuplicateDetails := make([]string, 0)
	landscapeFailedDetails := make([]string, 0)

	// Resolve default organization for landscapes
	org := orgMap["sap-cfs"]
	if org == nil {
		return fmt.Errorf("organization %s not found for landscapes", "sap-cfs")
	}

	for _, item := range landscapes {
		l, created, err := createLandscapeFromYAML(db, item, org.ID)
		if err != nil {
			log.Printf("⚠️  Warning: failed to create landscape %s: %v", item.Name, err)
			landscapeFailed++
			landscapeFailedDetails = append(landscapeFailedDetails, fmt.Sprintf("%s: %v", item.Name, err))
			continue
		}
		landscapeMap[item.Name] = l
		if created {
			landscapeCreated++
		} else {
			landscapeUpdated++
			landscapeDuplicateDetails = append(
				landscapeDuplicateDetails,
				fmt.Sprintf("%s (project=%s, env=%s)", item.Name, item.Project, strings.ToLower(item.Environment)),
			)
		}

	}
	log.Printf("📋 Landscapes: %d created, %d updated/existing, %d failed, %d total", landscapeCreated, landscapeUpdated, landscapeFailed, len(landscapes))
	if landscapeFailed > 0 {
		log.Printf("❗ Failed landscapes: %s", strings.Join(landscapeFailedDetails, "; "))
	}
	if len(landscapeDuplicateDetails) > 0 {
		log.Printf("ℹ️ Duplicates (already existed): %s", strings.Join(landscapeDuplicateDetails, "; "))
	}

	// Create project-component relationships

	return nil
}

func loadOrganizations(dataDir string) ([]OrganizationData, error) {
	var allOrgs []OrganizationData

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == "organizations.yaml" {
			path := filepath.Join(dataDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			var file OrganizationsFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				return nil, err
			}
			allOrgs = append(allOrgs, file.Organizations...)
		}
	}

	return allOrgs, nil
}

func loadGroups(dataDir string) ([]GroupData, error) {
	var allGroups []GroupData

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == "groups.yaml" {
			path := filepath.Join(dataDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			var file GroupsFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				return nil, err
			}
			allGroups = append(allGroups, file.Groups...)
		}
	}

	return allGroups, nil
}

func loadTeams(dataDir string) ([]TeamData, error) {
	var allTeams []TeamData

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == "teams.yaml" {
			path := filepath.Join(dataDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			var file TeamsFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				return nil, err
			}
			allTeams = append(allTeams, file.Teams...)
		}
	}

	return allTeams, nil
}

func loadUsers(dataDir string) ([]UserData, error) {
	var allUsers []UserData

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == "users.yaml" {
			path := filepath.Join(dataDir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			var file UsersFile
			if err := yaml.Unmarshal(data, &file); err != nil {
				return nil, err
			}
			allUsers = append(allUsers, file.Users...)
		}
	}

	return allUsers, nil
}

func loadComponents(dataDir string) ([]ComponentData, error) {
	var allComponents []ComponentData

	path := filepath.Join(dataDir, "components.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []ComponentData{}, nil
	}

	// Try wrapped format: { components: [...] }
	var file ComponentsFile
	if err := yaml.Unmarshal(data, &file); err == nil && len(file.Components) > 0 {
		allComponents = append(allComponents, file.Components...)
		return allComponents, nil
	}

	// Fallback: plain array format: [ {...}, {...} ]
	var list []ComponentData
	if err := yaml.Unmarshal(data, &list); err == nil && len(list) > 0 {
		allComponents = append(allComponents, list...)
		return allComponents, nil
	}

	// Neither format matched
	return nil, fmt.Errorf("unrecognized components YAML format: %s", path)
}

func loadProjects(dataDir string) ([]ProjectData, error) {
	path := filepath.Join(dataDir, "projects.yaml")
	var file ProjectsFile
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []ProjectData{}, nil
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return file.Projects, nil
}

func loadLandscapes(dataDir string) ([]LandscapeData, error) {
	path := filepath.Join(dataDir, "landscapes.yaml")
	var file LandscapesFile
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []LandscapeData{}, nil
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return file.Landscapes, nil
}

func loadCategories(dataDir string) ([]CategoryData, error) {
	path := filepath.Join(dataDir, "categories.yaml")
	var file CategoriesFile
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []CategoryData{}, nil
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return file.Categories, nil
}

// NEW: loadLinks reads links.yaml files
func loadLinks(dataDir string) ([]InitialLinkData, error) {
	path := filepath.Join(dataDir, "links.yaml")
	var file LinksFile
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []InitialLinkData{}, nil
	}

	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	// Normalize links: parse tags scalar into array
	for i := range file.Links {
		if len(file.Links[i].Tags) == 0 && file.Links[i].TagsRaw != nil {
			switch v := file.Links[i].TagsRaw.(type) {
			case string:
				parts := strings.Split(v, ",")
				file.Links[i].Tags = make([]string, 0, len(parts))
				for _, p := range parts {
					t := strings.TrimSpace(p)
					if t != "" {
						file.Links[i].Tags = append(file.Links[i].Tags, t)
					}
				}
			case []interface{}:
				file.Links[i].Tags = make([]string, 0, len(v))
				for _, it := range v {
					if s, ok := it.(string); ok && strings.TrimSpace(s) != "" {
						file.Links[i].Tags = append(file.Links[i].Tags, strings.TrimSpace(s))
					}
				}
			}
		}
	}

	return file.Links, nil
}

func slugifyTitle(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
		} else {
			if !prevHyphen {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "link"
	}
	return out
}

// NEW: createLink upserts a Link owned by a team
func createLink(db *gorm.DB, data InitialLinkData) (*models.Link, bool, error) {
	// Resolve owner user by fixed user_id 'cis.devops'
	var owner models.User
	if err := db.Where("user_id = ?", "cis.devops").First(&owner).Error; err != nil {
		return nil, false, fmt.Errorf("owner user with user_id 'cis.devops' not found for link %s: %w", data.Title, err)
	}

	name := slugifyTitle(data.Title)

	// Resolve category UUID by name
	var cat models.Category
	if err := db.Where("name = ?", data.Category).First(&cat).Error; err != nil {
		return nil, false, fmt.Errorf("category %s not found for link %s: %w", data.Category, data.Title, err)
	}
	catID := cat.ID

	var link models.Link
	if err := db.Where("name = ? AND owner = ?", name, owner.ID).First(&link).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tagsCSV := strings.Join(data.Tags, ",")
			link = models.Link{
				BaseModel: models.BaseModel{
					Name:        name,
					Title:       data.Title,
					Description: data.Description,
					CreatedBy:   "cis.devops",
				},
				Owner:      owner.ID,
				URL:        data.URL,
				CategoryID: catID,
				Tags:       tagsCSV,
			}

			if err := db.Create(&link).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create link: %w", err)
			}
			return &link, true, nil
		}
		return nil, false, fmt.Errorf("failed to query link: %w", err)
	}

	// Update mutable fields on existing link
	tagsCSV := strings.Join(data.Tags, ",")
	updates := map[string]interface{}{
		"url":         data.URL,
		"category_id": catID,
		"tags":        tagsCSV,
		"title":       data.Title,
		"description": data.Description,
	}
	if err := db.Model(&link).Updates(updates).Error; err != nil {
		log.Printf("⚠️  Warning: failed to update link %s: %v", data.Title, err)
	} else {
		link.URL = data.URL
		link.CategoryID = catID
		link.Tags = tagsCSV
	}

	return &link, false, nil
}

func createOrganization(db *gorm.DB, orgData OrganizationData) (*models.Organization, bool, error) {
	var org models.Organization
	if err := db.Where("name = ?", orgData.Name).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			org = models.Organization{
				BaseModel: models.BaseModel{
					Name:        orgData.Name,
					Title:       orgData.Title,
					Description: orgData.Description,
					CreatedBy:   "cis.devops",
				},
				Owner: orgData.Owner,
				Email: orgData.Email,
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
	org := orgMap[groupData.OrgName]
	if org == nil {
		return nil, false, fmt.Errorf("organization %s not found for group %s", groupData.OrgName, groupData.Name)
	}

	var group models.Group
	if err := db.Where("name = ? AND org_id = ?", groupData.Name, org.ID).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			group = models.Group{
				BaseModel: models.BaseModel{
					Name:        groupData.Name,
					Title:       groupData.Title,
					Description: groupData.Description,
					CreatedBy:   "cis.devops",
				},
				OrgID:      org.ID,
				Owner:      groupData.Owner,
				Email:      groupData.Email,
				PictureURL: groupData.Picture,
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
			// Validate required fields aligned with Team model
			if teamData.Owner == "" || teamData.Email == "" || teamData.Title == "" {
				return nil, false, fmt.Errorf("missing required team fields (owner, email, title) for team %s", teamData.Name)
			}

			metadataJSON, _ := json.Marshal(teamData.Metadata)
			team = models.Team{
				BaseModel: models.BaseModel{
					Name:        teamData.Name,
					Title:       teamData.Title,
					Description: teamData.Description,
					CreatedBy:   "cis.devops",
					Metadata:    metadataJSON,
				},
				GroupID:    group.ID,
				Owner:      teamData.Owner,
				Email:      teamData.Email,
				PictureURL: teamData.Picture,
			}

			if err := db.Create(&team).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create team: %w", err)
			}
			return &team, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query team: %w", err)
		}
	}

	// If team already exists, update metadata if provided in YAML
	if teamData.Metadata != nil {
		metadataJSON, _ := json.Marshal(teamData.Metadata)
		if err := db.Model(&team).Update("metadata", metadataJSON).Error; err != nil {
			log.Printf("⚠️  Warning: failed to update metadata for team %s: %v", teamData.Name, err)
		} else {
			team.Metadata = metadataJSON
		}
	}
	return &team, false, nil // created = false (existing)
}

func createUser(db *gorm.DB, userData UserData, teamMap map[string]*models.Team) (*models.User, bool, error) {
	team := teamMap[userData.TeamName]

	var teamID *uuid.UUID
	if userData.TeamName != "" && team != nil {
		if team := teamMap[userData.TeamName]; team != nil {
			teamID = &team.ID
		}
	}

	var user models.User
	if err := db.Where("email = ?", userData.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			metadataJSON, _ := json.Marshal(userData.Metadata)
			user = models.User{
				BaseModel: models.BaseModel{
					Name:        userData.UserID,
					Title:       userData.FirstName + " " + userData.LastName,
					Description: "",
					CreatedBy:   "cis.devops",
					Metadata:    metadataJSON,
				},
				UserID:     userData.UserID,
				TeamID:     teamID,
				FirstName:  userData.FirstName,
				LastName:   userData.LastName,
				Email:      userData.Email,
				Mobile:     userData.PhoneNumber,
				TeamDomain: models.TeamDomain(userData.TeamDomain),
				TeamRole:   models.TeamRole(userData.TeamRole),
				Metadata:   metadataJSON,
			}

			if err := db.Create(&user).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create user: %w", err)
			}

			return &user, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query user: %w", err)
		}
	}

	// If user already exists, update metadata if provided in YAML
	if userData.Metadata != nil {
		metadataJSON, _ := json.Marshal(userData.Metadata)

		// Force update by email to ensure row match and bypass any dirty-check issues
		tx := db.Model(&models.User{}).Where("email = ?", userData.Email).UpdateColumn("metadata", metadataJSON)
		if tx.Error != nil {
			log.Printf("⚠️  Warning: failed to update metadata for user %s by email: %v", userData.UserID, tx.Error)
		} else if tx.RowsAffected == 0 {
			log.Printf("⚠️  Warning: metadata update for user %s matched 0 rows by email", userData.UserID)
		} else {
			log.Printf("ℹ️  Updated metadata for user %s (%d rows) by email", userData.UserID, tx.RowsAffected)
			user.Metadata = metadataJSON
		}

		// Also attempt by user_id for completeness
		tx2 := db.Model(&models.User{}).Where("user_id = ?", userData.UserID).UpdateColumn("metadata", metadataJSON)
		if tx2.Error != nil {
			log.Printf("⚠️  Warning: failed to update metadata for user %s by user_id: %v", userData.UserID, tx2.Error)
		}
	}

	return &user, false, nil // created = false (existing)
}

func createComponent(db *gorm.DB, componentData ComponentData, orgMap map[string]*models.Organization, teamMap map[string]*models.Team) (*models.Component, bool, error) {
	// Resolve owner team -> OwnerID
	if componentData.Owner == "" {
		return nil, false, fmt.Errorf("owner team missing for component %s", componentData.Name)
	}
	team := teamMap[componentData.Owner]
	if team == nil {
		return nil, false, fmt.Errorf("owner team %s not found for component %s", componentData.Owner, componentData.Name)
	}
	ownerID := team.ID

	// Require project to exist (do not auto-create; projects seeded only from projects.yaml)
	var proj models.Project
	if err := db.Where("name = ?", componentData.Project).First(&proj).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, fmt.Errorf("project %s not found for component %s; projects must be defined in projects.yaml", componentData.Project, componentData.Name)
		}
		return nil, false, fmt.Errorf("failed to query project %s: %w", componentData.Project, err)
	}

	// Upsert component by name
	var component models.Component
	if err := db.Where("name = ?", componentData.Name).First(&component).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metadataJSON, _ := json.Marshal(componentData.Metadata)
			component = models.Component{
				BaseModel: models.BaseModel{
					Name:        componentData.Name,
					Title:       componentData.Title,
					Description: componentData.Description,
					CreatedBy:   "cis.devops",
					Metadata:    metadataJSON,
				},
				ProjectID: proj.ID,
				OwnerID:   ownerID,
			}
			if err := db.Create(&component).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create component: %w", err)
			}
			return &component, true, nil
		}
		return nil, false, fmt.Errorf("failed to query component: %w", err)
	}

	// Update fields on existing component
	updates := map[string]interface{}{
		"title":       componentData.Title,
		"description": componentData.Description,
		"project_id":  proj.ID,
		"owner_id":    ownerID,
	}
	if componentData.Metadata != nil {
		metadataJSON, _ := json.Marshal(componentData.Metadata)
		updates["metadata"] = metadataJSON
	}
	if err := db.Model(&component).Updates(updates).Error; err != nil {
		log.Printf("⚠️  Warning: failed to update component %s: %v", componentData.Name, err)
	} else {
		component.Title = componentData.Title
		component.Description = componentData.Description
		component.ProjectID = proj.ID
		component.OwnerID = ownerID
	}

	return &component, false, nil
}

func createProject(db *gorm.DB, projectData ProjectData, orgMap map[string]*models.Organization) (*models.Project, bool, error) {
	var project models.Project
	if err := db.Where("name = ?", projectData.Name).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			metadataJSON, _ := json.Marshal(projectData.Metadata)

			project = models.Project{
				BaseModel: models.BaseModel{
					Name:        projectData.Name,
					Title:       projectData.Title,
					Description: projectData.Description,
					CreatedBy:   "cis.devops",
					Metadata:    metadataJSON,
				},
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

func createLandscapeFromYAML(db *gorm.DB, item LandscapeData, orgID uuid.UUID) (*models.Landscape, bool, error) {
	// Merge domain into metadata
	md := item.Metadata
	if md == nil {
		md = map[string]interface{}{}
	}
	metadataJSON, _ := json.Marshal(md)

	env := strings.ToLower(item.Environment)
	var proj models.Project
	if err := db.Where("name = ?", item.Project).First(&proj).Error; err != nil {
		return nil, false, fmt.Errorf("project %s not found for landscape %s: %w", item.Project, item.Name, err)
	}

	var landscape models.Landscape
	if err := db.Where("name = ?", item.Name).First(&landscape).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			landscape = models.Landscape{
				BaseModel: models.BaseModel{
					Name:        item.Name,
					Title:       item.Title,
					Description: item.Description,
					CreatedBy:   "cis.devops",
					Metadata:    metadataJSON,
				},
				ProjectID:   proj.ID,
				Domain:      item.Domain,
				Environment: env,
			}
			if err := db.Create(&landscape).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create landscape: %w", err)
			}
			return &landscape, true, nil
		}
		return nil, false, fmt.Errorf("failed to query landscape: %w", err)
	}

	// Update mutable fields on existing landscape
	updates := map[string]interface{}{
		"title":       item.Title,
		"description": item.Description,
		"domain":      item.Domain,
		"environment": env,
		"project_id":  proj.ID,
		"metadata":    metadataJSON,
	}
	if err := db.Model(&landscape).Updates(updates).Error; err != nil {
		log.Printf("⚠️  Warning: failed to update landscape %s: %v", item.Name, err)
	} else {
		landscape.Title = item.Title
		landscape.Description = item.Description
		landscape.Domain = item.Domain
		landscape.Environment = env
		landscape.ProjectID = proj.ID
		landscape.Metadata = metadataJSON
	}

	return &landscape, false, nil
}

func createCategory(db *gorm.DB, catData CategoryData) (*models.Category, bool, error) {
	var cat models.Category
	if err := db.Where("name = ?", catData.Name).First(&cat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			cat = models.Category{
				BaseModel: models.BaseModel{
					Name:        catData.Name,
					Title:       catData.Title,
					Description: "",
					CreatedBy:   "cis.devops",
				},
				Icon:  catData.Icon,
				Color: catData.Color,
			}

			if err := db.Create(&cat).Error; err != nil {
				return nil, false, fmt.Errorf("failed to create category: %w", err)
			}
			return &cat, true, nil // created = true
		} else {
			return nil, false, fmt.Errorf("failed to query category: %w", err)
		}
	}

	// Update mutable fields on existing category
	updates := map[string]interface{}{
		"title": catData.Title,
		"icon":  catData.Icon,
		"color": catData.Color,
	}
	if err := db.Model(&cat).Updates(updates).Error; err != nil {
		log.Printf("⚠️  Warning: failed to update category %s: %v", catData.Name, err)
	} else {
		cat.Title = catData.Title
		cat.Icon = catData.Icon
		cat.Color = catData.Color
	}

	return &cat, false, nil
}
