package entities

import (
	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type (
	GitProvider string
	BuildStatus string
)

const (
	Github GitProvider = "github"
	Gitlab GitProvider = "gitlab"
)

const (
	BuildStatusPending BuildStatus = "pending"
	BuildStatusQueued  BuildStatus = "queued"
	BuildStatusRunning BuildStatus = "running"
	BuildStatusSuccess BuildStatus = "success"
	BuildStatusFailed  BuildStatus = "failed"
	BuildStatusError   BuildStatus = "error"
	BuildStatusIdle    BuildStatus = "idle"
)

type GitSource struct {
	Repository string      `json:"repository"`
	Branch     string      `json:"branch"`
	Provider   GitProvider `json:"provider"`
	WebhookId  *int        `json:"webhookId" graphql:"noinput"`
}

type Build struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	Name             string                    `json:"name"`
	CreatedBy        common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy    common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	Spec dbv1.BuildRunSpec `json:"spec"`

	Source GitSource `json:"source"`

	CredUser common.CreatedOrUpdatedBy `json:"credUser" graphql:"noinput"`

	ErrorMessages map[string]string `json:"errorMessages" graphql:"noinput"`
	Status        BuildStatus       `json:"status" graphql:"noinput"`
}

var BuildIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "spec.registry.repo.name", Value: repos.IndexAsc},
			{Key: "spec.accountName", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "spec.accountName", Value: repos.IndexAsc},
			{Key: "spec.cacheKeyName", Value: repos.IndexAsc},
		},
	},
}
