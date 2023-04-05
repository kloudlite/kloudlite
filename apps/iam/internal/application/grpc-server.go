package application

import (
	context "context"
	"fmt"
	"strings"

	"kloudlite.io/apps/iam/internal/domain/entities"
	t "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

type GrpcServer struct {
	iam.UnimplementedIAMServer
	logger         logging.Logger
	rbRepo         repos.DbRepo[*entities.RoleBinding]
	roleBindingMap map[t.Action][]t.Role
}

func (s *GrpcServer) findRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) (*entities.RoleBinding, error) {
	rb, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":      userId,
			"resource_ref": resourceRef,
		},
	)
	if err != nil {
		return nil, err
	}
	if rb == nil {
		return nil, fmt.Errorf("role binding for (userId=%s,  ResourceRef=%s) not found", userId, resourceRef)
	}
	return rb, nil
}

func (s *GrpcServer) ConfirmMembership(ctx context.Context, in *iam.ConfirmMembershipIn) (*iam.ConfirmMembershipOut, error) {
	rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceRef)
	if err != nil {
		return nil, err
	}

	if t.Role(in.Role) != rb.Role {
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
	rb, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":      in.UserId,
			"resource_ref": in.ResourceRef,
		},
	)
	if err != nil {
		return nil, err
	}

	if rb != nil {
		if string(rb.Role) == in.Role {
			s.logger.Infof("user %s already has role %s on resource %s", in.UserId, in.Role, in.ResourceRef)
			return &iam.AddMembershipOut{Result: true}, nil
		}
		rb.Role = t.Role(in.Role)
		rb.Accepted = false
		_, err := s.rbRepo.UpdateById(ctx, rb.Id, rb)
		if err != nil {
			return nil, err
		}
		return &iam.AddMembershipOut{Result: true}, nil
	}

	_, err = s.rbRepo.Create(
		ctx, &entities.RoleBinding{
			UserId:       in.UserId,
			ResourceType: t.ResourceType(in.ResourceType),
			ResourceRef:  in.ResourceRef,
			Role:         t.Role(in.Role),
			Accepted:     false,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not create rolebinding")
	}
	return &iam.AddMembershipOut{Result: true}, nil
}

func (s *GrpcServer) GetMembership(ctx context.Context, in *iam.GetMembershipIn) (*iam.GetMembershipOut, error) {
	rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceRef)
	if err != nil {
		return nil, err
	}
	return &iam.GetMembershipOut{
		UserId:      rb.UserId,
		ResourceRef: rb.ResourceRef,
		Role:        string(rb.Role),
		Accepted:    rb.Accepted,
	}, nil
}

func (s *GrpcServer) ListResourceMemberships(ctx context.Context, in *iam.ResourceMembershipsIn) (*iam.ListMembershipsOut, error) {
	filter := repos.Filter{}
	if in.ResourceRef != "" {
		filter["resource_ref"] = in.ResourceRef
	}
	if in.ResourceType != "" {
		filter["resource_type"] = in.ResourceType
	}

	rbs, err := s.rbRepo.Find(ctx, repos.Query{Filter: filter})
	if err != nil {
		return nil, errors.NewEf(err, "could not find memberships by (ResourceRef=%q, resourceType=%q)", in.ResourceRef, in.ResourceType)
	}

	var result []*iam.RoleBinding
	for _, rb := range rbs {
		result = append(
			result, &iam.RoleBinding{
				UserId:       rb.UserId,
				ResourceType: string(rb.ResourceType),
				ResourceRef:  rb.ResourceRef,
				Role:         string(rb.Role),
			},
		)
	}

	return &iam.ListMembershipsOut{
		RoleBindings: result,
	}, nil
}

