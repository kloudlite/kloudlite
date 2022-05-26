package constants

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const MsvcApiVersion = "msvc.kloudlite.io/v1"

const (
	HelmMongoDBKind string = "HelmMongoDB"

	HelmMySqlDBKind string = "HelmMySqlDB"
	HelmRedisKind   string = "HelmRedis"
)

var (
	DeploymentGroup metav1.GroupVersionKind = metav1.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}
	HelmMongoDBGroup metav1.GroupVersionKind = metav1.GroupVersionKind{
		Group:   "mongodb-standalone.msvc.kloudlite.io",
		Version: "v1",
		Kind:    "HelmMongoDB",
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
