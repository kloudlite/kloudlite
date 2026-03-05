package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kloudlite/kloudlite/api/pkg/imageref"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const snapshotMediaType = "application/vnd.kloudlite.snapshot.v1.tar+gzip"

// createTarGz creates a tar.gz archive of a directory
func createTarGz(srcDir string, destFile string) error {
	f, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %w", err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Linkname = link
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})
}

// extractTarGz extracts a tar.gz archive to a directory
func extractTarGz(srcFile string, destDir string) error {
	f, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		targetPath := filepath.Join(destDir, header.Name)

		// Ensure target path is within destDir (security check)
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid tar path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		case tar.TypeSymlink:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink: %w", err)
			}
		}
	}

	return nil
}

// parseImageRef extracts registry, repository and tag from an image reference
// e.g., "registry:5000/repo/path:tag" -> registry="registry:5000", repo="repo/path", tag="tag"
// This function uses robust imageref parser for proper OCI reference handling
// Returns an error if the image reference is invalid or uses digest instead of tag
func parseImageRef(imageRef string) (registry, repo, tag string, err error) {
	ref, err := imageref.Parse(imageRef)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse image reference %q: %w", imageRef, err)
	}

	// Handle both tagged and digested references
	// For digested references, we need a tag for ORAS operations
	// If digest is present, it's stored in ref.Digest
	if ref.Tag == "" {
		// If no tag is present but digest exists, this is a digest-only reference
		// ORAS operations require a tag, so this is an error condition
		if ref.Digest != "" {
			return "", "", "", fmt.Errorf("image reference %q uses digest but tag is required for ORAS operations", imageRef)
		}
		// No tag or digest - this is invalid
		return "", "", "", fmt.Errorf("image reference %q has no tag or digest", imageRef)
	}

	return ref.Registry, ref.Repository, ref.Tag, nil
}

// orasPushSnapshot pushes a directory as a snapshot to an OCI registry
func orasPushSnapshot(ctx context.Context, srcDir string, imageRef string, plainHTTP bool) error {
	registry, repo, tag, err := parseImageRef(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference: %w", err)
	}

	// Create repository reference
	repoRef, err := remote.NewRepository(registry + "/" + repo)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	repoRef.PlainHTTP = plainHTTP
	repoRef.Client = retry.DefaultClient

	// Create a temporary tar.gz file
	tmpFile := filepath.Join("/tmp", fmt.Sprintf("snapshot-%d.tar.gz", time.Now().UnixNano()))
	defer os.Remove(tmpFile)

	if err := createTarGz(srcDir, tmpFile); err != nil {
		return fmt.Errorf("failed to create tar.gz: %w", err)
	}

	// Create a file store
	fs, err := file.New("/tmp")
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	// Add file to store
	fileDesc, err := fs.Add(ctx, filepath.Base(tmpFile), snapshotMediaType, tmpFile)
	if err != nil {
		return fmt.Errorf("failed to add file to store: %w", err)
	}

	// Pack artifact using PackManifest
	manifestDesc, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, snapshotMediaType, oras.PackManifestOptions{
		Layers: []ocispec.Descriptor{fileDesc},
	})
	if err != nil {
		return fmt.Errorf("failed to pack artifact: %w", err)
	}

	// Tag manifest
	if err := fs.Tag(ctx, manifestDesc, tag); err != nil {
		return fmt.Errorf("failed to tag manifest: %w", err)
	}

	// Copy from file store to remote
	_, err = oras.Copy(ctx, fs, tag, repoRef, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to push to registry: %w", err)
	}

	return nil
}

// orasPullSnapshot pulls a snapshot from an OCI registry and extracts it to a directory
func orasPullSnapshot(ctx context.Context, imageRef string, destDir string, plainHTTP bool) error {
	registry, repo, tag, err := parseImageRef(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference: %w", err)
	}

	// Create repository reference
	repoRef, err := remote.NewRepository(registry + "/" + repo)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	repoRef.PlainHTTP = plainHTTP
	repoRef.Client = retry.DefaultClient

	// Create a temporary directory for pull
	tmpDir, err := os.MkdirTemp("/tmp", "oras-pull-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file store for pull
	fs, err := file.New(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to create file store: %w", err)
	}
	defer fs.Close()

	// Pull artifact
	_, err = oras.Copy(ctx, repoRef, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("failed to pull from registry: %w", err)
	}

	// Find and extract the tar.gz file
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read temp dir: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") {
			tarFile := filepath.Join(tmpDir, entry.Name())
			if err := extractTarGz(tarFile, destDir); err != nil {
				return fmt.Errorf("failed to extract tar.gz: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("no tar.gz file found in pulled artifact")
}
