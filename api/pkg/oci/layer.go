package oci

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

const (
	// DataFileName is the name of the btrfs stream file in the layer tar
	DataFileName = "data.btrfs"

	// MetadataFileName is the name of the metadata file in the layer tar
	MetadataFileName = "metadata.json"

	// LayerMediaType is the media type for snapshot layers
	LayerMediaType = "application/vnd.kloudlite.snapshot.layer.v1+tar"
)

// CreateSnapshotLayer creates an OCI layer from a btrfs snapshot
// The layer is a tar containing:
// - data.btrfs: btrfs send stream (full or incremental)
// - metadata.json: snapshot metadata
func CreateSnapshotLayer(snapshotPath, parentSnapshotPath string, metadata *SnapshotMetadata) (v1.Layer, error) {
	// Create a buffer for the tar
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	// Create btrfs send stream
	btrfsStream, err := createBtrfsStream(snapshotPath, parentSnapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create btrfs stream: %w", err)
	}

	// Add data.btrfs to tar
	if err := tw.WriteHeader(&tar.Header{
		Name: DataFileName,
		Mode: 0644,
		Size: int64(len(btrfsStream)),
	}); err != nil {
		return nil, fmt.Errorf("failed to write data header: %w", err)
	}
	if _, err := tw.Write(btrfsStream); err != nil {
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

// ExtractSnapshotLayer extracts a snapshot layer to the target directory using btrfs receive
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
	var btrfsStream []byte

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
			btrfsStream, err = io.ReadAll(tr)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read btrfs stream: %w", err)
			}
		}
	}

	if metadata == nil {
		return nil, "", fmt.Errorf("metadata.json not found in layer")
	}
	if btrfsStream == nil {
		return nil, "", fmt.Errorf("data.btrfs not found in layer")
	}

	// Receive btrfs stream
	snapshotPath, err := receiveBtrfsStream(btrfsStream, targetDir, metadata.Name)
	if err != nil {
		return nil, "", fmt.Errorf("failed to receive btrfs stream: %w", err)
	}

	return metadata, snapshotPath, nil
}

// createBtrfsStream creates a btrfs send stream from a snapshot
// If parentSnapshotPath is provided, creates an incremental stream
// Uses nsenter to run btrfs on the host filesystem
func createBtrfsStream(snapshotPath, parentSnapshotPath string) ([]byte, error) {
	var cmd *exec.Cmd

	if parentSnapshotPath != "" {
		// Incremental send with parent
		cmd = exec.Command("nsenter", "-t", "1", "-m", "--",
			"btrfs", "send", "-p", parentSnapshotPath, snapshotPath)
	} else {
		// Full send (no parent)
		cmd = exec.Command("nsenter", "-t", "1", "-m", "--",
			"btrfs", "send", snapshotPath)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("btrfs send failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// receiveBtrfsStream receives a btrfs stream to the target directory
// Uses nsenter to run btrfs on the host filesystem
// Returns the path where the snapshot was received
func receiveBtrfsStream(btrfsStream []byte, targetDir, snapshotName string) (string, error) {
	// Ensure target directory exists using nsenter
	mkdirCmd := exec.Command("nsenter", "-t", "1", "-m", "--", "mkdir", "-p", targetDir)
	if output, err := mkdirCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to create target dir: %w, output: %s", err, string(output))
	}

	// Receive btrfs stream
	cmd := exec.Command("nsenter", "-t", "1", "-m", "--", "btrfs", "receive", targetDir)
	cmd.Stdin = bytes.NewReader(btrfsStream)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("btrfs receive failed: %w, stderr: %s", err, stderr.String())
	}

	// The snapshot is received with its original name
	snapshotPath := filepath.Join(targetDir, snapshotName)

	// Make the received snapshot writable (btrfs receive creates read-only snapshots)
	// This is necessary so PVC provisioner can work with the data
	// Use -f to force even when received_uuid is set
	roCmd := exec.Command("nsenter", "-t", "1", "-m", "--", "btrfs", "property", "set", "-f", "-ts", snapshotPath, "ro", "false")
	if output, err := roCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to make snapshot writable: %w, output: %s", err, string(output))
	}

	return snapshotPath, nil
}
