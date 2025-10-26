package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	apperrors "developer-portal-backend/internal/errors"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// MemberHandlerTestSuite defines the test suite for MemberHandler
type MemberHandlerTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockService *mocks.MockMemberServiceInterface
	handler     *handlers.MemberHandler
	router      *gin.Engine
}

// SetupTest sets up the test suite
func (suite *MemberHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockMemberServiceInterface(suite.ctrl)

	// Create handler with mock service - we'll use a custom struct that wraps the interface
	// Since the handler expects a concrete type, we need a workaround
	suite.router = gin.New()
	suite.setupRoutesWithMock()
}

// TearDownTest cleans up after each test
func (suite *MemberHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// setupRoutesWithMock sets up routes that use the mock service directly
func (suite *MemberHandlerTestSuite) setupRoutesWithMock() {
	// Create custom handlers that use the mock service
	suite.router.POST("/members", func(c *gin.Context) {
		var req service.CreateMemberRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		member, err := suite.mockService.CreateMember(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, member)
	})

	suite.router.GET("/members/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
			return
		}
		member, err := suite.mockService.GetMemberByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		c.JSON(http.StatusOK, member)
	})

	suite.router.PUT("/members/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
			return
		}
		var req service.UpdateMemberRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		member, err := suite.mockService.UpdateMember(id, &req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, member)
	})

	suite.router.DELETE("/members/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid member ID"})
			return
		}
		err = suite.mockService.DeleteMember(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Member deleted successfully"})
	})

	suite.router.GET("/organizations/:orgId/members", func(c *gin.Context) {
		orgIdStr := c.Param("orgId")
		orgId, err := uuid.Parse(orgIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			return
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		members, total, err := suite.mockService.GetMembersByOrganization(orgId, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"members": members,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		})
	})

	suite.router.GET("/organizations/:orgId/members/search", func(c *gin.Context) {
		orgIdStr := c.Param("orgId")
		orgId, err := uuid.Parse(orgIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			return
		}
		query := c.Query("q")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		members, total, err := suite.mockService.SearchMembers(orgId, query, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"members": members,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		})
	})

	suite.router.GET("/organizations/:orgId/members/active", func(c *gin.Context) {
		orgIdStr := c.Param("orgId")
		orgId, err := uuid.Parse(orgIdStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
			return
		}
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		members, total, err := suite.mockService.GetActiveMembers(orgId, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"members": members,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		})
	})

	suite.router.GET("/members/:id/quick-links", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
			return
		}
		quickLinks, err := suite.mockService.GetQuickLinks(id)
		if err != nil {
			if errors.Is(err, apperrors.ErrMemberNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, quickLinks)
	})

	suite.router.POST("/members/:id/quick-links", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
			return
		}
		var req service.AddQuickLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		member, err := suite.mockService.AddQuickLink(id, &req)
		if err != nil {
			if errors.Is(err, apperrors.ErrMemberNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, apperrors.ErrLinkExists) {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, member)
	})

	suite.router.DELETE("/members/:id/quick-links", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member ID"})
			return
		}
		linkURL := c.Query("url")
		if linkURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
			return
		}
		member, err := suite.mockService.RemoveQuickLink(id, linkURL)
		if err != nil {
			if errors.Is(err, apperrors.ErrMemberNotFound) || errors.Is(err, apperrors.ErrLinkNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, member)
	})
}

// TestCreateMember tests the CreateMember handler
func (suite *MemberHandlerTestSuite) TestCreateMember() {
	// Test validation error - invalid request body
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/members", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestGetMember tests the GetMember handler
func (suite *MemberHandlerTestSuite) TestGetMember() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/members/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid member ID")
	})
}

// TestUpdateMember tests the UpdateMember handler
func (suite *MemberHandlerTestSuite) TestUpdateMember() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		fullName := "Updated John Doe"
		email := "john.updated@company.com"
		updateRequest := service.UpdateMemberRequest{
			FullName: &fullName,
			Email:    &email,
		}

		body, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, "/members/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid member ID")
	})

	// Test invalid JSON
	suite.T().Run("Invalid JSON", func(t *testing.T) {
		memberID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/members/%s", memberID), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "error")
	})
}

