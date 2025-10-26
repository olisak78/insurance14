package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ProjectRepositoryTestSuite tests the ProjectRepository
type ProjectRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *ProjectRepository
	factories     *testutils.FactorySet
}

// SetupSuite runs before all tests in the suite
func (suite *ProjectRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	suite.repo = NewProjectRepository(suite.baseTestSuite.DB)
	suite.factories = testutils.NewFactorySet()
}

// TearDownSuite runs after all tests in the suite
func (suite *ProjectRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *ProjectRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *ProjectRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// TestCreate tests creating a new project
func (suite *ProjectRepositoryTestSuite) TestCreate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test project
	project := suite.factories.Project.Create()
	project.OrganizationID = org.ID

	// Create the project
	err = suite.repo.Create(project)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, project.ID)
	suite.NotZero(project.CreatedAt)
	suite.NotZero(project.UpdatedAt)
}

// TestCreateDuplicateName tests creating a project with duplicate name in same organization
func (suite *ProjectRepositoryTestSuite) TestCreateDuplicateName() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create first project
	project1 := suite.factories.Project.WithName("duplicate-project")
	project1.OrganizationID = org.ID
	err = suite.repo.Create(project1)
	suite.NoError(err)

	// Create second project with same name in same organization - should succeed since no unique constraint
	project2 := suite.factories.Project.WithName("duplicate-project")
	project2.OrganizationID = org.ID

	err = suite.repo.Create(project2)
	suite.NoError(err) // Should succeed since there's no unique constraint on project names

	// Verify both projects exist
	projects, total, err := suite.repo.GetByOrganizationID(org.ID, 10, 0)
	suite.NoError(err)
	suite.Len(projects, 2)
	suite.Equal(int64(2), total)
}

// TestCreateSameNameDifferentOrg tests creating projects with same name in different organizations
func (suite *ProjectRepositoryTestSuite) TestCreateSameNameDifferentOrg() {
	// Create two organizations with unique domains
	org1 := suite.factories.Organization.WithName("org1")
	org1.Domain = "org1.test.com"
	org2 := suite.factories.Organization.WithName("org2")
	org2.Domain = "org2.test.com"

	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org1)
	suite.NoError(err)
	err = orgRepo.Create(org2)
	suite.NoError(err)

	// Create projects with same name in different organizations
	project1 := suite.factories.Project.WithName("same-project")
	project1.OrganizationID = org1.ID
	err = suite.repo.Create(project1)
	suite.NoError(err)

	project2 := suite.factories.Project.WithName("same-project")
	project2.OrganizationID = org2.ID
	err = suite.repo.Create(project2)
	suite.NoError(err) // Should succeed
}

// TestGetByID tests retrieving a project by ID
func (suite *ProjectRepositoryTestSuite) TestGetByID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test project
	project := suite.factories.Project.Create()
	project.OrganizationID = org.ID
	err = suite.repo.Create(project)
	suite.NoError(err)

	// Retrieve the project
	retrievedProject, err := suite.repo.GetByID(project.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedProject)
	suite.Equal(project.ID, retrievedProject.ID)
	suite.Equal(project.Name, retrievedProject.Name)
	suite.Equal(project.DisplayName, retrievedProject.DisplayName)
	suite.Equal(project.ProjectType, retrievedProject.ProjectType)
	suite.Equal(project.OrganizationID, retrievedProject.OrganizationID)
}

// TestGetByIDNotFound tests retrieving a non-existent project
func (suite *ProjectRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	project, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(project)
}

// TestGetByOrganizationID tests listing projects by organization
func (suite *ProjectRepositoryTestSuite) TestGetByOrganizationID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test projects
	project1 := suite.factories.Project.WithName("project-1")
	project1.OrganizationID = org.ID
	err = suite.repo.Create(project1)
	suite.NoError(err)

	project2 := suite.factories.Project.WithName("project-2")
	project2.OrganizationID = org.ID
	err = suite.repo.Create(project2)
	suite.NoError(err)

	project3 := suite.factories.Project.WithName("project-3")
	project3.OrganizationID = org.ID
	err = suite.repo.Create(project3)
	suite.NoError(err)

	// List projects by organization
	projects, total, err := suite.repo.GetByOrganizationID(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(projects, 3)
	suite.Equal(int64(3), total)

	// Verify projects are returned
	names := make([]string, len(projects))
	for i, project := range projects {
		names[i] = project.Name
	}
	suite.Contains(names, "project-1")
	suite.Contains(names, "project-2")
	suite.Contains(names, "project-3")
}

// TestGetByOrganizationIDWithPagination tests listing projects with pagination
func (suite *ProjectRepositoryTestSuite) TestGetByOrganizationIDWithPagination() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test projects
	for i := 0; i < 5; i++ {
		project := suite.factories.Project.WithName(suite.T().Name() + "-project-" + uuid.New().String()[:8])
		project.OrganizationID = org.ID
		err := suite.repo.Create(project)
		suite.NoError(err)
	}

	// Test first page
	projects, total, err := suite.repo.GetByOrganizationID(org.ID, 2, 0)
	suite.NoError(err)
	suite.Len(projects, 2)
	suite.Equal(int64(5), total)

	// Test second page
	projects, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 2)
	suite.NoError(err)
	suite.Len(projects, 2)
	suite.Equal(int64(5), total)

	// Test third page
	projects, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 4)
	suite.NoError(err)
	suite.Len(projects, 1) // Only one left
	suite.Equal(int64(5), total)
}

