package controllers

import (
	"context"
	"fmt"

	"github.com/compose-spec/compose-go/v2/loader"
	composego "github.com/compose-spec/compose-go/v2/types"
)

// ParseComposeFile parses docker-compose YAML content into a Project
func ParseComposeFile(composeContent string, projectName string) (*composego.Project, error) {
	if composeContent == "" {
		return nil, fmt.Errorf("compose content is empty")
	}

	// Parse the compose file
	configDetails := composego.ConfigDetails{
		ConfigFiles: []composego.ConfigFile{
			{
				Content: []byte(composeContent),
			},
		},
		Environment: make(map[string]string),
	}

	// Load and parse the project
	project, err := loader.LoadWithContext(context.Background(), configDetails, func(options *loader.Options) {
		options.SetProjectName(projectName, true)
		options.SkipConsistencyCheck = false
		options.SkipNormalization = false
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

	return project, nil
}
