//go:build integration
// +build integration

package repository

import (
	"testing"

	"developer-portal-backend/internal/database/models"
	"developer-portal-backend/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// UserRepositoryTestSuite tests the UserRepository
type UserRepositoryTestSuite struct {
	suite.Suite
	baseTestSuite *testutils.BaseTestSuite
	repo          *UserRepository
	factories     *testutils.FactorySet
}

// SetupSuite runs before all tests in the suite
func (suite *UserRepositoryTestSuite) SetupSuite() {
	// Initialize shared BaseTestSuite using the new API
	suite.baseTestSuite = testutils.SetupTestSuite(suite.T())

	// Init repository and factories
	suite.repo = NewUserRepository(suite.baseTestSuite.DB)
	suite.factories = testutils.NewFactorySet()
}

// TearDownSuite runs after all tests in the suite
func (suite *UserRepositoryTestSuite) TearDownSuite() {
	suite.baseTestSuite.TeardownTestSuite()
}

// SetupTest runs before each test
func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.baseTestSuite.SetupTest()
}

// TearDownTest runs after each test
func (suite *UserRepositoryTestSuite) TearDownTest() {
	suite.baseTestSuite.TearDownTest()
}

// TestCreate tests creating a new member
func (suite *UserRepositoryTestSuite) TestCreate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test member
	member := suite.factories.User.Create()
	member.OrganizationID = org.ID

	// Create the member
	err = suite.repo.Create(member)

	// Assertions
	suite.NoError(err)
	suite.NotEqual(uuid.Nil, member.ID)
	suite.NotZero(member.CreatedAt)
	suite.NotZero(member.UpdatedAt)
}

// TestCreateDuplicateEmail tests creating a member with duplicate email
func (suite *UserRepositoryTestSuite) TestCreateDuplicateEmail() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create first member
	member1 := suite.factories.User.WithEmail("test@example.com")
	member1.OrganizationID = org.ID
	err = suite.repo.Create(member1)
	suite.NoError(err)

	// Try to create second member with same email
	member2 := suite.factories.User.WithEmail("test@example.com")
	member2.OrganizationID = org.ID
	member2.FullName = "Different Name" // Different name
	member2.FirstName = "Different"
	member2.LastName = "Name"

	err = suite.repo.Create(member2)
	suite.Error(err)
	suite.Contains(err.Error(), "duplicate key value")
}

// TestGetByID tests retrieving a member by ID
func (suite *UserRepositoryTestSuite) TestGetByID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test member
	member := suite.factories.User.Create()
	member.OrganizationID = org.ID
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Retrieve the member
	retrievedMember, err := suite.repo.GetByID(member.ID)

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedMember)
	suite.Equal(member.ID, retrievedUser.ID)
	suite.Equal(member.Email, retrievedUser.Email)
	suite.Equal(member.FullName, retrievedUser.FullName)
	suite.Equal(member.FirstName, retrievedUser.FirstName)
	suite.Equal(member.LastName, retrievedUser.LastName)
	suite.Equal(member.Role, retrievedUser.Role)
}

// TestGetByIDNotFound tests retrieving a non-existent member
func (suite *UserRepositoryTestSuite) TestGetByIDNotFound() {
	nonExistentID := uuid.New()

	member, err := suite.repo.GetByID(nonExistentID)

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(member)
}

// TestGetByEmail tests retrieving a member by email
func (suite *UserRepositoryTestSuite) TestGetByEmail() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test member
	member := suite.factories.User.WithEmail("test@example.com")
	member.OrganizationID = org.ID
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Retrieve the member by email
	retrievedMember, err := suite.repo.GetByEmail("test@example.com")

	// Assertions
	suite.NoError(err)
	suite.NotNil(retrievedMember)
	suite.Equal(member.ID, retrievedUser.ID)
	suite.Equal("test@example.com", retrievedUser.Email)
}

