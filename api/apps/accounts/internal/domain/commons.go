package domain

import (
	"context"
	"fmt"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/repos"
)

type UserContext struct {
	context.Context
	UserId    repos.ID
	UserEmail string
}

func (d *domain) checkAccountAccess(ctx context.Context, accountName string, userId repos.ID, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId:       string(userId),
		ResourceRefs: []string{iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName)},
		Action:       string(action),
	})

	// if err != nil {
	// return err
	// }

	if err != nil || !co.Status {
		return fmt.Errorf("unauthorized to perform action: %s", action)
	}

	return nil
}
