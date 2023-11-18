package types

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`

	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	Generation        int64        `json:"generation" graphql:"noinput"`
	CreationTimestamp metav1.Time  `json:"creationTimestamp" graphql:"noinput"`
	DeletionTimestamp *metav1.Time `json:"deletionTimestamp,omitempty" graphql:"noinput"`
}

func MetadataToGraphqlFieldEntry(omitEmpty bool) string {
	required := ""
	if !omitEmpty {
		required = "!"
	}
	return fmt.Sprintf(`Metadata%s @goField(name: "objectMeta")`, required)
}

func MetadataToGraphqlInputEntry(omitEmpty bool) string {
	required := ""
	if !omitEmpty {
		required = "!"
	}
	return fmt.Sprintf(`MetadataIn%s`, required)
}

type TypeMeta struct {
	APIVersion string `json:"apiVersion" graphql:"noinput"`
	Kind       string `json:"kind" graphql:"noinput"`
}

//
// func TypeMetaToGraphqlFieldEntry(omitEmpty bool) string {
// 	required := ""
// 	if !omitEmpty {
// 		required = "!"
// 	}
// 	return fmt.Sprintf(`TypeMeta%s`, required)
// }
//
// func TypeMetaToGraphqlInputEntry(omitEmpty bool) string {
// 	required := ""
// 	if !omitEmpty {
// 		required = "!"
// 	}
// 	return fmt.Sprintf(`TypeMetaIn%s`, required)
// }
