package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidInput     = errors.New("invalid input")
	ErrCannotDelete     = errors.New("cannot delete")
	ErrCascadeDelete    = errors.New("cascade delete failed")
	ErrReassignFailed   = errors.New("reassignment failed")
)

type NotFoundError struct {
	Resource string
	ID       string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

func (e NotFoundError) Unwrap() error {
	return ErrNotFound
}

type AlreadyExistsError struct {
	Resource   string
	Identifier string
}

func (e AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s already exists: %s", e.Resource, e.Identifier)
}

func (e AlreadyExistsError) Unwrap() error {
	return ErrAlreadyExists
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

func (e ValidationError) Unwrap() error {
	return ErrValidationFailed
}

type CascadeDeleteError struct {
	Count   int
	Message string
}

func (e CascadeDeleteError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("failed to cascade delete %d record(s)", e.Count)
}

func (e CascadeDeleteError) Unwrap() error {
	return ErrCascadeDelete
}

type ReassignError struct {
	From    string
	To      string
	Count   int
	Message string
}

func (e ReassignError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("failed to reassign %d transaction(s) from %s to %s", e.Count, e.From, e.To)
}

func (e ReassignError) Unwrap() error {
	return ErrReassignFailed
}

type CannotDeleteError struct {
	Resource string
	Reason   string
}

func (e CannotDeleteError) Error() string {
	return fmt.Sprintf("cannot delete %s: %s", e.Resource, e.Reason)
}

func (e CannotDeleteError) Unwrap() error {
	return ErrCannotDelete
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

func IsValidationFailed(err error) bool {
	return errors.Is(err, ErrValidationFailed)
}

func IsCannotDelete(err error) bool {
	return errors.Is(err, ErrCannotDelete)
}
