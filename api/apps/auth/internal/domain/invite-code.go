package domain

import (
	"context"
	"github.com/kloudlite/api/apps/auth/internal/entities"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domainI) findInviteCodeData(ctx context.Context, inviteCode string) (*entities.InviteCode, error) {
	inv, err := d.inviteCodeRepo.FindOne(ctx, repos.Filter{
		"inviteCode": inviteCode,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if inv == nil {
		return nil, errors.Newf("no invite with code=%q found", inviteCode)
	}
	return inv, nil
}

func (d *domainI) CreateInviteCode(ctx context.Context, name string, inviteCode string) (*entities.InviteCode, error) {
	inv, err := d.inviteCodeRepo.Create(ctx, &entities.InviteCode{
		Name:       name,
		InviteCode: inviteCode,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return inv, nil
}

func (d *domainI) DeleteInviteCode(ctx context.Context, invCodeId string) error {
	err := d.inviteCodeRepo.DeleteOne(
		ctx,
		repos.Filter{"id": invCodeId},
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domainI) VerifyInviteCode(ctx context.Context, userId repos.ID, invitationCode string) (bool, error) {
	user, err := d.userRepo.FindById(ctx, userId)
	if err != nil {
		return false, errors.NewE(err)
	}

	inv, err := d.findInviteCodeData(ctx, invitationCode)
	if err != nil {
		return false, errors.NewE(err)
	}

	if inv.InviteCode == invitationCode {
		user.Approved = true
	}

	user, err = d.userRepo.UpdateById(ctx, userId, user)
	if err != nil {
		return false, errors.NewE(err)
	}
	return true, nil
}
