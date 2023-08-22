package parser

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"go/types"
	"golang.org/x/tools/go/packages"
	"regexp"
	"strings"
	"text/template"
)

var counter int

var charCounter int = 97

func getCharVariable() string {
	defer func() { charCounter++ }()
	return fmt.Sprintf("k%c", charCounter)
}

type ImportInfo struct {
	Alias       string
	PackagePath string
}

type Info struct {
	Imports         map[string]ImportInfo
	Implementations []string
	MockFunctions   []string
	StructName      string
}

func ExtractDefinitionArgs(funcDecl string, imports map[string]ImportInfo) (string, map[string]ImportInfo, error) {
	//pattern := `[(](.*)[)]\s|$?`
	pattern := `[(](.*)[)]`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", nil, err
	}

	x := re.FindStringSubmatch(funcDecl)
	if len(x) < 2 {
		return "", nil, fmt.Errorf("could not find first group with regex %q", pattern)
	}

	sp := strings.Split(x[1], ",")
	args := make([]string, 0, len(sp))

	for i := range sp {
		s := strings.TrimSpace(sp[i])
		if s == "" {
			continue
		}
		_sp2 := strings.SplitN(s, " ", 2)
		sp2 := [2]string{}
		if len(_sp2) != 2 {
			sp2[1] = _sp2[0]
			sp2[0] = getCharVariable()
			// return "", imports, fmt.Errorf("invalid entry %q, sp: %#v", s, sp)
		} else {
			sp2[0] = _sp2[0]
			sp2[1] = _sp2[1]
		}

		// to separate the type from package name
		// like, from (name string, age int, ex kloudlite.io/cmd/interface-impl/types.Example),
		// to (name string, age int, ex Example)
		s2 := strings.Split(sp2[1], "/")
		argType := s2[len(s2)-1]
		argTypeImport := ""
		if asp := strings.Split(argType, "."); len(asp) == 2 {
			argTypeImport = asp[0]
			if len(s2) > 1 {
				argTypeImport = strings.Join(s2[:len(s2)-1], "/") + "/" + asp[0]
			}

			hasPointer := false
			if strings.HasPrefix(argTypeImport, "*") {
				hasPointer = true
				argTypeImport = argTypeImport[1:]
			}
			hasVariadic := false
			if strings.HasPrefix(argTypeImport, "...") {
				hasVariadic = true
				argTypeImport = argTypeImport[3:]
			}
			_, ok := imports[argTypeImport]
			if !ok {
				counter++
				imports[argTypeImport] = ImportInfo{PackagePath: fmt.Sprintf("%q", argTypeImport), Alias: fmt.Sprintf("%s%d", asp[0], counter)}
			}
			argType = fmt.Sprintf("%s.%s", imports[argTypeImport].Alias, asp[1])
			if hasPointer {
				argType = "*" + argType
			}
			if hasVariadic {
				argType = "..." + argType
			}
		}

		args = append(args, fmt.Sprintf("%s %s", sp2[0], argType))
	}

	return fmt.Sprintf("(%s)", strings.Join(args, ", ")), imports, nil
}

func ExtractReturnValues(funcDecl string, imports map[string]ImportInfo) (string, map[string]ImportInfo, error) {
	// pattern := `[(].*[)]\s[(]?([\w*/.-]+)[)]?`
	pattern := `\(.*\)\s+[(]?([^)]+)[)]?`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", nil, err
	}

	x := re.FindStringSubmatch(funcDecl)
	if len(x) == 0 {
		return "", nil, nil
	}

	// fmt.Printf("%v %d\n", x, len(x))
	if len(x) < 2 {
		return "", nil, fmt.Errorf("could not find 1st group with regex %q", pattern)
	}

	sp := strings.Split(x[1], ",")
	args := make([]string, 0, len(sp))
	for i := range sp {
		s := strings.TrimSpace(sp[i])
		if s == "" {
			continue
		}
		sp2 := strings.Split(s, " ")
		var argName string
		t := sp2[0]
		if len(sp2) == 2 {
			argName = sp2[0]
			t = sp2[1]
		}

		// to separate the type from package name
		// like, from (name string, age int, ex kloudlite.io/cmd/interface-impl/types.Example),
		// to (name string, age int, ex Example)
		s2 := strings.Split(t, "/")
		argType := s2[len(s2)-1]
		argTypeImport := ""
		if asp := strings.Split(argType, "."); len(asp) == 2 {
			argTypeImport = asp[0]
			if len(s2) > 1 {
				argTypeImport = strings.Join(s2[:len(s2)-1], "/") + "/" + asp[0]
			}

			hasPointer := false
			if strings.HasPrefix(argTypeImport, "*") {
				hasPointer = true
				argTypeImport = argTypeImport[1:]
			}

			_, ok := imports[argTypeImport]
			if !ok {
				counter++
				imports[argTypeImport] = ImportInfo{PackagePath: fmt.Sprintf("%q", argTypeImport), Alias: fmt.Sprintf("%s%d", asp[0], counter)}
			}

			argType = fmt.Sprintf("%s.%s", imports[argTypeImport].Alias, asp[1])
			if hasPointer {
				argType = "*" + argType
			}
		}

		args = append(args, strings.TrimSpace(fmt.Sprintf("%s %s", argName, argType)))
	}

	return fmt.Sprintf("(%s)", strings.Join(args, ", ")), imports, nil
}

