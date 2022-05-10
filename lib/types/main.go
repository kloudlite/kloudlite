package types

// +kubebuilder:pruning:PreserveUnknownFields
// +kubebuilder:validation:EmbeddedResource
// +kubebuilder:validation:Schemaless
// type KV json.RawMessage
//
// func (k *KV) DeepCopyInto(out *KV) {
// 	if out != nil {
// 		u := unstructured.Unstructured(*k)
// 		dc := u.DeepCopy()
// 		out.Object = dc.Object
// 	}
// }
