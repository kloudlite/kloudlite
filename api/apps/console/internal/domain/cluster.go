package domain

import (
	"context"
	"encoding/base64"

	"kloudlite.io/apps/console/internal/domain/entities"
)

func (d *domain) AddNewCluster(ctx context.Context, name string, subDomain string, kubeConfig string) (*entities.Cluster, error) {
	cluster, err := d.clusterRepo.Create(
		ctx, &entities.Cluster{
			Name:       name,
			SubDomain:  subDomain,
			KubeConfig: base64.RawStdEncoding.EncodeToString([]byte(kubeConfig)),
		},
	)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
