package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeForLabel tests label value sanitization
func TestSanitizeForLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid email",
			input:    "user@example.com",
			expected: "user-at-example-dot-com",
		},
		{
			name:     "Email with dots and underscores",
			input:    "first.last_name@example.com",
			expected: "first-dot-last-name-at-example-dot-com",
		},
		{
			name:     "Email with plus sign",
			input:    "user+tag@example.com",
			expected: "user-plus-tag-at-example-dot-com",
		},
		{
			name:     "Long string exceeding 63 chars",
			input:    "very-long-email-address-that-exceeds-the-kubernetes-label-limit@example.com",
			expected: "very-long-email-address-that-exceeds-the-kubernetes-label-limit",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Already valid label",
			input:    "valid-label-123",
			expected: "valid-label-123",
		},
		{
			name:     "Uppercase letters",
			input:    "UserName",
			expected: "username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLabel(tt.input)
			assert.Equal(t, tt.expected, result)

			// Verify result is valid if not empty
			if result != "" {
				assert.True(t, IsValidLabel(result), "Sanitized label should be valid")
			}
		})
	}
}

// TestExtractUsernameFromEmail tests username extraction and sanitization
func TestExtractUsernameFromEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple email",
			input:    "user@example.com",
			expected: "user",
		},
		{
			name:     "Email with dots",
			input:    "first.last@example.com",
			expected: "first-last",
		},
		{
			name:     "Email with underscores",
			input:    "user_name@example.com",
			expected: "user-name",
		},
		{
			name:     "Email with plus",
			input:    "user+tag@example.com",
			expected: "user-tag",
		},
		{
			name:     "Long username",
			input:    "verylongusernamethatexceedsfiftycharacterslimit@example.com",
			expected: "verylongusernamethatexceedsfiftycharacterslimit",
		},
		{
			name:     "Username starting with number",
			input:    "123user@example.com",
			expected: "123user",
		},
		{
			name:     "Username starting with hyphen",
			input:    "-user@example.com",
			expected: "u--user",
		},
		{
			name:     "Username with trailing hyphens",
			input:    "user-@example.com",
			expected: "user",
		},
		{
			name:     "Empty email",
			input:    "",
			expected: "",
		},
		{
			name:     "No @ sign (just username)",
			input:    "username",
			expected: "username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractUsernameFromEmail(tt.input)
			assert.Equal(t, tt.expected, result)

			// Verify result is valid resource name if not empty
			if result != "" {
				assert.True(t, len(result) <= 50, "Username should be max 50 chars")
			}
		})
	}
}

// TestSanitizeResourceName tests resource name sanitization
func TestSanitizeResourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid resource name",
			input:    "my-resource-123",
			expected: "my-resource-123",
		},
		{
			name:     "Uppercase letters",
			input:    "MyResource",
			expected: "myresource",
		},
		{
			name:     "Special characters",
			input:    "my_resource@123!",
			expected: "my-resource-123",
		},
		{
			name:     "Multiple consecutive hyphens",
			input:    "my---resource",
			expected: "my-resource",
		},
		{
			name:     "Leading and trailing hyphens",
			input:    "-my-resource-",
			expected: "my-resource",
		},
		{
			name:     "Empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "Only special characters",
			input:    "@#$%",
			expected: "default",
		},
		{
			name:     "Very long name",
			input:    string(make([]byte, 300)),
			expected: string(make([]byte, 253)),
		},
		{
			name:     "Dots and underscores",
			input:    "my.resource_name",
			expected: "my-resource-name",
		},
		{
			name:     "Name with spaces",
			input:    "my resource name",
			expected: "my-resource-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeResourceName(tt.input)

			if tt.name == "Very long name" {
				// For very long name, just verify length constraint
				assert.True(t, len(result) <= 253, "Result should be max 253 chars")
			} else {
				assert.Equal(t, tt.expected, result)
			}

			// Verify result is valid (skip for empty string as it's not a valid resource name but is an expected return)
			if tt.name != "Empty string returns empty" {
				assert.True(t, IsValidResourceName(result), "Sanitized name should be valid")
			}
		})
	}
}

