package composition

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"unicode"

	"github.com/compose-spec/compose-go/v2/loader"
	composego "github.com/compose-spec/compose-go/v2/types"
	"go.uber.org/zap"
)

// ParseComposeFile parses docker-compose YAML content into a Project
func ParseComposeFile(composeContent string, projectName string, envData *EnvironmentData) (*composego.Project, error) {
	if composeContent == "" {
		return nil, fmt.Errorf("compose content is empty")
	}

	// Build environment map for compose parser
	environment := make(map[string]string)
	if envData != nil {
		// Add environment variables
		for k, v := range envData.EnvVars {
			environment[k] = v
		}
		// Add secrets
		for k, v := range envData.Secrets {
			environment[k] = v
		}
	}

	// Parse the compose file
	configDetails := composego.ConfigDetails{
		ConfigFiles: []composego.ConfigFile{
			{
				Content: []byte(composeContent),
			},
		},
		Environment: environment,
	}

	// Load and parse the project
	project, err := loader.LoadWithContext(context.Background(), configDetails, func(options *loader.Options) {
		options.SetProjectName(projectName, true)
		options.SkipConsistencyCheck = false
		options.SkipNormalization = true // Skip normalization to preserve /files/ volume references
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse compose file: %w", err)
	}

	if project == nil {
		return nil, fmt.Errorf("parsed project is nil")
	}

	// Validate ports in all services
	if err := validatePorts(project); err != nil {
		return nil, fmt.Errorf("port validation failed: %w", err)
	}

	// Validate service names, hostnames, and other resources
	if err := validateServices(project); err != nil {
		return nil, fmt.Errorf("service validation failed: %w", err)
	}

	// Validate volume names
	if err := validateVolumes(project); err != nil {
		return nil, fmt.Errorf("volume validation failed: %w", err)
	}

	// Validate network configuration
	if err := validateNetworks(project); err != nil {
		return nil, fmt.Errorf("network validation failed: %w", err)
	}

	// Allow empty services - user can add services later
	// No validation required here

	// Post-process to inject /files/ volume mounts that were filtered out during parsing
	// The compose parser filters out bind mounts with non-existent source paths
	// We need to manually parse the YAML to extract these special /files/ volumes
	logger := zap.L()
	if err := injectFilesVolumes(project, composeContent, envData, logger); err != nil {
		logger.Warn("Failed to inject /files/ volumes", zap.Error(err))
	}

	return project, nil
}

// validatePorts validates all port configurations in the compose project
func validatePorts(project *composego.Project) error {
	for serviceName, service := range project.Services {
		// Track seen ports to detect duplicates
		seenTargetPorts := make(map[uint32]int)
		seenPublishedPorts := make(map[string]int)

		for i, port := range service.Ports {
			// Validate port number (target/container port)
			if port.Target == 0 {
				return fmt.Errorf("service %s: port %d has invalid target port (must be 1-65535)", serviceName, i)
			}

			// Check if target port is in valid range
			if port.Target < 1 || port.Target > 65535 {
				return fmt.Errorf("service %s: port %d has target port %d outside valid range (1-65535)", serviceName, i, port.Target)
			}

			// Check for duplicate target ports
			if existingIdx, exists := seenTargetPorts[port.Target]; exists {
				return fmt.Errorf("service %s: port %d has duplicate target port %d (already defined at port %d)", serviceName, i, port.Target, existingIdx)
			}
			seenTargetPorts[port.Target] = i

			// Check for privileged ports if target is published
			// Ports below 1024 require special privileges in containers
			if port.Published != "" {
				// Validate published port format before parsing
				if err := validatePortFormat(port.Published); err != nil {
					return fmt.Errorf("service %s: port %d has invalid published port format '%s': %w", serviceName, i, port.Published, err)
				}

				publishedPort, err := strconv.ParseUint(port.Published, 10, 32)
				if err != nil {
					return fmt.Errorf("service %s: port %d has invalid published port format '%s': %w", serviceName, i, port.Published, err)
				}

				// Validate published port is in valid range
				if publishedPort < 1 || publishedPort > 65535 {
					return fmt.Errorf("service %s: port %d has published port %d outside valid range (1-65535)", serviceName, i, publishedPort)
				}

				// Check for duplicate published ports
				publishedPortKey := fmt.Sprintf("%s:%s", port.Published, port.HostIP)
				if existingIdx, exists := seenPublishedPorts[publishedPortKey]; exists {
					if port.HostIP != "" {
						return fmt.Errorf("service %s: port %d has duplicate published port %s on host IP %s (already defined at port %d)", serviceName, i, port.Published, port.HostIP, existingIdx)
					}
					return fmt.Errorf("service %s: port %d has duplicate published port %s (already defined at port %d)", serviceName, i, port.Published, existingIdx)
				}
				seenPublishedPorts[publishedPortKey] = i

				// Warn about privileged ports (1-1023) for published ports
				// These may require additional permissions or configuration
				if publishedPort < 1024 {
					// Privileged ports are technically valid but may not work in all environments
					// We allow them but the user should be aware of the implications
				}
			}

			// Validate protocol
			if port.Protocol != "" {
				if err := validateProtocol(port.Protocol); err != nil {
					return fmt.Errorf("service %s: port %d: %w", serviceName, i, err)
				}
			} else {
				// Default to TCP if not specified
				port.Protocol = "tcp"
			}

			// Validate mode (host or ingress)
			if port.Mode != "" {
				if err := validatePortMode(port.Mode); err != nil {
					return fmt.Errorf("service %s: port %d: %w", serviceName, i, err)
				}
			}

			// Validate port exposure type - ensure internal ports are properly configured
			if port.Published == "" && port.Mode == "" {
				// Internal port - this is valid, no validation needed
			} else if port.Published != "" && port.Mode == "host" {
				// Host mode with published port - valid
			} else if port.Published != "" && port.Mode == "" {
				// Published port without explicit mode - defaults to ingress/swarm mode
				// This is valid
			} else if port.Published == "" && port.Mode != "" {
				// Mode specified without published port - this is invalid
				return fmt.Errorf("service %s: port %d: mode '%s' specified without published port (mode only applies to published ports)", serviceName, i, port.Mode)
			}

			// Validate HostIP format if specified
			if port.HostIP != "" {
				if err := validateHostIP(port.HostIP); err != nil {
					return fmt.Errorf("service %s: port %d: %w", serviceName, i, err)
				}
			}
		}

		// Validate port count limits
		if len(service.Ports) == 0 {
			// Services without ports are valid (e.g., background jobs)
		} else if len(service.Ports) > 128 {
			return fmt.Errorf("service %s: too many ports defined (maximum 128, got %d)", serviceName, len(service.Ports))
		}
	}

	return nil
}

// validatePortFormat validates that a port string contains only valid characters
func validatePortFormat(portStr string) error {
	trimmed := strings.TrimSpace(portStr)
	if trimmed == "" {
		return fmt.Errorf("port cannot be empty")
	}

	// Check for invalid characters (only digits allowed)
	for _, ch := range trimmed {
		if ch < '0' || ch > '9' {
			return fmt.Errorf("port '%s' contains invalid characters (only digits allowed)", portStr)
		}
	}

	return nil
}

// validateProtocol validates the protocol is one of the supported values
func validateProtocol(protocol string) error {
	trimmed := strings.TrimSpace(protocol)
	if trimmed == "" {
		return fmt.Errorf("protocol cannot be empty")
	}

	lowerProtocol := strings.ToLower(trimmed)
	switch lowerProtocol {
	case "tcp", "udp", "sctp":
		return nil
	default:
		return fmt.Errorf("invalid protocol '%s' (must be tcp, udp, or sctp)", protocol)
	}
}

// validatePortMode validates the port mode
func validatePortMode(mode string) error {
	trimmed := strings.TrimSpace(mode)
	if trimmed == "" {
		return fmt.Errorf("mode cannot be empty")
	}

	lowerMode := strings.ToLower(trimmed)
	switch lowerMode {
	case "host", "ingress":
		return nil
	default:
		return fmt.Errorf("invalid mode '%s' (must be host or ingress)", mode)
	}
}

// validateHostIP validates the host IP address format
func validateHostIP(hostIP string) error {
	trimmed := strings.TrimSpace(hostIP)
	if trimmed == "" {
		return fmt.Errorf("host_ip cannot be empty")
	}

	// Check for invalid characters (no spaces allowed)
	if strings.Contains(trimmed, " ") {
		return fmt.Errorf("host_ip '%s' contains spaces", hostIP)
	}

	// Basic structural validation
	// - Must contain at least one digit
	// - Cannot be just dots or colons
	// - For IPv4: must have at least 3 dots for 4 octets
	// - For IPv6: must have at least one colon

	hasDigit := false
	hasDot := false
	hasColon := false
	for _, ch := range trimmed {
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		}
		if ch == '.' {
			hasDot = true
		}
		if ch == ':' {
			hasColon = true
		}
	}

	if !hasDigit {
		return fmt.Errorf("host_ip '%s' must contain at least one digit", hostIP)
	}

	// IPv4 validation (contains dots but no colons)
	if hasDot && !hasColon {
		// IPv4 addresses should only contain digits and dots (no letters)
		ipv4AllowedChars := "0123456789."
		for _, ch := range trimmed {
			if !strings.ContainsRune(ipv4AllowedChars, ch) {
				return fmt.Errorf("host_ip '%s' appears to be IPv4 but contains invalid character '%c' (IPv4 addresses only contain digits and dots)", hostIP, ch)
			}
		}

		// IPv4 should have at least 3 dots (e.g., 127.0.0.1)
		parts := strings.Split(trimmed, ".")
		if len(parts) < 4 {
			return fmt.Errorf("host_ip '%s' appears to be IPv4 but has invalid format (expected 4 octets, got %d)", hostIP, len(parts))
		}
	}

	// IPv6 validation (contains colons but no dots)
	if hasColon && !hasDot {
		// IPv6 addresses can contain letters (hex digits) and colons
		ipv6AllowedChars := "0123456789.:abcdefABCDEF"
		for _, ch := range trimmed {
			if !strings.ContainsRune(ipv6AllowedChars, ch) {
				return fmt.Errorf("host_ip '%s' appears to be IPv6 but contains invalid character '%c'", hostIP, ch)
			}
		}

		// IPv6 should have at least one colon
		if !strings.Contains(trimmed, ":") {
			return fmt.Errorf("host_ip '%s' appears to be IPv6 but has invalid format", hostIP)
		}
	}

	// Reject strings that are just dots
	if hasDot && !hasDigit {
		return fmt.Errorf("host_ip '%s' contains only dots and no digits", hostIP)
	}

	// Reject strings that mix dots and colons (invalid format)
	if hasDot && hasColon {
		return fmt.Errorf("host_ip '%s' cannot contain both dots and colons (mixed IPv4/IPv6 format is invalid)", hostIP)
	}

	return nil
}

