package composition

import (
	"testing"

	composego "github.com/compose-spec/compose-go/v2/types"
	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCommandEntrypointHandling(t *testing.T) {
	tests := []struct {
		name               string
		serviceName        string
		service            composego.ServiceConfig
		wantCommand        []string
		wantArgs           []string
		description        string
	}{
		{
			name:        "only command specified",
			serviceName: "web",
			service: composego.ServiceConfig{
				Image:   "nginx:latest",
				Command: []string{"nginx", "-g", "daemon off;"},
			},
			wantCommand: []string{"nginx", "-g", "daemon off;"},
			wantArgs:    nil,
			description: "Command should be set as container.Command when only command is specified",
		},
		{
			name:        "only entrypoint specified",
			serviceName: "web",
			service: composego.ServiceConfig{
				Image:      "nginx:latest",
				Entrypoint: []string{"/docker-entrypoint.sh"},
			},
			wantCommand: []string{"/docker-entrypoint.sh"},
			wantArgs:    nil,
			description: "Entrypoint should be set as container.Command when only entrypoint is specified",
		},
		{
			name:        "both command and entrypoint specified",
			serviceName: "web",
			service: composego.ServiceConfig{
				Image:      "nginx:latest",
				Entrypoint: []string{"/docker-entrypoint.sh"},
				Command:    []string{"nginx", "-g", "daemon off;"},
			},
			wantCommand: []string{"/docker-entrypoint.sh"},
			wantArgs:    []string{"nginx", "-g", "daemon off;"},
			description: "Entrypoint should be Command and Command should be Args when both are specified",
		},
		{
			name:        "neither command nor entrypoint specified",
			serviceName: "web",
			service: composego.ServiceConfig{
				Image: "nginx:latest",
			},
			wantCommand: nil,
			wantArgs:    nil,
			description: "Neither Command nor Args should be set when neither is specified",
		},
		{
			name:        "command with multiple args",
			serviceName: "app",
			service: composego.ServiceConfig{
				Image:   "node:18",
				Command: []string{"node", "server.js", "--port", "3000"},
			},
			wantCommand: []string{"node", "server.js", "--port", "3000"},
			wantArgs:    nil,
			description: "Command with multiple arguments should be preserved",
		},
		{
			name:        "entrypoint with command as args",
			serviceName: "db",
			service: composego.ServiceConfig{
				Image:      "postgres:15",
				Entrypoint: []string{"/docker-entrypoint.sh"},
				Command:    []string{"postgres"},
			},
			wantCommand: []string{"/docker-entrypoint.sh"},
			wantArgs:    []string{"postgres"},
			description: "Entrypoint wraps command as arguments",
		},
		{
			name:        "complex command and entrypoint",
			serviceName: "worker",
			service: composego.ServiceConfig{
				Image:      "python:3.11",
				Entrypoint: []string{"/app/entrypoint.sh"},
				Command:    []string{"python", "-m", "celery", "-A", "app", "worker"},
			},
			wantCommand: []string{"/app/entrypoint.sh"},
			wantArgs:    []string{"python", "-m", "celery", "-A", "app", "worker"},
			description: "Complex entrypoint/command combination should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal composition and environment for the conversion
			composition := &compositionsv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-composition",
					Namespace: "test-namespace",
				},
				Spec: compositionsv1.CompositionSpec{
					EnvVars: make(map[string]string),
				},
			}

			envData := &EnvironmentData{
				EnvVars:     make(map[string]string),
				Secrets:     make(map[string]string),
				ConfigFiles: make(map[string]string),
			}

			statefulSet, err := convertServiceToStatefulSet(
				tt.serviceName,
				tt.service,
				composition,
				"test-namespace",
				map[string]string{},
				envData,
				nil,
			)

			assert.NoError(t, err, "convertServiceToStatefulSet should not return an error")
			assert.NotNil(t, statefulSet, "StatefulSet should not be nil")
			assert.Len(t, statefulSet.Spec.Template.Spec.Containers, 1, "Should have exactly one container")

			container := statefulSet.Spec.Template.Spec.Containers[0]

			assert.Equal(t, tt.wantCommand, container.Command,
				"container.Command should match expected command - %s", tt.description)
			assert.Equal(t, tt.wantArgs, container.Args,
				"container.Args should match expected args - %s", tt.description)
		})
	}
}

