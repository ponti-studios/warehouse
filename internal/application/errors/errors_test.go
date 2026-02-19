package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundError(t *testing.T) {
	err := NotFoundError{Resource: "Account", ID: "123"}
	assert.Equal(t, "Account not found: 123", err.Error())
	assert.True(t, errors.Is(err, ErrNotFound))
}

func TestAlreadyExistsError(t *testing.T) {
	err := AlreadyExistsError{Resource: "Account", Identifier: "test-account"}
	assert.Equal(t, "Account already exists: test-account", err.Error())
	assert.True(t, errors.Is(err, ErrAlreadyExists))
}

func TestValidationError(t *testing.T) {
	err := ValidationError{Field: "name", Message: "is required"}
	assert.Equal(t, "name: is required", err.Error())
	assert.True(t, errors.Is(err, ErrValidationFailed))
}

func TestValidationErrorNoField(t *testing.T) {
	err := ValidationError{Message: "validation failed"}
	assert.Equal(t, "validation failed", err.Error())
}

func TestCascadeDeleteError(t *testing.T) {
	err := CascadeDeleteError{Count: 5}
	assert.Equal(t, "failed to cascade delete 5 record(s)", err.Error())
	assert.True(t, errors.Is(err, ErrCascadeDelete))
}

func TestCascadeDeleteErrorWithMessage(t *testing.T) {
	err := CascadeDeleteError{Count: 5, Message: "database error"}
	assert.Equal(t, "database error", err.Error())
}

func TestReassignError(t *testing.T) {
	err := ReassignError{From: "Food", To: "Uncategorized", Count: 10}
	assert.Equal(t, "failed to reassign 10 transaction(s) from Food to Uncategorized", err.Error())
	assert.True(t, errors.Is(err, ErrReassignFailed))
}

func TestCannotDeleteError(t *testing.T) {
	err := CannotDeleteError{Resource: "Uncategorized", Reason: "cannot delete default category"}
	assert.Equal(t, "cannot delete Uncategorized: cannot delete default category", err.Error())
	assert.True(t, errors.Is(err, ErrCannotDelete))
}

func TestIsNotFound(t *testing.T) {
	assert.True(t, IsNotFound(NotFoundError{Resource: "Account", ID: "1"}))
	assert.True(t, IsNotFound(ErrNotFound))
	assert.False(t, IsNotFound(AlreadyExistsError{Resource: "Account", Identifier: "test"}))
}

func TestIsAlreadyExists(t *testing.T) {
	assert.True(t, IsAlreadyExists(AlreadyExistsError{Resource: "Account", Identifier: "test"}))
	assert.True(t, IsAlreadyExists(ErrAlreadyExists))
	assert.False(t, IsAlreadyExists(NotFoundError{Resource: "Account", ID: "1"}))
}

func TestIsValidationFailed(t *testing.T) {
	assert.True(t, IsValidationFailed(ValidationError{Field: "name", Message: "is required"}))
	assert.True(t, IsValidationFailed(ErrValidationFailed))
	assert.False(t, IsValidationFailed(NotFoundError{Resource: "Account", ID: "1"}))
}

func TestIsCannotDelete(t *testing.T) {
	assert.True(t, IsCannotDelete(CannotDeleteError{Resource: "Category", Reason: "has children"}))
	assert.True(t, IsCannotDelete(ErrCannotDelete))
	assert.False(t, IsCannotDelete(NotFoundError{Resource: "Account", ID: "1"}))
}
