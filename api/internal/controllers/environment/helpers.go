package environment

import (
	"strings"
)

// joinErrors joins multiple errors into a single error message
func joinErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var sb strings.Builder
	for i, err := range errors {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return joinError(sb.String())
}

// joinError is a simple error type for joined error messages
type joinError string

func (e joinError) Error() string {
	return string(e)
}
