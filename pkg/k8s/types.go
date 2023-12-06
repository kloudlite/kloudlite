package k8s

import (
	"encoding/json"
	"fmt"
)

type InvalidSchemaError struct {
	err     error
	errMsgs []string
}

func (ise InvalidSchemaError) Error() string {
	m := map[string]any{
		"message":          ise.err.Error(),
		"type":             "InvalidData",
		"validationErrors": ise.errMsgs,
	}
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Println("[UNEXPECTED] failed to marshal InvalidSchemaError", err)
		return ise.err.Error()
	}
	return string(b)
}

func NewInvalidSchemaError(err error, errMsgs []string) InvalidSchemaError {
	return InvalidSchemaError{err: err, errMsgs: errMsgs}
}
