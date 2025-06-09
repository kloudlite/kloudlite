package domain

import "encoding/json"

type UnAuthorizedError struct {
	parentErr error
	debugMsg  string
}

func (e UnAuthorizedError) Error() string {
	b, _ := json.Marshal(map[string]any{
		"error": "unauthorized",
	})
	return string(b)
}
