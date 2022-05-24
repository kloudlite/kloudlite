package domain

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/functions"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func generateId(prefix string) string {
	id, e := functions.CleanerNanoid(28)
	if e != nil {
		panic(fmt.Errorf("could not get cleanerNanoid()"))
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(id))
}

type domainI struct {
	authClient             auth.AuthClient
	iamCli                 iam.IAMClient
	consoleCli             console.ConsoleClient
	accountRepo            repos.DbRepo[*Account]
	ciClient               ci.CIClient
	commsClient            comms.CommsClient
	billablesRepo          repos.DbRepo[*Billable]
	accountInviteTokenRepo cache.Repo[*AccountInviteToken]
}

func (domain *domainI) GetComputeInventoryByName(ctx context.Context, name string) (*InventoryItem, error) {
	file, err := ioutil.ReadFile("./inventory.yaml")
	if err != nil {
		return nil, err
	}
	items := make([]*InventoryItem, 0)
	err = yaml.Unmarshal(file, &items)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.Name == name {
			return i, nil
		}
	}
	return nil, errors.New("inventory item not found")
}

func (domain *domainI) GetComputeInventory(provider *string) ([]*InventoryItem, error) {
	file, err := ioutil.ReadFile("./inventory.yaml")
	if err != nil {
		return nil, err
	}
	items := make([]*InventoryItem, 0)
	err = yaml.Unmarshal(file, &items)
	if err != nil {
		return nil, err
	}
	filteredItems := make([]*InventoryItem, 0)
	for _, i := range items {
		if i.Provider == *provider && i.Type == "Compute" {
			filteredItems = append(filteredItems, i)
		}
	}
	return filteredItems, nil
}

func (domain *domainI) GetCurrentMonthBilling(ctx context.Context, accountID repos.ID) ([]*Billable, time.Time, error) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	find, err := domain.billablesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": accountID,
			"start_time": repos.Filter{
				"$gte": firstOfMonth,
			},
		},
	})
	if err != nil {
		return nil, firstOfMonth, err
	}
	return find, firstOfMonth, nil
}