func TestConvertComposeToK8s_CommandEntrypoint(t *testing.T) {
	tests := []struct {
		name        string
		composeYAML string
		serviceName string
		wantCommand []string
		wantArgs    []string
		wantErr     bool
		description string
	}{
		{
			name: "service with command only",
			composeYAML: `
services:
  web:
    image: nginx:latest
    command: ["nginx", "-g", "daemon off;"]
`,
			serviceName: "web",
			wantCommand: []string{"nginx", "-g", "daemon off;"},
			wantArgs:    nil,
			wantErr:     false,
			description: "Command should be set in container.Command",
		},
		{
			name: "service with entrypoint only",
			composeYAML: `
services:
  web:
    image: nginx:latest
    entrypoint: ["/docker-entrypoint.sh"]
`,
			serviceName: "web",
			wantCommand: []string{"/docker-entrypoint.sh"},
			wantArgs:    nil,
			wantErr:     false,
			description: "Entrypoint should be set in container.Command",
		},
		{
			name: "service with both command and entrypoint",
			composeYAML: `
services:
  web:
    image: nginx:latest
    entrypoint: ["/docker-entrypoint.sh"]
    command: ["nginx", "-g", "daemon off;"]
`,
			serviceName: "web",
			wantCommand: []string{"/docker-entrypoint.sh"},
			wantArgs:    []string{"nginx", "-g", "daemon off;"},
			wantErr:     false,
			description: "Entrypoint in Command, Command in Args",
		},
		{
			name: "multiple services with different command/entrypoint combos",
			composeYAML: `
services:
  web:
    image: nginx:latest
    command: ["nginx", "-g", "daemon off;"]
  api:
    image: node:18
    entrypoint: ["/app/start.sh"]
  worker:
    image: python:3.11
    entrypoint: ["/app/entrypoint.sh"]
    command: ["python", "worker.py"]
`,
			serviceName: "worker",
			wantCommand: []string{"/app/entrypoint.sh"},
			wantArgs:    []string{"python", "worker.py"},
			wantErr:     false,
			description: "Multiple services with different configurations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := ParseComposeFile(tt.composeYAML, "test-project", nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			composition := &compositionsv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-composition",
					Namespace: "test-namespace",
				},
				Spec: compositionsv1.CompositionSpec{
					EnvVars: make(map[string]string),
				},
			}

			envData := &EnvironmentData{
				EnvVars:     make(map[string]string),
				Secrets:     make(map[string]string),
				ConfigFiles: make(map[string]string),
			}

			resources, err := ConvertComposeToK8s(project, composition, "test-namespace", envData, nil)
			assert.NoError(t, err)
			assert.NotNil(t, resources)

			// Find the target service's StatefulSet
			var targetStatefulSet *corev1.Container
			for _, sts := range resources.StatefulSets {
				if sts.Name == tt.serviceName {
					if len(sts.Spec.Template.Spec.Containers) > 0 {
						targetStatefulSet = &sts.Spec.Template.Spec.Containers[0]
					}
					break
				}
			}

			assert.NotNil(t, targetStatefulSet, "Should find StatefulSet for service %s", tt.serviceName)

			assert.Equal(t, tt.wantCommand, targetStatefulSet.Command,
				"container.Command should match expected - %s", tt.description)
			assert.Equal(t, tt.wantArgs, targetStatefulSet.Args,
				"container.Args should match expected - %s", tt.description)
		})
	}
}