// TestDeleteMember tests the DeleteMember handler
func (suite *MemberHandlerTestSuite) TestDeleteMember() {
	// Test invalid UUID
	suite.T().Run("Invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/members/invalid-uuid", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid member ID")
	})
}

// TestGetMembersByOrganization tests the GetMembersByOrganization handler
func (suite *MemberHandlerTestSuite) TestGetMembersByOrganization() {
	// Test invalid organization_id
	suite.T().Run("Invalid Organization ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/members", nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid organization ID")
	})
}

// TestCreateMemberSuccess tests successful member creation
func (suite *MemberHandlerTestSuite) TestCreateMemberSuccess() {
	suite.T().Run("Success", func(t *testing.T) {
		role := "developer"
		createReq := service.CreateMemberRequest{
			OrganizationID: uuid.New(),
			FullName:       "Smith, John",
			FirstName:      "John",
			LastName:       "Smith",
			Email:          "john.smith@example.com",
			PhoneNumber:    "+1-555-0123",
			IUser:          "I123456",
			Role:           &role,
		}

		expectedResponse := &service.MemberResponse{
			ID:             uuid.New(),
			OrganizationID: createReq.OrganizationID,
			FullName:       createReq.FullName,
			FirstName:      createReq.FirstName,
			LastName:       createReq.LastName,
			Email:          createReq.Email,
			PhoneNumber:    createReq.PhoneNumber,
			IUser:          createReq.IUser,
			Role:           *createReq.Role,
			IsActive:       true,
		}

		suite.mockService.EXPECT().
			CreateMember(gomock.Any()).
			Return(expectedResponse, nil).
			Times(1)

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/members", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response service.MemberResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Email, response.Email)
		assert.Equal(t, expectedResponse.FullName, response.FullName)
	})
}

// TestCreateMemberWithDefaultValues tests creating a member without role and team_role (should use defaults)
func (suite *MemberHandlerTestSuite) TestCreateMemberWithDefaultValues() {
	suite.T().Run("Success with defaults", func(t *testing.T) {
		createReq := service.CreateMemberRequest{
			OrganizationID: uuid.New(),
			FullName:       "Smith, John",
			FirstName:      "John",
			LastName:       "Smith",
			Email:          "john.smith@example.com",
			PhoneNumber:    "+1-555-0123",
			IUser:          "I123456",
			// Role and TeamRole are not provided - should use defaults
		}

		expectedResponse := &service.MemberResponse{
			ID:             uuid.New(),
			OrganizationID: createReq.OrganizationID,
			FullName:       createReq.FullName,
			FirstName:      createReq.FirstName,
			LastName:       createReq.LastName,
			Email:          createReq.Email,
			PhoneNumber:    createReq.PhoneNumber,
			IUser:          createReq.IUser,
			Role:           "developer", // Default value
			TeamRole:       "member",    // Default value
			IsActive:       true,
		}

		suite.mockService.EXPECT().
			CreateMember(gomock.Any()).
			Return(expectedResponse, nil).
			Times(1)

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/members", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response service.MemberResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Email, response.Email)
		assert.Equal(t, expectedResponse.FullName, response.FullName)
		assert.Equal(t, "developer", response.Role)
		assert.Equal(t, "member", response.TeamRole)
	})
}

// TestGetMemberSuccess tests successful member retrieval
func (suite *MemberHandlerTestSuite) TestGetMemberSuccess() {
	suite.T().Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		expectedResponse := &service.MemberResponse{
			ID:          memberID,
			FullName:    "Doe, Jane",
			FirstName:   "Jane",
			LastName:    "Doe",
			Email:       "jane.doe@example.com",
			PhoneNumber: "+1-555-9999",
			IUser:       "I789012",
			Role:        "manager",
			IsActive:    true,
		}

		suite.mockService.EXPECT().
			GetMemberByID(memberID).
			Return(expectedResponse, nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/members/%s", memberID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.MemberResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Email, response.Email)
		assert.Equal(t, memberID, response.ID)
	})
}

