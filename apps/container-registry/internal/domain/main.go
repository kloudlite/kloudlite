package domain

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/kloudlite/container-registry-authorizer/admin"
	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/docker"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
)

type Impl struct {
	repositoryRepo repos.DbRepo[*entities.Repository]
	credentialRepo repos.DbRepo[*entities.Credential]
	buildRepo      repos.DbRepo[*entities.Build]
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

	if c == nil {
		return "", fmt.Errorf("credential not found")
	}

	if err := d.cacheClient.SetWithExpiry(ctx, username+"::"+accountname, []byte(c.TokenKey), time.Minute*5); err != nil {
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
func (d *Impl) CreateCredential(ctx RegistryContext, credential entities.Credential) (*entities.Credential, error) {

	pattern := `^([a-z])[a-z0-9_]+$`

	re := regexp.MustCompile(pattern)

	if !re.MatchString(credential.UserName) {
		return nil, fmt.Errorf("invalid credential name, must be lowercase alphanumeric with underscore")
	}

	key := nonce(12)

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
		return nil, fmt.Errorf("unauthorized to get credentials")
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
		return fmt.Errorf("unauthorized to delete credentials")
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

// CreateRepository implements Domain.
func (d *Impl) CreateRepository(ctx RegistryContext, repoName string) (*entities.Repository, error) {

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
		return nil, fmt.Errorf("unauthorized to create repository")
	}

	return d.repositoryRepo.Create(ctx, &entities.Repository{
		Name:        repoName,
		AccountName: ctx.AccountName,
		CreatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})
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

	res, err := d.tagRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"repository":  repoName,
			"accountName": ctx.AccountName,
		},
	})

	if err != nil {
		return err
	}

	if len(res) > 0 {
		return fmt.Errorf("repository %s is not empty, please delete all tags first", repoName)
	}

	return d.repositoryRepo.DeleteOne(ctx, repos.Filter{"name": repoName, "accountName": ctx.AccountName})
}

// DeleteRepositoryTag implements Domain.
func (d *Impl) DeleteRepositoryTag(ctx RegistryContext, repoName string, digest string) error {

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

	e, err := d.tagRepo.FindOne(ctx, repos.Filter{
		"digest":      digest,
		"repository":  repoName,
		"accountName": ctx.AccountName,
	})

	if err != nil {
		return err
	}

	if e == nil {
		return fmt.Errorf("%s not found in repository %s", digest, repoName)
	}

	dockerCli := docker.NewDockerClient(d.envs.RegistryUrl)
	if err := dockerCli.DeleteTag(fmt.Sprintf("%s/%s", ctx.AccountName, repoName), e.Digest); err != nil {
		return err
	}

	if _, err = d.tagRepo.Upsert(ctx, repos.Filter{
		"digest": digest,
	}, &entities.Tag{
		Deleting: true,
	}); err != nil {
		return err
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
		return nil, fmt.Errorf("unauthorized to check username availability")
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

func (d *Impl) AddBuild(ctx RegistryContext, build entities.Build) (*entities.Build, error) {
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
		return nil, fmt.Errorf("unauthorized to add build")
	}

	return d.buildRepo.Create(ctx, &entities.Build{
		Name:        build.Name,
		AccountName: ctx.AccountName,
		Repository:  build.Repository,
		Source:      build.Source,
		Tag:         build.Tag,
		CreatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})
}

func (d *Impl) UpdateBuild(ctx RegistryContext, id repos.ID, build entities.Build) (*entities.Build, error) {
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
		return nil, fmt.Errorf("unauthorized to update build")
	}

	return d.buildRepo.UpdateById(ctx, id, &entities.Build{
		Name:        build.Name,
		AccountName: ctx.AccountName,
		Repository:  build.Repository,
		Source:      build.Source,
		Tag:         build.Tag,
		LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})
}

func (d *Impl) ListBuilds(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Build], error) {
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
		return nil, fmt.Errorf("unauthorized to list builds")
	}

	filter := repos.Filter{"accountName": ctx.AccountName, "repository": repoName}

	return d.buildRepo.FindPaginated(ctx, d.buildRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *Impl) GetBuild(ctx RegistryContext, buildId repos.ID) (*entities.Build, error) {
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
		return nil, fmt.Errorf("unauthorized to get build")
	}

	b, err := d.buildRepo.FindOne(ctx, repos.Filter{"accountName": ctx.AccountName, "id": buildId})
	if err != nil {
		return nil, err
	}

	if b == nil {
		return nil, fmt.Errorf("build not found")
	}

	return b, nil
}

func (d *Impl) DeleteBuild(ctx RegistryContext, buildId repos.ID) error {
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
		return fmt.Errorf("unauthorized to delete build")
	}

	return d.buildRepo.DeleteOne(ctx, repos.Filter{"accountName": ctx.AccountName, "id": buildId})
}

func (d *Impl) TriggerBuild(ctx RegistryContext, buildId repos.ID) error {
	panic("implement me")
}

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

			t, err := d.tagRepo.FindOne(ctx, repos.Filter{
				"digest":      e.Target.Digest,
				"repository":  repoName,
				"accountName": accountName,
			})
			if err != nil {
				return err
			}

			if t == nil {
				if _, err := d.tagRepo.Create(ctx, &entities.Tag{
					Tags:        []string{tag},
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
					return err
				}
			} else {

				if _, err = d.tagRepo.Upsert(ctx, repos.Filter{
					"digest":      e.Target.Digest,
					"repository":  repoName,
					"accountName": accountName,
				}, &entities.Tag{
					Tags: func() []string {
						tags := []string{}
						for _, v := range t.Tags {
							if v == tag {
								return t.Tags
							}
						}
						tags = append(t.Tags, tag)
						return tags
					}(),
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

			}

		case "DELETE":

			if err := d.tagRepo.DeleteOne(ctx, repos.Filter{
				"digest":      e.Target.Digest,
				"repository":  repoName,
				"accountName": accountName,
			}); err != nil {
				d.logger.Errorf(err)
				return err
			}

		case "HEAD":
			log.Printf("HEAD %s:%s", e.Target.Repository, e.Target.Tag)

		case "GET":
			log.Printf("GET %s:%s", e.Target.Repository, e.Target.Tag)

		default:
			log.Println("unhandled method", e.Request.Method)
			return fmt.Errorf("unhandled method %s", e.Request.Method)
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
			buildRepo repos.DbRepo[*entities.Build],
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
				buildRepo:      buildRepo,
			}, nil
		}),
)
