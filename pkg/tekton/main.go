package tekton

import (
	"encoding/json"
	triggers "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"google.golang.org/grpc/codes"
	"net/http"
)

type Request struct {
	triggers.InterceptorRequest
}

type errT struct {
	Code int
	error
}

func (err errT) Error() string {
	return err.error.Error()
}

func NewError(code int, err error) error {
	return errT{code, err}
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

func (r *Response) Err(err error) *Response {
	e := func() errT {
		if tkErr, ok := err.(errT); ok {
			return tkErr
		}
		return errT{Code: http.StatusInternalServerError, error: err}
	}()

	r.InterceptorResponse.Continue = false

	r.InterceptorResponse.Status = triggers.Status{
		Code:    codes.Code(e.Code),
		Message: err.Error(),
	}

	return r
}

func (r *Response) Extend(m map[string]any) *Response {
	for k, v := range m {
		if r.InterceptorResponse.Extensions == nil {
			r.InterceptorResponse.Extensions = map[string]any{}
		}
		r.InterceptorResponse.Extensions[k] = v
	}
	return r
}

func (r *Response) Ok() *Response {
	r.InterceptorResponse.Continue = true
	// r.InterceptorResponse.Status = triggers.Status{
	// 	Code:    http.StatusOK,
	// 	Message: "success",
	// }
	return r
}

func (r *Response) ToJson() ([]byte, error) {
	m, err := json.Marshal(r.InterceptorResponse)
	if err != nil {
		return nil, err
	}
	return m, nil
}
