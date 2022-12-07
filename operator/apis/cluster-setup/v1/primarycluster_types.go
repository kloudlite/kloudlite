package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/pkg/operator"
)

type NamespacedReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type SecretReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type SecretKeyReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key"`
}

type SharedConstants struct {
	// Mongo
	MongoSvcName  string `json:"mongoSvcName"`
	AuthDbName    string `json:"authDbName"`
	ConsoleDbName string `json:"consoleDbName"`
	CiDbName      string `json:"ciDbName"`
	DnsDbName     string `json:"dnsDbName"`
	FinanceDbName string `json:"financeDbName"`
	IamDbName     string `json:"iamDbName"`
	CommsDbName   string `json:"commsDbName"`

	// Redis
	RedisSvcName     string `json:"redisSvcName"`
	AuthRedisName    string `json:"authRedisName"`
	ConsoleRedisName string `json:"consoleRedisName"`
	CiRedisName      string `json:"ciRedisName"`
	DnsRedisName     string `json:"dnsRedisName"`
	IamRedisName     string `json:"iamRedisName"`
	SocketRedisName  string `json:"socketRedisName"`

	// Apps

	// API
	AppAuthApi       string `json:"appAuthApi"`
	AppConsoleApi    string `json:"appConsoleApi"`
	AppCiApi         string `json:"appCiApi"`
	AppFinanceApi    string `json:"appFinanceApi"`
	AppCommsApi      string `json:"appCommsApi"`
	AppDnsApi        string `json:"appDnsApi"`
	AppIAMApi        string `json:"appIAMApi"`
	AppJsEvalApi     string `json:"appJsEval"`
	AppGqlGatewayApi string `json:"appGqlGatewayApi"`
	AppWebhooksApi   string `json:"appWebhooksApi"`

	// Web
	AppAuthWeb     string `json:"appAuthWeb"`
	AppAccountsWeb string `json:"appAccountsWeb"`
	AppConsoleWeb  string `json:"appConsoleWeb"`
	AppSocketWeb   string `json:"appSocketWeb"`
	CookieDomain   string `json:"cookieDomain"`

	// Secrets
	OAuthSecretName string `json:"oauthSecretName"`
	AuthWebDomain   string `json:"authWebDomain"`
}

// PrimaryClusterSpec defines the desired state of PrimaryCluster
type PrimaryClusterSpec struct {
	ClusterID    string `json:"clusterId"`
	Domain       string `json:"domain"`
	StorageClass string `json:"storageClass,omitempty"`

	CloudflareCreds   Cloudflare            `json:"cloudflareCreds"`
	HarborAdminCreds  SecretReference       `json:"harborAdminCreds,omitempty"`
	ImgPullSecrets    []NamespacedReference `json:"imagePullSecrets"`
	LokiValues        LokiValues            `json:"loki"`
	PrometheusValues  PrometheusValues      `json:"prometheus"`
	CertManagerValues CertManagerValues     `json:"certManager,omitempty"`
	IngressValues     IngressValues         `json:"ingress"`
	Operators         Operators             `json:"operators"`
	OAuthCreds        SecretReference       `json:"oAuthCreds"`

	SharedConstants *SharedConstants `json:"sharedConstants,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// PrimaryCluster is the Schema for the primaryclusters API
type PrimaryCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrimaryClusterSpec `json:"spec,omitempty"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (p *PrimaryCluster) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *PrimaryCluster) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (p *PrimaryCluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// PrimaryClusterList contains a list of PrimaryCluster
type PrimaryClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PrimaryCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PrimaryCluster{}, &PrimaryClusterList{})
}
