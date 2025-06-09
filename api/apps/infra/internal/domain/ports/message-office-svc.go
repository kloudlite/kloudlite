package ports

import "context"

type MessageOfficeService interface {
	GetAllocatedPlatformEdgeCluster(ctx context.Context, args *GetAllocatedPlatformEdgeClusterIn) (*GetAllocatedPlatformEdgeClusterOut, error)

	GetClusterToken(ctx context.Context, args *GetClusterTokenIn) (*GetClusterTokenOut, error)

	GenerateClusterToken(ctx context.Context, args *GenerateClusterTokenIn) (*GenerateClusterTokenOut, error)
}

type GetAllocatedPlatformEdgeClusterIn struct {
	AccountName string
	ClusterName string
}

type GetAllocatedPlatformEdgeClusterOut struct {
	PublicDNSHost string
}

type GetClusterTokenIn struct {
	AccountName string
	ClusterName string
}

type GetClusterTokenOut struct {
	ClusterToken string
}

type GenerateClusterTokenIn struct {
	AccountName string
	ClusterName string
}

type GenerateClusterTokenOut struct{
	ClusterToken string
}
