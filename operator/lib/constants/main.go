package constants

const MsvcApiVersion = "msvc.kloudlite.io/v1"

const (
	HelmMongoDBKind string = "HelmMongoDB"
)

var (
	ConditionReady = struct{ Type, InitReason, ErrorReason, SuccessReason string }{
		Type:          "Ready",
		ErrorReason:   "SomeChecksFailed",
		SuccessReason: "AllChecksCompleted",
	}
)

const (
	ConditionDefaultMessage string = "ErrWhileReconciling"
)
