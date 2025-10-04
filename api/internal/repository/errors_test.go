package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryError_Error(t *testing.T) {
	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := RepositoryError{
			Type:    ErrTypeNotFound,
			Message: "resource not found",
			Cause:   cause,
		}

		expected := "NOT_FOUND: resource not found (caused by: underlying error)"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error without cause", func(t *testing.T) {
		err := RepositoryError{
			Type:    ErrTypeNotFound,
			Message: "resource not found",
		}

		expected := "NOT_FOUND: resource not found"
		assert.Equal(t, expected, err.Error())
	})
}

func TestRepositoryError_Unwrap(t *testing.T) {
	t.Run("should unwrap cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := RepositoryError{
			Type:    ErrTypeNotFound,
			Message: "resource not found",
			Cause:   cause,
		}

		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("should return nil when no cause", func(t *testing.T) {
		err := RepositoryError{
			Type:    ErrTypeNotFound,
			Message: "resource not found",
		}

		assert.Nil(t, err.Unwrap())
	})
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("test resource not found")

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeNotFound, repoErr.Type)
	assert.Equal(t, "test resource not found", repoErr.Message)
}

func TestErrAlreadyExists(t *testing.T) {
	err := ErrAlreadyExists("test resource already exists")

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeAlreadyExists, repoErr.Type)
	assert.Equal(t, "test resource already exists", repoErr.Message)
}

func TestIsNotFound(t *testing.T) {
	t.Run("should return true for not found error", func(t *testing.T) {
		err := ErrNotFound("resource not found")
		assert.True(t, IsNotFound(err))
	})

	t.Run("should return false for other error types", func(t *testing.T) {
		err := ErrAlreadyExists("resource exists")
		assert.False(t, IsNotFound(err))
	})

	t.Run("should return false for non-repository errors", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, IsNotFound(err))
	})
}

func TestIsAlreadyExists(t *testing.T) {
	t.Run("should return true for already exists error", func(t *testing.T) {
		err := ErrAlreadyExists("resource exists")
		assert.True(t, IsAlreadyExists(err))
	})

	t.Run("should return false for other error types", func(t *testing.T) {
		err := ErrNotFound("resource not found")
		assert.False(t, IsAlreadyExists(err))
	})
}

func TestIsConflict(t *testing.T) {
	t.Run("should return true for conflict error", func(t *testing.T) {
		err := ErrConflict("resource conflict")
		assert.True(t, IsConflict(err))
	})

	t.Run("should return false for other error types", func(t *testing.T) {
		err := ErrNotFound("resource not found")
		assert.False(t, IsConflict(err))
	})
}

func TestErrConnection(t *testing.T) {
	cause := errors.New("network timeout")
	err := ErrConnection("connection failed", cause)

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeConnection, repoErr.Type)
	assert.Equal(t, "connection failed", repoErr.Message)
	assert.Equal(t, cause, repoErr.Cause)
}

func TestErrInternal(t *testing.T) {
	cause := errors.New("internal failure")
	err := ErrInternal("internal error", cause)

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeInternal, repoErr.Type)
	assert.Equal(t, "internal error", repoErr.Message)
	assert.Equal(t, cause, repoErr.Cause)
}
