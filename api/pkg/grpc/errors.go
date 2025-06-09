package grpc

import (
	"strings"

	"github.com/kloudlite/api/pkg/grpc/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Err(err error, comments ...string) error {
	s := status.New(codes.Unknown, err.Error())
	s2, e := s.WithDetails(&errors.Error{
		Message:  err.Error(),
		Comments: strings.Join(comments, "\n"),
	})
	if e != nil {
		panic(e)
	}
	return s2.Err()
}

func ParseErr(err error) *errors.Error {
	s := status.Convert(err)
	for _, detail := range s.Details() {
		e, ok := detail.(*errors.Error)
		if !ok {
			continue
		}
		return e
	}
	return nil
}
