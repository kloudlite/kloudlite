package composition

import (
	"testing"

	composego "github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestValidatePorts(t *testing.T) {
	tests := []struct {
		name        string
		project     *composego.Project
		wantErr     bool
		errContains string
	}{
		{
			name: "valid ports",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Published: "8080", Protocol: "tcp"},
							{Target: 443, Published: "8443", Protocol: "tcp"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid udp port",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 53, Published: "53", Protocol: "udp"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid sctp port",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 3868, Protocol: "sctp"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid target port - zero",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 0},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "invalid target port (must be 1-65535)",
		},
		{
			name: "invalid target port - too high",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 70000},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "target port 70000 outside valid range (1-65535)",
		},
		{
			name: "invalid target port - negative",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 1}, // uint32 can't be negative, but we test edge case
						},
					},
				},
			},
			wantErr: false, // uint32 can't be negative, so this is valid
		},
		{
			name: "invalid published port - non-numeric",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Published: "abc"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "invalid published port - too high",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Published: "99999"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "published port 99999 outside valid range (1-65535)",
		},
		{
			name: "invalid published port - zero",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Published: "0"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "published port 0 outside valid range (1-65535)",
		},
		{
			name: "invalid protocol",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Protocol: "invalid"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "invalid protocol 'invalid' (must be tcp, udp, or sctp)",
		},
		{
			name: "invalid mode",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Mode: "invalid"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "invalid mode 'invalid' (must be host or ingress)",
		},
		{
			name: "invalid host_ip - contains space",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, HostIP: "invalid ip"},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "contains spaces",
		},
		{
			name: "valid host mode",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Mode: "host"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid ingress mode",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Mode: "ingress"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no ports defined",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple services with valid ports",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Published: "8080", Protocol: "tcp"},
						},
					},
					"db": {
						Ports: []composego.ServicePortConfig{
							{Target: 3306, Published: "3306", Protocol: "tcp"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple services - one invalid",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Published: "8080", Protocol: "tcp"},
						},
					},
					"db": {
						Ports: []composego.ServicePortConfig{
							{Target: 99999}, // Invalid port
						},
					},
				},
			},
			wantErr:     true,
			errContains: "db",
		},
		{
			name: "protocol case insensitive",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Protocol: "TCP"}, // uppercase
							{Target: 81, Protocol: "Udp"}, // mixed case
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mode case insensitive",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Mode: "HOST"}, // uppercase
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePorts(tt.project)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePorts_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		project     *composego.Project
		wantErr     bool
		errContains string
	}{
		{
			name: "minimum valid port",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 1},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "maximum valid port",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 65535},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "port just above maximum",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 65536},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "65536 outside valid range",
		},
		{
			name: "privileged port warning (below 1024)",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Published: "80"},
						},
					},
				},
			},
			wantErr: false, // Privileged ports are allowed, just a warning
		},
		{
			name: "published privileged port",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 8080, Published: "80"},
						},
					},
				},
			},
			wantErr: false, // Privileged ports are allowed
		},
		{
			name: "empty protocol",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Protocol: ""},
						},
					},
				},
			},
			wantErr: false, // Empty protocol is valid (defaults to tcp)
		},
		{
			name: "empty mode",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, Mode: ""},
						},
					},
				},
			},
			wantErr: false, // Empty mode is valid
		},
		{
			name: "valid host_ip",
			project: &composego.Project{
				Services: map[string]composego.ServiceConfig{
					"web": {
						Ports: []composego.ServicePortConfig{
							{Target: 80, HostIP: "127.0.0.1"},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePorts(tt.project)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseComposeFile_PortValidation(t *testing.T) {
	tests := []struct {
		name        string
		composeYAML string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid compose with ports",
			composeYAML: `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
      - "8443:443"
`,
			wantErr: false,
		},
		{
			name: "valid compose with udp port",
			composeYAML: `
version: '3.8'
services:
  dns:
    image: bind9:latest
    ports:
      - "53:53/udp"
`,
			wantErr: false,
		},
		{
			name: "invalid target port zero - caught by parser",
			composeYAML: `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:0"
`,
			wantErr:     true,
			errContains: "failed to parse compose file",
		},
		{
			name: "invalid port too high - caught by parser",
			composeYAML: `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "70000:80"
`,
			wantErr:     true,
			errContains: "failed to parse compose file",
		},
		{
			name: "invalid protocol - caught by parser",
			composeYAML: `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80/invalid"
`,
			wantErr:     true,
			errContains: "failed to parse compose file",
		},
		{
			name: "invalid protocol - caught by our validation",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        protocol: "invalid"
`,
			wantErr:     true,
			errContains: "port validation failed",
		},
		{
			name: "valid compose with no ports",
			composeYAML: `
version: '3.8'
services:
  web:
    image: nginx:latest
`,
			wantErr: false,
		},
		{
			name: "valid compose with multiple services",
			composeYAML: `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  db:
    image: mysql:latest
    ports:
      - "3306:3306"
`,
			wantErr: false,
		},
		{
			name: "invalid compose - one service bad port",
			composeYAML: `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  db:
    image: mysql:latest
    ports:
      - "99999:3306"
`,
			wantErr:     true,
			errContains: "failed to parse compose file",
		},
		{
			name: "empty compose",
			composeYAML: ``,
			wantErr:     true,
			errContains: "compose content is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := ParseComposeFile(tt.composeYAML, "test-project", nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
			}
		})
	}
}

func TestParseComposeFile_WithEnvironmentData(t *testing.T) {
	envData := &EnvironmentData{
		EnvVars: map[string]string{
			"DB_HOST": "localhost",
			"DB_PORT": "3306",
		},
		Secrets: map[string]string{
			"API_KEY": "secret123",
		},
	}

	composeYAML := `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      - DB_HOST=${DB_HOST}
      - DB_PORT=${DB_PORT}
      - API_KEY=${API_KEY}
`

	project, err := ParseComposeFile(composeYAML, "test-project", envData)
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Len(t, project.Services, 1)
	assert.Equal(t, "nginx:latest", project.Services["web"].Image)
}

func TestValidatePortFormat(t *testing.T) {
	tests := []struct {
		name    string
		portStr string
		wantErr bool
	}{
		{"valid port", "8080", false},
		{"valid port - max", "65535", false},
		{"valid port - min", "1", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"whitespace with space around", " 8080 ", false}, // Should be handled by parsing, not format validation
		{"contains letter", "80a0", true},
		{"contains special char", "808!", true},
		{"contains dash", "808-0", true},
		{"contains plus", "808+0", true},
		{"contains colon", "808:0", true},
		{"negative sign", "-8080", true},
		{"decimal point", "8080.0", true},
		{"leading zeros", "0080", false}, // Valid format, parsing will handle
		{"zero", "0", false},            // Valid format, range validation will catch
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePortFormat(tt.portStr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		wantErr  bool
	}{
		{"valid tcp", "tcp", false},
		{"valid tcp uppercase", "TCP", false},
		{"valid tcp mixed case", "Tcp", false},
		{"valid udp", "udp", false},
		{"valid udp uppercase", "UDP", false},
		{"valid sctp", "sctp", false},
		{"valid sctp uppercase", "SCTP", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"whitespace with tcp", "  tcp  ", false}, // Trimmed, should be valid
		{"invalid protocol", "http", true},
		{"invalid protocol", "tls", true},
		{"invalid protocol with space", "tcp udp", true},
		{"protocol with number", "tcp4", true},
		{"protocol with slash", "tcp/udp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProtocol(tt.protocol)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePortMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr bool
	}{
		{"valid host", "host", false},
		{"valid host uppercase", "HOST", false},
		{"valid host mixed case", "Host", false},
		{"valid ingress", "ingress", false},
		{"valid ingress uppercase", "INGRESS", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"whitespace with host", "  host  ", false}, // Trimmed, should be valid
		{"invalid mode", "bridge", true},
		{"invalid mode", "overlay", true},
		{"invalid mode with space", "host ingress", true},
		{"mode with number", "host1", true},
		{"mode with slash", "host/ingress", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePortMode(tt.mode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHostIP(t *testing.T) {
	tests := []struct {
		name    string
		hostIP  string
		wantErr bool
	}{
		{"valid IPv4", "127.0.0.1", false},
		{"valid IPv4 different", "192.168.1.1", false},
		{"valid IPv4 all zeros", "0.0.0.0", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"whitespace around", " 127.0.0.1 ", false}, // Trimmed, should be valid
		{"contains space", "127 0.0.1", true},
		{"contains multiple spaces", "127.0.0 .1", true},
		{"contains tab", "127.0.0.\t1", true},
		{"contains invalid char", "127.0.0.a", true},
		{"contains special char", "127.0.0.1!", true},
		{"contains pipe", "127.0.0.1|", true},
		{"contains comma", "127.0.0.1,", true},
		{"contains semicolon", "127.0.0.1;", true},
		{"contains bracket", "127.0.0.1]", true},
		{"just a dot", ".", true},    // Now caught: must contain at least one digit
		{"multiple dots", "...", true}, // Now caught: must contain at least one digit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHostIP(tt.hostIP)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePorts_InvalidPortFormats(t *testing.T) {
	tests := []struct {
		name        string
		composeYAML string
		wantErr     bool
		errContains string
	}{
		{
			name: "published port with letters",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "abc"
`,
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "published port with special characters",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "808!"
`,
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "published port with dash",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "808-0"
`,
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "published port with negative sign",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "-8080"
`,
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "published port with decimal",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080.5"
`,
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "published port with plus",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "+8080"
`,
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "protocol with whitespace",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        protocol: " tcp "
`,
			wantErr: false, // Should be trimmed and valid
		},
		{
			name: "mode with whitespace",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        mode: " host "
`,
			wantErr: false, // Should be trimmed and valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := ParseComposeFile(tt.composeYAML, "test-project", nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
			}
		})
	}
}

func TestValidatePorts_PortRangeRecommendations(t *testing.T) {
	// These tests verify that ports in different ranges are properly validated
	// while allowing users to use privileged ports (1-1023) if needed
	tests := []struct {
		name        string
		composeYAML string
		wantErr     bool
		description string
	}{
		{
			name: "published port in ephemeral range (1024-65535) - recommended",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080"
`,
			wantErr:     false,
			description: "Ports 1024-65535 are recommended for published ports",
		},
		{
			name: "published port at lower bound (1024)",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "1024"
`,
			wantErr:     false,
			description: "Port 1024 is the lowest non-privileged port",
		},
		{
			name: "published port at upper bound (65535)",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "65535"
`,
			wantErr:     false,
			description: "Port 65535 is the highest valid port",
		},
		{
			name: "published port just above privileged range (1023)",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "1023"
`,
			wantErr:     false,
			description: "Port 1023 is privileged but allowed (may require special permissions)",
		},
		{
			name: "published port in privileged range (80)",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "80"
`,
			wantErr:     false,
			description: "Port 80 is privileged but commonly needed for HTTP",
		},
		{
			name: "published port in privileged range (443)",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 443
        published: "443"
`,
			wantErr:     false,
			description: "Port 443 is privileged but commonly needed for HTTPS",
		},
		{
			name: "target port in privileged range (22)",
			composeYAML: `
services:
  ssh:
    image: ssh:latest
    ports:
      - target: 22
        published: "2222"
`,
			wantErr:     false,
			description: "Target port 22 is privileged but valid for SSH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := ParseComposeFile(tt.composeYAML, "test-project", nil)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
			}
		})
	}
}

func TestValidatePorts_MultiplePortsMixedValidation(t *testing.T) {
	tests := []struct {
		name        string
		composeYAML string
		wantErr     bool
		errContains string
	}{
		{
			name: "multiple valid ports with different protocols",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080"
        protocol: tcp
      - target: 53
        published: "5353"
        protocol: udp
      - target: 443
        published: "8443"
        protocol: tcp
`,
			wantErr: false,
		},
		{
			name: "multiple ports - one invalid protocol",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080"
        protocol: tcp
      - target: 53
        published: "5353"
        protocol: invalid
      - target: 443
        published: "8443"
        protocol: tcp
`,
			wantErr:     true,
			errContains: "invalid protocol",
		},
		{
			name: "multiple ports - one invalid published port format",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080"
      - target: 81
        published: "abc"
      - target: 82
        published: "8082"
`,
			wantErr:     true,
			errContains: "invalid published port format",
		},
		{
			name: "multiple ports - one out of range",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080"
      - target: 99999
      - target: 82
        published: "8082"
`,
			wantErr:     true,
			errContains: "outside valid range",
		},
		{
			name: "multiple services - each with different port configurations",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080"
        protocol: tcp
  db:
    image: mysql:latest
    ports:
      - target: 3306
        published: "3306"
        protocol: tcp
  cache:
    image: redis:latest
    ports:
      - target: 6379
        published: "6379"
        protocol: tcp
`,
			wantErr: false,
		},
		{
			name: "multiple services - one has invalid port",
			composeYAML: `
services:
  web:
    image: nginx:latest
    ports:
      - target: 80
        published: "8080"
        protocol: tcp
  db:
    image: mysql:latest
    ports:
      - target: 3306
        published: "3306"
        protocol: tcp
  cache:
    image: redis:latest
    ports:
      - target: 6379
        published: "invalid"
        protocol: tcp
`,
			wantErr:     true,
			errContains: "cache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := ParseComposeFile(tt.composeYAML, "test-project", nil)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
			}
		})
	}
}
