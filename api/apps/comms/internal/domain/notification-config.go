package domain

import (
	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *Impl) GetNotificationConfig(ctx CommsContext) (*entities.NotificationConf, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId:       string(ctx.UserId),
		ResourceRefs: []string{iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName)},
		Action:       string(iamT.GetAccount),
	})
	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, errors.NewE(errors.Newf("user %s does not have permission to get account %s", ctx.UserId, ctx.AccountName))
	}

	nc, err := d.notificationConfigRepo.FindOne(ctx, repos.Filter{})
	if err != nil {
		return nil, err
	}

	if nc == nil {
		n, err := d.notificationConfigRepo.Create(ctx, &entities.NotificationConf{
			CreatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
			AccountName: ctx.AccountName,
		})
		if err != nil {
			return nil, err
		}

		nc = n
	}

	return nc, nil
}

func (d *Impl) UpdateNotificationConfig(ctx CommsContext, config entities.NotificationConf) (*entities.NotificationConf, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.UpdateAccount),
	})
	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, errors.NewE(errors.Newf("user %s does not have permission to update account %s", ctx.UserId, ctx.AccountName))
	}

	xnc, err := d.notificationConfigRepo.FindOne(ctx, repos.Filter{
		fields.AccountName: ctx.AccountName,
	})
	if err != nil {
		return nil, err
	}

	if xnc == nil {
		return nil, errors.NewE(errors.Newf("notification config not found"))
	}

	// TODO:(@abdheshnayak) - check for subscription

	xnc.Email = config.Email
	xnc.Slack = config.Slack
	xnc.Telegram = config.Telegram
	xnc.Webhook = config.Webhook

	return d.notificationConfigRepo.UpdateById(ctx, xnc.Id, xnc)
}