// TestGetByEmailNotFound tests retrieving a non-existent member by email
func (suite *UserRepositoryTestSuite) TestGetByEmailNotFound() {
	member, err := suite.repo.GetByEmail("nonexistent@example.com")

	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Nil(member)
}

// TestGetByOrganizationID tests listing members by organization
func (suite *UserRepositoryTestSuite) TestGetByOrganizationID() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test members
	member1 := suite.factories.User.WithEmail("member1@example.com")
	member1.OrganizationID = org.ID
	member1.FullName = "Member One"
	member1.FirstName = "Member"
	member1.LastName = "One"
	err = suite.repo.Create(member1)
	suite.NoError(err)

	member2 := suite.factories.User.WithEmail("member2@example.com")
	member2.OrganizationID = org.ID
	member2.FullName = "Member Two"
	member2.FirstName = "Member"
	member2.LastName = "Two"
	err = suite.repo.Create(member2)
	suite.NoError(err)

	member3 := suite.factories.User.WithEmail("member3@example.com")
	member3.OrganizationID = org.ID
	member3.FullName = "Member Three"
	member3.FirstName = "Member"
	member3.LastName = "Three"
	err = suite.repo.Create(member3)
	suite.NoError(err)

	// List members by organization
	members, total, err := suite.repo.GetByOrganizationID(org.ID, 10, 0)

	// Assertions
	suite.NoError(err)
	suite.Len(members, 3)
	suite.Equal(int64(3), total)

	// Verify members are returned
	emails := make([]string, len(members))
	for i, member := range members {
		emails[i] = member.Email
	}
	suite.Contains(emails, "member1@example.com")
	suite.Contains(emails, "member2@example.com")
	suite.Contains(emails, "member3@example.com")
}

// TestGetByOrganizationIDWithPagination tests listing members with pagination
func (suite *UserRepositoryTestSuite) TestGetByOrganizationIDWithPagination() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create multiple test members
	for i := 0; i < 5; i++ {
		member := suite.factories.User.WithEmail(suite.T().Name() + "-member" + uuid.New().String()[:8] + "@example.com")
		member.OrganizationID = org.ID
		uuidSuffix := uuid.New().String()[:6]
		member.FullName = "Test Member " + uuidSuffix
		member.FirstName = "Test"
		member.LastName = "Member " + uuidSuffix
		err := suite.repo.Create(member)
		suite.NoError(err)
	}

	// Test first page
	members, total, err := suite.repo.GetByOrganizationID(org.ID, 2, 0)
	suite.NoError(err)
	suite.Len(members, 2)
	suite.Equal(int64(5), total)

	// Test second page
	members, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 2)
	suite.NoError(err)
	suite.Len(members, 2)
	suite.Equal(int64(5), total)

	// Test third page
	members, total, err = suite.repo.GetByOrganizationID(org.ID, 2, 4)
	suite.NoError(err)
	suite.Len(members, 1) // Only one left
	suite.Equal(int64(5), total)
}

// TestUpdate tests updating a member
func (suite *UserRepositoryTestSuite) TestUpdate() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test member
	member := suite.factories.User.Create()
	member.OrganizationID = org.ID
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Update the member
	member.FullName = "Updated Name"
	member.FirstName = "Updated"
	member.LastName = "Name"
	member.Role = models.MemberRoleManager
	member.PhoneNumber = "+1-555-9999"

	err = suite.repo.Update(member)

	// Assertions
	suite.NoError(err)

	// Retrieve updated member
	updatedMember, err := suite.repo.GetByID(member.ID)
	suite.NoError(err)
	suite.Equal("Updated Name", updatedUser.FullName)
	suite.Equal("Updated", updatedUser.FirstName)
	suite.Equal("Name", updatedUser.LastName)
	suite.Equal(models.MemberRoleManager, updatedUser.Role)
	suite.Equal("+1-555-9999", updatedUser.PhoneNumber)
	suite.True(updatedUser.UpdatedAt.After(updatedUser.CreatedAt))
}

