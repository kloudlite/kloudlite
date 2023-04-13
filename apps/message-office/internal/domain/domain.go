package domain

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	moRepo repos.DbRepo[*MessageOfficeToken]
}

func (d *domain) GetClusterToken(ctx context.Context, accountName string, clusterName string) (string, error) {
	mot, err := d.moRepo.FindOne(ctx, repos.Filter{"accountName": accountName, "clusterName": clusterName})
	if err != nil {
		return "", err
	}
	if mot == nil {
		return "", fmt.Errorf("no token found")
	}
	return mot.Token, nil
}

func (d *domain) GenClusterToken(ctx context.Context, accountName, clusterName string) (string, error) {
	record, err := d.moRepo.Create(ctx, &MessageOfficeToken{
		AccountName: accountName,
		ClusterName: clusterName,
		Token:       fn.CleanerNanoidOrDie(40),
	})
	if err != nil {
		return "", err
	}
	return record.Token, nil
}

var Module = fx.Module(
	"domain",
	fx.Provide(func(moRepo repos.DbRepo[*MessageOfficeToken]) Domain {
		return &domain{
			moRepo: moRepo,
		}
	}),
)
