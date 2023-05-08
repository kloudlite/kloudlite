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
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeEnvironment:
		{
			p, err := d.environmentRepo.FindOne(ctx, repos.Filter{
				"accountName":   accountName,
				"metadata.name": name,
			})
			if err != nil {
				return &CheckNameAvailabilityOutput{Result: false}, err
			}
			if p == nil {
				return &CheckNameAvailabilityOutput{Result: true}, nil
			}
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeApp:
		{
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
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeConfig:
		{
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
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeSecret:
		{
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
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeRouter:
		{
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
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeManagedService:
		{
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
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	case ResTypeManagedResource:
		{
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
			return &CheckNameAvailabilityOutput{
				Result:         false,
				SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
			}, nil
		}
	default:
		{
			return nil, fmt.Errorf("resource type %q is not acknowledged", resType)
		}
	}
}
