package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ComponentDeploymentRepositoryTestSuite tests the ComponentDeploymentRepository
type ComponentDeploymentRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *ComponentDeploymentRepository
	factories     *testutils.FactorySet
}

// SetupSuite runs before all tests in the suite
func (suite *ComponentDeploymentRepositoryTestSuite) SetupSuite() {
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	suite.repo = NewComponentDeploymentRepository(suite.baseTestSuite.DB)
	suite.factories = testutils.NewFactorySet()
}

// TearDownSuite runs after all tests in the suite
func (suite *ComponentDeploymentRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *ComponentDeploymentRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *ComponentDeploymentRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// TestCreate tests creating a new component deployment
func (suite *ComponentDeploymentRepositoryTestSuite) TestCreate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
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

	// Create test component deployment
	deployment := suite.factories.ComponentDeployment.Create()
	deployment.ComponentID = component.ID
	deployment.LandscapeID = landscape.ID

	// Create the component deployment
	err = suite.repo.Create(deployment)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, deployment.ID)
	suite.NotZero(deployment.CreatedAt)
	suite.NotZero(deployment.UpdatedAt)
}

// TestCreateDuplicateComponentLandscape tests creating a deployment with duplicate component-landscape pair
func (suite *ComponentDeploymentRepositoryTestSuite) TestCreateDuplicateComponentLandscape() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
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

	// Create first deployment
	deployment1 := suite.factories.ComponentDeployment.Create()
	deployment1.ComponentID = component.ID
	deployment1.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment1)
	suite.NoError(err)

	// Try to create second deployment with same component-landscape pair
	deployment2 := suite.factories.ComponentDeployment.Create()
	deployment2.ComponentID = component.ID
	deployment2.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment2)

	// Should succeed as there may not be a unique constraint on (component_id, landscape_id)
	// This allows multiple deployments of the same component in the same landscape
	suite.NoError(err)
}

// TestGetByID tests retrieving a component deployment by ID
func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
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

	// Create test component deployment
	deployment := suite.factories.ComponentDeployment.Create()
	deployment.ComponentID = component.ID
	deployment.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment)
	suite.NoError(err)

	// Retrieve the deployment
	retrievedDeployment, err := suite.repo.GetByID(deployment.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedDeployment)
	suite.Equal(deployment.ID, retrievedDeployment.ID)
	suite.Equal(deployment.ComponentID, retrievedDeployment.ComponentID)
	suite.Equal(deployment.LandscapeID, retrievedDeployment.LandscapeID)
	suite.Equal(deployment.IsActive, retrievedDeployment.IsActive)
}

// TestGetByIDNotFound tests retrieving a non-existent component deployment
func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	deployment, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(deployment)
}

// TestGetByComponentID tests listing deployments by component
func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByComponentID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create component
	component := suite.factories.Component.Create()
	component.OrganizationID = org.ID
	componentRepo := NewComponentRepository(suite.baseTestSuite.DB)
	err = componentRepo.Create(component)
	suite.NoError(err)

	// Create multiple landscapes
	landscape1 := suite.factories.Landscape.WithName("landscape-1")
	landscape1.OrganizationID = org.ID
	landscapeRepo := NewLandscapeRepository(suite.baseTestSuite.DB)
	err = landscapeRepo.Create(landscape1)
	suite.NoError(err)

	landscape2 := suite.factories.Landscape.WithName("landscape-2")
	landscape2.OrganizationID = org.ID
	err = landscapeRepo.Create(landscape2)
	suite.NoError(err)

	landscape3 := suite.factories.Landscape.WithName("landscape-3")
	landscape3.OrganizationID = org.ID
	err = landscapeRepo.Create(landscape3)
	suite.NoError(err)

	// Create deployments for the component
	deployment1 := suite.factories.ComponentDeployment.Create()
	deployment1.ComponentID = component.ID
	deployment1.LandscapeID = landscape1.ID
	err = suite.repo.Create(deployment1)
	suite.NoError(err)

	deployment2 := suite.factories.ComponentDeployment.Create()
	deployment2.ComponentID = component.ID
	deployment2.LandscapeID = landscape2.ID
	err = suite.repo.Create(deployment2)
	suite.NoError(err)

	deployment3 := suite.factories.ComponentDeployment.Create()
	deployment3.ComponentID = component.ID
	deployment3.LandscapeID = landscape3.ID
	err = suite.repo.Create(deployment3)
	suite.NoError(err)

	// List deployments by component
	deployments, total, err := suite.repo.GetByComponentID(component.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(deployments, 3)
	suite.Equal(int64(3), total)

	// Verify all deployments belong to the component
	for _, deployment := range deployments {
		suite.Equal(component.ID, deployment.ComponentID)
	}
}

