package workmachine

// This file re-exports error types from the errors package for backward compatibility
// All error types have been moved to the errors subpackage to avoid circular imports

// import (
// 	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/errors"
// )
//
// // Re-export error interface and base types
// type (
// 	ProviderError                  = errors.ProviderError
// 	BaseProviderError              = errors.BaseProviderError
// 	ProviderNotConfiguredError     = errors.ProviderNotConfiguredError
// 	ProviderNotImplementedError    = errors.ProviderNotImplementedError
// 	PermissionDeniedError          = errors.PermissionDeniedError
// 	ResourceNotFoundError          = errors.ResourceNotFoundError
// 	ResourceAlreadyExistsError     = errors.ResourceAlreadyExistsError
// 	InvalidConfigurationError      = errors.InvalidConfigurationError
// 	ProviderAPIError               = errors.ProviderAPIError
// 	QuotaExceededError             = errors.QuotaExceededError
// 	InvalidInstanceStateError      = errors.InvalidInstanceStateError
// 	DNSError                       = errors.DNSError
// )
//
// // Re-export error constructor functions
// var (
// 	NewProviderNotConfiguredError  = errors.NewProviderNotConfiguredError
// 	NewProviderNotImplementedError = errors.NewProviderNotImplementedError
// 	NewPermissionDeniedError       = errors.NewPermissionDeniedError
// 	NewResourceNotFoundError       = errors.NewResourceNotFoundError
// 	NewResourceAlreadyExistsError  = errors.NewResourceAlreadyExistsError
// 	NewInvalidConfigurationError   = errors.NewInvalidConfigurationError
// 	NewProviderAPIError            = errors.NewProviderAPIError
// 	NewQuotaExceededError          = errors.NewQuotaExceededError
// 	NewInvalidInstanceStateError   = errors.NewInvalidInstanceStateError
// 	NewDNSError                    = errors.NewDNSError
// )
