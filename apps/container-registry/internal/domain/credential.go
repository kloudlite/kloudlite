package domain

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/pkg/errors"
	"regexp"
	"time"

	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/container-registry-authorizer/admin"
)

const (
	KL_ADMIN = "kloudlite"
)

func (d *Impl) GetTokenKey(ctx context.Context, username string, accountname string) (string, error) {

	if username == KL_ADMIN {
		return accountname, nil
	}

	b, err := d.cacheClient.Get(ctx, username+"::"+accountname)
	if err == nil {
		return string(b), nil
	}

	c, err := d.credentialRepo.FindOne(ctx, repos.Filter{
		"username":    username,
		"accountName": accountname,
	})
	if err != nil {
		return "", err
	}

	if c == nil {
		return "", errors.Newf("credential not found")
	}

	if err := d.cacheClient.SetWithExpiry(ctx, username+"::"+accountname, []byte(c.TokenKey), time.Minute*5); err != nil {
		return "", err
	}

	if c == nil {
		return "", errors.Newf("credential not found")
	}

	return c.TokenKey, nil
}

func (d *Impl) GetToken(ctx RegistryContext, username string) (string, error) {

	if username == KL_ADMIN {
		return "", errors.Newf("invalid credential name, %s is reserved", KL_ADMIN)
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.GetAccount),
	})

	if err != nil {
		return "", err
	}

	if !co.Status {
		return "", errors.Newf("unauthorized to get credentials")
	}

	c, err := d.credentialRepo.FindOne(ctx, repos.Filter{
		"username":    username,
		"accountName": ctx.AccountName,
	})
	if err != nil {
		return "", err
	}
	if c == nil {
		return "", errors.Newf("credential not found")
	}

	i, err := admin.GetExpirationTime(fmt.Sprintf("%d%s", c.Expiration.Value, c.Expiration.Unit))

	if err != nil {
		return "", err
	}

	token, err := admin.GenerateToken(c.UserName, ctx.AccountName, string(c.Access), i, d.envs.RegistrySecretKey+c.TokenKey)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (d *Impl) CheckUserNameAvailability(ctx RegistryContext, username string) (*CheckNameAvailabilityOutput, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.GetAccount),
	})

	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, errors.Newf("unauthorized to check username availability")
	}

	c, err := d.credentialRepo.FindOne(ctx, repos.Filter{
		"username":    username,
		"accountName": ctx.AccountName,
	})
	if err != nil {
		return nil, err
	}

	if c != nil {
		return &CheckNameAvailabilityOutput{
			SuggestedNames: generateUserNames(username, 5),
			Result:         false,
		}, nil
	}

	if isValidUserName(username) == nil {
		return &CheckNameAvailabilityOutput{
			Result: true,
		}, nil
	}

	return &CheckNameAvailabilityOutput{
		Result:         false,
		SuggestedNames: generateUserNames(username, 5),
	}, nil
}

// CreateCredential implements Domain.
func (d *Impl) CreateCredential(ctx RegistryContext, credential entities.Credential) (*entities.Credential, error) {

	pattern := `^([a-z])[a-z0-9_]+$`

	re := regexp.MustCompile(pattern)

	if !re.MatchString(credential.UserName) {
		return nil, errors.Newf("invalid credential name, must be lowercase alphanumeric with underscore")
	}

	if credential.UserName == KL_ADMIN {
		return nil, errors.Newf("invalid credential name, %s is reserved", KL_ADMIN)
	}

	key := Nonce(12)

	return d.credentialRepo.Create(ctx, &entities.Credential{
		Name:        credential.Name,
		Access:      credential.Access,
		AccountName: ctx.AccountName,
		UserName:    credential.UserName,
		TokenKey:    key,
		Expiration:  credential.Expiration,
		CreatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})
}

// ListCredentials implements Domain.
func (d *Impl) ListCredentials(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Credential], error) {

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.GetAccount),
	})

	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, errors.Newf("unauthorized to get credentials")
	}

	filter := repos.Filter{"accountName": ctx.AccountName}
	return d.credentialRepo.FindPaginated(ctx, d.credentialRepo.MergeMatchFilters(filter, search), pagination)
}

// DeleteCredential implements Domain.
func (d *Impl) DeleteCredential(ctx RegistryContext, userName string) error {

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.UpdateAccount),
	})

	if err != nil {
		return err
	}

	if !co.Status {
		return errors.Newf("unauthorized to delete credentials")
	}

	err = d.credentialRepo.DeleteOne(ctx, repos.Filter{
		"username":    userName,
		"accountName": ctx.AccountName,
	})
	if err != nil {
		return err
	}

	if _, err = d.cacheClient.Get(ctx, userName+"::"+ctx.AccountName); err != nil {
		return nil
	}

	return d.cacheClient.Drop(ctx, userName+"::"+ctx.AccountName)
}
