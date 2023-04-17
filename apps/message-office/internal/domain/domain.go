package domain

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	moRepo          repos.DbRepo[*MessageOfficeToken]
	accessTokenRepo repos.DbRepo[*AccessToken]
}

// ValidationAccessToken implements Domain
func (d *domain) ValidationAccessToken(ctx context.Context, accessToken string, accountName string, clusterName string) error {
	r, err := d.accessTokenRepo.FindOne(ctx, repos.Filter{
		"accessToken": accessToken,
		"accountName": accountName,
		"clusterName": clusterName,
	})
	if err != nil {
		return err
	}

	if r == nil {
		return fmt.Errorf("invalid access token")
	}

	return nil
}

func (d *domain) getClusterToken(ctx context.Context, accountName string, clusterName string) (string, error) {
	mot, err := d.moRepo.FindOne(ctx, repos.Filter{"accountName": accountName, "clusterName": clusterName})
	if err != nil {
		return "", err
	}
	if mot == nil {
		return "", nil
	}
	return mot.Token, nil
}

func (d *domain) GetClusterToken(ctx context.Context, accountName string, clusterName string) (string, error) {
	return d.getClusterToken(ctx, accountName, clusterName)
}

func (d *domain) GenClusterToken(ctx context.Context, accountName, clusterName string) (string, error) {
	token, err := d.getClusterToken(ctx, accountName, clusterName)
	if err != nil {
		return "", err
	}
	if token != "" {
		return token, nil
	}
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

func (d *domain) GenAccessToken(ctx context.Context, clusterToken string) (string, error) {
	mot, err := d.moRepo.FindOne(ctx, repos.Filter{"token": clusterToken})
	if err != nil {
		return "", err
	}
	if mot == nil {
		return "", fmt.Errorf("no such cluster token found")
	}

	record, err := d.accessTokenRepo.Upsert(ctx, repos.Filter{
		"accountName": mot.AccountName,
		"clusterName": mot.ClusterName,
	}, &AccessToken{
		AccountName: mot.AccountName,
		ClusterName: mot.ClusterName,
		AccessToken: fn.CleanerNanoidOrDie(40),
	})
	if err != nil {
		return "", err
	}

	if record == nil {
		return "", fmt.Errorf("failed to upsert into accessToken collection")
	}

	if err := d.moRepo.DeleteById(ctx, mot.Id); err != nil {
		return "", err
	}

	return record.AccessToken, nil
}

var Module = fx.Module(
	"domain",
	fx.Provide(func(
		moRepo repos.DbRepo[*MessageOfficeToken],
		accessTokenRepo repos.DbRepo[*AccessToken],
	) Domain {
		return &domain{
			moRepo:          moRepo,
			accessTokenRepo: accessTokenRepo,
		}
	}),
)
