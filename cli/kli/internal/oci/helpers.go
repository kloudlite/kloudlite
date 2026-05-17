package oci

// shortKey returns the first 8 characters of an installation key for resource naming
func shortKey(installationKey string) string {
	if len(installationKey) > 8 {
		return installationKey[:8]
	}
	return installationKey
}

// freeformTags returns the standard freeform tags for Kloudlite resources
func freeformTags(installationKey string) map[string]string {
	return map[string]string{
		"managed-by":      "kloudlite",
		"installation-id": installationKey,
		"project":         "kloudlite",
	}
}

func boolPtr(b bool) *bool {
	return &b
}
