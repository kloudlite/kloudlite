package constants

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const MsvcApiVersion = "msvc.kloudlite.io/v1"

const (
	HelmMongoDBKind string = "HelmMongoDB"

	HelmMySqlDBKind string = "HelmMySqlDB"
	HelmRedisKind   string = "HelmRedis"
)

const (
	CommonFinalizer     string = "finalizers.kloudlite.io"
	ForegroundFinalizer string = "foregroundDeletion"
)

type StorageClass string

const (
	DoBlockStorage    StorageClass = "do-block-storage"
	DoBlockStorageXFS StorageClass = "do-block-storage-xfs"
)

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
)

var (
	KnativeServiceType = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "serving.knative.dev/v1",
	}
)

var (
	ConditionReady = struct {
		Type, InitReason, InProgressReason, ErrorReason, SuccessReason string
	}{
		Type:             "Ready",
		InitReason:       "Initialized",
		InProgressReason: "ReconcilationInProgress",
		ErrorReason:      "SomeChecksFailed",
		SuccessReason:    "AllChecksCompleted",
	}
)

const (
	DefaultIngressClass  = "nginx"
	DefaultClusterIssuer = "kl-cert-issuer"
)
