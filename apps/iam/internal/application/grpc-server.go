package application

import (
	context "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"kloudlite.io/apps/iam/internal/domain/entities"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type GrpcServer struct {
	// iam.UnimplementedIAMServer
	rbRepo repos.DbRepo[*entities.RoleBinding]
}

func (s *GrpcServer) findRoleBinding(ctx context.Context, userId repos.ID, resourceId repos.ID) (*entities.RoleBinding, error) {
	rb, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":     userId,
			"resource_id": resourceId,
		},
	)
	if err != nil {
		return nil, err
	}
	if rb == nil {
		return nil, fmt.Errorf("role binding for (userId=%s,  resourceId=%s) not found", userId, resourceId)
	}
	return rb, nil
}

func (s *GrpcServer) ConfirmMembership(ctx context.Context, in *iam.ConfirmMembershipIn) (*iam.ConfirmMembershipOut, error) {
	rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceId)

	if in.Role != rb.Role {
		return nil, errors.New("The invitation has been updated")
	}

	rb.Accepted = true
	_, err = s.rbRepo.UpdateById(ctx, rb.Id, rb)
	if err != nil {
		return nil, err
	}
	return &iam.ConfirmMembershipOut{}, nil
}

func (s *GrpcServer) InviteMembership(ctx context.Context, in *iam.AddMembershipIn) (*iam.AddMembershipOut, error) {
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
		return &iam.AddMembershipOut{Result: true}, nil
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
	return &iam.AddMembershipOut{Result: true}, nil
}

func (s *GrpcServer) GetMembership(ctx context.Context, membership *iam.GetMembershipIn) (*iam.GetMembershipOut, error) {
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
	return &iam.GetMembershipOut{
		UserId:     one.UserId,
		ResourceId: one.ResourceId,
		Role:       one.Role,
		Accepted:   one.Accepted,
	}, nil
}

func (s *GrpcServer) ListResourceMemberships(ctx context.Context, in *iam.ResourceMembershipsIn) (*iam.ListMembershipsOut, error) {
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

	return &iam.ListMembershipsOut{
		RoleBindings: result,
	}, nil
}

func (s *GrpcServer) Can(ctx context.Context, in *iam.CanIn) (*iam.CanOut, error) {
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
		return &iam.CanOut{Status: false}, err
	}

	err = json.Unmarshal(file, &actionsConfig)

	if err != nil {
		return &iam.CanOut{Status: false}, err
	}

	if rb == nil {
		return &iam.CanOut{Status: false}, nil
	}

	if strings.HasPrefix(in.UserId, "sys-user") {
		return &iam.CanOut{Status: true}, nil
	}

	for _, v := range actionsConfig.Actions[in.Action] {
		fmt.Println(v, rb.Role)
		if v == rb.Role {
			return &iam.CanOut{Status: true}, nil
		}

	}

	for _, role := range constants.ActionMap[constants.Action(in.Action)] {
		if role == constants.Role(rb.Role) {
			return &iam.CanOut{Status: true}, nil
		}
	}

	return &iam.CanOut{Status: false}, nil
}

func (s *GrpcServer) ListUserMemberships(ctx context.Context, in *iam.UserMembershipsIn) (*iam.ListMembershipsOut, error) {
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

	return &iam.ListMembershipsOut{
		RoleBindings: result,
	}, nil
}

// Mutation

func (s *GrpcServer) AddMembership(ctx context.Context, in *iam.AddMembershipIn) (*iam.AddMembershipOut, error) {
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
	return &iam.AddMembershipOut{Result: true}, nil
}

func (s *GrpcServer) RemoveMembership(ctx context.Context, in *iam.RemoveMembershipIn) (*iam.RemoveMembershipOut, error) {
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

	return &iam.RemoveMembershipOut{Result: true}, nil
}

func (s *GrpcServer) RemoveResource(ctx context.Context, in *iam.RemoveResourceIn) (*iam.RemoveResourceOut, error) {
	err := s.rbRepo.DeleteMany(ctx, repos.Filter{"resource_id": in.ResourceId})
	if err != nil {
		return nil, errors.NewEf(err, "could not delete resources(id=%s)", in.ResourceId)
	}
	return &iam.RemoveResourceOut{Result: true}, nil
}

func (i *GrpcServer) Ping(ctx context.Context, in *iam.Message) (*iam.Message, error) {
	fmt.Println("", in.Message)
	return &iam.Message{
		Message: "asdfasdf",
	}, nil
}
