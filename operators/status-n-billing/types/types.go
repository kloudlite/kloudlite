package types

import (
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type StatusUpdate struct {
	AccountName string         `json:"accountName"`
	ClusterName string         `json:"clusterName,omitempty"`
	Object      map[string]any `json:"object"`
	Status      rApi.Status    `json:"status"`
}
