package oci

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

const (
	// DataFileName is the name of the btrfs data file in the layer tar
	DataFileName = "data.btrfs.gz"

	// MetadataFileName is the name of the metadata file in the layer tar
	MetadataFileName = "metadata.json"

	// LayerMediaType is the media type for snapshot layers
	LayerMediaType = "application/vnd.kloudlite.snapshot.layer.v1+tar"
)

// CreateSnapshotLayer creates an OCI layer from a btrfs snapshot
// The layer is a tar containing:
// - data.btrfs.gz: gzipped btrfs send stream
// - metadata.json: snapshot metadata
func CreateSnapshotLayer(snapshotPath, parentSnapshotPath string, metadata *SnapshotMetadata) (v1.Layer, error) {
	// Create a buffer for the tar
	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	// Create btrfs send stream
	btrfsData, err := createBtrfsSendStream(snapshotPath, parentSnapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create btrfs send stream: %w", err)
	}

	// Add data.btrfs.gz to tar
	if err := tw.WriteHeader(&tar.Header{
		Name: DataFileName,
		Mode: 0644,
		Size: int64(len(btrfsData)),
	}); err != nil {
		return nil, fmt.Errorf("failed to write data header: %w", err)
	}
	if _, err := tw.Write(btrfsData); err != nil {
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

// createBtrfsSendStream creates a gzipped btrfs send stream
func createBtrfsSendStream(snapshotPath, parentSnapshotPath string) ([]byte, error) {
	var cmd *exec.Cmd

	if parentSnapshotPath != "" {
		// Incremental send
		cmd = exec.Command("btrfs", "send", "-p", parentSnapshotPath, snapshotPath)
	} else {
		// Full send
		cmd = exec.Command("btrfs", "send", snapshotPath)
	}

	// Capture stdout
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("btrfs send failed: %w, stderr: %s", err, stderr.String())
	}

	// Compress with gzip
	var compressed bytes.Buffer
	gzw := gzip.NewWriter(&compressed)
	if _, err := io.Copy(gzw, &stdout); err != nil {
		return nil, fmt.Errorf("failed to compress: %w", err)
	}
	if err := gzw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}

	return compressed.Bytes(), nil
}

// ExtractSnapshotLayer extracts a snapshot layer to the target directory
// Returns the snapshot metadata and the path where the snapshot was received
func ExtractSnapshotLayer(layer v1.Layer, targetDir string) (*SnapshotMetadata, string, error) {
	// Get layer as tar reader
	rc, err := layer.Uncompressed()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get layer content: %w", err)
	}
	defer rc.Close()

	tr := tar.NewReader(rc)

	var metadata *SnapshotMetadata
	var btrfsData []byte

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
			btrfsData, err = io.ReadAll(tr)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read btrfs data: %w", err)
			}
		}
	}

	if metadata == nil {
		return nil, "", fmt.Errorf("metadata.json not found in layer")
	}
	if btrfsData == nil {
		return nil, "", fmt.Errorf("data.btrfs.gz not found in layer")
	}

	// Decompress and receive btrfs stream
	snapshotPath, err := receiveBtrfsStream(btrfsData, targetDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to receive btrfs stream: %w", err)
	}

	return metadata, snapshotPath, nil
}

// receiveBtrfsStream decompresses and receives a btrfs stream
func receiveBtrfsStream(compressedData []byte, targetDir string) (string, error) {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target dir: %w", err)
	}

	// Decompress
	gzr, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Run btrfs receive
	cmd := exec.Command("btrfs", "receive", targetDir)
	cmd.Stdin = gzr

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("btrfs receive failed: %w, stderr: %s", err, stderr.String())
	}

	// The received snapshot will be in targetDir with the original name
	// We need to find it - btrfs receive creates a subvolume with the original name
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return "", fmt.Errorf("failed to read target dir: %w", err)
	}

	// Find the newest entry (the one we just received)
	var snapshotPath string
	for _, entry := range entries {
		path := filepath.Join(targetDir, entry.Name())
		snapshotPath = path
	}

	if snapshotPath == "" {
		return "", fmt.Errorf("no snapshot found after btrfs receive")
	}

	return snapshotPath, nil
}
