package op_crds

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CloudProviderMetadata struct {
	Name            string                  `json:"name,omitempty"`
	Annotations     map[string]string       `json:"annotations,omitempty"`
	Labels          map[string]string       `json:"labels,omitempty"`
	OwnerReferences []metav1.OwnerReference `json:"ownerReferences,omitempty"`
}

type CloudProviderSpec struct {
}

const CloudProviderAPIVersion = "infra.kloudlite.io/v1"
const CloudProviderKind = "CloudProvider"

type CloudProvider struct {
	APIVersion string                `json:"apiVersion,omitempty"`
	Kind       string                `json:"kind,omitempty"`
	Metadata   CloudProviderMetadata `json:"metadata,omitempty"`
	Spec       CloudProviderSpec     `json:"spec,omitempty"`
}
