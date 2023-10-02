package v1

import (
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ChartRepo struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type JobVars struct {
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`
}

// HelmChartSpec defines the desired state of HelmChart
type HelmChartSpec struct {
	ChartRepo ChartRepo `json:"chartRepo"`

	ChartVersion string `json:"chartVersion"`
	ChartName    string `json:"chartName"`

	JobVars JobVars `json:"jobVars,omitempty"`

	// +kubebuilder:validation:Type=string
	ValuesYaml string `json:"valuesYaml"`
}

// HelmChartStatus defines the observed state of HelmChart
type HelmChartStatus struct {
	rApi.Status   `json:",inline"`
	ReleaseNotes  string `json:"releaseNotes"`
	ReleaseStatus string `json:"releaseStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.chartName",name=ChartName,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.chartRepo.url",name=ChartRepo,type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".status.releaseStatus",name=ReleaseStatus,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// HelmChart is the Schema for the helmcharts API
type HelmChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmChartSpec   `json:"spec,omitempty"`
	Status HelmChartStatus `json:"status,omitempty"`
}

func (p *HelmChart) EnsureGVK() {
	if p != nil {
		p.SetGroupVersionKind(GroupVersion.WithKind("HelmChart"))
	}
}

func (p *HelmChart) GetStatus() *rApi.Status {
	return &p.Status.Status
}

func (p *HelmChart) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (p *HelmChart) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// HelmChartList contains a list of HelmChart
type HelmChartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmChart `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelmChart{}, &HelmChartList{})
}
