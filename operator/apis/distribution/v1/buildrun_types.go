package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Repo struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type Registry struct {
	Username string `json:"username" graphql:"ignore"`
	Password string `json:"password" graphql:"ignore"`
	Host     string `json:"host" graphql:"ignore"`

	Repo Repo `json:"repo"`
}

type GitRepo struct {
	Url    string `json:"url"`
	Branch string `json:"branch"`
}

type BuildOptions struct {
	BuildArgs         map[string]string `json:"buildArgs,omitempty"`
	BuildContexts     map[string]string `json:"buildContexts,omitempty"`
	DockerfilePath    *string           `json:"dockerfilePath,omitempty"`
	DockerfileContent *string           `json:"dockerfileContent,omitempty"`
	TargetPlatforms   []string          `json:"targetPlatforms,omitempty"`
	ContextDir        *string           `json:"contextDir,omitempty"`
}

type Resource struct {
	Cpu        int `json:"cpu"`
	MemoryInMb int `json:"memoryInMb"`
}

// BuildRunSpec defines the desired state of BuildRun
type BuildRunSpec struct {
	CacheKeyName *string `json:"cacheKeyName,omitempty"`
	AccountName  string  `json:"accountName" graphql:"noinput"`

	Registry     Registry      `json:"registry"`
	GitRepo      GitRepo       `json:"gitRepo" graphql:"ignore"`
	BuildOptions *BuildOptions `json:"buildOptions,omitempty"`
	Resource     Resource      `json:"resource"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced

// BuildRun is the Schema for the buildruns API
type BuildRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BuildRunSpec `json:"spec,omitempty"`
	Status rApi.Status  `json:"status,omitempty"`
}

func (d *BuildRun) EnsureGVK() {
	if d != nil {
		d.SetGroupVersionKind(GroupVersion.WithKind("Build"))
	}
}

func (d *BuildRun) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *BuildRun) GetEnsuredLabels() map[string]string {

	if d.Spec.CacheKeyName != nil {
		return map[string]string{
			constants.CacheNameKey:   *d.Spec.CacheKeyName,
			constants.AccountNameKey: d.Spec.AccountName,
			constants.BuildNameKey:   d.Name,
		}
	}

	return map[string]string{
		constants.AccountNameKey: d.Spec.AccountName,
		constants.BuildNameKey:   d.Name,
	}
}

func (d *BuildRun) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Build").String(),
	}
}

//+kubebuilder:object:root=true

// BuildRunList contains a list of BuildRun
type BuildRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BuildRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BuildRun{}, &BuildRunList{})
}
