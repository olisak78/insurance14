package service

// LinkCategory represents a category for organizing links
type LinkCategory struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IconName string `json:"iconName"`
	Color    string `json:"color"`
}

// Link represents a useful link/resource
type Link struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	CategoryID  string   `json:"categoryId"`
	Tags        []string `json:"tags"`
	Favorite    bool     `json:"favorite"`
}

// LinksResponse represents the response structure for links
type LinksResponse struct {
	Links      []Link         `json:"links"`
	Categories []LinkCategory `json:"categories"`
}

// LinkService handles link-related business logic
type LinkService struct{}

// NewLinkService creates a new link service
func NewLinkService() *LinkService {
	return &LinkService{}
}

// GetMockCategories returns mock category data
func (s *LinkService) GetMockCategories() []LinkCategory {
	return []LinkCategory{
		{
			ID:       "ci-cd",
			Name:     "CI/CD & Build",
			IconName: "Code",
			Color:    "bg-blue-500",
		},
		{
			ID:       "security",
			Name:     "Security & Compliance",
			IconName: "Shield",
			Color:    "bg-red-500",
		},
		{
			ID:       "monitoring",
			Name:     "Monitoring & Observability",
			IconName: "Monitor",
			Color:    "bg-green-500",
		},
		{
			ID:       "project",
			Name:     "Project Management",
			IconName: "Users",
			Color:    "bg-purple-500",
		},
		{
			ID:       "documentation",
			Name:     "Documentation & Knowledge",
			IconName: "FileText",
			Color:    "bg-amber-500",
		},
		{
			ID:       "development",
			Name:     "Development Tools",
			IconName: "Wrench",
			Color:    "bg-indigo-500",
		},
		{
			ID:       "infrastructure",
			Name:     "Infrastructure & Cloud",
			IconName: "Cloud",
			Color:    "bg-cyan-500",
		},
		{
			ID:       "testing",
			Name:     "Testing & QA",
			IconName: "TestTube",
			Color:    "bg-emerald-500",
		},
		{
			ID:       "community",
			Name:     "Community & Support",
			IconName: "HelpCircle",
			Color:    "bg-orange-500",
		},
	}
}

