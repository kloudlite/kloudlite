package domain

import (
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
)

type K8sApplier struct {
	Apply func(s *batchv1.Job) error
}

type GqlClient struct {
	Request func(query string, variables map[string]interface{}) (*http.Request, error)
}
