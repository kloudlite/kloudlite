package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/repos"
	"math"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

type domainI struct {
	iamCli      iam.IAMClient
	consoleCli  console.ConsoleClient
	accountRepo repos.DbRepo[*Account]
}

func (domain *domainI) GetUserMemberships(ctx context.Context, id repos.ID) ([]*Membership, error) {
	rbs, err := domain.iamCli.ListResourceMemberships(ctx, &iam.InResourceMemberships{
		ResourceId:   string(id),
		ResourceType: common.ResourceAccount,
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
		ResourceType: common.ResourceAccount,
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
	billing *model.BillingInput,
) (*Account, error) {
	create, err := domain.accountRepo.Create(ctx, &Account{
		Name: name,
		Billing: Billing{
			StripeSetupIntentId: billing.StripeSetupIntentID,
			CardholderName:      billing.CardholderName,
			Address:             billing.Address,
		},
		IsActive:   true,
		CreatedAt:  time.Time{},
		ReadableId: repos.ID(generateReadable(name)),
	})

	if err != nil {
		return nil, err
	}

	_, err = domain.iamCli.AddMembership(ctx, &iam.InAddMembership{
		UserId:       string(userId),
		ResourceType: common.ResourceAccount,
		ResourceId:   string(create.Id),
		Role:         string(common.AccountOwner),
	})
	if err != nil {
		return nil, err
	}
	_, err = domain.consoleCli.CreateDefaultCluster(ctx, &console.CreateClusterIn{
		AccountId:   string(create.Id),
		AccountName: create.Name,
	})
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return create, err
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
		ResourceType: common.ResourceAccount,
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
		ResourceType: common.ResourceAccount,
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
	return domain.accountRepo.FindById(ctx, id)
}

func fxDomain(
	accountRepo repos.DbRepo[*Account],
	iamCli iam.IAMClient,
	consoleClient console.ConsoleClient,
) Domain {
	return &domainI{
		iamCli,
		consoleClient,
		accountRepo,
	}
}
