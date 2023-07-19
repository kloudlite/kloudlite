package app

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	proto_rpc "kloudlite.io/apps/message-office/internal/app/proto-rpc"
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/pkg/logging"
)

type vectorProxyServer struct {
	proto_rpc.UnimplementedVectorServer
	realVectorClient   proto_rpc.VectorClient
	logger             logging.Logger
	domain             domain.Domain
	pushEventsCounter  int
	healthCheckCounter int
}

const (
	AccessTokenHeader string = "kloudlite-access-token"
	AccountNameHeader string = "kloudlite-account-name"
	ClusterNameHeader string = "kloudlite-cluster-name"
)

func (v *vectorProxyServer) PushEvents(ctx context.Context, msg *proto_rpc.PushEventsRequest) (*proto_rpc.PushEventsResponse, error) {
	var accessToken, accountName, clusterName string

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to parse headers from incoming context")
	}

	// Access gRPC-specific metadata headers
	// Iterate through the headers if needed
	for key, values := range md {
		switch key {
		case AccessTokenHeader:
			accessToken = values[0]
		case AccountNameHeader:
			accountName = values[0]
		case ClusterNameHeader:
			clusterName = values[0]
		}
	}

	logger := v.logger.WithKV("accountName", accountName).WithKV("cluster", clusterName)
	v.pushEventsCounter++
	logger.Infof("[%v] received push-events message", v.pushEventsCounter)
	// logger.Infof("debugging message %+v\n", msg)

	if err := v.domain.ValidateAccessToken(ctx, accessToken, accountName, clusterName); err != nil {
		return nil, errors.Wrap(err, "failed to validate access token")
	}
	per, err := v.realVectorClient.PushEvents(ctx, msg)
	if err != nil {
		logger.Errorf(err)
		return nil, err
	}
	return per, nil
}

func (v *vectorProxyServer) HealthCheck(ctx context.Context, msg *proto_rpc.HealthCheckRequest) (*proto_rpc.HealthCheckResponse, error) {
	var accessToken, accountName, clusterName string

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to parse headers from incoming context")
	}

	for key, values := range md {
		switch key {
		case AccessTokenHeader:
			accessToken = values[0]
		case AccountNameHeader:
			accountName = values[0]
		case ClusterNameHeader:
			clusterName = values[0]
		}
	}

	logger := v.logger.WithKV("accountName", accountName).WithKV("cluster", clusterName)

	v.healthCheckCounter++
	logger.Infof("[%v] received health-check message", v.healthCheckCounter)

	if err := v.domain.ValidateAccessToken(ctx, accessToken, accountName, clusterName); err != nil {
		return nil, errors.Wrap(err, "failed to validate access token")
	}
	hcr, err := v.realVectorClient.HealthCheck(ctx, msg)
	if err != nil {
		logger.Errorf(err)
		return nil, err
	}
	return hcr, nil
}
