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

func Wrap(e error, msg string) error {
	return errors.Wrap(e, msg)
}

func NewEf(err error, msg string, a ...any) error {
	return yerrors.Errorf("%s as %+v", fmt.Sprintf(msg, a...), err)
}

func ErrMarshal(err error) error {
	return NewEf(err, "could not marshal into []byte")
}

func Newf(msg string, a ...any) error {
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

var NotLoggedIn error = fmt.Errorf("%d Not LoggedIn", http.StatusUnauthorized)
