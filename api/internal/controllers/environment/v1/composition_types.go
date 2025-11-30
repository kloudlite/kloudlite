package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PortMapping defines mapping between service and workspace ports for intercepts
type PortMapping struct {
	// ServicePort is the port exposed by the service
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ServicePort int32 `json:"servicePort"`

	// WorkspacePort is the port in the workspace pod
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	WorkspacePort int32 `json:"workspacePort"`

	// Protocol is the protocol used (TCP/UDP)
	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	// +kubebuilder:default=TCP
	// +optional
	Protocol corev1.Protocol `json:"protocol,omitempty"`
}

// ServiceInterceptConfig defines intercept configuration for a composition service
type ServiceInterceptConfig struct {
	// ServiceName is the name of the service in the composition to intercept
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ServiceName string `json:"serviceName"`

	// PortMappings defines how service ports map to workspace ports
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	PortMappings []PortMapping `json:"portMappings"`

	// Enabled indicates whether this intercept is currently active
	// +kubebuilder:default=false
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// WorkspaceRef references the workspace that will receive intercepted traffic
	// This is set when a workspace requests to intercept this service
	// +optional
	WorkspaceRef *corev1.ObjectReference `json:"workspaceRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={kloudlite,environments},shortName=comp
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Services",type=integer,JSONPath=`.status.servicesCount`
// +kubebuilder:printcolumn:name="Running",type=integer,JSONPath=`.status.runningCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Composition represents a Docker Compose application deployed in an environment
type Composition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CompositionSpec   `json:"spec,omitempty"`
	Status CompositionStatus `json:"status,omitempty"`
}

// CompositionSpec defines the desired state of Composition
type CompositionSpec struct {
	// DisplayName is the human-readable name for the composition
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	DisplayName string `json:"displayName"`

	// Description provides additional information about the composition
	// +kubebuilder:validation:MaxLength=500
	// +optional
	Description string `json:"description,omitempty"`

	// ComposeContent contains the docker-compose.yml file content
	// +kubebuilder:validation:Required
	ComposeContent string `json:"composeContent"`

	// ComposeFormat specifies the version/format of the compose file
	// +kubebuilder:validation:Enum=v2;v3;v3.1;v3.2;v3.3;v3.4;v3.5;v3.6;v3.7;v3.8;v3.9
	// +kubebuilder:default=v3.8
	// +optional
	ComposeFormat string `json:"composeFormat,omitempty"`

	// EnvVars are environment variables to inject into all services
	// +optional
	EnvVars map[string]string `json:"envVars,omitempty"`

	// EnvFrom references ConfigMaps or Secrets to use as environment variables
	// +optional
	EnvFrom []EnvFromSource `json:"envFrom,omitempty"`

	// AutoDeploy indicates whether changes should auto-deploy
	// +kubebuilder:default=false
	// +optional
	AutoDeploy bool `json:"autoDeploy,omitempty"`

	// ResourceOverrides allows overriding resource limits for specific services
	// +optional
	ResourceOverrides map[string]ServiceResourceOverride `json:"resourceOverrides,omitempty"`

	// Intercepts defines service intercept configurations for this composition
	// This allows workspace pods to intercept traffic destined for composition services
	// +optional
	Intercepts []ServiceInterceptConfig `json:"intercepts,omitempty"`

	// NodeName specifies the node where all composition services should run
	// Inherited from the Environment's NodeName
	// +optional
	NodeName string `json:"nodeName,omitempty"`
}

