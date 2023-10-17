package domain

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"go.uber.org/fx"
	"k8s.io/utils/strings/slices"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/common"
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
	digestRepo     repos.DbRepo[*entities.Digest]
	iamClient      iam.IAMClient
	envs           *env.Env
	logger         logging.Logger
	cacheClient    cache.Client

	authClient auth.AuthClient

	github Github
	gitlab Gitlab
}

func (d *Impl) ProcessRegistryEvents(ctx context.Context, events []entities.Event) error {

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
				return err
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
					d.digestRepo.DeleteById(ctx, digest.Id)
				} else {
					_, err := d.digestRepo.UpdateById(ctx, digest.Id, digest)
					if err != nil {
						return err
					}
				}

			}

			digest, err = d.digestRepo.FindOne(ctx, repos.Filter{
				"digest": e.Target.Digest,
			})

			if err != nil {
				return err
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
					return err
				}
			} else {

				if b := slices.Contains(digest.Tags, tag); !b {
					digest.Tags = append(digest.Tags, tag)
					_, err := d.digestRepo.UpdateById(ctx, digest.Id, digest)
					if err != nil {
						return err
					}
				}

				// if tag != "" {
				// 	digest.Tags = append(digest.Tags, tag)
				// }

			}

			ee, err := d.repositoryRepo.FindOne(ctx, repos.Filter{
				"accountName": accountName,
				"name":        repoName,
			})

			ee.LastUpdatedBy = common.CreatedOrUpdatedBy{
				UserName: e.Actor.Name,
			}

			if err != nil {
				return err
			} else {

				_, err := d.repositoryRepo.UpdateById(ctx, ee.Id, ee)
				if err != nil {
					log.Println(err)
				}
			}

		case "DELETE":

			log.Printf("DELETE %s:%s %s", e.Target.Repository, e.Target.Tag, e.Target.Digest)

			if err := d.digestRepo.DeleteOne(ctx, repos.Filter{
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
			tagRepo repos.DbRepo[*entities.Digest],
			iamClient iam.IAMClient,
			cacheClient cache.Client,
			authClient auth.AuthClient,
			github Github,
			gitlab Gitlab,
		) (Domain, error) {
			return &Impl{
				repositoryRepo: repositoryRepo,
				credentialRepo: credentialRepo,
				iamClient:      iamClient,
				envs:           e,
				digestRepo:     tagRepo,
				logger:         logger,
				cacheClient:    cacheClient,
				buildRepo:      buildRepo,
				authClient:     authClient,
				github:         github,
				gitlab:         gitlab,
			}, nil
		}),
)
