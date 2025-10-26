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

// MemberServiceTestSuite defines the test suite for MemberService
type MemberServiceTestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mockMemberRepo *mocks.MockMemberRepositoryInterface
	memberService  *service.MemberService
	validator      *validator.Validate
}

// SetupTest sets up the test suite
func (suite *MemberServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockMemberRepo = mocks.NewMockMemberRepositoryInterface(suite.ctrl)
	suite.validator = validator.New()

	// Create service with mock repository
	suite.memberService = service.NewMemberService(suite.mockMemberRepo, suite.validator)
}

// TearDownTest cleans up after each test
func (suite *MemberServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// TestCreateMember tests creating a member
func (suite *MemberServiceTestSuite) TestCreateMember() {
	role := "developer"
	req := &service.CreateMemberRequest{
		OrganizationID: uuid.New(),
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		PhoneNumber:    "+1-555-0123",
		IUser:          "I123456",
		Role:           &role,
	}

	// Mock GetByEmail to return not found (no existing member with same email)
	suite.mockMemberRepo.EXPECT().
		GetByEmail(req.Email).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	// Mock Create to succeed
	suite.mockMemberRepo.EXPECT().
		Create(gomock.Any()).
		Return(nil).
		Times(1)

	response, err := suite.memberService.CreateMember(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), req.FullName, response.FullName)
	assert.Equal(suite.T(), req.FirstName, response.FirstName)
	assert.Equal(suite.T(), req.LastName, response.LastName)
	assert.Equal(suite.T(), req.Email, response.Email)
	assert.Equal(suite.T(), *req.Role, response.Role)
}

// TestCreateMemberWithDefaultRoleAndTeamRole tests creating a member with default role and team role
func (suite *MemberServiceTestSuite) TestCreateMemberWithDefaultRoleAndTeamRole() {
	req := &service.CreateMemberRequest{
		OrganizationID: uuid.New(),
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		PhoneNumber:    "+1-555-0123",
		IUser:          "I123456",
		// Role and TeamRole are not provided - should use defaults
	}

	// Mock GetByEmail to return not found (no existing member with same email)
	suite.mockMemberRepo.EXPECT().
		GetByEmail(req.Email).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	// Mock Create to succeed
	suite.mockMemberRepo.EXPECT().
		Create(gomock.Any()).
		Return(nil).
		Times(1)

	response, err := suite.memberService.CreateMember(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), req.FullName, response.FullName)
	assert.Equal(suite.T(), "developer", response.Role)  // Default role
	assert.Equal(suite.T(), "member", response.TeamRole) // Default team role
}

// TestCreateMemberValidationError tests creating a member with validation error
func (suite *MemberServiceTestSuite) TestCreateMemberValidationError() {
	role := "developer"
	req := &service.CreateMemberRequest{
		OrganizationID: uuid.New(),
		FullName:       "", // Empty full name should fail validation
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		IUser:          "I123456",
		Role:           &role,
	}

	response, err := suite.memberService.CreateMember(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "validation failed")
}

// TestCreateMemberDuplicateEmail tests creating a member with duplicate email
func (suite *MemberServiceTestSuite) TestCreateMemberDuplicateEmail() {
	orgID := uuid.New()
	role := "developer"
	req := &service.CreateMemberRequest{
		OrganizationID: orgID,
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		IUser:          "I123456",
		Role:           &role,
	}

	existingMember := &models.Member{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		OrganizationID: orgID,
		FullName:       "Jane Doe",
		FirstName:      "Jane",
		LastName:       "Doe",
		Email:          req.Email,
		IUser:          "I789012",
		Role:           models.MemberRoleDeveloper,
	}

	// Mock GetByEmail to return existing member
	suite.mockMemberRepo.EXPECT().
		GetByEmail(req.Email).
		Return(existingMember, nil).
		Times(1)

	response, err := suite.memberService.CreateMember(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "member already exists with this email")
}

