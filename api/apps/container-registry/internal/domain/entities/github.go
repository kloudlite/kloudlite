package entities

import (
	"time"
)

type GithubUserAccount struct {
	ID        *int64  `json:"id,omitempty"`
	NodeID    *string `json:"node_id,omitempty"`
	Type      *string `json:"type,omitempty"`
	Login     *string `json:"login,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type GInst struct {
	ID              *int64            `json:"id,omitempty"`
	AppID           *int64            `json:"app_id,omitempty"`
	TargetType      *string           `json:"target_type,omitempty"`
	TargetID        *int64            `json:"target_id,omitempty"`
	NodeID          *string           `json:"node_id,omitempty"`
	RepositoriesURL *string           `json:"repositories_url,omitempty"`
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
	FullName      *string         `json:"full_name,omitempty"`
	Description   *string         `json:"description,omitempty"`
	DefaultBranch *string         `json:"default_branch,omitempty"`
	MasterBranch  *string         `json:"master_branch,omitempty"`
	CreatedAt     time.Time       `json:"created_at,omitempty"`
	PushedAt      time.Time       `json:"pushed_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at,omitempty"`
	HTMLURL       *string         `json:"html_url,omitempty"`
	CloneURL      *string         `json:"clone_url,omitempty"`
	GitURL        *string         `json:"git_url,omitempty"`
	MirrorURL     *string         `json:"mirror_url,omitempty"`
	Language      *string         `json:"language,omitempty"`
	Size          *int            `json:"size,omitempty"`
	Permissions   map[string]bool `json:"permissions,omitempty"`
	Archived      *bool           `json:"archived,omitempty"`
	Disabled      *bool           `json:"disabled,omitempty"`

	// Additional mutable fields when creating and editing a repository
	Private           *bool   `json:"private,omitempty"`
	GitignoreTemplate *string `json:"gitignore_template,omitempty"`

	// Creating an organization repository. Required for non-owners.
	TeamID *int64 `json:"team_id,omitempty"`

	Visibility *string `json:"visibility,omitempty"`
}

type GithubListRepository struct {
	TotalCount   *int                `json:"total_count,omitempty" graphql:"noinput"`
	Repositories []*GithubRepository `json:"repositories" graphql:"noinput"`
}

type GithubSearchRepository struct {
	Total             *int                `json:"total,omitempty" graphql:"noinput"`
	Repositories      []*GithubRepository `json:"repositories" graphql:"noinput"`
	IncompleteResults *bool               `json:"incomplete_results,omitempty" graphql:"noinput"`
}

type GithubBranch struct {
	Name      *string `json:"name,omitempty" graphql:"noinput"`
	Protected *bool   `json:"protected,omitempty" graphql:"noinput"`
}
