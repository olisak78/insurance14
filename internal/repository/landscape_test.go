package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// LandscapeRepositoryTestSuite tests the LandscapeRepository
type LandscapeRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *LandscapeRepository
	factories     *testutils.FactorySet
}

// SetupSuite runs before all tests in the suite
func (suite *LandscapeRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	suite.repo = NewLandscapeRepository(suite.baseTestSuite.DB)
	suite.factories = testutils.NewFactorySet()
}

// TearDownSuite runs after all tests in the suite
func (suite *LandscapeRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *LandscapeRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *LandscapeRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// TestCreate tests creating a new landscape
func (suite *LandscapeRepositoryTestSuite) TestCreate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test landscape
	landscape := suite.factories.Landscape.Create()
	landscape.OrganizationID = org.ID

	// Create the landscape
	err = suite.repo.Create(landscape)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, landscape.ID)
	suite.NotZero(landscape.CreatedAt)
	suite.NotZero(landscape.UpdatedAt)
}

// TestCreateDuplicateName tests creating a landscape with duplicate name
func (suite *LandscapeRepositoryTestSuite) TestCreateDuplicateName() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create first landscape
	landscape1 := suite.factories.Landscape.WithName("duplicate-landscape")
	landscape1.OrganizationID = org.ID
	err = suite.repo.Create(landscape1)
	suite.NoError(err)

	// Try to create second landscape with same name in same organization
	landscape2 := suite.factories.Landscape.WithName("duplicate-landscape")
	landscape2.OrganizationID = org.ID
	err = suite.repo.Create(landscape2)

	// Should fail due to unique constraint on (organization_id, name)
	suite.Error(err)
	suite.Contains(err.Error(), "duplicate key value")
}

// TestGetByID tests retrieving a landscape by ID
func (suite *LandscapeRepositoryTestSuite) TestGetByID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test landscape
	landscape := suite.factories.Landscape.Create()
	landscape.OrganizationID = org.ID
	err = suite.repo.Create(landscape)
	suite.NoError(err)

	// Retrieve the landscape
	retrievedLandscape, err := suite.repo.GetByID(landscape.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedLandscape)
	suite.Equal(landscape.ID, retrievedLandscape.ID)
	suite.Equal(landscape.Name, retrievedLandscape.Name)
	suite.Equal(landscape.DisplayName, retrievedLandscape.DisplayName)
	suite.Equal(landscape.Status, retrievedLandscape.Status)
}

// TestGetByIDNotFound tests retrieving a non-existent landscape
func (suite *LandscapeRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	landscape, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(landscape)
}

// TestGetByOrganizationID tests listing landscapes by organization
func (suite *LandscapeRepositoryTestSuite) TestGetByOrganizationID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test landscapes
	landscape1 := suite.factories.Landscape.WithName("landscape-1")
	landscape1.OrganizationID = org.ID
	err = suite.repo.Create(landscape1)
	suite.NoError(err)

	landscape2 := suite.factories.Landscape.WithName("landscape-2")
	landscape2.OrganizationID = org.ID
	err = suite.repo.Create(landscape2)
	suite.NoError(err)

	landscape3 := suite.factories.Landscape.WithName("landscape-3")
	landscape3.OrganizationID = org.ID
	err = suite.repo.Create(landscape3)
	suite.NoError(err)

	// List landscapes by organization
	landscapes, total, err := suite.repo.GetByOrganizationID(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(landscapes, 3)
	suite.Equal(int64(3), total)

	// Verify landscapes are returned
	names := make([]string, len(landscapes))
	for i, landscape := range landscapes {
		names[i] = landscape.Name
	}
	suite.Contains(names, "landscape-1")
	suite.Contains(names, "landscape-2")
	suite.Contains(names, "landscape-3")
}

// TestGetByOrganizationIDWithPagination tests listing landscapes with pagination
func (suite *LandscapeRepositoryTestSuite) TestGetByOrganizationIDWithPagination() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test landscapes
	for i := 0; i < 5; i++ {
		landscape := suite.factories.Landscape.WithName(suite.T().Name() + "-landscape-" + uuid.New().String()[:8])
		landscape.OrganizationID = org.ID
		err := suite.repo.Create(landscape)
		suite.NoError(err)
	}

	// Test first page
	landscapes, total, err := suite.repo.GetByOrganizationID(org.ID, 2, 0)
	suite.NoError(err)
	suite.Len(landscapes, 2)
	suite.Equal(int64(5), total)

	// Test second page
	landscapes, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 2)
	suite.NoError(err)
	suite.Len(landscapes, 2)
	suite.Equal(int64(5), total)

	// Test third page
	landscapes, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 4)
	suite.NoError(err)
	suite.Len(landscapes, 1) // Only one left
	suite.Equal(int64(5), total)
}

