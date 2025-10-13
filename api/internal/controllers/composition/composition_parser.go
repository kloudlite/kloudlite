package composition

import (
	"context"
	"fmt"
	"strings"

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

	// Validate that we have at least one service
	if len(project.Services) == 0 {
		return nil, fmt.Errorf("no services found in compose file")
	}

	// Post-process to inject /files/ volume mounts that were filtered out during parsing
	// The compose parser filters out bind mounts with non-existent source paths
	// We need to manually parse the YAML to extract these special /files/ volumes
	logger := zap.L()
	if err := injectFilesVolumes(project, composeContent, envData, logger); err != nil {
		logger.Warn("Failed to inject /files/ volumes", zap.Error(err))
	}

	return project, nil
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
