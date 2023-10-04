package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/container_registry"

	"github.com/kloudlite/operator/pkg/kubectl"
	"kloudlite.io/constants"

	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/stripe"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
)

func generateId(prefix string) string {
	id, e := fn.CleanerNanoid(28)
	if e != nil {
		panic(fmt.Errorf("could not get cleanerNanoid()"))
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(id))
}

func toK8sAccountCR(acc *Account) ([]byte, error) {
	kAcc := crdsv1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: acc.Name,
		},
		Spec: crdsv1.AccountSpec{
			HarborProjectName:      acc.Name,
			HarborUsername:         acc.Name,
			HarborSecretsNamespace: constants.NamespaceCore,
		},
	}
	kAcc.EnsureGVK()
	return json.Marshal(kAcc)
}

type domainI struct {
	invoiceRepo             repos.DbRepo[*BillingInvoice]
	authClient              auth.AuthClient
	iamClient               iam.IAMClient
	consoleClient           console.ConsoleClient
	containerRegistryClient container_registry.ContainerRegistryClient
	accountRepo             repos.DbRepo[*Account]
	commsClient             comms.CommsClient
	billablesRepo           repos.DbRepo[*AccountBilling]
	accountInviteTokenRepo  cache.Repo[*AccountInviteToken]
	stripeCli               *stripe.Client
	k8sYamlClient           kubectl.YAMLClient
	logger                  logging.Logger
}

func (d *domainI) ListAccounts(ctx FinanceContext) ([]*Account, error) {
	out, err := d.iamClient.ListMembershipsForUser(ctx, &iam.MembershipsForUserIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceAccount),
	})
	// out, err := d.iamClient.ListMembershipsByResource(ctx, &iam.MembershipsByResourceIn{
	// 	ResourceType: string(iamT.ResourceAccount),
	// })
	if err != nil {
		return nil, err
	}
	acc := make([]*Account, len(out.RoleBindings))
	for i := range out.RoleBindings {
		acc[i] = &Account{
			Name: strings.Split(out.RoleBindings[i].ResourceRef, "/")[0],
		}
	}
	return acc, nil
}

// ListInvitations implements Domain
func (d *domainI) ListInvitations(ctx FinanceContext, accountName string) ([]*Membership, error) {
	mems, err := d.iamClient.ListResourceMemberships(ctx, &iam.ResourceMembershipsIn{
		ResourceType: string(iamT.ResourceAccount),
		ResourceRef:  iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
	})
	if err != nil {
		return nil, err
	}

	m := make([]*Membership, len(mems.RoleBindings))

	for i := range mems.RoleBindings {
		//body
		m[i] = &Membership{
			AccountName: accountName,
			UserId:      repos.ID(mems.RoleBindings[i].UserId),
			Role:        iamT.Role(mems.RoleBindings[i].Role),
			Accepted:    mems.RoleBindings[i].Accepted,
		}
	}

	return m, nil
}

func JSONBytesEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

