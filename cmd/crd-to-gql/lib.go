package main

import (
	"bytes"
	"fmt"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"strings"
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

func navigateTree(tree *v1.JSONSchemaProps, name string, schemas map[string]string) {
	m := map[string]bool{}
	for i := range tree.Required {
		m[tree.Required[i]] = true
	}

	typeName := genTypeName(name)

	var tVar, iVar string

	tVar = fmt.Sprintf("type %s {\n", typeName)
	iVar = fmt.Sprintf("input %sIn {\n", typeName)

	//fmt.Printf("%q type: %s\n", typeName, tree.Type)

	for k, v := range tree.Properties {
		//fmt.Printf("[properties] %q type: %s\n", k, v.Type)

		if v.Type == "array" {
			if v.Items.Schema != nil && v.Items.Schema.Type == "object" {
				tVar += genFieldEntry(k, typeName+genTypeName(k), m[k])
				iVar += genFieldEntry(k, typeName+genTypeName(k)+"In", m[k])

				navigateTree(v.Items.Schema, typeName+genTypeName(k), schemas)
				continue
			}
			tVar += genFieldEntry(k, fmt.Sprintf("[%s]", genTypeName(v.Items.Schema.Type)), m[k])
			iVar += genFieldEntry(k, fmt.Sprintf("[%s]", genTypeName(v.Items.Schema.Type)), m[k])
			//schemas[name] += fmt.Sprintf("\t%s: [%s]\n", k, genTypeName(v.Items.Schema.Type))
			continue
		}

		if v.Type == "object" {
			if k == "metadata" {
				tVar += genFieldEntry(k, "Metadata!", m[k])
				iVar += genFieldEntry(k, "MetadataIn!", m[k])
				continue
			}

			if k == "status" {
				tVar += genFieldEntry(k, "Status", m[k])
				// INFO: removed as status is never going to be set via GraphQL
				//iVar += genFieldEntry(k, "StatusIn", m[k])
				continue
			}

			if len(v.Properties) == 0 {
				tVar += genFieldEntry(k, "Any", m[k])
				iVar += genFieldEntry(k, "Any", m[k])
				continue
			}

			tVar += genFieldEntry(k, typeName+genTypeName(k), m[k])
			iVar += genFieldEntry(k, typeName+genTypeName(k)+"In", m[k])
			//schemas[name] += fmt.Sprintf("\t%s: %s!\n", k, typeName+genTypeName(k))
			navigateTree(&v, typeName+genTypeName(k), schemas)
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
`

	metadata := `
type Metadata {
	Name: String!
	Namespace: String
	Labels: Json
}

input MetadataIn {
	Name: String!
	Namespace: String
	Labels: Json
}
`

	status := `
type Status {
	isReady: Boolean!
	checks: [Check]
	displayVars: Json
}

type Check {
	Status: Boolean
	Message: String
	Generation: Int
}
`

	b := bytes.NewBuffer(nil)
	b.WriteString(scalars)
	b.WriteString(metadata)
	b.WriteString(status)

	return b.Bytes(), nil
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
