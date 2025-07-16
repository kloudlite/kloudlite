package reconciler

const (
	ProjectDomain       = "kloudlite.io"
	OperatorLabelPrefix = "operator.kloudlite.io"
)

const (
	AnnotationShouldReconcileKey string = OperatorLabelPrefix + "/should-reconcile"
	AnnotationClearStatusKey     string = OperatorLabelPrefix + "/clear-status"
	AnnotationResetCheckKey      string = OperatorLabelPrefix + "/reset-check"
	AnnotationRestartKey         string = OperatorLabelPrefix + "/do-restart"
	// AnnotationDoHelmUpgrade      string = "kloudlite.io/do-helm-upgrade"
)

const (
	AnnotationResourceReady  string = OperatorLabelPrefix + "/resource.ready"
	AnnotationResourceChecks string = OperatorLabelPrefix + "/checks"
)

// Finalizers
const (
	Finalizer              string = "kloudlite.io/finalizer"
	ForegroundFinalizer    string = "foregroundDeletion"
	BillingFinalizer       string = "finalizers.kloudlite.io/billing-watcher"
	StatusWatcherFinalizer string = "finalizers.kloudlite.io/status-watcher"

	GenericFinalizer string = "kloudlite.io/finalizer"
)

// Generic Keys
const (
	AnnotationDescriptionKey string = "kloudlite.io/description"

	LastAppliedKey string = OperatorLabelPrefix + "/last-applied"
	GVKKey         string = "kloudlite.io/group-version-kind"
)

const (
	KloudliteDNSHostnameKey string = "kloudlite.io/dns.hostname"
)

const (
	ObservabilityAnnotationKey string = "kloudlite.io/observability"
)
