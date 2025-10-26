package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	t.Run("Error message", func(t *testing.T) {
		err := &NotFoundError{Entity: "team"}
		assert.Equal(t, "team not found", err.Error())
	})

	t.Run("errors.Is comparison with same entity", func(t *testing.T) {
		err1 := &NotFoundError{Entity: "team"}
		err2 := &NotFoundError{Entity: "team"}
		assert.True(t, errors.Is(err1, err2))
	})

	t.Run("errors.Is comparison with different entity", func(t *testing.T) {
		err1 := &NotFoundError{Entity: "team"}
		err2 := &NotFoundError{Entity: "component"}
		assert.False(t, errors.Is(err1, err2))
	})

	t.Run("errors.Is with predefined errors", func(t *testing.T) {
		assert.True(t, errors.Is(ErrTeamNotFound, ErrTeamNotFound))
		assert.False(t, errors.Is(ErrTeamNotFound, ErrComponentNotFound))
	})

	t.Run("IsNotFound helper", func(t *testing.T) {
		assert.True(t, IsNotFound(ErrTeamNotFound))
		assert.False(t, IsNotFound(ErrComponentAlreadyAssociated))
	})
}

func TestAlreadyExistsError(t *testing.T) {
	t.Run("Error message with context", func(t *testing.T) {
		err := &AlreadyExistsError{Entity: "team", Context: "in the organization"}
		assert.Equal(t, "team already exists in the organization", err.Error())
	})

	t.Run("Error message without context", func(t *testing.T) {
		err := &AlreadyExistsError{Entity: "team"}
		assert.Equal(t, "team already exists", err.Error())
	})

	t.Run("errors.Is comparison", func(t *testing.T) {
		err1 := &AlreadyExistsError{Entity: "team", Context: "in org"}
		err2 := &AlreadyExistsError{Entity: "team", Context: "in org"}
		assert.True(t, errors.Is(err1, err2))
	})

	t.Run("IsAlreadyExists helper", func(t *testing.T) {
		assert.True(t, IsAlreadyExists(ErrTeamExists))
		assert.False(t, IsAlreadyExists(ErrTeamNotFound))
	})
}

func TestValidationError(t *testing.T) {
	t.Run("Error message with field", func(t *testing.T) {
		err := &ValidationError{Field: "email", Message: "invalid format"}
		assert.Equal(t, "validation error: email - invalid format", err.Error())
	})

	t.Run("Error message without field", func(t *testing.T) {
		err := &ValidationError{Message: "invalid format"}
		assert.Equal(t, "validation error: invalid format", err.Error())
	})

	t.Run("IsValidation helper", func(t *testing.T) {
		err := NewValidationError("email", "invalid")
		assert.True(t, IsValidation(err))
		assert.False(t, IsValidation(ErrTeamNotFound))
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("NewNotFoundError", func(t *testing.T) {
		err := NewNotFoundError("custom entity")
		assert.Equal(t, "custom entity not found", err.Error())
		assert.True(t, IsNotFound(err))
	})

	t.Run("NewAlreadyExistsError", func(t *testing.T) {
		err := NewAlreadyExistsError("custom", "in scope")
		assert.Equal(t, "custom already exists in scope", err.Error())
		assert.True(t, IsAlreadyExists(err))
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("field", "message")
		assert.Equal(t, "validation error: field - message", err.Error())
		assert.True(t, IsValidation(err))
	})
}

func TestBusinessLogicErrors(t *testing.T) {
	t.Run("Association errors", func(t *testing.T) {
		assert.Error(t, ErrComponentAlreadyAssociated)
		assert.Error(t, ErrComponentNotAssociated)
		assert.Error(t, ErrLandscapeAlreadyAssociated)
		assert.Error(t, ErrLandscapeNotAssociated)
	})

	t.Run("Business logic errors", func(t *testing.T) {
		assert.Error(t, ErrInvalidStatus)
		assert.Error(t, ErrCallTimeInFuture)
		assert.Error(t, ErrInvalidTimeRange)
		assert.Error(t, ErrDeploymentDateInPast)
		assert.Error(t, ErrTimelineCodeExists)
		assert.Error(t, ErrScheduleConflict)
		assert.Error(t, ErrInvalidDutyRotation)
		assert.Error(t, ErrNoMembersInTeam)
	})
}
