package errors

import (
	"fmt"

	"github.com/yext/yerrors"
)

func NewEf(err error, msg string, a ...interface{}) error {
	return yerrors.WrapFrame(yerrors.Errorf("%s while %s", fmt.Sprintf(msg, a...), err.Error()), 1)
}

func Newf(msg string, a ...interface{}) error {
	if len(a) > 0 {
		return yerrors.Wrap(yerrors.Errorf(msg, a...))
	}
	return yerrors.New(msg)
}

func NewE(err error) error {
	return yerrors.Wrap(err)
}

func New(msg string) error {
	return yerrors.Wrap(yerrors.New(msg))
}
