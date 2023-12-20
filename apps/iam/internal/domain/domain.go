package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/iam/internal/entities"
	t "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

type Domain interface {
	AddRoleBinding(ctx context.Context, rb entities.RoleBinding) (*entities.RoleBinding, error)
	RemoveRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) error
	RemoveRoleBindingsForResource(ctx context.Context, resourceRef string) error
	UpdateRoleBinding(ctx context.Context, rb entities.RoleBinding) (*entities.RoleBinding, error)

	GetRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) (*entities.RoleBinding, error)

	ListRoleBindingsForResource(ctx context.Context, resourceType t.ResourceType, resourceRef string) ([]*entities.RoleBinding, error)
	ListRoleBindingsForUser(ctx context.Context, userId repos.ID, resourceType *t.ResourceType) ([]*entities.RoleBinding, error)

	Can(ctx context.Context, userId repos.ID, resourceRefs []string, action t.Action) (bool, error)
}

type domain struct {
	rbRepo         repos.DbRepo[*entities.RoleBinding]
	roleBindingMap map[t.Action][]t.Role
}

func (d domain) AddRoleBinding(ctx context.Context, rb entities.RoleBinding) (*entities.RoleBinding, error) {
	if err := rb.Validate(); err != nil {
		return nil, err
	}

	exists, err := d.rbRepo.FindOne(ctx, repos.Filter{
		"user_id":      rb.UserId,
		"resource_ref": rb.ResourceRef,
	})
	if err != nil {
		return nil, err
	}

	if exists != nil {
		return nil, errors.Newf("role binding for (userId=%s, ResourceRef=%s) already exists", rb.UserId, rb.ResourceRef)
	}

	nrb, err := d.rbRepo.Create(ctx, &rb)
	if err != nil {
		return nil, err
	}
	return nrb, nil
}

func (s domain) findRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) (*entities.RoleBinding, error) {
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
		return nil, errors.Newf("role binding for (userId=%s,  ResourceRef=%s) not found", userId, resourceRef)
	}

	return rb, nil
}

func (d domain) RemoveRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) error {
	if userId == "" || resourceRef == "" {
		return errors.Newf("userId or resourceRef is empty, rejecting")
	}

	rb, err := d.findRoleBinding(ctx, userId, resourceRef)
	if err != nil {
		return err
	}

	if err := d.rbRepo.DeleteById(ctx, rb.Id); err != nil {
		return errors.NewEf(err, "could not delete resource for (userId=%s, resourceRef=%s)", userId, resourceRef)
	}

	return nil
}

func (d domain) RemoveRoleBindingsForResource(ctx context.Context, resourceRef string) error {
	if err := d.rbRepo.DeleteMany(ctx, repos.Filter{"resource_ref": resourceRef}); err != nil {
		return errors.NewEf(err, "could not delete role bindings for (resourceRef=%s)", resourceRef)
	}
	return nil
}

// UpdateRoleBinding updates only the role for a user on an already specified resource_ref
func (d domain) UpdateRoleBinding(ctx context.Context, rb entities.RoleBinding) (*entities.RoleBinding, error) {
	currRb, err := d.rbRepo.FindOne(
		ctx, repos.Filter{
			"user_id":       rb.UserId,
			"resource_ref":  rb.ResourceRef,
			"resource_type": rb.ResourceType,
		},
	)
	if err != nil {
		return nil, err
	}
	if currRb == nil {
		return nil, errors.Newf("role binding for (userId=%q,  ResourceRef=%q, ResourceType=%q) not found", rb.UserId, rb.ResourceRef, rb.ResourceType)
	}

	currRb.Role = rb.Role
	return d.rbRepo.UpdateById(ctx, currRb.Id, currRb)
}

func (d domain) GetRoleBinding(ctx context.Context, userId repos.ID, resourceRef string) (*entities.RoleBinding, error) {
	return d.findRoleBinding(ctx, userId, resourceRef)
}

func (d domain) ListRoleBindingsForResource(ctx context.Context, resourceType t.ResourceType, resourceRef string) ([]*entities.RoleBinding, error) {
	filter := repos.Filter{
		"resource_type": resourceType,
		"resource_ref":  resourceRef,
	}

	return d.rbRepo.Find(ctx, repos.Query{Filter: filter})
}

func (d domain) ListRoleBindingsForUser(ctx context.Context, userId repos.ID, resourceType *t.ResourceType) ([]*entities.RoleBinding, error) {
	filter := repos.Filter{
		"user_id": userId,
	}

	if resourceType != nil {
		filter["resource_type"] = *resourceType
	}

	return d.rbRepo.Find(ctx, repos.Query{Filter: filter})
}

func (d domain) Can(ctx context.Context, userId repos.ID, resourceRefs []string, action t.Action) (bool, error) {
	if d.roleBindingMap == nil {
		return false, UnAuthorizedError{debugMsg: "action-role-binding map is empty"}
	}

	if strings.HasPrefix(string(userId), "sys-user") {
		return true, nil
	}

	rbs, err := d.rbRepo.Find(
		ctx, repos.Query{Filter: repos.Filter{
			"resource_ref": map[string]any{"$in": resourceRefs},
			"user_id":      userId,
		}},
	)

	if err != nil {
		return false, UnAuthorizedError{debugMsg: "db repository find() call error", parentErr: err}
	}

	if rbs == nil {
		return false, UnAuthorizedError{debugMsg: fmt.Sprintf("no rolebinding found for (userId=%s, resourceRefs=%s)", userId, strings.Join(resourceRefs, ","))}
	}

	for i := range rbs {
		// 2nd loop, but very small length (always < #roles), so it's not exactly O(n^2), much like XO(n)
		for _, role := range d.roleBindingMap[action] {
			if role == rbs[i].Role {
				return true, nil
			}
		}
	}

	return false, UnAuthorizedError{debugMsg: fmt.Sprintf("no role bindings allow user %q to perform action %q on resource %q", userId, action, resourceRefs)}
}

func NewDomain(rbRepo repos.DbRepo[*entities.RoleBinding], roleBindingMap map[t.Action][]t.Role) Domain {
	return &domain{
		rbRepo:         rbRepo,
		roleBindingMap: roleBindingMap,
	}
}
