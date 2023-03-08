package functions

import "k8s.io/apimachinery/pkg/types"

func NN(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}
