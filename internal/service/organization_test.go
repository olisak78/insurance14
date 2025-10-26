package service_test

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/mocks"
	"developer-portal-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

// OrganizationServiceTestSuite defines the test suite for OrganizationService
type OrganizationServiceTestSuite struct {
	suite.Suite
	ctrl                *gomock.Controller
	mockOrgRepo         *mocks.MockOrganizationRepositoryInterface
	organizationService *service.OrganizationService
	validator           *validator.Validate
}

// SetupTest sets up the test suite
func (suite *OrganizationServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockOrgRepo = mocks.NewMockOrganizationRepositoryInterface(suite.ctrl)
	suite.validator = validator.New()

	// Create a service with the mock repository
	suite.organizationService = service.NewOrganizationService(suite.mockOrgRepo, suite.validator)
}

// TearDownTest cleans up after each test
func (suite *OrganizationServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateOrganization tests creating an organization
func (suite *OrganizationServiceTestSuite) TestCreateOrganization() {
	req := &service.CreateOrganizationRequest{
		Name:        "test-org",
		DisplayName: "Test Organization",
		Description: "A test organization",
		Domain:      "test.com",
	}

	// Mock GetByName to return not found (no existing org with same name)
	suite.mockOrgRepo.EXPECT().
		GetByName(req.Name).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	// Mock GetByDomain to return not found (no existing org with same domain)
	suite.mockOrgRepo.EXPECT().
		GetByDomain(req.Domain).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	// Mock Create to succeed
	suite.mockOrgRepo.EXPECT().
		Create(gomock.Any()).
		Return(nil).
		Times(1)

	response, err := suite.organizationService.Create(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), req.Name, response.Name)
	assert.Equal(suite.T(), req.DisplayName, response.DisplayName)
	assert.Equal(suite.T(), req.Domain, response.Domain)
}

// TestCreateOrganizationValidationError tests creating an organization with validation error
func (suite *OrganizationServiceTestSuite) TestCreateOrganizationValidationError() {
	req := &service.CreateOrganizationRequest{
		Name:        "", // Empty name should fail validation
		DisplayName: "Test Organization",
		Domain:      "test.com",
	}

	response, err := suite.organizationService.Create(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "validation failed")
}

// TestCreateOrganizationDuplicateName tests creating an organization with duplicate name
func (suite *OrganizationServiceTestSuite) TestCreateOrganizationDuplicateName() {
	req := &service.CreateOrganizationRequest{
		Name:        "test-org",
		DisplayName: "Test Organization",
		Description: "A test organization",
		Domain:      "test.com",
	}

	existingOrg := &models.Organization{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		Name:        req.Name,
		DisplayName: "Existing Organization",
		Domain:      "existing.com",
	}

	// Mock GetByName to return existing organization
	suite.mockOrgRepo.EXPECT().
		GetByName(req.Name).
		Return(existingOrg, nil).
		Times(1)

	response, err := suite.organizationService.Create(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "organization already exists with this name or domain")
}

// TestCreateOrganizationDuplicateDomain tests creating an organization with duplicate domain
func (suite *OrganizationServiceTestSuite) TestCreateOrganizationDuplicateDomain() {
	req := &service.CreateOrganizationRequest{
		Name:        "test-org",
		DisplayName: "Test Organization",
		Description: "A test organization",
		Domain:      "test.com",
	}

	existingOrg := &models.Organization{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		Name:        "different-org",
		DisplayName: "Different Organization",
		Domain:      req.Domain,
	}

	// Mock GetByName to return not found
	suite.mockOrgRepo.EXPECT().
		GetByName(req.Name).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	// Mock GetByDomain to return existing organization
	suite.mockOrgRepo.EXPECT().
		GetByDomain(req.Domain).
		Return(existingOrg, nil).
		Times(1)

	response, err := suite.organizationService.Create(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "organization already exists with this name or domain")
}

// TestGetOrganizationByID tests getting an organization by ID
func (suite *OrganizationServiceTestSuite) TestGetOrganizationByID() {
	orgID := uuid.New()
	expectedOrg := &models.Organization{
		BaseModel: models.BaseModel{
			ID: orgID,
		},
		Name:        "test-org",
		DisplayName: "Test Organization",
		Description: "A test organization",
		Domain:      "test.com",
	}

	suite.mockOrgRepo.EXPECT().
		GetByID(orgID).
		Return(expectedOrg, nil).
		Times(1)

	response, err := suite.organizationService.GetByID(orgID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), expectedOrg.ID, response.ID)
	assert.Equal(suite.T(), expectedOrg.Name, response.Name)
	assert.Equal(suite.T(), expectedOrg.DisplayName, response.DisplayName)
}

// TestGetOrganizationByIDNotFound tests getting an organization by ID when not found
func (suite *OrganizationServiceTestSuite) TestGetOrganizationByIDNotFound() {
	orgID := uuid.New()

	suite.mockOrgRepo.EXPECT().
		GetByID(orgID).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	response, err := suite.organizationService.GetByID(orgID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "organization not found")
}

// TestGetOrganizationByName tests getting an organization by name
func (suite *OrganizationServiceTestSuite) TestGetOrganizationByName() {
	orgName := "test-org"
	expectedOrg := &models.Organization{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		Name:        orgName,
		DisplayName: "Test Organization",
		Description: "A test organization",
		Domain:      "test.com",
	}

	suite.mockOrgRepo.EXPECT().
		GetByName(orgName).
		Return(expectedOrg, nil).
		Times(1)

	response, err := suite.organizationService.GetByName(orgName)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), expectedOrg.Name, response.Name)
	assert.Equal(suite.T(), expectedOrg.DisplayName, response.DisplayName)
}

// TestUpdateOrganization tests updating an organization
func (suite *OrganizationServiceTestSuite) TestUpdateOrganization() {
	orgID := uuid.New()
	existingOrg := &models.Organization{
		BaseModel: models.BaseModel{
			ID: orgID,
		},
		Name:        "test-org",
		DisplayName: "Test Organization",
		Description: "A test organization",
		Domain:      "test.com",
	}

	req := &service.UpdateOrganizationRequest{
		DisplayName: "Updated Organization",
		Description: "An updated organization",
	}

	suite.mockOrgRepo.EXPECT().
		GetByID(orgID).
		Return(existingOrg, nil).
		Times(1)

	suite.mockOrgRepo.EXPECT().
		Update(gomock.Any()).
		Return(nil).
		Times(1)

	response, err := suite.organizationService.Update(orgID, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), req.DisplayName, response.DisplayName)
	assert.Equal(suite.T(), req.Description, response.Description)
}

// TestDeleteOrganization tests deleting an organization
func (suite *OrganizationServiceTestSuite) TestDeleteOrganization() {
	orgID := uuid.New()
	existingOrg := &models.Organization{
		BaseModel: models.BaseModel{
			ID: orgID,
		},
		Name:        "test-org",
		DisplayName: "Test Organization",
		Domain:      "test.com",
	}

	suite.mockOrgRepo.EXPECT().
		GetByID(orgID).
		Return(existingOrg, nil).
		Times(1)

	suite.mockOrgRepo.EXPECT().
		Delete(orgID).
		Return(nil).
		Times(1)

	err := suite.organizationService.Delete(orgID)

	assert.NoError(suite.T(), err)
}

// TestOrganizationServiceTestSuite runs the test suite
func TestOrganizationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationServiceTestSuite))
}
