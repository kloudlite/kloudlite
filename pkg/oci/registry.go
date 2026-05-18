package oci

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os/exec"
	"slices"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// Client handles OCI registry operations for snapshots
type Client struct {
	insecure bool
}

// NewClient creates a new OCI registry client
func NewClient(insecure bool) *Client {
	return &Client{
		insecure: insecure,
	}
}

// Push pushes a snapshot to the registry
// Each image contains ONLY its own layer - parent is referenced via config labels
func (c *Client) Push(opts PushOptions) (*PushResult, error) {
	// Parse the reference
	ref, err := c.parseReference(opts.RegistryURL, opts.Repository, opts.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	// Create the new snapshot layer using btrfs send
	// If ParentSnapshotPath is provided, creates incremental stream
	newLayer, err := CreateSnapshotLayer(opts.SnapshotPath, opts.ParentSnapshotPath, opts.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot layer: %w", err)
	}

	// Build the image with ONLY this layer (no parent layer copying)
	img := empty.Image

	// Create config with labels including parent reference
	// Parent chain is resolved via labels during pull
	configFile := &v1.ConfigFile{
		Architecture: "amd64",
		OS:           "linux",
		Config: v1.Config{
			Labels: map[string]string{
				"io.kloudlite.snapshot":      "true",
				"io.kloudlite.image-type":    "kloudlite-snapshot",
				"io.kloudlite.version":       "v2",
				"io.kloudlite.parent-image":  opts.ParentImageRef, // Parent reference for pull resolution
				"io.kloudlite.snapshot-name": opts.Metadata.Name,  // Snapshot name for identification
			},
		},
		RootFS: v1.RootFS{
			Type: "layers",
		},
	}

	// Set the config file
	img, err = mutate.ConfigFile(img, configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}

	// Append ONLY the new layer (single layer per image)
	img, err = mutate.AppendLayers(img, newLayer)
	if err != nil {
		return nil, fmt.Errorf("failed to append layer: %w", err)
	}

	// Push to registry
	remoteOpts := c.remoteOptions()
	if err := remote.Write(ref, img, remoteOpts...); err != nil {
		return nil, fmt.Errorf("failed to push image: %w", err)
	}

	// Get the digest
	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("failed to get digest: %w", err)
	}

	// Get layer info
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to get layers: %w", err)
	}

	layerDigests := make([]string, len(layers))
	var totalSize int64
	for i, layer := range layers {
		d, err := layer.Digest()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer digest: %w", err)
		}
		layerDigests[i] = d.String()

		size, err := layer.Size()
		if err == nil {
			totalSize += size
		}
	}

	return &PushResult{
		ImageRef:       ref.String(),
		Digest:         digest.String(),
		LayerDigests:   layerDigests,
		CompressedSize: totalSize,
	}, nil
}

