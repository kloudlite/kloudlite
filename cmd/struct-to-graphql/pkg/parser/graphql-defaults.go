package parser

import (
	"bytes"
	"strings"
)

func Directives() ([]byte, error) {
	directives := `extend schema @link(url: "https://specs.apollo.dev/federation/v2.0", import: ["@key", "@shareable"])

directive @goField(
	forceResolver: Boolean
	name: String
) on INPUT_FIELD_DEFINITION | FIELD_DEFINITION
`
	return []byte(directives), nil
}

func ScalarTypes() ([]byte, error) {
	scalars := `scalar Any
scalar Json
scalar Map
scalar Date
`

	b := bytes.NewBuffer(nil)
	b.WriteString(scalars)

	return b.Bytes(), nil
}

func KloudliteK8sTypes() ([]byte, error) {
	metadata := `
type Metadata @shareable {
	name: String!
	namespace: String
	labels: Json
	annotations: Json
	creationTimestamp: Date!
	deletionTimestamp: Date
	generation: Int!
}

input MetadataIn {
	name: String!
	namespace: String
	labels: Json
	annotations: Json
}
`

	// 	overrides := `
	// type Patch @shareable {
	// 	op: String!
	// 	path: String!
	// 	value: Any
	// }
	//
	// type Overrides @shareable{
	// 	applied: Boolean
	// 	patches: [Patch!]
	// }
	//
	// input PatchIn {
	// 	op: String!
	// 	path: String!
	// 	value: Any
	// }
	//
	// input OverridesIn {
	// 	patches: [PatchIn!]
	// }

	b := bytes.NewBuffer(nil)
	b.WriteString(strings.TrimSpace(metadata))
	b.WriteString("\n")
	return b.Bytes(), nil
}
