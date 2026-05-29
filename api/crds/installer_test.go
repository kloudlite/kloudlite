package crds

import (
	"os"
	"path/filepath"
	"testing"

	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

func TestLoadEmbeddedAssetsDoesNotIncludeStandaloneCompositionCRD(t *testing.T) {
	crds, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, crd := range crds {
		if crd.Name == "compositions.environments.kloudlite.io" {
			t.Fatalf("standalone Composition CRD should not be installed")
		}
	}
}

func TestLoadEmbeddedAssetsIncludesEnvironmentComposeContract(t *testing.T) {
	crds, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var environmentCRD *apixv1.CustomResourceDefinition
	for i := range crds {
		if crds[i].Name == "environments.environments.kloudlite.io" {
			environmentCRD = &crds[i]
			break
		}
	}
	if environmentCRD == nil {
		t.Fatal("expected embedded Environment CRD")
	}
	if environmentCRD.Spec.Scope != apixv1.NamespaceScoped {
		t.Fatalf("Environment CRD scope = %q, want %q", environmentCRD.Spec.Scope, apixv1.NamespaceScoped)
	}

	version := storageVersion(environmentCRD)
	if version == nil {
		t.Fatal("expected Environment CRD storage version")
	}
	if version.Subresources == nil || version.Subresources.Status == nil {
		t.Fatal("expected Environment CRD status subresource")
	}
	if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
		t.Fatal("expected Environment CRD OpenAPI v3 schema")
	}

	rootSchema := version.Schema.OpenAPIV3Schema
	specSchema := property(t, rootSchema, "spec")
	assertObjectSchema(t, specSchema, "spec")
	assertObjectSchema(t, property(t, specSchema, "compose"), "spec.compose")

	statusSchema := property(t, rootSchema, "status")
	assertObjectSchema(t, statusSchema, "status")
	assertObjectSchema(t, property(t, statusSchema, "composeStatus"), "status.composeStatus")
	conditionsSchema := property(t, statusSchema, "conditions")
	if conditionsSchema.Type != "array" {
		t.Fatalf("status.conditions schema type = %q, want array", conditionsSchema.Type)
	}
	if conditionsSchema.Items == nil || conditionsSchema.Items.Schema == nil {
		t.Fatal("expected status.conditions item schema")
	}
	conditionSchema := conditionsSchema.Items.Schema
	assertObjectSchema(t, conditionSchema, "status.conditions[]")
	for _, field := range []string{"lastTransitionTime", "message", "reason", "status", "type"} {
		property(t, conditionSchema, field)
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

func storageVersion(crd *apixv1.CustomResourceDefinition) *apixv1.CustomResourceDefinitionVersion {
	for i := range crd.Spec.Versions {
		if crd.Spec.Versions[i].Storage {
			return &crd.Spec.Versions[i]
		}
	}
	return nil
}

func property(t *testing.T, schema *apixv1.JSONSchemaProps, name string) *apixv1.JSONSchemaProps {
	t.Helper()
	prop, ok := schema.Properties[name]
	if !ok {
		t.Fatalf("expected schema property %q", name)
	}
	return &prop
}

func assertObjectSchema(t *testing.T, schema *apixv1.JSONSchemaProps, name string) {
	t.Helper()
	if schema.Type != "object" {
		t.Fatalf("%s schema type = %q, want object", name, schema.Type)
	}
}
