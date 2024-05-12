package entities

import (
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GlobalVPNDevice struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	metav1.ObjectMeta `json:"metadata"`

	common.ResourceMetadata `json:",inline"`

	AccountName   string `json:"accountName" graphql:"noinput"`
	GlobalVPNName string `json:"globalVPNName"`

	// Only needs to be set, if vpn device has a public IP
	PublicEndpoint *string `json:"publicEndpoint,omitempty" graphql:"noinput"`

	CreationMethod string `json:"creationMethod,omitempty"`

	IPAddr string `json:"ipAddr" graphql:"noinput"`

	PrivateKey string `json:"privateKey" graphql:"noinput"`
	PublicKey  string `json:"publicKey" graphql:"noinput"`
}

var GlobalVPNDeviceIndices = []repos.IndexField{
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
			{Key: fc.GlobalVPNDeviceGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.GlobalVPNDeviceGlobalVPNName, Value: repos.IndexAsc},
			{Key: fc.GlobalVPNDeviceIpAddr, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
