package service_test

import (
	"testing"

	"developer-portal-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// LinkServiceTestSuite defines the test suite for LinkService
type LinkServiceTestSuite struct {
	suite.Suite
	service *service.LinkService
}

// SetupTest sets up the test suite
func (suite *LinkServiceTestSuite) SetupTest() {
	suite.service = service.NewLinkService()
}

// TestNewLinkService tests the service constructor
func (suite *LinkServiceTestSuite) TestNewLinkService() {
	service := service.NewLinkService()
	assert.NotNil(suite.T(), service, "LinkService should not be nil")
}

// TestGetMockCategories tests retrieving mock categories
func (suite *LinkServiceTestSuite) TestGetMockCategories() {
	categories := suite.service.GetMockCategories()

	// Verify count
	assert.Equal(suite.T(), 9, len(categories), "Should return 9 categories")

	// Verify all categories have required fields
	for i, category := range categories {
		assert.NotEmpty(suite.T(), category.ID, "Category %d should have an ID", i)
		assert.NotEmpty(suite.T(), category.Name, "Category %d should have a name", i)
		assert.NotEmpty(suite.T(), category.IconName, "Category %d should have an icon name", i)
		assert.NotEmpty(suite.T(), category.Color, "Category %d should have a color", i)
	}

	// Verify specific categories
	expectedCategories := map[string]struct {
		name     string
		iconName string
		color    string
	}{
		"ci-cd":          {"CI/CD & Build", "Code", "bg-blue-500"},
		"security":       {"Security & Compliance", "Shield", "bg-red-500"},
		"monitoring":     {"Monitoring & Observability", "Monitor", "bg-green-500"},
		"project":        {"Project Management", "Users", "bg-purple-500"},
		"documentation":  {"Documentation & Knowledge", "FileText", "bg-amber-500"},
		"development":    {"Development Tools", "Wrench", "bg-indigo-500"},
		"infrastructure": {"Infrastructure & Cloud", "Cloud", "bg-cyan-500"},
		"testing":        {"Testing & QA", "TestTube", "bg-emerald-500"},
		"community":      {"Community & Support", "HelpCircle", "bg-orange-500"},
	}

	for _, category := range categories {
		expected, exists := expectedCategories[category.ID]
		assert.True(suite.T(), exists, "Category ID %s should be in expected list", category.ID)
		if exists {
			assert.Equal(suite.T(), expected.name, category.Name, "Category %s name mismatch", category.ID)
			assert.Equal(suite.T(), expected.iconName, category.IconName, "Category %s icon name mismatch", category.ID)
			assert.Equal(suite.T(), expected.color, category.Color, "Category %s color mismatch", category.ID)
		}
	}
}

// TestGetMockLinks tests retrieving mock links
func (suite *LinkServiceTestSuite) TestGetMockLinks() {
	links := suite.service.GetMockLinks()

	// Verify count
	assert.Equal(suite.T(), 23, len(links), "Should return 23 links")

	// Verify all links have required fields
	for i, link := range links {
		assert.NotEmpty(suite.T(), link.ID, "Link %d should have an ID", i)
		assert.NotEmpty(suite.T(), link.Title, "Link %d should have a title", i)
		assert.NotEmpty(suite.T(), link.URL, "Link %d should have a URL", i)
		assert.NotEmpty(suite.T(), link.Description, "Link %d should have a description", i)
		assert.NotEmpty(suite.T(), link.CategoryID, "Link %d should have a category ID", i)
		assert.NotNil(suite.T(), link.Tags, "Link %d should have tags array (can be empty)", i)
	}
}

// TestGetMockLinks_SpecificLinks tests specific link content
func (suite *LinkServiceTestSuite) TestGetMockLinks_SpecificLinks() {
	links := suite.service.GetMockLinks()

	// Test first link (JaaS Status)
	firstLink := links[0]
	assert.Equal(suite.T(), "1", firstLink.ID)
	assert.Equal(suite.T(), "JaaS Status", firstLink.Title)
	assert.Equal(suite.T(), "https://me.sap.com/cacv2/customer/2029347", firstLink.URL)
	assert.Equal(suite.T(), "Java as a Service status dashboard", firstLink.Description)
	assert.Equal(suite.T(), "ci-cd", firstLink.CategoryID)
	assert.Equal(suite.T(), []string{"jaas", "status", "dashboard"}, firstLink.Tags)
	assert.False(suite.T(), firstLink.Favorite)

	// Test a security link (Sonar)
	sonarLink := links[3] // ID "4"
	assert.Equal(suite.T(), "4", sonarLink.ID)
	assert.Equal(suite.T(), "Sonar", sonarLink.Title)
	assert.Equal(suite.T(), "security", sonarLink.CategoryID)

	// Test last link (SAP@Stackoverflow)
	lastLink := links[22]
	assert.Equal(suite.T(), "23", lastLink.ID)
	assert.Equal(suite.T(), "SAP@Stackoverflow", lastLink.Title)
	assert.Equal(suite.T(), "community", lastLink.CategoryID)
}

// TestGetMockLinks_CategoryDistribution tests that links are distributed across categories
func (suite *LinkServiceTestSuite) TestGetMockLinks_CategoryDistribution() {
	links := suite.service.GetMockLinks()
	categories := suite.service.GetMockCategories()

	// Count links per category
	categoryCount := make(map[string]int)
	for _, link := range links {
		categoryCount[link.CategoryID]++
	}

	// Verify all categories are represented
	for _, category := range categories {
		count := categoryCount[category.ID]
		assert.Greater(suite.T(), count, 0, "Category %s should have at least one link", category.ID)
	}

	// Verify specific counts
	assert.Equal(suite.T(), 3, categoryCount["ci-cd"], "CI/CD category should have 3 links")
	assert.Equal(suite.T(), 3, categoryCount["security"], "Security category should have 3 links")
	assert.Equal(suite.T(), 3, categoryCount["monitoring"], "Monitoring category should have 3 links")
	assert.Equal(suite.T(), 3, categoryCount["project"], "Project category should have 3 links")
	assert.Equal(suite.T(), 3, categoryCount["documentation"], "Documentation category should have 3 links")
	assert.Equal(suite.T(), 3, categoryCount["development"], "Development category should have 3 links")
	assert.Equal(suite.T(), 3, categoryCount["infrastructure"], "Infrastructure category should have 3 links")
	assert.Equal(suite.T(), 1, categoryCount["testing"], "Testing category should have 1 link")
	assert.Equal(suite.T(), 1, categoryCount["community"], "Community category should have 1 link")
}

// TestGetMockLinks_URLValidation tests that all URLs are properly formatted
func (suite *LinkServiceTestSuite) TestGetMockLinks_URLValidation() {
	links := suite.service.GetMockLinks()

	for _, link := range links {
		assert.Contains(suite.T(), link.URL, "://", "Link %s URL should contain protocol separator", link.ID)
		assert.True(suite.T(), 
			len(link.URL) > 10, 
			"Link %s URL should be a reasonable length", link.ID)
	}
}

// TestGetLinksByMemberID tests the main service method
func (suite *LinkServiceTestSuite) TestGetLinksByMemberID() {
	memberID := "123e4567-e89b-12d3-a456-426614174000"
	
	response := suite.service.GetLinksByMemberID(memberID)

	// Verify response is not nil
	assert.NotNil(suite.T(), response, "Response should not be nil")

	// Verify links array
	assert.NotNil(suite.T(), response.Links, "Links array should not be nil")
	assert.Equal(suite.T(), 23, len(response.Links), "Should return 23 links")

	// Verify categories array
	assert.NotNil(suite.T(), response.Categories, "Categories array should not be nil")
	assert.Equal(suite.T(), 9, len(response.Categories), "Should return 9 categories")
}

// TestGetLinksByMemberID_DifferentMemberIDs tests that different member IDs return same data
func (suite *LinkServiceTestSuite) TestGetLinksByMemberID_DifferentMemberIDs() {
	memberIDs := []string{
		"123e4567-e89b-12d3-a456-426614174000",
		"987e6543-e21b-34d5-a678-426614174111",
		"any-random-id",
		"",
	}

	var previousResponse *service.LinksResponse

	for _, memberID := range memberIDs {
		response := suite.service.GetLinksByMemberID(memberID)
		
		assert.NotNil(suite.T(), response, "Response should not be nil for member ID: %s", memberID)
		assert.Equal(suite.T(), 23, len(response.Links), "Should return 23 links for member ID: %s", memberID)
		assert.Equal(suite.T(), 9, len(response.Categories), "Should return 9 categories for member ID: %s", memberID)

		// Verify consistency across different member IDs
		if previousResponse != nil {
			assert.Equal(suite.T(), len(previousResponse.Links), len(response.Links), 
				"Link count should be consistent across member IDs")
			assert.Equal(suite.T(), len(previousResponse.Categories), len(response.Categories), 
				"Category count should be consistent across member IDs")
		}

		previousResponse = response
	}
}

// TestGetLinksByMemberID_ResponseStructure tests the complete response structure
func (suite *LinkServiceTestSuite) TestGetLinksByMemberID_ResponseStructure() {
	response := suite.service.GetLinksByMemberID("test-member-id")

	// Verify top-level structure
	assert.NotNil(suite.T(), response.Links)
	assert.NotNil(suite.T(), response.Categories)

	// Verify first link has all fields
	if len(response.Links) > 0 {
		firstLink := response.Links[0]
		assert.NotEmpty(suite.T(), firstLink.ID)
		assert.NotEmpty(suite.T(), firstLink.Title)
		assert.NotEmpty(suite.T(), firstLink.URL)
		assert.NotEmpty(suite.T(), firstLink.Description)
		assert.NotEmpty(suite.T(), firstLink.CategoryID)
		assert.NotNil(suite.T(), firstLink.Tags)
		// Favorite is a boolean, so just check it's set
		_ = firstLink.Favorite
	}

	// Verify first category has all fields
	if len(response.Categories) > 0 {
		firstCategory := response.Categories[0]
		assert.NotEmpty(suite.T(), firstCategory.ID)
		assert.NotEmpty(suite.T(), firstCategory.Name)
		assert.NotEmpty(suite.T(), firstCategory.IconName)
		assert.NotEmpty(suite.T(), firstCategory.Color)
	}
}

// TestGetLinksByMemberID_CategoryLinkRelationship tests that all link categories exist
func (suite *LinkServiceTestSuite) TestGetLinksByMemberID_CategoryLinkRelationship() {
	response := suite.service.GetLinksByMemberID("test-member-id")

	// Build a map of category IDs
	categoryIDs := make(map[string]bool)
	for _, category := range response.Categories {
		categoryIDs[category.ID] = true
	}

	// Verify all links reference valid categories
	for _, link := range response.Links {
		assert.True(suite.T(), categoryIDs[link.CategoryID], 
			"Link %s references category %s which should exist", link.ID, link.CategoryID)
	}
}

// TestLinkStructure tests the Link struct
func (suite *LinkServiceTestSuite) TestLinkStructure() {
	link := service.Link{
		ID:          "test-id",
		Title:       "Test Link",
		URL:         "https://example.com",
		Description: "Test description",
		CategoryID:  "test-category",
		Tags:        []string{"tag1", "tag2"},
		Favorite:    true,
	}

	assert.Equal(suite.T(), "test-id", link.ID)
	assert.Equal(suite.T(), "Test Link", link.Title)
	assert.Equal(suite.T(), "https://example.com", link.URL)
	assert.Equal(suite.T(), "Test description", link.Description)
	assert.Equal(suite.T(), "test-category", link.CategoryID)
	assert.Equal(suite.T(), []string{"tag1", "tag2"}, link.Tags)
	assert.True(suite.T(), link.Favorite)
}

// TestLinkCategoryStructure tests the LinkCategory struct
func (suite *LinkServiceTestSuite) TestLinkCategoryStructure() {
	category := service.LinkCategory{
		ID:       "test-id",
		Name:     "Test Category",
		IconName: "TestIcon",
		Color:    "bg-test-500",
	}

	assert.Equal(suite.T(), "test-id", category.ID)
	assert.Equal(suite.T(), "Test Category", category.Name)
	assert.Equal(suite.T(), "TestIcon", category.IconName)
	assert.Equal(suite.T(), "bg-test-500", category.Color)
}

// TestLinksResponseStructure tests the LinksResponse struct
func (suite *LinkServiceTestSuite) TestLinksResponseStructure() {
	links := []service.Link{
		{ID: "1", Title: "Link 1", URL: "https://link1.com", Description: "Desc 1", CategoryID: "cat1", Tags: []string{"tag1"}, Favorite: false},
	}
	categories := []service.LinkCategory{
		{ID: "cat1", Name: "Category 1", IconName: "Icon1", Color: "bg-blue-500"},
	}

	response := service.LinksResponse{
		Links:      links,
		Categories: categories,
	}

	assert.Equal(suite.T(), 1, len(response.Links))
	assert.Equal(suite.T(), 1, len(response.Categories))
	assert.Equal(suite.T(), "1", response.Links[0].ID)
	assert.Equal(suite.T(), "cat1", response.Categories[0].ID)
}

// TestInSuite runs all the tests in the suite
func TestLinkServiceTestSuite(t *testing.T) {
	suite.Run(t, new(LinkServiceTestSuite))
}