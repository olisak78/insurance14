package repository

import (
	"testing"

	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// OrganizationRepositoryTestSuite tests the OrganizationRepository
type OrganizationRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *OrganizationRepository
	factories     *testutils.FactorySet
}

// SetupSuite runs before all tests in the suite
func (suite *OrganizationRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	suite.repo = NewOrganizationRepository(suite.baseTestSuite.DB)
	suite.factories = testutils.NewFactorySet()
}

// TearDownSuite runs after all tests in the suite
func (suite *OrganizationRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *OrganizationRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *OrganizationRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// TestCreate tests creating a new organization
func (suite *OrganizationRepositoryTestSuite) TestCreate() {
	// Create test organization
	org := suite.factories.Organization.Create()

	// Create the organization
	err := suite.repo.Create(org)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, org.ID)
	suite.NotZero(org.CreatedAt)
	suite.NotZero(org.UpdatedAt)
}

// TestCreateDuplicateName tests creating an organization with duplicate name
func (suite *OrganizationRepositoryTestSuite) TestCreateDuplicateName() {
	// Create first organization
	org1 := suite.factories.Organization.WithName("test-org")
	err := suite.repo.Create(org1)
	suite.NoError(err)

	// Try to create second organization with same name
	org2 := suite.factories.Organization.WithName("test-org")
	org2.Domain = "different-domain.com" // Different domain to avoid domain constraint

	err = suite.repo.Create(org2)
	suite.Error(err)
	suite.Contains(err.Error(), "duplicate key value")
}

// TestCreateDuplicateDomain tests creating an organization with duplicate domain
func (suite *OrganizationRepositoryTestSuite) TestCreateDuplicateDomain() {
	// Create first organization
	org1 := suite.factories.Organization.WithDomain("test.com")
	err := suite.repo.Create(org1)
	suite.NoError(err)

	// Try to create second organization with same domain
	org2 := suite.factories.Organization.WithName("different-org")
	org2.Domain = "test.com" // Same domain

	err = suite.repo.Create(org2)
	suite.Error(err)
	suite.Contains(err.Error(), "duplicate key value")
}

// TestGetByID tests retrieving an organization by ID
func (suite *OrganizationRepositoryTestSuite) TestGetByID() {
	// Create test organization
	org := suite.factories.Organization.Create()
	err := suite.repo.Create(org)
	suite.NoError(err)

	// Retrieve the organization
	retrievedOrg, err := suite.repo.GetByID(org.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedOrg)
	suite.Equal(org.ID, retrievedOrg.ID)
	suite.Equal(org.Name, retrievedOrg.Name)
	suite.Equal(org.DisplayName, retrievedOrg.DisplayName)
	suite.Equal(org.Domain, retrievedOrg.Domain)
}

// TestGetByIDNotFound tests retrieving a non-existent organization
func (suite *OrganizationRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	org, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(org)
}

// TestGetByName tests retrieving an organization by name
func (suite *OrganizationRepositoryTestSuite) TestGetByName() {
	// Create test organization
	org := suite.factories.Organization.WithName("test-organization")
	err := suite.repo.Create(org)
	suite.NoError(err)

	// Retrieve the organization by name
	retrievedOrg, err := suite.repo.GetByName("test-organization")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedOrg)
	suite.Equal(org.ID, retrievedOrg.ID)
	suite.Equal("test-organization", retrievedOrg.Name)
}

// TestGetByNameNotFound tests retrieving a non-existent organization by name
func (suite *OrganizationRepositoryTestSuite) TestGetByNameNotFound() {
	org, err := suite.repo.GetByName("non-existent-org")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(org)
}

// TestGetByDomain tests retrieving an organization by domain
func (suite *OrganizationRepositoryTestSuite) TestGetByDomain() {
	// Create test organization
	org := suite.factories.Organization.WithDomain("example.com")
	err := suite.repo.Create(org)
	suite.NoError(err)

	// Retrieve the organization by domain
	retrievedOrg, err := suite.repo.GetByDomain("example.com")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedOrg)
	suite.Equal(org.ID, retrievedOrg.ID)
	suite.Equal("example.com", retrievedOrg.Domain)
}

