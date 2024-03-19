package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/sanity-io/litter"
	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Parser interface {
	GenerateGraphQLSchema(structName string, name string, t reflect.Type, parentTag GraphqlTag) error
	LoadStruct(name string, data any) error

	PrintTypes(w io.Writer)
	PrintCommonTypes(w io.Writer)

	DebugSchema(w io.Writer)
	DumpSchema(dir string) error
	WithPagination(types []string)
}

type GraphqlType string

const (
	Type  GraphqlType = "type"
	Input GraphqlType = "input"
	Enum  GraphqlType = "enum"
)

var scalarMappings = map[reflect.Type]string{
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

	reflect.Uint:   "Int",
	reflect.Uint8:  "Int",
	reflect.Uint16: "Int",
	reflect.Uint32: "Int",
	reflect.Uint64: "Int",

	reflect.Float32: "Float",
	reflect.Float64: "Float",

	reflect.Bool:      "Boolean",
	reflect.Interface: "Any",

	reflect.String: "String",
}

type Struct struct {
	TypeDirectives map[string]struct{}
	Types          map[string][]string
	Inputs         map[string][]string
	Enums          map[string][]string
}

func newStruct() *Struct {
	return &Struct{
		Types:          map[string][]string{},
		Inputs:         map[string][]string{},
		Enums:          map[string][]string{},
		TypeDirectives: map[string]struct{}{},
	}
}

type Field struct {
	ParentName  string
	Name        string
	PkgPath     string
	Type        reflect.Type
	StructName  string
	Fields      *[]string
	InputFields *[]string

	Parser *parser

	JsonTag
	GraphqlTag
}

const (
	commonLabel = "common-types"
)

type parser struct {
	structs map[string]*Struct
	// schemaCli    k8s.ExtendedK8sClient
	schemaCli SchemaClient
}

type JsonTag struct {
	Value     string
	OmitEmpty bool
	Inline    bool
}

var sanitizers map[string]string = map[string]string{
	".": "__",
	"/": "___",
	"-": "____",
}

func SanitizePackagePath(pkgPath string) string {
	replacements := make([]string, 0, len(sanitizers)*2)
	for k, v := range sanitizers {
		replacements = append(replacements, k, v)
	}

	return strings.NewReplacer(replacements...).Replace(pkgPath)
}

func RestoreSanitizedPackagePath(sanitizedPath string) string {
	replacements := make([]string, 0, len(sanitizers)*2)
	for k, v := range sanitizers {
		replacements = append(replacements, v, k)
	}

	return strings.NewReplacer(replacements...).Replace(sanitizedPath)
}

func parseJsonTag(field reflect.StructField) JsonTag {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return JsonTag{Value: field.Name, OmitEmpty: false, Inline: false}
	}

	var jt JsonTag
	sp := strings.Split(jsonTag, ",")
	jt.Value = sp[0]

	if jt.Value == "" {
		jt.Value = field.Name
	}

	for i := 1; i < len(sp); i++ {
		if sp[i] == "omitempty" {
			jt.OmitEmpty = true
		}
		if sp[i] == "inline" {
			jt.Inline = true
		}
	}

	return jt
}

type schemaFormat string

func toFieldType(fieldType string, isRequired bool) string {
	if isRequired {
		return fieldType + "!"
	}
	return fieldType
}

func (s *Struct) mergeParser(other *Struct, overKey string) (fields []string, inputFields []string) {
	for k, v := range other.Types {
		if k == overKey {
			fields = append(fields, v...)
			continue
		}
		s.Types[k] = v
	}

	for k, v := range other.Inputs {
		if k == overKey+"In" {
			inputFields = append(inputFields, v...)
			continue
		}
		s.Inputs[k] = v
	}

	for k, v := range other.Enums {
		s.Enums[k] = v
	}

	return fields, inputFields
}

