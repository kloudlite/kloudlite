package domain

import "context"

type Domain interface {
	GenClusterToken(ctx context.Context, accountName string, clusterName string) (string, error)
	FindClusterToken(ctx context.Context, clusterToken string) (*MessageOfficeToken, error)
	GetClusterToken(ctx context.Context, accountName string, clusterName string) (string, error)
	GenAccessToken(ctx context.Context, clusterToken string) (*AccessToken, error)
	ValidateAccessToken(ctx context.Context, accessToken, accountName, clusterName string) error
}
