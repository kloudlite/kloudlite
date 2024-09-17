package infra_service

import (
	"context"

	"github.com/kloudlite/api/apps/console/internal/domain/ports"
	"github.com/kloudlite/api/apps/infra/protobufs/infra"
)

type InfraService struct {
	infraClient infra.InfraClient
}

// EnsureGlobalVPNConnection implements ports.InfraService.
func (s *InfraService) EnsureGlobalVPNConnection(ctx context.Context, args ports.EnsureGlobalVPNConnectionIn) error {
	_, err := s.infraClient.EnsureGlobalVPNConnection(ctx, &infra.EnsureGlobalVPNConnectionIn{
		UserId:                   args.UserId,
		UserName:                 args.UserName,
		UserEmail:                args.UserEmail,
		AccountName:              args.AccountName,
		ClusterName:              args.ClusterName,
		GlobalVPNName:            args.GlobalVPNName,
		DispatchAddr_AccountName: args.DispatchAddrAccountName,
		DispatchAddr_ClusterName: args.DispatchAddrClusterName,
	})
	if err != nil {
		return err
	}

	return nil
}

var _ ports.InfraService = (*InfraService)(nil)

func NewInfraService(infraClient infra.InfraClient) ports.InfraService {
	return &InfraService{
		infraClient: infraClient,
	}
}
