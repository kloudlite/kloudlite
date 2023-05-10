package types

type ResourceUpdate struct {
	AuthToken   string         `json:"authToken"`
	AccountName string         `json:"accountName"`
	ClusterName string         `json:"clusterName"`
	Object      map[string]any `json:"object"`
}