func (s *GrpcServer) ListMembershipsByResource(ctx context.Context, in *iam.MembershipsByResourceIn) (*iam.ListMembershipsOut, error) {
	filter := repos.Filter{}
	if in.ResourceRef != "" {
		filter["resource_ref"] = in.ResourceRef
	}
	if in.ResourceType != "" {
		filter["resource_type"] = in.ResourceType
	}
	if in.Accepted != nil {
		filter["accepted"] = *in.Accepted
	}

	rbs, err := s.rbRepo.Find(ctx, repos.Query{Filter: filter})
	if err != nil {
		return nil, errors.NewEf(err, "could not find memberships by (ResourceRef=%q, resourceType=%q)", in.ResourceRef, in.ResourceType)
	}

	var result []*iam.RoleBinding
	for _, rb := range rbs {
		result = append(
			result, &iam.RoleBinding{
				UserId:       rb.UserId,
				ResourceType: string(rb.ResourceType),
				ResourceRef:  rb.ResourceRef,
				Role:         string(rb.Role),
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
			"resource_ref": map[string]any{"$in": in.ResourceRefs},
			"user_id":      in.UserId,
			"accepted":     true,
		},
	)

	if err != nil {
		return nil, errors.NewEf(err, "could not find rolebindings for (resourceRefs=%s)", strings.Join(in.ResourceRefs, ","))
	}

	if rb == nil {
		return nil, fmt.Errorf("no rolebinding found for (userId=%s, resourceRefs=%s)", in.UserId, strings.Join(in.ResourceRefs, ","))
	}

	if strings.HasPrefix(in.UserId, "sys-user") {
		return &iam.CanOut{Status: true}, nil
	}

	for _, role := range s.roleBindingMap[t.Action(in.Action)] {
		if role == rb.Role {
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
	for i := range rbs {
		result = append(
			result, &iam.RoleBinding{
				UserId:       rbs[i].UserId,
				ResourceType: string(rbs[i].ResourceType),
				ResourceRef:  rbs[i].ResourceRef,
				Role:         string(rbs[i].Role),
			},
		)
	}

	return &iam.ListMembershipsOut{
		RoleBindings: result,
	}, nil
}

func (s *GrpcServer) ListMembershipsForUser(ctx context.Context, in *iam.MembershipsForUserIn) (*iam.ListMembershipsOut, error) {
	filter := repos.Filter{"user_id": in.UserId}
	if in.ResourceType != "" {
		filter["resource_type"] = in.ResourceType
	}

	rbs, err := s.rbRepo.Find(ctx, repos.Query{Filter: filter})
	if err != nil {
		return nil, errors.NewEf(err, "could not find memberships by (userId=%q)", in.UserId)
	}

	result := []*iam.RoleBinding{}
	for i := range rbs {
		result = append(
			result, &iam.RoleBinding{
				UserId:       rbs[i].UserId,
				ResourceType: string(rbs[i].ResourceType),
				ResourceRef:  rbs[i].ResourceRef,
				Role:         string(rbs[i].Role),
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
			ResourceType: t.ResourceType(in.ResourceType),
			ResourceRef:  in.ResourceRef,
			Role:         t.Role(in.Role),
			Accepted:     true,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not create rolebinding")
	}
	return &iam.AddMembershipOut{Result: true}, nil
}

func (s *GrpcServer) RemoveMembership(ctx context.Context, in *iam.RemoveMembershipIn) (*iam.RemoveMembershipOut, error) {
	if in.UserId == "" || in.ResourceRef == "" {
		return nil, fmt.Errorf("userId or resourceRef is empty, rejecting")
	}

	rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceRef)
	if err != nil {
		return nil, err
	}

	if err := s.rbRepo.DeleteById(ctx, rb.Id); err != nil {
		return nil, errors.NewEf(err, "could not delete resource for (userId=%s, resourceRef=%s)", in.UserId, in.ResourceRef)
	}

	return &iam.RemoveMembershipOut{Result: true}, nil
}

func (s *GrpcServer) RemoveResource(ctx context.Context, in *iam.RemoveResourceIn) (*iam.RemoveResourceOut, error) {
	if err := s.rbRepo.DeleteMany(ctx, repos.Filter{"resource_ref": in.ResourceRef}); err != nil {
		return nil, errors.NewEf(err, "could not delete resources for (resourceRef=%s)", in.ResourceRef)
	}
	return &iam.RemoveResourceOut{Result: true}, nil
}

func (i *GrpcServer) Ping(ctx context.Context, in *iam.Message) (*iam.Message, error) {
	return &iam.Message{
		Message: "asdfasdf",
	}, nil
}
