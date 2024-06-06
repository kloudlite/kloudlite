package types

import (
	"fmt"

	"github.com/kloudlite/api/pkg/egob"
	"github.com/kloudlite/api/pkg/repos"
)

type NotificationType string

const (
	NotifyTypeAlert  NotificationType = "alert"
	NotifyTypeNotify NotificationType = "notification"
)

type NotifyContent struct {
	Title   string `json:"title" graphql:"noinput"`
	Subject string `json:"subject" graphql:"noinput"`
	Body    string `json:"body" graphql:"noinput"`
	Link    string `json:"link" graphql:"noinput"`
	Image   string `json:"image" graphql:"noinput"`
}

type Notification struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	Type             NotificationType `json:"notificationType" graphql:"noinput"`

	Content     NotifyContent `json:"content" graphql:"noinput"`
	Priority    int           `json:"priority" graphql:"noinput"`
	AccountName string        `json:"accountName" graphql:"noinput"`
	Read        bool          `json:"read" graphql:"noinput"`
}

func (obj *Notification) ToPlain() string {
	return fmt.Sprintf(`%s

%s

%s

Account: %s
`, obj.Type, obj.Content.Title, obj.Content.Body, obj.AccountName)
}

func (obj *Notification) ToBytes() ([]byte, error) {
	return egob.Marshal(obj)
}

func (obj *Notification) ParseBytes(data []byte) error {
	return egob.Unmarshal(data, obj)
}

var NotificationIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