// TestGetByDomainNotFound tests retrieving a non-existent organization by domain
func (suite *OrganizationRepositoryTestSuite) TestGetByDomainNotFound() {
	org, err := suite.repo.GetByDomain("nonexistent.com")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(org)
}

// TestGetAll tests listing organizations
func (suite *OrganizationRepositoryTestSuite) TestGetAll() {
	// Create multiple test organizations
	org1 := suite.factories.Organization.WithName("org-1")
	org1.Domain = "org1.com"
	err := suite.repo.Create(org1)
	suite.NoError(err)

	org2 := suite.factories.Organization.WithName("org-2")
	org2.Domain = "org2.com"
	err = suite.repo.Create(org2)
	suite.NoError(err)

	org3 := suite.factories.Organization.WithName("org-3")
	org3.Domain = "org3.com"
	err = suite.repo.Create(org3)
	suite.NoError(err)

	// List all organizations
	orgs, total, err := suite.repo.GetAll(10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(orgs, 3)
	suite.Equal(int64(3), total)

	// Verify organizations are returned
	names := make([]string, len(orgs))
	for i, org := range orgs {
		names[i] = org.Name
	}
	suite.Contains(names, "org-1")
	suite.Contains(names, "org-2")
	suite.Contains(names, "org-3")
}

// TestGetAllWithPagination tests listing organizations with pagination
func (suite *OrganizationRepositoryTestSuite) TestGetAllWithPagination() {
	// Create multiple test organizations
	for i := 0; i < 5; i++ {
		org := suite.factories.Organization.WithName(suite.T().Name() + "-org-" + uuid.New().String()[:8])
		org.Domain = uuid.New().String()[:8] + ".com"
		err := suite.repo.Create(org)
		suite.NoError(err)
	}

	// Test first page
	orgs, total, err := suite.repo.GetAll(2, 0)
	suite.NoError(err)
	suite.Len(orgs, 2)
	suite.Equal(int64(5), total)

	// Test second page
	orgs, total, err = suite.repo.GetAll(2, 2)
	suite.NoError(err)
	suite.Len(orgs, 2)
	suite.Equal(int64(5), total)

	// Test third page
	orgs, total, err = suite.repo.GetAll(2, 4)
	suite.NoError(err)
	suite.Len(orgs, 1) // Only one left
	suite.Equal(int64(5), total)
}

// TestUpdate tests updating an organization
func (suite *OrganizationRepositoryTestSuite) TestUpdate() {
	// Create test organization
	org := suite.factories.Organization.Create()
	err := suite.repo.Create(org)
	suite.NoError(err)

	// Update the organization
	org.DisplayName = "Updated Display Name"
	org.Description = "Updated description"

	err = suite.repo.Update(org)

	// Assertions
	suite.NoError(err)

	// Retrieve updated organization
	updatedOrg, err := suite.repo.GetByID(org.ID)
	suite.NoError(err)
	suite.Equal("Updated Display Name", updatedOrg.DisplayName)
	suite.Equal("Updated description", updatedOrg.Description)
	suite.True(updatedOrg.UpdatedAt.After(updatedOrg.CreatedAt))
}

// TestDelete tests deleting an organization
func (suite *OrganizationRepositoryTestSuite) TestDelete() {
	// Create test organization
	org := suite.factories.Organization.Create()
	err := suite.repo.Create(org)
	suite.NoError(err)

	// Delete the organization
	err = suite.repo.Delete(org.ID)
	suite.NoError(err)

	// Verify organization is deleted
	_, err = suite.repo.GetByID(org.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestDeleteNotFound tests deleting a non-existent organization
func (suite *OrganizationRepositoryTestSuite) TestDeleteNotFound() {
	nonExistentID := uuid.New()

	err := suite.repo.Delete(nonExistentID)

	// Should not error when deleting non-existent record
	suite.NoError(err)
}

// TestGetWithMembers tests retrieving organization with members
func (suite *OrganizationRepositoryTestSuite) TestGetWithMembers() {
	// Create organization and member
	org := suite.factories.Organization.Create()
	err := suite.repo.Create(org)
	suite.NoError(err)

	member := suite.factories.Member.Create()
	member.OrganizationID = org.ID
	memberRepo := NewMemberRepository(suite.baseTestSuite.DB)
	err = memberRepo.Create(member)
	suite.NoError(err)

	// Retrieve organization with members
	orgWithMembers, err := suite.repo.GetWithMembers(org.ID)

	suite.NoError(err)
	suite.NotNil(orgWithMembers)
	suite.NotEmpty(orgWithMembers.Members)
	suite.Equal(member.ID, orgWithMembers.Members[0].ID)
}

// TestGetWithGroups tests retrieving organization with groups
func (suite *OrganizationRepositoryTestSuite) TestGetWithGroups() {
	// Create organization and group
	org := suite.factories.Organization.Create()
	err := suite.repo.Create(org)
	suite.NoError(err)

	group := suite.factories.Group.Create()
	group.OrganizationID = org.ID
	groupRepo := NewGroupRepository(suite.baseTestSuite.DB)
	err = groupRepo.Create(group)
	suite.NoError(err)

	// Retrieve organization with groups
	orgWithGroups, err := suite.repo.GetWithGroups(org.ID)

	suite.NoError(err)
	suite.NotNil(orgWithGroups)
	suite.NotEmpty(orgWithGroups.Groups)
	suite.Equal(group.ID, orgWithGroups.Groups[0].ID)
}

// TestGetWithAllRelations tests retrieving organization with all relations
func (suite *OrganizationRepositoryTestSuite) TestGetWithAllRelations() {
	// Create organization
	org := suite.factories.Organization.Create()
	err := suite.repo.Create(org)
	suite.NoError(err)

	// Create member
	member := suite.factories.Member.Create()
	member.OrganizationID = org.ID
	memberRepo := NewMemberRepository(suite.baseTestSuite.DB)
	err = memberRepo.Create(member)
	suite.NoError(err)

	// Create group
	group := suite.factories.Group.Create()
	group.OrganizationID = org.ID
	groupRepo := NewGroupRepository(suite.baseTestSuite.DB)
	err = groupRepo.Create(group)
	suite.NoError(err)

	// Create project
	project := suite.factories.Project.Create()
	project.OrganizationID = org.ID
	projectRepo := NewProjectRepository(suite.baseTestSuite.DB)
	err = projectRepo.Create(project)
	suite.NoError(err)

	// Create component
	component := suite.factories.Component.Create()
	component.OrganizationID = org.ID
	componentRepo := NewComponentRepository(suite.baseTestSuite.DB)
	err = componentRepo.Create(component)
	suite.NoError(err)

	// Create landscape
	landscape := suite.factories.Landscape.Create()
	landscape.OrganizationID = org.ID
	landscapeRepo := NewLandscapeRepository(suite.baseTestSuite.DB)
	err = landscapeRepo.Create(landscape)
	suite.NoError(err)

	// Retrieve organization with all relations
	orgWithRelations, err := suite.repo.GetWithAllRelations(org.ID)

	suite.NoError(err)
	suite.NotNil(orgWithRelations)
	suite.NotEmpty(orgWithRelations.Members)
	suite.NotEmpty(orgWithRelations.Groups)
	suite.NotEmpty(orgWithRelations.Projects)
	suite.NotEmpty(orgWithRelations.Components)
	suite.NotEmpty(orgWithRelations.Landscapes)
}

// Run the test suite
func TestOrganizationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationRepositoryTestSuite))
}
