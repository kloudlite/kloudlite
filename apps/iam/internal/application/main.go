package application

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"kloudlite.io/apps/iam/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type server struct {
	iam.UnimplementedIAMServer
	rbRepo repos.DbRepo[*entities.RoleBinding]
}

func (s *server) Can(_ context.Context, _ *iam.InCan) (*iam.OutCan, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) ListMemberships(_ context.Context, _ *iam.InListMemberships) (*iam.OutListMemberships, error) {
	panic("not implemented") // TODO: Implement
}

// Mutation
func (s *server) AddMembership(ctx context.Context, in *iam.InAddMembership) (*iam.OutAddMembership, error) {
	_, err := s.rbRepo.Create(ctx, &entities.RoleBinding{
		UserId:       in.UserId,
		ResourceType: in.ResourceType,
		ResourceId:   in.ResourceId,
		Role:         in.Role,
	})
	if err != nil {
		return nil, errors.NewEf(err, "could not create rolebinding")
	}
	return &iam.OutAddMembership{Result: true}, nil
}

func (s *server) RemoveMembership(_ context.Context, _ *iam.InRemoveMembership) (*iam.OutRemoveMembership, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) RemoveResource(_ context.Context, _ *iam.InRemoveResource) (*iam.OutRemoveResource, error) {
	panic("not implemented") // TODO: Implement
}

func (s *server) mustEmbedUnimplementedIAMServer() {
	panic("not implemented") // TODO: Implement
}

func (i *server) Ping(ctx context.Context, in *iam.Message) (*iam.Message, error) {
	fmt.Println("Ping sadfasdfsadf", in.Message)
	return &iam.Message{
		Message: "asdfasdf",
	}, nil
}

func fxServer(rbRepo repos.DbRepo[*entities.RoleBinding]) iam.IAMServer {
	return &server{
		rbRepo: rbRepo,
	}
}

var Module = fx.Module("application",
	fx.Provide(fxServer),
	repos.NewFxMongoRepo[*entities.RoleBinding]("role_bindings", "rb", entities.RoleBindingIndexes),
	fx.Invoke(func(server *grpc.Server, iamService iam.IAMServer) {
		iam.RegisterIAMServer(server, iamService)
	}),
)
