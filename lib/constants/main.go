package constants

const MsvcApiVersion = "msvc.kloudlite.io/v1"

const (
	HelmMongoDBKind string = "HelmMongoDB"
	HelmMySqlDBKind string = "HelmMySqlDB"
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
