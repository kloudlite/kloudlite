package grpc

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
)

type ConnectOpts struct {
	SecureConnect bool
	Timeout       time.Duration
}

func Connect(url string, opts ConnectOpts) (*grpc.ClientConn, error) {
	ctx, cf := context.WithTimeout(context.TODO(), opts.Timeout)
	defer cf()
	if opts.SecureConnect {
		conn, err := grpc.DialContext(ctx, url, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
		})))
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to connect: %v, please retry", err)
		return nil, err
	}
	// conn, err := grpc.DialContext(ctx, url, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	conn, err := grpc.DialContext(ctx, url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err == nil {
		return conn, nil
	}
	log.Printf("Failed to connect: %v, please retry", err)
	return nil, err
}

func ConnectSecure(url string) (*grpc.ClientConn, error) {
	for {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: false,
		})), grpc.WithBlock())
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to connect: %v, retrying...", err)
		time.Sleep(2 * time.Second)
	}
}

type GrpcServerOpts struct {
	Logger logging.Logger
}

type GrpcServer struct {
	GrpcServer *grpc.Server
	logger     logging.Logger
}

func (g *GrpcServer) Listen(addr string) error {
	listen, err := net.Listen("tcp", addr)
	defer func() {
		g.logger.Infof("[GRPC Server] started on port (%s)", addr)
	}()
	if err != nil {
		return errors.NewEf(err, "could not listen to net/tcp server")
	}

	go func() error {
		err := g.GrpcServer.Serve(listen)
		if err != nil {
			return errors.NewEf(err, "could not start grpc server ")
		}
		return nil
	}()
	return nil
}

func (g *GrpcServer) Stop() {
	g.GrpcServer.Stop()
}

func NewGrpcServer(opts ...GrpcServerOpts) *GrpcServer {
	logger := func() logging.Logger {
		if len(opts) == 0 {
			return logging.NewOrDie(&logging.Options{Name: "grpc-server", Dev: false})
		}
		return opts[0].Logger
	}()

	server := grpc.NewServer(
		grpc.StreamInterceptor(func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			p, ok := peer.FromContext(stream.Context())
			if ok {
				logger.Debugf("New connection from %s", p.Addr.String())
			}
			return handler(srv, stream)
		}),
	)

	return &GrpcServer{GrpcServer: server, logger: logger}
}
