package domain

import (
	"context"
	tekton "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

type gitlab struct{}

func (g gitlab) Handle(ctx context.Context, req *tekton.InterceptorRequest) *tekton.InterceptorResponse {
	// TODO implement me
	panic("implement me")
}