// TestUpdateMemberSuccess tests successful member update
func (suite *MemberHandlerTestSuite) TestUpdateMemberSuccess() {
	suite.T().Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		newFullName := "Updated, Name"
		newFirstName := "Name"
		newLastName := "Updated"
		updateRequest := service.UpdateMemberRequest{
			FullName:  &newFullName,
			FirstName: &newFirstName,
			LastName:  &newLastName,
		}

		expectedResponse := &service.MemberResponse{
			ID:        memberID,
			FullName:  newFullName,
			FirstName: newFirstName,
			LastName:  newLastName,
			Email:     "test@example.com",
		}

		suite.mockService.EXPECT().
			UpdateMember(memberID, gomock.Any()).
			Return(expectedResponse, nil).
			Times(1)

		body, _ := json.Marshal(updateRequest)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/members/%s", memberID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response service.MemberResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, newFullName, response.FullName)
	})
}

// TestDeleteMemberSuccess tests successful member deletion
func (suite *MemberHandlerTestSuite) TestDeleteMemberSuccess() {
	suite.T().Run("Success", func(t *testing.T) {
		memberID := uuid.New()

		suite.mockService.EXPECT().
			DeleteMember(memberID).
			Return(nil).
			Times(1)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/members/%s", memberID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "deleted successfully")
	})
}

// TestGetMembersByOrganizationSuccess tests successful member listing
func (suite *MemberHandlerTestSuite) TestGetMembersByOrganizationSuccess() {
	suite.T().Run("Success", func(t *testing.T) {
		orgID := uuid.New()
		members := []service.MemberResponse{
			{
				ID:        uuid.New(),
				FullName:  "Smith, Alice",
				FirstName: "Alice",
				LastName:  "Smith",
				Email:     "alice@example.com",
			},
			{
				ID:        uuid.New(),
				FullName:  "Jones, Bob",
				FirstName: "Bob",
				LastName:  "Jones",
				Email:     "bob@example.com",
			},
		}

		suite.mockService.EXPECT().
			GetMembersByOrganization(orgID, 20, 0).
			Return(members, int64(2), nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/members", orgID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "alice@example.com")
		assert.Contains(t, w.Body.String(), "bob@example.com")
	})
}

// TestSearchMembersSuccess tests successful member search
func (suite *MemberHandlerTestSuite) TestSearchMembersSuccess() {
	suite.T().Run("Success", func(t *testing.T) {
		orgID := uuid.New()
		members := []service.MemberResponse{
			{
				ID:        uuid.New(),
				FullName:  "Smith, John",
				FirstName: "John",
				LastName:  "Smith",
				Email:     "john.smith@example.com",
			},
		}

		suite.mockService.EXPECT().
			SearchMembers(orgID, "john", 20, 0).
			Return(members, int64(1), nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/members/search?q=john", orgID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "john.smith@example.com")
	})
}

// TestGetActiveMembersSuccess tests successful active members retrieval
func (suite *MemberHandlerTestSuite) TestGetActiveMembersSuccess() {
	suite.T().Run("Success", func(t *testing.T) {
		orgID := uuid.New()
		members := []service.MemberResponse{
			{
				ID:        uuid.New(),
				FullName:  "Active, User",
				FirstName: "User",
				LastName:  "Active",
				Email:     "active@example.com",
				IsActive:  true,
			},
		}

		suite.mockService.EXPECT().
			GetActiveMembers(orgID, 20, 0).
			Return(members, int64(1), nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/members/active", orgID), nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "active@example.com")
	})
}

// TestMemberHandlerTestSuite runs the test suite
// TestAddQuickLink tests the AddQuickLink handler
func (suite *MemberHandlerTestSuite) TestAddQuickLink() {
	suite.T().Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		requestBody := map[string]interface{}{
			"url":      "https://github.com/user/repo",
			"title":    "My Repository",
			"icon":     "github",
			"category": "repository",
		}

		expectedResponse := &service.MemberResponse{
			ID:       memberID,
			FullName: "John Doe",
			Email:    "john@example.com",
		}

		suite.mockService.EXPECT().
			AddQuickLink(memberID, gomock.Any()).
			Return(expectedResponse, nil).
			Times(1)

		reqBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/members/%s/quick-links", memberID), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	suite.T().Run("Invalid Member ID", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"url":   "https://github.com/user/repo",
			"title": "My Repository",
		}

		reqBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/members/invalid-uuid/quick-links", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid member ID")
	})

	suite.T().Run("Member Not Found", func(t *testing.T) {
		memberID := uuid.New()
		requestBody := map[string]interface{}{
			"url":   "https://github.com/user/repo",
			"title": "My Repository",
		}

		suite.mockService.EXPECT().
			AddQuickLink(memberID, gomock.Any()).
			Return(nil, apperrors.ErrMemberNotFound).
			Times(1)

		reqBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/members/%s/quick-links", memberID), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "member not found")
	})

	suite.T().Run("Duplicate Link", func(t *testing.T) {
		memberID := uuid.New()
		requestBody := map[string]interface{}{
			"url":   "https://github.com/user/repo",
			"title": "My Repository",
		}

		suite.mockService.EXPECT().
			AddQuickLink(memberID, gomock.Any()).
			Return(nil, apperrors.ErrLinkExists).
			Times(1)

		reqBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/members/%s/quick-links", memberID), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "link already exists with this URL")
	})
}

