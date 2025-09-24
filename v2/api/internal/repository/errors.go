package repository

import "fmt"

// RepositoryError represents a repository-specific error
type RepositoryError struct {
	Type    string
	Message string
	Cause   error
}

func (e RepositoryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e RepositoryError) Unwrap() error {
	return e.Cause
}

// Error types
const (
	ErrTypeNotFound       = "NOT_FOUND"
	ErrTypeAlreadyExists  = "ALREADY_EXISTS"
	ErrTypeInvalidInput   = "INVALID_INPUT"
	ErrTypeConflict       = "CONFLICT"
	ErrTypeMultipleFound  = "MULTIPLE_FOUND"
	ErrTypeConnection     = "CONNECTION_ERROR"
	ErrTypeTimeout        = "TIMEOUT"
	ErrTypeUnauthorized   = "UNAUTHORIZED"
	ErrTypeForbidden      = "FORBIDDEN"
	ErrTypeInternal       = "INTERNAL_ERROR"
)

// ErrNotFound creates a not found error
func ErrNotFound(message string) error {
	return RepositoryError{
		Type:    ErrTypeNotFound,
		Message: message,
	}
}

// ErrAlreadyExists creates an already exists error
func ErrAlreadyExists(message string) error {
	return RepositoryError{
		Type:    ErrTypeAlreadyExists,
		Message: message,
	}
}

// ErrInvalidInput creates an invalid input error
func ErrInvalidInput(message string) error {
	return RepositoryError{
		Type:    ErrTypeInvalidInput,
		Message: message,
	}
}

// ErrConflict creates a conflict error
func ErrConflict(message string) error {
	return RepositoryError{
		Type:    ErrTypeConflict,
		Message: message,
	}
}

// ErrMultipleFound creates a multiple found error
func ErrMultipleFound(message string) error {
	return RepositoryError{
		Type:    ErrTypeMultipleFound,
		Message: message,
	}
}

// ErrConnection creates a connection error
func ErrConnection(message string, cause error) error {
	return RepositoryError{
		Type:    ErrTypeConnection,
		Message: message,
		Cause:   cause,
	}
}

// ErrTimeout creates a timeout error
func ErrTimeout(message string, cause error) error {
	return RepositoryError{
		Type:    ErrTypeTimeout,
		Message: message,
		Cause:   cause,
	}
}

// ErrUnauthorized creates an unauthorized error
func ErrUnauthorized(message string) error {
	return RepositoryError{
		Type:    ErrTypeUnauthorized,
		Message: message,
	}
}

// ErrForbidden creates a forbidden error
func ErrForbidden(message string) error {
	return RepositoryError{
		Type:    ErrTypeForbidden,
		Message: message,
	}
}

// ErrInternal creates an internal error
func ErrInternal(message string, cause error) error {
	return RepositoryError{
		Type:    ErrTypeInternal,
		Message: message,
		Cause:   cause,
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	if repoErr, ok := err.(RepositoryError); ok {
		return repoErr.Type == ErrTypeNotFound
	}
	return false
}

// IsAlreadyExists checks if an error is an already exists error
func IsAlreadyExists(err error) bool {
	if repoErr, ok := err.(RepositoryError); ok {
		return repoErr.Type == ErrTypeAlreadyExists
	}
	return false
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	if repoErr, ok := err.(RepositoryError); ok {
		return repoErr.Type == ErrTypeConflict
	}
	return false
}