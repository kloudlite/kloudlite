package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestParse_ComposeFile_WithFilesVolumes(t *testing.T) {
	composeContent := `services:
  web:
    image: nginx:latest
    volumes:
      - "/files/app.yml:/etc/nginx/nginx.conf"
      - "/files/config.json:/app/config.json"
`

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: map[string]string{
			"app.yml":     "test content",
			"config.json": "{}",
		},
	}

	project, err := ParseComposeFile(composeContent, "test", envData)
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, 1, len(project.Services))

	webService := project.Services["web"]
	assert.Equal(t, 2, len(webService.Volumes))
	assert.Equal(t, "/files/app.yml", webService.Volumes[0].Source)
	assert.Equal(t, "/etc/nginx/nginx.conf", webService.Volumes[0].Target)
	assert.Equal(t, "/files/config.json", webService.Volumes[1].Source)
	assert.Equal(t, "/app/config.json", webService.Volumes[1].Target)
}

func TestInjectFilesVolumes_BasicInjection(t *testing.T) {
	composeContent := `services:
  web:
    image: nginx:latest
    volumes:
      - "/files/app.yml:/etc/nginx/nginx.conf"
`

	project, err := ParseComposeFile(composeContent, "test", nil)
	assert.NoError(t, err)

	webService := project.Services["web"]
	assert.Equal(t, 1, len(webService.Volumes))
	assert.Equal(t, "/files/app.yml", webService.Volumes[0].Source)
	assert.Equal(t, "/etc/nginx/nginx.conf", webService.Volumes[0].Target)
	assert.Equal(t, "bind", webService.Volumes[0].Type)
}

func TestInjectFilesVolumes_MultipleServices(t *testing.T) {
	composeContent := `services:
  web:
    image: nginx:latest
    volumes:
      - "/files/nginx.conf:/etc/nginx/nginx.conf"
  app:
    image: node:latest
    volumes:
      - "/files/app.config:/app/config.json"
`

	project, err := ParseComposeFile(composeContent, "test", nil)
	assert.NoError(t, err)

	webService := project.Services["web"]
	assert.Equal(t, 1, len(webService.Volumes))
	assert.Equal(t, "/files/nginx.conf", webService.Volumes[0].Source)

	appService := project.Services["app"]
	assert.Equal(t, 1, len(appService.Volumes))
	assert.Equal(t, "/files/app.config", appService.Volumes[0].Source)
}

func TestInjectFilesVolumes_MixedVolumes(t *testing.T) {
	composeContent := `services:
  web:
    image: nginx:latest
    volumes:
      - "/files/nginx.conf:/etc/nginx/nginx.conf"
      - "data:/var/lib/data"

volumes:
  data:
`

	project, err := ParseComposeFile(composeContent, "test", nil)
	assert.NoError(t, err)

	webService := project.Services["web"]
	// Should have both /files/ and named volume
	assert.GreaterOrEqual(t, len(webService.Volumes), 1)

	// Check that /files/ volume was injected
	foundFilesVolume := false
	for _, vol := range webService.Volumes {
		if vol.Source == "/files/nginx.conf" {
			foundFilesVolume = true
			assert.Equal(t, "/etc/nginx/nginx.conf", vol.Target)
		}
	}
	assert.True(t, foundFilesVolume, "/files/ volume should be injected")
}

func TestInjectFilesVolumes_NoDuplicates(t *testing.T) {
	composeContent := `services:
  web:
    image: nginx:latest
    volumes:
      - "/files/app.yml:/etc/nginx/nginx.conf"
`

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: map[string]string{"app.yml": "content"},
	}

	// Parse twice to test deduplication
	project, err := ParseComposeFile(composeContent, "test", envData)
	assert.NoError(t, err)

	logger, _ := zap.NewDevelopment()
	// Inject again (simulating multiple reconciliations)
	err = injectFilesVolumes(project, composeContent, envData, logger)
	assert.NoError(t, err)

	webService := project.Services["web"]
	// Should still have only 1 volume (no duplicates)
	assert.Equal(t, 1, len(webService.Volumes))
}

func TestInjectFilesVolumes_QuotedVolumes(t *testing.T) {
	composeContent := `services:
  web:
    image: nginx:latest
    volumes:
      - "/files/app.yml:/etc/nginx/nginx.conf"
      - '/files/config.json:/app/config.json'
`

	project, err := ParseComposeFile(composeContent, "test", nil)
	assert.NoError(t, err)

	webService := project.Services["web"]
	assert.Equal(t, 2, len(webService.Volumes))
}

func TestParseComposeFile_WithEnvVars(t *testing.T) {
	composeContent := `services:
  web:
    image: nginx:latest
    environment:
      API_URL: $API_ENDPOINT
      SECRET: $DB_PASSWORD
`

	envData := &EnvironmentData{
		EnvVars: map[string]string{
			"API_ENDPOINT": "https://api.example.com",
		},
		Secrets: map[string]string{
			"DB_PASSWORD": "secret123",
		},
		ConfigFiles: make(map[string]string),
	}

	project, err := ParseComposeFile(composeContent, "test", envData)
	assert.NoError(t, err)
	assert.NotNil(t, project)

	webService := project.Services["web"]
	// Compose parser should resolve environment variables
	assert.NotNil(t, webService.Environment)
}
