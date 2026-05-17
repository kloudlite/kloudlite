package errors

import (
	"errors"
)

func Wrap(msg string, err ...error) error {
	return errors.Join(errors.New(msg), errors.Join(err...))
}

func New(msg string) error {
	return errors.New(msg)
}
