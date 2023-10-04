package domain

import (
	"fmt"
	"log"
	"regexp"

	"kloudlite.io/apps/container-registry/internal/domain/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/docker"
	"kloudlite.io/pkg/repos"
)

// CreateRepository implements Domain.
func (d *Impl) CreateRepository(ctx RegistryContext, repoName string) (*entities.Repository, error) {

	pattern := `^[a-z0-9]+([._/-][a-z0-9]+)*$`

	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Println(err)
	}

	if !re.MatchString(repoName) {
		return nil, fmt.Errorf("invalid repository name, must be lowercase alphanumeric with underscore")
	}

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
