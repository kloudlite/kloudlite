package errors

import "errors"

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

func New(text string) error {
	return errors.New(text)
}