func ExtractCallArgs(definitionArgs string) ([]string, error) {
	pattern := `[(]([^(]+)[)]`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	x := re.FindStringSubmatch(definitionArgs)
	// fmt.Println("definitionArgs:", definitionArgs, "\tx: ", x, len(x))

	if len(x) == 0 {
		return []string{}, nil
	}

	if len(x) < 2 {
		return nil, fmt.Errorf("could not find first group with regex %q", pattern)
	}

	sp := strings.Split(x[1], ",")
	args := make([]string, 0, len(sp))
	for i := range sp {
		s := strings.TrimSpace(sp[i])
		if s == "" {
			continue
		}
		sp2 := strings.SplitN(s, " ", 2)
		if len(sp2) != 2 {
			return nil, fmt.Errorf("invalid entry %q, sp: %#v", s, sp)
		}
		if strings.HasPrefix(sp2[1], "...") {
			sp2[0] = sp2[0] + "..."
		}
		args = append(args, sp2[0])
	}

	return args, nil
}

func FindAndParseInterface(packagePath string, interfaceName string) (*Info, error) {
	info := Info{Imports: make(map[string]ImportInfo)}
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
		// Mode: packages.LoadSyntax,
	}

	pkgs, err := packages.Load(cfg, packagePath)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages were loaded")
	}

	found := false

	for _, pkg := range pkgs {
		//for _, file := range pkg.Syntax {
		//	for _, importSpec := range file.Imports {
		//		importInfo := ImportInfo{PackagePath: importSpec.Path.Value}
		//		if importSpec.Name != nil && importSpec.Name.Name != "" {
		//			importInfo.Alias = importSpec.Name.Name
		//		}
		//		if _, ok := info.Imports[importSpec.Path.Value]; !ok {
		//			info.Imports[importSpec.Path.Value] = importInfo
		//		}
		//	}
		//}

		scope := pkg.Types.Scope()
		obj := scope.Lookup(interfaceName)

		sp := strings.Split(obj.Type().String(), "/")
		typeName := sp[len(sp)-1]
		nsp := strings.SplitN(typeName, ".", 2)
		interfaceName = nsp[1]

		if obj != nil {
			ifaceType, ok := obj.Type().Underlying().(*types.Interface)
			if !ok {
				return nil, fmt.Errorf("not an interface: %s", ifaceType.String())
			}
			found = true

			info.StructName = interfaceName
			receiverName := strings.ToLower(string(info.StructName[0]))
			receiverStructName := interfaceName

			// can be of format: Sample, Sample[T any, K any] etc.
			index := strings.Index(interfaceName, "[")
			if index != -1 {
				gsp := strings.Split(interfaceName[index+1:], ",")
				genericTypes := make([]string, 0, len(gsp))
				for i := range gsp {
					genericTypes = append(genericTypes, strings.TrimSpace(strings.SplitN(strings.TrimSpace(gsp[i]), " ", 2)[0]))
				}
				receiverStructName = fmt.Sprintf("%s[%s]", interfaceName[:index], strings.Join(genericTypes, ", "))
			}

			for i := 0; i < ifaceType.NumMethods(); i++ {
				if ifaceType.Method(i).Exported() {
					method := ifaceType.Method(i)
					ftype := method.Type().String()
					// sp := strings.SplitN(ftype, "func", 2)

					defArgs, extraImports, err := ExtractDefinitionArgs(ftype, info.Imports)
					if err != nil {
						return nil, err
					}

					for k, v := range extraImports {
						if _, ok := info.Imports[k]; !ok {
							info.Imports[k] = v
						}
					}

					retValues, extraImports, err := ExtractReturnValues(ftype, info.Imports)
					if err != nil {
						return nil, err
					}
					for k, v := range extraImports {
						if _, ok := info.Imports[k]; !ok {
							info.Imports[k] = v
						}
					}

					callArgs, err := ExtractCallArgs(defArgs)
					if err != nil {
						return nil, err
					}

					buff := new(bytes.Buffer)
					t := template.New("impl_gen")
					t.Funcs(sprig.TxtFuncMap())
					t.Parse(`func ({{.ReceiverName}} *{{.ReceiverStructName}}) {{.FunctionName}}{{.FunctionArgs}} {{.ReturnValues}} {
  if {{.ReceiverName}}.{{.MockFunctionName}} != nil {
    {{.ReceiverName}}.registerCall("{{.FunctionName}}", {{.CallArgs | join ", " | replace "..." "" }})
    {{if .ReturnValues}}return {{end}}{{.ReceiverName}}.{{.MockFunctionName}}({{.CallArgs | join ", "}})
  }
  panic("not implemented, yet")
}`)
					if err := t.ExecuteTemplate(buff, "impl_gen", map[string]any{
						"ReceiverName":       receiverName,
						"StructName":         info.StructName,
						"ReceiverStructName": receiverStructName,
						"FunctionName":       method.Name(),
						"FunctionArgs":       defArgs,
						"ReturnValues":       retValues,
						"CallArgs":           callArgs,
						"MockFunctionName":   fmt.Sprintf("Mock%s", method.Name()),
					}); err != nil {
						return nil, err
					}

					info.Implementations = append(info.Implementations, buff.String())
					info.MockFunctions = append(info.MockFunctions, fmt.Sprintf(`%s func%s %s`, "Mock"+method.Name(), defArgs, retValues))
				}
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("interface %q not found in  package %q", interfaceName, packagePath)
	}

	return &info, nil
}
