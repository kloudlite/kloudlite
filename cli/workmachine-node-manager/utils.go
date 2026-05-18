package main

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

// truncateError truncates an error message to maxLen characters
func truncateError(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// containsString checks if a string is present in a slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a string from a slice
func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// sanitizeLabelValue converts a string to a valid Kubernetes label value
// Kubernetes labels must match regex: (([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?
// and be at most 63 characters long
func sanitizeLabelValue(s string, maxLen int) string {
	if maxLen > 63 {
		maxLen = 63
	}
	if maxLen < 1 {
		maxLen = 1
	}

	// Replace spaces and invalid characters with hyphens
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, s)

	// Truncate to max length
	if len(sanitized) > maxLen {
		sanitized = sanitized[:maxLen]
	}

	// Ensure it starts and ends with alphanumeric
	// Trim leading/trailing non-alphanumeric chars
	sanitized = strings.TrimFunc(sanitized, func(r rune) bool {
		return !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	// If empty after sanitization, return a default
	if sanitized == "" {
		return "unknown"
	}

	return sanitized
}

// formatSnapshotSize formats bytes into human readable format
func formatSnapshotSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// parseQuantity parses a string into a Kubernetes resource.Quantity
func parseQuantity(value string) *resource.Quantity {
	q, err := resource.ParseQuantity(value)
	if err != nil {
		// Fallback to 0 if parsing fails
		return resource.NewQuantity(0, resource.DecimalSI)
	}
	return &q
}