func TestConvertCPU(t *testing.T) {
	tests := []struct {
		name      string
		nanoCPUs  float64
		want      string
		wantErr   bool
		precision bool // check if conversion is precise
	}{
		// Standard values
		{
			name:      "0.5 CPU cores to 500m",
			nanoCPUs:  0.5,
			want:      "500m",
			wantErr:   false,
			precision: true,
		},
		{
			name:      "1 CPU core to 1000m",
			nanoCPUs:  1.0,
			want:      "1000m",
			wantErr:   false,
			precision: true,
		},
		{
			name:      "2 CPU cores to 2000m",
			nanoCPUs:  2.0,
			want:      "2000m",
			wantErr:   false,
			precision: true,
		},
		{
			name:      "0.1 CPU cores to 100m",
			nanoCPUs:  0.1,
			want:      "100m",
			wantErr:   false,
			precision: true,
		},
		{
			name:      "0.001 CPU cores to 1m",
			nanoCPUs:  0.001,
			want:      "1m",
			wantErr:   false,
			precision: true,
		},

		// Edge cases - near zero
		{
			name:      "very small value 0.0001 CPU cores",
			nanoCPUs:  0.0001,
			want:      "1m",
			wantErr:   false,
			precision: false, // rounded up to 1m
		},
		{
			name:      "zero value returns default 1 CPU",
			nanoCPUs:  0.0,
			want:      "1",
			wantErr:   false,
			precision: false,
		},

		// Edge cases - negative values
		{
			name:      "negative value returns default 1 CPU",
			nanoCPUs:  -0.5,
			want:      "1",
			wantErr:   false,
			precision: false,
		},

		// Edge cases - NaN and infinity
		{
			name:      "NaN returns default 1 CPU",
			nanoCPUs:  0.0 / 0.0,
			want:      "1",
			wantErr:   false,
			precision: false,
		},
		{
			name:      "positive infinity returns default 1 CPU",
			nanoCPUs:  1.0 / 0.0,
			want:      "1",
			wantErr:   false,
			precision: false,
		},

		// Precision tests
		{
			name:      "0.999 CPU cores - ceiling to 999m",
			nanoCPUs:  0.999,
			want:      "999m",
			wantErr:   false,
			precision: true,
		},
		{
			name:      "0.001 CPU cores - ensure not truncated to 0",
			nanoCPUs:  0.0009,
			want:      "1m",
			wantErr:   false,
			precision: false, // rounded up
		},

		// Large values
		{
			name:      "large value 100 CPU cores",
			nanoCPUs:  100.0,
			want:      "100000m",
			wantErr:   false,
			precision: true,
		},

		// Overflow protection
		{
			name:      "value at max int32 boundary",
			nanoCPUs:  2147483.647, // max int32 / 1000
			want:      "2147483647m",
			wantErr:   false,
			precision: true,
		},
		{
			name:      "above max int32 (capped)",
			nanoCPUs:  2147484.0, // above max int32 / 1000
			want:      "2147483647m",
			wantErr:   false,
			precision: false,
		},

		// Fractional precision
		{
			name:      "0.0015 CPU cores - ceiling to 2m",
			nanoCPUs:  0.0015,
			want:      "2m",
			wantErr:   false,
			precision: false, // ceiling
		},
		{
			name:      "0.0006 CPU cores - ceiling to 1m",
			nanoCPUs:  0.0006,
			want:      "1m",
			wantErr:   false,
			precision: false, // ceiling from 0.6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertCPU(tt.nanoCPUs)

			if !tt.wantErr {
				assert.Equal(t, tt.want, got, "convertCPU(%f) = %v, want %v", tt.nanoCPUs, got, tt.want)

				// Verify the result can be parsed by Kubernetes resource parser
				if tt.precision && got != "1" {
					// Skip parsing for default "1" case and imprecise conversions
					_, err := strconv.ParseInt(strings.TrimSuffix(got, "m"), 10, 64)
					assert.NoError(t, err, "result should be parseable as integer")
				}
			}
		})
	}
}

