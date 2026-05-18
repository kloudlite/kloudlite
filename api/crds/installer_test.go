package crds

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesEmbeddedAssetsByDefault(t *testing.T) {
	crds, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(crds) == 0 {
		t.Fatal("expected embedded CRDs")
	}
}

func TestLoadKeepsDirectoryFallback(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.testing.kloudlite.io
spec:
  group: testing.kloudlite.io
  names:
    kind: Example
    plural: examples
  scope: Cluster
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
`)
	if err := os.WriteFile(filepath.Join(dir, "example.yaml"), data, 0644); err != nil {
		t.Fatal(err)
	}

	crds, err := Load(dir)
	if err != nil {
		t.Fatalf("Load(%q) error = %v", dir, err)
	}
	if len(crds) != 1 || crds[0].Name != "examples.testing.kloudlite.io" {
		t.Fatalf("unexpected CRDs: %#v", crds)
	}
}
