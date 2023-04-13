package domain

import "context"

type Domain interface {
	GenClusterToken(ctx context.Context, accountName string, clusterName string) (string, error)
	GetClusterToken(ctx context.Context, accountName string, clusterName string) (string, error)
}
