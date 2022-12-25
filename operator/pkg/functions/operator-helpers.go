package functions

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ContainsFinalizers(obj client.Object, finalizers ...string) bool {
	flist := obj.GetFinalizers()
	m := make(map[string]bool, len(flist))
	for i := range flist {
		m[flist[i]] = true
	}

	for i := range finalizers {
		_, ok := m[finalizers[i]]
		if !ok {
			return false
		}
	}
	return true
}

func ParseFromMap[T any, K any](m map[string]K) (*T, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var output T
	if err := json.Unmarshal(b, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func IntoMap(value any, targetMap any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &targetMap)
}

func ParseFromSecret[T any](secret *corev1.Secret) (*T, error) {
	x := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		x[k] = string(v)
	}

	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}

	var output T
	if err := json.Unmarshal(b, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
