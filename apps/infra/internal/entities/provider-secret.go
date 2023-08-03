package entities

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"kloudlite.io/pkg/repos"
)

type CloudProviderSecret struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	crdsv1.Secret    `json:",inline" graphql:"uri=k8s://secrets.crds.kloudlite.io"`
	//corev1.Secret `json:",inline" graphql:"uri=https://raw.githubusercontent.com/instrumenta/kubernetes-json-schema/master/v1.18.1/secret.json"`
	AccountName string `json:"accountName" graphql:"noinput"`
}

var SecretIndices = []repos.IndexField{
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
		},
		Unique: true,
	},

	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
