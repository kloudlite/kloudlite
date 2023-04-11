package types

type StatusUpdate struct {
	ClusterName string         `json:"clusterName,omitempty"`
	AccountName string         `json:"accountName,omitempty"`
	Object      map[string]any `json:"object,omitempty"`
}
