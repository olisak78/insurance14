package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// LinkHandlerTestSuite defines the test suite for LinkHandler
type LinkHandlerTestSuite struct {
	suite.Suite
	handler *handlers.LinkHandler
	service *service.LinkService
	router  *gin.Engine
}

// SetupTest sets up the test suite
func (suite *LinkHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	// Initialize service and handler
	suite.service = service.NewLinkService()
	suite.handler = handlers.NewLinkHandler(suite.service)

	// Setup router
	suite.router = gin.New()
	suite.setupRoutes()
}

// setupRoutes sets up the routes for testing
func (suite *LinkHandlerTestSuite) setupRoutes() {
	suite.router.GET("/links/:id", suite.handler.GetLinksByMemberID)
}

// TestNewLinkHandler tests the handler constructor
func (suite *LinkHandlerTestSuite) TestNewLinkHandler() {
	handler := handlers.NewLinkHandler(suite.service)
	assert.NotNil(suite.T(), handler, "LinkHandler should not be nil")
}

// TestGetLinksByMemberID_Success tests successful retrieval of links
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_Success() {
	// Create request
	memberID := "123e4567-e89b-12d3-a456-426614174000"
	req, _ := http.NewRequest("GET", "/links/"+memberID, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code, "Should return 200 OK")

	// Parse response
	var response service.LinksResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err, "Should be able to parse response")

	// Verify response structure
	assert.NotNil(suite.T(), response.Links, "Links should not be nil")
	assert.NotNil(suite.T(), response.Categories, "Categories should not be nil")
	assert.Equal(suite.T(), 23, len(response.Links), "Should return 23 links")
	assert.Equal(suite.T(), 9, len(response.Categories), "Should return 9 categories")
}

// TestGetLinksByMemberID_ResponseStructure tests the response JSON structure
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_ResponseStructure() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Parse as generic map to verify structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Verify top-level keys exist
	assert.Contains(suite.T(), response, "links", "Response should contain 'links' key")
	assert.Contains(suite.T(), response, "categories", "Response should contain 'categories' key")

	// Verify links is an array
	links, ok := response["links"].([]interface{})
	assert.True(suite.T(), ok, "links should be an array")
	assert.Equal(suite.T(), 23, len(links), "Should have 23 links")

	// Verify categories is an array
	categories, ok := response["categories"].([]interface{})
	assert.True(suite.T(), ok, "categories should be an array")
	assert.Equal(suite.T(), 9, len(categories), "Should have 9 categories")

	// Verify first link structure
	if len(links) > 0 {
		firstLink, ok := links[0].(map[string]interface{})
		assert.True(suite.T(), ok, "First link should be an object")
		assert.Contains(suite.T(), firstLink, "id")
		assert.Contains(suite.T(), firstLink, "title")
		assert.Contains(suite.T(), firstLink, "url")
		assert.Contains(suite.T(), firstLink, "description")
		assert.Contains(suite.T(), firstLink, "categoryId")
		assert.Contains(suite.T(), firstLink, "tags")
		assert.Contains(suite.T(), firstLink, "favorite")
	}

	// Verify first category structure
	if len(categories) > 0 {
		firstCategory, ok := categories[0].(map[string]interface{})
		assert.True(suite.T(), ok, "First category should be an object")
		assert.Contains(suite.T(), firstCategory, "id")
		assert.Contains(suite.T(), firstCategory, "name")
		assert.Contains(suite.T(), firstCategory, "iconName")
		assert.Contains(suite.T(), firstCategory, "color")
	}
}

// TestGetLinksByMemberID_ContentType tests the response content type
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_ContentType() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json; charset=utf-8", w.Header().Get("Content-Type"), 
		"Content-Type should be application/json")
}

