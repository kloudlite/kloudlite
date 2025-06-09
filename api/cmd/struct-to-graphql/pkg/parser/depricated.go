package parser

import (
	"encoding/json"
	"fmt"
	rApi "github.com/kloudlite/operator/pkg/operator"
	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

func (p *parser) NavigateTree(s *Struct, name string, tree *apiExtensionsV1.JSONSchemaProps, depth ...int) error {
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
	inputFields := make([]string, 0, len(tree.Properties))

	for k, v := range tree.Properties {
		if currDepth == 1 {
			if k == "apiVersion" || k == "kind" {
				// fields = append(fields, genFieldEntry(k, "String", m[k]))
				fields = append(fields, genFieldEntry(k, "String", true))
				// inputFields = append(inputFields, genFieldEntry(k, "String", m[k]))
				inputFields = append(inputFields, genFieldEntry(k, "String", m[k]))
				continue
			}
		}

		if v.Type == "array" {
			if v.Items.Schema != nil && v.Items.Schema.Type == "object" {
				fields = append(fields, genFieldEntry(k, fmt.Sprintf("[%s!]", typeName+genTypeName(k)), m[k]))
				inputFields = append(inputFields, genFieldEntry(k, fmt.Sprintf("[%sIn!]", typeName+genTypeName(k)), m[k]))
				if err := p.NavigateTree(s, typeName+genTypeName(k), v.Items.Schema, currDepth+1); err != nil {
					return err
				}
				continue
			}

			fields = append(fields, genFieldEntry(k, fmt.Sprintf("[%s!]", genTypeName(v.Items.Schema.Type)), m[k]))
			inputFields = append(inputFields, genFieldEntry(k, fmt.Sprintf("[%s!]", genTypeName(v.Items.Schema.Type)), m[k]))
			continue
		}

		if v.Type == "object" {
			if currDepth == 1 {
				// these types are common across all the types that will be generated
				if k == "metadata" {
					fields = append(fields, genFieldEntry(k, "Metadata! @goField(name: \"objectMeta\")", false))
					// fields = append(fields, genFieldEntry(k, "Metadata!", false))
					inputFields = append(inputFields, genFieldEntry(k, "MetadataIn!", false))

					metadata := struct {
						Name              string            `json:"name"`
						Namespace         string            `json:"namespace,omitempty"`
						Labels            map[string]string `json:"labels,omitempty"`
						Annotations       map[string]string `json:"annotations,omitempty"`
						Generation        int64             `json:"generation" graphql:"noinput"`
						CreationTimestamp metav1.Time       `json:"creationTimestamp" graphql:"noinput"`
						DeletionTimestamp *metav1.Time      `json:"deletionTimestamp,omitempty" graphql:"noinput"`
					}{}
					if err := p.GenerateGraphQLSchema(commonLabel, "Metadata", reflect.TypeOf(metadata), GraphqlTag{}); err != nil {
						return err
					}
					continue
				}

				if k == "status" {
					pkgPath := SanitizePackagePath("github.com/kloudlite/operator/pkg/operator")

					gType := genTypeName(pkgPath + "_" + "Status")

					fields = append(fields, genFieldEntry(k, gType, m[k]))

					p2 := newParser(p.schemaCli)
					if err := p2.GenerateGraphQLSchema(commonLabel, gType, reflect.TypeOf(rApi.Status{}), GraphqlTag{}); err != nil {
						return err
					}

					for _, v := range p2.structs {
						for k, v2 := range v.Types {
							p.structs[commonLabel].Types[k] = v2
						}
						for k, v2 := range v.Enums {
							p.structs[commonLabel].Enums[k] = v2
						}
					}

					continue
				}
			}

			if len(v.Properties) == 0 {
				fields = append(fields, genFieldEntry(k, "Map", m[k]))
				inputFields = append(inputFields, genFieldEntry(k, "Map", m[k]))
				continue
			}

			fields = append(fields, genFieldEntry(k, typeName+genTypeName(k), m[k]))
			inputFields = append(inputFields, genFieldEntry(k, typeName+genTypeName(k)+"In", m[k]))
			if err := p.NavigateTree(s, typeName+genTypeName(k), &v, currDepth+1); err != nil {
				return err
			}
			continue
		}

		if v.Type == "string" {
			if len(v.Enum) > 0 {
				fqtn := typeName + genTypeName(k)
				fields = append(fields, genFieldEntry(k, fqtn, m[k]))
				inputFields = append(inputFields, genFieldEntry(k, fqtn, m[k]))

				enums := make([]string, len(v.Enum))
				for i := range v.Enum {
					vjson, _ := v.Enum[i].MarshalJSON()
					var v string
					if err := json.Unmarshal(vjson, &v); err != nil {
						return nil
					}
					enums[i] = v
				}
				s.Enums[fqtn] = enums
				continue
			}
		}

		fields = append(fields, genFieldEntry(k, gqlTypeMap(v.Type), m[k]))
		inputFields = append(inputFields, genFieldEntry(k, gqlTypeMap(v.Type), m[k]))
	}

	s.Types[typeName] = fields
	s.Inputs[typeName+"In"] = inputFields
	return nil
}

func (p *parser) GenerateFromJsonSchema(s *Struct, name string, schema *apiExtensionsV1.JSONSchemaProps) error {
	return p.NavigateTree(s, name, schema)
}
