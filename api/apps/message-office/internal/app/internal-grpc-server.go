package app

import (
	"context"

	"github.com/kloudlite/api/apps/message-office/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	"github.com/kloudlite/api/pkg/errors"
)

type internalMsgServer struct {
	d domain.Domain
	message_office_internal.UnimplementedMessageOfficeInternalServer
}

func (s *internalMsgServer) GenerateClusterToken(ctx context.Context, in *message_office_internal.GenerateClusterTokenIn) (*message_office_internal.GenerateClusterTokenOut, error) {
	token, err := s.d.GenClusterToken(ctx, in.AccountName, in.ClusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &message_office_internal.GenerateClusterTokenOut{
		ClusterToken: token,
	}, nil
}

func (s *internalMsgServer) GetClusterToken(ctx context.Context, in *message_office_internal.GetClusterTokenIn) (*message_office_internal.GetClusterTokenOut, error) {
	token, err := s.d.GetClusterToken(ctx, in.AccountName, in.ClusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &message_office_internal.GetClusterTokenOut{
		ClusterToken: token,
	}, nil
}

func newInternalMsgServer(d domain.Domain) message_office_internal.MessageOfficeInternalServer {
	return &internalMsgServer{d: d}
}
