package entities

import (
	"time"
)

type GithubUserAccount struct {
	ID        *int64  `json:"id,omitempty"`
	NodeID    *string `json:"nodeId,omitempty"`
	Type      *string `json:"type,omitempty"`
	Login     *string `json:"login,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

type GInst struct {
	ID              *int64            `json:"id,omitempty"`
	AppID           *int64            `json:"appId,omitempty"`
	TargetType      *string           `json:"targetType,omitempty"`
	TargetID        *int64            `json:"targetId,omitempty"`
	NodeID          *string           `json:"nodeId,omitempty"`
	RepositoriesURL *string           `json:"repositoriesUrl,omitempty"`
	Account         GithubUserAccount `json:"account,omitempty"`
}

type GithubInstallation struct {
	GInst `json:",inline" graphql:"noinput"`
}

type GithubRepository struct {
	ID            *int64          `json:"id,omitempty"`
	NodeID        *string         `json:"node_id,omitempty"`
	URL           *string         `json:"url,omitempty"`
	Name          *string         `json:"name,omitempty"`
	FullName      *string         `json:"fullName,omitempty"`
	Description   *string         `json:"description,omitempty"`
	DefaultBranch *string         `json:"defaultBranch,omitempty"`
	MasterBranch  *string         `json:"masterBranch,omitempty"`
	CreatedAt     time.Time       `json:"createdAt,omitempty"`
	PushedAt      time.Time       `json:"pushedAt,omitempty"`
	UpdatedAt     time.Time       `json:"updatedAt,omitempty"`
	HTMLURL       *string         `json:"htmlUrl,omitempty"`
	CloneURL      *string         `json:"cloneUrl,omitempty"`
	GitURL        *string         `json:"gitUrl,omitempty"`
	MirrorURL     *string         `json:"mirrorUrl,omitempty"`
	Language      *string         `json:"language,omitempty"`
	Size          *int            `json:"size,omitempty"`
	Permissions   map[string]bool `json:"permissions,omitempty"`
	Archived      *bool           `json:"archived,omitempty"`
	Disabled      *bool           `json:"disabled,omitempty"`

	// Additional mutable fields when creating and editing a repository
	Private           *bool   `json:"private,omitempty"`
	GitignoreTemplate *string `json:"gitignoreTemplate,omitempty"`

	// Creating an organization repository. Required for non-owners.
	TeamID *int64 `json:"team_id,omitempty"`

	Visibility *string `json:"visibility,omitempty"`
}

type GithubListRepository struct {
	TotalCount   *int                `json:"totalCount,omitempty" graphql:"noinput"`
	Repositories []*GithubRepository `json:"repositories" graphql:"noinput"`
}

type GithubSearchRepository struct {
	Total             *int                `json:"total,omitempty" graphql:"noinput"`
	Repositories      []*GithubRepository `json:"repositories" graphql:"noinput"`
	IncompleteResults *bool               `json:"incompleteResults,omitempty" graphql:"noinput"`
}

type GitBranch struct {
	Name      string `json:"name,omitempty" graphql:"noinput"`
	Protected bool   `json:"protected,omitempty" graphql:"noinput"`
}
