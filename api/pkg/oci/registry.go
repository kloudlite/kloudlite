package oci

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

const (
	// ConfigMediaType is the media type for snapshot image config
	ConfigMediaType = "application/vnd.kloudlite.snapshot.config.v1+json"
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
func (c *Client) Push(opts PushOptions) (*PushResult, error) {
	// Parse the reference
	ref, err := c.parseReference(opts.RegistryURL, opts.Repository, opts.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	// Start with empty image or existing parent layers
	var img v1.Image
	var existingLayers []v1.Layer

	if len(opts.ParentLayers) > 0 {
		// Fetch existing image to get parent layers
		existingImg, err := c.pullImage(ref)
		if err == nil {
			existingLayers, _ = existingImg.Layers()
		}
		// If pull fails, we'll just create new layers
	}

	// Create the new snapshot layer
	newLayer, err := CreateSnapshotLayer(opts.SnapshotPath, opts.ParentSnapshotPath, opts.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot layer: %w", err)
	}

	// Build the image
	if len(existingLayers) > 0 {
		// Append new layer to existing layers
		img = empty.Image
		for _, layer := range existingLayers {
			img, err = mutate.AppendLayers(img, layer)
			if err != nil {
				return nil, fmt.Errorf("failed to append existing layer: %w", err)
			}
		}
	} else {
		img = empty.Image
	}

	// Append the new layer
	img, err = mutate.AppendLayers(img, newLayer)
	if err != nil {
		return nil, fmt.Errorf("failed to append new layer: %w", err)
	}

	// Set custom config
	configBytes, err := json.Marshal(ImageConfig{
		ImageType: "kloudlite-snapshot-chain",
		Version:   "v1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	img, err = mutate.ConfigFile(img, &v1.ConfigFile{
		Config: v1.Config{
			Labels: map[string]string{
				"io.kloudlite.snapshot": "true",
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}

	// Override the config media type by creating a custom manifest
	img = &customMediaTypeImage{
		Image:       img,
		configBytes: configBytes,
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

	// Get layer digests
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

// Pull pulls a snapshot chain from the registry
func (c *Client) Pull(opts PullOptions) (*PullResult, error) {
	// Parse the reference
	ref, err := c.parseReference(opts.RegistryURL, opts.Repository, opts.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %w", err)
	}

	// Pull the image
	img, err := c.pullImage(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	// Get all layers
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("failed to get layers: %w", err)
	}

	result := &PullResult{
		Snapshots:     make([]SnapshotMetadata, 0, len(layers)),
		SnapshotPaths: make(map[string]string),
	}

	// Extract each layer in order
	for i, layer := range layers {
		metadata, snapshotPath, err := ExtractSnapshotLayer(layer, opts.TargetDir)
		if err != nil {
			return nil, fmt.Errorf("failed to extract layer %d: %w", i, err)
		}

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

	var opts []name.Option
	if c.insecure {
		opts = append(opts, name.Insecure)
	}

	return name.ParseReference(refStr, opts...)
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

// customMediaTypeImage wraps an image to override media types for OCI format
type customMediaTypeImage struct {
	v1.Image
	configBytes []byte
}

func (c *customMediaTypeImage) MediaType() (types.MediaType, error) {
	return types.OCIManifestSchema1, nil
}

func (c *customMediaTypeImage) RawConfigFile() ([]byte, error) {
	return c.configBytes, nil
}

// ConfigName returns the digest of our custom config bytes
func (c *customMediaTypeImage) ConfigName() (v1.Hash, error) {
	digest, _, err := v1.SHA256(bytes.NewReader(c.configBytes))
	if err != nil {
		return v1.Hash{}, err
	}
	return digest, nil
}

// Manifest returns an OCI-format manifest instead of Docker format
func (c *customMediaTypeImage) Manifest() (*v1.Manifest, error) {
	// Get the underlying manifest
	m, err := c.Image.Manifest()
	if err != nil {
		return nil, err
	}

	// Create a copy with OCI media types
	ociManifest := &v1.Manifest{
		SchemaVersion: 2,
		MediaType:     types.OCIManifestSchema1,
		Config: v1.Descriptor{
			MediaType: ConfigMediaType, // Use our custom config media type
			Size:      int64(len(c.configBytes)),
			Digest:    m.Config.Digest, // This will be recalculated
		},
		Layers: make([]v1.Descriptor, len(m.Layers)),
	}

	// Convert layer media types to OCI format
	for i, layer := range m.Layers {
		ociManifest.Layers[i] = v1.Descriptor{
			MediaType: types.OCILayer,
			Size:      layer.Size,
			Digest:    layer.Digest,
		}
	}

	// Recalculate config digest from actual config bytes
	configDigest, _, err := v1.SHA256(bytes.NewReader(c.configBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate config digest: %w", err)
	}
	ociManifest.Config.Digest = configDigest

	return ociManifest, nil
}

// RawManifest returns the raw bytes of the OCI manifest
func (c *customMediaTypeImage) RawManifest() ([]byte, error) {
	m, err := c.Manifest()
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}
