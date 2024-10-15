package ports

import "context"

type InfraService interface {
	EnsureGlobalVPNConnection(ctx context.Context, args EnsureGlobalVPNConnectionIn) error
	GetByokClusterOwnedBy(ctx context.Context, args IsClusterLabelsIn) (string, error)
}

type IsClusterLabelsIn struct {
	UserId    string
	UserEmail string
	UserName  string

	AccountName string
	ClusterName string
}

type EnsureGlobalVPNConnectionIn struct {
	UserId    string
	UserEmail string
	UserName  string

	AccountName   string
	ClusterName   string
	GlobalVPNName string

	DispatchAddrAccountName string
	DispatchAddrClusterName string
}