// EnvFromSource represents a source for environment variables
type EnvFromSource struct {
	// Type specifies the source type (ConfigMap or Secret)
	// +kubebuilder:validation:Enum=ConfigMap;Secret
	Type string `json:"type"`

	// Name of the ConfigMap or Secret
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Prefix to prepend to all keys from this source
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// ServiceResourceOverride allows overriding resources for a specific service
type ServiceResourceOverride struct {
	// CPU limit override
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory limit override
	// +optional
	Memory string `json:"memory,omitempty"`

	// Replicas override for this service
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
}

// CompositionStatus defines the observed state of Composition
type CompositionStatus struct {
	// State represents the current state of the composition
	// +kubebuilder:validation:Enum=pending;deploying;running;degraded;stopped;failed;deleting
	State CompositionState `json:"state,omitempty"`

	// Message provides human-readable information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// ServicesCount is the total number of services defined
	// +optional
	ServicesCount int32 `json:"servicesCount,omitempty"`

	// RunningCount is the number of services currently running
	// +optional
	RunningCount int32 `json:"runningCount,omitempty"`

	// Services tracks the status of individual services
	// +optional
	Services []ServiceStatus `json:"services,omitempty"`

	// ActiveIntercepts tracks the status of active service intercepts
	// +optional
	ActiveIntercepts []InterceptStatus `json:"activeIntercepts,omitempty"`

	// Endpoints contains access URLs for exposed services
	// +optional
	Endpoints map[string]string `json:"endpoints,omitempty"`

	// LastDeployedTime is when the composition was last deployed
	// +optional
	LastDeployedTime *metav1.Time `json:"lastDeployedTime,omitempty"`

	// Conditions represent the latest available observations of the composition's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// EnvironmentActivated tracks whether the environment was activated when this composition was last reconciled
	// This is used to detect when environment activation state changes and trigger reconciliation
	// +optional
	EnvironmentActivated bool `json:"environmentActivated,omitempty"`

	// DeployedResources tracks the Kubernetes resources created for this composition
	// +optional
	DeployedResources *DeployedResources `json:"deployedResources,omitempty"`
}

// CompositionState represents the state of a composition
type CompositionState string

const (
	// CompositionStatePending means the composition is pending deployment
	CompositionStatePending CompositionState = "pending"

	// CompositionStateDeploying means the composition is being deployed
	CompositionStateDeploying CompositionState = "deploying"

	// CompositionStateRunning means all services are running
	CompositionStateRunning CompositionState = "running"

	// CompositionStateDegraded means some services are not running
	CompositionStateDegraded CompositionState = "degraded"

	// CompositionStateStopped means the composition is stopped
	CompositionStateStopped CompositionState = "stopped"

	// CompositionStateFailed means the composition deployment failed
	CompositionStateFailed CompositionState = "failed"

	// CompositionStateDeleting means the composition is being deleted
	CompositionStateDeleting CompositionState = "deleting"
)

// InterceptStatus tracks the status of an active service intercept
type InterceptStatus struct {
	// ServiceName is the service being intercepted
	ServiceName string `json:"serviceName"`

	// WorkspaceName is the workspace intercepting traffic
	// +optional
	WorkspaceName string `json:"workspaceName,omitempty"`

	// WorkspaceNamespace is the namespace of the workspace intercepting traffic
	// +optional
	WorkspaceNamespace string `json:"workspaceNamespace,omitempty"`

	// SOCATPodName is the name of the SOCAT forwarding pod
	// +optional
	SOCATPodName string `json:"socatPodName,omitempty"`

	// OriginalServiceSelector stores the original service selector before interception
	// +optional
	OriginalServiceSelector map[string]string `json:"originalServiceSelector,omitempty"`

	// WorkspaceHeadlessServiceName is the name of the headless service for the workspace
	// +optional
	WorkspaceHeadlessServiceName string `json:"workspaceHeadlessServiceName,omitempty"`

	// Phase represents the current phase of the intercept
	// +kubebuilder:validation:Enum=creating;active;failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides additional information about the intercept status
	// +optional
	Message string `json:"message,omitempty"`

	// InterceptStartTime when the intercept was activated
	// +optional
	InterceptStartTime *metav1.Time `json:"interceptStartTime,omitempty"`
}

// ServiceStatus tracks the status of an individual service
type ServiceStatus struct {
	// Name of the service
	Name string `json:"name"`

	// State of the service
	// +kubebuilder:validation:Enum=pending;starting;running;stopped;failed
	State string `json:"state"`

	// Replicas is the number of replicas for this service
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the number of ready replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Image used by this service
	// +optional
	Image string `json:"image,omitempty"`

	// Ports exposed by this service
	// +optional
	Ports []int32 `json:"ports,omitempty"`

	// Message provides additional status information
	// +optional
	Message string `json:"message,omitempty"`
}

// DeployedResources tracks Kubernetes resources created for the composition
type DeployedResources struct {
	// Deployments created
	// +optional
	Deployments []string `json:"deployments,omitempty"`

	// Services created
	// +optional
	Services []string `json:"services,omitempty"`

	// ConfigMaps created
	// +optional
	ConfigMaps []string `json:"configMaps,omitempty"`

	// Secrets created
	// +optional
	Secrets []string `json:"secrets,omitempty"`

	// PVCs created
	// +optional
	PVCs []string `json:"pvcs,omitempty"`

	// NetworkPolicies created
	// +optional
	NetworkPolicies []string `json:"networkPolicies,omitempty"`
}

// +kubebuilder:object:root=true

// CompositionList contains a list of Composition
type CompositionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Composition `json:"items"`
}
