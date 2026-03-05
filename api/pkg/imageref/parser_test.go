package imageref

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		ref           string
		wantRegistry  string
		wantRepo      string
		wantTag       string
		wantDigest    string
		expectError   bool
		errorContains string
	}{
		{
			name:         "full reference with tag",
			ref:          "registry.example.com:5000/myapp:v1.0.0",
			wantRegistry:  "registry.example.com:5000",
			wantRepo:      "myapp",
			wantTag:       "v1.0.0",
			wantDigest:    "",
			expectError:   false,
		},
		{
			name:         "full reference with digest",
			ref:          "registry.example.com:5000/myapp@sha256:abc123def456",
			wantRegistry:  "registry.example.com:5000",
			wantRepo:      "myapp",
			wantTag:       "",
			wantDigest:    "sha256:abc123def456",
			expectError:   false,
		},
		{
			name:         "Docker Hub official image",
			ref:          "docker.io/library/alpine:3.19",
			wantRegistry:  "docker.io",
			wantRepo:      "library/alpine",
			wantTag:       "3.19",
			wantDigest:    "",
			expectError:   false,
		},
		{
			name:         "localhost registry",
			ref:          "localhost:5000/myapp:latest",
			wantRegistry:  "localhost:5000",
			wantRepo:      "myapp",
			wantTag:       "latest",
			wantDigest:    "",
			expectError:   false,
		},
		{
			name:         "multi-level repository",
			ref:          "registry.example.com:5000/user/project/myapp:v1.0.0",
			wantRegistry:  "registry.example.com:5000",
			wantRepo:      "user/project/myapp",
			wantTag:       "v1.0.0",
			wantDigest:    "",
			expectError:   false,
		},
		{
			name:         "no tag (defaults to latest)",
			ref:          "registry.example.com:5000/myapp",
			wantRegistry:  "registry.example.com:5000",
			wantRepo:      "myapp",
			wantTag:       "latest",
			wantDigest:    "",
			expectError:   false,
		},
		{
			name:          "empty reference",
			ref:           "",
			expectError:    true,
			errorContains: "empty image reference",
		},
		{
			name:          "invalid reference format",
			ref:           "///invalid",
			expectError:    true,
			errorContains: "failed to parse image reference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.ref)

			if tt.expectError {
				if err == nil {
					t.Errorf("Parse() expected error but got nil")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Parse() error = %v, want error containing %q", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error = %v", err)
				return
			}

			if got.Registry != tt.wantRegistry {
				t.Errorf("Parse() registry = %v, want %v", got.Registry, tt.wantRegistry)
			}
			if got.Repository != tt.wantRepo {
				t.Errorf("Parse() repository = %v, want %v", got.Repository, tt.wantRepo)
			}
			if got.Tag != tt.wantTag {
				t.Errorf("Parse() tag = %v, want %v", got.Tag, tt.wantTag)
			}
			if got.Digest != tt.wantDigest {
				t.Errorf("Parse() digest = %v, want %v", got.Digest, tt.wantDigest)
			}
		})
	}
}

