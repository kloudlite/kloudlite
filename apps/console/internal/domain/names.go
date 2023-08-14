package domain

import (
	"context"
	"fmt"

	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CheckNameAvailability(ctx context.Context, resType ResType, accountName string, name string) (*CheckNameAvailabilityOutput, error) {
	switch resType {
	case ResTypeProject:
		{
			if fn.IsValidK8sResourceName(name) {
				p, err := d.projectRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if p == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: fn.GenValidK8sResourceNames(name, 3),
			}, nil
		}
	case ResTypeEnvironment:
		{
			if fn.IsValidK8sResourceName(name) {
				p, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if p == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: fn.GenValidK8sResourceNames(name, 3),
			}, nil
		}
	case ResTypeApp:
		{
			if fn.IsValidK8sResourceName(name) {
				a, err := d.appRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if a == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeConfig:
		{
			if fn.IsValidK8sResourceName(name) {
				c, err := d.configRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if c == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: fn.GenValidK8sResourceNames(name, 3),
			}, nil
		}
	case ResTypeSecret:
		{
			if fn.IsValidK8sResourceName(name) {
				s, err := d.secretRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if s == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: fn.GenValidK8sResourceNames(name, 3),
			}, nil
		}
	case ResTypeRouter:
		{
			if fn.IsValidK8sResourceName(name) {
				r, err := d.routerRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if r == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: fn.GenValidK8sResourceNames(name, 3),
			}, nil
		}
	case ResTypeManagedService:
		{
			if fn.IsValidK8sResourceName(name) {
				r, err := d.msvcRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if r == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: fn.GenValidK8sResourceNames(name, 3),
			}, nil
		}
	case ResTypeManagedResource:
		{
			if fn.IsValidK8sResourceName(name) {
				r, err := d.mresRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, err
				}
				if r == nil {
					return &CheckNameAvailabilityOutput{Result: true}, nil
				}
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: fn.GenValidK8sResourceNames(name, 3),
			}, nil
		}
	default:
		{
			return nil, fmt.Errorf("resource type %q is not acknowledged", resType)
		}
	}
}
