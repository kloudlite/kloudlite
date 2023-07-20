package domain

import (
	"context"
	"fmt"

	"go.uber.org/fx"

	"kloudlite.io/apps/message-office/internal/env"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	moRepo          repos.DbRepo[*MessageOfficeToken]
	env             *env.Env
	accessTokenRepo repos.DbRepo[*AccessToken]
	logger          logging.Logger
}

func (d *domain) ValidateAccessToken(ctx context.Context, accessToken string, accountName string, clusterName string) error {
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
	if accountName == "" || clusterName == "" {
		return "", fmt.Errorf("accountName and/or clusterName cannot be empty")
	}
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

func (d *domain) GenAccessToken(ctx context.Context, clusterToken string) (*AccessToken, error) {
	mot, err := d.moRepo.FindOne(ctx, repos.Filter{"token": clusterToken})
	if err != nil {
		return nil, err
	}

	if mot == nil {
		return nil, fmt.Errorf("no such cluster token found")
	}

	if mot.Granted != nil && *mot.Granted {
		d.logger.Infof("a valid access-token has already been issued for this cluster token, granting a new one, and removing the old one")
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
		return nil, err
	}

	if record == nil {
		return nil, fmt.Errorf("failed to upsert into accessToken collection")
	}

	mot.Granted = fn.New(true)
	if _, err := d.moRepo.UpdateById(ctx, mot.Id, mot); err != nil {
		return nil, err
	}

	return record, nil
}

var Module = fx.Module(
	"domain",
	fx.Provide(func(
		moRepo repos.DbRepo[*MessageOfficeToken],
		accessTokenRepo repos.DbRepo[*AccessToken],
		logger logging.Logger,
	) Domain {
		return &domain{
			moRepo:          moRepo,
			accessTokenRepo: accessTokenRepo,
			logger:          logger,
		}
	}),
)
