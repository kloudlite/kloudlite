package errors

import (
	"github.com/pkg/errors"
)

type ErrNotFound struct {
	Message string
	error
}

func (err ErrNotFound) Error() string {
	if err.Message != "" {
		return err.Message
	}
	return "not found"
}

func OfType[T error](err error) bool {
	var er T
	return errors.As(err, &er)
}
