package entities

// import (
// 	"github.com/kloudlite/api/common"
// 	"github.com/kloudlite/api/pkg/repos"
// 	t "github.com/kloudlite/api/pkg/types"
// 	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
// 	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
// )
//
// type GatewayDeviceRef struct {
// 	Name   string `json:"name"`
// 	IPAddr string `json:"ipAddr"`
// }
//
// type Gateway struct {
// 	repos.BaseEntity `json:",inline" graphql:"noinput"`
//
// 	networkingv1.Gateway `json:",inline"`
//
// 	GlobalVPNName string `json:"globalVPNName"`
//
// 	common.ResourceMetadata `json:",inline"`
//
// 	AccountName    string           `json:"accountName" graphql:"noinput"`
// 	ClusterName    string           `json:"clusterName" graphql:"noinput"`
// 	ClusterSvcCIDR string           `json:"clusterSvcCIDR" graphql:"noinput"`
// 	DeviceRef      GatewayDeviceRef `json:"deviceRef" graphql:"noinput"`
//
// 	ParsedWgParams *wgv1.WgParams `json:"parsedWgParams" graphql:"ignore"`
// 	SyncStatus     t.SyncStatus   `json:"syncStatus" graphql:"noinput"`
// }
//
// var GatewayIndices = []repos.IndexField{
// 	{
// 		Field: []repos.IndexKey{
// 			{Key: "id", Value: repos.IndexAsc},
// 		},
// 		Unique: true,
// 	},
// 	{
// 		Field: []repos.IndexKey{
// 			{Key: "metadata.name", Value: repos.IndexAsc},
// 			{Key: "accountName", Value: repos.IndexAsc},
// 			{Key: "clusterName", Value: repos.IndexAsc},
// 		},
// 		Unique: true,
// 	},
// }
