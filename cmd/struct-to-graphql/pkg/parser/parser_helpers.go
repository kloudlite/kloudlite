package parser

import (
	"fmt"
	"reflect"
	"strings"

	apiExtensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func (f *Field) handleString() (fieldType string, inputType string, err error) {
	childType := f.ParentName + f.Name
	if f.Enum != nil {
		f.Parser.structs[f.StructName].Enums[childType] = f.Enum
		return toFieldType(childType, !f.OmitEmpty), toFieldType(childType, !f.OmitEmpty), err
	}

	return toFieldType("String", !f.OmitEmpty), toFieldType("String", !f.OmitEmpty), err
}

func (f *Field) handleStruct() (fieldType string, inputFieldType string, err error) {
	pkgPath := fixPackagePath(f.PkgPath)

	childType := genTypeName(f.ParentName + f.Name)
	if pkgPath != "" {
		childType = genTypeName(pkgPath + "_" + f.Type.Name())
	}

	structName := func() string {
		if pkgPath == "" {
			return f.StructName
		}
		return commonLabel
	}()

	if f.Uri != nil {
		jsonSchema, err := func() (*apiExtensionsV1.JSONSchemaProps, error) {
			if strings.HasPrefix(*f.Uri, "http://") || strings.HasPrefix(*f.Uri, "https://") {
				return f.Parser.schemaCli.GetHttpJsonSchema(*f.Uri)
			}

			if strings.HasPrefix(*f.Uri, "k8s://") {
				k8sCrdName := strings.Split(*f.Uri, "k8s://")[1]
				return f.Parser.schemaCli.GetK8sJsonSchema(k8sCrdName)
			}

			return nil, fmt.Errorf("unknown schema for schema uri %q", *f.Uri)
		}()

		if err != nil {
			panic(err)
		}

		if f.Parser.structs[structName] == nil {
			f.Parser.structs[structName] = newStruct()
		}

		if f.Inline {
			p2 := newParser(f.Parser.schemaCli)
			p2.structs[structName] = newStruct()
			if err := p2.GenerateFromJsonSchema(p2.structs[structName], childType, jsonSchema); err != nil {
				return "", "", err
			}

			fields2, inputFields2 := f.Parser.structs[structName].mergeParser(p2.structs[structName], childType)

			if !f.GraphqlTag.OnlyInput {
				*f.Fields = append(*f.Fields, fields2...)
			}
			if !f.GraphqlTag.NoInput {
				*f.InputFields = append(*f.InputFields, inputFields2...)
			}

			return "", "", err
		}

		fieldType = toFieldType(childType, !f.OmitEmpty)
		inputFieldType = toFieldType(childType+"In", !f.OmitEmpty)
		if err := f.Parser.GenerateFromJsonSchema(f.Parser.structs[structName], childType, jsonSchema); err != nil {
			return "", "", err
		}
		return fieldType, inputFieldType, err
	}

	p2 := newParser(f.Parser.schemaCli)

	p2.structs[structName] = newStruct()
	if err := p2.GenerateGraphQLSchema(structName, childType, f.Type); err != nil {
		return "", "", err
	}

	if f.Inline {
		fields2, inputFields2 := f.Parser.structs[f.StructName].mergeParser(p2.structs[structName], childType)
		if !f.GraphqlTag.OnlyInput {
			*f.Fields = append(*f.Fields, fields2...)
		}

		if !f.GraphqlTag.NoInput {
			*f.InputFields = append(*f.InputFields, inputFields2...)
		}

		return "", "", err
	}

	if !f.GraphqlTag.OnlyInput {
		fieldType = toFieldType(childType, !f.OmitEmpty)
	}
	if !f.GraphqlTag.NoInput {
		inputFieldType = toFieldType(childType+"In", !f.OmitEmpty)
	}

	for k, v := range p2.structs {
		if _, ok := f.Parser.structs[k]; !ok {
			f.Parser.structs[k] = newStruct()
		}

		if !f.GraphqlTag.OnlyInput {
			for k2, v2 := range v.Types {
				f.Parser.structs[k].Types[k2] = v2
			}
		}

		for k2, v2 := range v.Enums {
			f.Parser.structs[k].Enums[k2] = v2
		}

		if !f.GraphqlTag.NoInput {
			for k2, v2 := range v.Inputs {
				f.Parser.structs[k].Inputs[k2] = v2
			}
		}
	}

	return fieldType, inputFieldType, err
}

func (f *Field) handleSlice() (fieldType string, inputFieldType string, err error) {
	if f.Type.Elem().Kind() == reflect.Struct {
		f2 := Field{
			ParentName:  f.ParentName,
			Name:        f.Name,
			PkgPath:     f.Type.Elem().PkgPath(),
			Type:        f.Type.Elem(),
			StructName:  f.StructName,
			Fields:      f.Fields,
			InputFields: f.InputFields,
			Parser:      f.Parser,
			JsonTag: JsonTag{
				Value:     f.JsonTag.Value,
				OmitEmpty: false,
				Inline:    false,
			},
			GraphqlTag: f.GraphqlTag,
		}

		fieldType, inputFieldType, _ := f2.handleStruct()

		return toFieldType(fmt.Sprintf("[%s]", fieldType), !f.JsonTag.OmitEmpty), toFieldType(fmt.Sprintf("[%s]", inputFieldType), !f.JsonTag.OmitEmpty), err
	}

	if f.Type.Elem().Kind() == reflect.Ptr {
		f2 := Field{
			ParentName:  f.ParentName,
			Name:        f.Name,
			PkgPath:     f.Type.Elem().PkgPath(),
			Type:        f.Type.Elem(),
			StructName:  f.StructName,
			Fields:      f.Fields,
			InputFields: f.InputFields,
			Parser:      f.Parser,
			JsonTag: JsonTag{
				Value:     f.JsonTag.Value,
				OmitEmpty: true,
				Inline:    false,
			},
			GraphqlTag: f.GraphqlTag,
		}

		fieldType, inputFieldType, _ := f2.handlePtr()
		return toFieldType(fmt.Sprintf("[%s]", fieldType), !f.JsonTag.OmitEmpty), toFieldType(fmt.Sprintf("[%s]", inputFieldType), !f.JsonTag.OmitEmpty), err
	}

	fieldType = toFieldType(fmt.Sprintf("[%s]", toFieldType(kindMap[f.Type.Elem().Kind()], true)), !f.JsonTag.OmitEmpty)
	inputFieldType = toFieldType(fmt.Sprintf("[%s]", toFieldType(kindMap[f.Type.Elem().Kind()], true)), !f.JsonTag.OmitEmpty)
	return fieldType, inputFieldType, err
}

func (f *Field) handleMap() (fieldType string, inputFieldType string, err error) {
	if f.Type.Elem().Kind() == reflect.Struct {
		pkgPath := f.Type.Elem().PkgPath()

		f2 := Field{
			ParentName:  f.ParentName,
			Name:        f.Name,
			PkgPath:     pkgPath,
			Type:        f.Type.Elem(),
			Fields:      f.Fields,
			InputFields: f.InputFields,
			Parser:      f.Parser,
			JsonTag: JsonTag{
				Value:     f.JsonTag.Value,
				OmitEmpty: false,
				Inline:    false,
			},
			GraphqlTag: f.GraphqlTag,
		}
		if _, _, err := f2.handleStruct(); err != nil {
			return "", "", err
		}
	}

	return toFieldType("Map", !f.JsonTag.OmitEmpty), toFieldType("Map", !f.JsonTag.OmitEmpty), err
}

func (f *Field) handlePtr() (fieldType string, inputFieldType string, err error) {
	if f.Type.Elem().Kind() == reflect.Struct {
		pkgPath := f.Type.Elem().PkgPath()

		f2 := Field{
			ParentName:  f.ParentName,
			Name:        f.Name,
			PkgPath:     pkgPath,
			Type:        f.Type.Elem(),
			Fields:      f.Fields,
			InputFields: f.InputFields,
			Parser:      f.Parser,
			JsonTag: JsonTag{
				Value:     f.JsonTag.Value,
				OmitEmpty: true, // because it is a pointer type
				Inline:    false,
			},
			GraphqlTag: f.GraphqlTag,
		}

		if pkgPath == "" {
			f2.StructName = f.StructName
			return f2.handleStruct()
		}
		f2.StructName = commonLabel
		return f2.handleStruct()
	}

	return kindMap[f.Type.Elem().Kind()], kindMap[f.Type.Elem().Kind()], err
}
