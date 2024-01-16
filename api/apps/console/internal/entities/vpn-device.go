package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wireguardV1 "github.com/kloudlite/operator/apis/wireguard/v1"
)

type ConsoleVPNDevice struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	wireguardV1.Device `json:",inline"`

	common.ResourceMetadata `json:",inline"`

	AccountName     string  `json:"accountName" graphql:"noinput"`
	ProjectName     *string `json:"projectName,omitempty"`
	EnvironmentName *string `json:"environmentName,omitempty"`

	WireguardConfig t.EncodedString `json:"wireguardConfig,omitempty" graphql:"noinput"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var VPNDeviceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

func ValidateVPNDevice(d *ConsoleVPNDevice) error {
	errMsgs := []string{}

	if d.DisplayName == "" {
		errMsgs = append(errMsgs, "displayName is required")
	}

	if len(errMsgs) > 0 {
		return errors.Newf("%v", errMsgs)
	}
	return nil
}
