package repository

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var (
	// K8s name validation regex: DNS-1123 subdomain
	k8sNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// IsValidK8sName checks if a string is a valid Kubernetes resource name
// Must be 3-63 characters, lowercase alphanumeric with hyphens,
// starting and ending with an alphanumeric character
func IsValidK8sName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	return k8sNameRegex.MatchString(name)
}

// GenerateUsernameFromEmail generates a valid k8s username from an email
// Example: john.doe@example.com -> john-doe
func GenerateUsernameFromEmail(email string) string {
	// Extract part before @
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return ""
	}

	username := parts[0]

	// Replace dots and underscores with hyphens
	username = strings.ReplaceAll(username, ".", "-")
	username = strings.ReplaceAll(username, "_", "-")

	// Convert to lowercase
	username = strings.ToLower(username)

	// Remove any invalid characters
	username = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(username, "")

	// Ensure it starts and ends with alphanumeric
	username = strings.Trim(username, "-")

	// Ensure minimum length
	if len(username) < 3 {
		username = username + "-user"
	}

	// Ensure maximum length
	if len(username) > 63 {
		username = username[:63]
	}

	// Trim trailing hyphens again in case we truncated
	username = strings.TrimRight(username, "-")

	return username
}

// GenerateUsernameWithSuffix generates a username with a random suffix
// Example: john-doe -> john-doe-x3p9
func GenerateUsernameWithSuffix(baseUsername string) string {
	// Generate 4-character random suffix
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 4)
	for i := range suffix {
		suffix[i] = charset[randomSource.Intn(len(charset))]
	}

	// Combine base with suffix
	suggested := baseUsername + "-" + string(suffix)

	// Ensure it doesn't exceed 63 characters
	if len(suggested) > 63 {
		// Truncate the base to make room for the suffix
		maxBase := 63 - 5 // 5 = len("-") + len(suffix)
		suggested = baseUsername[:maxBase] + "-" + string(suffix)
	}

	return suggested
}
