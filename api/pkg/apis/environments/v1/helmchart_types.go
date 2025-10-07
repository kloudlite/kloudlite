package v1

import (
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={kloudlite,environments},shortName=comp
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Services",type=integer,JSONPath=`.status.servicesCount`
// +kubebuilder:printcolumn:name="Running",type=integer,JSONPath=`.status.runningCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Composition represents a Docker Compose application deployed in an environment
type HelmChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`

	Spec   HelmChartSpec   `json:"spec,omitzero"`
	Status HelmChartStatus `json:"status,omitzero"`
}

type HelmChartInfo struct {
	URL     string `json:"url"`
	Version string `json:"version,omitempty"`
	Name    string `json:"name"`
}

type HelmJobVars struct {
	NodeSelector map[string]string           `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration         `json:"tolerations,omitempty"`
	Affinity     *corev1.Affinity            `json:"affinity,omitempty"`
	Resources    corev1.ResourceRequirements `json:"resources,omitzero"`
}

type HelmChartSpec struct {
	Chart HelmChartInfo `json:"chart"`

	HelmValues map[string]apiextensionsv1.JSON `json:"helmValues"`

	HelmJobVars *HelmJobVars `json:"jobVars,omitzero"`

	PreInstall  string `json:"preInstall,omitempty"`
	PostInstall string `json:"postInstall,omitempty"`

	PreUninstall  string `json:"preUninstall,omitempty"`
	PostUninstall string `json:"postUninstall,omitempty"`
}

type HelmChartStatus struct {
	// State represents the current state of the helm chart
	// +kubebuilder:validation:Enum=pending;installing;installed;upgrading;failed;uninstalling;deleting
	State HelmChartState `json:"state,omitempty"`

	// Message provides human-readable information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// ReleaseName is the Helm release name used for installation
	// +optional
	ReleaseName string `json:"releaseName,omitempty"`

	// InstalledVersion is the currently installed chart version
	// +optional
	InstalledVersion string `json:"installedVersion,omitempty"`

	// LastInstallTime is when the chart was last installed/upgraded
	// +optional
	LastInstallTime *metav1.Time `json:"lastInstallTime,omitempty"`

	// Conditions represent the latest available observations of the helm chart's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// DeployedResources tracks the Kubernetes resources created by this helm chart
	// +optional
	DeployedResources []string `json:"deployedResources,omitempty"`
}

// HelmChartState represents the state of a helm chart
type HelmChartState string

const (
	// HelmChartStatePending means the helm chart is pending installation
	HelmChartStatePending HelmChartState = "pending"

	// HelmChartStateInstalling means the helm chart is being installed
	HelmChartStateInstalling HelmChartState = "installing"

	// HelmChartStateInstalled means the helm chart is successfully installed
	HelmChartStateInstalled HelmChartState = "installed"

	// HelmChartStateUpgrading means the helm chart is being upgraded
	HelmChartStateUpgrading HelmChartState = "upgrading"

	// HelmChartStateFailed means the helm chart installation/upgrade failed
	HelmChartStateFailed HelmChartState = "failed"

	// HelmChartStateUninstalling means the helm chart is being uninstalled
	HelmChartStateUninstalling HelmChartState = "uninstalling"

	// HelmChartStateDeleting means the helm chart is being deleted
	HelmChartStateDeleting HelmChartState = "deleting"
)

// +kubebuilder:object:root=true

// HelmChartList contains a list of HelmChart
type HelmChartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmChart `json:"items"`
}
