package helm

import (
	"bytes"
	"encoding/json"

	"sigs.k8s.io/yaml"
)

func areHelmValuesEqual(releaseValues map[string]any, templateValues []byte) bool {
	b, err := json.Marshal(releaseValues)
	if err != nil {
		return false
	}

	tv, err := yaml.YAMLToJSON(templateValues)
	if err != nil {
		return false
	}

	if len(b) != len(tv) || !bytes.Equal(b, tv) {
		return false
	}
	return true
}

func AreHelmValuesEqual(releaseValues map[string]any, templateValues []byte) bool {
	return areHelmValuesEqual(releaseValues, templateValues)
}
