package grpc

import (
	"fmt"
	"net"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Server interface {
	grpc.ServiceRegistrar
	Listen(addr string) error
	Stop()
}

type ServerOpts struct {
	Logger logging.Logger
}

type grpcServer struct {
	*grpc.Server
	logger logging.Logger
}

func (g *grpcServer) Listen(addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.NewEf(err, "could not listen to net/tcp server")
	}
	g.logger.Infof("listening on %s", addr)
	return g.Serve(listen)
}

func (g *grpcServer) Stop() {
	g.Server.GracefulStop()
}

func NewGrpcServer(opts ServerOpts) (Server, error) {
	if opts.Logger == nil {
		lgr, err := logging.New(&logging.Options{Name: "grpc-server", Dev: false})
		if err != nil {
			return nil, err
		}
		opts.Logger = lgr
	}

	server := grpc.NewServer(
		grpc.StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			p, ok := peer.FromContext(stream.Context())
			if ok {
				fmt.Printf("[Stream] New connection from %s", p.Addr.String())
			}
			return handler(srv, stream)
		}),
	)

	return &grpcServer{Server: server, logger: opts.Logger}, nil
}

// Type guard to ensure grpcServer implements Server interface, at compile time
var _ Server = &grpcServer{}
