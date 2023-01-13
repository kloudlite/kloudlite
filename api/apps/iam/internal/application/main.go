package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kloudlite.io/constants"
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

func (s *server) ConfirmMembership(ctx context.Context, in *iam.InConfirmMembership) (*iam.OutConfirmMembership, error) {
	one, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":     in.UserId,
			"resource_id": in.ResourceId,
		},
	)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("role binding not found")
	}
	if in.Role != one.Role {
		return nil, errors.New("The invitation has been updated")
	}
	one.Accepted = true
	_, err = s.rbRepo.UpdateById(ctx, one.Id, one)
	if err != nil {
		return nil, err
	}
	return &iam.OutConfirmMembership{}, nil
}

func (s *server) InviteMembership(ctx context.Context, in *iam.InAddMembership) (*iam.OutAddMembership, error) {
	fmt.Println("InviteMembership", in)
	one, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":     in.UserId,
			"resource_id": in.ResourceId,
		},
	)
	if one != nil {
		if one.Role == in.Role {
			return nil, errors.New(fmt.Sprintf("user %s already has role %s on resource %s", in.UserId, in.Role, in.ResourceId))
		}
		one.Role = in.Role
		one.Accepted = false
		_, err := s.rbRepo.UpdateById(ctx, one.Id, one)
		if err != nil {
			return nil, err
		}
		return &iam.OutAddMembership{Result: true}, nil
	}

	_, err = s.rbRepo.Create(
		ctx, &entities.RoleBinding{
			UserId:       in.UserId,
			ResourceType: in.ResourceType,
			ResourceId:   in.ResourceId,
			Role:         in.Role,
			Accepted:     false,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not create rolebinding")
	}
	return &iam.OutAddMembership{Result: true}, nil
}

func (s *server) GetMembership(ctx context.Context, membership *iam.InGetMembership) (*iam.OutGetMembership, error) {
	one, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"resource_id": membership.ResourceId,
			"user_id":     membership.UserId,
		},
	)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New(fmt.Sprintf("role binding not found for resource %s and user %s", membership.ResourceId, membership.UserId))
	}
	return &iam.OutGetMembership{
		UserId:     one.UserId,
		ResourceId: one.ResourceId,
		Role:       one.Role,
		Accepted:   one.Accepted,
	}, nil
}

func (s *server) ListResourceMemberships(ctx context.Context, in *iam.InResourceMemberships) (*iam.OutListMemberships, error) {
	filter := repos.Filter{}
	if in.ResourceId != "" {
		filter["resource_id"] = in.ResourceId
	}
	if in.ResourceType != "" {
		filter["resource_type"] = in.ResourceType
	}

	rbs, err := s.rbRepo.Find(ctx, repos.Query{Filter: filter})
	if err != nil {
		return nil, errors.NewEf(err, "could not find memberships by (resourceId=%q, resourceType=%q)", in.ResourceId, in.ResourceType)
	}

	var result []*iam.RoleBinding
	for _, rb := range rbs {
		result = append(
			result, &iam.RoleBinding{
				UserId:       rb.UserId,
				ResourceType: rb.ResourceType,
				ResourceId:   rb.ResourceId,
				Role:         rb.Role,
			},
		)
	}

	return &iam.OutListMemberships{
		RoleBindings: result,
	}, nil
}

func (s *server) Can(ctx context.Context, in *iam.InCan) (*iam.OutCan, error) {
	rb, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"resource_id": map[string]interface{}{"$in": in.ResourceIds},
			"user_id":     in.UserId,
		},
	)

	if err != nil {
		return nil, errors.NewEf(err, "could not find resource(ids=%v)", in.ResourceIds)
	}

	var actionsConfig struct {
		Actions map[string][]string `json:"actions"`
	}

	file, err := ioutil.ReadFile("./configs/iam.json")
	if err != nil {
		return &iam.OutCan{Status: false}, err
	}

	err = json.Unmarshal(file, &actionsConfig)

	if err != nil {
		return &iam.OutCan{Status: false}, err
	}

	if rb == nil {
		return &iam.OutCan{Status: false}, nil
	}

	if strings.HasPrefix(in.UserId, "sys-user") {
		return &iam.OutCan{Status: true}, nil
	}

	fmt.Println("here", actionsConfig.Actions[in.Action])

	for _, v := range actionsConfig.Actions[in.Action] {
		fmt.Println(v, rb.Role)
		if v == rb.Role {
			return &iam.OutCan{Status: true}, nil
		}

	}

	for _, role := range constants.ActionMap[constants.Action(in.Action)] {
		if role == constants.Role(rb.Role) {
			return &iam.OutCan{Status: true}, nil
		}
	}

	return &iam.OutCan{Status: false}, nil
}

func (s *server) ListUserMemberships(ctx context.Context, in *iam.InUserMemberships) (*iam.OutListMemberships, error) {
	filter := repos.Filter{"user_id": in.UserId}
	if in.ResourceType != "" {
		filter["resource_type"] = in.ResourceType
	}

	rbs, err := s.rbRepo.Find(ctx, repos.Query{Filter: filter})
	if err != nil {
		return nil, errors.NewEf(err, "could not find memberships by (userId=%q)", in.UserId)
	}

	result := []*iam.RoleBinding{}
	for _, rb := range rbs {
		result = append(
			result, &iam.RoleBinding{
				UserId:       rb.UserId,
				ResourceType: rb.ResourceType,
				ResourceId:   rb.ResourceId,
				Role:         rb.Role,
			},
		)
	}

	return &iam.OutListMemberships{
		RoleBindings: result,
	}, nil
}

// Mutation

func (s *server) AddMembership(ctx context.Context, in *iam.InAddMembership) (*iam.OutAddMembership, error) {
	_, err := s.rbRepo.Create(
		ctx, &entities.RoleBinding{
			UserId:       in.UserId,
			ResourceType: in.ResourceType,
			ResourceId:   in.ResourceId,
			Role:         in.Role,
			Accepted:     true,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not create rolebinding")
	}
	return &iam.OutAddMembership{Result: true}, nil
}

func (s *server) RemoveMembership(ctx context.Context, in *iam.InRemoveMembership) (*iam.OutRemoveMembership, error) {
	var rb *entities.RoleBinding
	var err error

	if in.UserId != "" && in.ResourceId != "" {

		rb, err = s.rbRepo.FindOne(
			ctx, repos.Filter{
				"resource_id": in.ResourceId,
				"user_id":     in.UserId,
			},
		)
		if err != nil {
			return nil, errors.NewEf(err, "could not findone")
		}

	} else if in.ResourceId != "" {

		rb, err = s.rbRepo.FindOne(
			ctx, repos.Filter{
				"resource_id": in.ResourceId,
			},
		)
		if err != nil {
			return nil, errors.NewEf(err, "could not findone")
		}

	} else {
		return nil, errors.NewEf(err, "no resourceId provided")
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

var Module = fx.Module(
	"application",
	fx.Provide(fxServer),
	repos.NewFxMongoRepo[*entities.RoleBinding]("role_bindings", "rb", entities.RoleBindingIndices),
	fx.Invoke(
		func(server *grpc.Server, iamService iam.IAMServer) {
			iam.RegisterIAMServer(server, iamService)
		},
	),
)
