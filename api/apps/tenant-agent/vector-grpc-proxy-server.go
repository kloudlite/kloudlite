package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	proto_rpc "github.com/kloudlite/api/apps/tenant-agent/internal/proto-rpc"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type vectorGrpcProxyServer struct {
	proto_rpc.UnimplementedVectorServer

	realVectorClient proto_rpc.VectorClient
	connCancelFn     context.CancelFunc

	logger *slog.Logger

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
	logger := v.logger.With("request-id", fmt.Sprintf("%s-%s", fn.UUID(4), fn.UUID(4)))

	v.pushEventsCounter++
	logger.Debug("received push-events message")
	start := time.Now()
	per, err := v.realVectorClient.PushEvents(outgoingCtx, msg)
	if err != nil {
		v.connCancelFn()
		if status.Code(err) == codes.Canceled {
			return nil, err
		}
		v.logger.Error("FAILED to dispatch push-events message, got", "err", err)
		return nil, errors.NewE(err)
	}
	logger.Debug("DISPATCHED push-events message", "took", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))
	return per, nil
}

func (v *vectorGrpcProxyServer) HealthCheck(ctx context.Context, msg *proto_rpc.HealthCheckRequest) (*proto_rpc.HealthCheckResponse, error) {
	if v.realVectorClient == nil {
		return nil, errors.Newf("vector client is not yet connected to message-office")
	}

	outgoingCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", v.accessToken))

	logger := v.logger.With("request-id", fmt.Sprintf("%s-%s", fn.UUID(4), fn.UUID(4)))

	v.healthCheckCounter++
	logger.Debug("RECEIVED health-check message")
	start := time.Now()
	hcr, err := v.realVectorClient.HealthCheck(outgoingCtx, msg)
	if err != nil {
		v.connCancelFn()
		if status.Code(err) == codes.Canceled {
			return nil, err
		}
		v.logger.Error("FAILED to dispatch health-check message, got", "err", err)
		return nil, errors.NewE(err)
	}
	logger.Debug("DISPATCHED health-check message", "took", fmt.Sprintf("%.3fs", time.Since(start).Seconds()))
	return hcr, nil
}
