package op_crds

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type StatusMetadata struct {
	ClusterId        string                  `json:"clusterId,omitempty"`
	ProjectId        string                  `json:"projectId,omitempty"`
	ResourceId       string                  `json:"resourceId,omitempty"`
	GroupVersionKind metav1.GroupVersionKind `json:"groupVersionKind,omitempty"`
}

type StatusUpdate struct {
	Metadata        StatusMetadata     `json:"metadata,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
	ChildConditions []metav1.Condition `json:"childConditions,omitempty"`
	IsReady         bool               `json:"isReady,omitempty"`
	TobeDeleted     bool               `json:"tobeDeleted,omitempty"`
	Stage           string             `json:"stage,omitempty"`
}
