package op_crds

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type EdgeMetadata struct {
	Name            string                  `json:"name,omitempty"`
	Annotations     map[string]string       `json:"annotations,omitempty"`
	Labels          map[string]string       `json:"labels,omitempty"`
	OwnerReferences []metav1.OwnerReference `json:"ownerReferences,omitempty"`
}

type NodePool struct {
	Name   string `json:"name"`
	Config string `json:"config"`
	Min    int    `json:"min"`
	Max    int    `json:"max"`
}

type CredentialsRef struct {
	SecretName string `json:"secretName"`
	Namespace  string `json:"namespace"`
}

type EdgeSpec struct {
	AccountId      *string        `json:"accountId,omitempty"`
	Provider       string         `json:"provider"`
	Region         string         `json:"region"`
	Pools          []NodePool     `json:"pools"`
	CredentialsRef CredentialsRef `json:"credentialsRef"`
}

const EdgeAPIVersion = "infra.kloudlite.io/v1"
const EdgeKind = "Edge"

type Region struct {
	APIVersion string       `json:"apiVersion,omitempty"`
	Kind       string       `json:"kind,omitempty"`
	Metadata   EdgeMetadata `json:"metadata,omitempty"`
	Spec       EdgeSpec     `json:"spec,omitempty"`
}
