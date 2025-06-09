package domain

import (
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/pkg/errors"
	"k8s.io/utils/strings/slices"
)

func validateTag(tag string) error {
	if tag == "" {
		return errors.Newf("tag is required")
	}

	// re := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)
	//
	// if !re.MatchString(tag) {
	// 	return errors.Newf("tag is invalid")
	// }
	return nil
}

func validateRepository(repository string) error {
	if repository == "" {
		return errors.Newf("repository is required")
	}

	// re := regexp.MustCompile(`\b(?:http(?:s)?://)?(?:[a-zA-Z0-9]+(?::[a-zA-Z0-9]+)?@)?(?:www\.)?(?:github\.com|bitbucket\.org|gitlab\.com)[:/][a-zA-Z0-9_.\-]+(?:/[a-zA-Z0-9_.\-]+)*(?:\.git)?\b`)
	//
	// if !re.MatchString(repository) {
	// 	return errors.Newf("repository is invalid")
	// }
	return nil
}

func validateBranch(branch string) error {
	if branch == "" {
		return errors.Newf("branch is required")
	}

	// re := regexp.MustCompile(`^(?!\/)(?!.*\/\/)(?!.*\@\{)(?!.*\.\.)(?![\.\:~\^?*\[\]\\\/]).*(?<![\.\／])$`)
	//
	// if !re.MatchString(branch) {
	// 	return errors.Newf("branch is invalid")
	// }
	return nil
}

func validateProvider(provider string) error {
	if provider == "" {
		return errors.Newf("provider is required")
	}

	if !slices.Contains([]string{
		string(entities.Github),
		string(entities.Gitlab),
	}, provider) {
		return errors.Newf("provider is invalid")
	}

	return nil
}

func validateName(name string) error {
	if name == "" {
		return errors.Newf("name is required")
	}

	return nil
}

func validateBuild(build entities.Build) error {
	for _, v := range build.Spec.Registry.Repo.Tags {
		if err := validateTag(v); err != nil {
			return errors.NewE(err)
		}
	}

	if err := validateRepository(build.Spec.Registry.Repo.Name); err != nil {
		return errors.NewE(err)
	}

	if err := validateBranch(build.Source.Branch); err != nil {
		return errors.NewE(err)
	}

	if err := validateProvider(string(build.Source.Provider)); err != nil {
		return errors.NewE(err)
	}

	if err := validateName(build.Name); err != nil {
		return errors.NewE(err)
	}

	return nil
}