func (domain *domainI) ConfirmAccountMembership(ctx context.Context, invitationToken string) (bool, error) {
	existingToken, err := domain.accountInviteTokenRepo.Get(ctx, invitationToken)
	if err != nil {
		return false, err
	}
	if existingToken == nil {
		return false, errors.New("invitation token not found")
	}
	err = domain.accountInviteTokenRepo.Drop(ctx, invitationToken)
	if err != nil {
		return false, err
	}
	_, err = domain.iamCli.ConfirmMembership(ctx, &iam.InConfirmMembership{
		UserId:     string(existingToken.UserId),
		ResourceId: string(existingToken.AccountId),
		Role:       existingToken.Role,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (domain *domainI) StartBillable(
	ctx context.Context,
	accountId repos.ID,
	resourceType string,
	quantity float32,
) (*Billable, error) {
	create, err := domain.billablesRepo.Create(ctx, &Billable{
		AccountId:    accountId,
		ResourceType: resourceType,
		Quantity:     quantity,
		StartTime:    time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (domain *domainI) StopBillable(
	ctx context.Context,
	billableId repos.ID,
) error {
	id, err := domain.billablesRepo.FindById(ctx, billableId)
	if err != nil {
		return err
	}
	time := time.Now()
	id.EndTime = &time
	_, err = domain.billablesRepo.UpdateById(ctx, billableId, id)
	if err != nil {
		return err
	}
	return nil
}

func (domain *domainI) GetAccountMembership(ctx context.Context, userId repos.ID, accountId repos.ID) (*Membership, error) {
	membership, err := domain.iamCli.GetMembership(ctx, &iam.InGetMembership{
		UserId:       string(userId),
		ResourceType: "account",
		ResourceId:   string(accountId),
	})
	if err != nil {
		return nil, err
	}
	return &Membership{
		AccountId: repos.ID(membership.ResourceId),
		UserId:    repos.ID(membership.UserId),
		Role:      common.Role(membership.Role),
	}, nil
}

func (domain *domainI) GetUserMemberships(ctx context.Context, id repos.ID) ([]*Membership, error) {
	rbs, err := domain.iamCli.ListResourceMemberships(ctx, &iam.InResourceMemberships{
		ResourceId:   string(id),
		ResourceType: string(common.ResourceAccount),
	})
	if err != nil {
		return nil, err
	}
	var memberships []*Membership
	for _, rb := range rbs.RoleBindings {
		memberships = append(memberships, &Membership{
			AccountId: repos.ID(rb.ResourceId),
			UserId:    repos.ID(rb.UserId),
			Role:      common.Role(rb.Role),
		})
	}

	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func (domain *domainI) GetAccountMemberships(ctx context.Context, id repos.ID) ([]*Membership, error) {
	rbs, err := domain.iamCli.ListUserMemberships(ctx, &iam.InUserMemberships{
		UserId:       string(id),
		ResourceType: string(common.ResourceAccount),
	})
	if err != nil {
		return nil, err
	}
	var memberships []*Membership

	for _, rb := range rbs.RoleBindings {
		memberships = append(memberships, &Membership{
			AccountId: repos.ID(rb.ResourceId),
			UserId:    repos.ID(rb.UserId),
			Role:      common.Role(rb.Role),
			Accepted:  rb.Accepted,
		})
	}

	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func generateReadable(name string) string {
	compile := regexp.MustCompile(`[^0-9a-zA-Z:,/s]+`)
	allString := compile.ReplaceAllString(strings.ToLower(name), "")
	m := math.Min(10, float64(len(allString)))
	return fmt.Sprintf("%v_%v", allString[:int(m)], rand.Intn(9999))
}

func (domain *domainI) CreateAccount(
	ctx context.Context,
	userId repos.ID,
	name string,
	billing Billing,
	initialProvider string,
	initialRegion string,
) (*Account, error) {

	id := domain.accountRepo.NewId()
	_, err := domain.ciClient.CreateHarborProject(ctx, &ci.HarborProjectIn{Name: string(id)})
	if err != nil {
		return nil, errors.NewEf(err, "harbor account could not be created")
	}

	acc, err := domain.accountRepo.Create(ctx, &Account{
		BaseEntity: repos.BaseEntity{
			Id: id,
		},
		Name:         name,
		ContactEmail: "",
		Billing:      Billing{StripeSetupIntentId: billing.StripeSetupIntentId, CardholderName: billing.CardholderName, Address: billing.Address},
		IsActive:     true,
		IsDeleted:    false,
		CreatedAt:    time.Time{},
		ReadableId:   repos.ID(generateReadable(name)),
	})

	if err != nil {
		return nil, err
	}
	fmt.Println("sending message to console1")
	_, err = domain.iamCli.AddMembership(ctx, &iam.InAddMembership{
		UserId:       string(userId),
		ResourceType: string(common.ResourceAccount),
		ResourceId:   string(acc.Id),
		Role:         string(common.AccountOwner),
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("sending message to console")
	// TODO
	_, err = domain.consoleCli.SetupClusterForAccount(ctx, &console.AccountIn{
		AccountId:   string(acc.Id),
		AccountName: acc.Name,
		Provider:    initialProvider,
		Region:      initialRegion,
	})
	fmt.Println("sent message", err)
	if err != nil {
		return nil, err
	}

	return acc, err
}

func (domain *domainI) UpdateAccount(ctx context.Context, id repos.ID, name *string, email *string) (*Account, error) {
	acc, err := domain.accountRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		acc.Name = *name
	}
	if email != nil {
		acc.ContactEmail = *email
	}
	updated, err := domain.accountRepo.UpdateById(ctx, id, acc)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (domain *domainI) UpdateAccountBilling(ctx context.Context, id repos.ID, d *Billing) (*Account, error) {
	acc, err := domain.accountRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	acc.Billing = Billing{
		StripeSetupIntentId: d.StripeSetupIntentId,
		CardholderName:      d.CardholderName,
		Address:             d.Address,
	}
	updated, err := domain.accountRepo.UpdateById(ctx, id, acc)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (domain *domainI) AddAccountMember(
	ctx context.Context,
	accountId repos.ID,
	email string,
	role common.Role,
) (bool, error) {
	account, err := domain.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}

	byEmail, err := domain.authClient.EnsureUserByEmail(ctx, &auth.GetUserByEmailRequest{Email: email})
	if err != nil {
		return false, err
	}
	_, err = domain.iamCli.InviteMembership(ctx, &iam.InAddMembership{
		UserId:       byEmail.UserId,
		ResourceType: string(common.ResourceAccount),
		ResourceId:   string(accountId),
		Role:         string(role),
	})
	if err != nil {
		return false, err
	}
	token := generateId("acc-invite")
	err = domain.accountInviteTokenRepo.Set(ctx, token, &AccountInviteToken{
		Token:     token,
		UserId:    repos.ID(byEmail.UserId),
		Role:      string(role),
		AccountId: accountId,
	})
	if err != nil {
		return false, err
	}
	_, err = domain.commsClient.SendAccountMemberInviteEmail(ctx, &comms.AccountMemberInviteEmailInput{
		AccountName:     account.Name,
		InvitationToken: token,
		Email:           email,
		Name:            "",
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (domain *domainI) RemoveAccountMember(
	ctx context.Context,
	accountId repos.ID,
	userId repos.ID,
) (bool, error) {
	_, err := domain.iamCli.RemoveMembership(ctx, &iam.InRemoveMembership{
		UserId:     string(userId),
		ResourceId: string(accountId),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (domain *domainI) UpdateAccountMember(
	ctx context.Context,
	accountId repos.ID,
	userId repos.ID,
	role string,
) (bool, error) {
	_, err := domain.iamCli.AddMembership(ctx, &iam.InAddMembership{
		UserId:       string(userId),
		ResourceType: string(common.ResourceAccount),
		ResourceId:   string(accountId),
		Role:         role,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (domain *domainI) DeactivateAccount(ctx context.Context, accountId repos.ID) (bool, error) {
	matched, err := domain.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}
	matched.IsActive = false
	_, err = domain.accountRepo.UpdateById(ctx, accountId, matched)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (domain *domainI) ActivateAccount(ctx context.Context, accountId repos.ID) (bool, error) {
	matched, err := domain.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}
	matched.IsActive = true
	_, err = domain.accountRepo.UpdateById(ctx, accountId, matched)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (domain *domainI) DeleteAccount(ctx context.Context, accountId repos.ID) (bool, error) {
	// TODO: delete harbor project
	matched, err := domain.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}
	matched.IsDeleted = true
	_, err = domain.accountRepo.UpdateById(ctx, accountId, matched)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (domain *domainI) GetAccount(ctx context.Context, id repos.ID) (*Account, error) {
	fmt.Println("GetAccount", id)
	return domain.accountRepo.FindById(ctx, id)
}

func fxDomain(
	accountRepo repos.DbRepo[*Account],
	billablesRepo repos.DbRepo[*Billable],
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
	ciClient ci.CIClient,
	authClient auth.AuthClient,
	commsClient comms.CommsClient,
	accountInviteTokenRepo cache.Repo[*AccountInviteToken],
) Domain {
	return &domainI{
		authClient,
		iamCli,
		consoleClient,
		accountRepo,
		ciClient,
		commsClient,
		billablesRepo,
		accountInviteTokenRepo,
	}
}