// TestGetMemberByID tests getting a member by ID
func (suite *MemberServiceTestSuite) TestGetMemberByID() {
	memberID := uuid.New()
	expectedMember := &models.Member{
		BaseModel: models.BaseModel{
			ID: memberID,
		},
		OrganizationID: uuid.New(),
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		IUser:          "I123456",
		Role:           models.MemberRoleDeveloper,
		IsActive:       true,
	}

	suite.mockMemberRepo.EXPECT().
		GetByID(memberID).
		Return(expectedMember, nil).
		Times(1)

	response, err := suite.memberService.GetMemberByID(memberID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), expectedMember.ID, response.ID)
	assert.Equal(suite.T(), expectedMember.FullName, response.FullName)
	assert.Equal(suite.T(), expectedMember.FirstName, response.FirstName)
	assert.Equal(suite.T(), expectedMember.LastName, response.LastName)
	assert.Equal(suite.T(), expectedMember.Email, response.Email)
}

// TestGetMemberByIDNotFound tests getting a member by ID when not found
func (suite *MemberServiceTestSuite) TestGetMemberByIDNotFound() {
	memberID := uuid.New()

	suite.mockMemberRepo.EXPECT().
		GetByID(memberID).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	response, err := suite.memberService.GetMemberByID(memberID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "member not found")
}

// TestGetMembersByOrganization tests getting members by organization
func (suite *MemberServiceTestSuite) TestGetMembersByOrganization() {
	orgID := uuid.New()
	limit, offset := 20, 0
	expectedMembers := []models.Member{
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			OrganizationID: orgID,
			FullName:       "John Doe",
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john@example.com",
			IUser:          "I123456",
			Role:           models.MemberRoleDeveloper,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			OrganizationID: orgID,
			FullName:       "Jane Smith",
			FirstName:      "Jane",
			LastName:       "Smith",
			Email:          "jane@example.com",
			IUser:          "I789012",
			Role:           models.MemberRoleManager,
		},
	}
	expectedTotal := int64(2)

	suite.mockMemberRepo.EXPECT().
		GetByOrganizationID(orgID, limit, offset).
		Return(expectedMembers, expectedTotal, nil).
		Times(1)

	responses, total, err := suite.memberService.GetMembersByOrganization(orgID, limit, offset)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTotal, total)
	assert.Len(suite.T(), responses, 2)
	assert.Equal(suite.T(), expectedMembers[0].FullName, responses[0].FullName)
	assert.Equal(suite.T(), expectedMembers[0].FirstName, responses[0].FirstName)
	assert.Equal(suite.T(), expectedMembers[0].LastName, responses[0].LastName)
	assert.Equal(suite.T(), expectedMembers[1].FullName, responses[1].FullName)
	assert.Equal(suite.T(), expectedMembers[1].FirstName, responses[1].FirstName)
	assert.Equal(suite.T(), expectedMembers[1].LastName, responses[1].LastName)
}

// TestUpdateMember tests updating a member
func (suite *MemberServiceTestSuite) TestUpdateMember() {
	memberID := uuid.New()
	existingMember := &models.Member{
		BaseModel: models.BaseModel{
			ID: memberID,
		},
		OrganizationID: uuid.New(),
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		IUser:          "I123456",
		Role:           models.MemberRoleDeveloper,
		IsActive:       true,
	}

	newFullName := "John Updated"
	newFirstName := "John"
	newLastName := "Updated"
	newEmail := "john.updated@example.com"
	req := &service.UpdateMemberRequest{
		FullName:  &newFullName,
		FirstName: &newFirstName,
		LastName:  &newLastName,
		Email:     &newEmail,
	}

	suite.mockMemberRepo.EXPECT().
		GetByID(memberID).
		Return(existingMember, nil).
		Times(1)

	suite.mockMemberRepo.EXPECT().
		GetByEmail(newEmail).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	suite.mockMemberRepo.EXPECT().
		Update(gomock.Any()).
		Return(nil).
		Times(1)

	response, err := suite.memberService.UpdateMember(memberID, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), newFullName, response.FullName)
	assert.Equal(suite.T(), newFirstName, response.FirstName)
	assert.Equal(suite.T(), newLastName, response.LastName)
	assert.Equal(suite.T(), newEmail, response.Email)
}

