package retry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// RetryConfig defines the retry behavior for cloud operations
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int

	// InitialDelay is the initial delay before first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry
	Multiplier float64

	// Jitter is the random jitter factor to add to delays (0.0 to 1.0)
	Jitter float64

	// Timeout is the overall timeout for the operation
	Timeout time.Duration
}

// DefaultRetryConfig returns a sensible default configuration for cloud operations
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
		Timeout:      5 * time.Minute,
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err     error
	Message string
}

func (e *RetryableError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error, message string) error {
	if message == "" {
		message = err.Error()
	}
	return &RetryableError{
		Err:     err,
		Message: message,
	}
}

// IsRetryable checks if an error is retryable based on common cloud provider error patterns
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Common retryable error patterns across cloud providers
	retryablePatterns := []string{
		// AWS
		"requestlimitexceeded",
		"throttling",
		"throttled",
		"serviceunavailable",
		"internal failure",
		"network error",
		"connection reset",
		"connection refused",
		"connection timed out",
		"timeout exceeded",
		"request limit exceeded",
		"rate limit exceeded",

		// Azure
		"too many requests",
		"service unavailable",
		"operation timed out",
		"request cancelled",
		"429",
		"503",
		"504",

		// GCP
		"resource exhausted",
		"quota exceeded",
		"internal error",
		"deadline exceeded",

		// OCI
		"too many requests",
		"service unavailable",
		"internal server error",

		// General network/transient errors
		"temporary failure",
		"transient error",
		"try again later",
		"could not connect",
		"no such host",
		"dns error",
		"ssl error",
		"tls error",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	// Also check for explicit RetryableError
	var retryable *RetryableError
	return errors.As(err, &retryable)
}

// IsPermanent checks if an error is permanent and should not be retried
func IsPermanent(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Permanent error patterns
	permanentPatterns := []string{
		"not found",
		"does not exist",
		"invalid parameter",
		"invalid value",
		"malformed",
		"unauthorized",
		"access denied",
		"forbidden",
		"authentication failed",
		"permission denied",
		"quota limit",
		"insufficient",
		"already exists",
		"duplicate",
		"conflict",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// Operation represents a cloud operation that can be retried
type Operation func(ctx context.Context) error

// WithRetry executes an operation with retry logic
func WithRetry(ctx context.Context, operation Operation, config RetryConfig, operationName string) error {
	// Create a context with timeout if timeout is set
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Build k8s wait.Backoff from config
	backoff := wait.Backoff{
		Steps:    config.MaxAttempts,
		Duration: config.InitialDelay,
		Factor:   config.Multiplier,
		Jitter:   config.Jitter,
		Cap:      config.MaxDelay,
	}

	var lastErr error
	attempt := 0

	err := wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		attempt++
		if attempt > 1 {
			slog.Debug(fmt.Sprintf("[Cloud Retry] Attempt %d/%d for %s", attempt, config.MaxAttempts, operationName))
		}

		// Execute the operation
		err := operation(ctx)
		if err == nil {
			// Success
			if attempt > 1 {
				slog.Info(fmt.Sprintf("[Cloud Retry] %s succeeded on attempt %d", operationName, attempt))
			}
			return true, nil
		}

		lastErr = err

		// Check if error is permanent - don't retry
		if IsPermanent(err) {
			slog.Warn(fmt.Sprintf("[Cloud Retry] %s failed with permanent error: %v", operationName, err))
			return false, err
		}

		// Check if error is retryable
		if !IsRetryable(err) {
			// Unknown error - log and don't retry
			slog.Warn(fmt.Sprintf("[Cloud Retry] %s failed with non-retryable error: %v", operationName, err))
			return false, err
		}

		// Error is retryable - log and retry
		slog.Debug(fmt.Sprintf("[Cloud Retry] %s failed with retryable error: %v", operationName, err))
		return false, nil
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("%s timed out after %v: %w", operationName, config.Timeout, lastErr)
		}
		if err == wait.ErrWaitTimeout {
			return fmt.Errorf("%s failed after %d attempts: %w", operationName, config.MaxAttempts, lastErr)
		}
		return fmt.Errorf("%s failed: %w", operationName, err)
	}

	return nil
}

// WithRetryResult executes an operation that returns a result with retry logic
func WithRetryResult[T any](ctx context.Context, operation func(ctx context.Context) (T, error), config RetryConfig, operationName string) (T, error) {
	var zero T

	// Create a context with timeout if timeout is set
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Build k8s wait.Backoff from config
	backoff := wait.Backoff{
		Steps:    config.MaxAttempts,
		Duration: config.InitialDelay,
		Factor:   config.Multiplier,
		Jitter:   config.Jitter,
		Cap:      config.MaxDelay,
	}

	var lastErr error
	var result T
	attempt := 0

	err := wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		attempt++
		if attempt > 1 {
			slog.Debug(fmt.Sprintf("[Cloud Retry] Attempt %d/%d for %s", attempt, config.MaxAttempts, operationName))
		}

		// Execute the operation
		var err error
		result, err = operation(ctx)
		if err == nil {
			// Success
			if attempt > 1 {
				slog.Info(fmt.Sprintf("[Cloud Retry] %s succeeded on attempt %d", operationName, attempt))
			}
			return true, nil
		}

		lastErr = err

		// Check if error is permanent - don't retry
		if IsPermanent(err) {
			slog.Warn(fmt.Sprintf("[Cloud Retry] %s failed with permanent error: %v", operationName, err))
			return false, err
		}

		// Check if error is retryable
		if !IsRetryable(err) {
			// Unknown error - log and don't retry
			slog.Warn(fmt.Sprintf("[Cloud Retry] %s failed with non-retryable error: %v", operationName, err))
			return false, err
		}

		// Error is retryable - log and retry
		slog.Debug(fmt.Sprintf("[Cloud Retry] %s failed with retryable error: %v", operationName, err))
		return false, nil
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return zero, fmt.Errorf("%s timed out after %v: %w", operationName, config.Timeout, lastErr)
		}
		if err == wait.ErrWaitTimeout {
			return zero, fmt.Errorf("%s failed after %d attempts: %w", operationName, config.MaxAttempts, lastErr)
		}
		return zero, fmt.Errorf("%s failed: %w", operationName, err)
	}

	return result, nil
}