func (p *parser) GenerateGraphQLSchema(structName string, name string, t reflect.Type, parentTag GraphqlTag) error {
	var fields []string
	var inputFields []string

	if _, ok := p.structs[structName]; !ok {
		p.structs[structName] = newStruct()
	}

	typeDirectives := map[string]struct{}{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		jt := parseJsonTag(field)
		if jt.Value == "-" {
			continue
		}

		gt, err := parseGraphqlTag(field)
		if err != nil {
			return err
		}

		if gt.Ignore {
			continue
		}

		if parentTag.ChildrenRequired {
			jt.OmitEmpty = false
		}

		if parentTag.ChildrenOmitEmpty {
			jt.OmitEmpty = true
		}

		var fieldType string
		var inputFieldType string

		if scalar, ok := scalarMappings[field.Type]; ok {
			fieldType = toFieldType(scalar, !jt.OmitEmpty)
			inputFieldType = toFieldType(scalar, !jt.OmitEmpty)
		}

		if field.Type.Kind() != reflect.String {
			if v, ok := kindMap[field.Type.Kind()]; ok {
				fieldType = toFieldType(v, !jt.OmitEmpty)
				inputFieldType = toFieldType(v, !jt.OmitEmpty)
			}
		}

		f := Field{
			ParentName:  name,
			Name:        field.Name,
			PkgPath:     field.Type.PkgPath(),
			Type:        field.Type,
			StructName:  structName,
			Fields:      &fields,
			InputFields: &inputFields,
			Parser:      p,
			JsonTag:     jt,
			GraphqlTag:  gt,
		}

		if fieldType == "" {
			switch field.Type.Kind() {
			case reflect.String:
				{
					fieldType, inputFieldType, err = f.handleString()
					if err != nil {
						panic(err)
					}
				}
			case reflect.Struct:
				{
					fieldType, inputFieldType, err = f.handleStruct()
					if err != nil {
						panic(err)
					}
				}
			case reflect.Slice:
				{
					fieldType, inputFieldType, err = f.handleSlice()
					if err != nil {
						panic(err)
					}
				}
			case reflect.Ptr:
				{
					fieldType, inputFieldType, err = f.handlePtr()
					if err != nil {
						panic(err)
					}
				}
			case reflect.Map:
				{
					fieldType, inputFieldType, err = f.handleMap()
					if err != nil {
						panic(err)
					}
				}
			default:
				{
					fmt.Printf("default: name: %v (field-name: %v), type: %v, kind: %v\n", jt.Value, field.Name, field.Type, field.Type.Kind())
				}
			}
		}

		if fieldType != "" && !gt.OnlyInput {
			fields = append(fields, fmt.Sprintf("%s: %s", jt.Value, fieldType))
		}
		if inputFieldType != "" && !gt.NoInput {
			if gt.DefaultValue == nil {
				inputFields = append(inputFields, fmt.Sprintf("%s: %s", jt.Value, inputFieldType))
			} else {
				inputFields = append(inputFields, fmt.Sprintf("%s: %s = %v", jt.Value, inputFieldType, gt.DefaultValue))
			}
		}

		if f.GraphqlTag.ScalarType != nil && *f.GraphqlTag.ScalarType == "ID" {
			typeDirectives[`@key(fields: "id")`] = struct{}{}
		}
		typeDirectives[`@shareable`] = struct{}{}
	}

	if len(fields) > 0 {
		p.structs[structName].Types[name] = fields
	}

	if len(inputFields) > 0 {
		p.structs[structName].Inputs[name+"In"] = inputFields
	}

	if p.structs[structName].TypeDirectives == nil {
		p.structs[structName].TypeDirectives = map[string]struct{}{}
	}

	maps.Copy(p.structs[structName].TypeDirectives, typeDirectives)
	return nil
}

func (p *parser) LoadStruct(name string, data any) error {
	ty := reflect.TypeOf(data)
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	return p.GenerateGraphQLSchema(name, name, ty, GraphqlTag{})
}

