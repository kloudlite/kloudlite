package grpc

import (
	"google.golang.org/grpc"
	"net"
	"operators.kloudlite.io/lib/errors"
)

type Server struct {
	*grpc.Server
}

type Options struct {
	Addr string `json:"addr,omitempty"`
}

func NewServer(opts Options) (*Server, error) {
	listener, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}
	gServer := grpc.NewServer()
	go func() {
		if err := gServer.Serve(listener); err != nil {
			panic(errors.NewEf(err, "could not start grpc server"))
		}
	}()
	return &Server{Server: gServer}, nil
}
