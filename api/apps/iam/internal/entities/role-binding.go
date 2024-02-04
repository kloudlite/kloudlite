package entities

import (
	t "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type RoleBinding struct {
	repos.BaseEntity `json:",inline" bson:",inline"`
	UserId           string         `json:"user_id"`
	ResourceType     t.ResourceType `json:"resource_type"`
	ResourceRef      string         `json:"resource_ref"`
	Role             t.Role         `json:"role"`
}

func (rb *RoleBinding) Validate() error {
	verr := common.ValidationError{Label: "role_binding"}

	if rb.UserId == "" {
		verr.Errors = append(verr.Errors, "user_id is required")
	}
	if rb.ResourceType == "" {
		verr.Errors = append(verr.Errors, "resource_type is required")
	}
	if rb.ResourceRef == "" {
		verr.Errors = append(verr.Errors, "resource_ref is required")
	}
	if rb.Role == "" {
		verr.Errors = append(verr.Errors, "role is required")
	}

	if len(verr.Errors) > 0 {
		return verr
	}

	return nil
}

var RoleBindingIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "user_id", Value: repos.IndexDesc},
			{Key: "resource_ref", Value: repos.IndexDesc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "role", Value: repos.IndexAsc},
			{Key: "resource_type", Value: repos.IndexAsc},
		},
	},
}
