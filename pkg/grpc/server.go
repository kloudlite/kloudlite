package grpc

import (
	"context"
	"log/slog"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/kloudlite/api/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Server interface {
	grpc.ServiceRegistrar
	Listen(addr string) error
	Stop()
}

type ServerOpts struct {
	Logger *slog.Logger
}

type grpcServer struct {
	*grpc.Server
	logger *slog.Logger
}

func (g *grpcServer) Listen(addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.NewEf(err, "could not listen to net/tcp server")
	}
	g.logger.Info("grpc server listening", "at", addr)
	return g.Serve(listen)
}

func (g *grpcServer) Stop() {
	g.Server.GracefulStop()
}

func NewGrpcServer(opts ServerOpts) (Server, error) {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	grpcLogger := logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		opts.Logger.Log(ctx, slog.Level(lvl), msg, fields...)
	})

	grpcLoggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logging.UnaryServerInterceptor(grpcLogger, grpcLoggingOpts...)),
		grpc.ChainStreamInterceptor(logging.StreamServerInterceptor(grpcLogger, grpcLoggingOpts...)),

		grpc.StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			p, ok := peer.FromContext(stream.Context())
			if ok {
				_ = p.Addr.String()
				// if opts.Slogger != nil {
				// 	opts.Slogger.Debug("new grpc connection", "from", p.Addr.String())
				// } else {
				// 	opts.Logger.Debugf("[Stream] New connection from %s", p.Addr.String())
				// }
			}
			return handler(srv, stream)
		}),
	)

	return &grpcServer{Server: server, logger: opts.Logger}, nil
}

// Type guard to ensure grpcServer implements Server interface, at compile time
var _ Server = &grpcServer{}