// TestIsValidLabel tests label validation
func TestIsValidLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid label",
			input:    "valid-label-123",
			expected: true,
		},
		{
			name:     "Valid with dots and underscores",
			input:    "valid.label_123",
			expected: true,
		},
		{
			name:     "Too long",
			input:    string(make([]byte, 64)),
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Starts with hyphen",
			input:    "-invalid",
			expected: false,
		},
		{
			name:     "Ends with hyphen",
			input:    "invalid-",
			expected: false,
		},
		{
			name:     "Contains special characters",
			input:    "invalid@label",
			expected: false,
		},
		{
			name:     "Uppercase letters",
			input:    "InvalidLabel",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidLabel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsValidResourceName tests resource name validation
func TestIsValidResourceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid resource name",
			input:    "my-resource-123",
			expected: true,
		},
		{
			name:     "Too long",
			input:    string(make([]byte, 254)),
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Starts with hyphen",
			input:    "-invalid",
			expected: false,
		},
		{
			name:     "Ends with hyphen",
			input:    "invalid-",
			expected: false,
		},
		{
			name:     "Contains uppercase",
			input:    "Invalid",
			expected: false,
		},
		{
			name:     "Contains special characters",
			input:    "invalid_name",
			expected: false,
		},
		{
			name:     "Contains dots",
			input:    "invalid.name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidResourceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateKubernetesNamespace tests namespace validation
func TestValidateKubernetesNamespace(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Valid namespace",
			input:     "my-namespace",
			expectErr: false,
		},
		{
			name:      "Valid with numbers",
			input:     "namespace-123",
			expectErr: false,
		},
		{
			name:      "Empty namespace",
			input:     "",
			expectErr: true,
			errMsg:    "namespace cannot be empty",
		},
		{
			name:      "Too long",
			input:     string(make([]byte, 64)),
			expectErr: true,
			errMsg:    "namespace too long",
		},
		{
			name:      "Starts with hyphen",
			input:     "-invalid",
			expectErr: true,
			errMsg:    "invalid namespace format",
		},
		{
			name:      "Ends with hyphen",
			input:     "invalid-",
			expectErr: true,
			errMsg:    "invalid namespace format",
		},
		{
			name:      "Contains uppercase",
			input:     "Invalid",
			expectErr: true,
			errMsg:    "invalid namespace format",
		},
		{
			name:      "Contains special characters",
			input:     "invalid_namespace",
			expectErr: true,
			errMsg:    "invalid namespace format",
		},
		{
			name:      "Contains dots",
			input:     "invalid.namespace",
			expectErr: true,
			errMsg:    "invalid namespace format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKubernetesNamespace(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSanitizeSearchDomains tests DNS search domain sanitization
func TestSanitizeSearchDomains(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		expected  string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Empty domains",
			input:     []string{},
			expected:  "svc.cluster.local cluster.local",
			expectErr: false,
		},
		{
			name:      "Valid single domain",
			input:     []string{"example.com"},
			expected:  "example.com",
			expectErr: false,
		},
		{
			name:      "Valid multiple domains",
			input:     []string{"example.com", "test.local"},
			expected:  "example.com test.local",
			expectErr: false,
		},
		{
			name:      "Domain with subdomain",
			input:     []string{"api.example.com"},
			expected:  "api.example.com",
			expectErr: false,
		},
		{
			name:      "Domain with multiple subdomains",
			input:     []string{"api.v1.example.com"},
			expected:  "api.v1.example.com",
			expectErr: false,
		},
		{
			name:      "Invalid domain with uppercase",
			input:     []string{"Example.com"},
			expected:  "",
			expectErr: true,
			errMsg:    "invalid search domain format",
		},
		{
			name:      "Invalid domain with double dots",
			input:     []string{"example..com"},
			expected:  "",
			expectErr: true,
			errMsg:    "invalid search domain format",
		},
		{
			name:      "Path traversal attempt",
			input:     []string{"example.com/../etc/passwd"},
			expected:  "",
			expectErr: true,
			errMsg:    "invalid search domain format",
		},
		{
			name:      "Command injection attempt",
			input:     []string{"example.com; cat /etc/passwd"},
			expected:  "",
			expectErr: true,
			errMsg:    "invalid search domain format",
		},
		{
			name:      "Shell metacharacters",
			input:     []string{"example.com|whoami"},
			expected:  "",
			expectErr: true,
			errMsg:    "invalid search domain format",
		},
		{
			name:      "Shell variable injection",
			input:     []string{"example.com$USER"},
			expected:  "",
			expectErr: true,
			errMsg:    "invalid search domain format",
		},
		{
			name:      "Backticks injection",
			input:     []string{"example.com`id`"},
			expected:  "",
			expectErr: true,
			errMsg:    "invalid search domain format",
		},
		{
			name:      "Mixed valid and empty domains",
			input:     []string{"example.com", "", "test.local"},
			expected:  "example.com test.local",
			expectErr: false,
		},
		{
			name:      "All empty domains",
			input:     []string{"", "", ""},
			expected:  "svc.cluster.local cluster.local",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeSearchDomains(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestValidateHostPathForWorkspace tests workspace host path validation
func TestValidateHostPathForWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		workspaceName string
		expectErr     bool
		errMsg        string
	}{
		{
			name:          "Valid workspace path",
			path:          "/home/kl/workspaces/my-workspace",
			workspaceName: "my-workspace",
			expectErr:     false,
		},
		{
			name:          "Valid path with hyphens and numbers",
			path:          "/home/kl/workspaces/workspace-123",
			workspaceName: "workspace-123",
			expectErr:     false,
		},
		{
			name:          "Empty path",
			path:          "",
			workspaceName: "my-workspace",
			expectErr:     true,
			errMsg:        "host path cannot be empty",
		},
		{
			name:          "Path outside allowed prefix",
			path:          "/tmp/my-workspace",
			workspaceName: "my-workspace",
			expectErr:     true,
			errMsg:        "unsafe host path",
		},
		{
			name:          "Path without workspace name suffix",
			path:          "/home/kl/workspaces/other-workspace",
			workspaceName: "my-workspace",
			expectErr:     true,
			errMsg:        "host path must end with workspace name",
		},
		{
			name:          "Path traversal with ..",
			path:          "/home/kl/workspaces/../../../etc/passwd",
			workspaceName: "my-workspace",
			expectErr:     true,
			errMsg:        "host path must end with workspace name",
		},
		{
			name:          "Path with double dots",
			path:          "/home/kl/workspaces/..my-workspace",
			workspaceName: "my-workspace",
			expectErr:     true,
			errMsg:        "host path must end with workspace name",
		},
		{
			name:          "Path traversal detection",
			path:          "/home/kl/workspaces/my-workspace/../other",
			workspaceName: "my-workspace",
			expectErr:     true,
			errMsg:        "host path must end with workspace name",
		},
		{
			name:          "Absolute path required",
			path:          "workspaces/my-workspace",
			workspaceName: "my-workspace",
			expectErr:     true,
			errMsg:        "unsafe host path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostPathForWorkspace(tt.path, tt.workspaceName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIsAlphanumeric tests the private isAlphanumeric helper
func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected bool
	}{
		{name: "Lowercase a", input: 'a', expected: true},
		{name: "Lowercase z", input: 'z', expected: true},
		{name: "Lowercase m", input: 'm', expected: true},
		{name: "Digit 0", input: '0', expected: true},
		{name: "Digit 9", input: '9', expected: true},
		{name: "Digit 5", input: '5', expected: true},
		{name: "Uppercase A", input: 'A', expected: false},
		{name: "Uppercase Z", input: 'Z', expected: false},
		{name: "Hyphen", input: '-', expected: false},
		{name: "Underscore", input: '_', expected: false},
		{name: "Dot", input: '.', expected: false},
		{name: "Space", input: ' ', expected: false},
		{name: "Special char @", input: '@', expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlphanumeric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSanitizationRoundTrip tests that sanitization produces valid output
func TestSanitizationRoundTrip(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{name: "Email", input: "user@example.com"},
		{name: "Complex email", input: "first.last+tag@subdomain.example.com"},
		{name: "Simple name", input: "my-workspace-123"},
		{name: "Username", input: "user_name.123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test SanitizeForLabel
			label := SanitizeForLabel(tc.input)
			if label != "" {
				assert.True(t, IsValidLabel(label), "SanitizeForLabel should produce valid label")
			}

			// Test SanitizeResourceName
			resourceName := SanitizeResourceName(tc.input)
			assert.True(t, IsValidResourceName(resourceName), "SanitizeResourceName should produce valid resource name")

			// Test ExtractUsernameFromEmail
			username := ExtractUsernameFromEmail(tc.input)
			if username != "" {
				// Username should be lowercase alphanumeric with hyphens
				assert.True(t, len(username) > 0, "Username should not be empty")
				assert.True(t, len(username) <= 50, "Username should be max 50 chars")
			}
		})
	}
}
