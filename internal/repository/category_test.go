package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// CategoryRepositoryTestSuite tests the CategoryRepository
type CategoryRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *CategoryRepository
}

// SetupSuite runs before all tests in the suite
func (suite *CategoryRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())
	suite.repo = NewCategoryRepository(suite.baseTestSuite.DB)
}

// TearDownSuite runs after all tests in the suite
func (suite *CategoryRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *CategoryRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *CategoryRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// helper to insert a category directly via gorm
func (suite *CategoryRepositoryTestSuite) createCategory(name, title, icon, color string) *models.Category {
	c := &models.Category{
		BaseModel: models.BaseModel{
			ID:    uuid.New(),
			Name:  name,
			Title: title,
		},
		Icon:  icon,
		Color: color,
	}
	err := suite.baseTestSuite.DB.Create(c).Error
	suite.NoError(err)
	return c
}

// TestGetByID tests retrieving a category by ID
func (suite *CategoryRepositoryTestSuite) TestGetByID() {
	// Create test category
	category := suite.createCategory("cat-1", "Category 1", "icon-1", "red")

	// Retrieve the category
	retrieved, err := suite.repo.GetByID(category.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrieved)
	suite.Equal(category.ID, retrieved.ID)
	suite.Equal("cat-1", retrieved.Name)
	suite.Equal("Category 1", retrieved.Title)
	suite.Equal("icon-1", retrieved.Icon)
	suite.Equal("red", retrieved.Color)
}

// TestGetByIDNotFound tests retrieving a non-existent category
func (suite *CategoryRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	cat, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(cat)
}

// TestGetAll tests listing categories and ordering by title ascending
func (suite *CategoryRepositoryTestSuite) TestGetAll() {
	// Create multiple categories with different titles
	suite.createCategory("cat-a", "Alpha", "icon-a", "blue")
	suite.createCategory("cat-c", "Charlie", "icon-c", "green")
	suite.createCategory("cat-b", "Bravo", "icon-b", "red")

	// List all categories
	items, total, err := suite.repo.GetAll(10, 0)

	// Assertions
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(items, 3)

	// Verify ordering by title ASC: Alpha, Bravo, Charlie
	suite.Equal("Alpha", items[0].Title)
	suite.Equal("Bravo", items[1].Title)
	suite.Equal("Charlie", items[2].Title)
}

// TestGetAllWithPagination tests listing categories with pagination
func (suite *CategoryRepositoryTestSuite) TestGetAllWithPagination() {
	// Create multiple categories
	for i := 0; i < 5; i++ {
		name := "cat-" + uuid.New().String()[:6]
		title := "Title " + uuid.New().String()[:6]
		suite.createCategory(name, title, "icon-"+uuid.New().String()[:4], "color-"+uuid.New().String()[:4])
	}

	// First page
	items, total, err := suite.repo.GetAll(2, 0)
	suite.NoError(err)
	suite.Len(items, 2)
	suite.GreaterOrEqual(total, int64(5))

	// Second page
	items, total, err = suite.repo.GetAll(2, 2)
	suite.NoError(err)
	suite.Len(items, 2)
	suite.GreaterOrEqual(total, int64(5))

	// Third page
	items, total, err = suite.repo.GetAll(2, 4)
	suite.NoError(err)
	suite.True(len(items) >= 1)
	suite.GreaterOrEqual(total, int64(5))
}

// Run the test suite
func TestCategoryRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryRepositoryTestSuite))
}
