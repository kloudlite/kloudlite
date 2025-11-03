package functions

import (
	"maps"
	"slices"
	"strings"

	"github.com/gobuffalo/flect"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func MapHasKey[K comparable, T any](m map[K]T, k K) bool {
	_, ok := m[k]
	return ok
}

func IsOwner(obj client.Object, owner client.Object) bool {
	for _, ref := range obj.GetOwnerReferences() {
		if ref.Name == owner.GetName() &&
			ref.UID == owner.GetUID() &&
			ref.Kind == owner.GetObjectKind().GroupVersionKind().Kind &&
			ref.APIVersion == owner.GetObjectKind().GroupVersionKind().GroupVersion().String() {
			return true
		}
	}
	return false
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
		found := slices.Contains(objFinalizers, f)
		if !found {
			return false
		}
	}
	return true
}

func GVK(obj client.Object) metav1.GroupVersionKind {
	gvk := obj.GetObjectKind().GroupVersionKind()
	return metav1.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
}

func NN(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}

// RegularPlural is used to pluralize group of k8s CRDs from kind
// It is copied from https://github.com/kubernetes-sigs/kubebuilder/blob/afce6a0e8c2a6d5682be07bbe502e728dd619714/pkg/model/resource/utils.go#L71
func RegularPlural(singular string) string {
	return flect.Pluralize(strings.ToLower(singular))
}

func LabelValueEncoder(value string) string {
	return strings.ReplaceAll(value, "@", "-at-")
}

func LabelValueDecoder(value string) string {
	return strings.ReplaceAll(value, "-at-", "@")
}
