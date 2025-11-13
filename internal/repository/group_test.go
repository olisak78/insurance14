package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// GroupRepositoryTestSuite tests the GroupRepository
type GroupRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *GroupRepository
	factories     *testutils.FactorySet
}

// SetupSuite runs before all tests in the suite
func (suite *GroupRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	suite.repo = NewGroupRepository(suite.baseTestSuite.DB)
	suite.factories = testutils.NewFactorySet()
}

// TearDownSuite runs after all tests in the suite
func (suite *GroupRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *GroupRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *GroupRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// helper to create and persist an organization
func (suite *GroupRepositoryTestSuite) createOrganization() *models.Organization {
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)
	return org
}

// TestCreate tests creating a new group
func (suite *GroupRepositoryTestSuite) TestCreate() {
	// Create organization
	org := suite.createOrganization()

	// Create test group
	group := suite.factories.Group.Create()
	group.OrgID = org.ID

	// Create the group
	err := suite.repo.Create(group)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, group.ID)
	suite.NotZero(group.CreatedAt)
	suite.NotZero(group.UpdatedAt)
}

// TestGetByID tests retrieving a group by ID
func (suite *GroupRepositoryTestSuite) TestGetByID() {
	// Create organization
	org := suite.createOrganization()

	// Create test group
	group := suite.factories.Group.WithOrganization(org.ID)
	err := suite.repo.Create(group)
	suite.NoError(err)

	// Retrieve the group
	retrievedGroup, err := suite.repo.GetByID(group.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedGroup)
	suite.Equal(group.ID, retrievedGroup.ID)
	suite.Equal(group.Name, retrievedGroup.Name)
	suite.Equal(group.Title, retrievedGroup.Title)
	suite.Equal(group.OrgID, retrievedGroup.OrgID)
	suite.Equal(group.Owner, retrievedGroup.Owner)
}

// TestGetByIDNotFound tests retrieving a non-existent group
func (suite *GroupRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	group, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(group)
}

// TestGetByName tests retrieving a group by name within organization
func (suite *GroupRepositoryTestSuite) TestGetByName() {
	// Create organization
	org := suite.createOrganization()

	// Create test group
	group := suite.factories.Group.WithName("unique-group-name")
	group.OrgID = org.ID
	err := suite.repo.Create(group)
	suite.NoError(err)

	// Retrieve the group by name
	retrievedGroup, err := suite.repo.GetByName(org.ID, "unique-group-name")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedGroup)
	suite.Equal(group.ID, retrievedGroup.ID)
	suite.Equal("unique-group-name", retrievedGroup.Name)
}

// TestGetByNameNotFound tests retrieving a non-existent group by name
func (suite *GroupRepositoryTestSuite) TestGetByNameNotFound() {
	// Create organization
	org := suite.createOrganization()

	group, err := suite.repo.GetByName(org.ID, "nonexistent-group")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(group)
}

// TestGetByOrganizationID tests listing groups by organization with pagination
func (suite *GroupRepositoryTestSuite) TestGetByOrganizationID() {
	// Create organization
	org := suite.createOrganization()

	// Create multiple test groups under org
	group1 := suite.factories.Group.WithName("group-1")
	group1.OrgID = org.ID
	err := suite.repo.Create(group1)
	suite.NoError(err)

	group2 := suite.factories.Group.WithName("group-2")
	group2.OrgID = org.ID
	err = suite.repo.Create(group2)
	suite.NoError(err)

	group3 := suite.factories.Group.WithName("group-3")
	group3.OrgID = org.ID
	err = suite.repo.Create(group3)
	suite.NoError(err)

	// List all groups for the org
	groups, total, err := suite.repo.GetByOrganizationID(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(groups, 3)
	suite.Equal(int64(3), total)

	// Verify names are present
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Name
	}
	suite.Contains(names, "group-1")
	suite.Contains(names, "group-2")
	suite.Contains(names, "group-3")
}

