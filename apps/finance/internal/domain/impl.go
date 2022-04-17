package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/common"
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
	accountRepo repos.DbRepo[*Account]
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
	domain.iamCli.AddMembership(ctx, &iam.InAddMembership{
		UserId:       string(userId),
		ResourceType: common.ResourceAccount,
		ResourceId:   string(create.Id),
		Role:         string(iam.AccountOwner),
	})
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

func (domain *domainI) AddAccountMember(ctx context.Context, id string, email string, name string, role string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (domain *domainI) RemoveAccountMember(ctx context.Context, id repos.ID, id2 repos.ID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (domain *domainI) UpdateAccountMember(ctx context.Context, id repos.ID, id2 repos.ID, role string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (domain *domainI) DeactivateAccount(ctx context.Context, id repos.ID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (domain *domainI) ActivateAccount(ctx context.Context, id repos.ID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (domain *domainI) DeleteAccount(ctx context.Context, id repos.ID) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (domain *domainI) ListAccounts(ctx context.Context, id repos.ID) ([]*Account, error) {
	//TODO implement me
	panic("implement me")
}

func (domain *domainI) GetAccount(ctx context.Context, id repos.ID) (*Account, error) {
	return domain.accountRepo.FindById(ctx, id)
}

func fxDomain(
	accountRepo repos.DbRepo[*Account],
	iamCli iam.IAMClient,
) Domain {
	return &domainI{
		iamCli,
		accountRepo,
	}
}
