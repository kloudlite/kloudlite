package domain

import (
	"context"
	"fmt"
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
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (d *domain) deleteNamespaceForAccount(ctx context.Context, targetNamespace string) error {
	// panic("not implemented. Yet to decide if we want to delete namespace when account is deleted")
	return fmt.Errorf("not supported yet")
}

func (d *domain) ensureKloudliteRegistryCredentials(ctx UserContext, account *entities.Account) error {
	credentialsName := "kloudlite-registry-pull-creds"

	secret := &corev1.Secret{}
	if err := d.k8sClient.Get(ctx, fn.NN(account.TargetNamespace, credentialsName), secret); err != nil {
		secret = nil
	}

	if secret != nil {
		d.logger.Infof("kloudlite registry image pull secret already exists")
		return nil
	}

	// out, err := d.containerRegistryClient.CreateReadOnlyCredential(ctx, &container_registry.CreateReadOnlyCredentialIn{
	// 	AccountName:      account.Name,
	// 	UserId:           string(ctx.UserId),
	// 	CredentialName:   credentialsName,
	// 	RegistryUsername: fmt.Sprintf("account_%s", account.Name),
	// })
	// if err != nil {
	// 	return err
	// }

	// if err := d.k8sClient.Create(ctx, &corev1.Secret{
	// 	TypeMeta: metav1.TypeMeta{
	// 		Kind:       "Secret",
	// 		APIVersion: "v1",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      credentialsName,
	// 		Namespace: account.TargetNamespace,
	// 	},
	// 	Immutable: new(bool),
	// 	Data: map[string][]byte{
	// 		corev1.DockerConfigJsonKey: []byte(out.DockerConfigJson),
	// 	},
	// 	Type: corev1.SecretTypeDockerConfigJson,
	// }); err != nil {
	// 	return err
	// }

	return nil
}

func (d *domain) EnsureKloudliteRegistryCredentials(ctx UserContext, accountName string) error {
	a, err := d.findAccount(ctx, accountName)
	if err != nil {
		return errors.NewE(err)
	}

	return d.ensureKloudliteRegistryCredentials(ctx, a)
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

	// if err := d.ensureNamespaceForAccount(ctx, account.Name, account.TargetNamespace); err != nil {
	// 	return nil, errors.NewE(err)
	// }

	// if err := d.ensureKloudliteRegistryCredentials(ctx, &account); err != nil {
	// 	return nil, errors.NewE(err)
	// }

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

func (d *domain) ResyncAccount(ctx UserContext, name string) error {
	acc, err := d.findAccount(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}
	if err := d.ensureNamespaceForAccount(ctx, acc.Name, acc.TargetNamespace); err != nil {
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
