package app

import (
	"context"
	"fmt"
	"kloudlite.io/apps/iam/internal/entities"
	"kloudlite.io/pkg/grpc"
	"strings"

	t "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

type IAMGrpcServer grpc.Server

type GrpcService struct {
	iam.UnimplementedIAMServer
	logger         logging.Logger
	rbRepo         repos.DbRepo[*entities.RoleBinding]
	roleBindingMap map[t.Action][]t.Role
}

// UpdateMembership updates only the role for a user on already specified resource_ref
func (s *GrpcService) UpdateMembership(ctx context.Context, in *iam.UpdateMembershipIn) (*iam.UpdateMembershipOut, error) {
	rb, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":       in.UserId,
			"resource_ref":  in.ResourceRef,
			"resource_type": in.ResourceType,
		},
	)
	if err != nil {
		return nil, err
	}
	if rb == nil {
		return nil, fmt.Errorf("role binding for (userId=%q,  ResourceRef=%q, ResourceType=%q) not found", in.UserId, in.ResourceRef, in.ResourceType)
	}

	rb.Role = t.Role(in.Role)

	if _, err = s.rbRepo.UpdateById(ctx, rb.Id, rb); err != nil {
		return nil, err
	}

	return &iam.UpdateMembershipOut{
		Result: true,
	}, nil
}

func (s *GrpcService) findRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) (*entities.RoleBinding, error) {
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

//func (s *GrpcService) ConfirmMembership(ctx context.Context, in *iam.ConfirmMembershipIn) (*iam.ConfirmMembershipOut, error) {
//	//rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceRef)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//if t.Role(in.Role) != rb.Role {
//	//	return nil, errors.New("The invitation has been updated")
//	//}
//	//
//	//_, err = s.rbRepo.UpdateById(ctx, rb.Id, rb)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//return &iam.ConfirmMembershipOut{}, nil
//	return nil, fmt.Errorf("not implemented")
//}

//func (s *GrpcService) InviteMembership(ctx context.Context, in *iam.AddMembershipIn) (*iam.AddMembershipOut, error) {
//	rb, err := s.rbRepo.FindOne(
//		ctx, repos.Filter{
//			"user_id":      in.UserId,
//			"resource_ref": in.ResourceRef,
//		},
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	if rb != nil {
//		if string(rb.Role) == in.Role {
//			s.logger.Infof("user %s already has role %s on resource %s", in.UserId, in.Role, in.ResourceRef)
//			return &iam.AddMembershipOut{Result: true}, nil
//		}
//		rb.Role = t.Role(in.Role)
//		//rb.Accepted = false
//		_, err := s.rbRepo.UpdateById(ctx, rb.Id, rb)
//		if err != nil {
//			return nil, err
//		}
//		return &iam.AddMembershipOut{Result: true}, nil
//	}
//
//	_, err = s.rbRepo.Create(
//		ctx, &entities.RoleBinding{
//			UserId:       in.UserId,
//			ResourceType: t.ResourceType(in.ResourceType),
//			ResourceRef:  in.ResourceRef,
//			Role:         t.Role(in.Role),
//			//Accepted:     false,
//		},
//	)
//	if err != nil {
//		return nil, errors.NewEf(err, "could not create rolebinding")
//	}
//	return &iam.AddMembershipOut{Result: true}, nil
//}

func (s *GrpcService) GetMembership(ctx context.Context, in *iam.GetMembershipIn) (*iam.GetMembershipOut, error) {
	rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceRef)
	if err != nil {
		return nil, err
	}
	return &iam.GetMembershipOut{
		UserId:      rb.UserId,
		ResourceRef: rb.ResourceRef,
		Role:        string(rb.Role),
	}, nil
}

func (s *GrpcService) ListMembershipsForResource(ctx context.Context, in *iam.MembershipsForResourceIn) (*iam.ListMembershipsOut, error) {
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

func (s *GrpcService) Can(ctx context.Context, in *iam.CanIn) (*iam.CanOut, error) {
	rbs, err := s.rbRepo.Find(
		ctx, repos.Query{Filter: repos.Filter{
			"resource_ref": map[string]any{"$in": in.ResourceRefs},
			"user_id":      in.UserId,
		}},
	)

	if err != nil {
		return nil, errors.NewEf(err, "could not find rolebindings for (resourceRefs=%s)", strings.Join(in.ResourceRefs, ","))
	}

	if rbs == nil {
		return nil, fmt.Errorf("no rolebinding found for (userId=%s, resourceRefs=%s)", in.UserId, strings.Join(in.ResourceRefs, ","))
	}

	if strings.HasPrefix(in.UserId, "sys-user") {
		return &iam.CanOut{Status: true}, nil
	}

	for i := range rbs {
		// 2nd loop, but very small length (always < #roles), so it's not exactly O(n^2), much like XO(n)
		for _, role := range s.roleBindingMap[t.Action(in.Action)] {
			if role == rbs[i].Role {
				return &iam.CanOut{Status: true}, nil
			}
		}
	}

	return &iam.CanOut{Status: false}, nil
}

func (s *GrpcService) ListMembershipsForUser(ctx context.Context, in *iam.MembershipsForUserIn) (*iam.ListMembershipsOut, error) {
	filter := repos.Filter{"user_id": in.UserId}
	if in.ResourceType != "" {
		filter["resource_type"] = in.ResourceType
	}

	rbs, err := s.rbRepo.Find(ctx, repos.Query{Filter: filter})
	if err != nil {
		return nil, errors.NewEf(err, "could not find memberships by (userId=%q)", in.UserId)
	}

	result := make([]*iam.RoleBinding, 0, len(rbs))
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

func (s *GrpcService) AddMembership(ctx context.Context, in *iam.AddMembershipIn) (*iam.AddMembershipOut, error) {
	_, err := s.rbRepo.Create(
		ctx, &entities.RoleBinding{
			UserId:       in.UserId,
			ResourceType: t.ResourceType(in.ResourceType),
			ResourceRef:  in.ResourceRef,
			Role:         t.Role(in.Role),
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not create rolebinding")
	}
	return &iam.AddMembershipOut{Result: true}, nil
}

func (s *GrpcService) RemoveMembership(ctx context.Context, in *iam.RemoveMembershipIn) (*iam.RemoveMembershipOut, error) {
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

func (s *GrpcService) RemoveResource(ctx context.Context, in *iam.RemoveResourceIn) (*iam.RemoveResourceOut, error) {
	if err := s.rbRepo.DeleteMany(ctx, repos.Filter{"resource_ref": in.ResourceRef}); err != nil {
		return nil, errors.NewEf(err, "could not delete resources for (resourceRef=%s)", in.ResourceRef)
	}
	return &iam.RemoveResourceOut{Result: true}, nil
}

func (s *GrpcService) Ping(ctx context.Context, in *iam.Message) (*iam.Message, error) {
	return &iam.Message{
		Message: "ping",
	}, nil
}

func newIAMGrpcService(logger logging.Logger, rbRepo repos.DbRepo[*entities.RoleBinding], roleBindingMap map[t.Action][]t.Role) iam.IAMServer {
	return &GrpcService{
		logger:         logger,
		rbRepo:         rbRepo,
		roleBindingMap: roleBindingMap,
	}
}
