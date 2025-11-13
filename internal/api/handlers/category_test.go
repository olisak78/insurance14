package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"developer-portal-backend/internal/api/handlers"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// CategoryHandlerTestSuite defines the test suite for CategoryHandler
type CategoryHandlerTestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mockCategorySv *mocks.MockCategoryServiceInterface
	handler        *handlers.CategoryHandler
	router         *gin.Engine
}

func (suite *CategoryHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockCategorySv = mocks.NewMockCategoryServiceInterface(suite.ctrl)
	suite.handler = handlers.NewCategoryHandler(suite.mockCategorySv)

	suite.router = gin.New()
	suite.router.GET("/categories", suite.handler.ListCategories)
}

func (suite *CategoryHandlerTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *CategoryHandlerTestSuite) TestListCategories_DefaultPagination_Success() {
	// Expect defaults page=1, page_size=1000
	resp := &service.CategoryListResponse{
		Categories: []service.CategoryResponse{
			{
				ID:          uuid.New(),
				Name:        "platform",
				Title:       "Platform",
				Description: "Platform capabilities",
				Icon:        "cube",
				Color:       "#3366ff",
			},
		},
		Total:    1,
		Page:     1,
		PageSize: 1000,
	}
	suite.mockCategorySv.EXPECT().GetAll(1, 1000).Return(resp, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got service.CategoryListResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), got.Total)
	assert.Equal(suite.T(), 1, got.Page)
	assert.Equal(suite.T(), 1000, got.PageSize)
	assert.Len(suite.T(), got.Categories, 1)
	assert.Equal(suite.T(), "platform", got.Categories[0].Name)
	assert.Equal(suite.T(), "Platform", got.Categories[0].Title)
}

func (suite *CategoryHandlerTestSuite) TestListCategories_CustomPagination_Success() {
	resp := &service.CategoryListResponse{
		Categories: []service.CategoryResponse{},
		Total:      0,
		Page:       2,
		PageSize:   10,
	}
	suite.mockCategorySv.EXPECT().GetAll(2, 10).Return(resp, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories?page=2&page_size=10", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got service.CategoryListResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, got.Page)
	assert.Equal(suite.T(), 10, got.PageSize)
	assert.Equal(suite.T(), int64(0), got.Total)
	assert.Len(suite.T(), got.Categories, 0)
}

func (suite *CategoryHandlerTestSuite) TestListCategories_BoundsNormalization_Success() {
	// page=0 should normalize to 1; page_size=5001 should normalize to 1000
	resp := &service.CategoryListResponse{
		Categories: []service.CategoryResponse{},
		Total:      0,
		Page:       1,
		PageSize:   1000,
	}
	suite.mockCategorySv.EXPECT().GetAll(1, 1000).Return(resp, nil)

	req := httptest.NewRequest(http.MethodGet, "/categories?page=0&page_size=5001", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var got service.CategoryListResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, got.Page)
	assert.Equal(suite.T(), 1000, got.PageSize)
}

func (suite *CategoryHandlerTestSuite) TestListCategories_ServiceError() {
	suite.mockCategorySv.EXPECT().GetAll(1, 1000).Return(nil, errors.New("db failure"))

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	body := w.Body.String()
	assert.Contains(suite.T(), body, "Failed to get categories")
	assert.Contains(suite.T(), body, "db failure")
}

func TestCategoryHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryHandlerTestSuite))
}
