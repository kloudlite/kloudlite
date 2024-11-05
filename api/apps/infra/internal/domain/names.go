package domain

import (
	"context"

	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) SuggestName(ctx context.Context, seed *string) string {
	return fn.GenReadableName(fn.DefaultIfNil(seed, ""))
}

type ResType string

const (
	ResTypeCluster               ResType = "cluster"
	ResTypeBYOKCluster           ResType = "byok_cluster"
	ResTypeGlobalVPNDevice       ResType = "global_vpn_device"
	ResTypeClusterManagedService ResType = "cluster_managed_service"
	ResTypeProviderSecret        ResType = "providersecret"
	ResTypeNodePool              ResType = "nodepool"
	ResTypeHelmRelease           ResType = "helm_release"
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
		SuggestedNames: fn.GenValidK8sResourceNames(filters[fields.MetadataName].(string), 3),
	}, nil
}

func (d *domain) CheckNameAvailability(ctx InfraContext, typeArg ResType, clusterName *string, name string) (*CheckNameAvailabilityOutput, error) {
	if !fn.IsValidK8sResourceName(name) {
		return &CheckNameAvailabilityOutput{Result: false, SuggestedNames: fn.GenValidK8sResourceNames(name, 3)}, nil
	}

	switch typeArg {
	// FIXME: remove me after web fixes it
	case ResTypeClusterManagedService:
		{
			return &CheckNameAvailabilityOutput{Result: true}, nil
		}
	case ResTypeCluster:
		{
			return checkResourceName(ctx, repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.MetadataName: name,
			}, d.clusterRepo)
		}
	case ResTypeBYOKCluster:
		{
			return checkResourceName(ctx, repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.MetadataName: name,
			}, d.byokClusterRepo)
		}
	case ResTypeProviderSecret:
		{
			return checkResourceName(ctx, repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.MetadataName: name,
			}, d.secretRepo)
		}
	case ResTypeGlobalVPNDevice:
		{
			return checkResourceName(ctx, repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.MetadataName: name,
			}, d.gvpnDevicesRepo)
		}
	case ResTypeNodePool:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, errors.Newf("clusterName is required for checking name availability for %s", ResTypeHelmRelease)
			}
			return checkResourceName(ctx, repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.ClusterName:  clusterName,
				fields.MetadataName: name,
			}, d.nodePoolRepo)
		}
	case ResTypeHelmRelease:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, errors.Newf("clusterName is required for checking name availability for %s", ResTypeNodePool)
			}
			return checkResourceName(ctx, repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.ClusterName:  clusterName,
				fields.MetadataName: name,
			}, d.helmReleaseRepo)
		}
	default:
		{
			return &CheckNameAvailabilityOutput{Result: false}, errors.Newf("unknown resource type provided: %q", typeArg)
		}
	}
}
