package domain

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/apps/accounts/internal/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	nanoid "github.com/matoous/go-nanoid/v2"
)

func (d *domain) findInvitation(ctx context.Context, accountName string, invitationId repos.ID) (*entities.Invitation, error) {
	inv, err := d.invitationRepo.FindOne(ctx, repos.Filter{
		"accountName": accountName,
		"id":          invitationId,
	})
	if err != nil {
		return nil, err
	}

	if inv == nil {
		return nil, fmt.Errorf("no invitation found with id=%s", invitationId)
	}

	return inv, nil
}

func (d *domain) findInvitationByInviteToken(ctx context.Context, accountName string, userEmail string, inviteToken string) (*entities.Invitation, error) {
	inv, err := d.invitationRepo.FindOne(ctx, repos.Filter{
		"userEmail":   userEmail,
		"accountName": accountName,
		"inviteToken": inviteToken,
	})
	if err != nil {
		return nil, err
	}

	if inv == nil {
		return nil, fmt.Errorf("no invitation found, with given invite token")
	}

	return inv, nil
}

func (d *domain) InviteMembers(ctx UserContext, accountName string, invitations []*entities.Invitation) ([]*entities.Invitation, error) {
	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, iamT.InviteAccountMember); err != nil {
		return nil, err
	}

	_, err := d.findAccount(ctx, accountName)
	if err != nil {
		return nil, err
	}

	results := make([]*entities.Invitation, len(invitations))

	for i := range invitations {
		invitations[i].InviteToken, err = nanoid.New(64)
		if err != nil {
			return nil, errors.NewEf(err, "failed to generate invite token")
		}

		user, err := d.authClient.GetUser(ctx, &auth.GetUserIn{
			UserId: string(ctx.UserId),
		})
		if err != nil {
			return nil, err
		}

		invitations[i].InvitedBy = user.Name
		invitations[i].AccountName = accountName

		inv, err := d.invitationRepo.Create(ctx, invitations[i])
		if err != nil {
			return nil, errors.NewEf(err, "failed to create invitation")
		}

		if _, err := d.commsClient.SendAccountMemberInviteEmail(ctx, &comms.AccountMemberInviteEmailInput{
			AccountName:     inv.AccountName,
			InvitationToken: inv.InviteToken,
			InvitedBy:       inv.InvitedBy,
			Email:           inv.UserEmail,
			// TODO: verify user name, if it is not empty, then use it, otherwise use email
			Name: inv.UserName,
		}); err != nil {
			return nil, err
		}

		results[i] = inv
	}

	return results, nil
}

func (d *domain) ResendInviteEmail(ctx UserContext, accountName string, invitationId repos.ID) (bool, error) {
	inv, err := d.findInvitation(ctx, accountName, invitationId)
	if err != nil {
		return false, err
	}

	action := iamT.InviteAccountMember
	if inv.UserRole == iamT.RoleAccountAdmin {
		action = iamT.InviteAccountAdmin
	}

	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, action); err != nil {
		return false, err
	}

	if _, err := d.commsClient.SendAccountMemberInviteEmail(ctx, &comms.AccountMemberInviteEmailInput{
		AccountName:     accountName,
		InvitationToken: inv.InviteToken,
		InvitedBy:       inv.InvitedBy,
		Email:           inv.UserEmail,
		Name:            accountName,
	}); err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) ListInvitations(ctx UserContext, accountName string) ([]*entities.Invitation, error) {
	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, iamT.ListAccountInvitations); err != nil {
		return nil, err
	}

	return d.invitationRepo.Find(ctx, repos.Query{Filter: repos.Filter{"accountName": accountName}})
}

func (d *domain) ListInvitationsForUser(ctx UserContext, onlyPending bool) ([]*entities.Invitation, error) {
	var filters repos.Filter = map[string]any{
		"userEmail": ctx.UserEmail,
	}

	if onlyPending {
		filters["accepted"] = nil
		filters["rejected"] = nil
	}

	return d.invitationRepo.Find(ctx, repos.Query{Filter: filters})
}

func (d *domain) GetInvitation(ctx UserContext, accountName string, invitationId repos.ID) (*entities.Invitation, error) {
	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, iamT.ListAccountInvitations); err != nil {
		return nil, err
	}

	return d.invitationRepo.FindById(ctx, invitationId)
}

func (d *domain) DeleteInvitation(ctx UserContext, accountName string, invitationId repos.ID) (bool, error) {
	if err := d.checkAccountAccess(ctx, accountName, ctx.UserId, iamT.DeleteAccountInvitation); err != nil {
		return false, err
	}

	inv, err := d.findInvitation(ctx, accountName, invitationId)
	if err != nil {
		return false, err
	}

	if err := d.invitationRepo.DeleteById(ctx, inv.Id); err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) AcceptInvitation(ctx UserContext, accountName string, inviteToken string) (bool, error) {
	inv, err := d.findInvitationByInviteToken(ctx, accountName, ctx.UserEmail, inviteToken)
	if err != nil {
		return false, err
	}

	if inv.Accepted != nil || inv.Rejected != nil {
		return false, fmt.Errorf("invitation already accepted or rejected, won't process further")
	}

	inv.Accepted = fn.New(true)
	if _, err := d.invitationRepo.UpdateById(ctx, inv.Id, inv); err != nil {
		return false, err
	}

	if err := d.addMembership(ctx, accountName, ctx.UserId, inv.UserRole); err != nil {
		return false, err
	}

	return true, nil
}

func (d *domain) RejectInvitation(ctx UserContext, accountName string, inviteToken string) (bool, error) {
	inv, err := d.findInvitationByInviteToken(ctx, accountName, ctx.UserEmail, inviteToken)
	if err != nil {
		return false, err
	}

	if inv.Accepted != nil || inv.Rejected != nil {
		return false, fmt.Errorf("invitation already accepted or rejected, won't process further")
	}

	inv.Rejected = fn.New(true)
	if _, err := d.invitationRepo.UpdateById(ctx, inv.Id, inv); err != nil {
		return false, err
	}

	return true, nil
}