// TestUpdate tests updating a project
func (suite *ProjectRepositoryTestSuite) TestUpdate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test project
	project := suite.factories.Project.Create()
	project.OrganizationID = org.ID
	err = suite.repo.Create(project)
	suite.NoError(err)

	// Update the project
	project.DisplayName = "Updated Project Display Name"
	project.Description = "Updated project description"
	project.ProjectType = models.ProjectTypeLibrary
	project.Status = models.ProjectStatusInactive
	project.SortOrder = 100

	err = suite.repo.Update(project)

	// Assertions
	suite.NoError(err)

	// Retrieve updated project
	updatedProject, err := suite.repo.GetByID(project.ID)
	suite.NoError(err)
	suite.Equal("Updated Project Display Name", updatedProject.DisplayName)
	suite.Equal("Updated project description", updatedProject.Description)
	suite.Equal(models.ProjectTypeLibrary, updatedProject.ProjectType)
	suite.Equal(models.ProjectStatusInactive, updatedProject.Status)
	suite.Equal(100, updatedProject.SortOrder)
	suite.True(updatedProject.UpdatedAt.After(updatedProject.CreatedAt))
}

// TestDelete tests deleting a project
func (suite *ProjectRepositoryTestSuite) TestDelete() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test project
	project := suite.factories.Project.Create()
	project.OrganizationID = org.ID
	err = suite.repo.Create(project)
	suite.NoError(err)

	// Delete the project
	err = suite.repo.Delete(project.ID)
	suite.NoError(err)

	// Verify project is deleted
	_, err = suite.repo.GetByID(project.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestDeleteNotFound tests deleting a non-existent project
func (suite *ProjectRepositoryTestSuite) TestDeleteNotFound() {
	nonExistentID := uuid.New()

	err := suite.repo.Delete(nonExistentID)

	// Should not error when deleting non-existent record
	suite.NoError(err)
}

// TestGetByName tests retrieving a project by name within organization
func (suite *ProjectRepositoryTestSuite) TestGetByName() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test project
	project := suite.factories.Project.WithName("unique-project-name")
	project.OrganizationID = org.ID
	err = suite.repo.Create(project)
	suite.NoError(err)

	// Retrieve the project by name
	retrievedProject, err := suite.repo.GetByName(org.ID, "unique-project-name")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedProject)
	suite.Equal(project.ID, retrievedProject.ID)
	suite.Equal("unique-project-name", retrievedProject.Name)
}

// TestGetByNameNotFound tests retrieving a non-existent project by name
func (suite *ProjectRepositoryTestSuite) TestGetByNameNotFound() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	project, err := suite.repo.GetByName(org.ID, "nonexistent-project")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(project)
}

// TestGetByStatus tests retrieving projects by status
func (suite *ProjectRepositoryTestSuite) TestGetByStatus() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create projects with different statuses
	active1 := suite.factories.Project.WithName("active-1")
	active1.OrganizationID = org.ID
	active1.Status = models.ProjectStatusActive
	err = suite.repo.Create(active1)
	suite.NoError(err)

	active2 := suite.factories.Project.WithName("active-2")
	active2.OrganizationID = org.ID
	active2.Status = models.ProjectStatusActive
	err = suite.repo.Create(active2)
	suite.NoError(err)

	inactive1 := suite.factories.Project.WithName("inactive-1")
	inactive1.OrganizationID = org.ID
	inactive1.Status = models.ProjectStatusInactive
	err = suite.repo.Create(inactive1)
	suite.NoError(err)

	// Get projects by active status
	activeProjects, total, err := suite.repo.GetByStatus(org.ID, models.ProjectStatusActive, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(activeProjects, 2)
	suite.Equal(int64(2), total)

	// Verify all returned projects are active
	for _, project := range activeProjects {
		suite.Equal(models.ProjectStatusActive, project.Status)
	}
}

// TestGetActiveProjects tests retrieving active projects
func (suite *ProjectRepositoryTestSuite) TestGetActiveProjects() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create projects with different statuses
	active1 := suite.factories.Project.WithName("active-1")
	active1.OrganizationID = org.ID
	active1.Status = models.ProjectStatusActive
	err = suite.repo.Create(active1)
	suite.NoError(err)

	active2 := suite.factories.Project.WithName("active-2")
	active2.OrganizationID = org.ID
	active2.Status = models.ProjectStatusActive
	err = suite.repo.Create(active2)
	suite.NoError(err)

	inactive1 := suite.factories.Project.WithName("inactive-1")
	inactive1.OrganizationID = org.ID
	inactive1.Status = models.ProjectStatusInactive
	err = suite.repo.Create(inactive1)
	suite.NoError(err)

	// Get active projects
	activeProjects, total, err := suite.repo.GetActiveProjects(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(activeProjects, 2)
	suite.Equal(int64(2), total)

	// Verify all returned projects are active
	for _, project := range activeProjects {
		suite.Equal(models.ProjectStatusActive, project.Status)
	}
}

// Run the test suite
func TestProjectRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectRepositoryTestSuite))
}
