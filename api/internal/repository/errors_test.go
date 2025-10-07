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

func TestErrInvalidInput(t *testing.T) {
	err := ErrInvalidInput("invalid request parameters")

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeInvalidInput, repoErr.Type)
	assert.Equal(t, "invalid request parameters", repoErr.Message)
	assert.Nil(t, repoErr.Cause)
}

func TestErrMultipleFound(t *testing.T) {
	err := ErrMultipleFound("found 3 resources matching criteria")

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeMultipleFound, repoErr.Type)
	assert.Equal(t, "found 3 resources matching criteria", repoErr.Message)
	assert.Nil(t, repoErr.Cause)
}

func TestErrTimeout(t *testing.T) {
	cause := errors.New("context deadline exceeded")
	err := ErrTimeout("operation timed out", cause)

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeTimeout, repoErr.Type)
	assert.Equal(t, "operation timed out", repoErr.Message)
	assert.Equal(t, cause, repoErr.Cause)
}

func TestErrUnauthorized(t *testing.T) {
	err := ErrUnauthorized("user not authenticated")

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeUnauthorized, repoErr.Type)
	assert.Equal(t, "user not authenticated", repoErr.Message)
	assert.Nil(t, repoErr.Cause)
}

func TestErrForbidden(t *testing.T) {
	err := ErrForbidden("user lacks required permissions")

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeForbidden, repoErr.Type)
	assert.Equal(t, "user lacks required permissions", repoErr.Message)
	assert.Nil(t, repoErr.Cause)
}

func TestErrConflict(t *testing.T) {
	err := ErrConflict("resource version mismatch")

	repoErr, ok := err.(RepositoryError)
	assert.True(t, ok)
	assert.Equal(t, ErrTypeConflict, repoErr.Type)
	assert.Equal(t, "resource version mismatch", repoErr.Message)
	assert.Nil(t, repoErr.Cause)
}

func TestIsAlreadyExists_EdgeCases(t *testing.T) {
	t.Run("should return false for nil error", func(t *testing.T) {
		assert.False(t, IsAlreadyExists(nil))
	})

	t.Run("should return false for standard errors", func(t *testing.T) {
		err := errors.New("some standard error")
		assert.False(t, IsAlreadyExists(err))
	})

	t.Run("should return false for different repository error types", func(t *testing.T) {
		err := ErrInvalidInput("bad input")
		assert.False(t, IsAlreadyExists(err))
	})
}

func TestIsConflict_EdgeCases(t *testing.T) {
	t.Run("should return false for nil error", func(t *testing.T) {
		assert.False(t, IsConflict(nil))
	})

	t.Run("should return false for standard errors", func(t *testing.T) {
		err := errors.New("some standard error")
		assert.False(t, IsConflict(err))
	})

	t.Run("should return false for different repository error types", func(t *testing.T) {
		err := ErrTimeout("timeout", nil)
		assert.False(t, IsConflict(err))
	})
}