func TestConvertCPUIntegration(t *testing.T) {
	// Integration test: verify convertCPU works in actual conversion
	tests := []struct {
		name           string
		composeYAML    string
		serviceName    string
		expectedCPU    string
		description    string
	}{
		{
			name: "service with 0.5 CPU limit",
			composeYAML: `
services:
  web:
    image: nginx:latest
    deploy:
      resources:
        limits:
          cpus: '0.5'
`,
			serviceName: "web",
			expectedCPU: "500m",
			description: "0.5 CPU should convert to 500m",
		},
		{
			name: "service with 2 CPU limit",
			composeYAML: `
services:
  api:
    image: node:18
    deploy:
      resources:
        limits:
          cpus: '2'
`,
			serviceName: "api",
			expectedCPU: "2000m",
			description: "2 CPU should convert to 2000m",
		},
		{
			name: "service with 0.1 CPU limit",
			composeYAML: `
services:
  worker:
    image: python:3.11
    deploy:
      resources:
        limits:
          cpus: '0.1'
`,
			serviceName: "worker",
			expectedCPU: "100m",
			description: "0.1 CPU should convert to 100m",
		},
		{
			name: "service with 0.001 CPU limit (minimum)",
			composeYAML: `
services:
  tiny:
    image: alpine:latest
    deploy:
      resources:
        limits:
          cpus: '0.001'
`,
			serviceName: "tiny",
			expectedCPU: "1m",
			description: "0.001 CPU should convert to 1m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := ParseComposeFile(tt.composeYAML, "test-project", nil)
			assert.NoError(t, err)

			composition := &compositionsv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-composition",
					Namespace: "test-namespace",
				},
				Spec: compositionsv1.CompositionSpec{
					EnvVars: make(map[string]string),
				},
			}

			envData := &EnvironmentData{
				EnvVars:     make(map[string]string),
				Secrets:     make(map[string]string),
				ConfigFiles: make(map[string]string),
			}

			resources, err := ConvertComposeToK8s(project, composition, "test-namespace", envData, nil)
			assert.NoError(t, err)
			assert.NotNil(t, resources)

			// Find the target service's StatefulSet
			var container *corev1.Container
			for _, sts := range resources.StatefulSets {
				if sts.Name == tt.serviceName && len(sts.Spec.Template.Spec.Containers) > 0 {
					container = &sts.Spec.Template.Spec.Containers[0]
					break
				}
			}

			assert.NotNil(t, container, "Should find container for service %s", tt.serviceName)

			// Check CPU limit
			cpuLimit, exists := container.Resources.Limits[corev1.ResourceCPU]
			assert.True(t, exists, "CPU limit should be set - %s", tt.description)
			assert.Equal(t, tt.expectedCPU, cpuLimit.String(),
				"CPU limit should match expected - %s", tt.description)
		})
	}
}

func TestCommandEntrypointRegressionTests(t *testing.T) {
	// These tests specifically check for the bug where entrypoint was overwriting command
	tests := []struct {
		name           string
		composeYAML    string
		serviceName    string
		expectedCmd    string
		expectedArg    string
		description    string
	}{
		{
			name: "regression test - command not lost when both specified",
			composeYAML: `
services:
  app:
    image: myapp:latest
    entrypoint: ["/entrypoint.sh"]
    command: ["run", "--port", "8080"]
`,
			serviceName: "app",
			expectedCmd: "/entrypoint.sh",
			expectedArg: "run",
			description: "Command should be preserved as Args, not lost",
		},
		{
			name: "regression test - entrypoint does not replace command",
			composeYAML: `
services:
  db:
    image: postgres:15
    command: ["postgres", "-c", "max_connections=200"]
`,
			serviceName: "db",
			expectedCmd: "postgres",
			expectedArg: "",
			description: "Command without entrypoint should work correctly",
		},
		{
			name: "regression test - entrypoint-only works",
			composeYAML: `
services:
  web:
    image: nginx:latest
    entrypoint: ["/custom-entrypoint.sh"]
`,
			serviceName: "web",
			expectedCmd: "/custom-entrypoint.sh",
			expectedArg: "",
			description: "Entrypoint-only should work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := ParseComposeFile(tt.composeYAML, "test-project", nil)
			assert.NoError(t, err)

			composition := &compositionsv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-composition",
					Namespace: "test-namespace",
				},
				Spec: compositionsv1.CompositionSpec{
					EnvVars: make(map[string]string),
				},
			}

			envData := &EnvironmentData{
				EnvVars:     make(map[string]string),
				Secrets:     make(map[string]string),
				ConfigFiles: make(map[string]string),
			}

			resources, err := ConvertComposeToK8s(project, composition, "test-namespace", envData, nil)
			assert.NoError(t, err)

			// Find the target service's StatefulSet
			var container *corev1.Container
			for _, sts := range resources.StatefulSets {
				if sts.Name == tt.serviceName && len(sts.Spec.Template.Spec.Containers) > 0 {
					container = &sts.Spec.Template.Spec.Containers[0]
					break
				}
			}

			assert.NotNil(t, container, "Should find container for service %s", tt.serviceName)

			// Check that Command is set correctly
			assert.NotEmpty(t, container.Command, "Command should not be empty - %s", tt.description)
			assert.Equal(t, tt.expectedCmd, container.Command[0],
				"First command element should match - %s", tt.description)

			// Check Args if expected
			if tt.expectedArg != "" {
				assert.NotEmpty(t, container.Args, "Args should not be empty when command is specified with entrypoint - %s", tt.description)
				assert.Equal(t, tt.expectedArg, container.Args[0],
					"First arg element should match - %s", tt.description)
			} else {
				assert.Empty(t, container.Args, "Args should be empty when no command with entrypoint - %s", tt.description)
			}
		})
	}
}