// GetMockLinks returns mock link data
func (s *LinkService) GetMockLinks() []Link {
	return []Link{
		{
			ID:          "1",
			Title:       "JaaS Status",
			URL:         "https://me.sap.com/cacv2/customer/2029347",
			Description: "Java as a Service status dashboard",
			CategoryID:  "ci-cd",
			Tags:        []string{"jaas", "status", "dashboard"},
			Favorite:    false,
		},
		{
			ID:          "2",
			Title:       "xMake Nova",
			URL:         "https://xmake-nova.wdf.sap.corp",
			Description: "Next-generation build system",
			CategoryID:  "ci-cd",
			Tags:        []string{"build", "xmake", "nova"},
			Favorite:    false,
		},
		{
			ID:          "3",
			Title:       "CIS 2.0",
			URL:         "https://gkecfsmicroservice.jaas-gcp.cloud.sap.corp/",
			Description: "Modern microservice CI system",
			CategoryID:  "ci-cd",
			Tags:        []string{"cis", "microservice", "ci"},
			Favorite:    false,
		},
		{
			ID:          "4",
			Title:       "Sonar",
			URL:         "https://sonar.tools.sap/",
			Description: "Code quality and security analysis",
			CategoryID:  "security",
			Tags:        []string{"quality", "security", "analysis"},
			Favorite:    false,
		},
		{
			ID:          "5",
			Title:       "Fortify",
			URL:         "https://fortify.tools.sap/ssc",
			Description: "Static application security testing",
			CategoryID:  "security",
			Tags:        []string{"security", "testing", "static"},
			Favorite:    false,
		},
		{
			ID:          "6",
			Title:       "Vault",
			URL:         "https://vault.tools.sap/ui/vault/auth?with=oidc%2F",
			Description: "Secrets management",
			CategoryID:  "security",
			Tags:        []string{"secrets", "vault", "management"},
			Favorite:    false,
		},
		{
			ID:          "7",
			Title:       "Copernicus",
			URL:         "https://copernicus.cfapps.sap.hana.ondemand.com/",
			Description: "Application monitoring dashboard",
			CategoryID:  "monitoring",
			Tags:        []string{"monitoring", "dashboard", "application"},
			Favorite:    false,
		},
		{
			ID:          "8",
			Title:       "Cloud Availability Center",
			URL:         "https://me.sap.com/systemsprovisioning/availability",
			Description: "Cloud service availability dashboard",
			CategoryID:  "monitoring",
			Tags:        []string{"availability", "cloud", "dashboard"},
			Favorite:    false,
		},
		{
			ID:          "9",
			Title:       "Splunk",
			URL:         "https://portal.victorops.com/",
			Description: "Log analysis and monitoring",
			CategoryID:  "monitoring",
			Tags:        []string{"logs", "analysis", "monitoring"},
			Favorite:    false,
		},
		{
			ID:          "10",
			Title:       "Jira Tools",
			URL:         "https://jira.tools.sap/",
			Description: "Issue tracking and project management",
			CategoryID:  "project",
			Tags:        []string{"jira", "issues", "tracking"},
			Favorite:    false,
		},
		{
			ID:          "11",
			Title:       "Project Portal",
			URL:         "https://projectportal.tools.sap/",
			Description: "Project management and tracking",
			CategoryID:  "project",
			Tags:        []string{"project", "portal", "management"},
			Favorite:    false,
		},
		{
			ID:          "12",
			Title:       "SNOW",
			URL:         "https://itsm.services.sap/sp",
			Description: "ServiceNow ITSM platform",
			CategoryID:  "project",
			Tags:        []string{"servicenow", "itsm", "platform"},
			Favorite:    false,
		},
		{
			ID:          "13",
			Title:       "CIS Delivery Wiki",
			URL:         "https://wiki.wdf.sap.corp/wiki/pages/viewpage.action?pageId=1893349895#Development,CI&Delivery(CIS/CISAdmin/CISConfig)-Delivery",
			Description: "CIS delivery process documentation",
			CategoryID:  "documentation",
			Tags:        []string{"wiki", "cis", "delivery"},
			Favorite:    false,
		},
		{
			ID:          "14",
			Title:       "On Duty Tasks",
			URL:         "https://wiki.wdf.sap.corp/wiki/x/6_jyc",
			Description: "On-duty task documentation",
			CategoryID:  "documentation",
			Tags:        []string{"duty", "tasks", "documentation"},
			Favorite:    false,
		},
		{
			ID:          "15",
			Title:       "Cloud Engineer OnDuty",
			URL:         "https://wiki.wdf.sap.corp/wiki/display/CloudEng/Cloud+Engineer+on+Duty",
			Description: "Cloud engineer on-duty procedures",
			CategoryID:  "documentation",
			Tags:        []string{"cloud", "engineer", "procedures"},
			Favorite:    false,
		},
		{
			ID:          "16",
			Title:       "GitHub Tools",
			URL:         "https://github.tools.sap/",
			Description: "GitHub tools and utilities",
			CategoryID:  "development",
			Tags:        []string{"github", "tools", "utilities"},
			Favorite:    false,
		},
		{
			ID:          "17",
			Title:       "Artifactory (internal)",
			URL:         "https://int.repositories.cloud.sap/ui",
			Description: "Internal artifact repository",
			CategoryID:  "development",
			Tags:        []string{"artifactory", "repository", "internal"},
			Favorite:    false,
		},
		{
			ID:          "18",
			Title:       "SAP API Business Hub",
			URL:         "https://api.sap.com/package/SAPCloudPlatformCoreServices",
			Description: "API documentation and testing",
			CategoryID:  "development",
			Tags:        []string{"api", "documentation", "testing"},
			Favorite:    false,
		},
		{
			ID:          "19",
			Title:       "Gardener - Live",
			URL:         "https://dashboard.garden.live.k8s.ondemand.com/",
			Description: "Kubernetes gardener live environment",
			CategoryID:  "infrastructure",
			Tags:        []string{"kubernetes", "gardener", "live"},
			Favorite:    false,
		},
		{
			ID:          "20",
			Title:       "Hyperspace Portal",
			URL:         "https://portal.hyperspace.tools.sap",
			Description: "Hyperspace platform portal",
			CategoryID:  "infrastructure",
			Tags:        []string{"hyperspace", "platform", "portal"},
			Favorite:    false,
		},
		{
			ID:          "21",
			Title:       "Converged Cloud",
			URL:         "https://dashboard.eu-de-2.cloud.sap/monsoon3/home",
			Description: "Converged cloud management dashboard",
			CategoryID:  "infrastructure",
			Tags:        []string{"cloud", "converged", "dashboard"},
			Favorite:    false,
		},
		{
			ID:          "22",
			Title:       "E2E CLI Staging ALI",
			URL:         "https://gkecfsautomation.jaas-gcp.cloud.sap.corp/job/cli-e2e-tests-on-stagingac/",
			Description: "End-to-end CLI testing on staging ALI",
			CategoryID:  "testing",
			Tags:        []string{"e2e", "cli", "staging"},
			Favorite:    false,
		},
		{
			ID:          "23",
			Title:       "SAP@Stackoverflow",
			URL:         "https://sap.stackenterprise.co/",
			Description: "SAP Stack Overflow community",
			CategoryID:  "community",
			Tags:        []string{"stackoverflow", "community", "support"},
			Favorite:    false,
		},
	}
}

// GetLinksByMemberID returns all links for a member (currently returns all mock links)
func (s *LinkService) GetLinksByMemberID(memberID string) *LinksResponse {
	return &LinksResponse{
		Links:      s.GetMockLinks(),
		Categories: s.GetMockCategories(),
	}
}