// TestDeleteMember tests deleting a member
func (suite *MemberServiceTestSuite) TestDeleteMember() {
	memberID := uuid.New()
	existingMember := &models.Member{
		BaseModel: models.BaseModel{
			ID: memberID,
		},
		OrganizationID: uuid.New(),
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		IUser:          "I123456",
		Role:           models.MemberRoleDeveloper,
	}

	suite.mockMemberRepo.EXPECT().
		GetByID(memberID).
		Return(existingMember, nil).
		Times(1)

	suite.mockMemberRepo.EXPECT().
		Delete(memberID).
		Return(nil).
		Times(1)

	err := suite.memberService.DeleteMember(memberID)

	assert.NoError(suite.T(), err)
}

// TestSearchMembers tests searching for members
func (suite *MemberServiceTestSuite) TestSearchMembers() {
	orgID := uuid.New()
	query := "john"
	limit, offset := 20, 0
	expectedMembers := []models.Member{
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			OrganizationID: orgID,
			FullName:       "Doe, John",
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john.doe@example.com",
			IUser:          "I123456",
			Role:           models.MemberRoleDeveloper,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			OrganizationID: orgID,
			FullName:       "Johnson, Mary",
			FirstName:      "Mary",
			LastName:       "Johnson",
			Email:          "mary.johnson@example.com",
			IUser:          "I789012",
			Role:           models.MemberRoleManager,
		},
	}
	expectedTotal := int64(2)

	suite.mockMemberRepo.EXPECT().
		SearchByOrganization(orgID, query, limit, offset).
		Return(expectedMembers, expectedTotal, nil).
		Times(1)

	responses, total, err := suite.memberService.SearchMembers(orgID, query, limit, offset)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTotal, total)
	assert.Len(suite.T(), responses, 2)
	assert.Equal(suite.T(), expectedMembers[0].FullName, responses[0].FullName)
	assert.Equal(suite.T(), expectedMembers[0].Email, responses[0].Email)
	assert.Equal(suite.T(), expectedMembers[1].FullName, responses[1].FullName)
	assert.Equal(suite.T(), expectedMembers[1].Email, responses[1].Email)
}

// TestSearchMembersError tests searching for members with error
func (suite *MemberServiceTestSuite) TestSearchMembersError() {
	orgID := uuid.New()
	query := "test"
	limit, offset := 20, 0

	suite.mockMemberRepo.EXPECT().
		SearchByOrganization(orgID, query, limit, offset).
		Return(nil, int64(0), gorm.ErrInvalidDB).
		Times(1)

	responses, total, err := suite.memberService.SearchMembers(orgID, query, limit, offset)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), responses)
	assert.Equal(suite.T(), int64(0), total)
	assert.Contains(suite.T(), err.Error(), "failed to search members")
}

// TestGetActiveMembers tests getting active members
func (suite *MemberServiceTestSuite) TestGetActiveMembers() {
	orgID := uuid.New()
	limit, offset := 20, 0
	expectedMembers := []models.Member{
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			OrganizationID: orgID,
			FullName:       "Smith, Active",
			FirstName:      "Active",
			LastName:       "Smith",
			Email:          "active.smith@example.com",
			IUser:          "I123456",
			Role:           models.MemberRoleDeveloper,
			IsActive:       true,
		},
		{
			BaseModel: models.BaseModel{
				ID: uuid.New(),
			},
			OrganizationID: orgID,
			FullName:       "Jones, Active",
			FirstName:      "Active",
			LastName:       "Jones",
			Email:          "active.jones@example.com",
			IUser:          "I789012",
			Role:           models.MemberRoleManager,
			IsActive:       true,
		},
	}
	expectedTotal := int64(2)

	suite.mockMemberRepo.EXPECT().
		GetActiveByOrganization(orgID, limit, offset).
		Return(expectedMembers, expectedTotal, nil).
		Times(1)

	responses, total, err := suite.memberService.GetActiveMembers(orgID, limit, offset)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedTotal, total)
	assert.Len(suite.T(), responses, 2)
	assert.True(suite.T(), responses[0].IsActive)
	assert.True(suite.T(), responses[1].IsActive)
	assert.Equal(suite.T(), expectedMembers[0].Email, responses[0].Email)
	assert.Equal(suite.T(), expectedMembers[1].Email, responses[1].Email)
}

