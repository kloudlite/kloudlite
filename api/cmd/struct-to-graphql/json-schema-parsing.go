package main

import (
	"fmt"
	"sort"
	"strings"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
		return fmt.Sprintf("%s: %s!", k, t)
	}
	return fmt.Sprintf("%s: %s", k, t)
}

func navigateTree(tree *v1.JSONSchemaProps, name string, schemaMap map[string][]string, depth ...int) {
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

	fields := make([]string, 0, len(tree.Properties))

	for k, v := range tree.Properties {
		// fmt.Printf("[properties] %q type: %s\n", k, v.Type)

		if v.Type == "array" {
			if v.Items.Schema != nil && v.Items.Schema.Type == "object" {
				fields = append(fields, genFieldEntry(k, fmt.Sprintf("[%s]", typeName+genTypeName(k)), m[k]))
				// iVar += genFieldEntry(k, fmt.Sprintf("[%s]", typeName+genTypeName(k)+"In"), m[k])

				navigateTree(v.Items.Schema, typeName+genTypeName(k), schemaMap, currDepth+1)
				continue
			}
			fields = append(fields, genFieldEntry(k, fmt.Sprintf("[%s]", genTypeName(v.Items.Schema.Type)), m[k]))
			continue
		}

		if v.Type == "object" {
			if currDepth == 1 {
				if k == "metadata" {
					fields = append(fields, genFieldEntry(k, "Metadata! @goField(name: \"objectMeta\")", m[k]))
					continue
				}

				// if k == "status" {
				// 	fields = append(fields, genFieldEntry(k, "Status", m[k]))
				// 	// INFO: removed as status is never going to be set via GraphQL
				// 	// iVar += genFieldEntry(k, "StatusIn", m[k])
				// 	continue
				// }

				// if k == "overrides" {
				// 	tVar += genFieldEntry(k, "Overrides", m[k])
				// 	iVar += genFieldEntry(k, "OverridesIn", m[k])
				// 	continue
				// }

				// if !hasAddedSyncStatus {
				// 	// TODO: added a custom sync status for everything k8s related
				// 	tVar += genFieldEntry("syncStatus", "SyncStatus", false)
				// 	hasAddedSyncStatus = true
				// }
			}

			if len(v.Properties) == 0 {
				fields = append(fields, genFieldEntry(k, "Map", m[k]))
				continue
			}

			fields = append(fields, genFieldEntry(k, typeName+genTypeName(k), m[k]))
			navigateTree(&v, typeName+genTypeName(k), schemaMap, currDepth+1)
			continue
		}

		fields = append(fields, genFieldEntry(k, gqlTypeMap(v.Type), m[k]))
	}

	sort.Strings(fields)
	schemaMap[name] = fields
}

func Convert(schema *v1.JSONSchemaProps, name string, schemaMap map[string][]string) error {
	navigateTree(schema, name, schemaMap)
	return nil
}
