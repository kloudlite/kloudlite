package domain

import (
	"context"
	"log/slog"

	"github.com/kloudlite/api/pkg/errors"
	"go.uber.org/fx"

	platform_edge "github.com/kloudlite/api/apps/message-office/internal/domain/platform-edge"
	"github.com/kloudlite/api/apps/message-office/internal/entities"
	"github.com/kloudlite/api/apps/message-office/internal/env"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

type PlatformEdgeDomain struct {
	*platform_edge.Repo
}

type domain struct {
	moRepo repos.DbRepo[*MessageOfficeToken]
	env    *env.Env
	logger *slog.Logger

	PlatformEdgeDomain
}

// func (d *domain) ValidateAccessToken(ctx context.Context, accessToken string, accountName string, clusterName string) error {
// 	r, err := d.accessTokenRepo.FindOne(ctx, repos.Filter{
// 		"accessToken": accessToken,
// 		"accountName": accountName,
// 		"clusterName": clusterName,
// 	})
// 	if err != nil {
// 		return errors.NewE(err)
// 	}
//
// 	if r == nil {
// 		return errors.Newf("invalid access token")
// 	}
//
// 	return nil
// }

func (d *domain) getClusterToken(ctx context.Context, accountName string, clusterName string) (string, error) {
	if accountName == "" || clusterName == "" {
		return "", errors.Newf("accountName and/or clusterName cannot be empty")
	}
	mot, err := d.moRepo.FindOne(ctx, repos.Filter{"accountName": accountName, "clusterName": clusterName})
	if err != nil {
		return "", errors.NewE(err)
	}
	if mot == nil {
		return "", nil
	}
	return mot.Token, nil
}

func (d *domain) FindClusterToken(ctx context.Context, clusterToken string) (*MessageOfficeToken, error) {
	if clusterToken == "" {
		return nil, errors.Newf("clusterToken cannot be empty")
	}
	mot, err := d.moRepo.FindOne(ctx, repos.Filter{"token": clusterToken})
	if err != nil {
		return nil, errors.NewE(err)
	}
	return mot, nil
}

func (d *domain) GetClusterToken(ctx context.Context, accountName string, clusterName string) (string, error) {
	return d.getClusterToken(ctx, accountName, clusterName)
}

func (d *domain) GenClusterToken(ctx context.Context, accountName, clusterName string) (string, error) {
	token, err := d.getClusterToken(ctx, accountName, clusterName)
	if err != nil {
		return "", errors.NewE(err)
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
		return "", errors.NewE(err)
	}
	return record.Token, nil
}

// func (d *domain) GenAccessToken(ctx context.Context, clusterToken string) (*AccessToken, error) {
// 	mot, err := d.moRepo.FindOne(ctx, repos.Filter{"token": clusterToken})
// 	if err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	if mot == nil {
// 		return nil, errors.Newf("no such cluster token found")
// 	}
//
// 	if mot.Granted != nil && *mot.Granted {
// 		d.logger.Infof("a valid access-token has already been issued for this cluster token, granting a new one, and removing the old one")
// 	}
//
// 	record, err := d.accessTokenRepo.Upsert(ctx, repos.Filter{
// 		"accountName": mot.AccountName,
// 		"clusterName": mot.ClusterName,
// 	}, &AccessToken{
// 		AccountName: mot.AccountName,
// 		ClusterName: mot.ClusterName,
// 		AccessToken: fn.CleanerNanoidOrDie(40),
// 	})
// 	if err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	if record == nil {
// 		return nil, errors.Newf("failed to upsert into accessToken collection")
// 	}
//
// 	mot.Granted = fn.New(true)
// 	if _, err := d.moRepo.UpdateById(ctx, mot.Id, mot); err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	return record, nil
// }

var Module = fx.Module(
	"domain",
	fx.Provide(func(
		ev *env.Env,
		moRepo repos.DbRepo[*MessageOfficeToken],
		edgeClustersRepo repos.DbRepo[*entities.PlatformEdgeCluster],
		allocatedClustersRepo repos.DbRepo[*entities.ClusterAllocation],
		// accessTokenRepo repos.DbRepo[*AccessToken],
		logger *slog.Logger,
	) Domain {
		return &domain{
			moRepo: moRepo,
			env:    ev,
			logger: logger,
			PlatformEdgeDomain: PlatformEdgeDomain{Repo: &platform_edge.Repo{
				EdgeClusters:      edgeClustersRepo,
				AllocatedClusters: allocatedClustersRepo,
			}},
		}
	}),
)
