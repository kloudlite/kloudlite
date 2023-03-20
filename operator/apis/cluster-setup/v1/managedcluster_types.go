package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KloudliteDnsApi struct {
	PublicHttpUri  string          `json:"publicHttpUri"`
	BasicAuthCreds SecretReference `json:"basicAuthCreds"`
}

type GitlabRunner struct {
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:default=gitlab-runner
	ReleaseName string `json:"releaseName,omitempty"`
	RunnerToken string `json:"runnerToken"`
	// +kubebuilder:default=helm-gitlab-runner
	Namespace string `json:"namespace,omitempty"`
}

type CertManager struct {
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:default=cert-manager
	ReleaseName string `json:"releaseName,omitempty"`
	// +kubebuilder:default=helm-cert-manager
	Namespace string `json:"namespace,omitempty"`
}

// ManagedClusterSpec defines the desired state of ManagedCluster
type ManagedClusterSpec struct {
	Domain         *string         `json:"domain,omitempty"`
	KloudliteCreds SecretReference `json:"kloudliteCreds,omitempty"`
	KlOperators    KlOperators     `json:"kloudliteOperators,omitempty"`
	CertManager    *CertManager    `json:"certManager,omitempty"`
	GitlabRunner   *GitlabRunner   `json:"gitlabRunner,omitempty"`
	Loki           *Loki           `json:"loki,omitempty"`
	Prometheus     *Prometheus     `json:"prometheus,omitempty"`
}

type KlOperators struct {
	// +kubebuilder:validation:Enum:=development;production
	InstallationMode string `json:"installationMode"`
	// +kubebuilder:default=kl-init-operators
	Namespace string `json:"namespace,omitempty"`
	// +kubebuilder:default=support@kloudlite.io
	ACMEEmail string `json:"acmeEmail,omitempty"`
	// +kubebuilder:default="kl-cert-issuer"
	ClusterIssuerName string `json:"clusterIssuerName,omitempty"`
	// +kubebuilder:default=kloudlite-cluster-svc-account
	ClusterSvcAccount string `json:"clusterSvcAccount,omitempty"`
	// +kubebuilder:default=v1.0.5
	ImageTag string `json:"imageTag,omitempty"`
	// +kubebuilder:default=Always
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}

type KloudliteCreds struct {
	DnsApiEndpoint string `json:"dnsApiEndpoint,omitempty" validate:"required"`
	DnsApiUsername string `json:"dnsApiUsername" validate:"required"`
	DnsApiPassword string `json:"dnsApiPassword" validate:"required"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ManagedCluster is the Schema for the managedclusters API
type ManagedCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedClusterSpec `json:"spec,omitempty"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (mc *ManagedCluster) EnsureGVK() {
	if mc != nil {
		mc.SetGroupVersionKind(GroupVersion.WithKind("ManagedCluster"))
	}
}

func (mc *ManagedCluster) GetStatus() *rApi.Status {
	return &mc.Status
}

func (mc *ManagedCluster) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ClusterSetupType: constants.ManagedClusterSetup,
	}
}

func (mc *ManagedCluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// ManagedClusterList contains a list of ManagedCluster
type ManagedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedCluster{}, &ManagedClusterList{})
}