func (d *domainI) ConfirmAccountMembership(ctx FinanceContext, invitationToken string) (bool, error) {
	existingToken, err := d.accountInviteTokenRepo.Get(ctx, invitationToken)
	if err != nil {
		return false, err
	}

	if existingToken == nil {
		return false, errors.New("invitation token not found")
	}

	err = d.accountInviteTokenRepo.Drop(ctx, invitationToken)
	if err != nil {
		return false, err
	}
	_, err = d.iamClient.ConfirmMembership(
		ctx, &iam.ConfirmMembershipIn{
			UserId: string(existingToken.UserId),
			// ResourceId: string(existingToken.AccountName),
			// ResourceRef: iamT.NewResourceRef(clusterName string, kind string, namespace string, name string),
			Role: existingToken.Role,
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) GetAccountMembership(ctx FinanceContext, accountName string) (*Membership, error) {
	membership, err := d.iamClient.GetMembership(
		ctx, &iam.GetMembershipIn{
			UserId:       string(ctx.UserId),
			ResourceType: string(iamT.ResourceAccount),
			ResourceRef:  iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
	)
	if err != nil {
		return nil, err
	}
	return &Membership{
		AccountName: accountName,
		UserId:      repos.ID(membership.UserId),
		Role:        iamT.Role(membership.Role),
	}, nil
}

func (d *domainI) GetUserMemberships(ctx FinanceContext, resourceRef string) ([]*Membership, error) {
	rbs, err := d.iamClient.ListResourceMemberships(
		ctx, &iam.ResourceMembershipsIn{
			ResourceType: string(iamT.ResourceAccount),
			ResourceRef:  resourceRef,
		},
	)
	if err != nil {
		return nil, err
	}

	memberships := make([]*Membership, 0, len(rbs.RoleBindings))
	for _, rb := range rbs.RoleBindings {
		memberships = append(
			memberships, &Membership{
				AccountName: "",
				UserId:      repos.ID(rb.UserId),
				Role:        iamT.Role(rb.Role),
			},
		)
	}

	return memberships, nil
}

func (d *domainI) GetAccountMemberships(ctx FinanceContext) ([]*Membership, error) {
	rbs, err := d.iamClient.ListUserMemberships(
		ctx, &iam.UserMembershipsIn{
			UserId:       string(ctx.UserId),
			ResourceType: string(iamT.ResourceAccount),
		},
	)
	if err != nil {
		return nil, err
	}

	var memberships []*Membership
	for _, rb := range rbs.RoleBindings {
		memberships = append(
			memberships, &Membership{
				AccountName: strings.Split(rb.ResourceRef, "/")[0],
				UserId:      repos.ID(rb.UserId),
				Role:        iamT.Role(rb.Role),
				Accepted:    rb.Accepted,
			},
		)
	}

	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func generateReadable(name string) string {
	compile := regexp.MustCompile(`[^\da-zA-Z:,/s]+`)
	allString := compile.ReplaceAllString(strings.ToLower(name), "")
	m := math.Min(10, float64(len(allString)))
	return fmt.Sprintf("%v_%v", allString[:int(m)], rand.Intn(9999))
}

func (d *domainI) CreateAccount(ctx FinanceContext, name string, displayName string) (*Account, error) {
	uid, err := GetUser(ctx)
	if err != nil {
		return nil, err
	}
	if uid != string(ctx.UserId) {
		return nil, errors.New("you don't have permission to perform this operation")
	}

	id := d.accountRepo.NewId()

	acc, err := d.accountRepo.Create(
		ctx, &Account{
			BaseEntity:  repos.BaseEntity{Id: id},
			Name:        name,
			DisplayName: displayName,
			IsActive:    fn.New(true),
			CreatedAt:   time.Now(),
			ReadableId:  repos.ID(generateReadable(name)),
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = d.iamClient.AddMembership(
		ctx, &iam.AddMembershipIn{
			UserId:       string(ctx.UserId),
			ResourceType: string(iamT.ResourceAccount),
			ResourceRef:  iamT.NewResourceRef(acc.Name, iamT.ResourceAccount, acc.Name),
			Role:         string(iamT.RoleAccountAdmin),
		},
	)
	if err != nil {
		return nil, err
	}

	b, err := toK8sAccountCR(acc)
	if err != nil {
		return nil, err
	}

	if _, err = d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
		return nil, err
	}

	// creating account namespace
	ns, err := d.k8sYamlClient.K8sClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "acc-" + acc.Name,
			Labels: map[string]string{
				constants.AccountNameKey: acc.Name,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	d.logger.Infof("created namespace (%s) for account (%s)", ns.Name, acc.Name)

	return acc, err
}

func (d *domainI) UpdateAccount(ctx FinanceContext, accountName string, contactEmail *string) (*Account, error) {
	if err := d.checkAccountAccess(ctx, accountName, iamT.UpdateAccount); err != nil {
		return nil, err
	}

	acc, err := d.findAccount(ctx, accountName)
	if err != nil {
		return nil, err
	}

	if contactEmail != nil {
		acc.ContactEmail = *contactEmail
	}
	return d.accountRepo.UpdateById(ctx, acc.Id, acc)
}

func (d *domainI) UpdateAccountBilling(ctx FinanceContext, accountName string, b *Billing) (*Account, error) {
	if err := d.checkAccountAccess(ctx, accountName, "update_account"); err != nil {
		return nil, err
	}

	acc, err := d.findAccount(ctx, accountName)
	if err != nil {
		return nil, err
	}

	acc.Billing = Billing{
		PaymentMethodId: b.PaymentMethodId,
		CardholderName:  b.CardholderName,
		Address:         b.Address,
	}

	return d.accountRepo.UpdateById(ctx, acc.Id, acc)
}

// Invitation

func (d *domainI) DeleteInvitation(ctx FinanceContext, email string) (bool, error) {
	panic("not implemented")
}

func (d *domainI) InviteUser(ctx FinanceContext, accountName string, email string, role iamT.Role) (bool, error) {
	switch role {
	case "account-member":
		if err := d.checkAccountAccess(ctx, accountName, iamT.InviteAccountMember); err != nil {
			return false, err
		}
	case "account-admin":
		if err := d.checkAccountAccess(ctx, accountName, iamT.InviteAccountAdmin); err != nil {
			return false, err
		}
	default:
		return false, errors.New("role must be one of [account-member, account-admin]")
	}

	acc, err := d.findAccount(ctx, accountName)
	if err != nil {
		return false, err
	}

	invitedUser, err := d.authClient.EnsureUserByEmail(ctx, &auth.GetUserByEmailRequest{Email: email})
	if err != nil {
		return false, err
	}

	_, err = d.iamClient.InviteMembership(
		ctx, &iam.AddMembershipIn{
			UserId:       string(invitedUser.UserId),
			ResourceType: string(iamT.ResourceAccount),
			ResourceRef:  iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
			Role:         string(role),
		},
	)
	if err != nil {
		return false, err
	}

	token := generateId("acc-invite")
	if err := d.accountInviteTokenRepo.Set(
		ctx, token, &AccountInviteToken{
			Token:       token,
			UserId:      ctx.UserId,
			Role:        string(role),
			AccountName: accountName,
		},
	); err != nil {
		return false, err
	}

	if _, err = d.commsClient.SendAccountMemberInviteEmail(
		ctx, &comms.AccountMemberInviteEmailInput{
			AccountName:     acc.Name,
			InvitationToken: token,
			Email:           email,
			Name:            "",
		},
	); err != nil {
		return false, err
	}

	return true, nil
}

func (d *domainI) RemoveAccountMember(
	ctx FinanceContext,
	accountName string,
	userId repos.ID,
) (bool, error) {
	_, err := d.iamClient.RemoveMembership(
		ctx, &iam.RemoveMembershipIn{
			UserId:      string(userId),
			ResourceRef: iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) checkAccountAccess(ctx FinanceContext, accountName string, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId:       string(ctx.UserId),
		ResourceRefs: []string{iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName)},
		Action:       string(iamT.GetAccount),
	})

	if err != nil {
		return err
	}

	if !co.Status {
		return fmt.Errorf("unauthorized to access this account")
	}

	return nil
}

func (d *domainI) findAccount(ctx FinanceContext, name string) (*Account, error) {
	acc, err := d.accountRepo.FindOne(ctx, repos.Filter{"name": name})
	if err != nil {
		return nil, err
	}

	if acc == nil {
		return nil, fmt.Errorf("account with name=%q not found", name)
	}
	return acc, nil
}

func (d *domainI) UpdateAccountMember(
	ctx FinanceContext,
	accountName string,
	userId repos.ID,
	role string,
) (bool, error) {

	if err := d.checkAccountAccess(ctx, accountName, iamT.UpdateAccount); err != nil {
		return false, err
	}

	_, err := d.iamClient.AddMembership(
		ctx, &iam.AddMembershipIn{
			UserId:       string(userId),
			ResourceType: string(constants.ResourceAccount),
			Role:         role,
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) DeactivateAccount(ctx FinanceContext, accountName string) (bool, error) {
	if err := d.checkAccountAccess(ctx, accountName, iamT.ActivateAccount); err != nil {
		return false, err
	}

	acc, err := d.findAccount(ctx, accountName)
	if err != nil {
		return false, err
	}

	acc.IsActive = fn.New(false)
	_, err = d.accountRepo.UpdateById(ctx, acc.Id, acc)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) ActivateAccount(ctx FinanceContext, accountName string) (bool, error) {
	if err := d.checkAccountAccess(ctx, accountName, iamT.ActivateAccount); err != nil {
		return false, err
	}

	acc, err := d.findAccount(ctx, accountName)
	if err != nil {
		return false, err
	}

	acc.IsActive = fn.New(true)
	_, err = d.accountRepo.UpdateById(ctx, acc.Id, acc)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *domainI) DeleteAccount(ctx FinanceContext, accountName string) (bool, error) {
	// TODO: delete harbor project
	if err := d.checkAccountAccess(ctx, accountName, iamT.DeleteAccount); err != nil {
		return false, err
	}

	acc, err := d.findAccount(ctx, accountName)
	if err != nil {
		return false, err
	}
	acc.IsDeleted = fn.New(true)
	_, err = d.accountRepo.UpdateById(ctx, acc.Id, acc)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) GetAccount(ctx FinanceContext, accountName string) (*Account, error) {
	// _, err := GetUser(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	if err := d.checkAccountAccess(ctx, accountName, iamT.GetAccount); err != nil {
		return nil, err
	}
	return d.findAccount(ctx, accountName)
}

func (d *domainI) ReSyncToK8s(ctx FinanceContext, accountName string) error {
	if err := d.checkAccountAccess(ctx, accountName, iamT.GetAccount); err != nil {
		return err
	}
	acc, err := d.findAccount(ctx, accountName)
	if err != nil {
		return err
	}

	b, err := toK8sAccountCR(acc)
	if err != nil {
		return err
	}

	if _, err := d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
		return err
	}

	return nil
}

func fxDomain(
	accountRepo repos.DbRepo[*Account],
	billablesRepo repos.DbRepo[*AccountBilling],
	invoiceRepo repos.DbRepo[*BillingInvoice],
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
	containerRegistryClient container_registry.ContainerRegistryClient,
	authClient auth.AuthClient,
	commsClient comms.CommsClient,
	accountInviteTokenRepo cache.Repo[*AccountInviteToken],
	// stripeCli *stripe.Client,
	k8sYamlClient kubectl.YAMLClient,
	logger logging.Logger,
) Domain {
	return &domainI{
		invoiceRepo:             invoiceRepo,
		authClient:              authClient,
		iamClient:               iamCli,
		consoleClient:           consoleClient,
		containerRegistryClient: containerRegistryClient,
		accountRepo:             accountRepo,
		commsClient:             commsClient,
		accountInviteTokenRepo:  accountInviteTokenRepo,
		k8sYamlClient:           k8sYamlClient,
		logger:                  logger,
	}
}
