package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type SecretStatus string

const (
	SecretStateSyncing = SecretStatus("sync-in-progress")
	SecretStateLive    = SecretStatus("live")
	SecretStateError   = SecretStatus("error")
	SecretStateDown    = SecretStatus("down")
)

type Secret struct {
	repos.BaseEntity `bson:",inline"`
	ClusterId        repos.ID           `json:"cluster_id" bson:"cluster_id"`
	ProjectId        repos.ID           `json:"project_id" bson:"project_id"`
	Name             string             `json:"name" bson:"name"`
	Namespace        string             `json:"namespace" bson:"namespace"`
	Description      *string            `json:"description" bson:"description"`
	Data             []*Entry           `json:"data" bson:"data"`
	Status           SecretStatus       `json:"status" bson:"status"`
	Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

var SecretIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "name", Value: repos.IndexAsc},
			{Key: "namespace", Value: repos.IndexAsc},
			{Key: "cluster_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
