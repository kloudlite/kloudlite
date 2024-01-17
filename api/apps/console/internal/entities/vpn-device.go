package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
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

	WireguardConfig t.EncodedString `json:"wireguardConfig,omitempty" graphql:"noinput" struct-json-path:",ignore-nesting"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var VPNDeviceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
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
