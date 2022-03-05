package app

import "net/http"

type GqlClient struct {
	Request func(query string, variables map[string]interface{}) (*http.Request, error)
}
