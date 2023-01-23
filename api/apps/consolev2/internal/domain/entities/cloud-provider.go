package entities

import (
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
	"io"

	// "gopkg.in/validator.v2"
	op_crds "kloudlite.io/apps/consolev2/internal/domain/op-crds"
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
	op_crds.CloudProvider `bson:",inline" json:",inline"`

	// sync status
	SyncStatus CloudProviderStatus `json:"sync_status" bson:"sync_status,omitempty" validate:"nonzero"`
}

func (c *CloudProvider) UnmarshalGQL(v interface{}) error {
	if err := json.Unmarshal([]byte(v.(string)), c); err != nil {
		return err
	}

	// if err := validator.Validate(*c); err != nil {
	// 	return err
	// }

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
