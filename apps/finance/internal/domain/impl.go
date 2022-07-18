package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/stripe"
	"math"
	"math/rand"
	"reflect"
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
	billablesRepo          repos.DbRepo[*AccountBilling]
	accountInviteTokenRepo cache.Repo[*AccountInviteToken]
	inventoryPath          string
	stripeCli              *stripe.Client
}

func (d *domainI) GetOutstandingAmount(ctx context.Context, accountId repos.ID) (float64, error) {
	accountBillings, err := d.billablesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": accountId,
			"month":      nil,
		},
	})
	if err != nil {
		return 0, err
	}
	var billableTotal float64
	for _, ab := range accountBillings {
		if ab.EndTime == nil {
			bill, err := d.calculateBill(ctx, ab.Billables, ab.StartTime, time.Now())
			if err != nil {
				return 0, err
			}
			billableTotal = billableTotal + bill
		} else {
			billableTotal = billableTotal + ab.BillAmount
		}
	}
	return billableTotal, nil
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

func (d *domainI) calculateBill(ctx context.Context, billables []Billable, startTime time.Time, endTime time.Time) (float64, error) {
	var billableTotal float64
	for _, billable := range billables {
		if billable.ResourceType == "Pod" {
			plan, err := d.GetComputePlanByName(ctx, billable.Plan)
			if err != nil {
				return 0, err
			}
			billableTotal = billableTotal + func() float64 {
				if billable.IsShared {
					return float64(billable.Count) * billable.Quantity * (plan.SharedPrice / (28 * 24 * 60 * 60)) * endTime.Sub(startTime).Seconds()
				} else {
					return float64(billable.Count) * billable.Quantity * (plan.DedicatedPrice / (28 * 24 * 60 * 60)) * endTime.Sub(startTime).Seconds()
				}
			}()
		}
	}
	return billableTotal, nil

}

func (d *domainI) TriggerBillingEvent(
	ctx context.Context,
	accountId repos.ID,
	resourceId repos.ID,
	projectId repos.ID,
	eventType string,
	billables []Billable,
	timeStamp time.Time,
) error {
	one, err := d.billablesRepo.FindOne(ctx, repos.Filter{
		"account_id":  accountId,
		"resource_id": resourceId,
		"end_time":    nil,
	})

	if err != nil {
		return err
	}

	if one == nil {
		_, err := d.billablesRepo.Create(ctx, &AccountBilling{
			AccountId:  accountId,
			ResourceId: resourceId,
			ProjectId:  projectId,
			Billables:  billables,
			StartTime:  timeStamp,
		})
		return err
	}

	if eventType == "end" {
		one.EndTime = &timeStamp
		bill, err := d.calculateBill(ctx, one.Billables, one.StartTime, timeStamp)
		if err != nil {
			return err
		}
		one.BillAmount = bill
		_, err = d.billablesRepo.UpdateById(ctx, one.Id, one)
		return err
	}

	billablesBytes, err := json.Marshal(billables)
	if err != nil {
		return err
	}
	oneBytes, err := json.Marshal(one.Billables)
	if err != nil {
		return err
	}

	isEqual, err := JSONBytesEqual(billablesBytes, oneBytes)
	if err != nil {
		return err
	}

	if isEqual {
		return nil
	}

	one.EndTime = &timeStamp
	bill, err := d.calculateBill(ctx, one.Billables, one.StartTime, timeStamp)
	if err != nil {
		return err
	}
	one.BillAmount = bill
	_, err = d.billablesRepo.UpdateById(ctx, one.Id, one)

	if err != nil {
		return err
	}

	_, err = d.billablesRepo.Create(ctx, &AccountBilling{
		AccountId:  accountId,
		ResourceId: resourceId,
		ProjectId:  projectId,
		Billables:  billables,
		StartTime:  timeStamp,
	})

	return err
}

func (d *domainI) GetStoragePlanByName(ctx context.Context, name string) (*StoragePlan, error) {
	fileData, err := ioutil.ReadFile(fmt.Sprint(d.inventoryPath, "/storage.yaml"))
	if err != nil {
		return nil, err
	}
	var items []StoragePlan
	err = yaml.Unmarshal(fileData, &items)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.Name == name {
			return &i, nil
		}
	}
	return nil, errors.New("inventory item not found")
}

