package domain

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/kloudlite/container-registry-authorizer/admin"
	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

type Impl struct {
	repositoryRepo repos.DbRepo[*entities.Repository]
	credentialRepo repos.DbRepo[*entities.Credential]
	tagRepo        repos.DbRepo[*entities.Tag]
	iamClient      iam.IAMClient
	envs           *env.Env
	logger         logging.Logger
	cacheClient    cache.Client
}

// nonce generates a random string of length size
func nonce(size int) string {
	chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	nonceBytes := make([]byte, size)

	for i := range nonceBytes {
		nonceBytes[i] = chars[rand.Intn(len(chars))]
	}

	return string(nonceBytes)
}

func (d *Impl) GetTokenKey(ctx context.Context, username string, accountname string) (string, error) {

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

	if err := d.cacheClient.Set(ctx, username+"::"+accountname, []byte(c.TokenKey)); err != nil {
		return "", err
	}

	if c == nil {
		return "", fmt.Errorf("credential not found")
	}

	return c.TokenKey, nil
}

func (d *Impl) GetToken(ctx RegistryContext, username string) (string, error) {

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
		return "", fmt.Errorf("unauthorized to get credentials")
	}

	c, err := d.credentialRepo.FindOne(ctx, repos.Filter{
		"username":    username,
		"accountName": ctx.AccountName,
	})
	if err != nil {
		return "", err
	}
	if c == nil {
		return "", fmt.Errorf("credential not found")
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

// CreateCredential implements Domain.
func (d *Impl) CreateCredential(ctx RegistryContext, credential entities.Credential) error {

	pattern := `^([a-z])[a-z0-9_]+$`

	re := regexp.MustCompile(pattern)

	if !re.MatchString(credential.Name) {
		return fmt.Errorf("invalid credential name, must be lowercase alphanumeric with underscore")
	}

	key := nonce(12)
	_, err := d.credentialRepo.Create(ctx, &entities.Credential{
		Name:        credential.Name,
		Access:      credential.Access,
		AccountName: ctx.AccountName,
		UserName:    credential.UserName,
		TokenKey:    key,
		Expiration:  credential.Expiration,
	})
	if err != nil {
		return err
	}
	return nil
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
		return nil, fmt.Errorf("unauthorized to get credentials")
	}

	filter := repos.Filter{"accountName": ctx.AccountName}
	return d.credentialRepo.FindPaginated(ctx, d.credentialRepo.MergeMatchFilters(filter, search), pagination)
}

// DeleteCredential implements Domain.
func (d *Impl) DeleteCredential(ctx RegistryContext, credName string, userName string) error {

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
		return fmt.Errorf("unauthorized to delete credentials")
	}

	err = d.credentialRepo.DeleteOne(ctx, repos.Filter{
		"name":        credName,
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

// CreateRepository implements Domain.
func (d *Impl) CreateRepository(ctx RegistryContext, repoName string) error {

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
		return fmt.Errorf("unauthorized to create repository")
	}

	_, err = d.repositoryRepo.Create(ctx, &entities.Repository{
		Name:        repoName,
		AccountName: ctx.AccountName,
	})
	return err
}

// DeleteRepository implements Domain.
func (d *Impl) DeleteRepository(ctx RegistryContext, repoName string) error {

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
		return fmt.Errorf("unauthorized to delete repository")
	}

	if _, err = d.repositoryRepo.FindOne(ctx, repos.Filter{
		"name":        repoName,
		"accountName": ctx.AccountName,
	}); err != nil {
		return err
	}

	return d.repositoryRepo.DeleteOne(ctx, repos.Filter{"name": repoName, "accountName": ctx.AccountName})
}

// DeleteRepositoryTag implements Domain.
func (d *Impl) DeleteRepositoryTag(ctx RegistryContext, repoName string, tagName string) error {

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
		return fmt.Errorf("unauthorized to delete repository tag")
	}

	return nil
}

// ListRepositories implements Domain.
func (d *Impl) ListRepositories(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Repository], error) {

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
		return nil, fmt.Errorf("unauthorized to list repositories")
	}

	filter := repos.Filter{"accountName": ctx.AccountName}
	return d.repositoryRepo.FindPaginated(ctx, d.repositoryRepo.MergeMatchFilters(filter, search), pagination)
}

// ListRepositoryTags implements Domain.
func (d *Impl) ListRepositoryTags(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Tag], error) {
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
		return nil, fmt.Errorf("unauthorized to list repository tags")
	}

	filter := repos.Filter{"accountName": ctx.AccountName, "repository": repoName}
	return d.tagRepo.FindPaginated(ctx, d.tagRepo.MergeMatchFilters(filter, search), pagination)
}

// ListRepositoryTags implements Domain.

func (d *Impl) ProcessEvents(ctx context.Context, events []entities.Event) error {

	pattern := `.*[^\/].*\/.*$`

	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Println(err)
	}

	for _, e := range events {

		r := e.Target.Repository

		if !re.MatchString(r) {
			return fmt.Errorf("invalid repository name %s", r)
		}

		rArray := strings.Split(r, "/")

		accountName := rArray[0]
		repoName := strings.Join(rArray[1:], "/")
		tag := e.Target.Tag

		switch e.Request.Method {
		case "PUT":

			if _, err := d.repositoryRepo.Upsert(ctx, repos.Filter{
				"name":        repoName,
				"accountName": accountName,
			}, &entities.Repository{
				AccountName: accountName,
				Name:        repoName,
			}); err != nil {
				d.logger.Errorf(err)
				return err
			}

			if _, err = d.tagRepo.Upsert(ctx, repos.Filter{
				"name":        tag,
				"repository":  repoName,
				"accountName": accountName,
			}, &entities.Tag{
				Name:        tag,
				AccountName: accountName,
				Repository:  repoName,
				Actor:       e.Actor.Name,
				Digest:      e.Target.Digest,
				Size:        e.Target.Size,
				Length:      e.Target.Length,
				MediaType:   e.Target.MediaType,
				URL:         e.Target.URL,
				References:  e.Target.References,
			}); err != nil {
				d.logger.Errorf(err)
				return err
			}

		default:
			log.Println("unhandled method", e.Request.Method)
			return nil
		}

	}
	return nil
}

var Module = fx.Module(
	"domain",
	fx.Provide(
		func(e *env.Env,
			logger logging.Logger,
			repositoryRepo repos.DbRepo[*entities.Repository],
			credentialRepo repos.DbRepo[*entities.Credential],
			tagRepo repos.DbRepo[*entities.Tag],
			iamClient iam.IAMClient,
			cacheClient cache.Client,
		) (Domain, error) {
			return &Impl{
				repositoryRepo: repositoryRepo,
				credentialRepo: credentialRepo,
				iamClient:      iamClient,
				envs:           e,
				tagRepo:        tagRepo,
				logger:         logger,
				cacheClient:    cacheClient,
			}, nil
		}),
)
