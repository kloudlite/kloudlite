package op_crds

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ProjectSpec struct {
	DisplayName string `json:"displayName,omitempty"`
}

type Status struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type Project struct {
	Name      string      `json:"name,omitempty"`
	NameSpace string      `json:"name,omitempty"`
	ClusterId string      `json:"cluster_id"`
	Spec      ProjectSpec `json:"spec,omitempty"`
	Status    Status      `json:"status,omitempty"`
}