// injectFilesVolumes manually parses YAML to find /files/ volume mounts that were filtered out
func injectFilesVolumes(project *composego.Project, composeContent string, envData *EnvironmentData, logger *zap.Logger) error {
	// Simple YAML parsing to extract volumes starting with /files/
	// This is a workaround for the compose parser filtering out non-existent bind mounts
	lines := strings.Split(composeContent, "\n")
	var currentService string
	inVolumesSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect service name
		if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			// Top-level key
			if strings.HasPrefix(trimmed, "services:") {
				continue
			}
		}

		// Detect service name under services
		if strings.HasPrefix(line, "  ") && strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(line, "    ") {
			currentService = strings.TrimSuffix(trimmed, ":")
			inVolumesSection = false
			continue
		}

		// Detect volumes section
		if strings.HasPrefix(line, "    ") && trimmed == "volumes:" {
			inVolumesSection = true
			continue
		}

		// Parse volume entries
		if inVolumesSection && strings.HasPrefix(line, "      - ") {
			volumeSpec := strings.TrimPrefix(trimmed, "- ")
			volumeSpec = strings.Trim(volumeSpec, "\"'")

			// Check if this is a /files/ volume
			if strings.HasPrefix(volumeSpec, "/files/") {
				parts := strings.SplitN(volumeSpec, ":", 2)
				if len(parts) == 2 {
					source := parts[0]
					target := parts[1]

					// Inject this volume into the project
					if svc, ok := project.Services[currentService]; ok {
						// Check if this volume already exists to avoid duplicates
						exists := false
						for _, v := range svc.Volumes {
							if v.Source == source && v.Target == target {
								exists = true
								break
							}
						}

						if !exists {
							svc.Volumes = append(svc.Volumes, composego.ServiceVolumeConfig{
								Type:   "bind",
								Source: source,
								Target: target,
							})
							project.Services[currentService] = svc
						}
					}
				}
			}
		} else if inVolumesSection && !strings.HasPrefix(line, "      ") {
			inVolumesSection = false
		}
	}

	return nil
}

