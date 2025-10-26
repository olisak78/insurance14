package errors

import (
	"errors"
	"fmt"
)

// NotFoundError represents an error when an entity is not found
type NotFoundError struct {
	Entity string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Entity)
}

// Is enables errors.Is() comparison for NotFoundError
func (e *NotFoundError) Is(target error) bool {
	t, ok := target.(*NotFoundError)
	if !ok {
		return false
	}
	return e.Entity == t.Entity
}

// AlreadyExistsError represents an error when an entity already exists
type AlreadyExistsError struct {
	Entity  string
	Context string // Additional context like "in organization"
}

func (e *AlreadyExistsError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s already exists %s", e.Entity, e.Context)
	}
	return fmt.Sprintf("%s already exists", e.Entity)
}

// Is enables errors.Is() comparison for AlreadyExistsError
func (e *AlreadyExistsError) Is(target error) bool {
	t, ok := target.(*AlreadyExistsError)
	if !ok {
		return false
	}
	return e.Entity == t.Entity
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// AuthenticationError represents authentication-related errors
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// AuthorizationError represents authorization-related errors
type AuthorizationError struct {
	Message string
}

func (e *AuthorizationError) Error() string {
	return e.Message
}

// ConfigurationError represents configuration-related errors
type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string {
	return e.Message
}

// Entity Not Found Errors
var (
	ErrOrganizationNotFound           = &NotFoundError{Entity: "organization"}
	ErrTeamNotFound                   = &NotFoundError{Entity: "team"}
	ErrComponentNotFound              = &NotFoundError{Entity: "component"}
	ErrMemberNotFound                 = &NotFoundError{Entity: "member"}
	ErrProjectNotFound                = &NotFoundError{Entity: "project"}
	ErrLandscapeNotFound              = &NotFoundError{Entity: "landscape"}
	ErrGroupNotFound                  = &NotFoundError{Entity: "group"}
	ErrComponentDeploymentNotFound    = &NotFoundError{Entity: "component deployment"}
	ErrOutageCallNotFound             = &NotFoundError{Entity: "outage call"}
	ErrDeploymentTimelineNotFound     = &NotFoundError{Entity: "deployment timeline entry"}
	ErrDutyScheduleNotFound           = &NotFoundError{Entity: "duty schedule"}
	ErrLeaderNotFound                 = &NotFoundError{Entity: "leader"}
	ErrLinkNotFound                   = &NotFoundError{Entity: "link"}
	ErrTeamComponentOwnershipNotFound = &NotFoundError{Entity: "team-component ownership"}
	ErrProjectComponentNotFound       = &NotFoundError{Entity: "project-component relationship"}
	ErrProjectLandscapeNotFound       = &NotFoundError{Entity: "project-landscape relationship"}
	ErrOutageCallAssigneeNotFound     = &NotFoundError{Entity: "outage call assignee"}
)

// Already Exists Errors
var (
	ErrOrganizationExists              = &AlreadyExistsError{Entity: "organization", Context: "with this name or domain"}
	ErrTeamExists                      = &AlreadyExistsError{Entity: "team", Context: "with this name in the group"}
	ErrComponentExists                 = &AlreadyExistsError{Entity: "component", Context: "with this name in the organization"}
	ErrMemberExists                    = &AlreadyExistsError{Entity: "member", Context: "with this email"}
	ErrProjectExists                   = &AlreadyExistsError{Entity: "project", Context: "with this name in the organization"}
	ErrLandscapeExists                 = &AlreadyExistsError{Entity: "landscape", Context: "with this name"}
	ErrGroupExists                     = &AlreadyExistsError{Entity: "group", Context: "with this name in the organization"}
	ErrLinkExists                      = &AlreadyExistsError{Entity: "link", Context: "with this URL"}
	ErrComponentDeploymentExists       = &AlreadyExistsError{Entity: "component deployment", Context: "for this component and landscape"}
	ErrActiveComponentDeploymentExists = &AlreadyExistsError{Entity: "active component deployment", Context: "for this component and landscape"}
	ErrTeamComponentOwnershipExists    = &AlreadyExistsError{Entity: "team-component ownership", Context: ""}
	ErrProjectComponentExists          = &AlreadyExistsError{Entity: "project-component relationship", Context: ""}
	ErrProjectLandscapeExists          = &AlreadyExistsError{Entity: "project-landscape relationship", Context: ""}
	ErrOutageCallAssigneeExists        = &AlreadyExistsError{Entity: "outage call assignee", Context: ""}
)

// Association Errors
var (
	ErrComponentAlreadyAssociated = errors.New("component is already associated with this project")
	ErrComponentNotAssociated     = errors.New("component is not associated with this project")
	ErrLandscapeAlreadyAssociated = errors.New("landscape is already associated with this project")
	ErrLandscapeNotAssociated     = errors.New("landscape is not associated with this project")
	ErrMemberAlreadyAssigned      = errors.New("member is already assigned to this outage call")
	ErrMemberNotAssigned          = errors.New("member is not assigned to this outage call")
	ErrActiveDeploymentNotFound   = errors.New("active deployment not found")
)

// Business Logic Errors
var (
	ErrInvalidStatus              = errors.New("invalid status")
	ErrCallTimeInFuture           = errors.New("call time cannot be in the future")
	ErrInvalidTimeRange           = errors.New("invalid time range")
	ErrDeploymentDateInPast       = errors.New("scheduled deployment date is in the past")
	ErrTimelineCodeExists         = errors.New("timeline code already exists")
	ErrScheduleConflict           = errors.New("schedule conflict detected")
	ErrInvalidDutyRotation        = errors.New("invalid duty rotation configuration")
	ErrNoMembersInTeam            = errors.New("team has no members")
	ErrInvalidPaginationParams    = errors.New("invalid pagination parameters")
	ErrGitHubAPIRateLimitExceeded = errors.New("GitHub API rate limit exceeded")
)

// Authentication Errors
var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	
	// AI Core specific authentication errors
	ErrUserEmailNotFound     = &AuthenticationError{Message: "user email not found in context"}
	ErrUserNotAssignedToTeam = &AuthorizationError{Message: "user is not assigned to any team"}
	ErrUserNotFoundInDB      = &AuthorizationError{Message: "user not found in database"}
	ErrTeamNotFoundInDB      = &AuthorizationError{Message: "team not found in database"}
)

