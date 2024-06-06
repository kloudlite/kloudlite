package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type Email struct {
	Enabled     bool   `json:"enabled"`
	MailAddress string `json:"mailAddress"`
}

type Slack struct {
	Enabled bool   `json:"enabled"`
	Url     string `json:"url"`
}

type Telegram struct {
	Enabled bool   `json:"enabled"`
	Token   string `json:"token"`
	ChatID  string `json:"chatId"`
}

type Webhook struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

type NotificationConf struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	CreatedBy        common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy    common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	Email    *Email    `json:"email"`
	Slack    *Slack    `json:"slack"`
	Telegram *Telegram `json:"telegram"`
	Webhook  *Webhook  `json:"webhook"`

	AccountName string `json:"accountName" graphql:"noinput"`
}

var NotificationConfIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
