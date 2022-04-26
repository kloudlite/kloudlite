package domain

import "kloudlite.io/pkg/repos"

type Pipeline struct {
	repos.BaseEntity     `bson:",inline"`
	Name                 string            `json:"name,omitempty"`
	ImageName            string            `json:"image_name,omitempty"`
	PipelineEnv          string            `json:"pipeline_env,omitempty"`
	GitProvider          string            `json:"git_provider,omitempty"`
	GitRepoUrl           string            `json:"git_repo_url,omitempty"`
	DockerFile           string            `json:"docker_file,omitempty"`
	ContextDir           string            `json:"context_dir,omitempty"`
	GithubInstallationId string            `json:"github_installation_id,omitempty"`
	GitlabTokenId        string            `json:"gitlab_token,omitempty"`
	BuildArgs            map[string]string `json:"build_args,omitempty"`
}

var PipelineIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
