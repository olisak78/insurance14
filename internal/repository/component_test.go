package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
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
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.OrganizationID = org.ID

	// Create the component
	err = suite.repo.Create(component)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, component.ID)
	suite.NotZero(component.CreatedAt)
	suite.NotZero(component.UpdatedAt)
}

// TestCreateDuplicateName tests creating a component with duplicate name
func (suite *ComponentRepositoryTestSuite) TestCreateDuplicateName() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create first component
	component1 := suite.factories.Component.WithName("duplicate-component")
	component1.OrganizationID = org.ID
	err = suite.repo.Create(component1)
	suite.NoError(err)

	// Try to create second component with same name in same organization
	component2 := suite.factories.Component.WithName("duplicate-component")
	component2.OrganizationID = org.ID
	err = suite.repo.Create(component2)

	// Should fail due to unique constraint on (organization_id, name)
	suite.Error(err)
	suite.Contains(err.Error(), "duplicate key value")
}

// TestGetByID tests retrieving a component by ID
func (suite *ComponentRepositoryTestSuite) TestGetByID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.OrganizationID = org.ID
	err = suite.repo.Create(component)
	suite.NoError(err)

	// Retrieve the component
	retrievedComponent, err := suite.repo.GetByID(component.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedComponent)
	suite.Equal(component.ID, retrievedComponent.ID)
	suite.Equal(component.Name, retrievedComponent.Name)
	suite.Equal(component.DisplayName, retrievedComponent.DisplayName)
	suite.Equal(component.ComponentType, retrievedComponent.ComponentType)
	suite.Equal(component.Status, retrievedComponent.Status)
}

// TestGetByIDNotFound tests retrieving a non-existent component
func (suite *ComponentRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	component, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(component)
}

// TestGetByOrganizationID tests listing components by organization
func (suite *ComponentRepositoryTestSuite) TestGetByOrganizationID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test components
	component1 := suite.factories.Component.WithName("component-1")
	component1.OrganizationID = org.ID
	err = suite.repo.Create(component1)
	suite.NoError(err)

	component2 := suite.factories.Component.WithName("component-2")
	component2.OrganizationID = org.ID
	err = suite.repo.Create(component2)
	suite.NoError(err)

	component3 := suite.factories.Component.WithName("component-3")
	component3.OrganizationID = org.ID
	err = suite.repo.Create(component3)
	suite.NoError(err)

	// List components by organization
	components, total, err := suite.repo.GetByOrganizationID(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(components, 3)
	suite.Equal(int64(3), total)

	// Verify components are returned
	names := make([]string, len(components))
	for i, component := range components {
		names[i] = component.Name
	}
	suite.Contains(names, "component-1")
	suite.Contains(names, "component-2")
	suite.Contains(names, "component-3")
}

// TestGetByOrganizationIDWithPagination tests listing components with pagination
func (suite *ComponentRepositoryTestSuite) TestGetByOrganizationIDWithPagination() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test components
	for i := 0; i < 5; i++ {
		component := suite.factories.Component.WithName(suite.T().Name() + "-component-" + uuid.New().String()[:8])
		component.OrganizationID = org.ID
		err := suite.repo.Create(component)
		suite.NoError(err)
	}

	// Test first page
	components, total, err := suite.repo.GetByOrganizationID(org.ID, 2, 0)
	suite.NoError(err)
	suite.Len(components, 2)
	suite.Equal(int64(5), total)

	// Test second page
	components, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 2)
	suite.NoError(err)
	suite.Len(components, 2)
	suite.Equal(int64(5), total)

	// Test third page
	components, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 4)
	suite.NoError(err)
	suite.Len(components, 1) // Only one left
	suite.Equal(int64(5), total)
}

// TestUpdate tests updating a component
func (suite *ComponentRepositoryTestSuite) TestUpdate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.OrganizationID = org.ID
	err = suite.repo.Create(component)
	suite.NoError(err)

	// Update the component
	component.DisplayName = "Updated Component Display Name"
	component.Description = "Updated component description"
	component.ComponentType = models.ComponentTypeLibrary
	component.Status = models.ComponentStatusMaintenance

	err = suite.repo.Update(component)

	// Assertions
	suite.NoError(err)

	// Retrieve updated component
	updatedComponent, err := suite.repo.GetByID(component.ID)
	suite.NoError(err)
	suite.Equal("Updated Component Display Name", updatedComponent.DisplayName)
	suite.Equal("Updated component description", updatedComponent.Description)
	suite.Equal(models.ComponentTypeLibrary, updatedComponent.ComponentType)
	suite.Equal(models.ComponentStatusMaintenance, updatedComponent.Status)
	suite.True(updatedComponent.UpdatedAt.After(updatedComponent.CreatedAt))
}

