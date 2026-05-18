package assets

import (
	"fmt"
	"sort"

	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func All() ([]apixv1.CustomResourceDefinition, error) {
	names := make([]string, 0, len(crdYAMLDocuments))
	for name := range crdYAMLDocuments {
		names = append(names, name)
	}
	sort.Strings(names)

	crds := make([]apixv1.CustomResourceDefinition, 0, len(names))
	for _, name := range names {
		crd := apixv1.CustomResourceDefinition{}
		if err := yaml.Unmarshal([]byte(crdYAMLDocuments[name]), &crd); err != nil {
			return nil, fmt.Errorf("parse embedded CRD %q: %w", name, err)
		}
		if crd.Kind != "CustomResourceDefinition" || crd.Name == "" {
			return nil, fmt.Errorf("embedded CRD %q is not a CustomResourceDefinition", name)
		}
		crds = append(crds, crd)
	}
	if len(crds) == 0 {
		return nil, fmt.Errorf("no embedded CRDs found")
	}
	return crds, nil
}