// validateServices validates all service configurations in the project
func validateServices(project *composego.Project) error {
	for serviceName, service := range project.Services {
		// Validate service name
		if err := validateServiceName(serviceName); err != nil {
			return fmt.Errorf("service name '%s': %w", serviceName, err)
		}

		// Validate hostname if specified
		if service.Hostname != "" {
			if err := validateHostname(service.Hostname); err != nil {
				return fmt.Errorf("service %s: hostname '%s': %w", serviceName, service.Hostname, err)
			}
		}

		// Validate container name if specified
		if service.ContainerName != "" {
			if err := validateContainerName(service.ContainerName); err != nil {
				return fmt.Errorf("service %s: container_name '%s': %w", serviceName, service.ContainerName, err)
			}
		}

		// Validate domain name if specified
		if service.DomainName != "" {
			if err := validateDomainname(service.DomainName); err != nil {
				return fmt.Errorf("service %s: domainname '%s': %w", serviceName, service.DomainName, err)
			}
		}

		// Validate network mode if specified
		if service.Net != "" {
			if err := validateNetworkMode(service.Net); err != nil {
				return fmt.Errorf("service %s: network_mode '%s': %w", serviceName, service.Net, err)
			}
		}

		// Validate environment variables
		for envKey := range service.Environment {
			if err := validateEnvVarName(envKey); err != nil {
				return fmt.Errorf("service %s: environment variable '%s': %w", serviceName, envKey, err)
			}
		}

		// Validate image reference
		if service.Image == "" {
			return fmt.Errorf("service %s: image is required", serviceName)
		}
		if err := validateImageReference(service.Image); err != nil {
			return fmt.Errorf("service %s: image '%s': %w", serviceName, service.Image, err)
		}

		// Validate volume mounts
		for i, volume := range service.Volumes {
			if err := validateVolumeMount(volume); err != nil {
				return fmt.Errorf("service %s: volume %d: %w", serviceName, i, err)
			}
		}

		// Validate command and entrypoint (for security concerns)
		if len(service.Command) > 0 {
			for i, cmdPart := range service.Command {
				if err := validateCommandString(cmdPart, "command", i); err != nil {
					return fmt.Errorf("service %s: %w", serviceName, err)
				}
			}
		}
		if len(service.Entrypoint) > 0 {
			for i, entryPart := range service.Entrypoint {
				if err := validateCommandString(entryPart, "entrypoint", i); err != nil {
					return fmt.Errorf("service %s: %w", serviceName, err)
				}
			}
		}

		// Validate restart policy if specified
		if service.Restart != "" {
			if err := validateRestartPolicy(service.Restart); err != nil {
				return fmt.Errorf("service %s: restart policy '%s': %w", serviceName, service.Restart, err)
			}
		}

		// Validate working directory if specified
		if service.WorkingDir != "" {
			if err := validateWorkingDir(service.WorkingDir); err != nil {
				return fmt.Errorf("service %s: working_dir '%s': %w", serviceName, service.WorkingDir, err)
			}
		}

		// Validate user if specified
		if service.User != "" {
			if err := validateUser(service.User); err != nil {
				return fmt.Errorf("service %s: user '%s': %w", serviceName, service.User, err)
			}
		}
	}

	return nil
}

