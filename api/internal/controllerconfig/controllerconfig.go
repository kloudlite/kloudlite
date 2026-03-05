package controllerconfig

import "time"

// ControllerConfig contains configuration for all controllers
type ControllerConfig struct {
	// Workspace controller configuration
	Workspace WorkspaceConfig

	// Environment controller configuration
	Environment EnvironmentConfig

	// WorkMachine controller configuration
	WorkMachine WorkMachineConfig

	// WMIngress controller configuration
	WMIngress WMIngressConfig

	// Snapshot controller configuration
	Snapshot SnapshotConfig
}

// WorkspaceConfig contains workspace controller configuration
type WorkspaceConfig struct {
	// DefaultIdleTimeoutMinutes is default idle timeout before auto-stopping a workspace
	// Default: 30 minutes
	DefaultIdleTimeoutMinutes int

	// RequeueIntervalMinutes is how often to requeue workspaces for idle checking
	// Default: 1 minute
	RequeueIntervalMinutes int

	// RBACCleanupIntervalMinutes is how often to run orphaned RBAC cleanup
	// Default: 60 minutes
	RBACCleanupIntervalMinutes int

	// KubectlImage is the image used for kubectl operations
	// Default: bitnami/kubectl:latest
	KubectlImage string

	// GitImage is the image used for git operations
	// Default: alpine/git:latest
	GitImage string

	// AlpineImage is the image used for Alpine-based operations
	// Default: alpine:latest
	AlpineImage string

	// CleanupPodTTLSeconds is how long cleanup pods are kept
	// Default: 300 seconds (5 minutes)
	CleanupPodTTLSeconds int64

	// VSCodeVersion is the default VS Code version for workspaces
	// Default: latest
	VSCodeVersion string

	// Derived fields (not from env vars)
	DefaultRequeueInterval     time.Duration
	RBACCleanupRetryInterval time.Duration
}

// EnvironmentConfig contains environment controller configuration
type EnvironmentConfig struct {
	// PodTerminationRetryInterval is how long to wait between pod termination checks
	// Default: 2 seconds
	PodTerminationRetryInterval time.Duration

	// SnapshotRestoreRetryInterval is how long to wait between snapshot restore retries
	// Default: 2 seconds
	SnapshotRestoreRetryInterval time.Duration

	// SnapshotRequestRetryInterval is how long to wait between snapshot request retries
	// Default: 2 seconds
	SnapshotRequestRetryInterval time.Duration

	// ForkRetryInterval is how long to wait between fork operation retries
	// Default: 5 seconds
	ForkRetryInterval time.Duration

	// StatusUpdateRetryInterval is how long to wait between status update retries
	// Default: 5 seconds
	StatusUpdateRetryInterval time.Duration

	// DeletionRetryInterval is how long to wait between deletion retries
	// Default: 5 seconds
	DeletionRetryInterval time.Duration

	// LifecycleRetryInterval is how long to wait between lifecycle operation retries
	// Default: 5 seconds
	LifecycleRetryInterval time.Duration

	// Derived fields (not from env vars)
	DefaultRequeueInterval      time.Duration
	StatefulSetScaleTimeout   time.Duration
	SnapshotRestoreWaitInterval time.Duration
}

