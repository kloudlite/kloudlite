package reconciler

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:generate=true
type CheckDefinition struct {
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Debug       bool    `json:"debug,omitempty"`
	Hide        bool    `json:"hide,omitempty"`
}

// +kubebuilder:object:generate=true
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
type Status struct {
	// +kubebuilder:validation:Optional
	IsReady bool `json:"isReady"`
	// Resources []ResourceRef `json:"resources,omitempty"`

	CheckList           []CheckDefinition      `json:"checkList,omitempty"`
	Checks              map[string]CheckResult `json:"checks,omitempty"`
	LastReadyGeneration int64                  `json:"lastReadyGeneration,omitempty"`
	LastReconcileTime   *metav1.Time           `json:"lastReconcileTime,omitempty"`
}
