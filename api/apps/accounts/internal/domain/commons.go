package domain

import (
	"context"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

type UserContext struct {
	context.Context
	UserId    repos.ID
	UserName  string
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

	if err != nil {
		d.logger.Errorf(err, "iam.can check for action: ", action)
		return errors.Newf("unauthorized to perform action: %s", action)
	}

	if !co.Status {
		return errors.Newf("unauthorized to perform action: %s", action)
	}

	return nil
}
