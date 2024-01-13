package domain

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) SuggestName(ctx context.Context, seed *string) string {
	return fn.GenReadableName(fn.DefaultIfNil(seed, ""))
}

type ResType string

const (
	ResTypeCluster        ResType = "cluster"
	ResTypeClusterManagedService        ResType = "cluster_managed_service"
	ResTypeProviderSecret ResType = "providersecret"
	ResTypeNodePool       ResType = "nodepool"
	ResTypeHelmRelease       ResType = "helm_release"
	ResTypeVPNDevice      ResType = "vpn_device"
)

type CheckNameAvailabilityOutput struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

func checkResourceName[T repos.Entity](ctx context.Context, filters repos.Filter, repo repos.DbRepo[T]) (*CheckNameAvailabilityOutput, error) {
	res, err := repo.FindOne(ctx, filters)
	if err != nil {
		return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
	}

	if fn.IsNil(res) {
		return &CheckNameAvailabilityOutput{Result: true}, nil
	}

	return &CheckNameAvailabilityOutput{
		Result:         false,
		SuggestedNames: fn.GenValidK8sResourceNames(filters["metadata.name"].(string), 3),
	}, nil
}


func (d *domain) CheckNameAvailability(ctx InfraContext, typeArg ResType, clusterName *string, name string) (*CheckNameAvailabilityOutput, error) {

	if !fn.IsValidK8sResourceName(name) {
		return &CheckNameAvailabilityOutput{Result: false, SuggestedNames: fn.GenValidK8sResourceNames(name, 3)}, nil
	}

	switch typeArg {
	case ResTypeCluster:
		{
			return checkResourceName(ctx, repos.Filter{
				"accountName": ctx.AccountName,
				"metadata.name": name,
			}, d.clusterRepo)
		}
	case ResTypeProviderSecret:
		{
			return checkResourceName(ctx, repos.Filter{
				"accountName":        ctx.AccountName,
				"metadata.name": name,
			}, d.secretRepo)
		}
	case ResTypeNodePool:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, errors.Newf("clusterName is required for checking name availability for %s", ResTypeHelmRelease)
			}
			return checkResourceName(ctx, repos.Filter{
				"accountName":   ctx.AccountName,
				"clusterName":   clusterName,
				"metadata.name": name,
			}, d.nodePoolRepo)
		}
	case ResTypeHelmRelease:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, errors.Newf("clusterName is required for checking name availability for %s", ResTypeNodePool)
			}
			return checkResourceName(ctx, repos.Filter{
				"accountName":   ctx.AccountName,
				"clusterName":   clusterName,
				"metadata.name": name,
			}, d.helmReleaseRepo)
		}
	case ResTypeClusterManagedService:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, errors.Newf("clusterName is required for checking name availability for %s", ResTypeNodePool)
			}
			return checkResourceName(ctx, repos.Filter{
				"accountName":   ctx.AccountName,
				"clusterName":   clusterName,
				"metadata.name": name,
			}, d.clusterManagedServiceRepo)
		}
	case ResTypeVPNDevice:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, errors.Newf("clusterName is required for checking name availability for %s", ResTypeVPNDevice)
			}

			return checkResourceName(ctx, repos.Filter{
				"accountName":   ctx.AccountName,
				"clusterName":   clusterName,
				"metadata.name": name,
			}, d.vpnDeviceRepo)
		}
	default:
		{
			return &CheckNameAvailabilityOutput{Result: false}, errors.Newf("unknown resource type provided: %q", typeArg)
		}
	}
}
