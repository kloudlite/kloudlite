package functions

import (
	"encoding/json"
	"regexp"

	"k8s.io/apimachinery/pkg/types"
	"kloudlite.io/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func NN(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}

func K8sObjToYAML(obj client.Object) ([]byte, error) {
	return yaml.Marshal(obj)
}

func K8sObjToMap(obj client.Object) (map[string]any, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

var nameValidator *regexp.Regexp = regexp.MustCompile(constants.K8sNameValidatorRegex)

func IsValidK8sResourceName(name string) bool {
	return nameValidator.Match([]byte(name))
}
