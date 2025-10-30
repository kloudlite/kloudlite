// +k8s:deepcopy-gen=package
// +groupName=domains.kloudlite.io

// Package v1 contains API Schema definitions for the domains v1 API group
//
// This package defines types in the domains.kloudlite.io API group:
// - DomainRequest: Represents domain registration and certificate management for installations
//
// The DomainRequest controller manages:
// - IP address registration with console.kloudlite.io
// - DNS record creation via Cloudflare
// - TLS certificate generation using Cloudflare Origin CA
// - Certificate storage and renewal
package v1
