package entities

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TeamMembership struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	
	common.ResourceMetadata `json:",inline"`

	UserId repos.ID      `json:"userId"`
	TeamId repos.ID      `json:"teamId"`
	Role   iamT.Role    `json:"role"`
}

var TeamMembershipIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "userId", Value: repos.IndexAsc},
			{Key: "teamId", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "teamId", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "userId", Value: repos.IndexAsc},
		},
	},
}