// validateServiceName validates a service name follows Kubernetes naming conventions
func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if len(name) > 63 {
		return fmt.Errorf("service name exceeds maximum length of 63 characters (got %d)", len(name))
	}

	// Service names must be lowercase alphanumeric or '-', and must start and end with alphanumeric
	// RFC 1123 subdomain format
	for i, ch := range name {
		if !(unicode.IsLower(ch) || unicode.IsDigit(ch) || ch == '-') {
			return fmt.Errorf("service name contains invalid character '%c' at position %d (only lowercase alphanumeric and '-' allowed)", ch, i)
		}
		if i == 0 || i == len(name)-1 {
			if ch == '-' {
				return fmt.Errorf("service name cannot start or end with '-'")
			}
		}
	}

	return nil
}

// validateHostname validates a hostname follows RFC 1123
func validateHostname(hostname string) error {
	trimmed := strings.TrimSpace(hostname)
	if trimmed == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if len(trimmed) > 253 {
		return fmt.Errorf("hostname exceeds maximum length of 253 characters (got %d)", len(trimmed))
	}

	// Hostname must be a valid DNS hostname
	// Split into labels (separated by dots)
	labels := strings.Split(trimmed, ".")
	for _, label := range labels {
		if label == "" {
			return fmt.Errorf("hostname contains empty label")
		}

		if len(label) > 63 {
			return fmt.Errorf("hostname label '%s' exceeds maximum length of 63 characters", label)
		}

		// Each label must start and end with alphanumeric, contain only alphanumeric or hyphen
		for i, ch := range label {
			if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-') {
				return fmt.Errorf("hostname label '%s' contains invalid character '%c' at position %d", label, ch, i)
			}
			if i == 0 || i == len(label)-1 {
				if ch == '-' {
					return fmt.Errorf("hostname label '%s' cannot start or end with '-'", label)
				}
			}
		}
	}

	return nil
}

// validateContainerName validates a container name
func validateContainerName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("container_name cannot be empty")
	}

	if len(trimmed) > 128 {
		return fmt.Errorf("container_name exceeds maximum length of 128 characters (got %d)", len(trimmed))
	}

	// Container names follow similar rules to hostnames but can contain underscores
	// Pattern: [a-zA-Z0-9][a-zA-Z0-9_.-]*
	for i, ch := range trimmed {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '-' || ch == '.') {
			return fmt.Errorf("container_name contains invalid character '%c' at position %d", ch, i)
		}
		if i == 0 && !(unicode.IsLetter(ch) || unicode.IsDigit(ch)) {
			return fmt.Errorf("container_name must start with a letter or digit")
		}
	}

	return nil
}

