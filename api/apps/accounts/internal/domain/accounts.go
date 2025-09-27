package domain

import (
	"context"
	"strings"

	fc "github.com/kloudlite/api/apps/accounts/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/accounts/internal/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findAccount(ctx context.Context, name string) (*entities.Account, error) {
	result, err := d.accountRepo.FindOne(ctx, repos.Filter{
		fields.MetadataName:      name,
		fields.MarkedForDeletion: repos.Filter{"$ne": true},
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if result == nil {
		return nil, errors.Newf("account with name %q not found", name)
	}

	return result, nil
}

func (d *domain) ListAccounts(ctx UserContext) ([]*entities.Account, error) {
	out, err := d.iamClient.ListMembershipsForUser(ctx, &iam.MembershipsForUserIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceAccount),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	accountNames := make([]string, len(out.RoleBindings))
	for i := range out.RoleBindings {
		accountNames[i] = strings.Split(out.RoleBindings[i].ResourceRef, "/")[0]
	}

	return d.accountRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		fields.MetadataName:      repos.Filter{"$in": accountNames},
		fields.MarkedForDeletion: repos.Filter{"$ne": true},
	}})
}

func (d *domain) GetAccount(ctx UserContext, name string) (*entities.Account, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.GetAccount); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findAccount(ctx, name)
}


func (d *domain) CreateAccount(ctx UserContext, account entities.Account) (*entities.Account, error) {
	account.TargetNamespace = constants.GetAccountTargetNamespace(account.Name)
	account.IsActive = fn.New(true)
	account.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	account.LastUpdatedBy = account.CreatedBy

	acc, err := d.accountRepo.Create(ctx, &account)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.addMembership(ctx, acc.Name, ctx.UserId, iamT.RoleAccountOwner); err != nil {
		return nil, errors.NewE(err)
	}

	return acc, nil
}

func (d *domain) UpdateAccount(ctx UserContext, accountIn entities.Account) (*entities.Account, error) {
	if err := d.checkAccountAccess(ctx, accountIn.Name, ctx.UserId, iamT.UpdateAccount); err != nil {
		return nil, errors.NewE(err)
	}

	account, err := d.findAccount(ctx, accountIn.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if account.IsActive != nil && !*account.IsActive {
		return nil, errors.Newf("accountIn %q is not active, could not update", accountIn.Name)
	}

	if account.IsMarkedForDeletion() {
		return nil, errors.Newf("accountIn %q is marked for deletion, could not update", accountIn.Name)
	}

	uAcc, err := d.accountRepo.PatchById(ctx, account.Id, repos.Document{
		fields.MetadataLabels:  accountIn.Labels,
		fields.DisplayName:     accountIn.DisplayName,
		fc.AccountLogo:         accountIn.Logo,
		fc.AccountContactEmail: accountIn.ContactEmail,
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	return uAcc, nil
}

func (d *domain) DeleteAccount(ctx UserContext, name string) (bool, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.DeleteAccount); err != nil {
		return false, errors.NewE(err)
	}

	if _, err := d.accountRepo.Patch(
		ctx,
		repos.Filter{
			fields.MetadataName:      name,
			fields.MarkedForDeletion: repos.Filter{"$ne": true},
		},
		repos.Document{
			fields.MarkedForDeletion: fn.New(true),
			fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
		}); err != nil {
		return false, errors.NewE(err)
	}

	return true, nil
}
func (d *domain) ActivateAccount(ctx UserContext, name string) (bool, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.ActivateAccount); err != nil {
		return false, errors.NewE(err)
	}

	account, err := d.findAccount(ctx, name)
	if err != nil {
		return false, errors.NewE(err)
	}

	if account.IsActive != nil && *account.IsActive {
		return false, errors.Newf("account %q is already active", name)
	}

	if _, err := d.accountRepo.PatchById(ctx, account.Id, repos.Document{
		fc.AccountIsActive: fn.New(true),
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}); err != nil {
		return false, errors.NewE(err)
	}

	return true, nil
}

func (d *domain) DeactivateAccount(ctx UserContext, name string) (bool, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.DeactivateAccount); err != nil {
		return false, errors.NewE(err)
	}

	account, err := d.findAccount(ctx, name)
	if err != nil {
		return false, errors.NewE(err)
	}

	if account.IsActive != nil && !*account.IsActive {
		return false, errors.Newf("account %q is already deactive", name)
	}

	if _, err := d.accountRepo.PatchById(ctx, account.Id, repos.Document{
		fc.AccountIsActive: fn.New(false),
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}); err != nil {
		return false, errors.NewE(err)
	}

	return true, nil
}

type AvailableKloudliteRegion struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// AvailableKloudliteRegions implements Domain.
func (d *domain) AvailableKloudliteRegions(ctx UserContext) ([]*AvailableKloudliteRegion, error) {
	regions := make([]*AvailableKloudliteRegion, 0, len(d.Env.AvailableKloudliteRegions))
	for i := range d.Env.AvailableKloudliteRegions {
		regions = append(regions, &AvailableKloudliteRegion{
			ID:          d.Env.AvailableKloudliteRegions[i].ID,
			DisplayName: d.Env.AvailableKloudliteRegions[i].DisplayName,
		})
	}

	return regions, nil
}
