package domain

import (
	"context"

	field_constants "github.com/kloudlite/api/apps/comms/internal/domain/entities/field-constants"
	"github.com/kloudlite/api/apps/comms/types"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *Impl) ListNotifications(ctx CommsContext, pagination repos.CursorPagination) (*repos.PaginatedRecord[*types.Notification], error) {
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

	n, err := d.notificationRepo.FindPaginated(ctx, repos.Filter{
		fields.AccountName: ctx.AccountName,
	}, pagination)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (d *Impl) MarkNotificationAsRead(ctx CommsContext, id repos.ID) (*types.Notification, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId:       string(ctx.UserId),
		ResourceRefs: []string{iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName)},
		Action:       string(iamT.UpdateAccount),
	})

	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, errors.NewE(errors.Newf("user %s does not have permission to update account %s", ctx.UserId, ctx.AccountName))
	}

	xnotf, err := d.notificationRepo.FindOne(ctx, repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.Id:          id,
	})
	if err != nil {
		return nil, err
	}

	if xnotf == nil {
		return nil, errors.NewE(errors.Newf("notification with id %s not found", id))
	}

	xnotf.Read = true

	n, err := d.notificationRepo.UpdateById(ctx, id, xnotf)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (d *Impl) MarkAllNotificationsAsRead(ctx CommsContext) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId:       string(ctx.UserId),
		ResourceRefs: []string{iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName)},
		Action:       string(iamT.UpdateAccount),
	})

	if err != nil {
		return err
	}

	if !co.Status {
		return errors.NewE(errors.Newf("user %s does not have permission to update account %s", ctx.UserId, ctx.AccountName))
	}

	if err := d.notificationRepo.UpdateMany(ctx, repos.Filter{
		fields.AccountName: ctx.AccountName,
	}, map[string]any{
		field_constants.NotificationRead: true,
	}); err != nil {
		return err
	}

	return nil

}

func (d *Impl) Notify(ctx context.Context, notification *types.Notification) error {
	_, err := d.notificationRepo.Create(ctx, notification)
	if err != nil {
		return err
	}

	nc, err := d.notificationConfigRepo.FindOne(ctx, repos.Filter{})
	if err != nil {
		return err
	}

	if nc == nil {
		return errors.NewE(errors.Newf("notification config not found"))
	}

	np := newNotificationProcessor(context.Background(), d, notification, nc)
	if err := np.Send(); err != nil {
		return err
	}

	return nil
}
