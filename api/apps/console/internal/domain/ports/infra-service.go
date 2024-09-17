package ports

import "context"

type InfraService interface {
	EnsureGlobalVPNConnection(ctx context.Context, args EnsureGlobalVPNConnectionIn) error
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