// TestGetByLandscapeID tests listing deployments by landscape
func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByLandscapeID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create landscape
	landscape := suite.factories.Landscape.Create()
	landscape.OrganizationID = org.ID
	landscapeRepo := NewLandscapeRepository(suite.baseTestSuite.DB)
	err = landscapeRepo.Create(landscape)
	suite.NoError(err)

	// Create multiple components
	component1 := suite.factories.Component.WithName("component-1")
	component1.OrganizationID = org.ID
	componentRepo := NewComponentRepository(suite.baseTestSuite.DB)
	err = componentRepo.Create(component1)
	suite.NoError(err)

	component2 := suite.factories.Component.WithName("component-2")
	component2.OrganizationID = org.ID
	err = componentRepo.Create(component2)
	suite.NoError(err)

	component3 := suite.factories.Component.WithName("component-3")
	component3.OrganizationID = org.ID
	err = componentRepo.Create(component3)
	suite.NoError(err)

	// Create deployments for the landscape
	deployment1 := suite.factories.ComponentDeployment.Create()
	deployment1.ComponentID = component1.ID
	deployment1.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment1)
	suite.NoError(err)

	deployment2 := suite.factories.ComponentDeployment.Create()
	deployment2.ComponentID = component2.ID
	deployment2.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment2)
	suite.NoError(err)

	deployment3 := suite.factories.ComponentDeployment.Create()
	deployment3.ComponentID = component3.ID
	deployment3.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment3)
	suite.NoError(err)

	// List deployments by landscape
	deployments, total, err := suite.repo.GetByLandscapeID(landscape.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(deployments, 3)
	suite.Equal(int64(3), total)

	// Verify all deployments belong to the landscape
	for _, deployment := range deployments {
		suite.Equal(landscape.ID, deployment.LandscapeID)
	}
}

// TestGetByComponentAndLandscape tests retrieving deployment by component and landscape
func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByComponentAndLandscape() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
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

	// Create test component deployment
	deployment := suite.factories.ComponentDeployment.Create()
	deployment.ComponentID = component.ID
	deployment.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment)
	suite.NoError(err)

	// Retrieve the deployment by component and landscape
	retrievedDeployment, err := suite.repo.GetByComponentAndLandscape(component.ID, landscape.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedDeployment)
	suite.Equal(deployment.ID, retrievedDeployment.ID)
	suite.Equal(component.ID, retrievedDeployment.ComponentID)
	suite.Equal(landscape.ID, retrievedDeployment.LandscapeID)
}

// TestGetByComponentAndLandscapeNotFound tests retrieving non-existent deployment
func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByComponentAndLandscapeNotFound() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
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

	// Try to retrieve non-existent deployment
	deployment, err := suite.repo.GetByComponentAndLandscape(component.ID, landscape.ID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(deployment)
}

// TestUpdate tests updating a component deployment
func (suite *ComponentDeploymentRepositoryTestSuite) TestUpdate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
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

	// Create test component deployment
	deployment := suite.factories.ComponentDeployment.Create()
	deployment.ComponentID = component.ID
	deployment.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment)
	suite.NoError(err)

	// Update the deployment
	deployment.IsActive = false
	deployment.Version = "v2.0.0"
	deployment.GitCommitID = "abc123def456"

	err = suite.repo.Update(deployment)

	// Assertions
	suite.NoError(err)

	// Retrieve updated deployment
	updatedDeployment, err := suite.repo.GetByID(deployment.ID)
	suite.NoError(err)
	suite.Equal(false, updatedDeployment.IsActive)
	suite.Equal("v2.0.0", updatedDeployment.Version)
	suite.Equal("abc123def456", updatedDeployment.GitCommitID)
	suite.True(updatedDeployment.UpdatedAt.After(updatedDeployment.CreatedAt))
}

