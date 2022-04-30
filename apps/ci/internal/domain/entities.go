package domain

import (
	"golang.org/x/oauth2"
	"kloudlite.io/pkg/repos"
)

type Pipeline struct {
	repos.BaseEntity     `bson:",inline"`
	Name                 string                 `json:"name,omitempty"`
	ProjectId            string                 `json:"project_id,omitempty"`
	ImageName            string                 `json:"image_name,omitempty"`
	PipelineEnv          string                 `json:"pipeline_env,omitempty"`
	GitProvider          string                 `json:"git_provider,omitempty"`
	GitRepoUrl           string                 `json:"git_repo_url,omitempty"`
	DockerFile           *string                `json:"docker_file,omitempty"`
	ContextDir           *string                `json:"context_dir,omitempty"`
	GithubInstallationId *int                   `json:"github_installation_id,omitempty"`
	GitlabTokenId        string                 `json:"gitlab_token,omitempty"`
	BuildArgs            map[string]interface{} `json:"build_args,omitempty"`
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
