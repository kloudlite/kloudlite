package domain

import (
	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	fc "github.com/kloudlite/api/apps/comms/internal/domain/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *Impl) isMailAddressValid(ctx CommsContext, id, mailAddress string) bool {
	if mailAddress == "" {
		return false
	}

	if id == "" {
		return false
	}

	s, err := d.subscriptionRepo.FindOne(ctx, repos.Filter{
		fc.SubscriptionMailAddress: mailAddress,
		fields.Id:                  id,
	})

	if err != nil {
		return false
	}

	return s != nil
}

func (d *Impl) UpdateSubscriptionConfig(ctx CommsContext, id repos.ID, config entities.Subscription) (*entities.Subscription, error) {
	if config.MailAddress == "" {
		return nil, errors.NewE(errors.New("mail address is required"))
	}

	b := d.isMailAddressValid(ctx, string(id), config.MailAddress)
	if !b {
		return nil, errors.NewE(errors.New("mail address is not valid"))
	}

	return d.subscriptionRepo.UpdateOne(ctx, repos.Filter{
		fields.Id: id,
	}, &config)
}

func (d *Impl) GetSubscriptionConfig(ctx CommsContext, id repos.ID) (*entities.Subscription, error) {
	return d.subscriptionRepo.FindOne(ctx, repos.Filter{fields.Id: id})
}
