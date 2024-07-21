package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	proto_rpc "github.com/kloudlite/api/apps/tenant-agent/internal/proto-rpc"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"google.golang.org/grpc/metadata"
)

type vectorGrpcProxyServer struct {
	proto_rpc.UnimplementedVectorServer
	realVectorClient proto_rpc.VectorClient
	logger           *slog.Logger

	errCh chan error

	accessToken string
	accountName string
	clusterName string

	pushEventsCounter  int
	healthCheckCounter int
}

func (v *vectorGrpcProxyServer) PushEvents(ctx context.Context, msg *proto_rpc.PushEventsRequest) (*proto_rpc.PushEventsResponse, error) {
	if v.realVectorClient == nil {
		return nil, errors.Newf("vector client is not yet connected to message-office")
	}

	outgoingCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", v.accessToken))
	logger := v.logger.With("request-id", fn.UUID())

	v.pushEventsCounter++
	logger.Debug("received push-events message")
	start := time.Now()
	defer logger.Info("dispatched push-events message", "took", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))

	per, err := v.realVectorClient.PushEvents(outgoingCtx, msg)
	if err != nil {
		v.logger.Error("while pushing events got", "err", err)
		if v.errCh != nil {
			v.errCh <- err
		}
		return nil, errors.NewE(err)
	}
	return per, nil
}

func (v *vectorGrpcProxyServer) HealthCheck(ctx context.Context, msg *proto_rpc.HealthCheckRequest) (*proto_rpc.HealthCheckResponse, error) {
	if v.realVectorClient == nil {
		return nil, errors.Newf("vector client is not yet connected to message-office")
	}

	outgoingCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", v.accessToken))

	logger := v.logger.With("request-id", fn.UUID())

	v.healthCheckCounter++
	logger.Debug("received health-check message")
	start := time.Now()
	defer logger.Debug("dispatched health-check message", "took", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))
	hcr, err := v.realVectorClient.HealthCheck(outgoingCtx, msg)
	if err != nil {
		v.logger.Error("while health-checking got", "err", err)
		if v.errCh != nil {
			v.errCh <- err
		}
		return nil, errors.NewE(err)
	}
	return hcr, nil
}
