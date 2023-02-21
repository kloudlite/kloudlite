package constants

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const MsvcApiVersion = "msvc.kloudlite.io/v1"

const (
	CommonFinalizer        string = "finalizers.kloudlite.io"
	ForegroundFinalizer    string = "foregroundDeletion"
	BillingFinalizer       string = "finalizers.kloudlite.io/billing-watcher"
	StatusWatcherFinalizer string = "finalizers.kloudlite.io/status-watcher"
)

var LabelKeys = struct {
	HarborProjectRef string
	Freeze           string
	IsIntercepted    string
	DeviceRef        string
	ProjectName      string
	AppName          string
	CsiForEdge       string
}{
	HarborProjectRef: "artifacts.kloudlite.io/harbor-project-ref",
	Freeze:           "kloudlite.io/freeze",
	IsIntercepted:    "kloudlite.io/is-intercepted",
	DeviceRef:        "kloudlite.io/intercept.device-ref",

	ProjectName: "kloudlite.io/project.name",
	AppName:     "kloudlite.io/app.name",
	CsiForEdge:  "kloudlite.io/csi-for-edge",
}

var AnnotationKeys = struct {
	AccountRef       string
	ProjectRef       string
	ResourceRef      string
	BillingPlan      string
	BillableQuantity string
	GroupVersionKind string
	IsShared         string

	Restart string
}{
	AccountRef:       "kloudlite.io/account-ref",
	ProjectRef:       "kloudlite.io/project-ref",
	ResourceRef:      "kloudlite.io/resource-ref",
	BillingPlan:      "kloudlite.io/billing-watcher-plan",
	BillableQuantity: "kloudlite.io/billable-quantity",
	GroupVersionKind: "kloudlite.io/group-version-kind",
	IsShared:         "kloudlite.io/is-shared",

	Restart: "kloudlite.io/do-restart",
}

const (
	AccountRef      string = "kloudlite.io/account-ref"
	ProjectRef      string = "kloudlite.io/project-ref"
	ProviderRef     string = "kloudlite.io/provider-ref"
	EnvironmentRef  string = "kloudlite.io/environment-ref"
	ResourceRef     string = "kloudlite.io/resource-ref"
	ShouldReconcile string = "kloudlite.io/should-reconcile"

	ProjectNameKey       string = "kloudlite.io/project.name"
	BlueprintNameKey     string = "kloudlite.io/blueprint.name"
	MsvcNameKey          string = "kloudlite.io/msvc.name"
	IsMresOutput         string = "kloudlite.io/is-mres-output"
	MresNameKey          string = "kloudlite.io/mres.name"
	AppNameKey           string = "kloudlite.io/app.name"
	RouterNameKey        string = "kloudlite.io/router.name"
	LambdaNameKey        string = "kloudlite.io/lambda.name"
	AccountRouterNameKey string = "kloudlite.io/account-router.name"
	EdgeNameKey          string = "kloudlite.io/edge.name"
	EdgeRouterNameKey    string = "kloudlite.io/edge-router.name"
	EnvNameKey           string = "kloudlite.io/env.name"
	CsiDriverNameKey     string = "kloudlite.io/csi-driver.name"

	ClearStatusKey string = "kloudlite.io/clear-status"
	ResetCheckKey  string = "kloudlite.io/reset-check"
	RestartKey     string = "kloudlite.io/do-restart"
	RegionKey      string = "kloudlite.io/region"

	IsBluePrintKey    string = "kloudlite.io/is-blueprint"
	MarkedAsBlueprint string = "kloudlite.io/marked-as-blueprint"

	LastAppliedKey string = "kloudlite.io/last-applied"

	GVKKey string = "kloudlite.io/group-version-kind"

	ClusterSetupType string = "kloudlite.io/cluster.setup-type"
)

// ClusterSetupTypes
const (
	ManagedClusterSetup   = "managed"
	PrimaryClusterSetup   = "primary"
	SecondaryClusterSetup = "secondary"
)

var (
	K8sConfigType = metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	}

	K8sSecretType = metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}
)

var (
	HelmMongoDBType = metav1.TypeMeta{
		APIVersion: MsvcApiVersion,
		Kind:       "HelmMongoDB",
	}

	HelmRedisType = metav1.TypeMeta{
		APIVersion: MsvcApiVersion,
		Kind:       "HelmRedis",
	}

	HelmMysqlType = metav1.TypeMeta{
		APIVersion: MsvcApiVersion,
		Kind:       "HelmMySqlDB",
	}

	HelmElasticType = metav1.TypeMeta{
		Kind:       "HelmElasticSearch",
		APIVersion: MsvcApiVersion,
	}

	HelmKibanaType = metav1.TypeMeta{
		Kind:       "HelmKibana",
		APIVersion: MsvcApiVersion,
	}

	HelmOpenSearchType = metav1.TypeMeta{
		Kind:       "HelmOpenSearch",
		APIVersion: MsvcApiVersion,
	}

	HelmZookeeperType = metav1.TypeMeta{
		Kind:       "HelmZookeeper",
		APIVersion: MsvcApiVersion,
	}

	HelmInfluxDBType = metav1.TypeMeta{
		Kind:       "HelmInfluxDB",
		APIVersion: MsvcApiVersion,
	}

	RedpandaClusterType = metav1.TypeMeta{
		Kind:       "Cluster",
		APIVersion: "redpanda.vectorized.io/v1alpha1",
	}

	HelmNeo4JStandaloneType = metav1.TypeMeta{
		Kind:       "HelmNeo4jStandalone",
		APIVersion: MsvcApiVersion,
	}

	EdgeInfraType = metav1.TypeMeta{
		Kind:       "Edge",
		APIVersion: "infra.kloudlite.io/v1",
	}
	CloudProviderType = metav1.TypeMeta{
		Kind:       "CloudProvider",
		APIVersion: "infra.kloudlite.io/v1",
	}
	DeviceType = metav1.TypeMeta{
		Kind:       "Device",
		APIVersion: "management.kloudlite.io/v1",
	}

	HelmAwsEbsCsiKind = metav1.TypeMeta{
		Kind:       "AwsEbsCsiDriver",
		APIVersion: "csi.helm.kloudlite.io/v1",
	}
	HelmDigitaloceanCsiKind = metav1.TypeMeta{
		Kind:       "DigitaloceanCSIDriver",
		APIVersion: "csi.helm.kloudlite.io/v1",
	}
)

var (
	HelmIngressNginx = metav1.TypeMeta{
		Kind:       "Nginx",
		APIVersion: "ingress.kloudlite.io/v1",
	}

	TektonPipelineRunKind = metav1.TypeMeta{
		Kind:       "PipelineRun",
		APIVersion: "tekton.dev/v1beta1",
	}
)

var (
	KloudliteAccountType = metav1.TypeMeta{
		Kind:       "Account",
		APIVersion: "management.kloudlite.io/v1",
	}
)

var (
	KnativeServiceType = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "serving.knative.dev/v1",
	}

	ClusterIssuerType = metav1.TypeMeta{
		Kind:       "ClusterIssuer",
		APIVersion: "cert-manager.io/v1",
	}

	StorageClassType = metav1.TypeMeta{
		APIVersion: "storage.k8s.io/v1",
		Kind:       "StorageClass",
	}
)

const (
	DefaultIngressClass  = "nginx"
	DefaultClusterIssuer = "kl-cert-issuer"
)
