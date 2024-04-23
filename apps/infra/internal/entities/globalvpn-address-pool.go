package entities

import (
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type GlobalVPNDeviceAddressPool struct {
	repos.BaseEntity `json:",inline"`

	AccountName   string `json:"accountName"`
	GlobalVPNName string `json:"globalVPNName"`

	CIDR      string `json:"cidr"`
	MinOffset int    `json:"minOffset"`
	MaxOffset int    `json:"maxOffset"`

	RunningOffset int `json:"runningOffset"`
	// ReservedIPs     map[string]string   `json:"reservedIPs"`
	// FreeAddressPool map[string]struct{} `json:"freeAddressPool"`
}

var DeviceAddressPoolIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "globalVPNName", Value: repos.IndexAsc},
		},
	},
}

type IPClaim struct {
	repos.BaseEntity `json:",inline"`

	AccountName   string `json:"accountName"`
	GlobalVPNName string `json:"globalVPNName"`

	IPAddr         string `json:"ipAddr"`
	ReservationKey string `json:"reservationKey"`
}

var IPClaimIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.IPClaimIpAddr, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.FreeIPGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.IPClaimReservationKey, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.FreeIPGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type FreeIP struct {
	repos.BaseEntity `json:",inline"`

	AccountName   string `json:"accountName"`
	GlobalVPNName string `json:"globalVPNName"`

	IPAddr string `json:"ipAddr"`
}

var FreeIPIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.FreeIPIpAddr, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.FreeIPGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