func (d *domainI) GetLambdaPlanByName(ctx context.Context, name string) (*LamdaPlan, error) {
	fileData, err := ioutil.ReadFile(fmt.Sprint(d.inventoryPath, "/lamda.yaml"))
	if err != nil {
		return nil, err
	}
	var items []LamdaPlan
	err = yaml.Unmarshal(fileData, &items)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.Name == name {
			return &i, nil
		}
	}
	return nil, errors.New("inventory item not found")
}

func (d *domainI) GetComputePlanByName(ctx context.Context, name string) (*ComputePlan, error) {
	fileData, err := ioutil.ReadFile(fmt.Sprint(d.inventoryPath, "/compute.yaml"))
	if err != nil {
		return nil, err
	}
	var items []ComputePlan
	err = yaml.Unmarshal(fileData, &items)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.Name == name {
			return &i, nil
		}
	}
	return nil, errors.New("inventory item not found")
}

func (d *domainI) GetCurrentMonthBilling(ctx context.Context, accountID repos.ID) ([]*AccountBilling, time.Time, error) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	find, err := d.billablesRepo.Find(ctx, repos.Query{
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

func (d *domainI) ConfirmAccountMembership(ctx context.Context, invitationToken string) (bool, error) {
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
	_, err = d.iamCli.ConfirmMembership(ctx, &iam.InConfirmMembership{
		UserId:     string(existingToken.UserId),
		ResourceId: string(existingToken.AccountId),
		Role:       existingToken.Role,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) StartBillable(
	ctx context.Context,
	accountId repos.ID,
	resourceType string,
	quantity float32,
) (*AccountBilling, error) {
	create, err := d.billablesRepo.Create(ctx, &AccountBilling{
		AccountId: accountId,
		//ResourceType: resourceType,
		//Quantity:     quantity,
		StartTime: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (d *domainI) StopBillable(
	ctx context.Context,
	billableId repos.ID,
) error {
	id, err := d.billablesRepo.FindById(ctx, billableId)
	if err != nil {
		return err
	}
	time := time.Now()
	id.EndTime = &time
	_, err = d.billablesRepo.UpdateById(ctx, billableId, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *domainI) GetAccountMembership(ctx context.Context, userId repos.ID, accountId repos.ID) (*Membership, error) {
	membership, err := d.iamCli.GetMembership(
		ctx, &iam.InGetMembership{
			UserId:       string(userId),
			ResourceType: "account",
			ResourceId:   string(accountId),
		},
	)
	if err != nil {
		return nil, err
	}
	return &Membership{
		AccountId: repos.ID(membership.ResourceId),
		UserId:    repos.ID(membership.UserId),
		Role:      common.Role(membership.Role),
	}, nil
}

func (d *domainI) GetUserMemberships(ctx context.Context, id repos.ID) ([]*Membership, error) {
	rbs, err := d.iamCli.ListResourceMemberships(
		ctx, &iam.InResourceMemberships{
			ResourceId:   string(id),
			ResourceType: string(common.ResourceAccount),
		},
	)
	if err != nil {
		return nil, err
	}
	var memberships []*Membership
	for _, rb := range rbs.RoleBindings {
		memberships = append(
			memberships, &Membership{
				AccountId: repos.ID(rb.ResourceId),
				UserId:    repos.ID(rb.UserId),
				Role:      common.Role(rb.Role),
			},
		)
	}

	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func (d *domainI) GetAccountMemberships(ctx context.Context, id repos.ID) ([]*Membership, error) {
	rbs, err := d.iamCli.ListUserMemberships(
		ctx, &iam.InUserMemberships{
			UserId:       string(id),
			ResourceType: string(common.ResourceAccount),
		},
	)
	if err != nil {
		return nil, err
	}
	var memberships []*Membership

	for _, rb := range rbs.RoleBindings {
		memberships = append(
			memberships, &Membership{
				AccountId: repos.ID(rb.ResourceId),
				UserId:    repos.ID(rb.UserId),
				Role:      common.Role(rb.Role),
				Accepted:  rb.Accepted,
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

func (d *domainI) CreateAccount(
	ctx context.Context,
	userId repos.ID,
	name string,
	billing Billing,
) (*Account, error) {
	id := d.accountRepo.NewId()
	customer, err := d.stripeCli.NewCustomer(string(id), billing.PaymentMethodId)
	if err != nil {
		return nil, err
	}
	billing.StripeCustomerId = customer.Str()
	acc, err := d.accountRepo.Create(
		ctx, &Account{
			BaseEntity: repos.BaseEntity{
				Id: id,
			},
			Name:         name,
			ContactEmail: "",
			Billing:      billing,
			IsActive:     true,
			IsDeleted:    false,
			CreatedAt:    time.Now(),
			ReadableId:   repos.ID(generateReadable(name)),
		},
	)
	if err != nil {
		return nil, err
	}
	_, err = d.iamCli.AddMembership(
		ctx, &iam.InAddMembership{
			UserId:       string(userId),
			ResourceType: string(common.ResourceAccount),
			ResourceId:   string(acc.Id),
			Role:         string(common.AccountOwner),
		},
	)
	if err != nil {
		return nil, err
	}

	return acc, err
}

func (d *domainI) UpdateAccount(ctx context.Context, id repos.ID, name *string, email *string) (*Account, error) {
	acc, err := d.accountRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		acc.Name = *name
	}
	if email != nil {
		acc.ContactEmail = *email
	}
	updated, err := d.accountRepo.UpdateById(ctx, id, acc)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *domainI) UpdateAccountBilling(ctx context.Context, id repos.ID, b *Billing) (*Account, error) {
	acc, err := d.accountRepo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	acc.Billing = Billing{
		PaymentMethodId: b.PaymentMethodId,
		CardholderName:  b.CardholderName,
		Address:         b.Address,
	}
	updated, err := d.accountRepo.UpdateById(ctx, id, acc)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (d *domainI) AddAccountMember(
	ctx context.Context,
	accountId repos.ID,
	email string,
	role common.Role,
) (bool, error) {
	account, err := d.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}
	byEmail, err := d.authClient.EnsureUserByEmail(ctx, &auth.GetUserByEmailRequest{Email: email})
	if err != nil {
		return false, err
	}
	_, err = d.iamCli.InviteMembership(
		ctx, &iam.InAddMembership{
			UserId:       byEmail.UserId,
			ResourceType: string(common.ResourceAccount),
			ResourceId:   string(accountId),
			Role:         string(role),
		},
	)
	if err != nil {
		return false, err
	}
	token := generateId("acc-invite")
	err = d.accountInviteTokenRepo.Set(ctx, token, &AccountInviteToken{
		Token:     token,
		UserId:    repos.ID(byEmail.UserId),
		Role:      string(role),
		AccountId: accountId,
	})
	if err != nil {
		return false, err
	}
	_, err = d.commsClient.SendAccountMemberInviteEmail(ctx, &comms.AccountMemberInviteEmailInput{
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

func (d *domainI) RemoveAccountMember(
	ctx context.Context,
	accountId repos.ID,
	userId repos.ID,
) (bool, error) {
	_, err := d.iamCli.RemoveMembership(
		ctx, &iam.InRemoveMembership{
			UserId:     string(userId),
			ResourceId: string(accountId),
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) UpdateAccountMember(
	ctx context.Context,
	accountId repos.ID,
	userId repos.ID,
	role string,
) (bool, error) {
	_, err := d.iamCli.AddMembership(
		ctx, &iam.InAddMembership{
			UserId:       string(userId),
			ResourceType: string(common.ResourceAccount),
			ResourceId:   string(accountId),
			Role:         role,
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) DeactivateAccount(ctx context.Context, accountId repos.ID) (bool, error) {
	matched, err := d.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}
	matched.IsActive = false
	_, err = d.accountRepo.UpdateById(ctx, accountId, matched)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) ActivateAccount(ctx context.Context, accountId repos.ID) (bool, error) {
	matched, err := d.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}
	matched.IsActive = true
	_, err = d.accountRepo.UpdateById(ctx, accountId, matched)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) DeleteAccount(ctx context.Context, accountId repos.ID) (bool, error) {
	// TODO: delete harbor project
	matched, err := d.accountRepo.FindById(ctx, accountId)
	if err != nil {
		return false, err
	}
	matched.IsDeleted = true
	_, err = d.accountRepo.UpdateById(ctx, accountId, matched)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domainI) GetAccount(ctx context.Context, id repos.ID) (*Account, error) {
	return d.accountRepo.FindById(ctx, id)
}

func fxDomain(
	accountRepo repos.DbRepo[*Account],
	billablesRepo repos.DbRepo[*AccountBilling],
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
	ciClient ci.CIClient,
	authClient auth.AuthClient,
	env *Env,
	//commsClient comms.CommsClient,
	accountInviteTokenRepo cache.Repo[*AccountInviteToken],
	stripeCli *stripe.Client,
) Domain {
	return &domainI{
		authClient,
		iamCli,
		consoleClient,
		accountRepo,
		ciClient,
		nil,
		billablesRepo,
		accountInviteTokenRepo,
		env.InventoryPath,
		stripeCli,
	}
}
