package v1

import (
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobVars struct {
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Affinity     *corev1.Affinity    `json:"affinity,omitempty"`
	// +kubebuilder:default=1
	BackOffLimit *int32 `json:"backOffLimit,omitempty"`
}

// HelmChartSpec defines the desired state of HelmChart
type HelmChartSpec struct {
	ChartRepoURL string `json:"chartRepoURL"`

	// find chartVersion by running command `helm search repo <chartName> --versions` 2nd column is the chartVersion
	ChartVersion string `json:"chartVersion"`

	ChartName string `json:"chartName"`

	JobVars JobVars `json:"jobVars,omitempty"`

	PreInstall  string `json:"preInstall,omitempty"`
	PostInstall string `json:"postInstall,omitempty"`

	PreUninstall  string `json:"preUninstall,omitempty"`
	PostUninstall string `json:"postUninstall,omitempty"`

	Values map[string]apiextensionsv1.JSON `json:"values"`
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
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".status.releaseStatus",name=ReleaseStatus,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// HelmChart is the Schema for the helmcharts API
type HelmChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmChartSpec   `json:"spec,omitempty"`
	Status HelmChartStatus `json:"status,omitempty" graphql:"noinput"`
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