// TestGetByOrganizationIDWithPagination tests listing groups with pagination
func (suite *GroupRepositoryTestSuite) TestGetByOrganizationIDWithPagination() {
	// Create organization
	org := suite.createOrganization()

	// Create multiple test groups under org (short names)
	for i := 0; i < 5; i++ {
		g := suite.factories.Group.WithName("grp-" + uuid.New().String()[:8])
		g.OrgID = org.ID
		err := suite.repo.Create(g)
		suite.NoError(err)
	}

	// First page
	groups, total, err := suite.repo.GetByOrganizationID(org.ID, 2, 0)
	suite.NoError(err)
	suite.Len(groups, 2)
	suite.GreaterOrEqual(total, int64(5))

	// Second page
	groups, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 2)
	suite.NoError(err)
	suite.Len(groups, 2)
	suite.GreaterOrEqual(total, int64(5))

	// Third page
	groups, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 4)
	suite.NoError(err)
	suite.True(len(groups) >= 1) // At least one left
	suite.GreaterOrEqual(total, int64(5))
}

// TestUpdate tests updating a group using map updates
func (suite *GroupRepositoryTestSuite) TestUpdate() {
	// Create organization
	org := suite.createOrganization()

	// Create test group
	group := suite.factories.Group.WithName("group-to-update")
	group.OrgID = org.ID
	err := suite.repo.Create(group)
	suite.NoError(err)

	// Update the group
	updates := map[string]interface{}{
		"title":       "Updated Group Title",
		"description": "Updated group description",
	}
	err = suite.repo.Update(group.ID, updates)

	// Assertions
	suite.NoError(err)

	// Retrieve updated group
	updatedGroup, err := suite.repo.GetByID(group.ID)
	suite.NoError(err)
	suite.Equal("Updated Group Title", updatedGroup.Title)
	suite.Equal("Updated group description", updatedGroup.Description)
	suite.True(updatedGroup.UpdatedAt.After(updatedGroup.CreatedAt))
}

// TestDelete tests deleting a group
func (suite *GroupRepositoryTestSuite) TestDelete() {
	// organization
	org := suite.createOrganization()

	// Create test group
	group := suite.factories.Group.WithName("group-to-delete")
	group.OrgID = org.ID
	err := suite.repo.Create(group)
	suite.NoError(err)

	// Delete the group
	err = suite.repo.Delete(group.ID)
	suite.NoError(err)

	// Verify group is deleted
	_, err = suite.repo.GetByID(group.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestDeleteNotFound tests deleting a non-existent group
func (suite *GroupRepositoryTestSuite) TestDeleteNotFound() {
	nonExistentID := uuid.New()

	err := suite.repo.Delete(nonExistentID)

	// Should not error when deleting non-existent record
	suite.NoError(err)
}

// TestSearch tests searching groups by name/title/description within an organization
func (suite *GroupRepositoryTestSuite) TestSearch() {
	// Create organization
	org := suite.createOrganization()

	// Create groups with different names/titles/descriptions
	alpha := suite.factories.Group.WithName("alpha-group")
	alpha.OrgID = org.ID
	alpha.Description = "Alpha description"
	err := suite.repo.Create(alpha)
	suite.NoError(err)

	beta := suite.factories.Group.WithName("beta-group")
	beta.OrgID = org.ID
	beta.Title = "Beta Title"
	err = suite.repo.Create(beta)
	suite.NoError(err)

	noMatch := suite.factories.Group.WithName("gamma-group")
	noMatch.OrgID = org.ID
	noMatch.Description = "No match here"
	err = suite.repo.Create(noMatch)
	suite.NoError(err)

	// Search by keyword "alpha" should return only alpha-group
	results, total, err := suite.repo.Search(org.ID, "alpha", 10, 0)
	suite.NoError(err)
	suite.Len(results, 1)
	suite.Equal(int64(1), total)
	suite.Equal("alpha-group", results[0].Name)

	// Search by keyword "beta" should return only beta-group
	results, total, err = suite.repo.Search(org.ID, "beta", 10, 0)
	suite.NoError(err)
	suite.Len(results, 1)
	suite.Equal(int64(1), total)
	suite.Equal("beta-group", results[0].Name)
}

// Run the test suite
func TestGroupRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(GroupRepositoryTestSuite))
}
