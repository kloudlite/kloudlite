// +k8s:deepcopy-gen=package
// +groupName=environments.kloudlite.io

// Package v1 contains API Schema definitions for the environments v1 API group
//
// This package defines types in the environments.kloudlite.io API group:
// - Environment: Represents workspace environments with namespace management
// - Composition: Represents Docker Compose applications deployed in environments
//
// While these types are part of the same API group, they are managed by separate controllers:
// - Environment controller: manages namespaces and environment lifecycle
// - Composition controller: manages Docker Compose deployments
//
// Controllers import this package using descriptive aliases for clarity:
//
//	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1" // for Environment types
//	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1" // for Composition types
package v1
