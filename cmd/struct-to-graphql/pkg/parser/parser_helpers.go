package parser

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

func (f *Field) handleString() (fieldType string, inputType string) {
	childType := f.ParentName + f.Name
	if f.Enum != nil {
		f.Parser.structs[f.StructName].Enums[childType] = f.Enum
		return toFieldType(childType, !f.OmitEmpty), toFieldType(childType, !f.OmitEmpty)
	}

	return toFieldType("String", !f.OmitEmpty), toFieldType("String", !f.OmitEmpty)
}

func (f *Field) handleStruct() (fieldType string, inputFieldType string) {
	pkgPath := fixPackagePath(f.PkgPath)

	childType := genTypeName(f.ParentName + f.Name)
	if pkgPath != "" {
		childType = genTypeName(pkgPath + "_" + f.Type.Name())
	}

	if f.Uri != nil {
		if strings.HasPrefix(*f.Uri, "k8s://") {
			k8sCrdName := strings.Split(*f.Uri, "k8s://")[1]
			jsonSchema, err := f.Parser.kCli.GetCRDJsonSchema(context.TODO(), k8sCrdName)
			if err != nil {
				panic(err)
			}

			structName := func() string {
				if pkgPath == "" {
					return f.StructName
				}
				return commonLabel
			}()

			if f.Inline {
				p2 := newParser(f.Parser.kCli)
				p2.structs[structName] = newStruct()
				p2.GenerateFromJsonSchema(p2.structs[structName], childType, jsonSchema)

				if f.Parser.structs[structName] == nil {
					f.Parser.structs[structName] = newStruct()
				}

				fields2, inputFields2 := f.Parser.structs[structName].mergeParser(p2.structs[structName], childType)

				*f.Fields = append(*f.Fields, fields2...)
				if !f.GraphqlTag.NoInput {
					*f.InputFields = append(*f.InputFields, inputFields2...)
				}

				return "", ""
			}

			fieldType = toFieldType(childType, !f.OmitEmpty)
			inputFieldType = toFieldType(childType+"In", !f.OmitEmpty)
			f.Parser.GenerateFromJsonSchema(f.Parser.structs[structName], childType, jsonSchema)
			return fieldType, inputFieldType
		}

		return "", ""
	}

	p2 := newParser(f.Parser.kCli)

	structName := func() string {
		if pkgPath == "" {
			return f.StructName
		}
		return commonLabel
	}()

	p2.structs[structName] = newStruct()
	p2.GenerateGraphQLSchema(structName, childType, f.Type)

	if f.Inline {
		fields2, inputFields2 := f.Parser.structs[f.StructName].mergeParser(p2.structs[structName], childType)
		*f.Fields = append(*f.Fields, fields2...)

		if !f.GraphqlTag.NoInput {
			*f.InputFields = append(*f.InputFields, inputFields2...)
		}

		return "", ""
	}

	fieldType = toFieldType(childType, !f.OmitEmpty)
	if !f.GraphqlTag.NoInput {
		inputFieldType = toFieldType(childType+"In", !f.OmitEmpty)
	}

	for k, v := range p2.structs {
		if _, ok := f.Parser.structs[k]; !ok {
			f.Parser.structs[k] = newStruct()
		}

		for k2, v2 := range v.Types {
			f.Parser.structs[k].Types[k2] = v2
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

	return fieldType, inputFieldType
}

func (f *Field) handleSlice() (fieldType string, inputFieldType string) {
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

		fieldType, inputFieldType := f2.handleStruct()

		return toFieldType(fmt.Sprintf("[%s]", fieldType), !f.JsonTag.OmitEmpty), toFieldType(fmt.Sprintf("[%s]", inputFieldType), !f.JsonTag.OmitEmpty)
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

		fieldType, inputFieldType := f2.handlePtr()
		return toFieldType(fmt.Sprintf("[%s]", fieldType), !f.JsonTag.OmitEmpty), toFieldType(fmt.Sprintf("[%s]", inputFieldType), !f.JsonTag.OmitEmpty)
	}

	fieldType = toFieldType(fmt.Sprintf("[%s]", toFieldType(kindMap[f.Type.Elem().Kind()], true)), !f.JsonTag.OmitEmpty)
	inputFieldType = toFieldType(fmt.Sprintf("[%s]", toFieldType(kindMap[f.Type.Elem().Kind()], true)), !f.JsonTag.OmitEmpty)
	return fieldType, inputFieldType
}

func (f *Field) handleMap() (fieldType string, inputFieldType string) {
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
		f2.handleStruct()
	}

	return toFieldType("Map", !f.JsonTag.OmitEmpty), toFieldType("Map", !f.JsonTag.OmitEmpty)
}

func (f *Field) handlePtr() (fieldType string, inputFieldType string) {
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

	return kindMap[f.Type.Elem().Kind()], kindMap[f.Type.Elem().Kind()]
}
