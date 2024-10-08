package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/iam/internal/entities"
	"github.com/kloudlite/api/pkg/grpc"

	t "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
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
		return nil, errors.NewE(err)
	}
	if rb == nil {
		return nil, errors.Newf("role binding for (userId=%q,  ResourceRef=%q, ResourceType=%q) not found", in.UserId, in.ResourceRef, in.ResourceType)
	}

	rb.Role = t.Role(in.Role)

	if _, err = s.rbRepo.UpdateById(ctx, rb.Id, rb); err != nil {
		return nil, errors.NewE(err)
	}

	return &iam.UpdateMembershipOut{
		Result: true,
	}, nil
}

var ErrRoleBindingNotFound error = fmt.Errorf("role binding not found")

func (s *GrpcService) findRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) (*entities.RoleBinding, error) {
	rb, err := s.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":      userId,
			"resource_ref": resourceRef,
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if rb == nil {
		return nil, ErrRoleBindingNotFound
	}
	return rb, nil
}

func (s *GrpcService) GetMembership(ctx context.Context, in *iam.GetMembershipIn) (*iam.GetMembershipOut, error) {
	rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceRef)
	if err != nil {
		return nil, errors.NewE(err)
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
	if strings.HasPrefix(in.UserId, "sys-user") {
		return &iam.CanOut{Status: true}, nil
	}

	rb, ok := s.roleBindingMap[t.Action(in.Action)]
	if !ok {
		return &iam.CanOut{Status: false}, nil
	}

	var hasAccountMemberRole bool

	canFilter := repos.Filter{
		"resource_ref": map[string]any{"$in": in.ResourceRefs},
		"user_id":      in.UserId,
	}

	for i := range rb {
		if rb[i] == t.RoleAccountMember {
			hasAccountMemberRole = true

			rr := make([]map[string]any, 0, len(in.ResourceRefs))

			for i := range in.ResourceRefs {
				accountName, _, _, err := t.ParseResourceRef(in.ResourceRefs[i])
				if err != nil {
					return nil, err
				}

				if strings.TrimSpace(accountName) == "" {
					return nil, fmt.Errorf("accountName must be provided")
				}

				nf := s.rbRepo.MergeMatchFilters(repos.Filter{}, map[string]repos.MatchFilter{
					"resource_ref": {
						MatchType: repos.MatchTypeRegex,
						Regex:     fn.New(t.NewResourceRef(accountName, "*", "*")),
					},
				})
				rr = append(rr, map[string]any{"resource_ref": nf["resource_ref"]})
			}

			delete(canFilter, "resource_ref")
			canFilter["$or"] = rr
		}
	}

	rbs, err := s.rbRepo.Find(
		ctx, repos.Query{Filter: canFilter},
	)
	if err != nil {
		return nil, errors.NewEf(err, "could not find rolebindings for (resourceRefs=%s)", strings.Join(in.ResourceRefs, ","))
	}

	if rbs == nil {
		return nil, errors.Newf("no rolebinding found for (userId=%s, resourceRefs=%s)", in.UserId, strings.Join(in.ResourceRefs, ","))
	}

	if hasAccountMemberRole && len(rbs) > 0 {
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
	s.logger.Debugf("received request for membership deletion for userid=%s, resourceRef=%s", in.UserId, in.ResourceRef)
	if in.UserId == "" || in.ResourceRef == "" {
		return nil, errors.Newf("userId or resourceRef is empty, rejecting")
	}

	rb, err := s.findRoleBinding(ctx, repos.ID(in.UserId), in.ResourceRef)
	if err != nil {
		if errors.Is(err, ErrRoleBindingNotFound) {
			s.logger.WithKV("userID", in.UserId, "resourceRef", in.ResourceRef).Infof("role binding might already have been deleted")
			return &iam.RemoveMembershipOut{Result: true}, nil
		}
		return nil, errors.NewE(err)
	}

	if err := s.rbRepo.DeleteById(ctx, rb.Id); err != nil {
		return nil, errors.NewEf(err, "could not delete resource for (userId=%s, resourceRef=%s)", in.UserId, in.ResourceRef)
	}
	s.logger.Debugf("removed user (%s) membership resourceRef=%s", in.UserId, in.ResourceRef)

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
