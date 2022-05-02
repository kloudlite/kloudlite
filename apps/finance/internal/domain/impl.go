package domain

import (
	"context"
	"fmt"
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

type domainI struct {
	iamCli      iam.IAMClient
	consoleCli  console.ConsoleClient
	accountRepo repos.DbRepo[*Account]
	ciClient    ci.CIClient
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
	// _, err = domain.consoleCli.CreateDefaultCluster(ctx, &console.CreateClusterIn{
	// 	AccountId:   string(acc.HarborId),
	// 	AccountName: acc.Name,
	// })
	// fmt.Println("sent message", err)
	// if err != nil {
	// 	return nil, err
	// }

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
	userId repos.ID,
	role common.Role,
) (bool, error) {
	_, err := domain.iamCli.AddMembership(ctx, &iam.InAddMembership{
		UserId:       string(userId),
		ResourceType: string(common.ResourceAccount),
		ResourceId:   string(accountId),
		Role:         string(role),
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
		Role:         string(role),
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
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
	ciClient ci.CIClient,
) Domain {
	return &domainI{
		iamCli,
		consoleClient,
		accountRepo,
		ciClient,
	}
}
