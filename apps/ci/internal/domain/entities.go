package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"kloudlite.io/common"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type ContainerImageBuild struct {
	BaseImage string `json:"base_image,omitempty" bson:"base_image,omitempty"`
	Cmd       string `json:"cmd,omitempty" bson:"cmd,omitempty"`
}

type ContainerImageRun struct {
	BaseImage string `json:"base_image,omitempty" bson:"base_image,omitempty"`
	Cmd       string `json:"cmd,omitempty" bson:"cmd,omitempty"`
}

type ArtifactRef struct {
	DockerImageName string `json:"docker_image_name,omitempty" bson:"docker_image_name,omitempty"`
	DockerImageTag  string `json:"docker_image_tag,omitempty" bson:"docker_image_tag,omitempty"`
}

type Pipeline struct {
	repos.BaseEntity `bson:",inline"`
	Name             string `json:"name,omitempty" bson:"name"`
	ProjectId        string `json:"project_id,omitempty" bson:"project_id"`

	GitProvider string `json:"git_provider,omitempty" bson:"git_provider"`
	GitRepoUrl  string `json:"git_repo_url,omitempty" bson:"git_repo_url"`
	GitBranch   string `json:"git_branch" bson:"git_branch"`

	GitlabTokenId string `json:"gitlab_token,omitempty" bson:"gitlab_token_id"`

	Build ContainerImageBuild `json:"build,omitempty" bson:"build,omitempty"`
	Run   ContainerImageRun   `json:"run,omitempty" bson:"run,omitempty"`

	ArtifactRef ArtifactRef `json:"artifact_ref,omitempty" bson:"artifact_ref,omitempty"`

	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata"`
}

type TektonVars struct {
	GitRepo       string `json:"git-repo"`
	GitUser       string `json:"git-user"`
	GitPassword   string `json:"git-password"`
	GitRef        string `json:"git-ref"`
	GitCommitHash string `json:"git-commit_hash"`

	BuildBaseImage string `json:"build-base_image"`
	BuildCmd       string `json:"build-cmd"`

	RunBaseImage string `json:"run-base_image"`
	RunCmd       string `json:"run-cmd"`

	ArtifactDockerImageName string `json:"artifact_ref-docker_image_name"`
	ArtifactDockerImageTag  string `json:"artifact_ref-docker_image_tag"`
}

func (t *TektonVars) ToJson() (map[string]any, error) {
	marshal, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(marshal, &m); err != nil {
		return nil, err
	}
	return m, nil
}

const (
	gitlabWebhook string = "https://webhooks.dev.madhouselabs.io/gitlab"
	githubWebhook string = "https://webhooks.dev.madhouselabs.io/github"
)

func (p *Pipeline) TriggerHook() error {
	var req *http.Request
	if p.GitProvider == common.ProviderGithub {
		body := t.M{
			"ref":   fmt.Sprintf("refs/heads/%s", p.GitBranch),
			"after": "",
			"repository": t.M{
				"html_url": p.GitRepoUrl,
			},
		}

		b, err := json.Marshal(body)
		if err != nil {
			return errors.ErrMarshal(err)
		}
		req, err = http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s?pipelineId=%s", githubWebhook, p.Id),
			bytes.NewBuffer(b),
		)
		if err != nil {
			return errors.NewEf(err, "could not build http request")
		}
	}

	if p.GitProvider == common.ProviderGitlab {

		body := t.M{
			"ref":          fmt.Sprintf("refs/heads/%s", p.GitBranch),
			"checkout_sha": "",
			"repository": t.M{
				"git_http_url": p.GitRepoUrl,
			},
		}
		b, err := json.Marshal(body)
		if err != nil {
			return errors.ErrMarshal(err)
		}
		req, err = http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s?pipelineId=%s", gitlabWebhook, p.Id),
			bytes.NewBuffer(b),
		)
		if err != nil {
			return errors.NewEf(err, "could not build http request")
		}
	}

	if req != nil {
		fmt.Println("HERE......")
		r, err := http.DefaultClient.Do(req)
		fmt.Printf("r: %+v | err: %v\n", r, err)
		if err != nil {
			return errors.NewEf(err, "while making request")
		}
		fmt.Println("HERE response......")
		if r.StatusCode == http.StatusAccepted {
			return nil
		}
		return errors.Newf("trigger for repo=%s failed as received StatusCode=%s", p.GitRepoUrl, r.StatusCode)
	}
	return errors.Newf("unknown gitProvider=%s, aborting trigger", p.GitProvider)
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
