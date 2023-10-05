package entities

import "time"

type GitlabGroup struct {
	Id        string `json:"id" graphql:"noinput"`
	FullName  string `json:"fullName" graphql:"noinput"`
	AvatarUrl string `json:"avatarUrl" graphql:"noinput"`
}

type GitlabProject struct {
	ID                int        `json:"id" graphql:"noinput"`
	Description       string     `json:"description" graphql:"noinput"`
	DefaultBranch     string     `json:"defaultBranch" graphql:"noinput"`
	Public            bool       `json:"public" graphql:"noinput"`
	SSHURLToRepo      string     `json:"sshUrlToRepo" graphql:"noinput"`
	HTTPURLToRepo     string     `json:"httpUrlToRepo" graphql:"noinput"`
	WebURL            string     `json:"webUrl" graphql:"noinput"`
	TagList           []string   `json:"tagList" graphql:"noinput"`
	Topics            []string   `json:"topics" graphql:"noinput"`
	Name              string     `json:"name" graphql:"noinput"`
	NameWithNamespace string     `json:"nameWithNamespace" graphql:"noinput"`
	Path              string     `json:"path" graphql:"noinput"`
	PathWithNamespace string     `json:"pathWithNamespace" graphql:"noinput"`
	CreatedAt         *time.Time `json:"createdAt,omitempty" graphql:"noinput"`
	LastActivityAt    *time.Time `json:"lastActivityAt,omitempty" graphql:"noinput"`
	CreatorID         int        `json:"creatorId" graphql:"noinput"`
	EmptyRepo         bool       `json:"emptyRepo" graphql:"noinput"`
	Archived          bool       `json:"archived" graphql:"noinput"`
	AvatarURL         string     `json:"avatarUrl" graphql:"noinput"`
}
