package domain

import (
	"context"
	"github.com/kloudlite/api/apps/webhook/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/pkg/errors"
	"go.uber.org/fx"
)

type domain struct {
	env         *env.Env
	commsClient comms.CommsClient
}

func (d *domain) SendContactUsEmail(ctx context.Context, contactUsData *ContactUsData) error {
	_, err := d.commsClient.SendContactUsEmail(ctx, &comms.SendContactUsEmailInput{
		Email:        contactUsData.Email,
		Name:         contactUsData.Name,
		CompanyName:  contactUsData.CompanyName,
		Country:      contactUsData.Country,
		Message:      contactUsData.Message,
		MobileNumber: contactUsData.MobileNumber,
	})
	if err != nil {
		return errors.NewE(err)
	}

	return nil
}

var Module = fx.Module("domain",
	fx.Provide(func(
		env *env.Env,
		commsClient comms.CommsClient,
	) (Domain, error) {
		return &domain{
			env:         env,
			commsClient: commsClient,
		}, nil
	}),
)