// Configuration Errors
var (
	ErrJiraConfigMissing    = errors.New("jira configuration missing: JIRA_DOMAIN, JIRA_USER or JIRA_PASSWORD")
	ErrJenkinsTokenNotFound = errors.New("jenkins token not found")
	ErrJenkinsUserNotFound  = errors.New("jenkins username not found")
	
	// AI Core specific configuration errors
	ErrAICoreCredentialsNotSet      = &ConfigurationError{Message: "AI_CORE_CREDENTIALS environment variable not set"}
	ErrAICoreCredentialsInvalid     = &ConfigurationError{Message: "failed to parse AI_CORE_CREDENTIALS"}
	ErrAICoreCredentialsNotFound    = &ConfigurationError{Message: "no credentials found for team"}
	ErrAICoreAPIRequestFailed       = errors.New("AI Core API request failed")
	ErrAICoreDeploymentNotFound     = &NotFoundError{Entity: "deployment"}
)

// Helper Functions

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	var notFoundErr *NotFoundError
	return errors.Is(err, &NotFoundError{}) || errors.As(err, &notFoundErr)
}

// IsAlreadyExists checks if an error is an AlreadyExistsError
func IsAlreadyExists(err error) bool {
	var existsErr *AlreadyExistsError
	return errors.Is(err, &AlreadyExistsError{}) || errors.As(err, &existsErr)
}

// IsValidation checks if an error is a ValidationError
func IsValidation(err error) bool {
	var validationErr *ValidationError
	return errors.Is(err, &ValidationError{}) || errors.As(err, &validationErr)
}

// IsAuthentication checks if an error is an AuthenticationError
func IsAuthentication(err error) bool {
	var authErr *AuthenticationError
	return errors.Is(err, &AuthenticationError{}) || errors.As(err, &authErr)
}

// IsAuthorization checks if an error is an AuthorizationError
func IsAuthorization(err error) bool {
	var authzErr *AuthorizationError
	return errors.Is(err, &AuthorizationError{}) || errors.As(err, &authzErr)
}

// IsConfiguration checks if an error is a ConfigurationError
func IsConfiguration(err error) bool {
	var configErr *ConfigurationError
	return errors.Is(err, &ConfigurationError{}) || errors.As(err, &configErr)
}

// NewNotFoundError creates a new NotFoundError for a custom entity
func NewNotFoundError(entity string) error {
	return &NotFoundError{Entity: entity}
}

// NewAlreadyExistsError creates a new AlreadyExistsError for a custom entity
func NewAlreadyExistsError(entity, context string) error {
	return &AlreadyExistsError{Entity: entity, Context: context}
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

// NewAuthenticationError creates a new AuthenticationError
func NewAuthenticationError(message string) error {
	return &AuthenticationError{Message: message}
}

// NewAuthorizationError creates a new AuthorizationError
func NewAuthorizationError(message string) error {
	return &AuthorizationError{Message: message}
}

// NewConfigurationError creates a new ConfigurationError
func NewConfigurationError(message string) error {
	return &ConfigurationError{Message: message}
}

// NewAICoreCredentialsNotFoundError creates a specific error for missing team credentials
func NewAICoreCredentialsNotFoundError(teamName string) error {
	return &ConfigurationError{Message: fmt.Sprintf("no credentials found for team: %s", teamName)}
}
