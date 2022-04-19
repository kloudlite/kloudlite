package entities

import "kloudlite.io/pkg/repos"

type Pipeline struct {
	repos.BaseEntity `bson:",inline"`
}

type GitProvider string

const (
	GitHub GitProvider = "github"
	GitLab GitProvider = "gitlab"
)

type ProviderData struct {
	Provider       GitProvider `json:"provider" bson:"provider"`
	Token          string      `json:"token" bson:"token"`
	InstallationId string      `json:"installation_id" bson:"installation_id"`
}

type PipelineStatus string

const (
	PipelineStateSyncing = PipelineStatus("sync-in-progress")
	PipelineStateLive    = PipelineStatus("live")
	PipelineStateError   = PipelineStatus("error")
	PipelineStateDown    = PipelineStatus("down")
)

type PipelineBuild struct {
	repos.BaseEntity `bson:",inline"`
	Name             string            `json:"name" bson:"name"`
	GitRepoUrl       string            `json:"git_repo_url" bson:"git_repo_url"`
	GitRef           string            `json:"git_ref" bson:"git_ref"`
	BuildArgs        map[string]string `json:"build_args" bson:"build_args"`
	DockerFile       string            `json:"docker_file" bson:"docker_file"`
	ContextDir       string            `json:"context_dir" bson:"context_dir"`
	PullSecret       string            `json:"pull_secret" bson:"pull_secret"`
	WebhookId        repos.ID          `json:"webhook_id" bson:"webhook_id"`
	Status           PipelineStatus    `json:"status" bson:"status"`
}
