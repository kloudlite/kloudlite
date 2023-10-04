package domain

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/cache"
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

	authClient auth.AuthClient

	github Github
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
			authClient auth.AuthClient,
			github Github,
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
				authClient:     authClient,
				github:         github,
			}, nil
		}),
)