// Pull pulls a snapshot chain from the registry by resolving parent references
// Images in v2 format have single layer each with parent-image label
// Images in v1 format have all layers in one image (backwards compatible)
func (c *Client) Pull(opts PullOptions) (*PullResult, error) {
	result := &PullResult{
		Snapshots:     make([]SnapshotMetadata, 0),
		SnapshotPaths: make(map[string]string),
	}

	// Track created subvolumes for cleanup on failure
	var createdSubvolumes []string
	cleanup := func() {
		for _, path := range createdSubvolumes {
			deleteCmd := exec.Command("nsenter", "-t", "1", "-m", "--", "btrfs", "subvolume", "delete", path)
			_ = deleteCmd.Run() // Best effort cleanup
		}
	}

	// Build initial reference
	currentRef := fmt.Sprintf("%s/%s:%s", opts.RegistryURL, opts.Repository, opts.Tag)
	var imagesToProcess []string

	// 1. Walk up the parent chain collecting all image refs (leaf to root)
	for currentRef != "" {
		imagesToProcess = append(imagesToProcess, currentRef)

		ref, err := c.parseReferenceFromString(currentRef)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to parse reference %s: %w", currentRef, err)
		}

		img, err := c.pullImage(ref)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to pull image %s: %w", currentRef, err)
		}

		// Check version and get parent reference
		cfg, err := img.ConfigFile()
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to get config for %s: %w", currentRef, err)
		}

		version := cfg.Config.Labels["io.kloudlite.version"]
		if version == "v1" || version == "" {
			// v1 format: all layers in single image, no parent chain
			// Extract all layers from this single image
			layers, err := img.Layers()
			if err != nil {
				cleanup()
				return nil, fmt.Errorf("failed to get layers: %w", err)
			}

			for i, layer := range layers {
				metadata, snapshotPath, err := ExtractSnapshotLayer(layer, opts.TargetDir)
				if err != nil {
					cleanup()
					return nil, fmt.Errorf("failed to extract layer %d: %w", i, err)
				}
				createdSubvolumes = append(createdSubvolumes, snapshotPath)
				result.Snapshots = append(result.Snapshots, *metadata)
				result.SnapshotPaths[metadata.Name] = snapshotPath
			}
			return result, nil
		}

		// v2 format: check for parent reference
		parentRef := cfg.Config.Labels["io.kloudlite.parent-image"]
		if parentRef == "" {
			break // Reached root
		}
		currentRef = parentRef
	}

	// 2. Reverse to get root-to-leaf order (required for btrfs receive)
	slices.Reverse(imagesToProcess)

	// 3. Extract layers in order
	for _, imgRef := range imagesToProcess {
		ref, err := c.parseReferenceFromString(imgRef)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to parse reference %s: %w", imgRef, err)
		}

		img, err := c.pullImage(ref)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to pull image %s: %w", imgRef, err)
		}

		layers, err := img.Layers()
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to get layers for %s: %w", imgRef, err)
		}

		// v2 format: each image has exactly ONE layer
		if len(layers) != 1 {
			cleanup()
			return nil, fmt.Errorf("expected 1 layer in v2 image %s, got %d", imgRef, len(layers))
		}

		metadata, snapshotPath, err := ExtractSnapshotLayer(layers[0], opts.TargetDir)
		if err != nil {
			cleanup()
			return nil, fmt.Errorf("failed to extract layer from %s: %w", imgRef, err)
		}

		createdSubvolumes = append(createdSubvolumes, snapshotPath)
		result.Snapshots = append(result.Snapshots, *metadata)
		result.SnapshotPaths[metadata.Name] = snapshotPath
	}

	return result, nil
}

// GetImageLayers returns the layer digests for an existing image
func (c *Client) GetImageLayers(registryURL, repository, tag string) ([]string, error) {
	ref, err := c.parseReference(registryURL, repository, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	img, err := c.pullImage(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to get layers: %w", err)
	}

	digests := make([]string, len(layers))
	for i, layer := range layers {
		d, err := layer.Digest()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer digest: %w", err)
		}
		digests[i] = d.String()
	}

	return digests, nil
}

// parseReference creates a name.Reference from components
func (c *Client) parseReference(registryURL, repository, tag string) (name.Reference, error) {
	refStr := fmt.Sprintf("%s/%s:%s", registryURL, repository, tag)
	return c.parseReferenceFromString(refStr)
}

// parseReferenceFromString parses a full image reference string
func (c *Client) parseReferenceFromString(refStr string) (name.Reference, error) {
	var opts []name.Option
	if c.insecure {
		opts = append(opts, name.Insecure)
	}
	return name.ParseReference(refStr, opts...)
}

// Delete deletes an image from the registry
func (c *Client) Delete(imageRef string) error {
	ref, err := c.parseReferenceFromString(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse reference: %w", err)
	}

	if err := remote.Delete(ref, c.remoteOptions()...); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// Tag creates an additional tag for an existing image
// This is used to add user-provided tags in addition to snapshot name tags
func (c *Client) Tag(sourceRef, targetRef string) error {
	src, err := c.parseReferenceFromString(sourceRef)
	if err != nil {
		return fmt.Errorf("failed to parse source ref: %w", err)
	}

	img, err := c.pullImage(src)
	if err != nil {
		return fmt.Errorf("failed to pull source image: %w", err)
	}

	dst, err := c.parseReferenceFromString(targetRef)
	if err != nil {
		return fmt.Errorf("failed to parse target ref: %w", err)
	}

	return remote.Write(dst, img, c.remoteOptions()...)
}

// pullImage pulls an image from the registry
func (c *Client) pullImage(ref name.Reference) (v1.Image, error) {
	return remote.Image(ref, c.remoteOptions()...)
}

// remoteOptions returns the remote options for registry operations
func (c *Client) remoteOptions() []remote.Option {
	var opts []remote.Option

	if c.insecure {
		// Create a custom transport that skips TLS verification
		// and handles redirects to external hostnames
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		opts = append(opts, remote.WithTransport(transport))
	}

	return opts
}
