package functions

import (
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
