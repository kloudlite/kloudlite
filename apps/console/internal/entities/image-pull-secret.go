package entities

import (
	"fmt"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImagePullSecretFormat string

const (
	DockerConfigJsonFormat ImagePullSecretFormat = "dockerConfigJson"
	ParamsFormat           ImagePullSecretFormat = "params"
)

type ImagePullSecret struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	metav1.ObjectMeta `json:"metadata"`

	Format           ImagePullSecretFormat `json:"format"`
	DockerConfigJson *string               `json:"dockerConfigJson,omitempty"`

	RegistryUsername *string `json:"registryUsername,omitempty"`
	RegistryPassword *string `json:"registryPassword,omitempty"`
	RegistryURL      *string `json:"registryURL,omitempty"`

	GeneratedK8sSecret corev1.Secret `json:"generatedK8sSecret,omitempty" graphql:"ignore"`
	//corev1.Secret    `json:",inline"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	ProjectName     string `json:"projectName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (ips *ImagePullSecret) GetResourceType() ResourceType {
	return ResourceTypeImagePullSecret
}

func (ips *ImagePullSecret) Validate() error {
	if ips == nil {
		return fmt.Errorf("image pull secret is nil")
	}
	switch ips.Format {
	case DockerConfigJsonFormat:
		if ips.DockerConfigJson == nil {
			return fmt.Errorf("when format is %s, field: dockerConfigJson must be set", DockerConfigJsonFormat)
		}
	case ParamsFormat:
		if ips.RegistryUsername == nil || ips.RegistryPassword == nil || ips.RegistryURL == nil {
			return fmt.Errorf("when format is %s, fields: registryUsername, registryPassword, registryURL must be set", ParamsFormat)
		}
	}

	return nil
}

var ImagePullSecretIndexes = []repos.IndexField{
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
			{Key: "projectName", Value: repos.IndexAsc},
			{Key: "environmentName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
