package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Config is the Schema for the configs API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	ProjectName string            `json:"projectName,omitempty"`
	Data        map[string]string `json:"data,omitempty"`

	// +kubebuilder:default=true
	Enabled   bool        `json:"enabled,omitempty"`
	Overrides *JsonPatch  `json:"overrides,omitempty"`
	Status    rApi.Status `json:"status,omitempty"`
}

func (cfg *Config) EnsureGVK() {
	if cfg != nil {
		cfg.SetGroupVersionKind(GroupVersion.WithKind("Config"))
	}
}

func (cfg *Config) GetStatus() *rApi.Status {
	return &cfg.Status
}

func (cfg *Config) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (cfg *Config) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: cfg.GroupVersionKind().String(),
	}
}

//+kubebuilder:object:root=true

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
