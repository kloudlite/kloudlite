package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestRequest represents a test HTTP request
type TestRequest struct {
	Method  string
	URL     string
	Body    interface{}
	Headers map[string]string
}

// TestResponse represents a test HTTP response
type TestResponse struct {
	Code int
	Body []byte
}

// SetupTestRouter creates a test router
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	return router
}

// PerformRequest performs a test HTTP request
func PerformRequest(router *gin.Engine, req TestRequest) TestResponse {
	var body io.Reader
	if req.Body != nil {
		jsonBytes, _ := json.Marshal(req.Body)
		body = bytes.NewBuffer(jsonBytes)
	}

	httpReq, _ := http.NewRequest(req.Method, req.URL, body)

	// Set headers
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	return TestResponse{
		Code: w.Code,
		Body: w.Body.Bytes(),
	}
}

// ParseJSONResponse parses JSON response body
func ParseJSONResponse(t *testing.T, body []byte, target interface{}) {
	err := json.Unmarshal(body, target)
	require.NoError(t, err, "Failed to parse JSON response")
}

// AssertJSONResponse asserts JSON response matches expected
func AssertJSONResponse(t *testing.T, expected, actual interface{}) {
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)

	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)

	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}

// CreateTestConfig creates a test configuration
func CreateTestConfig() map[string]string {
	return map[string]string{
		"PORT":        "8080",
		"ENVIRONMENT": "test",
		"LOG_LEVEL":   "debug",
	}
}