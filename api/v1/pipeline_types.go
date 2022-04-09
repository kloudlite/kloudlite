package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	GitProvider string         `json:"gitProvider"`
	GitRepoUrl  string         `json:"gitRepoUrl"`
	GitRef      string         `json:"gitRef"`
	BuildArgs   []buildArg     `json:"buildArgs,omitempty"`
	Dockerfile  string         `json:"dockerfile,omitempty"`
	ContextDir  string         `json:"contextDir,omitempty"`
	Github      pipelineGithub `json:"github,omitempty"`
	Gitlab      pipelineGitlab `json:"gitlab,omitempty"`
}

type buildArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type pipelineGithub struct {
	InstallationId string `json:"installationId"`
	TokenId        string `json:"tokenId"`
}

type pipelineGitlab struct {
	TokenId string `json:"tokenId"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Pipeline is the Schema for the pipelines API
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
