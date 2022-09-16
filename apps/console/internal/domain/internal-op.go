package domain

import (
	"context"
	"fmt"

	internal_crds "kloudlite.io/apps/console/internal/domain/op-crds/internal-crds"
	kldns "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/dns"
	"kloudlite.io/pkg/repos"
)

func (d *domain) SetupAccount(ctx context.Context, accountID repos.ID) (bool, error) {
	domainsOut, err := d.dnsClient.GetAccountDomains(ctx, &kldns.GetAccountDomainsIn{AccountId: string(accountID)})
	if err != nil {
		return false, err
	}

	err = d.workloadMessenger.SendAction("apply", string(accountID), &internal_crds.Account{
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
			OwnedDomains: func() []string {
				st := make([]string, 0)
				for _, v := range domainsOut.Domains {
					st = append(st, fmt.Sprintf("*.%s", v))
				}
				return st
			}(),
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