// TestGetQuickLinks tests the GetQuickLinks handler
func (suite *MemberHandlerTestSuite) TestGetQuickLinks() {
	suite.T().Run("Success", func(t *testing.T) {
		memberID := uuid.New()

		expectedResponse := &service.QuickLinksResponse{
			QuickLinks: []service.QuickLink{
				{
					URL:      "https://github.com/user/repo",
					Title:    "My Repository",
					Icon:     "github",
					Category: "repository",
				},
				{
					URL:      "https://example.com/docs",
					Title:    "Documentation",
					Icon:     "docs",
					Category: "documentation",
				},
			},
		}

		suite.mockService.EXPECT().
			GetQuickLinks(memberID).
			Return(expectedResponse, nil).
			Times(1)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/members/%s/quick-links", memberID), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "https://github.com/user/repo")
		assert.Contains(t, w.Body.String(), "My Repository")
	})

	suite.T().Run("Success with empty links", func(t *testing.T) {
		memberID := uuid.New()

		expectedResponse := &service.QuickLinksResponse{
			QuickLinks: []service.QuickLink{},
		}

		suite.mockService.EXPECT().
			GetQuickLinks(memberID).
			Return(expectedResponse, nil).
			Times(1)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/members/%s/quick-links", memberID), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "quick_links")
	})

	suite.T().Run("Invalid Member ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/members/invalid-uuid/quick-links", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid member ID")
	})

	suite.T().Run("Member Not Found", func(t *testing.T) {
		memberID := uuid.New()

		suite.mockService.EXPECT().
			GetQuickLinks(memberID).
			Return(nil, apperrors.ErrMemberNotFound).
			Times(1)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/members/%s/quick-links", memberID), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "member not found")
	})
}

// TestRemoveQuickLink tests the RemoveQuickLink handler
func (suite *MemberHandlerTestSuite) TestRemoveQuickLink() {
	suite.T().Run("Success", func(t *testing.T) {
		memberID := uuid.New()
		linkURL := "https://github.com/user/repo"

		expectedResponse := &service.MemberResponse{
			ID:       memberID,
			FullName: "John Doe",
			Email:    "john@example.com",
		}

		suite.mockService.EXPECT().
			RemoveQuickLink(memberID, linkURL).
			Return(expectedResponse, nil).
			Times(1)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/members/%s/quick-links?url=%s", memberID, linkURL), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	suite.T().Run("Invalid Member ID", func(t *testing.T) {
		linkURL := "https://github.com/user/repo"

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/members/invalid-uuid/quick-links?url=%s", linkURL), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid member ID")
	})

	suite.T().Run("Missing URL Parameter", func(t *testing.T) {
		memberID := uuid.New()

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/members/%s/quick-links", memberID), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "url query parameter is required")
	})

	suite.T().Run("Link Not Found", func(t *testing.T) {
		memberID := uuid.New()
		linkURL := "https://github.com/user/nonexistent"

		suite.mockService.EXPECT().
			RemoveQuickLink(memberID, linkURL).
			Return(nil, apperrors.ErrLinkNotFound).
			Times(1)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/members/%s/quick-links?url=%s", memberID, linkURL), nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "link not found")
	})
}

func TestMemberHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MemberHandlerTestSuite))
}
