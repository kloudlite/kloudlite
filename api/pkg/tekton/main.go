package tekton

import (
	triggers "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"google.golang.org/grpc/codes"
	fn "kloudlite.io/pkg/functions"
	"net/http"
)

type Request struct {
	triggers.InterceptorRequest
}

type Response struct {
	triggers.InterceptorResponse
}

func NewResponse(req *Request) *Response {
	return &Response{
		InterceptorResponse: triggers.InterceptorResponse{
			Extensions: req.Extensions,
			// Continue:   false,
			// Status:     triggers.Status{},
		},
	}
}

func (r *Response) Err(err error, statusCode ...codes.Code) *Response {
	r.InterceptorResponse.Continue = false

	sc := codes.Code(http.StatusInternalServerError)
	if len(statusCode) > 0 {
		sc = statusCode[0]
	}

	r.InterceptorResponse.Status = triggers.Status{
		Code:    sc,
		Message: err.Error(),
	}

	return r
}

func (r *Response) Extend(m map[string]any) *Response {
	for k, v := range m {
		r.InterceptorResponse.Extensions[k] = v
	}
	return r
}

func (r *Response) Ok(message ...string) *Response {
	r.InterceptorResponse.Status = triggers.Status{
		Code:    http.StatusOK,
		Message: fn.First(message),
	}
	return r
}
