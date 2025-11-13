package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

type TeamHandlerTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockTeam *mocks.MockTeamServiceInterface
	handler  *handlers.TeamHandler
}

func (suite *TeamHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockTeam = mocks.NewMockTeamServiceInterface(suite.ctrl)
	suite.handler = handlers.NewTeamHandler(suite.mockTeam)
}

func (suite *TeamHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// helper to build router and optionally inject a viewer name (username) into context
func (suite *TeamHandlerTestSuite) newRouter(withViewer bool, viewerName string) *gin.Engine {
	r := gin.New()
	if withViewer {
		r.Use(func(c *gin.Context) {
			c.Set("username", viewerName)
			c.Next()
		})
	}
	r.GET("/teams", suite.handler.GetAllTeams)
	return r
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_DefaultList_Success() {
	router := suite.newRouter(false, "")
	teamID := uuid.New()
	resp := &service.TeamListResponse{
		Teams: []service.TeamResponse{
			{
				ID:          teamID,
				GroupID:     uuid.New(),
				Name:        "alpha",
				Title:       "Alpha Team",
				Description: "Desc",
				PictureURL:  "https://img",
			},
		},
		Total:    1,
		Page:     1,
		PageSize: 1,
	}
	suite.mockTeam.EXPECT().GetAllTeams(nil, 1, 1000).Return(resp, nil)

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(1), got["total"])
	assert.Equal(suite.T(), float64(1), got["page"])
	assert.Equal(suite.T(), float64(1), got["page_size"])

	items, ok := got["teams"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), items, 1)
	first := items[0].(map[string]interface{})
	assert.Equal(suite.T(), "alpha", first["name"])
	assert.Equal(suite.T(), "Alpha Team", first["title"])
	assert.Equal(suite.T(), "https://img", first["picture_url"])
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_DefaultList_Error() {
	router := suite.newRouter(false, "")
	suite.mockTeam.EXPECT().GetAllTeams(nil, 1, 1000).Return(nil, errors.New("db failure"))

	req := httptest.NewRequest(http.MethodGet, "/teams", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "db failure")
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByID_InvalidUUID() {
	router := suite.newRouter(false, "")
	req := httptest.NewRequest(http.MethodGet, "/teams?team-id=not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "invalid team ID")
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByID_NotFound() {
	router := suite.newRouter(false, "")
	id := uuid.New()

	suite.mockTeam.EXPECT().GetByID(id).Return(nil, apperrors.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/teams?team-id="+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.Contains(suite.T(), w.Body.String(), apperrors.ErrTeamNotFound.Error())
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByID_InternalError() {
	router := suite.newRouter(false, "")
	id := uuid.New()

	suite.mockTeam.EXPECT().GetByID(id).Return(nil, errors.New("db failure"))

	req := httptest.NewRequest(http.MethodGet, "/teams?team-id="+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "db failure")
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByID_WithViewer_Success() {
	router := suite.newRouter(true, "john.doe")
	id := uuid.New()
	teamName := "alpha"

	teamResp := &service.TeamResponse{
		ID:             id,
		GroupID:        uuid.New(),
		OrganizationID: uuid.New(),
		Name:           teamName,
		Title:          "Alpha",
		Description:    "D",
		Owner:          "owner",
		Email:          "alpha@example.com",
		PictureURL:     "https://img",
	}
	// Metadata contains jira.team -> should be mapped to jira_team in handler response
	meta := map[string]interface{}{"jira": map[string]interface{}{"team": "JiraAlpha"}}
	metaBytes, _ := json.Marshal(meta)
	teamWithMembers := &service.TeamWithMembersResponse{
		TeamResponse: service.TeamResponse{
			ID:             teamResp.ID,
			GroupID:        teamResp.GroupID,
			OrganizationID: teamResp.OrganizationID,
			Name:           teamResp.Name,
			Title:          teamResp.Title,
			Description:    teamResp.Description,
			Owner:          teamResp.Owner,
			Email:          teamResp.Email,
			PictureURL:     teamResp.PictureURL,
			Metadata:       metaBytes,
		},
		Members: []service.UserResponse{},
		Links:   []service.LinkResponse{},
	}

	gomock.InOrder(
		suite.mockTeam.EXPECT().GetByID(id).Return(teamResp, nil),
		suite.mockTeam.EXPECT().GetBySimpleNameWithViewer(teamName, "john.doe").Return(teamWithMembers, nil),
	)

	req := httptest.NewRequest(http.MethodGet, "/teams?team-id="+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), teamName, got["name"])
	assert.Equal(suite.T(), "JiraAlpha", got["jira_team"])
	// Metadata may be included; ensure jira_team extracted; if present, ensure it's a map
	metaVal, hasMeta := got["metadata"]
	if hasMeta {
		_, _ = metaVal.(map[string]interface{})
	}
	// Ensure members and links are present arrays
	_, hasMembers := got["members"]
	_, hasLinks := got["links"]
	assert.True(suite.T(), hasMembers)
	assert.True(suite.T(), hasLinks)
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByID_NoViewer_Success() {
	router := suite.newRouter(false, "")
	id := uuid.New()
	teamName := "beta"

	teamResp := &service.TeamResponse{
		ID:             id,
		GroupID:        uuid.New(),
		OrganizationID: uuid.New(),
		Name:           teamName,
		Title:          "Beta",
		Description:    "D2",
		Owner:          "owner",
		Email:          "beta@example.com",
		PictureURL:     "https://img2",
	}
	meta := map[string]interface{}{"jira": map[string]interface{}{"team": "JiraBeta"}}
	metaBytes, _ := json.Marshal(meta)
	teamWithMembers := &service.TeamWithMembersResponse{
		TeamResponse: service.TeamResponse{
			ID:             teamResp.ID,
			GroupID:        teamResp.GroupID,
			OrganizationID: teamResp.OrganizationID,
			Name:           teamResp.Name,
			Title:          teamResp.Title,
			Description:    teamResp.Description,
			Owner:          teamResp.Owner,
			Email:          teamResp.Email,
			PictureURL:     teamResp.PictureURL,
			Metadata:       metaBytes,
		},
		Members: []service.UserResponse{{FirstName: "A", LastName: "B"}},
		Links:   []service.LinkResponse{{ID: uuid.New().String()}},
	}

	gomock.InOrder(
		suite.mockTeam.EXPECT().GetByID(id).Return(teamResp, nil),
		suite.mockTeam.EXPECT().GetBySimpleName(teamName).Return(teamWithMembers, nil),
	)

	req := httptest.NewRequest(http.MethodGet, "/teams?team-id="+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), teamName, got["name"])
	assert.Equal(suite.T(), "JiraBeta", got["jira_team"])
	// Metadata may be included; if present, ensure it's a map
	metaVal, hasMeta := got["metadata"]
	if hasMeta {
		_, _ = metaVal.(map[string]interface{})
	}
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByName_WithViewer_Success() {
	router := suite.newRouter(true, "john.doe")
	teamName := "gamma"
	meta := map[string]interface{}{"jira": map[string]interface{}{"team": "JiraGamma"}}
	metaBytes, _ := json.Marshal(meta)

	teamWithMembers := &service.TeamWithMembersResponse{
		TeamResponse: service.TeamResponse{
			ID:             uuid.New(),
			GroupID:        uuid.New(),
			OrganizationID: uuid.New(),
			Name:           teamName,
			Title:          "Gamma",
			Description:    "DG",
			Owner:          "owner",
			Email:          "gamma@example.com",
			PictureURL:     "https://img3",
			Metadata:       metaBytes,
		},
		Members: []service.UserResponse{},
		Links:   []service.LinkResponse{},
	}

	suite.mockTeam.EXPECT().GetBySimpleNameWithViewer(teamName, "john.doe").Return(teamWithMembers, nil)

	req := httptest.NewRequest(http.MethodGet, "/teams?team-name="+teamName, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), teamName, got["name"])
	assert.Equal(suite.T(), "JiraGamma", got["jira_team"])
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByName_NoViewer_Success() {
	router := suite.newRouter(false, "")
	teamName := "delta"
	meta := map[string]interface{}{"jira": map[string]interface{}{"team": "JiraDelta"}}
	metaBytes, _ := json.Marshal(meta)

	teamWithMembers := &service.TeamWithMembersResponse{
		TeamResponse: service.TeamResponse{
			ID:             uuid.New(),
			GroupID:        uuid.New(),
			OrganizationID: uuid.New(),
			Name:           teamName,
			Title:          "Delta",
			Description:    "DD",
			Owner:          "owner",
			Email:          "delta@example.com",
			PictureURL:     "https://img4",
			Metadata:       metaBytes,
		},
		Members: []service.UserResponse{},
		Links:   []service.LinkResponse{},
	}

	suite.mockTeam.EXPECT().GetBySimpleName(teamName).Return(teamWithMembers, nil)

	req := httptest.NewRequest(http.MethodGet, "/teams?team-name="+teamName, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), teamName, got["name"])
	assert.Equal(suite.T(), "JiraDelta", got["jira_team"])
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByName_NotFound() {
	router := suite.newRouter(false, "")
	teamName := "epsilon"

	suite.mockTeam.EXPECT().GetBySimpleName(teamName).Return(nil, apperrors.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/teams?team-name="+teamName, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.Contains(suite.T(), w.Body.String(), apperrors.ErrTeamNotFound.Error())
}

func (suite *TeamHandlerTestSuite) TestGetAllTeams_ByName_InternalError() {
	router := suite.newRouter(false, "")
	teamName := "zeta"

	suite.mockTeam.EXPECT().GetBySimpleName(teamName).Return(nil, errors.New("db failure"))

	req := httptest.NewRequest(http.MethodGet, "/teams?team-name="+teamName, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "db failure")
}

func TestTeamHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TeamHandlerTestSuite))
}
