package domain

import (
	"context"
	"fmt"

	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
)

func (d *domain) SuggestName(ctx context.Context, seed *string) string {
	return fn.GenReadableName(fn.DefaultIfNil(seed, ""))
}

type ResType string

const (
	ResTypeCluster        ResType = "cluster"
	ResTypeProviderSecret ResType = "providersecret"
	ResTypeNodePool       ResType = "nodepool"
)

type CheckNameAvailabilityOutput struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

func (d *domain) CheckNameAvailability(ctx InfraContext, typeArg ResType, clusterName *string, name string) (*CheckNameAvailabilityOutput, error) {
	switch typeArg {
	case ResTypeCluster:
		{
			if fn.IsValidK8sResourceName(name) {
				cp, err := d.clusterRepo.FindOne(ctx, repos.Filter{
					"accountName":        ctx.AccountName,
					"metadata.name":      name,
					"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}

				if cp == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}

			return &CheckNameAvailabilityOutput{Result: false, SuggestedNames: fn.GenValidK8sResourceNames(name, 3)}, nil
		}
	case ResTypeProviderSecret:
		{
			if fn.IsValidK8sResourceName(name) {
				cp, err := d.secretRepo.FindOne(ctx, repos.Filter{
					"accountName":        ctx.AccountName,
					"metadata.name":      name,
					"metadata.namespace": d.getAccountNamespace(ctx.AccountName),
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}

				if cp == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}

			return &CheckNameAvailabilityOutput{Result: false, SuggestedNames: fn.GenValidK8sResourceNames(name, 3)}, nil
		}
	case ResTypeNodePool:
		{
			if clusterName == nil {
				return nil, fmt.Errorf("clusterName is required for checking name availability for nodepool")
			}

			if fn.IsValidK8sResourceName(name) {
				cp, err := d.nodePoolRepo.FindOne(ctx, repos.Filter{
					"accountName":   ctx.AccountName,
					"clusterName":   clusterName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}

				if cp == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}

			return &CheckNameAvailabilityOutput{Result: false, SuggestedNames: fn.GenValidK8sResourceNames(name, 3)}, nil
		}
	default:
		{
			return &CheckNameAvailabilityOutput{Result: false}, fmt.Errorf("unknown resource type provided: %q", typeArg)
		}
	}
}
