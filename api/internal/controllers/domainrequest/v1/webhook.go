package v1

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SetupWebhookWithManager sets up the webhook with the manager
func (r *DomainRequest) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/webhooks/validate/domainrequests,mutating=false,failurePolicy=fail,sideEffects=None,groups=domains.kloudlite.io,resources=domainrequests,verbs=create;update,versions=v1,name=domainrequests.kloudlite.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &DomainRequest{}

// ValidateCreate implements webhook.CustomValidator
func (r *DomainRequest) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return r.validateDomainRequest()
}

// ValidateUpdate implements webhook.CustomValidator
func (r *DomainRequest) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	// Use newObj which is the updated DomainRequest
	req, ok := newObj.(*DomainRequest)
	if !ok {
		return nil, fmt.Errorf("expected DomainRequest but got %T", newObj)
	}
	return req.validateDomainRequest()
}

// ValidateDelete implements webhook.CustomValidator
func (r *DomainRequest) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed for deletion
	return nil, nil
}

// validateDomainRequest validates the DomainRequest resource
func (r *DomainRequest) validateDomainRequest() (admission.Warnings, error) {
	var warnings admission.Warnings

	// Validate that all domain routes are covered by origin certificate hostnames
	if len(r.Spec.DomainRoutes) > 0 && len(r.Spec.OriginCertificateHostnames) > 0 {
		uncoveredDomains := []string{}

		for _, route := range r.Spec.DomainRoutes {
			if !isDomainCovered(route.Domain, r.Spec.OriginCertificateHostnames) {
				uncoveredDomains = append(uncoveredDomains, route.Domain)
			}
		}

		if len(uncoveredDomains) > 0 {
			return nil, fmt.Errorf(
				"domain routes contain domains not covered by origin certificate hostnames: %s. "+
					"All domains in domainRoutes must be present in originCertificateHostnames",
				strings.Join(uncoveredDomains, ", "),
			)
		}
	}

	// Validate that origin certificate hostnames don't contain multiple wildcards
	for _, hostname := range r.Spec.OriginCertificateHostnames {
		if strings.Count(hostname, "*") > 1 {
			return nil, fmt.Errorf(
				"invalid origin certificate hostname '%s': Cloudflare only allows ONE wildcard per hostname at the beginning (e.g., '*.example.com'). "+
					"Multiple wildcards like '*.*.example.com' are not supported",
				hostname,
			)
		}

		// Check for wildcards not at the beginning
		if strings.Contains(hostname, "*") && !strings.HasPrefix(hostname, "*.") {
			return nil, fmt.Errorf(
				"invalid origin certificate hostname '%s': wildcard must be at the beginning (e.g., '*.example.com'). "+
					"Patterns like 'test.*.example.com' are not supported by Cloudflare",
				hostname,
			)
		}
	}

	return warnings, nil
}

// isDomainCovered checks if a domain is covered by one of the certificate hostnames
// Supports exact match and wildcard matching (e.g., *.example.com covers sub.example.com)
func isDomainCovered(domain string, certHostnames []string) bool {
	for _, certHostname := range certHostnames {
		// Exact match
		if domain == certHostname {
			return true
		}

		// Wildcard match: *.example.com covers sub.example.com but not example.com
		if strings.HasPrefix(certHostname, "*.") {
			wildcardBase := certHostname[2:] // Remove "*."
			// Check if domain ends with .wildcardBase (e.g., sub.example.com ends with .example.com)
			if strings.HasSuffix(domain, "."+wildcardBase) {
				return true
			}
		}
	}

	return false
}
