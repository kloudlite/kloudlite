package errors

import (
	"fmt"

	"github.com/ztrue/tracerr"
	"net/http"

	"github.com/pkg/errors"
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
	return tracerr.Wrap(fmt.Errorf("%s while %s", fmt.Sprintf(msg, args...), err.Error()))
}

func ErrMarshal(err error) error {
	return NewEf(err, "could not marshal into []byte")
}

func Newf(msg string, a ...any) error {
	if len(a) > 0 {
		return tracerr.Wrap(fmt.Errorf(msg, a...))
	}
	return tracerr.New(msg)
}

func NewE(err error) error {
	if err == nil {
		return nil
	}
	return tracerr.Wrap(err)
}

func New(msg string) error {
	return tracerr.Wrap(tracerr.New(msg))
}

var NotLoggedIn error = fmt.Errorf("%d Not LoggedIn", http.StatusUnauthorized)
