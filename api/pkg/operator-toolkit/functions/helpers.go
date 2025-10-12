package functions

import (
	"maps"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MapContains checks if parent map contains all key-value pairs from child map
func MapContains(parent, child map[string]string) bool {
	if parent == nil {
		return len(child) == 0
	}
	for k, v := range child {
		if parent[k] != v {
			return false
		}
	}
	return true
}

// MapEqual checks if two maps are equal
func MapEqual(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	return MapContains(m1, m2)
}

// MapKeys returns keys from a map
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapMerge merges two maps into a new one and returns it
func MapMerge[K comparable, V any](inputs ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range inputs {
		maps.Copy(result, m)
	}

	return result
}

func MapFilter[K comparable, V any](input map[K]V, filter func(k K, v V) bool) map[K]V {
	result := make(map[K]V, len(input))
	for k, v := range input {
		if filter(k, v) {
			result[k] = v
		}
	}

	return result
}

func AsOwner(r client.Object, controller ...bool) metav1.OwnerReference {
	ctrler := false
	if len(controller) > 0 {
		ctrler = controller[0]
	}
	return metav1.OwnerReference{
		APIVersion: r.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       r.GetObjectKind().GroupVersionKind().Kind,
		Name:       r.GetName(),
		UID:        r.GetUID(),
		Controller: &ctrler,
		// BlockOwnerDeletion: New(false),
		BlockOwnerDeletion: &ctrler,
	}
}

// ContainsFinalizers checks if object has all specified finalizers
func ContainsFinalizers(obj client.Object, finalizers ...string) bool {
	objFinalizers := obj.GetFinalizers()
	for _, f := range finalizers {
		found := false
		for _, of := range objFinalizers {
			if of == f {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
