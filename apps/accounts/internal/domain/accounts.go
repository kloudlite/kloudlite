package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"kloudlite.io/apps/accounts/internal/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/yaml"
	"strings"
)

func (d *domain) findAccount(ctx context.Context, name string) (*entities.Account, error) {
	result, err := d.accountRepo.FindOne(ctx, repos.Filter{
		"metadata.name": name,
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("account with name %q not found", name)
	}

	return result, nil
}

func (d *domain) applyAccountOnCluster(ctx context.Context, account *entities.Account) error {
	b, err := json.Marshal(account.Account)
	if err != nil {
		return err
	}
	y, err := yaml.JSONToYAML(b)
	if err != nil {
		return err
	}

	if _, err := d.k8sYamlClient.ApplyYAML(ctx, y); err != nil {
		return err
	}

	return nil
}

func (d *domain) ListAccounts(ctx UserContext) ([]*entities.Account, error) {
	out, err := d.iamClient.ListMembershipsForUser(ctx, &iam.MembershipsForUserIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceAccount),
	})
	if err != nil {
		return nil, err
	}

	accountNames := make([]string, len(out.RoleBindings))
	for i := range out.RoleBindings {
		accountNames[i] = strings.Split(out.RoleBindings[i].ResourceRef, "/")[0]
	}

	return d.accountRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"metadata.name": map[string]any{"$in": accountNames},
	}})
}

func (d *domain) GetAccount(ctx UserContext, name string) (*entities.Account, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.GetAccount); err != nil {
		return nil, err
	}
	return d.findAccount(ctx, name)
}

func (d *domain) CreateAccount(ctx UserContext, account entities.Account) (*entities.Account, error) {
	account.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &account.Account); err != nil {
		return nil, err
	}

	account.IsActive = fn.New(true)
	account.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	account.LastUpdatedBy = account.CreatedBy

	acc, err := d.accountRepo.Create(ctx, &account)
	if err != nil {
		return nil, err
	}

	if err := d.addMembership(ctx, acc.Name, ctx.UserId, iamT.RoleAccountOwner); err != nil {
		return nil, err
	}

	if err := d.applyAccountOnCluster(ctx, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (d *domain) UpdateAccount(ctx UserContext, account entities.Account) (*entities.Account, error) {
	if err := d.checkAccountAccess(ctx, account.Name, ctx.UserId, iamT.UpdateAccount); err != nil {
		return nil, err
	}

	account.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &account.Account); err != nil {
		return nil, err
	}

	acc, err := d.findAccount(ctx, account.Name)
	if err != nil {
		return nil, err
	}

	if acc.IsActive != nil && !*acc.IsActive {
		return nil, fmt.Errorf("account %q is not active, could not update", account.Name)
	}

	if acc.IsMarkedForDeletion() {
		return nil, fmt.Errorf("account %q is marked for deletion, could not update", account.Name)
	}

	acc.Labels = account.Labels
	acc.IsActive = account.IsActive
	acc.DisplayName = account.DisplayName

	acc.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	uAcc, err := d.accountRepo.UpdateById(ctx, acc.Id, acc)
	if err != nil {
		return nil, err
	}

	if err := d.applyAccountOnCluster(ctx, uAcc); err != nil {
		return nil, err
	}
	return uAcc, nil
}

func (d *domain) DeleteAccount(ctx UserContext, name string) (bool, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.DeleteAccount); err != nil {
		return false, err
	}

	account, err := d.findAccount(ctx, name)
	if err != nil {
		return false, err
	}

	account.MarkedForDeletion = fn.New(true)
	if _, err := d.accountRepo.UpdateById(ctx, account.Id, account); err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) ResyncAccount(ctx UserContext, name string) error {
	//TODO implement me
	panic("implement me")
}

func (d *domain) ActivateAccount(ctx UserContext, name string) (bool, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.ActivateAccount); err != nil {
		return false, err
	}

	account, err := d.findAccount(ctx, name)
	if err != nil {
		return false, err
	}

	if account.IsActive != nil && *account.IsActive {
		return false, fmt.Errorf("account %q is already active", name)
	}

	account.IsActive = fn.New(true)

	if _, err := d.accountRepo.UpdateById(ctx, account.Id, account); err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) DeactivateAccount(ctx UserContext, name string) (bool, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.DeactivateAccount); err != nil {
		return false, err
	}

	account, err := d.findAccount(ctx, name)
	if err != nil {
		return false, err
	}

	if account.IsActive != nil && !*account.IsActive {
		return false, fmt.Errorf("account %q is already deactive", name)
	}

	account.IsActive = fn.New(false)

	if _, err := d.accountRepo.UpdateById(ctx, account.Id, account); err != nil {
		return false, err
	}

	return true, nil
}
