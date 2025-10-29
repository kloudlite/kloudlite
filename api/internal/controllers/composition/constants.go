package composition

const (
	// Finalizers
	compositionFinalizer = "dockercompositions.environments.kloudlite.io/finalizer"

	// Labels and annotations
	dockerCompositionLabel     = "kloudlite.io/docker-composition"
	originalReplicasAnnotation = "kloudlite.io/original-replicas"
	managedLabel               = "kloudlite.io/managed"
	serviceLabel               = "kloudlite.io/service"
	volumeLabel                = "kloudlite.io/volume"

	// Environment configuration
	envConfigConfigMapName = "env-config"
	envSecretSecretName    = "env-secret"
	envFileTypeLabel       = "kloudlite.io/file-type"
	envFileConfigMapPrefix = "env-file-"

	// Status condition types
	readyConditionType = "Ready"

	// Requeue intervals (in nanoseconds for time.Duration)
	deletionRequeueInterval  = 5 * 1000000000
	deployingRequeueInterval = 10 * 1000000000
)
