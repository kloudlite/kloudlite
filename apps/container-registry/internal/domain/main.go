package domain

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
	"k8s.io/utils/strings/slices"
)

type Impl struct {
	repositoryRepo repos.DbRepo[*entities.Repository]
	credentialRepo repos.DbRepo[*entities.Credential]
	buildRepo      repos.DbRepo[*entities.Build]
	buildCacheRepo repos.DbRepo[*entities.BuildCacheKey]
	digestRepo     repos.DbRepo[*entities.Digest]
	buildRunRepo   repos.DbRepo[*entities.BuildRun]
	iamClient      iam.IAMClient
	envs           *env.Env
	logger         logging.Logger
	cacheClient    kv.BinaryDataRepo

	authClient auth.AuthClient

	github                 Github
	gitlab                 Gitlab
	resourceEventPublisher ResourceEventPublisher
	dispatcher             ResourceDispatcher
}

func (d *Impl) ProcessRegistryEvents(ctx context.Context, events []entities.Event, logger logging.Logger) error {
	l := logger.WithName("registry-event")

	pattern := `.*[^\/].*\/.*$`

	re, err := regexp.Compile(pattern)
	if err != nil {
		l.Errorf(err)
		return errors.NewE(err)
	}

	for _, e := range events {

		r := e.Target.Repository

		if !re.MatchString(r) {
			l.Warnf("invalid repository name %s\n, ignoring", r)
			return nil
		}

		rArray := strings.Split(r, "/")

		accountName := rArray[0]
		repoName := strings.Join(rArray[1:], "/")
		tag := e.Target.Tag

		switch e.Request.Method {
		case "PUT":

			if tag == "" {
				fmt.Println("tag is empty with digest", e.Target.Digest)
				return nil
			}

			digest, err := d.digestRepo.FindOne(ctx, repos.Filter{
				"tags": map[string]any{
					"$in": []string{tag},
				},
				"repository":  repoName,
				"accountName": accountName,
			})
			if err != nil {
				return errors.NewE(err)
			}

			if digest == nil {
			} else {
				digest.Tags = func() []string {
					tags := []string{}

					for _, v := range digest.Tags {
						if v != tag {
							tags = append(tags, v)
						}
					}

					return tags
				}()

				if len(digest.Tags) == 0 {
					if err := d.digestRepo.DeleteById(ctx, digest.Id); err != nil {
						d.logger.Errorf(err)
					}
				} else {
					_, err := d.digestRepo.UpdateById(ctx, digest.Id, digest)
					if err != nil {
						return errors.NewE(err)
					}
				}

			}

			digest, err = d.digestRepo.FindOne(ctx, repos.Filter{
				"digest":      e.Target.Digest,
				"repository":  repoName,
				"accountName": accountName,
			})

			if err != nil {
				return errors.NewE(err)
			}

			if digest == nil {
				if _, err := d.digestRepo.Create(ctx, &entities.Digest{
					Tags: func() []string {
						if tag != "" {
							return []string{tag}
						}
						return []string{}
					}(),
					AccountName: accountName,
					Repository:  repoName,
					Actor:       e.Actor.Name,
					Digest:      e.Target.Digest,
					Size:        e.Target.Size,
					Length:      e.Target.Length,
					MediaType:   e.Target.MediaType,
					URL:         e.Target.URL,
				}); err != nil {
					return errors.NewE(err)
				}
			} else {
				if b := slices.Contains(digest.Tags, tag); !b {
					digest.Tags = append(digest.Tags, tag)
					_, err := d.digestRepo.UpdateById(ctx, digest.Id, digest)
					if err != nil {
						return errors.NewE(err)
					}
				}
			}

			ee, err := d.repositoryRepo.FindOne(ctx, repos.Filter{
				"accountName": accountName,
				"name":        repoName,
			})
			if err != nil {
				return errors.NewE(err)
			}

			if ee == nil {
				_, err := d.repositoryRepo.Create(ctx, &entities.Repository{})
				if err != nil {
					return errors.NewE(err)
				}
				return nil
			}

			ee.LastUpdatedBy = common.CreatedOrUpdatedBy{
				UserName: e.Actor.Name,
			}

			if _, err := d.repositoryRepo.UpdateById(ctx, ee.Id, ee); err != nil {
				d.logger.Errorf(err)
			}

		case "DELETE":

			l.Infof("DELETE %s:%s %s", e.Target.Repository, e.Target.Tag, e.Target.Digest)

			if err := d.digestRepo.DeleteOne(ctx, repos.Filter{
				"digest":      e.Target.Digest,
				"repository":  repoName,
				"accountName": accountName,
			}); err != nil {
				d.logger.Errorf(err)
				return errors.NewE(err)
			}

		case "HEAD":
			l.Infof("HEAD %s:%s", e.Target.Repository, e.Target.Tag)

		case "GET":
			l.Infof("GET %s:%s", e.Target.Repository, e.Target.Tag)

		default:
			l.Infof("unhandled method", e.Request.Method)
			return errors.Newf("unhandled method %s", e.Request.Method)
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
			buildCacheRepo repos.DbRepo[*entities.BuildCacheKey],
			tagRepo repos.DbRepo[*entities.Digest],
			buildRunRepo repos.DbRepo[*entities.BuildRun],
			iamClient iam.IAMClient,
			cacheClient kv.BinaryDataRepo,
			authClient auth.AuthClient,
			github Github,
			gitlab Gitlab,
			resourceEventPublisher ResourceEventPublisher,
			dispatcher ResourceDispatcher,
		) (Domain, error) {
			return &Impl{
				repositoryRepo:         repositoryRepo,
				credentialRepo:         credentialRepo,
				iamClient:              iamClient,
				envs:                   e,
				digestRepo:             tagRepo,
				logger:                 logger,
				cacheClient:            cacheClient,
				buildRepo:              buildRepo,
				buildCacheRepo:         buildCacheRepo,
				buildRunRepo:           buildRunRepo,
				authClient:             authClient,
				github:                 github,
				gitlab:                 gitlab,
				resourceEventPublisher: resourceEventPublisher,
				dispatcher:             dispatcher,
			}, nil
		}),
)
