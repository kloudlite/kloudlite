package assets

import "testing"

func TestAllReturnsEmbeddedCRDs(t *testing.T) {
	crds, err := All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(crds) == 0 {
		t.Fatal("expected embedded CRDs")
	}

	foundUsers := false
	for _, crd := range crds {
		if crd.Name == "users.platform.kloudlite.io" {
			foundUsers = true
		}
	}
	if !foundUsers {
		t.Fatal("expected users.platform.kloudlite.io CRD")
	}
}
