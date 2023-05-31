package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	_ "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// DONE: read omitempty from json tag
// DONE: read inline from json tag and flatten struct
// WONT FIX: lowercase first letter of field name
// DONE: unexport field name should be ignored, i.e. with `json:"-"`

var typeMap = map[reflect.Type]string{
	reflect.TypeOf(metav1.Time{}):      "Date",
	reflect.TypeOf(&metav1.Time{}):     "Date",
	reflect.TypeOf(time.Time{}):        "Date",
	reflect.TypeOf(&time.Time{}):       "Date",
	reflect.TypeOf(json.RawMessage{}):  "Any",
	reflect.TypeOf(&json.RawMessage{}): "Any",
}

var kindMap = map[reflect.Kind]string{
	reflect.Int:   "Int",
	reflect.Int8:  "Int",
	reflect.Int16: "Int",
	reflect.Int32: "Int",
	reflect.Int64: "Int",

	reflect.Float32: "Float",
	reflect.Float64: "Float",

	reflect.Bool:   "Boolean",
	reflect.String: "String",
}

func genFieldType(fieldType string, omitempty bool) string {
	if omitempty {
		return fieldType
	}
	return fieldType + "!"
}

type Project struct {
	repos.BaseEntity `json:",inline"`
	crdsv1.Project   `json:",inline" json-schema:"k8s://projects.crds.kloudlite.io"`
	AccountName      string       `json:"accountName"`
	ClusterName      string       `json:"clusterName"`
	SyncStatus       t.SyncStatus `json:"syncStatus"`
}

// type Person struct {
// 	ID             int    `json:"id"`
// 	Name           string `json:"name"`
// 	Age            int    `json:"age"`
// 	crdsv1.Project `json:",inline" json-schema:"k8s://projects.crds.kloudlite.io"`
// 	Email          string `json:"email"`
// }

func GenerateGraphQLSchema(name string, data interface{}, kCli k8s.ExtendedK8sClient) (map[string][]string, error) {
	ty := reflect.TypeOf(data)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	schemaMap := map[string][]string{}

	getGraphQLFields(ty, name, schemaMap, kCli)

	return schemaMap, nil
}

