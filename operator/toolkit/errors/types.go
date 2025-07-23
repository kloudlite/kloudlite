package errors

import (
	"fmt"

	std_errors "errors"
	"github.com/pkg/errors"
)

func NewEf(err error, msg string, a ...any) error {
	return errors.Wrapf(err, "[while] %s [encountered] %s", fmt.Sprintf(msg, a...), err.Error())
	// return errors.Wrapf(err, msg, a...)
	// return yerrors.WrapFrame(yerrors.Errorf("[while] %s [encountered] %s", fmt.Sprintf(msg, a...), err.Error()), 1)
}

func Newf(msg string, a ...any) error {
	return errors.Errorf(msg, a...)
}

func NewE(err error) error {
	return errors.WithStack(err)
}

func New(msg string) error {
	return errors.New(msg)
}

func Join(errs ...error) error {
	return std_errors.Join(errs...)
}

func NotInLocals(key string) error {
	return errors.New(fmt.Sprintf("key %s not found in req.locals", key))
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
