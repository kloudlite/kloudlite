package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/kloudlite/api/pkg/errors"

	proto_rpc "github.com/kloudlite/api/apps/message-office/internal/app/proto-rpc"
	"github.com/kloudlite/api/apps/message-office/internal/domain"
)

type vectorProxyServer struct {
	proto_rpc.UnimplementedVectorServer
	realVectorClient   proto_rpc.VectorClient
	logger             *slog.Logger
	domain             domain.Domain
	tokenHashingSecret string

	sync.Mutex
	pushEventsCounter int
}

func (v *vectorProxyServer) PushEvents(ctx context.Context, msg *proto_rpc.PushEventsRequest) (*proto_rpc.PushEventsResponse, error) {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(ctx, v.tokenHashingSecret)
	if err != nil {
		return nil, errors.NewE(err)
	}

	v.Lock()
	v.pushEventsCounter++
	v.Unlock()

	logger := v.logger.With("account", accountName, "cluster", clusterName, "counter", v.pushEventsCounter)

	logger.Debug("RECEIVED push-events message")

	nctx, cf := context.WithTimeout(ctx, 3*time.Second)
	defer cf()

	per, err := v.realVectorClient.PushEvents(nctx, msg)
	if err != nil {
		logger.Error("FAILED to dispatch push-events message, got", "err", err)
		return nil, errors.NewE(err)
	}
	logger.Debug("DISPATCHED push-events message")
	return per, nil
}

func (v *vectorProxyServer) HealthCheck(ctx context.Context, msg *proto_rpc.HealthCheckRequest) (*proto_rpc.HealthCheckResponse, error) {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(ctx, v.tokenHashingSecret)
	if err != nil {
		return nil, errors.NewE(err)
	}

	logger := v.logger.With("account", accountName, "cluster", clusterName)
	logger.Debug("RECEIVED health-check message")
	hcr, err := v.realVectorClient.HealthCheck(ctx, msg)
	if err != nil {
		logger.Error("FAILED to dispatch health-check message, got", "err", err)
		return nil, errors.NewE(err)
	}
	logger.Debug("DISPATCHED health-check message")
	return hcr, nil
}
