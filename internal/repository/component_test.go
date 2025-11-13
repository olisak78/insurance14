package repository

import (
	"testing"

	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ComponentRepositoryTestSuite tests the ComponentRepository
type ComponentRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *ComponentRepository
	factories     *testutils.FactorySet
}

// SetupSuite runs before all tests in the suite
func (suite *ComponentRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	suite.repo = NewComponentRepository(suite.baseTestSuite.DB)
	suite.factories = testutils.NewFactorySet()
}

// TearDownSuite runs after all tests in the suite
func (suite *ComponentRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *ComponentRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *ComponentRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// TestCreate tests creating a new component
func (suite *ComponentRepositoryTestSuite) TestCreate() {
	// Create project
	project := suite.factories.Project.Create()
	projectRepo := NewProjectRepository(suite.baseTestSuite.DB)
	err := projectRepo.Create(project)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.ProjectID = project.ID

	// Create the component
	err = suite.repo.Create(component)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, component.ID)
	suite.NotZero(component.CreatedAt)
	suite.NotZero(component.UpdatedAt)
}


// TestGetByID tests retrieving a component by ID
func (suite *ComponentRepositoryTestSuite) TestGetByID() {
	// Create project
	project := suite.factories.Project.Create()
	projectRepo := NewProjectRepository(suite.baseTestSuite.DB)
	err := projectRepo.Create(project)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.ProjectID = project.ID
	err = suite.repo.Create(component)
	suite.NoError(err)

	// Retrieve the component
	retrievedComponent, err := suite.repo.GetByID(component.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedComponent)
	suite.Equal(component.ID, retrievedComponent.ID)
	suite.Equal(component.Name, retrievedComponent.Name)
	suite.Equal(component.Title, retrievedComponent.Title)
	suite.Equal(component.ProjectID, retrievedComponent.ProjectID)
	suite.Equal(component.OwnerID, retrievedComponent.OwnerID)
}

// TestGetByIDNotFound tests retrieving a non-existent component
func (suite *ComponentRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	component, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(component)
}

// TestUpdate tests updating a component
func (suite *ComponentRepositoryTestSuite) TestUpdate() {
	// Create project
	project := suite.factories.Project.Create()
	projectRepo := NewProjectRepository(suite.baseTestSuite.DB)
	err := projectRepo.Create(project)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.ProjectID = project.ID
	err = suite.repo.Create(component)
	suite.NoError(err)

	// Update the component
	component.Title = "Updated Component Title"
	component.Description = "Updated component description"

	err = suite.repo.Update(component)

	// Assertions
	suite.NoError(err)

	// Retrieve updated component
	updatedComponent, err := suite.repo.GetByID(component.ID)
	suite.NoError(err)
	suite.Equal("Updated Component Title", updatedComponent.Title)
	suite.Equal("Updated component description", updatedComponent.Description)
	suite.True(updatedComponent.UpdatedAt.After(updatedComponent.CreatedAt))
}

// TestDelete tests deleting a component
func (suite *ComponentRepositoryTestSuite) TestDelete() {
	// Create project
	project := suite.factories.Project.Create()
	projectRepo := NewProjectRepository(suite.baseTestSuite.DB)
	err := projectRepo.Create(project)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.ProjectID = project.ID
	err = suite.repo.Create(component)
	suite.NoError(err)

	// Delete the component
	err = suite.repo.Delete(component.ID)
	suite.NoError(err)

	// Verify component is deleted
	_, err = suite.repo.GetByID(component.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestDeleteNotFound tests deleting a non-existent component
func (suite *ComponentRepositoryTestSuite) TestDeleteNotFound() {
	nonExistentID := uuid.New()

	err := suite.repo.Delete(nonExistentID)

	// Should not error when deleting non-existent record
	suite.NoError(err)
}

// TestGetByName tests retrieving a component by name within project
func (suite *ComponentRepositoryTestSuite) TestGetByName() {
	// Create project
	project := suite.factories.Project.Create()
	projectRepo := NewProjectRepository(suite.baseTestSuite.DB)
	err := projectRepo.Create(project)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.WithName("unique-component-name")
	component.ProjectID = project.ID
	err = suite.repo.Create(component)
	suite.NoError(err)

	// Retrieve the component by name
	retrievedComponent, err := suite.repo.GetByName(project.ID, "unique-component-name")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedComponent)
	suite.Equal(component.ID, retrievedComponent.ID)
	suite.Equal("unique-component-name", retrievedComponent.Name)
}

// TestGetByNameNotFound tests retrieving a non-existent component by name
func (suite *ComponentRepositoryTestSuite) TestGetByNameNotFound() {
	// Create project
	project := suite.factories.Project.Create()
	projectRepo := NewProjectRepository(suite.baseTestSuite.DB)
	err := projectRepo.Create(project)
	suite.NoError(err)

	component, err := suite.repo.GetByName(project.ID, "nonexistent-component")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(component)
}

// Run the test suite
func TestComponentRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentRepositoryTestSuite))
}
