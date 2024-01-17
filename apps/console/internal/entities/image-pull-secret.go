package entities

import (
	"fmt"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PullSecretFormat string

const (
	DockerConfigJsonFormat PullSecretFormat = "dockerConfigJson"
	ParamsFormat           PullSecretFormat = "params"
)

type ImagePullSecret struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	metav1.ObjectMeta `json:"metadata"`

	Format           PullSecretFormat `json:"format"`
	DockerConfigJson *string          `json:"dockerConfigJson,omitempty"`

	RegistryUsername *string `json:"registryUsername,omitempty"`
	RegistryPassword *string `json:"registryPassword,omitempty"`
	RegistryURL      *string `json:"registryURL,omitempty"`

	GeneratedK8sSecret corev1.Secret `json:"generatedK8sSecret,omitempty" graphql:"ignore" struct-json-path:",ignore-nesting"`

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
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.MetadataNamespace, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
			{Key: fc.ProjectName, Value: repos.IndexAsc},
			{Key: fc.EnvironmentName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