// WorkMachineConfig contains workmachine controller configuration
type WorkMachineConfig struct {
	// WMIngressControllerImage is the image for wm-ingress-controller
	// Default: ghcr.io/kloudlite/kloudlite/wm-ingress-controller:development
	WMIngressControllerImage string

	// SSHUserName is the username for SSH access to workmachine nodes
	// Default: kloudlite
	SSHUserName string

	// DefaultWildcardCertName is the default wildcard TLS certificate secret name
	// Default: kloudlite-wildcard-cert-tls
	DefaultWildcardCertName string

	// CloudOperationRetryInterval is how long to wait between cloud operation retries
	// Default: 5 seconds
	CloudOperationRetryInterval time.Duration

	// MachineStatusCheckInterval is how long to wait between machine status checks
	// Default: 5 seconds
	MachineStatusCheckInterval time.Duration

	// MachineStartupRetryInterval is how long to wait between machine startup retries
	// Default: 10 seconds
	MachineStartupRetryInterval time.Duration

	// NodeJoinRetryInterval is how long to wait between node join checks
	// Default: 10 seconds
	NodeJoinRetryInterval time.Duration

	// VolumeResizeRetryInterval is how long to wait between volume resize checks
	// Default: 10 seconds
	VolumeResizeRetryInterval time.Duration

	// MachineTypeChangeRetryInterval is how long to wait between machine type change retries
	// Default: 5 seconds
	MachineTypeChangeRetryInterval time.Duration

	// AutoShutdownCheckInterval is how often to check for auto-shutdown
	// Default: 5 minutes
	AutoShutdownCheckInterval time.Duration

	// AutoShutdownIdleThresholdMinutes is how long a workmachine can be idle before auto-shutdown
	// Default: 30 minutes
	AutoShutdownIdleThresholdMinutes int

	// AutoShutdownWarningMinutes is how many minutes before shutdown to send warning
	// Default: 5 minutes
	AutoShutdownWarningMinutes int

	// CloudMachineCreationRetryInterval is how long to wait after creating a cloud machine
	// Default: 2 seconds
	CloudMachineCreationRetryInterval time.Duration

	// CloudMachineStopRetryInterval is how long to wait after stopping a machine
	// Default: 10 seconds
	CloudMachineStopRetryInterval time.Duration

	// CloudMachineStartRetryInterval is how long to wait after starting a machine
	// Default: 10 seconds
	CloudMachineStartRetryInterval time.Duration

	// NodeDrainRetryInterval is how long to wait between node drain retries
	// Default: 5 seconds
	NodeDrainRetryInterval time.Duration

	// NodeDeleteRetryInterval is how long to wait after forcing node deletion
	// Default: 2 seconds
	NodeDeleteRetryInterval time.Duration

	// VolumeResizeCheckInterval is how long to wait for volume resize checks
	// Default: 10 seconds
	VolumeResizeCheckInterval time.Duration

	// NodeJoinCheckInterval is how long to wait for node to join cluster
	// Default: 10 seconds
	NodeJoinCheckInterval time.Duration

	// NodeReadyRetryInterval is how long to wait for node to be ready
	// Default: 5 seconds
	NodeReadyRetryInterval time.Duration

	// AutoShutdownTriggerRetryInterval is how long to wait after triggering auto-shutdown
	// Default: 5 seconds
	AutoShutdownTriggerRetryInterval time.Duration
}

// WMIngressConfig contains wm-ingress controller configuration
type WMIngressConfig struct {
	// HTTPPort is the HTTP port for the ingress controller
	// Default: 80
	HTTPPort int

	// HTTPSPort is the HTTPS port for the ingress controller
	// Default: 443
	HTTPSPort int

	// WildcardDomain is the wildcard domain for TLS certificates (e.g., khost.dev)
	// Empty string means no domain filtering
	WildcardDomain string

	// WildcardSecretName is the name of the wildcard TLS certificate secret
	// Default: kloudlite-wildcard-cert-tls
	WildcardSecretName string

	// WildcardSecretNamespace is the namespace of the wildcard TLS certificate secret
	// Default: kloudlite
	WildcardSecretNamespace string

	// RegistryUsername is the username for registry path access control
	// When set, write operations to cr.* domains are restricted to /v2/{username}/*
	// Empty string means no access control
	RegistryUsername string

	// ForceFullRebuild forces full rebuild on every event (for debugging)
	// Default: false
	ForceFullRebuild bool

	// ProxyTimeout is the timeout for HTTP proxy connections
	// Default: 30 seconds
	ProxyTimeout time.Duration

	// ProxyKeepAlive is the keep-alive duration for HTTP proxy connections
	// Default: 30 seconds
	ProxyKeepAlive time.Duration

	// ProxyIdleConnTimeout is the idle connection timeout for HTTP proxy
	// Default: 90 seconds
	ProxyIdleConnTimeout time.Duration

	// ProxyTLSHandshakeTimeout is the TLS handshake timeout for HTTP proxy
	// Default: 10 seconds
	ProxyTLSHandshakeTimeout time.Duration

	// ProxyExpectContinueTimeout is the expect continue timeout for HTTP proxy
	// Default: 1 second
	ProxyExpectContinueTimeout time.Duration

	// ProxyMaxIdleConns is the maximum number of idle connections
	// Default: 100
	ProxyMaxIdleConns int
}

// SnapshotConfig contains snapshot controller configuration
type SnapshotConfig struct {
	// DefaultRequeueInterval is the default requeue interval for snapshot operations
	// Default: 30 seconds
	DefaultRequeueInterval time.Duration

	// StorageCleanupRetryInterval is how long to wait between storage cleanup retries
	// Default: 5 seconds
	StorageCleanupRetryInterval time.Duration

	// SnapshotReadyCheckInterval is how long to wait between snapshot ready checks
	// Default: 5 seconds
	SnapshotReadyCheckInterval time.Duration

	// SnapshotRestoreWaitInterval is how long to wait for snapshot to become ready before restore
	// Default: 5 seconds
	SnapshotRestoreWaitInterval time.Duration

	// SnapshotRestoreStatusRetryInterval is how long to wait between snapshot restore status updates
	// Default: 5 seconds
	SnapshotRestoreStatusRetryInterval time.Duration
}
