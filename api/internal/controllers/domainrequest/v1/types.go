package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="IP Address",type=string,JSONPath=`.spec.ipAddress`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.status.domain`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// DomainRequest represents a request to register an IP address with console.kloudlite.io
// and fetch TLS certificates for the domain
type DomainRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DomainRequestSpec   `json:"spec,omitempty"`
	Status DomainRequestStatus `json:"status,omitempty"`
}

// DomainRequestSpec defines the desired state of DomainRequest
type DomainRequestSpec struct {
	// IPAddress is the IP address to register
	// If not provided, will be auto-detected from the LoadBalancer service
	// +optional
	IPAddress string `json:"ipAddress,omitempty"`

	// NodeName is the name of the node where HAProxy pod should be scheduled
	// The node should have the public IP address
	// +optional
	NodeName string `json:"nodeName,omitempty"`

	// LoadBalancerServiceName is the name of the LoadBalancer service to watch for IP
	// Used for auto-detecting the IP address
	// +optional
	LoadBalancerServiceName string `json:"loadBalancerServiceName,omitempty"`

	// LoadBalancerServiceNamespace is the namespace of the LoadBalancer service
	// +optional
	LoadBalancerServiceNamespace string `json:"loadBalancerServiceNamespace,omitempty"`

	// CertificateScope defines the scope for certificate generation
	// +kubebuilder:validation:Enum=installation;workmachine;workspace
	// +kubebuilder:default=installation
	CertificateScope string `json:"certificateScope"`

	// CertificateScopeIdentifier is the identifier for the certificate scope
	// (e.g., workmachine name or workspace name)
	// +optional
	CertificateScopeIdentifier string `json:"certificateScopeIdentifier,omitempty"`

	// CertificateParentScopeIdentifier is the parent scope identifier
	// (e.g., workmachine name for workspace certificates)
	// +optional
	CertificateParentScopeIdentifier string `json:"certificateParentScopeIdentifier,omitempty"`

	// OriginCertificateHostnames specifies the hostnames to include in the origin certificate
	// If not provided, defaults to ["subdomain.domain", "*.subdomain.domain"]
	// Note: Cloudflare only allows ONE wildcard per hostname at the beginning
	// Valid: ["example.com", "*.example.com"]
	// Invalid: ["*.*.example.com", "test.*.example.com"]
	// +optional
	OriginCertificateHostnames []string `json:"originCertificateHostnames,omitempty"`

	// SSHProxyEnabled enables SSH proxy functionality (port 22)
	// Will be handled by a sidecar container in future implementation
	// +kubebuilder:default=false
	// +optional
	SSHProxyEnabled bool `json:"sshProxyEnabled,omitempty"`

	// IngressBackend defines the backend service to route traffic to
	// +optional
	IngressBackend *IngressBackendConfig `json:"ingressBackend,omitempty"`

	// DomainRoutes defines domain-based routing rules for HAProxy
	// +optional
	DomainRoutes []DomainRoute `json:"domainRoutes,omitempty"`
}

// IngressBackendConfig defines a simple service:port mapping for traffic routing
type IngressBackendConfig struct {
	// ServiceName is the name of the backend service
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName"`

	// ServiceNamespace is the namespace of the backend service
	// +kubebuilder:validation:Required
	ServiceNamespace string `json:"serviceNamespace"`

	// ServicePort is the port of the backend service
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ServicePort int32 `json:"servicePort"`
}

// DomainRoute defines a domain-based routing rule
type DomainRoute struct {
	// Domain is the domain name to match (e.g., "example.khost.dev")
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// ServiceName is the name of the backend service for this domain
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName"`

	// ServiceNamespace is the namespace of the backend service
	// +kubebuilder:validation:Required
	ServiceNamespace string `json:"serviceNamespace"`

	// ServicePort is the port of the backend service
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ServicePort int32 `json:"servicePort"`
}

// DomainRequestStatus defines the observed state of DomainRequest
type DomainRequestStatus struct {
	// State represents the current state of the DomainRequest
	// +kubebuilder:validation:Enum=Pending;CertificateDownloading;CertificateGenerated;CertificateReady;HAProxyCreating;HAProxyReady;IPRegistering;Ready;Failed
	// +kubebuilder:default=Pending
	State string `json:"state"`

	// Message provides additional information about the current state
	// +optional
	Message string `json:"message,omitempty"`

	// Domain is the registered domain name
	// +optional
	Domain string `json:"domain,omitempty"`

	// Subdomain is the subdomain assigned to this registration
	// +optional
	Subdomain string `json:"subdomain,omitempty"`

	// DNSRecordIDs contains the Cloudflare DNS record IDs created
	// +optional
	DNSRecordIDs []string `json:"dnsRecordIds,omitempty"`

	// OriginCertificateSecretName is the name of the Kubernetes Secret containing the installation's origin certificate
	// This certificate is shared across all DomainRequests for the same installation
	// +optional
	OriginCertificateSecretName string `json:"originCertificateSecretName,omitempty"`

	// LastIPRegistrationTime is when the IP was last registered
	// +optional
	LastIPRegistrationTime *metav1.Time `json:"lastIPRegistrationTime,omitempty"`

	// HAProxyPodName is the name of the HAProxy pod created for this domain
	// +optional
	HAProxyPodName string `json:"haProxyPodName,omitempty"`

	// HAProxyReady indicates if the HAProxy pod is ready and serving traffic
	// +optional
	HAProxyReady bool `json:"haProxyReady,omitempty"`

	// CertificateID is the ID of the generated certificate (deprecated - kept for backward compatibility)
	// +optional
	CertificateID string `json:"certificateId,omitempty"`

	// CertificateSecretName is the name of the Kubernetes Secret containing the certificate (deprecated - kept for backward compatibility)
	// +optional
	CertificateSecretName string `json:"certificateSecretName,omitempty"`

	// LastCertificateGenerationTime is when the certificate was last generated (deprecated - kept for backward compatibility)
	// +optional
	LastCertificateGenerationTime *metav1.Time `json:"lastCertificateGenerationTime,omitempty"`

	// CertificateExpiresAt is when the certificate expires (deprecated - kept for backward compatibility)
	// +optional
	CertificateExpiresAt *metav1.Time `json:"certificateExpiresAt,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// DomainRequestList contains a list of DomainRequest
type DomainRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DomainRequest `json:"items"`
}
