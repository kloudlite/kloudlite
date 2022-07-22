package domain

import (
	"context"
	internal_crds "kloudlite.io/apps/console/internal/domain/op-crds/internal-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) SetupAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	err := d.workloadMessenger.SendAction("apply", string(accountID), &internal_crds.Account{
		APIVersion: internal_crds.AccountAPIVersion,
		Kind:       internal_crds.AccountKind,
		Metadata: internal_crds.AccountMetadata{
			Name: string(accountID),
			Annotations: map[string]string{
				"kloudlite.io/account-ref": string(accountID),
			},
		},
		Spec: internal_crds.AccountSpec{
			AccountId: string(accountID),
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
