package domain

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"

	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) CheckNameAvailability(ctx context.Context, name string) (*CheckNameAvailabilityOutput, error) {
	if fn.IsValidK8sResourceName(name) {
		p, err := d.accountRepo.FindOne(ctx, repos.Filter{
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
