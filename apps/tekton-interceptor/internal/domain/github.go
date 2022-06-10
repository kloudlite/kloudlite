package domain

import (
	"context"
	tekton "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"kloudlite.io/pkg/errors"
	"net/http"
	"net/url"
)

type github struct{}

func (gh github) Handle(ctx context.Context, req *tekton.InterceptorRequest) *tekton.InterceptorResponse {
	reqUrl, err := url.Parse(req.Context.EventURL)
	if err != nil {
		return SendResponse(req).Err(err, http.StatusBadRequest)
	}

	if !reqUrl.Query().Has("pipelineId") {
		return SendResponse(req).Err(
			errors.NewEf(err, "url does not have query params 'pipelineId'"),
			http.StatusBadRequest,
		)
	}

	return SendResponse(req).Err(errors.Newf("bad request", http.StatusBadRequest))
}
