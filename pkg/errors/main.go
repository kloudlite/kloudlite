package errors

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/yext/yerrors"
)

func HandleErr(e *error) {
	if r := recover(); r != nil {
		if y, ok := r.(error); ok {
			*e = y
		}
		return
	}
}

func Assert(condition bool, err error) {
	if !condition {
		panic(err)
	}
}

func AssertNoError(err error, msg error) {
	if err != nil {
		panic(msg)
	}
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func NewEf(err error, msg string, args ...any) error {
	return yerrors.WrapFrame(yerrors.Errorf("%s while %s", fmt.Sprintf(msg, args...), err.Error()), 1)
}

func ErrMarshal(err error) error {
	return NewEf(err, "could not marshal into []byte")
}

func Newf(msg string, a ...any) error {
	if len(a) > 0 {
		return yerrors.WrapFrame(yerrors.Errorf(msg, a...), 1)
	}
	return yerrors.New(msg)
}

func NewE(err error) error {
	return yerrors.Wrap(err)
}

func New(msg string) error {
	return yerrors.Wrap(yerrors.New(msg))
}

var NotLoggedIn error = fmt.Errorf("%d Not LoggedIn", http.StatusUnauthorized)