func TestParseWithDefaults(t *testing.T) {
	tests := []struct {
		name         string
		ref          string
		defaultTag   string
		wantTag      string
		expectError  bool
	}{
		{
			name:        "reference without tag uses default",
			ref:         "registry.example.com:5000/myapp",
			defaultTag:  "stable",
			wantTag:     "stable",
			expectError: false,
		},
		{
			name:        "reference with tag ignores default",
			ref:         "registry.example.com:5000/myapp:v1.0.0",
			defaultTag:  "stable",
			wantTag:     "v1.0.0",
			expectError: false,
		},
		{
			name:        "empty default tag preserves latest",
			ref:         "registry.example.com:5000/myapp",
			defaultTag:  "",
			wantTag:     "latest",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWithDefaults(tt.ref, tt.defaultTag)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseWithDefaults() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseWithDefaults() unexpected error = %v", err)
				return
			}

			if got.Tag != tt.wantTag {
				t.Errorf("ParseWithDefaults() tag = %v, want %v", got.Tag, tt.wantTag)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("valid reference", func(t *testing.T) {
		got := MustParse("docker.io/library/alpine:3.19")
		if got.Registry != "docker.io" {
			t.Errorf("MustParse() registry = %v, want docker.io", got.Registry)
		}
		if got.Tag != "3.19" {
			t.Errorf("MustParse() tag = %v, want 3.19", got.Tag)
		}
	})

	t.Run("panics on invalid reference", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustParse() expected panic but did not panic")
			}
		}()
		MustParse("")
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		wantErr bool
	}{
		{
			name:    "valid reference",
			ref:     "docker.io/library/alpine:3.19",
			wantErr: false,
		},
		{
			name:    "invalid reference",
			ref:     "///invalid",
			wantErr: true,
		},
		{
			name:    "empty reference",
			ref:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReferenceIsStandardRegistry(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected bool
	}{
		{
			name:     "Docker Hub",
			ref:      "docker.io/library/alpine:3.19",
			expected: true,
		},
		{
			name:     "GHCR",
			ref:      "ghcr.io/user/myapp:v1.0",
			expected: true,
		},
		{
			name:     "GCR",
			ref:      "gcr.io/myproject/myapp:latest",
			expected: true,
		},
		{
			name:     "ECR",
			ref:      "123456789.dkr.ecr.us-east-1.amazonaws.com/myapp:v1.0",
			expected: true,
		},
		{
			name:     "custom registry",
			ref:      "registry.example.com:5000/myapp:v1.0",
			expected: false,
		},
		{
			name:     "localhost registry",
			ref:      "localhost:5000/myapp:v1.0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := Parse(tt.ref)
			if err != nil {
				t.Fatalf("Parse() failed: %v", err)
			}

			if got := ref.IsStandardRegistry(); got != tt.expected {
				t.Errorf("Reference.IsStandardRegistry() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReferenceWithTag(t *testing.T) {
	ref, err := Parse("registry.example.com:5000/myapp:v1.0.0")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	newRef, err := ref.WithTag("v2.0.0")
	if err != nil {
		t.Fatalf("WithTag() failed: %v", err)
	}

	if newRef.Tag != "v2.0.0" {
		t.Errorf("WithTag() tag = %v, want v2.0.0", newRef.Tag)
	}
	if newRef.Registry != ref.Registry {
		t.Errorf("WithTag() registry changed, should remain same")
	}
	if newRef.Repository != ref.Repository {
		t.Errorf("WithTag() repository changed, should remain same")
	}
}

func TestReferenceWithDigest(t *testing.T) {
	ref, err := Parse("registry.example.com:5000/myapp:v1.0.0")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	newRef, err := ref.WithDigest("sha256:newhash123")
	if err != nil {
		t.Fatalf("WithDigest() failed: %v", err)
	}

	if newRef.Digest != "sha256:newhash123" {
		t.Errorf("WithDigest() digest = %v, want sha256:newhash123", newRef.Digest)
	}
}

func TestReferenceWithoutTag(t *testing.T) {
	ref, err := Parse("registry.example.com:5000/myapp:v1.0.0")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	newRef := ref.WithoutTag()

	if newRef.Tag != "" {
		t.Errorf("WithoutTag() tag = %v, want empty", newRef.Tag)
	}
	if newRef.Digest != "" {
		t.Errorf("WithoutTag() digest = %v, want empty", newRef.Digest)
	}
	if newRef.Registry != ref.Registry {
		t.Errorf("WithoutTag() registry changed, should remain same")
	}
}

func TestImageNameRegistry(t *testing.T) {
	registry := NewImageNameRegistry()

	t.Run("lookup known image", func(t *testing.T) {
		meta, ok := registry.Lookup("alpine")
		if !ok {
			t.Errorf("Lookup() returned false for known image 'alpine'")
		}
		if meta.Name != "alpine" {
			t.Errorf("Lookup() name = %v, want alpine", meta.Name)
		}
		if meta.Repository != "library/alpine" {
			t.Errorf("Lookup() repository = %v, want library/alpine", meta.Repository)
		}
	})

	t.Run("lookup unknown image", func(t *testing.T) {
		_, ok := registry.Lookup("unknown-image")
		if ok {
			t.Errorf("Lookup() returned true for unknown image")
		}
	})

	t.Run("resolve full reference", func(t *testing.T) {
		ref, err := registry.Resolve("registry.example.com:5000/myapp:v1.0")
		if err != nil {
			t.Errorf("Resolve() failed: %v", err)
		}
		if ref.Registry != "registry.example.com:5000" {
			t.Errorf("Resolve() registry = %v, want registry.example.com:5000", ref.Registry)
		}
	})

	t.Run("resolve short name", func(t *testing.T) {
		ref, err := registry.Resolve("alpine")
		if err != nil {
			t.Errorf("Resolve() failed: %v", err)
		}
		if ref.Registry != "docker.io" {
			t.Errorf("Resolve() registry = %v, want docker.io", ref.Registry)
		}
		if ref.Tag != "3.19" {
			t.Errorf("Resolve() tag = %v, want 3.19", ref.Tag)
		}
	})

	t.Run("resolve unknown short name", func(t *testing.T) {
		_, err := registry.Resolve("unknown-image")
		if err == nil {
			t.Errorf("Resolve() expected error for unknown image but got nil")
		}
	})

	t.Run("resolve with custom default tag", func(t *testing.T) {
		ref, err := registry.ResolveWithDefaultTag("alpine", "3.20")
		if err != nil {
			t.Errorf("ResolveWithDefaultTag() failed: %v", err)
		}
		if ref.Tag != "3.20" {
			t.Errorf("ResolveWithDefaultTag() tag = %v, want 3.20", ref.Tag)
		}
	})

	t.Run("register custom image", func(t *testing.T) {
		registry.Register(&ImageMetadata{
			Name:        "my-custom-image",
			Registry:     "registry.example.com",
			Repository:   "custom/my-image",
			DefaultTag:   "v1.0",
			Description:  "Custom test image",
		})

		meta, ok := registry.Lookup("my-custom-image")
		if !ok {
			t.Errorf("Lookup() returned false for custom image")
		}
		if meta.Name != "my-custom-image" {
			t.Errorf("Lookup() name = %v, want my-custom-image", meta.Name)
		}

		ref, err := registry.Resolve("my-custom-image")
		if err != nil {
			t.Errorf("Resolve() failed for custom image: %v", err)
		}
		if ref.Registry != "registry.example.com" {
			t.Errorf("Resolve() registry = %v, want registry.example.com", ref.Registry)
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || len(s) > len(substr) && s[:len(substr)] == substr || containsString(s[1:], substr))
}
