package app

import (
	"context"
	"kloudlite.io/apps/message-office/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
)

type internalMsgServer struct {
	d domain.Domain
	message_office_internal.UnimplementedMessageOfficeInternalServer
}

func (s *internalMsgServer) GenerateClusterToken(ctx context.Context, in *message_office_internal.GenerateClusterTokenIn) (*message_office_internal.GenerateClusterTokenOut, error) {
	token, err := s.d.GenClusterToken(ctx, in.AccountName, in.ClusterName)
	if err != nil {
		return nil, err
	}
	return &message_office_internal.GenerateClusterTokenOut{
		ClusterToken: token,
	}, nil
}

func newInternalMsgServer(d domain.Domain) *internalMsgServer {
	return &internalMsgServer{d: d}
}
