package types

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:validation:EmbeddedResource
type KV struct {
	unstructured.Unstructured `json:",inline"`
}

func (k *KV) Set(key string, value interface{}) {
	k.Object[key] = value
}

func (k *KV) SetFromMap(m *map[string]interface{}) {
	k.Object = *m
}

func (k *KV) DeepCopyInto(out *KV) {
	if out != nil {
		dc := k.Unstructured.DeepCopy()
		out.Object = dc.Object
	}
}
