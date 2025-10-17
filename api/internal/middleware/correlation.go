package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Correlation ID context key
type correlationKey string

const (
	CorrelationIDKey correlationKey = "correlation_id"
	CorrelationHeader = "X-Correlation-ID"
	RequestIDHeader   = "X-Request-ID"
)

// CorrelationMiddleware adds correlation IDs to requests for tracing
func CorrelationMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get correlation ID from headers
		correlationID := c.GetHeader(CorrelationHeader)
		if correlationID == "" {
			// Fall back to X-Request-ID header
			correlationID = c.GetHeader(RequestIDHeader)
		}

		// Generate a new correlation ID if none exists
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}

		// Add correlation ID to context
		c.Set(string(CorrelationIDKey), correlationID)

		// Add correlation ID to response headers
		c.Header(CorrelationHeader, correlationID)

		// Add correlation ID to logger
		logger = logger.With(zap.String("correlation_id", correlationID))
		c.Set("logger", logger)

		// Add correlation ID to request context for use in services
		ctx := context.WithValue(c.Request.Context(), CorrelationIDKey, correlationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetCorrelationID extracts correlation ID from gin context
func GetCorrelationID(c *gin.Context) string {
	if id, exists := c.Get(string(CorrelationIDKey)); exists {
		if correlationID, ok := id.(string); ok {
			return correlationID
		}
	}
	return ""
}

// GetCorrelationIDFromContext extracts correlation ID from context
func GetCorrelationIDFromContext(ctx context.Context) string {
	if id := ctx.Value(CorrelationIDKey); id != nil {
		if correlationID, ok := id.(string); ok {
			return correlationID
		}
	}
	return ""
}

// generateCorrelationID generates a new correlation ID
func generateCorrelationID() string {
	// Generate UUID and remove dashes for shorter ID
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error       string `json:"error"`
	Message     string `json:"message,omitempty"`
	Code        string `json:"code,omitempty"`
	RequestID   string `json:"request_id"`
	Timestamp   string `json:"timestamp"`
	TraceID     string `json:"trace_id,omitempty"`
}

// NewErrorResponse creates a new error response with correlation ID
func NewErrorResponse(c *gin.Context, message string, details ...string) ErrorResponse {
	return ErrorResponse{
		Error:     message,
		Message:   joinStrings(details...),
		RequestID: GetCorrelationID(c),
		Timestamp: getCurrentTimestamp(),
	}
}

// ErrorWithCode creates an error response with error code
func ErrorWithCode(c *gin.Context, code, message string, details ...string) ErrorResponse {
	return ErrorResponse{
		Error:     message,
		Code:      code,
		Message:   joinStrings(details...),
		RequestID: GetCorrelationID(c),
		Timestamp: getCurrentTimestamp(),
	}
}

// SendErrorResponse sends a standardized error response
func SendErrorResponse(c *gin.Context, statusCode int, message string, details ...string) {
	response := NewErrorResponse(c, message, details...)
	c.JSON(statusCode, response)
}

// SendErrorResponseWithCode sends an error response with error code
func SendErrorResponseWithCode(c *gin.Context, statusCode int, code, message string, details ...string) {
	response := ErrorWithCode(c, code, message, details...)
	c.JSON(statusCode, response)
}

// joinStrings helper to join optional detail strings
func joinStrings(details ...string) string {
	var nonEmpty []string
	for _, detail := range details {
		if detail != "" {
			nonEmpty = append(nonEmpty, detail)
		}
	}
	return strings.Join(nonEmpty, "; ")
}

// getCurrentTimestamp returns current timestamp in ISO format
func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}