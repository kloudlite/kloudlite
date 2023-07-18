package main

import (
	"context"
	"fmt"

	proto_rpc "github.com/kloudlite/operator/agent/internal/proto-rpc"
	"github.com/kloudlite/operator/pkg/logging"
	"google.golang.org/grpc/metadata"
)

const (
	AccessTokenHeader string = "kloudlite-access-token"
	AccountNameHeader string = "kloudlite-account-name"
	ClusterNameHeader string = "kloudlite-cluster-name"
)

type vectorGrpcProxyServer struct {
	proto_rpc.UnimplementedVectorServer
	realVectorClient proto_rpc.VectorClient
	logger           logging.Logger

	accessToken string
	accountName string
	clusterName string

	pushEventsCounter  int
	healthCheckCounter int
}

func (v *vectorGrpcProxyServer) PushEvents(ctx context.Context, msg *proto_rpc.PushEventsRequest) (*proto_rpc.PushEventsResponse, error) {
	if v.realVectorClient == nil {
		return nil, fmt.Errorf("real vector client is not established yet")
	}

	md := metadata.Pairs(AccessTokenHeader, v.accessToken)
	md.Append(AccountNameHeader, v.accountName)
	md.Append(ClusterNameHeader, v.clusterName)

	outgoingCtx := metadata.NewOutgoingContext(ctx, md)

	v.pushEventsCounter++
	v.logger.Infof("[%v] received push-events message", v.pushEventsCounter)

	per, err := v.realVectorClient.PushEvents(outgoingCtx, msg)
	if err != nil {
		v.logger.Error(err)
		return nil, err
	}
	return per, nil

}

func (v *vectorGrpcProxyServer) HealthCheck(ctx context.Context, msg *proto_rpc.HealthCheckRequest) (*proto_rpc.HealthCheckResponse, error) {
	if v.realVectorClient == nil {
		return nil, fmt.Errorf("real vector client is not established yet")
	}

	md := metadata.Pairs(AccessTokenHeader, v.accessToken)
	md.Append(AccountNameHeader, v.accountName)
	md.Append(ClusterNameHeader, v.clusterName)

	outgoingCtx := metadata.NewOutgoingContext(ctx, md)

	v.healthCheckCounter++
	v.logger.Infof("[%v] received health-check message", v.healthCheckCounter)
	hcr, err := v.realVectorClient.HealthCheck(outgoingCtx, msg)
	if err != nil {
		v.logger.Error(err)
		return nil, err
	}
	return hcr, nil
}
