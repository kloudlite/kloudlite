package errors

import (
	"fmt"

	"github.com/yext/yerrors"
)

func NewEf(err error, msg string, a ...interface{}) error {
	return yerrors.WrapFrame(yerrors.Errorf("[while] %s [encountered] %s", fmt.Sprintf(msg, a...), err.Error()), 1)
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

func NotInLocals(key string) error {
	return yerrors.New(fmt.Sprintf("key %s not found in req.locals", key))
}

type HttpError struct {
	Code int
	Msg  string
}

func (h HttpError) Error() string {
	return fmt.Sprintf("[ CODE: %d ] %s", h.Code, h.Msg)
}

func NewHttpError[T []byte | string](code int, msg T) error {
	return &HttpError{Code: code, Msg: string(msg)}
}
