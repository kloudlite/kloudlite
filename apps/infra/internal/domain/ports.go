package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type InfraClient interface {
	CreateCluster(cxt context.Context, action SetupClusterAction) (publicIp string, publicKey string, err error)
	UpdateCluster(cxt context.Context, action UpdateClusterAction) (e error)
	DeleteCluster(cxt context.Context, action DeleteClusterAction) (e error)
	AddPeer(cxt context.Context, action AddPeerAction) (e error)
	DeletePeer(cxt context.Context, action DeletePeerAction) (e error)
	GetResourceOutput(ctx context.Context, clusterId repos.ID, resName string, namespace string) ([]byte, error)
}

type InfraJobResponder interface {
	SendCreateClusterResponse(cxt context.Context, action SetupClusterResponse) error
	SendUpdateClusterResponse(cxt context.Context, action UpdateClusterResponse) error
	SendDeleteClusterResponse(cxt context.Context, action DeleteClusterResponse) error
	SendAddPeerResponse(cxt context.Context, action AddPeerResponse) error
	SendDeletePeerResponse(cxt context.Context, action DeletePeerResponse) error
}