// validateDomainname validates a domain name
func validateDomainname(domainname string) error {
	trimmed := strings.TrimSpace(domainname)
	if trimmed == "" {
		return fmt.Errorf("domainname cannot be empty")
	}

	// Domain names follow similar rules to hostnames
	return validateHostname(trimmed)
}

// validateNetworkMode validates the network mode
func validateNetworkMode(mode string) error {
	trimmed := strings.TrimSpace(mode)
	if trimmed == "" {
		return fmt.Errorf("network_mode cannot be empty")
	}

	lowerMode := strings.ToLower(trimmed)

	// Standard Docker network modes
	standardModes := []string{"bridge", "host", "none", "service:", "container:"}

	for _, validMode := range standardModes {
		if strings.HasPrefix(lowerMode, strings.TrimSuffix(validMode, ":")) {
			// For service: and container:, we need to validate the reference
			if strings.HasPrefix(lowerMode, "service:") {
				serviceRef := strings.TrimPrefix(lowerMode, "service:")
				if serviceRef == "" {
					return fmt.Errorf("network_mode 'service:' requires a service name reference")
				}
				if err := validateServiceName(serviceRef); err != nil {
					return fmt.Errorf("network_mode 'service:' has invalid service reference: %w", err)
				}
			}
			if strings.HasPrefix(lowerMode, "container:") {
				containerRef := strings.TrimPrefix(lowerMode, "container:")
				if containerRef == "" {
					return fmt.Errorf("network_mode 'container:' requires a container name reference")
				}
				if err := validateContainerName(containerRef); err != nil {
					return fmt.Errorf("network_mode 'container:' has invalid container reference: %w", err)
				}
			}
			return nil
		}
	}

	return fmt.Errorf("invalid network_mode '%s' (must be one of: bridge, host, none, service:<name>, container:<name>)", mode)
}

// validateEnvVarName validates an environment variable name
func validateEnvVarName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("environment variable name cannot be empty")
	}

	// POSIX environment variable names: [A-Za-z_][A-Za-z0-9_]*
	if len(trimmed) > 255 {
		return fmt.Errorf("environment variable name exceeds maximum length of 255 characters (got %d)", len(trimmed))
	}

	// First character must be letter or underscore
	first := trimmed[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return fmt.Errorf("environment variable name must start with a letter or underscore (got '%c')", first)
	}

	// Remaining characters must be alphanumeric or underscore
	for i := 1; i < len(trimmed); i++ {
		ch := trimmed[i]
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return fmt.Errorf("environment variable name contains invalid character '%c' at position %d (only letters, digits, and underscore allowed)", ch, i)
		}
	}

	return nil
}

// validateImageReference validates a Docker image reference
func validateImageReference(image string) error {
	trimmed := strings.TrimSpace(image)
	if trimmed == "" {
		return fmt.Errorf("image cannot be empty")
	}

	if len(trimmed) > 255 {
		return fmt.Errorf("image reference exceeds maximum length of 255 characters (got %d)", len(trimmed))
	}

	// Check for invalid characters that could be security concerns
	// We allow: letters, digits, dots, dashes, underscores, colons, slashes, at signs, plus, tildes
	allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_/:@+~"
	for _, ch := range trimmed {
		if !strings.ContainsRune(allowedChars, ch) {
			return fmt.Errorf("image reference contains invalid character '%c'", ch)
		}
	}

	// Check for suspicious patterns that could be command injection attempts
	// For example: `$(...)`, backticks, etc.
	if strings.Contains(trimmed, "$(") || strings.Contains(trimmed, "`") {
		return fmt.Errorf("image reference contains potentially dangerous characters")
	}

	// Validate format: [registry/]image[:tag] or [registry/]image[@digest]
	parts := strings.Split(trimmed, ":")
	if len(parts) > 2 {
		// More than one colon suggests invalid format (unless it's an IPv6 address)
		// IPv6 addresses in registry URLs are valid: [::1]:5000/image
		if !strings.Contains(trimmed, "[") || !strings.Contains(trimmed, "]") {
			return fmt.Errorf("image reference has invalid format (too many colons)")
		}
	}

	// If tag is specified, validate it
	if len(parts) == 2 {
		tag := parts[1]
		// Tags can contain alphanumeric, dots, underscores, and hyphens
		for _, ch := range tag {
			if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' || ch == '_' || ch == '-') {
				return fmt.Errorf("image tag contains invalid character '%c'", ch)
			}
		}
	}

	return nil
}

