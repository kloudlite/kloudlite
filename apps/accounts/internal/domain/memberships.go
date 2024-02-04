package domain

import (
	"context"
	"github.com/kloudlite/api/apps/accounts/internal/entities"
	"strings"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) addMembership(ctx context.Context, accountName string, userId repos.ID, role iamT.Role) error {
	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(userId),
		ResourceType: string(iamT.ResourceAccount),
		ResourceRef:  iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		Role:         string(role),
	}); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) RemoveAccountMembership(ctx UserContext, accountName string, memberId repos.ID) (bool, error) {
	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, iamT.RemoveAccountMembership); err != nil {
		return false, errors.NewE(err)
	}

	account, err := d.findAccount(ctx, accountName)
	if err != nil {
		return false, errors.NewE(err)
	}

	if (account.IsActive != nil && !*account.IsActive) || account.IsMarkedForDeletion() {
		return false, errors.Newf("account %q is not active, or marked for deletion, aborting request", accountName)
	}

	out, err := d.iamClient.RemoveMembership(ctx, &iam.RemoveMembershipIn{
		UserId:      string(memberId),
		ResourceRef: iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
	})
	if err != nil {
		return false, errors.NewE(err)
	}

	return out.Result, nil
}

func (d *domain) UpdateAccountMembership(ctx UserContext, accountName string, memberId repos.ID, role iamT.Role) (bool, error) {
	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, iamT.UpdateAccountMembership); err != nil {
		return false, errors.NewE(err)
	}

	account, err := d.findAccount(ctx, accountName)
	if err != nil {
		return false, errors.NewE(err)
	}

	if (account.IsActive != nil && !*account.IsActive) || account.IsMarkedForDeletion() {
		return false, errors.Newf("account %q is not active, or marked for deletion, aborting request", accountName)
	}

	out, err := d.iamClient.UpdateMembership(ctx, &iam.UpdateMembershipIn{
		UserId:       string(memberId),
		ResourceType: string(iamT.ResourceAccount),
		ResourceRef:  iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		Role:         string(role),
	})

	if err != nil {
		return false, errors.NewE(err)
	}

	return out.Result, nil
}

func (d *domain) ListMembershipsForAccount(ctx UserContext, accountName string, role *iamT.Role) ([]*entities.AccountMembership, error) {
	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, iamT.ListMembershipsForAccount); err != nil {
		return nil, errors.NewE(err)
	}

	out, err := d.iamClient.ListMembershipsForResource(ctx, &iam.MembershipsForResourceIn{
		ResourceType: string(iamT.ResourceAccount),
		ResourceRef:  iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	memberships := make([]*entities.AccountMembership, len(out.RoleBindings))
	for i := range out.RoleBindings {
		memberships[i] = &entities.AccountMembership{
			AccountName: accountName,
			UserId:      repos.ID(out.RoleBindings[i].UserId),
			Role:        iamT.Role(out.RoleBindings[i].Role),
		}
	}

	return memberships, nil
}

func (d *domain) ListMembershipsForUser(ctx UserContext) ([]*entities.AccountMembership, error) {
	out, err := d.iamClient.ListMembershipsForUser(ctx, &iam.MembershipsForUserIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceAccount),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	memberships := make([]*entities.AccountMembership, len(out.RoleBindings))
	for i := range out.RoleBindings {
		memberships[i] = &entities.AccountMembership{
			AccountName: strings.Split(out.RoleBindings[i].ResourceRef, "/")[0],
			UserId:      repos.ID(out.RoleBindings[i].UserId),
			Role:        iamT.Role(out.RoleBindings[i].Role),
		}
	}

	return memberships, nil
}

func (d *domain) GetAccountMembership(ctx UserContext, accountName string) (*entities.AccountMembership, error) {
	membership, err := d.iamClient.GetMembership(
		ctx, &iam.GetMembershipIn{
			UserId:       string(ctx.UserId),
			ResourceType: string(iamT.ResourceAccount),
			ResourceRef:  iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &entities.AccountMembership{
		AccountName: accountName,
		UserId:      repos.ID(membership.UserId),
		Role:        iamT.Role(membership.Role),
	}, nil
}
