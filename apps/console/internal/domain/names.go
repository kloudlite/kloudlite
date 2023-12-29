package domain

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) CheckNameAvailability(ctx context.Context, resType ResType, accountName string, namespace *string, name string) (*CheckNameAvailabilityOutput, error) {
	errNamespaceRequired := func() error {
		return errors.Newf("namespace is required for resource type %q", resType)
	}

	switch resType {
	case ResTypeProject:
		{
			if fn.IsValidK8sResourceName(name) {
				p, err := d.projectRepo.FindOne(ctx, repos.Filter{
					"accountName":   accountName,
					"metadata.name": name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
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
	case ResTypeEnvironment, ResTypeWorkspace:
		{
			if namespace == nil {
				return nil, errNamespaceRequired()
			}
			if fn.IsValidK8sResourceName(name) {
				p, err := d.workspaceRepo.FindOne(ctx, repos.Filter{
					"accountName":        accountName,
					"metadata.namespace": namespace,
					"metadata.name":      name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
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
			if namespace == nil {
				return nil, errNamespaceRequired()
			}
			if fn.IsValidK8sResourceName(name) {
				a, err := d.appRepo.FindOne(ctx, repos.Filter{
					"accountName":        accountName,
					"metadata.namespace": namespace,
					"metadata.name":      name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
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
			if namespace == nil {
				return nil, errNamespaceRequired()
			}
			if fn.IsValidK8sResourceName(name) {
				c, err := d.configRepo.FindOne(ctx, repos.Filter{
					"accountName":        accountName,
					"metadata.namespace": namespace,
					"metadata.name":      name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
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
			if namespace == nil {
				return nil, errNamespaceRequired()
			}
			if fn.IsValidK8sResourceName(name) {
				s, err := d.secretRepo.FindOne(ctx, repos.Filter{
					"accountName":        accountName,
					"metadata.namespace": namespace,
					"metadata.name":      name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
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
			if namespace == nil {
				return nil, errNamespaceRequired()
			}
			if fn.IsValidK8sResourceName(name) {
				r, err := d.routerRepo.FindOne(ctx, repos.Filter{
					"accountName":        accountName,
					"metadata.namespace": namespace,
					"metadata.name":      name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
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
			if namespace == nil {
				return nil, errNamespaceRequired()
			}
			if fn.IsValidK8sResourceName(name) {
				r, err := d.mresRepo.FindOne(ctx, repos.Filter{
					"accountName":        accountName,
					"metadata.namespace": namespace,
					"metadata.name":      name,
				})
				if err != nil {
					return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
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
			return nil, errors.Newf("resource type %q is not acknowledged", resType)
		}
	}
}