// validateVolumeMount validates a volume mount configuration
func validateVolumeMount(volume composego.ServiceVolumeConfig) error {
	if volume.Source == "" && volume.Type != "tmpfs" {
		return fmt.Errorf("volume source cannot be empty (except for tmpfs volumes)")
	}

	if volume.Target == "" {
		return fmt.Errorf("volume target cannot be empty")
	}

	// Validate mount path format
	if strings.HasPrefix(volume.Target, "/") {
		// Absolute path - validate it's a reasonable path
		if len(volume.Target) > 4096 {
			return fmt.Errorf("volume target path exceeds maximum length of 4096 characters (got %d)", len(volume.Target))
		}

		// Check for path traversal attempts
		if strings.Contains(volume.Target, "../") || strings.Contains(volume.Target, "..\\") {
			return fmt.Errorf("volume target contains path traversal sequence '..'")
		}

		// Validate characters in path
		for _, ch := range volume.Target {
			if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '/' || ch == '_' || ch == '-' || ch == '.' || ch == '~' || ch == '@' || ch == ':' || ch == ' ') {
				return fmt.Errorf("volume target contains invalid character '%c'", ch)
			}
		}
	}

	// Validate volume type
	if volume.Type != "" {
		validTypes := []string{"volume", "bind", "tmpfs", "npipe"}
		isValidType := false
		for _, validType := range validTypes {
			if volume.Type == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			return fmt.Errorf("invalid volume type '%s' (must be one of: volume, bind, tmpfs, npipe)", volume.Type)
		}
	}

	return nil
}

// validateRestartPolicy validates the restart policy
func validateRestartPolicy(policy string) error {
	trimmed := strings.TrimSpace(policy)
	if trimmed == "" {
		return fmt.Errorf("restart policy cannot be empty")
	}

	lowerPolicy := strings.ToLower(trimmed)

	// Standard Docker restart policies
	standardPolicies := []string{"no", "always", "on-failure", "unless-stopped"}

	for _, validPolicy := range standardPolicies {
		if lowerPolicy == validPolicy {
			return nil
		}
	}

	// Check for on-failure with max retries: "on-failure:N"
	if strings.HasPrefix(lowerPolicy, "on-failure:") {
		retryStr := strings.TrimPrefix(lowerPolicy, "on-failure:")
		if retryStr == "" {
			return fmt.Errorf("restart policy 'on-failure:' requires a maximum retry count")
		}
		retries, err := strconv.Atoi(retryStr)
		if err != nil {
			return fmt.Errorf("restart policy 'on-failure:' has invalid retry count '%s' (must be a number)", retryStr)
		}
		if retries < 0 {
			return fmt.Errorf("restart policy 'on-failure:' retry count must be non-negative (got %d)", retries)
		}
		return nil
	}

	return fmt.Errorf("invalid restart policy '%s' (must be one of: no, always, on-failure[:N], unless-stopped)", policy)
}

// validateWorkingDir validates the working directory path
func validateWorkingDir(workingDir string) error {
	trimmed := strings.TrimSpace(workingDir)
	if trimmed == "" {
		return fmt.Errorf("working_dir cannot be empty")
	}

	if len(trimmed) > 4096 {
		return fmt.Errorf("working_dir exceeds maximum length of 4096 characters (got %d)", len(trimmed))
	}

	// Check for path traversal attempts
	if strings.Contains(trimmed, "../") || strings.Contains(trimmed, "..\\") {
		return fmt.Errorf("working_dir contains path traversal sequence '..'")
	}

	// Validate characters in path
	for _, ch := range trimmed {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '/' || ch == '\\' || ch == '_' || ch == '-' || ch == '.' || ch == '~' || ch == '@' || ch == ':' || ch == ' ') {
			return fmt.Errorf("working_dir contains invalid character '%c'", ch)
		}
	}

	return nil
}

// validateUser validates the user specification
func validateUser(user string) error {
	trimmed := strings.TrimSpace(user)
	if trimmed == "" {
		return fmt.Errorf("user cannot be empty")
	}

	if len(trimmed) > 64 {
		return fmt.Errorf("user specification exceeds maximum length of 64 characters (got %d)", len(trimmed))
	}

	// User can be: username, uid, or uid:gid
	// Pattern: [a-zA-Z0-9_.-]+|[0-9]+(:[0-9]+)?
	// We'll do basic validation

	// Check for invalid characters that could be security concerns
	// We allow: letters, digits, dots, dashes, underscores, colons
	allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_:"
	for _, ch := range trimmed {
		if !strings.ContainsRune(allowedChars, ch) {
			return fmt.Errorf("user specification contains invalid character '%c'", ch)
		}
	}

	// If it contains a colon, validate the gid part
	if strings.Contains(trimmed, ":") {
		parts := strings.Split(trimmed, ":")
		if len(parts) != 2 {
			return fmt.Errorf("user specification must have at most one colon (uid:gid format)")
		}
		if parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("user specification in uid:gid format cannot have empty uid or gid")
		}

		// Validate uid and gid are numbers (or usernames)
		for i, part := range parts {
			// Check if it's all digits (UID/GID)
			allDigits := true
			for _, ch := range part {
				if !unicode.IsDigit(ch) {
					allDigits = false
					break
				}
			}

			// If not all digits, it's a username which we already validated characters for
			if allDigits {
				num, err := strconv.ParseUint(part, 10, 32)
				if err != nil {
					return fmt.Errorf("user specification part %d is not a valid number: %s", i+1, part)
				}
				if num > 4294967295 { // Max 32-bit unsigned int
					return fmt.Errorf("user specification part %d exceeds maximum value (4294967295)", i+1)
				}
			}
		}
	}

	return nil
}