// TestGetLinksByMemberID_DifferentMemberIDs tests with various member IDs
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_DifferentMemberIDs() {
	testCases := []struct {
		name     string
		memberID string
	}{
		{
			name:     "UUID format",
			memberID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:     "Simple numeric ID",
			memberID: "123",
		},
		{
			name:     "Alphanumeric ID",
			memberID: "member-abc-123",
		},
		{
			name:     "Long ID",
			memberID: "very-long-member-id-with-many-characters-12345678910",
		},
		{
			name:     "Empty string",
			memberID: "",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/links/"+tc.memberID, nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK for member ID: %s", tc.memberID)

			var response service.LinksResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Should be able to parse response for member ID: %s", tc.memberID)

			// All member IDs should return the same mock data
			assert.Equal(t, 23, len(response.Links), "Should return 23 links for member ID: %s", tc.memberID)
			assert.Equal(t, 9, len(response.Categories), "Should return 9 categories for member ID: %s", tc.memberID)
		})
	}
}

// TestGetLinksByMemberID_ResponseConsistency tests that multiple calls return consistent data
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_ResponseConsistency() {
	memberID := "test-member-id"
	
	// Make multiple requests
	responses := make([]service.LinksResponse, 3)
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/links/"+memberID, nil)
		w := httptest.NewRecorder()
		
		suite.router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		err := json.Unmarshal(w.Body.Bytes(), &responses[i])
		assert.NoError(suite.T(), err)
	}

	// Verify all responses are consistent
	for i := 1; i < len(responses); i++ {
		assert.Equal(suite.T(), len(responses[0].Links), len(responses[i].Links), 
			"Link count should be consistent across requests")
		assert.Equal(suite.T(), len(responses[0].Categories), len(responses[i].Categories), 
			"Category count should be consistent across requests")
		
		// Verify first link is the same
		if len(responses[0].Links) > 0 && len(responses[i].Links) > 0 {
			assert.Equal(suite.T(), responses[0].Links[0].ID, responses[i].Links[0].ID, 
				"First link ID should be consistent")
			assert.Equal(suite.T(), responses[0].Links[0].Title, responses[i].Links[0].Title, 
				"First link title should be consistent")
		}
	}
}

// TestGetLinksByMemberID_SpecificLinkData tests that specific expected links are present
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_SpecificLinkData() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	var response service.LinksResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Create a map of links by ID for easy lookup
	linksByID := make(map[string]service.Link)
	for _, link := range response.Links {
		linksByID[link.ID] = link
	}

	// Test specific links
	testCases := []struct {
		id          string
		title       string
		categoryID  string
		shouldExist bool
	}{
		{"1", "JaaS Status", "ci-cd", true},
		{"4", "Sonar", "security", true},
		{"10", "Jira Tools", "project", true},
		{"23", "SAP@Stackoverflow", "community", true},
		{"999", "Non-existent", "none", false},
	}

	for _, tc := range testCases {
		link, exists := linksByID[tc.id]
		if tc.shouldExist {
			assert.True(suite.T(), exists, "Link with ID %s should exist", tc.id)
			if exists {
				assert.Equal(suite.T(), tc.title, link.Title, "Link %s title mismatch", tc.id)
				assert.Equal(suite.T(), tc.categoryID, link.CategoryID, "Link %s category mismatch", tc.id)
			}
		} else {
			assert.False(suite.T(), exists, "Link with ID %s should not exist", tc.id)
		}
	}
}

