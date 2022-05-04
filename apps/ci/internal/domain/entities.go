package domain

import (
	"golang.org/x/oauth2"
	"kloudlite.io/pkg/repos"
)

type Pipeline struct {
	repos.BaseEntity     `bson:",inline"`
	Name                 string                 `json:"name,omitempty"`
	ProjectId            string                 `json:"project_id,omitempty" bson:"project_id"`
	ImageName            string                 `json:"image_name,omitempty" bson:"image_name"`
	PipelineEnv          string                 `json:"pipeline_env,omitempty" bson:"pipeline_env"`
	GitProvider          string                 `json:"git_provider,omitempty" bson:"git_provider"`
	GitRepoUrl           string                 `json:"git_repo_url,omitempty" bson:"git_repo_url"`
	DockerFile           *string                `json:"docker_file,omitempty" bson:"docker_file"`
	ContextDir           *string                `json:"context_dir,omitempty" bson:"context_dir"`
	GithubInstallationId *int                   `json:"github_installation_id,omitempty" bson:"github_installation_id"`
	GitlabTokenId        string                 `json:"gitlab_token,omitempty" bson:"gitlab_token_id"`
	GitlabRepoId         *int                   `json:"gitlab_repo_id,omitempty" bson:"gitlab_repo_id"`
	BuildArgs            map[string]interface{} `json:"build_args,omitempty" bson:"build_args"`
	RepoName             string                 `json:"repo_name,omitempty" bson:"repo_name"`
	Metadata             map[string]interface{} `json:"metadata,omitempty" bson:"metadata"`
}

var PipelineIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type HarborAccount struct {
	repos.BaseEntity `bson:",inline"`
	HarborId         int    `json:"harbor_id"`
	ProjectName      string `json:"project_name"`
	Username         string `json:"username"`
	Password         string `json:"password"`
}

type AccessToken struct {
	Id       repos.ID       `json:"_id"`
	UserId   repos.ID       `json:"user_id" bson:"user_id"`
	Email    string         `json:"email" bson:"email"`
	Provider string         `json:"provider" bson:"provider"`
	Token    *oauth2.Token  `json:"token" bson:"token"`
	Data     map[string]any `json:"data" bson:"data"`
}
