package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// SanitizeForLabel sanitizes a string to be used as a Kubernetes label value.
// It follows Kubernetes label value constraints:
// - Must be 63 characters or less
// - Must start and end with alphanumeric character
// - May contain dashes, underscores, and dots in between
func SanitizeForLabel(value string) string {
	if value == "" {
		return ""
	}

	// Replace special characters with hyphens for label value
	sanitized := strings.ReplaceAll(value, "@", "-at-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-dot-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, "+", "-plus-")
	sanitized = strings.ToLower(sanitized)

	// Ensure it starts and ends with alphanumeric
	sanitized = strings.Trim(sanitized, "-")

	// Limit length to 63 characters (Kubernetes label value limit)
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
	}

	// Ensure it ends with alphanumeric by trimming trailing hyphens
	sanitized = strings.TrimRight(sanitized, "-")

	return sanitized
}

// ExtractUsernameFromEmail extracts the username part from an email address
// and sanitizes it for use in Kubernetes resource names.
func ExtractUsernameFromEmail(email string) string {
	if email == "" {
		return ""
	}

	username := email
	if idx := strings.Index(username, "@"); idx > 0 {
		username = username[:idx]
	}

	// Replace dots and special characters with hyphens for valid k8s names
	username = strings.ReplaceAll(username, ".", "-")
	username = strings.ReplaceAll(username, "_", "-")
	username = strings.ReplaceAll(username, "+", "-")
	username = strings.ToLower(username)

	// Ensure the username starts with a letter or number
	if len(username) > 0 && !isAlphanumeric(username[0]) {
		username = "u-" + username
	}

	// Limit length to ensure the full name stays within k8s limits
	if len(username) > 50 {
		username = username[:50]
	}

	// Trim trailing hyphens
	username = strings.TrimRight(username, "-")

	return username
}

// SanitizeResourceName sanitizes a string for use as a Kubernetes resource name.
// Follows RFC 1123 subdomain naming rules:
// - No more than 253 characters
// - Lowercase alphanumeric characters and '-'
// - Start and end with alphanumeric
func SanitizeResourceName(name string) string {
	if name == "" {
		return ""
	}

	// Convert to lowercase and replace invalid characters
	sanitized := strings.ToLower(name)

	// Replace invalid characters with hyphens
	var result strings.Builder
	for _, r := range sanitized {
		if unicode.IsLower(r) || unicode.IsDigit(r) || r == '-' {
			result.WriteRune(r)
		} else {
			result.WriteRune('-')
		}
	}

	sanitized = result.String()

	// Remove consecutive hyphens
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}

	// Ensure it starts and ends with alphanumeric
	sanitized = strings.Trim(sanitized, "-")
	if sanitized == "" {
		return "default"
	}

	// Limit length to 253 characters
	if len(sanitized) > 253 {
		sanitized = sanitized[:253]
		// Ensure it still ends with alphanumeric after truncation
		sanitized = strings.TrimRight(sanitized, "-")
		if sanitized == "" {
			return "default"
		}
	}

	return sanitized
}

// isAlphanumeric checks if a byte is alphanumeric (a-z, 0-9)
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

// IsValidLabel checks if a string is a valid Kubernetes label value
func IsValidLabel(value string) bool {
	if len(value) > 63 || len(value) == 0 {
		return false
	}

	// Must start and end with alphanumeric
	if !isAlphanumeric(value[0]) || !isAlphanumeric(value[len(value)-1]) {
		return false
	}

	// Can contain alphanumeric, dashes, underscores, and dots in between
	for _, r := range value {
		if isAlphanumeric(byte(r)) || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}

	return true
}

// IsValidResourceName checks if a string is a valid Kubernetes resource name
func IsValidResourceName(name string) bool {
	if len(name) > 253 || len(name) == 0 {
		return false
	}

	// Must start and end with alphanumeric
	if !isAlphanumeric(name[0]) || !isAlphanumeric(name[len(name)-1]) {
		return false
	}

	// Can contain lowercase alphanumeric and hyphens
	for _, r := range name {
		if unicode.IsLower(r) || unicode.IsDigit(r) || r == '-' {
			continue
		}
		return false
	}

	return true
}

// ValidateKubernetesNamespace validates that a string is a safe Kubernetes namespace name
// Follows RFC 1123 subdomain naming rules for namespaces
func ValidateKubernetesNamespace(ns string) error {
	if ns == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	if len(ns) > 63 {
		return fmt.Errorf("namespace too long (max 63 characters): %s", ns)
	}

	// Must match regex: ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$
	namespaceRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !namespaceRegex.MatchString(ns) {
		return fmt.Errorf("invalid namespace format: %s (must be lowercase alphanumeric with hyphens)", ns)
	}

	return nil
}

// SanitizeSearchDomains sanitizes search domains for DNS configuration
// Prevents DNS injection attacks by validating each domain component
func SanitizeSearchDomains(domains []string) (string, error) {
	if len(domains) == 0 {
		return "svc.cluster.local cluster.local", nil
	}

	var validDomains []string
	domainRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$`)

	for _, domain := range domains {
		if domain == "" {
			continue
		}

		// Validate each domain component
		if !domainRegex.MatchString(domain) {
			return "", fmt.Errorf("invalid search domain format: %s", domain)
		}

		// Additional safety checks
		if strings.Contains(domain, "..") {
			return "", fmt.Errorf("potentially malicious domain (contains '..'): %s", domain)
		}

		// Prevent path traversal attempts
		if strings.ContainsAny(domain, "/\\;|&`'\"$(){}[]<>") {
			return "", fmt.Errorf("potentially malicious characters in domain: %s", domain)
		}

		validDomains = append(validDomains, domain)
	}

	// Default fallback if no valid domains
	if len(validDomains) == 0 {
		return "svc.cluster.local cluster.local", nil
	}

	return strings.Join(validDomains, " "), nil
}

// ValidateHostPathForWorkspace validates host paths for workspace directories
// Prevents path traversal attacks beyond allowed directories
func ValidateHostPathForWorkspace(path, workspaceName string) error {
	if path == "" {
		return fmt.Errorf("host path cannot be empty")
	}

	// Only allow paths within /home/kl/workspaces/
	allowedPrefix := "/home/kl/workspaces/"
	if !strings.HasPrefix(path, allowedPrefix) {
		return fmt.Errorf("unsafe host path: %s (must be within %s)", path, allowedPrefix)
	}

	// Extract workspace name from path for validation
	expectedSuffix := "/" + workspaceName
	if !strings.HasSuffix(path, expectedSuffix) {
		return fmt.Errorf("host path must end with workspace name: expected suffix %s, got %s", expectedSuffix, path)
	}

	// Clean the path and verify it's still within allowed bounds
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, allowedPrefix) {
		return fmt.Errorf("path traversal detected: %s resolves to %s", path, cleanPath)
	}

	// Additional safety checks
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	return nil
}
