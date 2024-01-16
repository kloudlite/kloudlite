package main

import (
	"bytes"
	"fmt"
	"go/format"
	"go/types"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	ast_parser "github.com/kloudlite/api/cmd/struct-json-path/ast-parser"

	fn "github.com/kloudlite/api/pkg/functions"
	flag "github.com/spf13/pflag"
)

const (
	StructJsonpathKey     = "struct-json-path"
	IgnoreTagValue        = "ignore"
	IgnoreNestingTagValue = "ignore-nesting"
)

type Tag struct {
	Value  string
	Params map[string]struct{}
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

		tag := Tag{Value: tagparams[0], Params: map[string]struct{}{}}
		for i := 1; i < len(tagparams); i++ {
			tag.Params[tagparams[i]] = struct{}{}
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

func extractEmbeddedStruct(field *types.Var) (*types.Struct, bool) {
	if es, ok := field.Type().Underlying().(*types.Struct); ok {
		return es, true
	}

	if tp, ok := field.Type().Underlying().(*types.Pointer); ok {
		if es, ok := tp.Elem().Underlying().(*types.Struct); ok {
			return es, true
		}
	}

	return nil, false
}

func getFieldPackagePath(field *types.Var) (string, bool) {
	if origin, ok := field.Origin().Type().(*types.Named); ok {
		pkgName := origin.Obj().Pkg().Path()
		typeName := origin.Obj().Name()

		pkgPath := fmt.Sprintf("%s.%s", pkgName, typeName)
		if pkgName == "" {
			pkgPath = typeName
		}

		return pkgPath, true
	}
	return "", false
}

func traverseStruct(s *types.Struct, ignoreNestingForPkgs map[string]struct{}) map[string][]string {
	paths := make(map[string][]string)

	ignoredNesting := make(map[string]struct{}, len(ignoreNestingForPkgs))
	for k, v := range ignoreNestingForPkgs {
		ignoredNesting[k] = v
	}

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		fieldName := field.Name()

		tags := extractTag(s.Tag(i))
		jsonTag, ok := tags["json"]
		if ok {
			fieldName = jsonTag.Value
		}

		if strings.TrimSpace(jsonTag.Value) == "-" {
			// ignore this field, as this does not convert into a json field
			continue
		}

		if _, ok := jsonTag.Params["inline"]; ok {
			fieldName = ""
		}

		if structToJsonpathTag, ok := tags[StructJsonpathKey]; ok {
			if _, ignore := structToJsonpathTag.Params[IgnoreTagValue]; ignore {
				continue
			}

			if _, ok := structToJsonpathTag.Params[IgnoreNestingTagValue]; ok {
				if pkgPath, ok := getFieldPackagePath(field); ok {
					ignoredNesting[pkgPath] = struct{}{}
				}
			}
		}

		if paths[fieldName] == nil {
			paths[fieldName] = []string{}
		}

		if pkgPath, ok := getFieldPackagePath(field); ok {
			if _, ok := ignoredNesting[pkgPath]; ok {
				continue
			}
		}

		if es, ok := extractEmbeddedStruct(field); ok {
			childKeys := flattenChildKeys(traverseStruct(es, ignoreNestingForPkgs))
			paths[fieldName] = mergeUniqueKeys(childKeys, paths[fieldName])
		}
	}

	return paths
}

func generateVariableName(value string) string {
	splits := strings.Split(value, "")
	result := make([]string, 0, len(splits))
	result = append(result, strings.ToUpper(splits[0]))
	for prev, curr := 0, 1; curr < len(value); prev, curr = prev+1, curr+1 {
		v := splits[curr]

		if splits[curr] == "." || splits[curr] == "_" {
			continue
		}

		if splits[prev] == "." || splits[prev] == "_" {
			v = strings.ToUpper(v)
		}
		result = append(result, v)
	}

	return strings.Join(result, "")
}

func genVariables(out io.Writer, structName string, jsonPaths []string) {
	fmt.Fprintf(out, `
// constant vars generated for struct %s
const (
`, structName)

	sort.Strings(jsonPaths)
	for i := range jsonPaths {
		fmt.Fprintf(out, "%s = %q\n", generateVariableName(structName+"."+jsonPaths[i]), jsonPaths[i])
	}

	fmt.Fprintf(out, ")")
}

func main() {
	var structPaths []string
	var pkg string
	var outFile string
	var banner string
	var ignoreNestingForPkgs []string
	flag.StringSliceVar(&structPaths, "struct", nil, "--struct <package-path>.<struct-name>")
	flag.StringVar(&pkg, "pkg", "", "--pkg <package-path>")
	flag.StringVar(&outFile, "out-file", "", "--out-file")
	flag.StringVar(&banner, "banner", "", "--banner")
	flag.StringSliceVar(&ignoreNestingForPkgs, "ignore-nesting", nil, "--ignore-nesting")
	flag.Parse()

	nestingIgnored := make(map[string]struct{})
	for i := range ignoreNestingForPkgs {
		nestingIgnored[ignoreNestingForPkgs[i]] = struct{}{}
	}

	parser := ast_parser.NewASTParser()

	buff := new(bytes.Buffer)
	buff.WriteString(fmt.Sprintf(`
// DO NOT EDIT. generated by "github.com/kloudlite/api/cmd/struct-json-path"

%s
`, banner))

	for i := range structPaths {
		sp := strings.SplitN(fn.StringReverse(structPaths[i]), ".", 2)
		if len(sp) != 2 {
			panic("invalid struct path")
		}

		structName, pkgPath := fn.StringReverse(sp[0]), fn.StringReverse(sp[1])

		structObj, err := parser.FindStruct(pkgPath, structName)
		if err != nil {
			panic(err)
		}

		m := traverseStruct(structObj, nestingIgnored)
		genVariables(buff, structName, flattenChildKeys(m))
	}

	if pkg != "" {
		structObjs, err := parser.FindAllStructs(pkg)
		if err != nil {
			panic(err)
		}

		for i := range structObjs {
			m := traverseStruct(structObjs[i].Struct, nestingIgnored)
			genVariables(buff, structObjs[i].Name, flattenChildKeys(m))
		}
	}

	source, err := format.Source(buff.Bytes())
	if err != nil {
		log.Println("error formatting source:")
		log.Println(buff.String())
		log.Fatal(err)
	}

	if err := os.Chmod(outFile, 0600); err != nil {
		panic(err)
	}
	file, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := os.Chmod(outFile, 0400); err != nil {
		panic(err)
	}
	fmt.Fprintf(file, "%s", source)
}