// TestDelete tests deleting a member
func (suite *UserRepositoryTestSuite) TestDelete() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create test member
	member := suite.factories.User.Create()
	member.OrganizationID = org.ID
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Delete the member
	err = suite.repo.Delete(member.ID)
	suite.NoError(err)

	// Verify member is deleted
	_, err = suite.repo.GetByID(member.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestDeleteNotFound tests deleting a non-existent member
func (suite *UserRepositoryTestSuite) TestDeleteNotFound() {
	nonExistentID := uuid.New()

	err := suite.repo.Delete(nonExistentID)

	// Should not error when deleting non-existent record
	suite.NoError(err)
}

// TestGetWithOrganization tests retrieving member with organization details
func (suite *UserRepositoryTestSuite) TestGetWithOrganization() {
	// Create organization first
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create member
	member := suite.factories.User.Create()
	member.OrganizationID = org.ID
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Retrieve member with organization details
	memberWithOrg, err := suite.repo.GetWithOrganization(member.ID)

	suite.NoError(err)
	suite.NotNil(memberWithOrg)
	suite.Equal(member.ID, memberWithOrg.ID)
	suite.NotNil(memberWithOrg.Organization)
	suite.Equal(org.ID, memberWithOrg.Organization.ID)
	suite.Equal(org.Name, memberWithOrg.Organization.Name)
}

// TestGetByTeamID tests retrieving members by team ID
func (suite *UserRepositoryTestSuite) TestGetByTeamID() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create team
	group := suite.factories.Group.WithOrganization(org.ID)
	groupRepo := NewGroupRepository(suite.baseTestSuite.DB)
	err = groupRepo.Create(group)
	suite.NoError(err)

	team := suite.factories.Team.WithGroup(group.ID)
	teamRepo := NewTeamRepository(suite.baseTestSuite.DB)
	err = teamRepo.Create(team)
	suite.NoError(err)

	// Create members with the team
	member1 := suite.factories.User.WithEmail("team1@example.com")
	member1.OrganizationID = org.ID
	member1.TeamID = &team.ID
	err = suite.repo.Create(member1)
	suite.NoError(err)

	member2 := suite.factories.User.WithEmail("team2@example.com")
	member2.OrganizationID = org.ID
	member2.TeamID = &team.ID
	err = suite.repo.Create(member2)
	suite.NoError(err)

	// Get members by team ID
	members, total, err := suite.repo.GetByTeamID(team.ID, 10, 0)

	suite.NoError(err)
	suite.Len(members, 2)
	suite.Equal(int64(2), total)
}

// TestGetByRole tests retrieving members by role
func (suite *UserRepositoryTestSuite) TestGetByRole() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create members with different roles
	manager1 := suite.factories.User.WithRole(models.MemberRoleManager)
	manager1.OrganizationID = org.ID
	manager1.Email = "manager1@example.com"
	err = suite.repo.Create(manager1)
	suite.NoError(err)

	manager2 := suite.factories.User.WithRole(models.MemberRoleManager)
	manager2.OrganizationID = org.ID
	manager2.Email = "manager2@example.com"
	err = suite.repo.Create(manager2)
	suite.NoError(err)

	developer := suite.factories.User.WithRole(models.MemberRoleDeveloper)
	developer.OrganizationID = org.ID
	developer.Email = "developer@example.com"
	err = suite.repo.Create(developer)
	suite.NoError(err)

	// Get managers only
	managers, total, err := suite.repo.GetByRole(org.ID, models.MemberRoleManager, 10, 0)

	suite.NoError(err)
	suite.Len(managers, 2)
	suite.Equal(int64(2), total)
	for _, m := range managers {
		suite.Equal(models.MemberRoleManager, m.Role)
	}
}

