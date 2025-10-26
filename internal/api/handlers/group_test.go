package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// GroupHandlerTestSuite tests the GroupHandler
type GroupHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	ctrl        *gomock.Controller
	mockService *mocks.MockGroupServiceInterface
	handler     *GroupHandler
}

// SetupSuite sets up the test suite
func (suite *GroupHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

// SetupTest sets up each individual test
func (suite *GroupHandlerTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockService = mocks.NewMockGroupServiceInterface(suite.ctrl)
	suite.handler = NewGroupHandler(suite.mockService)

	suite.router = gin.New()

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	{
		groups := v1.Group("/groups")
		{
			groups.POST("", suite.handler.CreateGroup)
			groups.GET("/:id", suite.handler.GetGroup)
			groups.PUT("/:id", suite.handler.UpdateGroup)
			groups.DELETE("/:id", suite.handler.DeleteGroup)
		}

		organizations := v1.Group("/organizations")
		{
			organizations.GET("/:id/groups", suite.handler.GetGroupsByOrganization)
		}
	}
}

// TearDownTest cleans up after each test
func (suite *GroupHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateGroup tests creating a new group
func (suite *GroupHandlerTestSuite) TestCreateGroup() {
	orgID := uuid.New()
	groupID := uuid.New()

	request := service.CreateGroupRequest{
		OrganizationID: orgID,
		Name:           "test-group",
		DisplayName:    "Test Group",
		Description:    "Test group description",
	}

	expectedResponse := &service.GroupResponse{
		ID:             groupID,
		OrganizationID: orgID,
		Name:           "test-group",
		DisplayName:    "Test Group",
		Description:    "Test group description",
	}

	suite.mockService.EXPECT().
		Create(gomock.Any()).
		Return(expectedResponse, nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response service.GroupResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), groupID, response.ID)
	assert.Equal(suite.T(), "test-group", response.Name)
}

// TestCreateGroupInvalidInput tests creating a group with invalid input
func (suite *GroupHandlerTestSuite) TestCreateGroupInvalidInput() {
	request := service.CreateGroupRequest{
		// Missing required fields
		Name: "",
	}

	// The handler will still call the service, so we need to expect it and return an error
	suite.mockService.EXPECT().
		Create(gomock.Any()).
		Return(nil, assert.AnError)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestGetGroup tests retrieving a group by ID
func (suite *GroupHandlerTestSuite) TestGetGroup() {
	groupID := uuid.New()
	orgID := uuid.New()

	expectedResponse := &service.GroupResponse{
		ID:             groupID,
		OrganizationID: orgID,
		Name:           "test-group",
		DisplayName:    "Test Group",
	}

	suite.mockService.EXPECT().
		GetByID(groupID).
		Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/"+groupID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.GroupResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), groupID, response.ID)
	assert.Equal(suite.T(), "test-group", response.Name)
}

// TestUpdateGroup tests updating a group
func (suite *GroupHandlerTestSuite) TestUpdateGroup() {
	groupID := uuid.New()
	orgID := uuid.New()

	request := service.UpdateGroupRequest{
		DisplayName: "Updated Group",
		Description: "Updated description",
	}

	expectedResponse := &service.GroupResponse{
		ID:             groupID,
		OrganizationID: orgID,
		Name:           "test-group",
		DisplayName:    "Updated Group",
		Description:    "Updated description",
	}

	suite.mockService.EXPECT().
		Update(groupID, gomock.Any()).
		Return(expectedResponse, nil)

	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/groups/"+groupID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.GroupResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Group", response.DisplayName)
}

// TestDeleteGroup tests deleting a group
func (suite *GroupHandlerTestSuite) TestDeleteGroup() {
	groupID := uuid.New()

	suite.mockService.EXPECT().
		Delete(groupID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/groups/"+groupID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

// TestListGroupsByOrganization tests listing groups by organization
func (suite *GroupHandlerTestSuite) TestListGroupsByOrganization() {
	orgID := uuid.New()
	group1ID := uuid.New()
	group2ID := uuid.New()

	expectedResponse := &service.GroupListResponse{
		Groups: []service.GroupResponse{
			{
				ID:             group1ID,
				OrganizationID: orgID,
				Name:           "group-1",
				DisplayName:    "Group 1",
			},
			{
				ID:             group2ID,
				OrganizationID: orgID,
				Name:           "group-2",
				DisplayName:    "Group 2",
			},
		},
		Total: 2,
		Page:  1,
	}

	suite.mockService.EXPECT().
		GetByOrganization(orgID, 1, 20).
		Return(expectedResponse, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+orgID.String()+"/groups?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response service.GroupListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(response.Groups))
	assert.Equal(suite.T(), int64(2), response.Total)
}

// TestInvalidUUID tests endpoints with invalid UUID parameters
func (suite *GroupHandlerTestSuite) TestInvalidUUID() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups/invalid-uuid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// Run the test suite
func TestGroupHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(GroupHandlerTestSuite))
}
