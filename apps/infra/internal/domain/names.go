package domain

import (
	"context"
	"fmt"

	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) SuggestName(ctx context.Context, seed *string) string {
	return fn.GenReadableName(fn.DefaultIfNil(seed, ""))
}

type ResType string

const (
	ResTypeCluster        ResType = "cluster"
	ResTypeProviderSecret ResType = "providersecret"
	ResTypeNodePool       ResType = "nodepool"
	ResTypeVPNDevice      ResType = "vpn_device"
)

type CheckNameAvailabilityOutput struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

func (d *domain) CheckNameAvailability(ctx InfraContext, typeArg ResType, clusterName *string, name string) (*CheckNameAvailabilityOutput, error) {
	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, err
	}

	if !fn.IsValidK8sResourceName(name) {
		return &CheckNameAvailabilityOutput{Result: false, SuggestedNames: fn.GenValidK8sResourceNames(name, 3)}, nil
	}

	fromFindOneResult := func(data any, err error) (*CheckNameAvailabilityOutput, error) {
		if err != nil {
			return &CheckNameAvailabilityOutput{Result: false}, err
		}

		if data == nil {
			return &CheckNameAvailabilityOutput{Result: true}, nil
		}

		return &CheckNameAvailabilityOutput{Result: false, SuggestedNames: fn.GenValidK8sResourceNames(name, 3)}, nil
	}

	switch typeArg {
	case ResTypeCluster:
		{
			cp, err := d.clusterRepo.FindOne(ctx, repos.Filter{
				"accountName":        ctx.AccountName,
				"metadata.name":      name,
				"metadata.namespace": accNs,
			})

			return fromFindOneResult(cp, err)
		}
	case ResTypeProviderSecret:
		{
			cp, err := d.secretRepo.FindOne(ctx, repos.Filter{
				"accountName":        ctx.AccountName,
				"metadata.name":      name,
				"metadata.namespace": accNs,
			})

			return fromFindOneResult(cp, err)
		}
	case ResTypeNodePool:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, fmt.Errorf("clusterName is required for checking name availability for %s", ResTypeNodePool)
			}

			cp, err := d.nodePoolRepo.FindOne(ctx, repos.Filter{
				"accountName":   ctx.AccountName,
				"clusterName":   clusterName,
				"metadata.name": name,
			})

			return fromFindOneResult(cp, err)
		}
	case ResTypeVPNDevice:
		{
			if clusterName == nil || *clusterName == "" {
				return nil, fmt.Errorf("clusterName is required for checking name availability for %s", ResTypeVPNDevice)
			}

			cp, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
				"accountName":   ctx.AccountName,
				"clusterName":   clusterName,
				"metadata.name": name,
			})

			return fromFindOneResult(cp, err)
		}
	default:
		{
			return &CheckNameAvailabilityOutput{Result: false}, fmt.Errorf("unknown resource type provided: %q", typeArg)
		}
	}
}
