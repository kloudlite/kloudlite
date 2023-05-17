package main

import (
	"bytes"
	"fmt"
	"strings"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func gqlTypeMap(jsonType string) string {
	switch jsonType {
	case "boolean":
		return "Boolean"
	case "integer":
		return "Int"
	case "object":
		return "Object"
	case "string":
		return "String"
	case "array":
		return "Array"
	default:
		return "Any"
	}
}

func genTypeName(n string) string {
	return strings.ToUpper(n[0:1]) + n[1:]
}

func genFieldEntry(k string, t string, required bool) string {
	if required {
		return fmt.Sprintf("\t%s: %s!\n", k, t)
	}
	return fmt.Sprintf("\t%s: %s\n", k, t)
}

func navigateTree(tree *v1.JSONSchemaProps, name string, schemas map[string]string, depth ...int) {
	currDepth := func() int {
		if len(depth) == 0 {
			return 1
		}
		return depth[0]
	}()

	m := map[string]bool{}
	for i := range tree.Required {
		m[tree.Required[i]] = true
	}

	typeName := genTypeName(name)

	var tVar, iVar string

	tVar = fmt.Sprintf("type %s @shareable {\n", typeName)
	iVar = fmt.Sprintf("input %sIn {\n", typeName)

	//fmt.Printf("%q type: %s\n", typeName, tree.Type)

	hasAddedSyncStatus := false

	for k, v := range tree.Properties {
		//fmt.Printf("[properties] %q type: %s\n", k, v.Type)

		if v.Type == "array" {
			if v.Items.Schema != nil && v.Items.Schema.Type == "object" {
				tVar += genFieldEntry(k, fmt.Sprintf("[%s]", typeName+genTypeName(k)), m[k])
				iVar += genFieldEntry(k, fmt.Sprintf("[%s]", typeName+genTypeName(k)+"In"), m[k])

				navigateTree(v.Items.Schema, typeName+genTypeName(k), schemas, currDepth+1)
				continue
			}
			tVar += genFieldEntry(k, fmt.Sprintf("[%s]", genTypeName(v.Items.Schema.Type)), m[k])
			iVar += genFieldEntry(k, fmt.Sprintf("[%s]", genTypeName(v.Items.Schema.Type)), m[k])
			//schemas[name] += fmt.Sprintf("\t%s: [%s]\n", k, genTypeName(v.Items.Schema.Type))
			continue
		}

		if v.Type == "object" {
			if currDepth == 1 {
				if k == "metadata" {
					tVar += genFieldEntry(k, "Metadata! @goField(name: \"objectMeta\")", m[k])
					iVar += genFieldEntry(k, "MetadataIn! @goField(name: \"objectMeta\")", m[k])
					continue
				}

				if k == "status" {
					tVar += genFieldEntry(k, "Status", m[k])
					// INFO: removed as status is never going to be set via GraphQL
					//iVar += genFieldEntry(k, "StatusIn", m[k])
					continue
				}

				if k == "overrides" {
					tVar += genFieldEntry(k, "Overrides", m[k])
					iVar += genFieldEntry(k, "OverridesIn", m[k])
					continue
				}

				if !hasAddedSyncStatus {
					// TODO: added a custom sync status for everything k8s related
					tVar += genFieldEntry("syncStatus", "SyncStatus", false)
					hasAddedSyncStatus = true
				}
			}

			if len(v.Properties) == 0 {
				tVar += genFieldEntry(k, "Map", m[k])
				iVar += genFieldEntry(k, "Map", m[k])
				continue
			}

			tVar += genFieldEntry(k, typeName+genTypeName(k), m[k])
			iVar += genFieldEntry(k, typeName+genTypeName(k)+"In", m[k])
			//schemas[name] += fmt.Sprintf("\t%s: %s!\n", k, typeName+genTypeName(k))
			navigateTree(&v, typeName+genTypeName(k), schemas, currDepth+1)
			continue
		}

		tVar += genFieldEntry(k, gqlTypeMap(v.Type), m[k])
		iVar += genFieldEntry(k, gqlTypeMap(v.Type), m[k])
	}
	tVar += "}"
	iVar += "}"

	schemas[name] = tVar
	schemas["input-"+name] = iVar
}

func ScalarTypes() ([]byte, error) {
	scalars := `
scalar Any
scalar Json
scalar Map
scalar Date
`

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

	status := `
type Status @shareable {
	isReady: Boolean!
	checks: Map
	displayVars: Json
}

type Check @shareable {
	status: Boolean
	message: String
	generation: Int
}
`

	overrides := `
type Patch @shareable {
	op: String!
	path: String!
	value: Any
}

type Overrides @shareable{
	applied: Boolean
	patches: [Patch!]
}

input PatchIn {
	op: String!
	path: String!
	value: Any
}

input OverridesIn{
	patches: [PatchIn!]
}
`

	syncStatus := `
enum SyncAction {
	APPLY
	DELETE
}

enum SyncState {
	IDLE
	IN_PROGRESS
	READY
	NOT_READY
}

type SyncStatus @shareable{
	syncScheduledAt: Date!
	lastSyncedAt: Date
	action: SyncAction!
	generation: Int!
	state: SyncState!
	error: String
}
`

	b := bytes.NewBuffer(nil)
	b.WriteString(scalars)
	b.WriteString(metadata)
	b.WriteString(status)
	b.WriteString(overrides)
	b.WriteString(syncStatus)

	return b.Bytes(), nil
}

func Directives() ([]byte, error) {
	directives := `
extend schema @link(url: "https://specs.apollo.dev/federation/v2.0", import: ["@key", "@shareable"])

directive @goField(
	forceResolver: Boolean
	name: String
) on INPUT_FIELD_DEFINITION | FIELD_DEFINITION
`
	return []byte(directives), nil
}

func Convert(schema *v1.JSONSchemaProps, name string) ([]byte, error) {
	schemas := map[string]string{}
	navigateTree(schema, name, schemas)
	b := bytes.NewBuffer(nil)
	//b.WriteString("scalar Any\n")
	for s := range schemas {
		b.WriteString(schemas[s])
		b.WriteString("\n\n")
	}
	return b.Bytes(), nil
}