// TestUpdate tests updating a landscape
func (suite *LandscapeRepositoryTestSuite) TestUpdate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test landscape
	landscape := suite.factories.Landscape.Create()
	landscape.OrganizationID = org.ID
	err = suite.repo.Create(landscape)
	suite.NoError(err)

	// Update the landscape
	landscape.DisplayName = "Updated Landscape Display Name"
	landscape.Description = "Updated landscape description"
	landscape.Status = models.LandscapeStatusRetired

	err = suite.repo.Update(landscape)

	// Assertions
	suite.NoError(err)

	// Retrieve updated landscape
	updatedLandscape, err := suite.repo.GetByID(landscape.ID)
	suite.NoError(err)
	suite.Equal("Updated Landscape Display Name", updatedLandscape.DisplayName)
	suite.Equal("Updated landscape description", updatedLandscape.Description)
	suite.Equal(models.LandscapeStatusRetired, updatedLandscape.Status)
	suite.True(updatedLandscape.UpdatedAt.After(updatedLandscape.CreatedAt))
}

// TestDelete tests deleting a landscape
func (suite *LandscapeRepositoryTestSuite) TestDelete() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test landscape
	landscape := suite.factories.Landscape.Create()
	landscape.OrganizationID = org.ID
	err = suite.repo.Create(landscape)
	suite.NoError(err)

	// Delete the landscape
	err = suite.repo.Delete(landscape.ID)
	suite.NoError(err)

	// Verify landscape is deleted
	_, err = suite.repo.GetByID(landscape.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestDeleteNotFound tests deleting a non-existent landscape
func (suite *LandscapeRepositoryTestSuite) TestDeleteNotFound() {
	nonExistentID := uuid.New()

	err := suite.repo.Delete(nonExistentID)

	// Should not error when deleting non-existent record
	suite.NoError(err)
}

// TestGetByName tests retrieving a landscape by name within organization
func (suite *LandscapeRepositoryTestSuite) TestGetByName() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test landscape
	landscape := suite.factories.Landscape.WithName("unique-landscape-name")
	landscape.OrganizationID = org.ID
	err = suite.repo.Create(landscape)
	suite.NoError(err)

	// Retrieve the landscape by name
	retrievedLandscape, err := suite.repo.GetByName(org.ID, "unique-landscape-name")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedLandscape)
	suite.Equal(landscape.ID, retrievedLandscape.ID)
	suite.Equal("unique-landscape-name", retrievedLandscape.Name)
}

// TestGetByNameNotFound tests retrieving a non-existent landscape by name
func (suite *LandscapeRepositoryTestSuite) TestGetByNameNotFound() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	landscape, err := suite.repo.GetByName(org.ID, "nonexistent-landscape")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(landscape)
}

// TestGetByStatus tests retrieving landscapes by status
func (suite *LandscapeRepositoryTestSuite) TestGetByStatus() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create landscapes with different statuses
	active1 := suite.factories.Landscape.WithName("active-landscape-1")
	active1.OrganizationID = org.ID
	active1.Status = models.LandscapeStatusActive
	err = suite.repo.Create(active1)
	suite.NoError(err)

	active2 := suite.factories.Landscape.WithName("active-landscape-2")
	active2.OrganizationID = org.ID
	active2.Status = models.LandscapeStatusActive
	err = suite.repo.Create(active2)
	suite.NoError(err)

	inactive1 := suite.factories.Landscape.WithName("inactive-landscape-1")
	inactive1.OrganizationID = org.ID
	inactive1.Status = models.LandscapeStatusInactive
	err = suite.repo.Create(inactive1)
	suite.NoError(err)

	// Get landscapes by active status
	activeLandscapes, total, err := suite.repo.GetByStatus(org.ID, models.LandscapeStatusActive, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(activeLandscapes, 2)
	suite.Equal(int64(2), total)

	// Verify all returned landscapes are active
	for _, landscape := range activeLandscapes {
		suite.Equal(models.LandscapeStatusActive, landscape.Status)
	}
}

// TestGetActiveLandscapes tests retrieving active landscapes
func (suite *LandscapeRepositoryTestSuite) TestGetActiveLandscapes() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create landscapes with different statuses
	active1 := suite.factories.Landscape.WithName("active-for-test-1")
	active1.OrganizationID = org.ID
	active1.Status = models.LandscapeStatusActive
	err = suite.repo.Create(active1)
	suite.NoError(err)

	active2 := suite.factories.Landscape.WithName("active-for-test-2")
	active2.OrganizationID = org.ID
	active2.Status = models.LandscapeStatusActive
	err = suite.repo.Create(active2)
	suite.NoError(err)

	inactive1 := suite.factories.Landscape.WithName("inactive-for-test-1")
	inactive1.OrganizationID = org.ID
	inactive1.Status = models.LandscapeStatusInactive
	err = suite.repo.Create(inactive1)
	suite.NoError(err)

	// Get active landscapes
	activeLandscapes, total, err := suite.repo.GetActiveLandscapes(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(activeLandscapes, 2)
	suite.Equal(int64(2), total)

	// Verify all returned landscapes are active
	for _, landscape := range activeLandscapes {
		suite.Equal(models.LandscapeStatusActive, landscape.Status)
	}
}

// Run the test suite
func TestLandscapeRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(LandscapeRepositoryTestSuite))
}
