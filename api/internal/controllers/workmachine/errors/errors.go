package errors

import (
	"fmt"
)

// ProviderError is the base interface for all cloud provider errors
type ProviderError interface {
	error
	IsRetryable() bool
}

// BaseProviderError implements the ProviderError interface
type BaseProviderError struct {
	Message   string
	Retryable bool
	Cause     error
}

func (e *BaseProviderError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *BaseProviderError) IsRetryable() bool {
	return e.Retryable
}

func (e *BaseProviderError) Unwrap() error {
	return e.Cause
}

// ProviderNotConfiguredError indicates that no cloud provider is configured
type ProviderNotConfiguredError struct {
	BaseProviderError
}

func NewProviderNotConfiguredError() *ProviderNotConfiguredError {
	return &ProviderNotConfiguredError{
		BaseProviderError: BaseProviderError{
			Message:   "cloud provider not configured for this WorkMachine",
			Retryable: false,
		},
	}
}

// ProviderNotImplementedError indicates that a cloud provider is not yet implemented
type ProviderNotImplementedError struct {
	BaseProviderError
	Provider string
}

func NewProviderNotImplementedError(provider string) *ProviderNotImplementedError {
	return &ProviderNotImplementedError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("cloud provider %q is not yet implemented", provider),
			Retryable: false,
		},
		Provider: provider,
	}
}

// PermissionDeniedError indicates that required permissions are missing
type PermissionDeniedError struct {
	BaseProviderError
	Action          string
	MissingActions  []string
	RequiredActions []string
}

func NewPermissionDeniedError(action string, missingActions []string, requiredActions []string, cause error) *PermissionDeniedError {
	return &PermissionDeniedError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("permission denied for action %q", action),
			Retryable: false,
			Cause:     cause,
		},
		Action:          action,
		MissingActions:  missingActions,
		RequiredActions: requiredActions,
	}
}

// ResourceNotFoundError indicates that a cloud resource was not found
type ResourceNotFoundError struct {
	BaseProviderError
	ResourceType string
	ResourceID   string
}

func NewResourceNotFoundError(resourceType, resourceID string) *ResourceNotFoundError {
	return &ResourceNotFoundError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("%s %q not found", resourceType, resourceID),
			Retryable: false,
		},
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// ResourceAlreadyExistsError indicates that a cloud resource already exists
type ResourceAlreadyExistsError struct {
	BaseProviderError
	ResourceType string
	ResourceID   string
}

func NewResourceAlreadyExistsError(resourceType, resourceID string) *ResourceAlreadyExistsError {
	return &ResourceAlreadyExistsError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("%s %q already exists", resourceType, resourceID),
			Retryable: false,
		},
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
}

// InvalidConfigurationError indicates that the provider configuration is invalid
type InvalidConfigurationError struct {
	BaseProviderError
	Field  string
	Reason string
}

func NewInvalidConfigurationError(field, reason string) *InvalidConfigurationError {
	return &InvalidConfigurationError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("invalid configuration for field %q: %s", field, reason),
			Retryable: false,
		},
		Field:  field,
		Reason: reason,
	}
}

// ProviderAPIError indicates a generic error from the cloud provider API
type ProviderAPIError struct {
	BaseProviderError
	Operation string
	Code      string
}

func NewProviderAPIError(operation string, code string, message string, retryable bool, cause error) *ProviderAPIError {
	return &ProviderAPIError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("cloud provider API error during %s: %s", operation, message),
			Retryable: retryable,
			Cause:     cause,
		},
		Operation: operation,
		Code:      code,
	}
}

// QuotaExceededError indicates that a cloud provider quota has been exceeded
type QuotaExceededError struct {
	BaseProviderError
	ResourceType string
	QuotaName    string
}

func NewQuotaExceededError(resourceType, quotaName string, cause error) *QuotaExceededError {
	return &QuotaExceededError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("quota exceeded for %s: %s", resourceType, quotaName),
			Retryable: true, // User might be able to delete other resources
			Cause:     cause,
		},
		ResourceType: resourceType,
		QuotaName:    quotaName,
	}
}

// InvalidInstanceStateError indicates that an operation cannot be performed in the current instance state
type InvalidInstanceStateError struct {
	BaseProviderError
	InstanceID    string
	CurrentState  string
	RequiredState string
	Operation     string
}

func NewInvalidInstanceStateError(instanceID string, currentState, requiredState string, operation string) *InvalidInstanceStateError {
	return &InvalidInstanceStateError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("cannot %s instance %q: current state is %q, required state is %q", operation, instanceID, currentState, requiredState),
			Retryable: true, // State might change
		},
		InstanceID:    instanceID,
		CurrentState:  currentState,
		RequiredState: requiredState,
		Operation:     operation,
	}
}

// DNSError indicates an error with DNS operations
type DNSError struct {
	BaseProviderError
	Domain    string
	Operation string
}

func NewDNSError(domain, operation string, cause error) *DNSError {
	return &DNSError{
		BaseProviderError: BaseProviderError{
			Message:   fmt.Sprintf("DNS error for domain %q during %s", domain, operation),
			Retryable: true,
			Cause:     cause,
		},
		Domain:    domain,
		Operation: operation,
	}
}