// TestGetActiveMembers tests retrieving active members
func (suite *UserRepositoryTestSuite) TestGetActiveMembers() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create active members
	active1 := suite.factories.User.WithEmail("active1@example.com")
	active1.OrganizationID = org.ID
	active1.IsActive = true
	err = suite.repo.Create(active1)
	suite.NoError(err)

	active2 := suite.factories.User.WithEmail("active2@example.com")
	active2.OrganizationID = org.ID
	active2.IsActive = true
	err = suite.repo.Create(active2)
	suite.NoError(err)

	// Create inactive member
	inactive := suite.factories.User.WithEmail("inactive@example.com")
	inactive.OrganizationID = org.ID
	inactive.IsActive = false
	err = suite.repo.Create(inactive)
	suite.NoError(err)

	// Get active members only
	activeMembers, total, err := suite.repo.GetActiveMembers(org.ID, 10, 0)

	suite.NoError(err)
	suite.Len(activeMembers, 2)
	suite.Equal(int64(2), total)
	for _, m := range activeMembers {
		suite.True(m.IsActive)
	}
}

// TestGetWithTeam tests retrieving member with team details
func (suite *UserRepositoryTestSuite) TestGetWithTeam() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create team
	group := suite.factories.Group.WithOrganization(org.ID)
	groupRepo := NewGroupRepository(suite.baseTestSuite.DB)
	err = groupRepo.Create(group)
	suite.NoError(err)

	team := suite.factories.Team.WithGroup(group.ID)
	teamRepo := NewTeamRepository(suite.baseTestSuite.DB)
	err = teamRepo.Create(team)
	suite.NoError(err)

	// Create member with team
	member := suite.factories.User.WithEmail("team-member@example.com")
	member.OrganizationID = org.ID
	member.TeamID = &team.ID
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Retrieve member with team details
	memberWithTeam, err := suite.repo.GetWithTeam(member.ID)

	suite.NoError(err)
	suite.NotNil(memberWithTeam)
	suite.Equal(member.ID, memberWithTeam.ID)
	suite.NotNil(memberWithTeam.Team)
	suite.Equal(team.ID, memberWithTeam.Team.ID)
	suite.Equal(team.Name, memberWithTeam.Team.Name)
}

// TestAssignToTeam tests assigning a member to a team
func (suite *UserRepositoryTestSuite) TestAssignToTeam() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create team
	group := suite.factories.Group.WithOrganization(org.ID)
	groupRepo := NewGroupRepository(suite.baseTestSuite.DB)
	err = groupRepo.Create(group)
	suite.NoError(err)

	team := suite.factories.Team.WithGroup(group.ID)
	teamRepo := NewTeamRepository(suite.baseTestSuite.DB)
	err = teamRepo.Create(team)
	suite.NoError(err)

	// Create member without team
	member := suite.factories.User.WithEmail("assign-test@example.com")
	member.OrganizationID = org.ID
	member.TeamID = nil
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Assign to team
	err = suite.repo.AssignToTeam(member.ID, team.ID)
	suite.NoError(err)

	// Verify assignment
	updatedMember, err := suite.repo.GetByID(member.ID)
	suite.NoError(err)
	suite.NotNil(updatedUser.TeamID)
	suite.Equal(team.ID, *updatedUser.TeamID)
}

// TestRemoveFromTeam tests removing a member from a team
func (suite *UserRepositoryTestSuite) TestRemoveFromTeam() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create team
	group := suite.factories.Group.WithOrganization(org.ID)
	groupRepo := NewGroupRepository(suite.baseTestSuite.DB)
	err = groupRepo.Create(group)
	suite.NoError(err)

	team := suite.factories.Team.WithGroup(group.ID)
	teamRepo := NewTeamRepository(suite.baseTestSuite.DB)
	err = teamRepo.Create(team)
	suite.NoError(err)

	// Create member with team
	member := suite.factories.User.WithEmail("remove-test@example.com")
	member.OrganizationID = org.ID
	member.TeamID = &team.ID
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Remove from team
	err = suite.repo.RemoveFromTeam(member.ID)
	suite.NoError(err)

	// Verify removal
	updatedMember, err := suite.repo.GetByID(member.ID)
	suite.NoError(err)
	suite.Nil(updatedUser.TeamID)
}

