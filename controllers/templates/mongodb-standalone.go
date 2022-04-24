package templates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	msvcv1 "operators.kloudlite.io/apis/msvc/v1"
)

func NewMongoDBStandalone(namespace, name string) {
	_ = msvcv1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}
