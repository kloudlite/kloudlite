package entities

import (
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type FreeDeviceIP struct {
	repos.BaseEntity `json:",inline"`

	AccountName   string `json:"accountName"`
	GlobalVPNName string `json:"globalVPNName"`

	IPAddr string `json:"ipAddr"`
}

var FreeDeviceIPIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.FreeDeviceIPIpAddr, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.FreeDeviceIPGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type ClaimDeviceIP struct {
	repos.BaseEntity `json:",inline"`

	AccountName   string `json:"accountName"`
	GlobalVPNName string `json:"globalVPNName"`

	IPAddr    string `json:"ipAddr"`
	ClaimedBy string `json:"claimedBy"`
}

var ClaimDeviceIPIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClaimDeviceIPIpAddr, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.ClaimDeviceIPGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClaimDeviceIPClaimedBy, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.ClaimDeviceIPGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