// TestDelete tests deleting a component
func (suite *ComponentRepositoryTestSuite) TestDelete() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.Create()
	component.OrganizationID = org.ID
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

// TestGetByName tests retrieving a component by name within organization
func (suite *ComponentRepositoryTestSuite) TestGetByName() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test component
	component := suite.factories.Component.WithName("unique-component-name")
	component.OrganizationID = org.ID
	err = suite.repo.Create(component)
	suite.NoError(err)

	// Retrieve the component by name
	retrievedComponent, err := suite.repo.GetByName(org.ID, "unique-component-name")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedComponent)
	suite.Equal(component.ID, retrievedComponent.ID)
	suite.Equal("unique-component-name", retrievedComponent.Name)
}

// TestGetByNameNotFound tests retrieving a non-existent component by name
func (suite *ComponentRepositoryTestSuite) TestGetByNameNotFound() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	component, err := suite.repo.GetByName(org.ID, "nonexistent-component")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(component)
}

// TestGetByType tests retrieving components by type
func (suite *ComponentRepositoryTestSuite) TestGetByType() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create components with different types
	service1 := suite.factories.Component.WithName("service-1")
	service1.OrganizationID = org.ID
	service1.ComponentType = models.ComponentTypeService
	err = suite.repo.Create(service1)
	suite.NoError(err)

	service2 := suite.factories.Component.WithName("service-2")
	service2.OrganizationID = org.ID
	service2.ComponentType = models.ComponentTypeService
	err = suite.repo.Create(service2)
	suite.NoError(err)

	library1 := suite.factories.Component.WithName("library-1")
	library1.OrganizationID = org.ID
	library1.ComponentType = models.ComponentTypeLibrary
	err = suite.repo.Create(library1)
	suite.NoError(err)

	// Get components by service type
	serviceComponents, total, err := suite.repo.GetByType(org.ID, models.ComponentTypeService, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(serviceComponents, 2)
	suite.Equal(int64(2), total)

	// Verify all returned components are services
	for _, component := range serviceComponents {
		suite.Equal(models.ComponentTypeService, component.ComponentType)
	}
}

// TestGetByStatus tests retrieving components by status
func (suite *ComponentRepositoryTestSuite) TestGetByStatus() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create components with different statuses
	active1 := suite.factories.Component.WithName("active-component-1")
	active1.OrganizationID = org.ID
	active1.Status = models.ComponentStatusActive
	err = suite.repo.Create(active1)
	suite.NoError(err)

	active2 := suite.factories.Component.WithName("active-component-2")
	active2.OrganizationID = org.ID
	active2.Status = models.ComponentStatusActive
	err = suite.repo.Create(active2)
	suite.NoError(err)

	inactive1 := suite.factories.Component.WithName("inactive-component-1")
	inactive1.OrganizationID = org.ID
	inactive1.Status = models.ComponentStatusInactive
	err = suite.repo.Create(inactive1)
	suite.NoError(err)

	// Get components by active status
	activeComponents, total, err := suite.repo.GetByStatus(org.ID, models.ComponentStatusActive, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(activeComponents, 2)
	suite.Equal(int64(2), total)

	// Verify all returned components are active
	for _, component := range activeComponents {
		suite.Equal(models.ComponentStatusActive, component.Status)
	}
}

// TestGetActiveComponents tests retrieving active components
func (suite *ComponentRepositoryTestSuite) TestGetActiveComponents() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create components with different statuses
	active1 := suite.factories.Component.WithName("active-for-test-1")
	active1.OrganizationID = org.ID
	active1.Status = models.ComponentStatusActive
	err = suite.repo.Create(active1)
	suite.NoError(err)

	active2 := suite.factories.Component.WithName("active-for-test-2")
	active2.OrganizationID = org.ID
	active2.Status = models.ComponentStatusActive
	err = suite.repo.Create(active2)
	suite.NoError(err)

	inactive1 := suite.factories.Component.WithName("inactive-for-test-1")
	inactive1.OrganizationID = org.ID
	inactive1.Status = models.ComponentStatusInactive
	err = suite.repo.Create(inactive1)
	suite.NoError(err)

	// Get active components
	activeComponents, total, err := suite.repo.GetActiveComponents(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(activeComponents, 2)
	suite.Equal(int64(2), total)

	// Verify all returned components are active
	for _, component := range activeComponents {
		suite.Equal(models.ComponentStatusActive, component.Status)
	}
}

// Run the test suite
func TestComponentRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentRepositoryTestSuite))
}
