package apiclient

//
//import (
//	fn "github.com/kloudlite/kl/pkg/functions"
//)
//
//type BYOKCluster struct {
//	DisplayName string   `json:"displayName"`
//	Metadata    Metadata `json:"metadata"`
//}
//
//func (apic *apiClient) ListBYOKClusters(teamName string) ([]BYOKCluster, error) {
//	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
//	if err != nil {
//		return nil, fn.NewE(err)
//	}
//
//	respData, err := klFetch("cli_listAccountClusters", map[string]any{
//		"pq": map[string]any{
//			"orderBy":       "updateTime",
//			"sortDirection": "ASC",
//			"first":         99999999,
//		},
//	}, &cookie)
//
//	if err != nil {
//		return nil, fn.NewE(err)
//	}
//
//	if fromResp, err := GetFromRespForEdge[BYOKCluster](respData); err != nil {
//		return nil, fn.NewE(err)
//	} else {
//		return fromResp, fn.NewE(err)
//	}
//}