func (s *Struct) WriteSchema(w io.Writer) {
	keys := make([]string, 0, len(s.Types))
	for k := range s.Types {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for i := range keys {
		// directives := maps.Keys(s.TypeDirectives)
		// sort.Strings(directives)
		//
		// directivesStr := fmt.Sprintf(" %s", strings.Join(directives, " "))

		// if strings.HasSuffix(keys[i], "PaginatedRecords") || strings.HasSuffix(keys[i], "Edge") {
		// directivesStr = ""
		// }

		io.WriteString(w, fmt.Sprintf("type %s @shareable {\n", keys[i]))

		sort.Slice(s.Types[keys[i]], func(p, q int) bool {
			return strings.ToLower(s.Types[keys[i]][p]) < strings.ToLower(s.Types[keys[i]][q])
		})
		io.WriteString(w, fmt.Sprintf("  %s\n", strings.Join(s.Types[keys[i]], "\n  ")))
		io.WriteString(w, "}\n\n")
	}

	keys = make([]string, 0, len(s.Inputs))
	for k := range s.Inputs {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for i := range keys {
		io.WriteString(w, fmt.Sprintf("input %s {\n", keys[i]))
		sort.Slice(s.Inputs[keys[i]], func(p, q int) bool {
			return strings.ToLower(s.Inputs[keys[i]][p]) < strings.ToLower(s.Inputs[keys[i]][q])
		})
		io.WriteString(w, fmt.Sprintf("  %s\n", strings.Join(s.Inputs[keys[i]], "\n  ")))
		io.WriteString(w, "}\n\n")
	}

	keys = make([]string, 0, len(s.Enums))
	for k := range s.Enums {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for i := range keys {
		io.WriteString(w, fmt.Sprintf("enum %s {\n", keys[i]))
		sort.Slice(s.Enums[keys[i]], func(p, q int) bool {
			return strings.ToLower(s.Enums[keys[i]][p]) < strings.ToLower(s.Enums[keys[i]][q])
		})
		io.WriteString(w, fmt.Sprintf("  %s\n", strings.Join(s.Enums[keys[i]], "\n  ")))
		io.WriteString(w, "}\n\n")
	}
}

func (p *parser) PrintTypes(w io.Writer) {
	keys := make([]string, 0, len(p.structs))
	for k := range p.structs {
		keys = append(keys, k)
	}

	for _, v := range keys {
		if v != commonLabel {
			p.structs[v].WriteSchema(w)
		}
	}
}

func (p *parser) PrintCommonTypes(w io.Writer) {
	if v, ok := p.structs[commonLabel]; ok {
		v.WriteSchema(w)
	}
}

func (p *parser) WithPagination(types []string) {
	for i := range types {
		k := types[i]
		v, ok := p.structs[types[i]]
		if !ok {
			continue
		}

		paginatedTypes := map[string][]string{
			fmt.Sprintf("%sPaginatedRecords", k): {
				"totalCount: Int!",
				fmt.Sprintf("edges: [%sEdge!]!", k),
				"pageInfo: PageInfo!",
			},
			fmt.Sprintf("%sEdge", k): {
				fmt.Sprintf("node: %v!", k),
				"cursor: String!",
			},
		}

		if _, ok := p.structs[commonLabel]; !ok {
			p.structs[commonLabel] = newStruct()
		}

		for i := range paginatedTypes {
			v.Types[i] = paginatedTypes[i]
		}
	}

	if len(types) > 0 {
		if p.structs[commonLabel] == nil {
			p.structs[commonLabel] = newStruct()
		}

		p.structs[commonLabel].Types["PageInfo"] = []string{
			"hasNextPage: Boolean",
			"hasPreviousPage: Boolean",
			"startCursor: String",
			"endCursor: String",
		}
	}
}

func (p *parser) DebugSchema(w io.Writer) {
	for k, v := range p.structs {
		io.WriteString(w, fmt.Sprintf("struct: %v\n", k))
		io.WriteString(w, litter.Sdump(v))
		io.WriteString(w, "\n")
	}
}

func (p *parser) DumpSchema(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(dir, 0o766); err != nil {
			return err
		}
	}

	for k, v := range p.structs {
		f, err := os.Create(filepath.Join(dir, strings.ToLower(k)+".graphqls"))
		if err != nil {
			return err
		}

		v.WriteSchema(f)
		f.Close()
	}
	return nil
}

type SchemaClient interface {
	GetK8sJsonSchema(name string) (*apiExtensionsV1.JSONSchemaProps, error)
	GetHttpJsonSchema(url string) (*apiExtensionsV1.JSONSchemaProps, error)
}

func newParser(schemaCli SchemaClient) *parser {
	return &parser{
		structs: map[string]*Struct{
			commonLabel: {
				Types:  map[string][]string{},
				Inputs: map[string][]string{},
				Enums:  map[string][]string{},
			},
		},
		schemaCli: schemaCli,
	}
}

func NewParser(cli SchemaClient) Parser {
	return newParser(cli)
}

func NewUnsafeParser(strucs map[string]*Struct, cli SchemaClient) Parser {
	return &parser{
		structs: strucs,
	}
}