// TestDelete tests deleting a component deployment
func (suite *ComponentDeploymentRepositoryTestSuite) TestDelete() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
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

	// Create test component deployment
	deployment := suite.factories.ComponentDeployment.Create()
	deployment.ComponentID = component.ID
	deployment.LandscapeID = landscape.ID
	err = suite.repo.Create(deployment)
	suite.NoError(err)

	// Delete the deployment
	err = suite.repo.Delete(deployment.ID)
	suite.NoError(err)

	// Verify deployment is deleted
	_, err = suite.repo.GetByID(deployment.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestDeleteNotFound tests deleting a non-existent component deployment
func (suite *ComponentDeploymentRepositoryTestSuite) TestDeleteNotFound() {
	nonExistentID := uuid.New()

	err := suite.repo.Delete(nonExistentID)

	// Should not error when deleting non-existent record
	suite.NoError(err)
}

// // TestGetByActiveStatus tests retrieving deployments by active status
// func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByActiveStatus() {
// 	// Create organization first
// 	org := suite.factories.Organization.Create()
// 	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
// 	err := orgRepo.Create(org)
// 	suite.NoError(err)

// 	// Create a dedicated component for this test to avoid interference
// 	component := suite.factories.Component.WithName("active-status-test-component")
// 	component.OrganizationID = org.ID
// 	componentRepo := NewComponentRepository(suite.baseTestSuite.DB)
// 	err = componentRepo.Create(component)
// 	suite.NoError(err)

// 	// Create landscapes
// 	landscape1 := suite.factories.Landscape.WithName("active-test-landscape-1")
// 	landscape1.OrganizationID = org.ID
// 	landscapeRepo := NewLandscapeRepository(suite.baseTestSuite.DB)
// 	err = landscapeRepo.Create(landscape1)
// 	suite.NoError(err)

// 	landscape2 := suite.factories.Landscape.WithName("active-test-landscape-2")
// 	landscape2.OrganizationID = org.ID
// 	err = landscapeRepo.Create(landscape2)
// 	suite.NoError(err)

// 	landscape3 := suite.factories.Landscape.WithName("active-test-landscape-3")
// 	landscape3.OrganizationID = org.ID
// 	err = landscapeRepo.Create(landscape3)
// 	suite.NoError(err)

// 	// Create deployments with different active statuses
// 	active1 := suite.factories.ComponentDeployment.Create()
// 	active1.ComponentID = component.ID
// 	active1.LandscapeID = landscape1.ID
// 	active1.IsActive = true
// 	err = suite.repo.Create(active1)
// 	suite.NoError(err)

// 	active2 := suite.factories.ComponentDeployment.Create()
// 	active2.ComponentID = component.ID
// 	active2.LandscapeID = landscape2.ID
// 	active2.IsActive = true
// 	err = suite.repo.Create(active2)
// 	suite.NoError(err)

// 	inactive1 := suite.factories.ComponentDeployment.Create()
// 	inactive1.ComponentID = component.ID
// 	inactive1.LandscapeID = landscape3.ID
// 	inactive1.IsActive = false
// 	err = suite.repo.Create(inactive1)
// 	suite.NoError(err)

// 	// Test GetByActiveStatus method for active deployments
// 	allActiveDeployments, _, err := suite.repo.GetByActiveStatus(true, 100, 0)
// 	suite.NoError(err)

// 	// Verify all returned deployments are active
// 	for _, deployment := range allActiveDeployments {
// 		suite.Equal(true, deployment.IsActive)
// 	}

// 	// Count our component's active deployments in the results
// 	componentActiveCount := 0
// 	for _, deployment := range allActiveDeployments {
// 		if deployment.ComponentID == component.ID && deployment.IsActive {
// 			componentActiveCount++
// 		}
// 	}

// 	// Verify our component has exactly 2 active deployments in the global results
// 	suite.Equal(2, componentActiveCount)

// 	// Test GetByActiveStatus method for inactive deployments
// 	allInactiveDeployments, _, err := suite.repo.GetByActiveStatus(false, 100, 0)
// 	suite.NoError(err)

// 	// Verify all returned deployments are inactive
// 	for _, deployment := range allInactiveDeployments {
// 		suite.Equal(false, deployment.IsActive)
// 	}

// 	// Count our component's inactive deployments in the results
// 	componentInactiveCount := 0
// 	for _, deployment := range allInactiveDeployments {
// 		if deployment.ComponentID == component.ID && deployment.IsActive == false {
// 			componentInactiveCount++
// 		}
// 	}

// 	// Verify our component has exactly 1 inactive deployment in the global results
// 	suite.Equal(1, componentInactiveCount)
// }

// TestGetByActiveStatus tests retrieving deployments by active status
func (suite *ComponentDeploymentRepositoryTestSuite) TestGetByActiveStatus() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create a dedicated component for this test to avoid interference
	component := suite.factories.Component.WithName("active-status-test-component")
	component.OrganizationID = org.ID
	componentRepo := NewComponentRepository(suite.baseTestSuite.DB)
	err = componentRepo.Create(component)
	suite.NoError(err)

	// Create landscapes
	landscape1 := suite.factories.Landscape.WithName("active-test-landscape-1")
	landscape1.OrganizationID = org.ID
	landscapeRepo := NewLandscapeRepository(suite.baseTestSuite.DB)
	err = landscapeRepo.Create(landscape1)
	suite.NoError(err)

	landscape2 := suite.factories.Landscape.WithName("active-test-landscape-2")
	landscape2.OrganizationID = org.ID
	err = landscapeRepo.Create(landscape2)
	suite.NoError(err)

	landscape3 := suite.factories.Landscape.WithName("active-test-landscape-3")
	landscape3.OrganizationID = org.ID
	err = landscapeRepo.Create(landscape3)
	suite.NoError(err)

	// ---- FIX #1: build (do not persist via factory) and call repo.Create exactly once ----
	active1 := &models.ComponentDeployment{
		ComponentID: component.ID,
		LandscapeID: landscape1.ID,
		IsActive:    true,
	}
	suite.NoError(suite.repo.Create(active1))

	active2 := &models.ComponentDeployment{
		ComponentID: component.ID,
		LandscapeID: landscape2.ID,
		IsActive:    true,
	}
	suite.NoError(suite.repo.Create(active2))

	inactive1 := &models.ComponentDeployment{
		ComponentID: component.ID,
		LandscapeID: landscape3.ID,
		IsActive:    false,
	}
	suite.NoError(suite.repo.Create(inactive1))

	// Test GetByActiveStatus method for active deployments
	allActiveDeployments, _, err := suite.repo.GetByActiveStatus(true, 100, 0)
	suite.NoError(err)

	// ---- FIX #4: clearer assertions + verify our specific IDs appear ----
	for _, deployment := range allActiveDeployments {
		suite.True(deployment.IsActive, "expected all returned deployments to be active")
	}

	// Verify our component has exactly 2 active deployments in the global results
	componentActiveIDs := make(map[uuid.UUID]bool)
	for _, d := range allActiveDeployments {
		if d.ComponentID == component.ID && d.IsActive {
			componentActiveIDs[d.ID] = true
		}
	}
	suite.Equal(2, len(componentActiveIDs), "expected exactly 2 active deployments for our component")
	suite.Contains(componentActiveIDs, active1.ID, "expected active1 to be returned")
	suite.Contains(componentActiveIDs, active2.ID, "expected active2 to be returned")

	// Test GetByActiveStatus method for inactive deployments
	allInactiveDeployments, _, err := suite.repo.GetByActiveStatus(false, 100, 0)
	suite.NoError(err)

	for _, deployment := range allInactiveDeployments {
		suite.False(deployment.IsActive, "expected all returned deployments to be inactive")
	}

	// Verify our component has exactly 1 inactive deployment in the global results
	componentInactiveIDs := make(map[uuid.UUID]bool)
	for _, d := range allInactiveDeployments {
		if d.ComponentID == component.ID && !d.IsActive {
			componentInactiveIDs[d.ID] = true
		}
	}
	suite.Equal(1, len(componentInactiveIDs), "expected exactly 1 inactive deployment for our component")
	suite.Contains(componentInactiveIDs, inactive1.ID, "expected inactive1 to be returned")
}

// Run the test suite
func TestComponentDeploymentRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentDeploymentRepositoryTestSuite))
}
