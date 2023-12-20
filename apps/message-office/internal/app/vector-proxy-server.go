package app

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"

	proto_rpc "github.com/kloudlite/api/apps/message-office/internal/app/proto-rpc"
	"github.com/kloudlite/api/apps/message-office/internal/domain"
	"github.com/kloudlite/api/pkg/logging"
)

type vectorProxyServer struct {
	proto_rpc.UnimplementedVectorServer
	realVectorClient   proto_rpc.VectorClient
	logger             logging.Logger
	domain             domain.Domain
	tokenHashingSecret string
	pushEventsCounter  int
	healthCheckCounter int
}

func (v *vectorProxyServer) PushEvents(ctx context.Context, msg *proto_rpc.PushEventsRequest) (*proto_rpc.PushEventsResponse, error) {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(ctx, v.tokenHashingSecret)
	if err != nil {
		return nil, errors.NewE(err)
	}

	logger := v.logger.WithKV("accountName", accountName, "clusterName", clusterName)
	v.pushEventsCounter++
	logger.Infof("[%v] received push-events message", v.pushEventsCounter)
	defer logger.Infof("[%v] dispatched push-events message to vector aggregator", v.pushEventsCounter)

	per, err := v.realVectorClient.PushEvents(ctx, msg)
	if err != nil {
		logger.Errorf(err)
		return nil, errors.NewE(err)
	}
	return per, nil
}

func (v *vectorProxyServer) HealthCheck(ctx context.Context, msg *proto_rpc.HealthCheckRequest) (*proto_rpc.HealthCheckResponse, error) {
	accountName, clusterName, err := validateAndDecodeFromGrpcContext(ctx, v.tokenHashingSecret)
	if err != nil {
		return nil, errors.NewE(err)
	}

	logger := v.logger.WithKV("accountName", accountName, "clusterName", clusterName)
	v.healthCheckCounter++
	logger.Infof("[%v] received health-check message", v.healthCheckCounter)
	defer logger.Infof("[%v] dispatched health-check message to vector aggregator", v.healthCheckCounter)

	hcr, err := v.realVectorClient.HealthCheck(ctx, msg)
	if err != nil {
		logger.Errorf(err)
		return nil, errors.NewE(err)
	}
	return hcr, nil
}
