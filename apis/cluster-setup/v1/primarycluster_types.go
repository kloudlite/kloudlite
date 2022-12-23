package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/pkg/operator"
)

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
	SubDomain string `json:"subDomain,omitempty"`

	// Mongo
	MongoSvcName  string `json:"mongoSvcName,omitempty"`
	AuthDbName    string `json:"authDbName,omitempty"`
	ConsoleDbName string `json:"consoleDbName,omitempty"`
	CiDbName      string `json:"ciDbName,omitempty"`
	DnsDbName     string `json:"dnsDbName,omitempty"`
	FinanceDbName string `json:"financeDbName,omitempty"`
	IamDbName     string `json:"iamDbName,omitempty"`
	CommsDbName   string `json:"commsDbName,omitempty"`

	// Redis
	RedisSvcName     string `json:"redisSvcName,omitempty"`
	AuthRedisName    string `json:"authRedisName,omitempty"`
	ConsoleRedisName string `json:"consoleRedisName,omitempty"`
	CiRedisName      string `json:"ciRedisName,omitempty"`
	FinanceRedisName string `json:"financeRedisName,omitempty"`
	DnsRedisName     string `json:"dnsRedisName,omitempty"`
	IamRedisName     string `json:"iamRedisName,omitempty"`
	SocketRedisName  string `json:"socketRedisName,omitempty"`

	// Apps

	// API
	AppAuthApi       string `json:"appAuthApi,omitempty"`
	AppConsoleApi    string `json:"appConsoleApi,omitempty"`
	AppCiApi         string `json:"appCiApi,omitempty"`
	AppFinanceApi    string `json:"appFinanceApi,omitempty"`
	AppCommsApi      string `json:"appCommsApi,omitempty"`
	AppDnsApi        string `json:"appDnsApi,omitempty"`
	AppIAMApi        string `json:"appIAMApi,omitempty"`
	AppJsEvalApi     string `json:"appJsEval,omitempty"`
	AppGqlGatewayApi string `json:"appGqlGatewayApi,omitempty"`
	AppWebhooksApi   string `json:"appWebhooksApi,omitempty"`
	AppKlAgent       string `json:"appKlAgent,omitempty"`

	// Web
	AppAuthWeb     string `json:"appAuthWeb,omitempty"`
	AppAccountsWeb string `json:"appAccountsWeb,omitempty"`
	AppConsoleWeb  string `json:"appConsoleWeb,omitempty"`
	AppSocketWeb   string `json:"appSocketWeb,omitempty"`
	CookieDomain   string `json:"cookieDomain,omitempty"`

	// Secrets
	OAuthSecretName string `json:"oauthSecretName,omitempty"`

	// Images
	ImageAuthApi       string `json:"imageAuthApi,omitempty"`
	ImageConsoleApi    string `json:"imageConsoleApi,omitempty"`
	ImageCiApi         string `json:"imageCiApi,omitempty"`
	ImageFinanceApi    string `json:"imageFinanceApi,omitempty"`
	ImageCommsApi      string `json:"imageCommsApi,omitempty"`
	ImageDnsApi        string `json:"imageDnsApi,omitempty"`
	ImageIAMApi        string `json:"imageIAMApi,omitempty"`
	ImageJsEvalApi     string `json:"ImageJsEvalApi,omitempty"`
	ImageGqlGatewayApi string `json:"imageGqlGatewayApi,omitempty"`
	ImageWebhooksApi   string `json:"imageWebhooksApi,omitempty"`
	ImageAuthWeb       string `json:"imageAuthWeb,omitempty"`
	ImageAccountsWeb   string `json:"imageAccountsWeb,omitempty"`
	ImageConsoleWeb    string `json:"imageConsoleWeb,omitempty"`
	ImageKlAgent       string `json:"imageKlAgent,omitempty"`

	ImageSocketWeb               string `json:"imageSocketWeb,omitempty"`
	RedpandaAdminSecretName      string `json:"redpandaAdminSecretName,omitempty"`
	HarborAdminCredsSecretName   string `json:"harborAdminCredsSecretName,omitempty"`
	KafkaTopicGitWebhooks        string `json:"KafkaTopicGitWebhooks,omitempty"`
	KafkaTopicPipelineRunUpdates string `json:"KafkaTopicPipelineRunUpdates,omitempty"`
	KafkaTopicsStatusUpdates     string `json:"KafkaTopicsStatusUpdates,omitempty"`
	KafkaTopicBillingUpdates     string `json:"KafkaTopicBillingUpdates,omitempty"`
	KafkaTopicHarborWebhooks     string `json:"kafkaTopicHarborWebhooks,omitempty"`
	StatefulPriorityClass        string `json:"statefulPriorityClass,omitempty"`
	WebhookAuthzSecretName       string `json:"webhookAuthzSecretName,omitempty"`
	StripeSecretName             string `json:"stripeSecretName,omitempty"`

	// Routers
	AuthWebDomain     string `json:"authWebDomain,omitempty"`
	ConsoleWebDomain  string `json:"consoleWebDomain,omitempty"`
	AccountsWebDomain string `json:"accountsWebDomain,omitempty"`
	SocketWebDomain   string `json:"socketWebDomain,omitempty"`
	WebhookApiDomain  string `json:"webhookApiDomain,omitempty"`
}

// PrimaryClusterSpec defines the desired state of PrimaryCluster
type PrimaryClusterSpec struct {
	ClusterID    string            `json:"clusterId"`
	Networking   NetworkingValues  `json:"networking"`
	Domain       string            `json:"domain"`
	StorageClass string            `json:"storageClass,omitempty"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Tolerations  map[string]string `json:"tolerations,omitempty"`

	SecondaryClusterId     string `json:"secondaryClusterId,omitempty"`
	ShouldInstallSecondary bool   `json:"shouldInstallSecondary,omitempty"`

	StripeCreds       SecretReference   `json:"stripeCreds"`
	CloudflareCreds   Cloudflare        `json:"cloudflareCreds"`
	HarborAdminCreds  SecretReference   `json:"harborAdminCreds,omitempty"`
	WebhookAuthzCreds SecretReference   `json:"webhookAuthzCreds,omitempty"`
	ImgPullSecrets    []SecretReference `json:"imagePullSecrets"`
	LokiValues        LokiValues        `json:"loki"`
	PrometheusValues  PrometheusValues  `json:"prometheus"`
	CertManagerValues CertManagerValues `json:"certManager,omitempty"`
	IngressValues     IngressValues     `json:"ingress"`
	Operators         Operators         `json:"operators"`
	OAuthCreds        SecretReference   `json:"oAuthCreds"`
	RedpandaValues    RedpandaValues    `json:"redpanda"`

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
