package oci

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

const (
	// DataFileName is the name of the snapshot data file in the layer tar
	DataFileName = "data.tar.gz"

	// MetadataFileName is the name of the metadata file in the layer tar
	MetadataFileName = "metadata.json"

	// LayerMediaType is the media type for snapshot layers
	LayerMediaType = "application/vnd.kloudlite.snapshot.layer.v1+tar"
)

// CreateSnapshotLayer creates an OCI layer from a snapshot
// The layer is a tar containing:
// - data.tar.gz: gzipped tar archive of snapshot directory
// - metadata.json: snapshot metadata
func CreateSnapshotLayer(snapshotPath, parentSnapshotPath string, metadata *SnapshotMetadata) (v1.Layer, error) {
	// Create a buffer for the tar
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	// Create tar archive of snapshot directory
	snapshotData, err := createTarArchive(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tar archive: %w", err)
	}

	// Add data.tar.gz to tar
	if err := tw.WriteHeader(&tar.Header{
		Name: DataFileName,
		Mode: 0644,
		Size: int64(len(snapshotData)),
	}); err != nil {
		return nil, fmt.Errorf("failed to write data header: %w", err)
	}
	if _, err := tw.Write(snapshotData); err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	// Add metadata.json to tar
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: MetadataFileName,
		Mode: 0644,
		Size: int64(len(metadataBytes)),
	}); err != nil {
		return nil, fmt.Errorf("failed to write metadata header: %w", err)
	}
	if _, err := tw.Write(metadataBytes); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar: %w", err)
	}

	// Create layer from tar bytes
	layer, err := tarball.LayerFromReader(&tarBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to create layer: %w", err)
	}

	return layer, nil
}

// ExtractSnapshotLayer extracts a snapshot layer to the target directory
// Returns the snapshot metadata and the path where the snapshot was extracted
func ExtractSnapshotLayer(layer v1.Layer, targetDir string) (*SnapshotMetadata, string, error) {
	// Get layer as tar reader
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get layer content: %w", err)
	}
	defer rc.Close()

	tr := tar.NewReader(rc)

	var metadata *SnapshotMetadata
	var snapshotData []byte

	// Extract files from tar
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to read tar: %w", err)
		}

		switch header.Name {
		case MetadataFileName:
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read metadata: %w", err)
			}
			metadata = &SnapshotMetadata{}
			if err := json.Unmarshal(data, metadata); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal metadata: %w", err)
			}

		case DataFileName:
			snapshotData, err = io.ReadAll(tr)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read snapshot data: %w", err)
			}
		}
	}

	if metadata == nil {
		return nil, "", fmt.Errorf("metadata.json not found in layer")
	}
	if snapshotData == nil {
		return nil, "", fmt.Errorf("data.tar.gz not found in layer")
	}

	// Extract tar archive
	snapshotPath, err := extractTarArchive(snapshotData, targetDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract snapshot: %w", err)
	}

	return metadata, snapshotPath, nil
}

// createTarArchive creates a gzipped tar archive of a directory
// Uses nsenter to run tar on the host
func createTarArchive(snapshotPath string) ([]byte, error) {
	// Use tar with gzip compression via nsenter
	// -C changes to parent dir, then archives the basename
	parentDir := filepath.Dir(snapshotPath)
	baseName := filepath.Base(snapshotPath)

	tarCmd := fmt.Sprintf("tar -czf - -C %s %s", parentDir, baseName)
	cmd := exec.Command("nsenter", "-t", "1", "-m", "--", "sh", "-c", tarCmd)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tar failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// extractTarArchive extracts a gzipped tar archive to target directory
// Uses nsenter to run tar on the host
func extractTarArchive(compressedData []byte, targetDir string) (string, error) {
	// Ensure target directory exists using nsenter
	mkdirCmd := exec.Command("nsenter", "-t", "1", "-m", "--", "mkdir", "-p", targetDir)
	if output, err := mkdirCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create target dir: %w, output: %s", err, string(output))
	}

	// Extract tar using nsenter
	cmd := exec.Command("nsenter", "-t", "1", "-m", "--", "tar", "-xzf", "-", "-C", targetDir)
	cmd.Stdin = bytes.NewReader(compressedData)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tar extract failed: %w, stderr: %s", err, stderr.String())
	}

	// Find extracted directory
	lsCmd := exec.Command("nsenter", "-t", "1", "-m", "--", "ls", targetDir)
	lsOutput, err := lsCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list target dir: %w", err)
	}

	var snapshotPath string
	entries := strings.Split(strings.TrimSpace(string(lsOutput)), "\n")
	for _, entry := range entries {
		if entry != "" {
			snapshotPath = filepath.Join(targetDir, entry)
		}
	}

	if snapshotPath == "" {
		return "", fmt.Errorf("no directory found after tar extract")
	}

	return snapshotPath, nil
}
