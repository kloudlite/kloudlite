package entities

import (
	"encoding/json"
	"github.com/vektah/gqlparser/v2/validator"
	"io"

	infrav1 "github.com/kloudlite/internal_operator_v2/apis/infra/v1"
	"gopkg.in/validator.v2"
	"kloudlite.io/pkg/repos"
)

type CloudProviderStatus string

const (
	CloudProviderStateSyncing = CloudProviderStatus("sync-in-progress")
	CloudProviderStateLive    = CloudProviderStatus("live")
	CloudProviderStateError   = CloudProviderStatus("error")
	CloudProviderStateDown    = CloudProviderStatus("down")
)

type CloudProvider struct {
	repos.BaseEntity      `bson:",inline"`
	infrav1.CloudProvider `bson:",inline" json:",inline"`

	// sync status
	SyncStatus CloudProviderStatus `json:"sync_status" bson:"sync_status,omitempty" validate:"nonzero"`
}

func (c *CloudProvider) UnmarshalGQL(v interface{}) error {
	if err := json.Unmarshal([]byte(v.(string)), c); err != nil {
		return err
	}

	if err := validator.Validate(*c); err != nil {
		return err
	}

	return nil
}

func (c CloudProvider) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(c)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var CloudProviderIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "spec.accountId", Value: repos.IndexAsc},
		},
	},
}
