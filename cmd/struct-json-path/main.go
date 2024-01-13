package main

import (
	"fmt"
	"go/types"
	"io"
	"os"
	"strings"

	ast_parser "github.com/kloudlite/api/cmd/struct-json-path/ast-parser"

	fn "github.com/kloudlite/api/pkg/functions"
	flag "github.com/spf13/pflag"
)

type Tag struct {
	Value  string
	Params []string
}

func extractTag(tagstr string) map[string]Tag {
	tags := strings.Split(tagstr, " ")
	m := make(map[string]Tag, len(tags))
	for i := range tags {
		sp := strings.SplitN(tags[i], ":", 2)
		if len(sp) != 2 {
			continue
		}

		tagparams := strings.Split(sp[1][1:len(sp[1])-1], ",")

		tag := Tag{Value: tagparams[0]}
		if len(tagparams) > 1 {
			tag.Params = tagparams[1:]
		}

		m[sp[0]] = tag
	}

	return m
}

func flattenChildKeys(child map[string][]string) []string {
	keys := make([]string, 0, len(child)*2)
	m := make(map[string]struct{})
	for key, values := range child {
		if key != "" {
			keys = append(keys, key)
		}
		for _, name := range values {
			if _, ok := m[name]; ok {
				continue
			}
			nameref := fmt.Sprintf("%s.%s", key, name)
			if key == "" {
				nameref = name
			}
			keys = append(keys, nameref)
		}
	}
	return keys
}

func mergeUniqueKeys(keys1 []string, keys2 []string) []string {
	m := make(map[string]struct{})
	for _, v := range keys1 {
		m[v] = struct{}{}
	}

	for _, v := range keys2 {
		m[v] = struct{}{}
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func traverseStruct(s *types.Struct) map[string][]string {
	paths := make(map[string][]string)

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		fieldName := field.Name()

		tags := extractTag(s.Tag(i))
		v, ok := tags["json"]
		if ok {
			fieldName = v.Value
		}

		if strings.TrimSpace(v.Value) == "-" {
			// ignore this field, as this does not convert into a json field
			continue
		}

		if _, ok := tags["inline"]; ok {
			fieldName = ""
		}

		if paths[fieldName] == nil {
			paths[fieldName] = []string{}
		}
		if es, ok := field.Type().Underlying().(*types.Struct); ok {
			childKeys := flattenChildKeys(traverseStruct(es))
			paths[fieldName] = mergeUniqueKeys(childKeys, paths[fieldName])
		}
	}

	return paths
}

func variableGenerator(out io.Writer, value string) {
	out.Write([]byte(value + "\n"))
}

func main() {
	var structPaths []string
	var out string
	flag.StringSliceVar(&structPaths, "struct", nil, "--struct")
	flag.StringVar(&out, "out", "", "--out")
	flag.Parse()

	parser := ast_parser.NewASTParser()

	for i := range structPaths {
		sp := strings.SplitN(fn.StringReverse(structPaths[i]), ".", 2)
		if len(sp) != 2 {
			panic("invalid struct path")
		}

		structName, pkgPath := fn.StringReverse(sp[0]), fn.StringReverse(sp[1])

		// fmt.Printf("structName: %s, pkgPath: %s\n", structName, pkgPath)

		structObj, err := parser.FindStruct(pkgPath, structName)
		if err != nil {
			panic(err)
		}

		file, err := os.Create(fmt.Sprintf("%s/%s-jsonpath.go", out, structName))
		if err != nil {
			panic(err)
		}

		m := traverseStruct(structObj)

		for _, v := range flattenChildKeys(m) {
			key := structName + "." + v
			variableGenerator(file, key)
		}

		file.Close()
	}
}