// TestGetActiveMembersError tests getting active members with error
func (suite *MemberServiceTestSuite) TestGetActiveMembersError() {
	orgID := uuid.New()
	limit, offset := 20, 0

	suite.mockMemberRepo.EXPECT().
		GetActiveByOrganization(orgID, limit, offset).
		Return(nil, int64(0), gorm.ErrInvalidDB).
		Times(1)

	responses, total, err := suite.memberService.GetActiveMembers(orgID, limit, offset)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), responses)
	assert.Equal(suite.T(), int64(0), total)
	assert.Contains(suite.T(), err.Error(), "failed to get active members")
}

// TestUpdateMemberNotFound tests updating a member that doesn't exist
func (suite *MemberServiceTestSuite) TestUpdateMemberNotFound() {
	memberID := uuid.New()
	newFullName := "John Updated"
	req := &service.UpdateMemberRequest{
		FullName: &newFullName,
	}

	suite.mockMemberRepo.EXPECT().
		GetByID(memberID).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	response, err := suite.memberService.UpdateMember(memberID, req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "member not found")
}

// TestUpdateMemberEmailConflict tests updating a member with a conflicting email
func (suite *MemberServiceTestSuite) TestUpdateMemberEmailConflict() {
	memberID := uuid.New()
	existingMember := &models.Member{
		BaseModel: models.BaseModel{
			ID: memberID,
		},
		OrganizationID: uuid.New(),
		FullName:       "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		IUser:          "I123456",
		Role:           models.MemberRoleDeveloper,
		IsActive:       true,
	}

	conflictingEmail := "taken@example.com"
	conflictingMember := &models.Member{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		OrganizationID: existingMember.OrganizationID,
		Email:          conflictingEmail,
	}

	req := &service.UpdateMemberRequest{
		Email: &conflictingEmail,
	}

	suite.mockMemberRepo.EXPECT().
		GetByID(memberID).
		Return(existingMember, nil).
		Times(1)

	suite.mockMemberRepo.EXPECT().
		GetByEmail(conflictingEmail).
		Return(conflictingMember, nil).
		Times(1)

	response, err := suite.memberService.UpdateMember(memberID, req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), response)
	assert.Contains(suite.T(), err.Error(), "member already exists")
}

// TestDeleteMemberNotFound tests deleting a member that doesn't exist
func (suite *MemberServiceTestSuite) TestDeleteMemberNotFound() {
	memberID := uuid.New()

	suite.mockMemberRepo.EXPECT().
		GetByID(memberID).
		Return(nil, gorm.ErrRecordNotFound).
		Times(1)

	err := suite.memberService.DeleteMember(memberID)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "member not found")
}

// TestMemberServiceTestSuite runs the test suite
// TestAddQuickLinkValidation tests the validation logic for adding a quick link
func TestAddQuickLinkValidation(t *testing.T) {
	validator := validator.New()

	testCases := []struct {
		name        string
		request     *service.AddQuickLinkRequest
		expectError bool
	}{
		{
			name: "Valid quick link",
			request: &service.AddQuickLinkRequest{
				URL:      "https://github.com/user/repo",
				Title:    "My Repository",
				Icon:     "github",
				Category: "repository",
			},
			expectError: false,
		},
		{
			name: "Valid quick link without optional fields",
			request: &service.AddQuickLinkRequest{
				URL:   "https://example.com",
				Title: "Example",
			},
			expectError: false,
		},
		{
			name: "Missing URL",
			request: &service.AddQuickLinkRequest{
				Title:    "My Repository",
				Icon:     "github",
				Category: "repository",
			},
			expectError: true,
		},
		{
			name: "Invalid URL",
			request: &service.AddQuickLinkRequest{
				URL:   "not-a-url",
				Title: "My Repository",
			},
			expectError: true,
		},
		{
			name: "Missing title",
			request: &service.AddQuickLinkRequest{
				URL:      "https://github.com/user/repo",
				Icon:     "github",
				Category: "repository",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.Struct(tc.request)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMemberServiceTestSuite(t *testing.T) {
	suite.Run(t, new(MemberServiceTestSuite))
}
