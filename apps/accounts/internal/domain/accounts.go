package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/api/pkg/errors"
	"strings"

	"github.com/kloudlite/api/apps/accounts/internal/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func (d *domain) findAccount(ctx context.Context, name string) (*entities.Account, error) {
	result, err := d.accountRepo.FindOne(ctx, repos.Filter{
		"metadata.name": name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if result == nil {
		return nil, errors.Newf("account with name %q not found", name)
	}

	return result, nil
}

func (d *domain) applyAccountOnCluster(ctx context.Context, account *entities.Account) error {
	b, err := json.Marshal(account.Account)
	if err != nil {
		return errors.NewE(err)
	}
	y, err := yaml.JSONToYAML(b)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.k8sClient.ApplyYAML(ctx, y); err != nil {
		return errors.NewE(err)
	}

	return nil
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
		"metadata.name": map[string]any{"$in": accountNames},
	}})
}

func (d *domain) GetAccount(ctx UserContext, name string) (*entities.Account, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.GetAccount); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findAccount(ctx, name)
}

func (d *domain) ensureNamespaceForAccount(ctx context.Context, accountName string, targetNamespace string) error {
	if err := d.k8sClient.Get(ctx, fn.NN("", targetNamespace), &corev1.Namespace{}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return errors.NewE(err)
		}
	}

	if err := d.k8sClient.Create(ctx, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNamespace,
			Labels: map[string]string{
				constants.AccountNameKey: accountName,
			},
		},
	}); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) CreateAccount(ctx UserContext, account entities.Account) (*entities.Account, error) {
	account.EnsureGVK()

	if account.Spec.TargetNamespace == nil {
		account.Spec.TargetNamespace = fn.New(fmt.Sprintf("kl-account-%s", account.Name))
	}

	if err := d.k8sClient.ValidateObject(ctx, &account.Account); err != nil {
		return nil, errors.NewE(err)
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
		return nil, errors.NewE(err)
	}

	if err := d.addMembership(ctx, acc.Name, ctx.UserId, iamT.RoleAccountOwner); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.ensureNamespaceForAccount(ctx, account.Name, *account.Spec.TargetNamespace); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyAccountOnCluster(ctx, acc); err != nil {
		return nil, errors.NewE(err)
	}
	return acc, nil
}

func (d *domain) UpdateAccount(ctx UserContext, accountIn entities.Account) (*entities.Account, error) {
	if err := d.checkAccountAccess(ctx, accountIn.Name, ctx.UserId, iamT.UpdateAccount); err != nil {
		return nil, errors.NewE(err)
	}

	accountIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &accountIn.Account); err != nil {
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
		"labels":      accountIn.Labels,
		"isActive":    accountIn.IsActive,
		"displayName": accountIn.DisplayName,
		"logo":        accountIn.Logo,
		"contactEmail": accountIn.ContactEmail,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})

	if err := d.applyAccountOnCluster(ctx, uAcc); err != nil {
		return nil, errors.NewE(err)
	}
	return uAcc, nil
}

func (d *domain) DeleteAccount(ctx UserContext, name string) (bool, error) {
	if err := d.checkAccountAccess(ctx, name, ctx.UserId, iamT.DeleteAccount); err != nil {
		return false, errors.NewE(err)
	}

	account, err := d.findAccount(ctx, name)
	if err != nil {
		return false, errors.NewE(err)
	}

	if _, err := d.accountRepo.PatchById(ctx, account.Id, repos.Document{
		"markedForDeletion": fn.New(true),
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}); err != nil {
		return false, errors.NewE(err)
	}

	return true, nil
}

func (d *domain) ResyncAccount(ctx UserContext, name string) error {
	acc, err := d.findAccount(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.ensureNamespaceForAccount(ctx, acc.Name, *acc.Spec.TargetNamespace); err != nil {
		return errors.NewE(err)
	}

	if err := d.applyAccountOnCluster(ctx, acc); err != nil {
		return errors.NewE(err)
	}

	return nil
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
		"isActive": fn.New(true),
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
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
		"isActive": fn.New(false),
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}); err != nil {
		return false, errors.NewE(err)
	}

	return true, nil
}
