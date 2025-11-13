package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

// LinkRepositoryTestSuite tests the LinkRepository
type LinkRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *LinkRepository
}

// SetupSuite runs before all tests in the suite
func (suite *LinkRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	suite.repo = NewLinkRepository(suite.baseTestSuite.DB)
}

// TearDownSuite runs after all tests in the suite
func (suite *LinkRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *LinkRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *LinkRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// helper to insert a category (FK) directly via gorm
func (suite *LinkRepositoryTestSuite) createCategory(name, title, icon, color string) *models.Category {
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

// helper to insert a link directly via gorm
func (suite *LinkRepositoryTestSuite) createLink(owner uuid.UUID, title, url string, categoryID uuid.UUID, tags string) *models.Link {
	l := &models.Link{
		BaseModel: models.BaseModel{
			ID:    uuid.New(),
			Name:  title, // keep name aligned with title
			Title: title,
		},
		Owner:      owner,
		URL:        url,
		CategoryID: categoryID,
		Tags:       tags,
	}
	err := suite.baseTestSuite.DB.Create(l).Error
	suite.NoError(err)
	return l
}

// TestCreate tests creating a new link
func (suite *LinkRepositoryTestSuite) TestCreate() {
	cat := suite.createCategory("cat-1", "Category 1", "icon-1", "red")
	owner := uuid.New()

	link := &models.Link{
		BaseModel: models.BaseModel{
			ID:    uuid.New(),
			Name:  "alpha",
			Title: "Alpha",
		},
		Owner:      owner,
		URL:        "https://example.com/alpha",
		CategoryID: cat.ID,
		Tags:       "tag1,tag2",
	}

	err := suite.repo.Create(link)

	suite.NoError(err)
	suite.NotEqual(uuid.Nil, link.ID)
	suite.NotZero(link.CreatedAt)
	suite.NotZero(link.UpdatedAt)
}

// TestGetByOwner tests retrieving links by owner ordered by title ASC
func (suite *LinkRepositoryTestSuite) TestGetByOwner() {
	cat := suite.createCategory("cat-2", "Category 2", "icon-2", "blue")
	owner := uuid.New()

	// Create multiple links with out-of-order titles
	_ = suite.createLink(owner, "Charlie", "https://example.com/charlie", cat.ID, "")
	_ = suite.createLink(owner, "Alpha", "https://example.com/alpha", cat.ID, "")
	_ = suite.createLink(owner, "Bravo", "https://example.com/bravo", cat.ID, "")

	links, err := suite.repo.GetByOwner(owner)

	suite.NoError(err)
	suite.Len(links, 3)
	// Verify ordering by title ASC: Alpha, Bravo, Charlie
	suite.Equal("Alpha", links[0].Title)
	suite.Equal("Bravo", links[1].Title)
	suite.Equal("Charlie", links[2].Title)
}

// TestGetByIDs tests retrieving links by IDs, ordered by title ASC
func (suite *LinkRepositoryTestSuite) TestGetByIDs() {
	cat := suite.createCategory("cat-3", "Category 3", "icon-3", "green")
	owner := uuid.New()

	l1 := suite.createLink(owner, "Zeta", "https://example.com/zeta", cat.ID, "")
	l2 := suite.createLink(owner, "Eta", "https://example.com/eta", cat.ID, "")
	l3 := suite.createLink(owner, "Theta", "https://example.com/theta", cat.ID, "")

	ids := []uuid.UUID{l1.ID, l2.ID, l3.ID}
	links, err := suite.repo.GetByIDs(ids)

	suite.NoError(err)
	suite.Len(links, 3)
	// Ordered by title ASC: Eta, Theta, Zeta
	suite.Equal("Eta", links[0].Title)
	suite.Equal("Theta", links[1].Title)
	suite.Equal("Zeta", links[2].Title)
}

// TestGetByIDs_Empty tests retrieving with empty ID slice returns empty result
func (suite *LinkRepositoryTestSuite) TestGetByIDs_Empty() {
	links, err := suite.repo.GetByIDs([]uuid.UUID{})
	suite.NoError(err)
	suite.Len(links, 0)
}

// TestDelete tests deleting a link
func (suite *LinkRepositoryTestSuite) TestDelete() {
	cat := suite.createCategory("cat-4", "Category 4", "icon-4", "yellow")
	owner := uuid.New()

	l := suite.createLink(owner, "DeleteMe", "https://example.com/del", cat.ID, "")

	err := suite.repo.Delete(l.ID)
	suite.NoError(err)

	// Verify link is deleted by fetching by IDs
	found, err := suite.repo.GetByIDs([]uuid.UUID{l.ID})
	suite.NoError(err)
	suite.Len(found, 0)
}

// TestDeleteNotFound tests deleting a non-existent link (should not)
func (suite *LinkRepositoryTestSuite) TestDeleteNotFound() {
	err := suite.repo.Delete(uuid.New())
	suite.NoError(err)
}

// Run the test suite
func TestLinkRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(LinkRepositoryTestSuite))
}
