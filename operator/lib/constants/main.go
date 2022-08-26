package constants

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const MsvcApiVersion = "msvc.kloudlite.io/v1"

const (
	CommonFinalizer        string = "finalizers.kloudlite.io"
	ForegroundFinalizer    string = "foregroundDeletion"
	BillingFinalizer       string = "finalizers.kloudlite.io/billing"
	StatusWatcherFinalizer string = "finalizers.kloudlite.io/status-watcher"
)

var LabelKeys = struct {
	HarborProjectRef string
	Freeze           string
	IsIntercepted    string
	DeviceRef        string
}{
	HarborProjectRef: "artifacts.kloudlite.io/harbor-project-ref",
	Freeze:           "kloudlite.io/freeze",
	IsIntercepted:    "kloudlite.io/is-intercepted",
	DeviceRef:        "kloudlite.io/intercept.device-ref",
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
	BillingPlan:      "kloudlite.io/billing-plan",
	BillableQuantity: "kloudlite.io/billable-quantity",
	GroupVersionKind: "kloudlite.io/group-version-kind",
	IsShared:         "kloudlite.io/is-shared",

	Restart: "kloudlite.io/do-restart",
}

var (
	PodGroup = metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Pod",
	}

	DeploymentType = metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}

	StatefulsetType = metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "StatefulSet",
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

	HelmOpenSearchType = metav1.TypeMeta{
		Kind:       "HelmOpenSearch",
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
)

var (
	KnativeServiceType = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "serving.knative.dev/v1",
	}
)

const (
	DefaultIngressClass  = "nginx"
	DefaultClusterIssuer = "kl-cert-issuer"
)
