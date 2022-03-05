package domain

import (
	batchv1 "k8s.io/api/batch/v1"
)

type K8sApplier struct {
	Apply func(s *batchv1.Job) error
}
