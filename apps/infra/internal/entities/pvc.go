package entities

import (
	corev1 "k8s.io/api/core/v1"
	"kloudlite.io/pkg/repos"
)

type PersistentVolumeClaim struct {
	repos.BaseEntity             `json:",inline" graphql:"noinput"`
	corev1.PersistentVolumeClaim `json:",inline" graphql:"noinput"`
	AccountName                  string `json:"accountName" graphql:"noinput"`
	ClusterName                  string `json:"clusterName" graphql:"noinput"`
}

var PersistentVolumeClaimIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