func getGraphQLFields(t reflect.Type, name string, dataMap map[string][]string, kCli k8s.ExtendedK8sClient) {
	var fields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			// unexported field
			continue
		}

		jsonSchemaTag := field.Tag.Get("json-schema")
		fieldName := field.Tag.Get("json")
		sp := strings.Split(fieldName, ",")

		omitempty := false
		inline := false

		if len(sp) >= 1 && sp[0] == "-" {
			// this field does not want to be included in the schema
			continue
		}

		// iterating from 1 as the first element is the field name, it would always be there
		for i := 1; i < len(sp); i++ {
			if sp[i] == "omitempty" {
				omitempty = true
			}
			if sp[i] == "inline" {
				inline = true
			}
		}

		if len(sp) > 1 {
			fieldName = sp[0]
		}

		if fieldName == "" {
			fieldName = field.Name
		}

		fieldType := ""

		hasSpecialCase := false

		if tf, ok := typeMap[field.Type]; ok {
			hasSpecialCase = true
			fieldType = tf
		}

		if !hasSpecialCase {
			switch field.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fieldType = genFieldType("Int", omitempty)
			case reflect.Float32, reflect.Float64:
				fieldType = genFieldType("Float", omitempty)
			case reflect.Bool:
				fieldType = genFieldType("Boolean", omitempty)
			case reflect.String:
				fieldType = genFieldType("String", omitempty)
			case reflect.Struct:
				fieldType = field.Name
				// fmt.Println("fieldType: ", fieldType)
				if jsonSchemaTag != "" {
					if strings.HasPrefix(jsonSchemaTag, "k8s://") {
						crdName := strings.Split(jsonSchemaTag, "k8s://")[1]
						jp, err := kCli.GetCRDJsonSchema(context.TODO(), crdName)
						if err != nil {
							panic(err)
						}

						if inline {
							dMap := map[string][]string{}
							Convert(jp, field.Type.Name(), dMap)
							for k, v := range dMap {
								if k == field.Type.Name() {
									fields = append(fields, v...)
									continue
								}
								sort.Strings(v)
								dataMap[k] = v
							}
							continue
						} else {
							fieldType = genFieldType("K8s"+field.Type.Name(), omitempty)
							Convert(jp, "K8s"+field.Type.Name(), dataMap)
							// continue
						}
					}
				} else {
					if inline {
						dMap := map[string][]string{}
						getGraphQLFields(field.Type, field.Name, dMap, kCli)
						for k, v := range dMap {
							if k == field.Name {
								fields = append(fields, v...)
								continue
							}
							sort.Strings(v)
							dataMap[k] = v
						}
						continue
					} else {
						fieldType = name + field.Name
						getGraphQLFields(field.Type, name+field.Name, dataMap, kCli)
					}
				}
			case reflect.Slice:
				{
					if field.Type.Elem().Kind() == reflect.Struct {
						// fieldType = fmt.Sprintf("[%s]", field.Name)
						if jsonSchemaTag != "" {
							if strings.HasPrefix(jsonSchemaTag, "k8s://") {
								fieldType = genFieldType(fmt.Sprintf("[%s]", "K8s"+field.Type.Name()), omitempty)
								crdName := strings.Split(jsonSchemaTag, "k8s://")[1]
								jp, err := kCli.GetCRDJsonSchema(context.TODO(), crdName)
								if err != nil {
									panic(err)
								}
								Convert(jp, "K8s"+field.Type.Name(), dataMap)
								// continue
							}
						} else {
							getGraphQLFields(field.Type.Elem(), name+field.Name, dataMap, kCli)
						}
					} else {
						fieldType = fmt.Sprintf("[%s]", kindMap[field.Type.Elem().Kind()])
					}
				}
			case reflect.Ptr:
				{
					if field.Type.Elem().Kind() == reflect.Struct {
						// fmt.Println("type: ", field.Type, field.Name)
						fieldType = field.Name
						getGraphQLFields(field.Type.Elem(), name+field.Name, dataMap, kCli)
					} else {
						fieldType = kindMap[field.Type.Elem().Kind()]
					}
				}
			case reflect.Map:
				{
					fieldType = "Map"
					if field.Type.Elem().Kind() == reflect.Struct {
						getGraphQLFields(field.Type.Elem(), name+field.Name, dataMap, kCli)
					}
				}
			default:
				{
					fmt.Printf("default: name: %v (%v), type: %v, kind: %v\n", fieldName, field.Name, field.Type, field.Type.Kind())
				}
				// fields = append(fields, fmt.Sprintf("%s { %s }", fieldName, strings.Join(nestedFields, " ")))
				// fields = append(fields, fieldName)
				// continue
			}
		}

		if fieldType != "" {
			fields = append(fields, fmt.Sprintf("%s: %s", fieldName, fieldType))
			continue
		}
		// fmt.Printf("hello: %v\n", fieldName)
	}

	sort.Strings(fields)
	dataMap[name] = fields

	// return fields
}

func WriteSchema(schema map[string][]string, writer io.Writer) error {
	for k, v := range schema {
		if _, err := fmt.Fprintf(writer, "\ntype %s {\n", k); err != nil {
			return err
		}
		for _, f := range v {
			if _, err := fmt.Fprintf(writer, "\t%s\n", f); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(writer, "}"); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	project := Project{}

	kCli, err := func() (k8s.ExtendedK8sClient, error) {
		return k8s.NewExtendedK8sClient(&rest.Config{Host: "localhost:8080"})
	}()
	if err != nil {
		panic(err)
	}

	schemaMap, err := GenerateGraphQLSchema("Project", project, kCli)
	if err != nil {
		fmt.Printf("Failed to generate GraphQL schema: %v", err)
		panic(err)
	}

	fmt.Printf("%#v\n", schemaMap)

	if err := WriteSchema(schemaMap, os.Stdout); err != nil {
		panic(err)
	}
}
