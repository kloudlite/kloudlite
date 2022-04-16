package application

import (
	"context"
	"fmt"
	"strings"

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

func (s *server) Can(ctx context.Context, in *iam.InCan) (*iam.OutCan, error) {
	rb, err := s.rbRepo.FindOne(ctx, repos.Query{
		Filter: repos.Filter{
			"resource_id": map[string]interface{}{"$in": in.ResourceIds},
			"user_id":     in.UserId,
		},
	})

	if err != nil {
		if rb == nil {
			return &iam.OutCan{Status: false}, nil
		}
		return nil, errors.NewEf(err, "could not find resource(ids=%v)", in.ResourceIds)
	}

	fmt.Println("HERE2")
	if strings.HasPrefix(in.UserId, "sys-user") {
		return &iam.OutCan{Status: true}, nil
	}

	for _, role := range ActionMap[Action(in.Action)] {
		if role == Role(rb.Role) {
			return &iam.OutCan{Status: true}, nil
		}
	}

	return &iam.OutCan{Status: false}, nil
}

func (s *server) ListMemberships(ctx context.Context, in *iam.InListMemberships) (*iam.OutListMemberships, error) {
	rbs, err := s.rbRepo.Find(ctx, repos.Query{Filter: repos.Filter{"user_id": in.UserId}})
	if err != nil {
		return nil, errors.NewEf(err, "could not find memberships by (userId=%q)", in.UserId)
	}

	result := []*iam.RoleBinding{}
	for _, rb := range rbs {
		result = append(result, &iam.RoleBinding{
			UserId:       rb.UserId,
			ResourceType: rb.ResourceType,
			ResourceId:   rb.ResourceId,
			Role:         rb.Role,
		})
	}

	return &iam.OutListMemberships{
		RoleBindings: result,
	}, nil
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

func (s *server) RemoveMembership(ctx context.Context, in *iam.InRemoveMembership) (*iam.OutRemoveMembership, error) {
	rb, err := s.rbRepo.FindOne(ctx, repos.Query{
		Filter: map[string]interface{}{},
		Sort:   map[string]interface{}{},
	})
	if err != nil {
		return nil, errors.NewEf(err, "could not findone")
	}

	err = s.rbRepo.DeleteById(ctx, rb.Id)
	if err != nil {
		return nil, errors.NewEf(err, "could not delete resource(id=%s)", rb.Id)
	}

	return &iam.OutRemoveMembership{Result: true}, nil
}

func (s *server) RemoveResource(ctx context.Context, in *iam.InRemoveResource) (*iam.OutRemoveResource, error) {
	err := s.rbRepo.DeleteMany(ctx, repos.Filter{"resource_id": in.ResourceId})
	if err != nil {
		return nil, errors.NewEf(err, "could not delete resources(id=%s)", in.ResourceId)
	}
	return &iam.OutRemoveResource{Result: true}, nil
}

func (i *server) Ping(ctx context.Context, in *iam.Message) (*iam.Message, error) {
	fmt.Println("", in.Message)
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
	repos.NewFxMongoRepo[*entities.RoleBinding]("role_bindings", "rb", entities.RoleBindingIndices),
	fx.Invoke(func(server *grpc.Server, iamService iam.IAMServer) {
		iam.RegisterIAMServer(server, iamService)
	}),
)
