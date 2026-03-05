package workspace

import (
	"testing"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	corev1 "k8s.io/api/core/v1"
)

// TestPortKey tests the portKey struct for correct equality checks
func TestPortKey(t *testing.T) {
	tests := []struct {
		name     string
		key1     portKey
		key2     portKey
		expected bool
	}{
		{
			name:     "identical keys",
			key1:     portKey{protocol: corev1.ProtocolTCP, port: 8080},
			key2:     portKey{protocol: corev1.ProtocolTCP, port: 8080},
			expected: true,
		},
		{
			name:     "different protocols",
			key1:     portKey{protocol: corev1.ProtocolTCP, port: 8080},
			key2:     portKey{protocol: corev1.ProtocolUDP, port: 8080},
			expected: false,
		},
		{
			name:     "different ports",
			key1:     portKey{protocol: corev1.ProtocolTCP, port: 8080},
			key2:     portKey{protocol: corev1.ProtocolTCP, port: 9090},
			expected: false,
		},
		{
			name:     "both different",
			key1:     portKey{protocol: corev1.ProtocolTCP, port: 8080},
			key2:     portKey{protocol: corev1.ProtocolUDP, port: 9090},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.key1 == tt.key2
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestBuildWorkspaceServiceName tests the service name builder
func TestBuildWorkspaceServiceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "myworkspace",
			expected: "ws-myworkspace",
		},
		{
			name:     "name with numbers",
			input:    "workspace123",
			expected: "ws-workspace123",
		},
		{
			name:     "name with hyphens",
			input:    "my-workspace",
			expected: "ws-my-workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildWorkspaceServiceName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestBuildHeadlessServiceName tests the headless service name builder
func TestBuildHeadlessServiceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "myworkspace",
			expected: "ws-myworkspace-headless",
		},
		{
			name:     "name with numbers",
			input:    "workspace123",
			expected: "ws-workspace123-headless",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildHeadlessServiceName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestBuildHashInput tests the hash input builder
func TestBuildHashInput(t *testing.T) {
	tests := []struct {
		name      string
		owner     string
		workspace string
		expected  string
	}{
		{
			name:      "simple values",
			owner:     "user1",
			workspace: "workspace1",
			expected:  "user1-workspace1",
		},
		{
			name:      "with hyphens",
			owner:     "user-1",
			workspace: "my-workspace",
			expected:  "user-1-my-workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildHashInput(tt.owner, tt.workspace)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestBuildHostname tests the hostname builder
func TestBuildHostname(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		hash      string
		subdomain string
		expected  string
	}{
		{
			name:      "vscode hostname",
			prefix:    "vscode",
			hash:      "a1b2c3d4",
			subdomain: "example.com",
			expected:  "vscode-a1b2c3d4.example.com",
		},
		{
			name:      "claude hostname",
			prefix:    "claude",
			hash:      "e5f6g7h8",
			subdomain: "khost.dev",
			expected:  "claude-e5f6g7h8.khost.dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildHostname(tt.prefix, tt.hash, tt.subdomain)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestBuildPortHostname tests the port hostname builder with validation
func TestBuildPortHostname(t *testing.T) {
	tests := []struct {
		name      string
		port      int32
		hash      string
		subdomain string
		expected  string
	}{
		{
			name:      "valid port 3000",
			port:      3000,
			hash:      "a1b2c3d4",
			subdomain: "example.com",
			expected:  "p3000-a1b2c3d4.example.com",
		},
		{
			name:      "valid port 8080",
			port:      8080,
			hash:      "e5f6g7h8",
			subdomain: "khost.dev",
			expected:  "p8080-e5f6g7h8.khost.dev",
		},
		{
			name:      "minimum valid port",
			port:      1,
			hash:      "a1b2c3d4",
			subdomain: "example.com",
			expected:  "p1-a1b2c3d4.example.com",
		},
		{
			name:      "maximum valid port",
			port:      65535,
			hash:      "a1b2c3d4",
			subdomain: "example.com",
			expected:  "p65535-a1b2c3d4.example.com",
		},
		{
			name:      "invalid port 0",
			port:      0,
			hash:      "a1b2c3d4",
			subdomain: "example.com",
			expected:  "p0-a1b2c3d4.example.com",
		},
		{
			name:      "invalid port negative",
			port:      -1,
			hash:      "a1b2c3d4",
			subdomain: "example.com",
			expected:  "p0-a1b2c3d4.example.com",
		},
		{
			name:      "invalid port too high",
			port:      65536,
			hash:      "a1b2c3d4",
			subdomain: "example.com",
			expected:  "p0-a1b2c3d4.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPortHostname(tt.port, tt.hash, tt.subdomain)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestGenerateHash tests hash generation
func TestGenerateHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple input",
			input:    "test-input",
			expected: "6f8b794f", // Precomputed SHA256 of "test-input"
		},
		{
			name:     "different input",
			input:    "different",
			expected: "5d5b09f6", // Precomputed SHA256 of "different"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateHash(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}

	// Ensure consistent hashing
	hash1 := generateHash("test")
	hash2 := generateHash("test")
	if hash1 != hash2 {
		t.Error("hash generation is not consistent")
	}

	// Ensure different inputs produce different hashes
	hash3 := generateHash("test2")
	if hash1 == hash3 {
		t.Error("different inputs produce same hash")
	}
}

// TestBuildServicePorts tests the service port builder with port key optimization
func TestBuildServicePorts(t *testing.T) {
	workspace := &workspacev1.Workspace{
		Spec: workspacev1.WorkspaceSpec{
			Expose: []workspacev1.ExposedPort{
				{Port: 3000},
				{Port: 8080},
				{Port: 22}, // Should be skipped as SSH is already defined
			},
		},
	}

	reconciler := &WorkspaceReconciler{}
	ports := reconciler.buildServicePorts(workspace)

	// Check default ports exist
	defaultPorts := []struct {
		name string
		port int32
	}{
		{"ssh", 22},
		{"code-server", 8080},
		{"ttyd", 7681},
		{"claude-ttyd", 7682},
		{"opencode-ttyd", 7683},
		{"codex-ttyd", 7684},
	}

	for _, dp := range defaultPorts {
		found := false
		for _, p := range ports {
			if p.Name == dp.name && p.Port == dp.port {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("default port %s:%d not found", dp.name, dp.port)
		}
	}

	// Check exposed ports
	found3000 := false
	found8080 := false
	for _, p := range ports {
		if p.Name == "exposed-3000" && p.Port == 3000 {
			found3000 = true
		}
		if p.Name == "exposed-8080" && p.Port == 8080 {
			found8080 = true
		}
	}

	if !found3000 {
		t.Error("exposed port 3000 not found")
	}
	if !found8080 {
		t.Error("exposed port 8080 not found")
	}

	// Ensure no duplicate ports exist
	portMap := make(map[portKey]bool)
	for _, p := range ports {
		key := portKey{protocol: p.Protocol, port: p.Port}
		if portMap[key] {
			t.Errorf("duplicate port found: %d", p.Port)
		}
		portMap[key] = true
	}
}