// validateCommandString validates a command or entrypoint string for security concerns
func validateCommandString(cmd string, cmdType string, index int) error {
	if cmd == "" {
		return fmt.Errorf("%s part %d cannot be empty", cmdType, index)
	}

	// Maximum command length to prevent excessively long commands
	if len(cmd) > 4096 {
		return fmt.Errorf("%s part %d exceeds maximum length of 4096 characters (got %d)", cmdType, index, len(cmd))
	}

	// Check for potentially dangerous shell metacharacters
	// We allow most characters but warn about some that could be injection risks
	// Note: This is a basic check - actual command execution happens in the container

	// Check for null bytes
	if strings.Contains(cmd, "\x00") {
		return fmt.Errorf("%s part %d contains null byte (possible injection attempt)", cmdType, index)
	}

	// We're conservative: allow printable ASCII and basic Unicode
	for i, ch := range cmd {
		if ch < 32 && ch != '\t' && ch != '\n' && ch != '\r' {
			// Control characters (except tab, newline, carriage return) are suspicious
			return fmt.Errorf("%s part %d contains control character at position %d", cmdType, index, i)
		}
	}

	return nil
}

// validateVolumes validates all volume definitions in the project
func validateVolumes(project *composego.Project) error {
	for volumeName, volume := range project.Volumes {
		// Validate volume name
		if err := validateVolumeName(volumeName); err != nil {
			return fmt.Errorf("volume name '%s': %w", volumeName, err)
		}

		// Validate volume driver if specified
		if volume.Driver != "" {
			if err := validateVolumeDriver(volume.Driver); err != nil {
				return fmt.Errorf("volume '%s' driver '%s': %w", volumeName, volume.Driver, err)
			}
		}

		// Validate volume options if specified
		for optionKey, optionValue := range volume.DriverOpts {
			if err := validateVolumeOption(optionKey, optionValue); err != nil {
				return fmt.Errorf("volume '%s' option '%s=%s': %w", volumeName, optionKey, optionValue, err)
			}
		}
	}

	return nil
}

// validateVolumeName validates a volume name
func validateVolumeName(name string) error {
	if name == "" {
		return fmt.Errorf("volume name cannot be empty")
	}

	if len(name) > 128 {
		return fmt.Errorf("volume name exceeds maximum length of 128 characters (got %d)", len(name))
	}

	// Volume names follow similar rules to service names (RFC 1123)
	for i, ch := range name {
		if !(unicode.IsLower(ch) || unicode.IsDigit(ch) || ch == '-') {
			return fmt.Errorf("volume name contains invalid character '%c' at position %d (only lowercase alphanumeric and '-' allowed)", ch, i)
		}
		if i == 0 || i == len(name)-1 {
			if ch == '-' {
				return fmt.Errorf("volume name cannot start or end with '-'")
			}
		}
	}

	return nil
}

// validateVolumeDriver validates a volume driver name
func validateVolumeDriver(driver string) error {
	trimmed := strings.TrimSpace(driver)
	if trimmed == "" {
		return fmt.Errorf("volume driver cannot be empty")
	}

	if len(trimmed) > 64 {
		return fmt.Errorf("volume driver exceeds maximum length of 64 characters (got %d)", len(trimmed))
	}

	// Driver names are alphanumeric with dots and hyphens
	for _, ch := range trimmed {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("volume driver contains invalid character '%c'", ch)
		}
	}

	return nil
}

// validateVolumeOption validates a volume option key and value
func validateVolumeOption(key, value string) error {
	// Validate key
	if key == "" {
		return fmt.Errorf("option key cannot be empty")
	}

	if len(key) > 256 {
		return fmt.Errorf("option key exceeds maximum length of 256 characters (got %d)", len(key))
	}

	// Key format: alphanumeric with dots, hyphens, underscores
	for _, ch := range key {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("option key contains invalid character '%c'", ch)
		}
	}

	// Validate value (can be more flexible)
	if len(value) > 1024 {
		return fmt.Errorf("option value exceeds maximum length of 1024 characters (got %d)", len(value))
	}

	// Check for suspicious patterns in value
	if strings.Contains(value, "\x00") {
		return fmt.Errorf("option value contains null byte")
	}

	return nil
}

