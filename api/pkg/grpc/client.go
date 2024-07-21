package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	grpc.ClientConnInterface
	Close() error
}

type grpcClient struct {
	*grpc.ClientConn
}

func (g *grpcClient) Close() error {
	return g.ClientConn.Close()
}

func NewGrpcClient(serverAddr string) (Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &grpcClient{ClientConn: conn}, nil
}

type GrpcConnectOpts struct {
	TLSConnect bool
	Logger     *slog.Logger

	ReconnectCheckInterval *time.Duration
}

func NewGrpcClientV2(serverAddr string, opts GrpcConnectOpts) (Client, error) {
	tc := insecure.NewCredentials()
	if opts.TLSConnect {
		tc = credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
	}

	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.ReconnectCheckInterval == nil {
		rafter := 2 * time.Second
		opts.ReconnectCheckInterval = &rafter
	}

	opts.Logger.Debug("ATTEMPTING to connect")
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(tc))
	if err != nil {
		return nil, errors.NewE(err)
	}

	gc := &grpcClient{ClientConn: conn}

	go func() {
		attempt := 0
		prevState := connectivity.Idle
		reconnecting := false
		start := time.Now()
		for {
			cstate := conn.GetState()

			if cstate != connectivity.Ready {
				opts.Logger.Warn("connection is not in ready state", "current", cstate)

				if cstate == connectivity.Shutdown {
					opts.Logger.Info("connection is closing, shutting down...")
					return
				}

				if cstate != connectivity.Connecting && !reconnecting {
					start = time.Now()
					conn.Connect() // reconnecting
					opts.Logger.Debug("ATTEMPTING re-connect")
					reconnecting = true
					attempt++
					continue
				}
			}
			if cstate == connectivity.Ready && prevState != connectivity.Ready {
				if attempt > 0 {
					reconnecting = false
					opts.Logger.Info("RE-CONNECTED", "attempt", attempt, "took", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))
				}
				opts.Logger.Info("CONNECTED", "took", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))
			}

			prevState = cstate
			// <-time.After(*opts.ReconnectCheckInterval)
			ctx, cf := context.WithTimeout(context.TODO(), *opts.ReconnectCheckInterval)
			conn.WaitForStateChange(ctx, cstate)
			cf()
		}
	}()

	return gc, nil
}
