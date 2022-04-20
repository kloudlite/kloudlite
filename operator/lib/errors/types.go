package errors

import (
	"fmt"

	"github.com/yext/yerrors"
)

func NewEf(err error, msg string, a ...interface{}) error {
	return yerrors.Errorf("%s as %v", fmt.Sprintf(msg, a...), err)
}

func Newf(msg string, a ...interface{}) error {
	if len(a) > 0 {
		return yerrors.Wrap(yerrors.Errorf(msg, a))
	}
	return yerrors.New(msg)
}

func NewE(err error) error {
	return yerrors.Wrap(err)
}

func New(msg string) error {
	return yerrors.Wrap(yerrors.New(msg))
}

func StatusUpdate(err error) error {
	return NewEf(err, "resource status update failed")
}

func ConditionUpdate(err error) error {
	return Newf("job condition update failed as %v", err)
}