// validateNetworks validates all network configurations in the project
func validateNetworks(project *composego.Project) error {
	for networkName, network := range project.Networks {
		// Validate network name
		if err := validateNetworkName(networkName); err != nil {
			return fmt.Errorf("network name '%s': %w", networkName, err)
		}

		// Validate network driver if specified
		if network.Driver != "" {
			if err := validateNetworkDriver(network.Driver); err != nil {
				return fmt.Errorf("network '%s' driver '%s': %w", networkName, network.Driver, err)
			}
		}

		// Validate IPAM configuration if specified
		if len(network.Ipam.Config) > 0 {
			for _, config := range network.Ipam.Config {
				if config.Subnet != "" {
					if err := validateIPSubnet(config.Subnet); err != nil {
						return fmt.Errorf("network '%s' IPAM subnet '%s': %w", networkName, config.Subnet, err)
					}
				}
				if config.IPRange != "" {
					if err := validateIPSubnet(config.IPRange); err != nil {
						return fmt.Errorf("network '%s' IPAM range '%s': %w", networkName, config.IPRange, err)
					}
				}
				if config.Gateway != "" {
					if err := validateIPAddress(config.Gateway); err != nil {
						return fmt.Errorf("network '%s' IPAM gateway '%s': %w", networkName, config.Gateway, err)
					}
				}
			}
		}
	}

	return nil
}

// validateNetworkName validates a network name
func validateNetworkName(name string) error {
	if name == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	if len(name) > 31 {
		return fmt.Errorf("network name exceeds maximum length of 31 characters (got %d)", len(name))
	}

	// Network names follow similar rules to service names
	for i, ch := range name {
		if !(unicode.IsLower(ch) || unicode.IsDigit(ch) || ch == '-') {
			return fmt.Errorf("network name contains invalid character '%c' at position %d (only lowercase alphanumeric and '-' allowed)", ch, i)
		}
		if i == 0 || i == len(name)-1 {
			if ch == '-' {
				return fmt.Errorf("network name cannot start or end with '-'")
			}
		}
	}

	return nil
}

// validateNetworkDriver validates a network driver name
func validateNetworkDriver(driver string) error {
	trimmed := strings.TrimSpace(driver)
	if trimmed == "" {
		return fmt.Errorf("network driver cannot be empty")
	}

	if len(trimmed) > 32 {
		return fmt.Errorf("network driver exceeds maximum length of 32 characters (got %d)", len(trimmed))
	}

	// Standard Docker network drivers
	standardDrivers := []string{"bridge", "overlay", "host", "macvlan", "ipvlan", "none"}

	for _, validDriver := range standardDrivers {
		if strings.ToLower(trimmed) == validDriver {
			return nil
		}
	}

	// Custom drivers are allowed, but validate the format
	for _, ch := range trimmed {
		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '.' || ch == '-' || ch == '_') {
			return fmt.Errorf("network driver contains invalid character '%c'", ch)
		}
	}

	return nil
}

// validateIPSubnet validates an IP subnet (CIDR notation)
func validateIPSubnet(subnet string) error {
	trimmed := strings.TrimSpace(subnet)
	if trimmed == "" {
		return fmt.Errorf("IP subnet cannot be empty")
	}

	// Use net.ParseCIDR to validate CIDR notation
	_, ipNet, err := net.ParseCIDR(trimmed)
	if err != nil {
		return fmt.Errorf("invalid IP subnet '%s': %w", trimmed, err)
	}

	// Validate the IP is not multicast or reserved
	if ipNet.IP.IsMulticast() {
		return fmt.Errorf("IP subnet '%s' is a multicast address which is not allowed", trimmed)
	}
	if ipNet.IP.IsLinkLocalUnicast() {
		return fmt.Errorf("IP subnet '%s' is a link-local address which is not recommended", trimmed)
	}

	return nil
}

// validateIPAddress validates an IP address
func validateIPAddress(ipAddr string) error {
	trimmed := strings.TrimSpace(ipAddr)
	if trimmed == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	// Use net.ParseIP to validate IP address
	ip := net.ParseIP(trimmed)
	if ip == nil {
		return fmt.Errorf("invalid IP address '%s'", trimmed)
	}

	// Validate the IP is not multicast or reserved
	if ip.IsMulticast() {
		return fmt.Errorf("IP address '%s' is a multicast address which is not allowed", trimmed)
	}

	return nil
}