// TestGetLinksByMemberID_SpecificCategoryData tests that specific expected categories are present
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_SpecificCategoryData() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	var response service.LinksResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Create a map of categories by ID for easy lookup
	categoriesByID := make(map[string]service.LinkCategory)
	for _, category := range response.Categories {
		categoriesByID[category.ID] = category
	}

	// Test specific categories
	testCases := []struct {
		id          string
		name        string
		iconName    string
		color       string
		shouldExist bool
	}{
		{"ci-cd", "CI/CD & Build", "Code", "bg-blue-500", true},
		{"security", "Security & Compliance", "Shield", "bg-red-500", true},
		{"monitoring", "Monitoring & Observability", "Monitor", "bg-green-500", true},
		{"project", "Project Management", "Users", "bg-purple-500", true},
		{"documentation", "Documentation & Knowledge", "FileText", "bg-amber-500", true},
		{"development", "Development Tools", "Wrench", "bg-indigo-500", true},
		{"infrastructure", "Infrastructure & Cloud", "Cloud", "bg-cyan-500", true},
		{"testing", "Testing & QA", "TestTube", "bg-emerald-500", true},
		{"community", "Community & Support", "HelpCircle", "bg-orange-500", true},
		{"non-existent", "Non-existent Category", "None", "bg-none", false},
	}

	for _, tc := range testCases {
		category, exists := categoriesByID[tc.id]
		if tc.shouldExist {
			assert.True(suite.T(), exists, "Category with ID %s should exist", tc.id)
			if exists {
				assert.Equal(suite.T(), tc.name, category.Name, "Category %s name mismatch", tc.id)
				assert.Equal(suite.T(), tc.iconName, category.IconName, "Category %s icon name mismatch", tc.id)
				assert.Equal(suite.T(), tc.color, category.Color, "Category %s color mismatch", tc.id)
			}
		} else {
			assert.False(suite.T(), exists, "Category with ID %s should not exist", tc.id)
		}
	}
}

// TestGetLinksByMemberID_LinkCategoryRelationship tests that all links reference valid categories
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_LinkCategoryRelationship() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	var response service.LinksResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Build a set of valid category IDs
	validCategoryIDs := make(map[string]bool)
	for _, category := range response.Categories {
		validCategoryIDs[category.ID] = true
	}

	// Verify all links reference valid categories
	for _, link := range response.Links {
		assert.True(suite.T(), validCategoryIDs[link.CategoryID], 
			"Link %s (ID: %s) references invalid category: %s", 
			link.Title, link.ID, link.CategoryID)
	}
}

// TestGetLinksByMemberID_ValidJSON tests that the response is valid JSON
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_ValidJSON() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify it's valid JSON by unmarshaling into interface{}
	var jsonData interface{}
	err := json.Unmarshal(w.Body.Bytes(), &jsonData)
	assert.NoError(suite.T(), err, "Response should be valid JSON")

	// Verify it's an object (not an array)
	_, ok := jsonData.(map[string]interface{})
	assert.True(suite.T(), ok, "Response should be a JSON object")
}

// TestGetLinksByMemberID_NoExtraFields tests that response has only expected fields
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_NoExtraFields() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Should only have "links" and "categories" keys
	assert.Equal(suite.T(), 2, len(response), "Response should only have 2 fields")
	assert.Contains(suite.T(), response, "links")
	assert.Contains(suite.T(), response, "categories")
}

// TestGetLinksByMemberID_HTTPMethodsNotAllowed tests that only GET is allowed
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_HTTPMethodsNotAllowed() {
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		suite.T().Run(method, func(t *testing.T) {
			req, _ := http.NewRequest(method, "/links/test-member-id", nil)
			w := httptest.NewRecorder()

			suite.router.ServeHTTP(w, req)

			// Should return 404 since route is not defined for these methods
			assert.Equal(t, http.StatusNotFound, w.Code, 
				"%s method should not be allowed", method)
		})
	}
}

// TestGetLinksByMemberID_EmptyResponseCheck ensures arrays are never nil
func (suite *LinkHandlerTestSuite) TestGetLinksByMemberID_EmptyResponseCheck() {
	req, _ := http.NewRequest("GET", "/links/test-member-id", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	var response service.LinksResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Arrays should not be nil (even if empty, they should be empty arrays)
	assert.NotNil(suite.T(), response.Links, "Links array should not be nil")
	assert.NotNil(suite.T(), response.Categories, "Categories array should not be nil")
}

// TestInSuite runs all the tests in the suite
func TestLinkHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LinkHandlerTestSuite))
}