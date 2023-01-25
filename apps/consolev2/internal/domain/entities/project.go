package entities

import (
	"encoding/json"
	"io"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type ProjectStatus string

const (
	ProjectStateSyncing = ProjectStatus("sync-in-progress")
	ProjectStateLive    = ProjectStatus("live")
	ProjectStateError   = ProjectStatus("error")
	ProjectStateDown    = ProjectStatus("down")
)

type Project struct {
	repos.BaseEntity `bson:",inline" json:",inline"`
	crdsv1.Project   `json:",inline"`

	// IsDeleting  bool               `json:"is_deleting" bson:"is_deleting"`
	// AccountId   repos.ID           `json:"account_id" bson:"account_id"`
	// Name        string             `json:"name" bson:"name"`
	// DisplayName string             `json:"display_name" bson:"display_name"`
	// Description *string            `json:"description" bson:"description"`
	// Logo        *string            `json:"logo" bson:"logo"`
	// ReadableId  repos.ID           `json:"readable_id" bson:"readable_id"`
	// Status      ProjectStatus      `json:"status" bson:"status"`
	// RegionId    *repos.ID          `json:"region_id" bson:"region_id"`
	// Conditions  []metav1.Condition `json:"conditions" bson:"conditions"`
}

func (c *Project) UnmarshalGQL(v interface{}) error {
	switch res := v.(type) {
	case map[string]any:
		b, err := json.Marshal(res)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, c); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(v.(string)), c); err != nil {
			return err
		}
	}

	//if err := json.Unmarshal([]byte(v.(string)), c); err != nil {
	//	return err
	//}
	//
	//if err := validator.Validate(*c); err != nil {
	//	return err
	//}
	//
	return nil
}

func (c Project) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(c)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var ProjectIndexes = []repos.IndexField{
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

type ProjectMembership struct {
	ProjectId repos.ID
	UserId    repos.ID
	Role      common.Role
	Accepted  bool
}
