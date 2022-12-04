package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func Get[T any](req *http.Request) (*T, *http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, resp, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp, errors.New("bad status code, should be >= 200 & < 300")
	}

	var b T
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	if err := json.Unmarshal(body, &b); err != nil {
		return nil, resp, err
	}
	return &b, resp, err
}
