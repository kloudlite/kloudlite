package domain

import (
	"fmt"

	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"k8s.io/utils/strings/slices"
)

func validateTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag is required")
	}

	// re := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)
	//
	// if !re.MatchString(tag) {
	// 	return fmt.Errorf("tag is invalid")
	// }
	return nil
}

func validateRepository(repository string) error {
	if repository == "" {
		return fmt.Errorf("repository is required")
	}

	// re := regexp.MustCompile(`\b(?:http(?:s)?://)?(?:[a-zA-Z0-9]+(?::[a-zA-Z0-9]+)?@)?(?:www\.)?(?:github\.com|bitbucket\.org|gitlab\.com)[:/][a-zA-Z0-9_.\-]+(?:/[a-zA-Z0-9_.\-]+)*(?:\.git)?\b`)
	//
	// if !re.MatchString(repository) {
	// 	return fmt.Errorf("repository is invalid")
	// }
	return nil
}

func validateBranch(branch string) error {
	if branch == "" {
		return fmt.Errorf("branch is required")
	}

	// re := regexp.MustCompile(`^(?!\/)(?!.*\/\/)(?!.*\@\{)(?!.*\.\.)(?![\.\:~\^?*\[\]\\\/]).*(?<![\.\／])$`)
	//
	// if !re.MatchString(branch) {
	// 	return fmt.Errorf("branch is invalid")
	// }
	return nil
}

func validateProvider(provider string) error {
	if provider == "" {
		return fmt.Errorf("provider is required")
	}

	if !slices.Contains([]string{
		string(entities.Github),
		string(entities.Gitlab),
	}, provider) {
		return fmt.Errorf("provider is invalid")
	}

	return nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}

	return nil
}

func validateBuild(build entities.Build) error {
	for _, v := range build.Spec.Registry.Repo.Tags {
		if err := validateTag(v); err != nil {
			return err
		}
	}

	if err := validateRepository(build.Spec.Registry.Repo.Name); err != nil {
		return err
	}

	if err := validateBranch(build.Source.Branch); err != nil {
		return err
	}

	if err := validateProvider(string(build.Source.Provider)); err != nil {
		return err
	}

	if err := validateName(build.Name); err != nil {
		return err
	}

	return nil
}
