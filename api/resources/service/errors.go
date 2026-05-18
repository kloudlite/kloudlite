package service

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorKind string

const (
	ErrUnknownResource ErrorKind = "unknown_resource"
	ErrWrongScope      ErrorKind = "wrong_scope"
	ErrNotFound        ErrorKind = "not_found"
	ErrCacheNotReady   ErrorKind = "cache_not_ready"
	ErrBadRequest      ErrorKind = "bad_request"
	ErrConflict        ErrorKind = "conflict"
	ErrNotImplemented  ErrorKind = "not_implemented"
)

type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.Err }

func NewError(kind ErrorKind, message string, err error) error {
	return &Error{Kind: kind, Message: message, Err: err}
}

func IsKind(err error, kind ErrorKind) bool {
	var serviceErr *Error
	return errors.As(err, &serviceErr) && serviceErr.Kind == kind
}

func UnknownResource(alias string) error {
	return NewError(ErrUnknownResource, fmt.Sprintf("unknown resource %q", alias), nil)
}

func HTTPStatus(err error) int {
	var serviceErr *Error
	if !errors.As(err, &serviceErr) {
		return http.StatusInternalServerError
	}

	switch serviceErr.Kind {
	case ErrUnknownResource, ErrNotFound:
		return http.StatusNotFound
	case ErrWrongScope, ErrBadRequest:
		return http.StatusBadRequest
	case ErrConflict:
		return http.StatusConflict
	case ErrCacheNotReady:
		return http.StatusServiceUnavailable
	case ErrNotImplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}
