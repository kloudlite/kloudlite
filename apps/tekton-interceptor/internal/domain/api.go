package domain

import (
	"context"
	tekton "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

type Domain interface {
	Handle(ctx context.Context, req *tekton.InterceptorRequest) *tekton.InterceptorResponse
}