// TestUpdateRole tests updating a member's role
func (suite *UserRepositoryTestSuite) TestUpdateRole() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create member with developer role
	member := suite.factories.User.WithRole(models.MemberRoleDeveloper)
	member.OrganizationID = org.ID
	member.Email = "role-test@example.com"
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Update to manager role
	err = suite.repo.UpdateRole(member.ID, models.MemberRoleManager)
	suite.NoError(err)

	// Verify role update
	updatedMember, err := suite.repo.GetByID(member.ID)
	suite.NoError(err)
	suite.Equal(models.MemberRoleManager, updatedUser.Role)
}

// TestSetActiveStatus tests setting the active status of a member
func (suite *UserRepositoryTestSuite) TestSetActiveStatus() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create active member
	member := suite.factories.User.WithEmail("status-test@example.com")
	member.OrganizationID = org.ID
	member.IsActive = true
	err = suite.repo.Create(member)
	suite.NoError(err)

	// Set to inactive
	err = suite.repo.SetActiveStatus(member.ID, false)
	suite.NoError(err)

	// Verify status update
	updatedMember, err := suite.repo.GetByID(member.ID)
	suite.NoError(err)
	suite.False(updatedUser.IsActive)

	// Set back to active
	err = suite.repo.SetActiveStatus(member.ID, true)
	suite.NoError(err)

	// Verify status update
	updatedMember, err = suite.repo.GetByID(member.ID)
	suite.NoError(err)
	suite.True(updatedUser.IsActive)
}

// TestSearch tests searching members by name or email
func (suite *UserRepositoryTestSuite) TestSearch() {
	// Create organization
	org := suite.factories.Organization.Create()
	orgRepo := NewOrganizationRepository(suite.baseTestSuite.DB)
	err := orgRepo.Create(org)
	suite.NoError(err)

	// Create members with searchable names/emails
	alice := suite.factories.User.WithEmail("alice.smith@example.com")
	alice.OrganizationID = org.ID
	alice.FirstName = "Alice"
	alice.LastName = "Smith"
	alice.FullName = "Smith, Alice"
	err = suite.repo.Create(alice)
	suite.NoError(err)

	bob := suite.factories.User.WithEmail("bob.jones@example.com")
	bob.OrganizationID = org.ID
	bob.FirstName = "Bob"
	bob.LastName = "Jones"
	bob.FullName = "Jones, Bob"
	err = suite.repo.Create(bob)
	suite.NoError(err)

	charlie := suite.factories.User.WithEmail("charlie.brown@example.com")
	charlie.OrganizationID = org.ID
	charlie.FirstName = "Charlie"
	charlie.LastName = "Brown"
	charlie.FullName = "Brown, Charlie"
	err = suite.repo.Create(charlie)
	suite.NoError(err)

	// Search by partial name
	results, total, err := suite.repo.Search(org.ID, "alice", 10, 0)
	suite.NoError(err)
	suite.Len(results, 1)
	suite.Equal(int64(1), total)
	suite.Equal(alice.Email, results[0].Email)

	// Search by partial email
	results, total, err = suite.repo.Search(org.ID, "bob.jones", 10, 0)
	suite.NoError(err)
	suite.Len(results, 1)
	suite.Equal(int64(1), total)
	suite.Equal(bob.Email, results[0].Email)

	// Search that matches multiple
	results, total, err = suite.repo.Search(org.ID, "example.com", 10, 0)
	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 3)
	suite.GreaterOrEqual(total, int64(3))
}

// Run the test suite
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
