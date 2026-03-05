package snapshot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kloudlite/kloudlite/api/pkg/imageref"
	"go.uber.org/zap"
)

// DefaultSnapshotOperator implements SnapshotOperator using OCI registry API
type DefaultSnapshotOperator struct {
	Logger     *zap.Logger
	HTTPClient *http.Client
}

// NewDefaultSnapshotOperator creates a new DefaultSnapshotOperator
func NewDefaultSnapshotOperator(logger *zap.Logger) *DefaultSnapshotOperator {
	return &DefaultSnapshotOperator{
		Logger: logger,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DeleteFromRegistry removes the snapshot from OCI registry
// imageRef format: registry:port/repo/path:tag
func (o *DefaultSnapshotOperator) DeleteFromRegistry(ctx context.Context, imageRef string) error {
	if imageRef == "" {
		return nil
	}

	o.Logger.Info("Deleting snapshot from registry", zap.String("imageRef", imageRef))

	// Parse image reference using robust parser
	// Example: image-registry.kloudlite.svc.cluster.local:5000/snapshots/karthik/snapshot-name:snapshot-name
	ref, err := imageref.Parse(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}

	registry := ref.Registry
	repository := ref.Repository
	tag := ref.Tag

	// Validate that we have a tag (not a digest)
	if tag == "" && ref.Digest != "" {
		return fmt.Errorf("cannot delete by digest, use tag-based reference: %s", imageRef)
	}
	if tag == "" {
		return fmt.Errorf("no tag found in image reference: %s", imageRef)
	}

	// Step 1: Get the manifest digest for this tag
	manifestURL := fmt.Sprintf("http://%s/v2/%s/manifests/%s", registry, repository, tag)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, manifestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create manifest request: %w", err)
	}

	// Accept OCI manifest types
	req.Header.Set("Accept", "application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json")

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		o.Logger.Info("Manifest not found, already deleted", zap.String("imageRef", imageRef))
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get manifest: status %d, body: %s", resp.StatusCode, string(body))
	}

	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return fmt.Errorf("no digest returned for manifest")
	}

	o.Logger.Info("Got manifest digest", zap.String("digest", digest))

	// Step 2: Delete the manifest by digest
	deleteURL := fmt.Sprintf("http://%s/v2/%s/manifests/%s", registry, repository, digest)

	deleteReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	deleteResp, err := o.HTTPClient.Do(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to delete manifest: %w", err)
	}
	defer deleteResp.Body.Close()

	if deleteResp.StatusCode == http.StatusNotFound {
		o.Logger.Info("Manifest already deleted", zap.String("imageRef", imageRef))
		return nil
	}

	if deleteResp.StatusCode != http.StatusAccepted && deleteResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(deleteResp.Body)
		return fmt.Errorf("failed to delete manifest: status %d, body: %s", deleteResp.StatusCode, string(body))
	}

	o.Logger.Info("Successfully deleted snapshot from registry", zap.String("imageRef", imageRef))
	return nil
}
