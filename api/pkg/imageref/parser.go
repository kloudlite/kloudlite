package imageref

import (
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

// Reference represents a parsed OCI image reference with all components
type Reference struct {
	// Full is the complete image reference
	Full string

	// Registry is the registry hostname (e.g., "registry.example.com:5000")
	Registry string

	// Repository is the repository path (e.g., "username/myapp")
	Repository string

	// Tag is the image tag (e.g., "v1.0.0", defaults to "latest")
	Tag string

	// Digest is the image digest (e.g., "sha256:abc123...")
	Digest string
}

// Parse parses an image reference string into a Reference struct
// Supports both tagged and digested references:
// - "registry:5000/repo/path:tag"
// - "registry:5000/repo/path@sha256:..."
// - "docker.io/library/alpine:3.19"
// - "localhost:5000/myapp:latest"
// - "alpine" (uses Docker Hub defaults)
func Parse(ref string) (*Reference, error) {
	// Handle empty reference
	if ref == "" {
		return nil, fmt.Errorf("empty image reference")
	}

	// Parse using go-containerregistry library
	parsedRef, err := name.ParseReference(ref, name.StrictValidation)
	if err != nil {
		// Try with non-strict validation for backwards compatibility
		parsedRef, err = name.ParseReference(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to parse image reference %q: %w", ref, err)
		}
	}

	result := &Reference{
		Full:      ref,
		Registry:  parsedRef.Context().RegistryStr(),
		Repository: parsedRef.Context().RepositoryStr(),
	}

	// Extract tag if present
	switch t := parsedRef.(type) {
	case *name.Tag:
		result.Tag = t.TagStr()
	case *name.Digest:
		result.Digest = t.DigestStr()
	}

	return result, nil
}

// ParseWithDefaults parses an image reference with default values
// If tag is not specified, defaults to the provided defaultTag (or "latest" if empty)
func ParseWithDefaults(ref, defaultTag string) (*Reference, error) {
	parsed, err := Parse(ref)
	if err != nil {
		return nil, err
	}

	// Apply default tag if none present
	if parsed.Tag == "" && defaultTag != "" {
		parsed.Tag = defaultTag
	}

	return parsed, nil
}

// MustParse parses an image reference and panics on error
// Use only in initialization code where the reference is known to be valid
func MustParse(ref string) *Reference {
	parsed, err := Parse(ref)
	if err != nil {
		panic(fmt.Sprintf("failed to parse image reference %q: %v", ref, err))
	}
	return parsed
}

// Validate checks if an image reference string is valid
func Validate(ref string) error {
	_, err := Parse(ref)
	return err
}

// IsStandardRegistry checks if the image uses a standard public registry
// Returns true for Docker Hub, GHCR, GCR, ECR, ACR, etc.
func (r *Reference) IsStandardRegistry() bool {
	registry := strings.ToLower(r.Registry)

	// List of standard public registries
	standardRegistries := []string{
		"registry-1.docker.io",
		"docker.io",
		"docker.io/library",
		"ghcr.io",
		"gcr.io",
		"*.gcr.io",
		"*.amazonaws.com",
		"azurecr.io",
		"*.azurecr.io",
		"mcr.microsoft.com",
		"quay.io",
		"*.quay.io",
		"registry.access.redhat.com",
		"registry.redhat.io",
	}

	for _, std := range standardRegistries {
		if strings.HasPrefix(std, "*.") {
			// Wildcard match: check if registry ends with the suffix
			if strings.HasSuffix(registry, strings.TrimPrefix(std, "*.")) {
				return true
			}
		} else if registry == std {
			return true
		}
	}

	return false
}

// String returns the full image reference string
func (r *Reference) String() string {
	return r.Full
}

// WithTag returns a new Reference with the specified tag
func (r *Reference) WithTag(tag string) (*Reference, error) {
	newRef := &Reference{
		Registry:  r.Registry,
		Repository: r.Repository,
		Tag:       tag,
		Digest:    r.Digest,
	}

	// Build the full reference
	var refStr string
	if r.Digest != "" {
		refStr = fmt.Sprintf("%s/%s@%s", r.Registry, r.Repository, r.Digest)
	} else {
		refStr = fmt.Sprintf("%s/%s:%s", r.Registry, r.Repository, tag)
	}

	newRef.Full = refStr
	return newRef, nil
}

// WithDigest returns a new Reference with the specified digest
func (r *Reference) WithDigest(digest string) (*Reference, error) {
	newRef := &Reference{
		Registry:  r.Registry,
		Repository: r.Repository,
		Tag:       r.Tag,
		Digest:    digest,
	}

	// Build the full reference (digest takes precedence over tag)
	refStr := fmt.Sprintf("%s/%s@%s", r.Registry, r.Repository, digest)
	newRef.Full = refStr
	return newRef, nil
}

// WithoutTag returns the reference without a tag or digest
func (r *Reference) WithoutTag() *Reference {
	newRef := &Reference{
		Registry:  r.Registry,
		Repository: r.Repository,
		Tag:       "",
		Digest:    "",
	}

	// Build the full reference without tag/digest
	newRef.Full = fmt.Sprintf("%s/%s", r.Registry, r.Repository)
	return newRef
}

// ImageNameRegistry provides a centralized registry of known image names and their properties
type ImageNameRegistry struct {
	// knownImages maps image names to their metadata
	knownImages map[string]*ImageMetadata
}

// ImageMetadata contains metadata about a known image
type ImageMetadata struct {
	// Name is the canonical image name
	Name string

	// Registry is the default registry for this image
	Registry string

	// Repository is the full repository path
	Repository string

	// DefaultTag is the default tag to use
	DefaultTag string

	// Description describes this image
	Description string
}

// NewImageNameRegistry creates a new image name registry with default known images
func NewImageNameRegistry() *ImageNameRegistry {
	registry := &ImageNameRegistry{
		knownImages: make(map[string]*ImageMetadata),
	}

	// Register known images
	registry.Register(&ImageMetadata{
		Name:        "alpine",
		Registry:     "docker.io",
		Repository:   "library/alpine",
		DefaultTag:   "3.19",
		Description:  "Lightweight Linux distribution for containers",
	})

	registry.Register(&ImageMetadata{
		Name:        "ubuntu",
		Registry:     "docker.io",
		Repository:   "library/ubuntu",
		DefaultTag:   "22.04",
		Description:  "Ubuntu Linux distribution",
	})

	registry.Register(&ImageMetadata{
		Name:        "nginx",
		Registry:     "docker.io",
		Repository:   "library/nginx",
		DefaultTag:   "latest",
		Description:  "High-performance web server and reverse proxy",
	})

	registry.Register(&ImageMetadata{
		Name:        "bitnami/postgresql",
		Registry:     "docker.io",
		Repository:   "bitnami/postgresql",
		DefaultTag:   "15",
		Description:  "PostgreSQL database by Bitnami",
	})

	registry.Register(&ImageMetadata{
		Name:        "docker",
		Registry:     "docker.io",
		Repository:   "library/docker",
		DefaultTag:   "latest",
		Description:  "Official Docker image",
	})

	return registry
}

// Register adds a new image to the registry
func (r *ImageNameRegistry) Register(meta *ImageMetadata) {
	r.knownImages[meta.Name] = meta
}

// Lookup finds an image by name (short name like "alpine" or "nginx")
func (r *ImageNameRegistry) Lookup(name string) (*ImageMetadata, bool) {
	meta, ok := r.knownImages[name]
	return meta, ok
}

// Resolve resolves a short image name to a full image reference
// If the name is already a full reference, it's parsed as-is
// If it's a short name, the registry is looked up and the default tag is applied
func (r *ImageNameRegistry) Resolve(name string) (*Reference, error) {
	// Try to parse as a full reference first
	if strings.Contains(name, "/") || strings.Contains(name, ":") {
		return Parse(name)
	}

	// Look up in registry
	meta, ok := r.Lookup(name)
	if !ok {
		// Return error for unknown image names
		return nil, fmt.Errorf("unknown image name %q, use full reference (e.g., registry/image:tag)", name)
	}

	// Build full reference from metadata
	ref := fmt.Sprintf("%s/%s:%s", meta.Registry, meta.Repository, meta.DefaultTag)
	return Parse(ref)
}

// ResolveWithDefaultTag resolves a short image name with a custom default tag
func (r *ImageNameRegistry) ResolveWithDefaultTag(name, defaultTag string) (*Reference, error) {
	// Try to parse as a full reference first
	if strings.Contains(name, "/") || strings.Contains(name, ":") {
		return Parse(name)
	}

	// Look up in registry
	meta, ok := r.Lookup(name)
	if !ok {
		// Return error for unknown image names
		return nil, fmt.Errorf("unknown image name %q, use full reference (e.g., registry/image:tag)", name)
	}

	// Build full reference from metadata with custom tag
	ref := fmt.Sprintf("%s/%s:%s", meta.Registry, meta.Repository, defaultTag)
	return Parse(ref)
}

// Global default image name registry
var DefaultRegistry = NewImageNameRegistry()
