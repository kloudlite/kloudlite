package templates

import (
	corev1 "k8s.io/api/core/v1"
)

type HelmChartInstallJobSpecParams struct {
	BackOffLimit int

	PodLabels      map[string]string
	PodAnnotations map[string]string

	ReleaseName      string
	ReleaseNamespace string

	Image           string
	ImagePullPolicy string

	ServiceAccountName string
	Tolerations        []corev1.Toleration
	Affinity           corev1.Affinity
	NodeSelector       map[string]string

	ChartRepoURL string
	ChartName    string
	ChartVersion string

	PreInstall  string
	PostInstall string

	HelmValuesYAML string
}

type HelmChartUninstallJobSpecParams struct {
	BackOffLimit int

	PodLabels      map[string]string
	PodAnnotations map[string]string

	ReleaseName      string
	ReleaseNamespace string

	Image           string
	ImagePullPolicy string

	ServiceAccountName string
	Tolerations        []corev1.Toleration
	Affinity           corev1.Affinity
	NodeSelector       map[string]string

	ChartRepoURL string
	ChartName    string
	ChartVersion string

	PreUninstall  string
	PostUninstall string
}

// type HelmPipelineInstallJobVars struct {
// 	JobMetadata    metav1.ObjectMeta
// 	PodTolerations []corev1.Toleration
// 	PodAnnotations map[string]string
//
// 	Pipeline []v1.PipelineStep
// }
