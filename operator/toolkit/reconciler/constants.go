package reconciler

const (
	AnnotationShouldReconcileKey string = "kloudlite.io/operator.should-reconcile"
	AnnotationClearStatusKey     string = "kloudlite.io/operator.clear-status"
	AnnotationResetCheckKey      string = "kloudlite.io/operator.reset-check"
	AnnotationRestartKey         string = "kloudlite.io/do-restart"
	// AnnotationDoHelmUpgrade      string = "kloudlite.io/do-helm-upgrade"
)

const (
	AnnotationResourceReady  string = "kloudlite.io/operator.resource.ready"
	AnnotationResourceChecks string = "kloudlite.io/operator.checks"
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

	LastAppliedKey string = "kloudlite.io/last-applied"
	GVKKey         string = "kloudlite.io/group-version-kind"
)

const (
	KloudliteDNSHostnameKey string = "kloudlite.io/dns.hostname"
)
