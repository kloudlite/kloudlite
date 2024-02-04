package parser

import (
	"bytes"
	"fmt"
	"go/types"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"golang.org/x/tools/go/packages"
)

type ImportInfo struct {
	Alias       string
	PackagePath string
}

type Parser struct {
	pkgImports  map[string]ImportInfo
	Imports     map[string]ImportInfo
	counter     int
	charCounter int
}

func NewParser() *Parser {
	return &Parser{
		pkgImports:  map[string]ImportInfo{},
		Imports:     map[string]ImportInfo{},
		counter:     1,
		charCounter: 97,
	}
}

type Info struct {
	Imports            map[string]ImportInfo
	Implementations    []string
	MockFunctions      []string
	StructName         string
	ReceiverStructName string
}

func (p *Parser) getCharVariable() string {
	defer func() { p.charCounter++ }()
	return fmt.Sprintf("k%c", p.charCounter)
}

func (p *Parser) getCounter() int {
	defer func() { p.counter++ }()
	return p.counter
}

func (p *Parser) ExtractGenericConstraints(t types.Type) ([]string, []string, error) {
	named, ok := t.(*types.Named)
	if !ok {
		return nil, nil, fmt.Errorf("not a named type: %s", t.String())
	}

	tp := named.TypeParams()

	constraints := make([]string, 0, tp.Len())
	constraintVars := make([]string, 0, tp.Len())

	for i := 0; i < tp.Len(); i++ {
		constraint := tp.At(i).Obj().Name()
		constraintVars = append(constraintVars, constraint)

		ct, err := p.TypeExtracter(tp.At(i).Constraint(), false)
		if err != nil {
			return nil, nil, err
		}

		constraints = append(constraints, fmt.Sprintf("%s %s", constraint, ct))
	}

	return constraints, constraintVars, nil
}

func (p *Parser) TypeExtracter(ut types.Type, variadic bool) (string, error) {
	switch t := ut.(type) {
	case *types.Named:
		if t.Obj().Pkg() != nil {
			pkgPath := t.Obj().Pkg().Path()
			pkgName := t.Obj().Pkg().Name()

			alias := pkgName

			if _, ok := p.Imports[pkgPath]; !ok {
				v, ok2 := p.pkgImports[pkgPath]
				if ok2 {
					if v.Alias != "" {
						alias = v.Alias
					}
				}

				// pkgName = fmt.Sprintf("%s%d", pkgName, p.getCounter())
				p.Imports[pkgPath] = ImportInfo{PackagePath: pkgPath, Alias: alias}
			}

			targs := make([]string, 0, t.TypeArgs().Len())
			for i := 0; i < t.TypeArgs().Len(); i++ {
				ta, err := p.TypeExtracter(t.TypeArgs().At(i), false)
				if err != nil {
					return "", err
				}
				targs = append(targs, ta)
			}

			if len(targs) > 0 {
				return fmt.Sprintf("%s.%s[%s]", alias, t.Obj().Name(), strings.Join(targs, ", ")), nil
			}
			return fmt.Sprintf("%s.%s", alias, t.Obj().Name()), nil
		}

		return t.Obj().Name(), nil
	case *types.Pointer:
		t2, err := p.TypeExtracter(t.Elem(), variadic)
		if err != nil {
			return "", err
		}
		return "*" + t2, nil
	case *types.Slice:
		t2, err := p.TypeExtracter(t.Elem(), variadic)
		if err != nil {
			return "", err
		}

		if variadic && !strings.HasPrefix(t2, "...") {
			return "..." + t2, nil
		}

		if strings.HasPrefix(t2, "...") {
			return "...[]" + strings.Replace(t2, "...", "", 1), nil
		}
		return "[]" + t2, nil
	case *types.Map:
		k, err := p.TypeExtracter(t.Elem(), false)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("map[%s]%s", t.Key(), k), nil
	default:
		return t.String(), nil
	}
}

func (p *Parser) ExtractResults(signature *types.Signature) ([]string, error) {
	results := signature.Results()
	res := make([]string, 0, results.Len())
	for i := 0; i < results.Len(); i++ {
		// variadic is false, as result arguments cannot be variadic
		t, err := p.TypeExtracter(results.At(i).Type(), false)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}

	return res, nil
}

func (p *Parser) ExtractParameters(signature *types.Signature) ([]string, []string, error) {
	// typeParams := signature.TypeParams()
	params := signature.Params()
	plist := make([]string, 0, params.Len())
	callVarsList := make([]string, 0, params.Len())

	for i := 0; i < params.Len(); i++ {
		pt := params.At(i)
		x := pt.String()

		// because, only last element in parameter list, can be variadic
		variadic := i == params.Len()-1 && signature.Variadic()

		_ = x

		t, err := p.TypeExtracter(pt.Type(), variadic)
		if err != nil {
			return nil, nil, err
		}

		name := pt.Name()

		if name == "" {
			name = p.getCharVariable()
		}

		callVarsList = append(callVarsList, func() string {
			if variadic {
				return fmt.Sprintf("%s...", name)
			}
			return name
		}())
		plist = append(plist, fmt.Sprintf("%s %s", name, t))
	}

	return plist, callVarsList, nil
}

type ImplementationArgs struct {
	ReceiverName       string
	StructName         string
	ReceiverStructName string
	FunctionName       string
	FunctionParams     string
	FunctionReturns    []string
	FunctionCallArgs   []string
	MockFunctionName   string
}

func (args *ImplementationArgs) FormatReturnValues() string {
	if len(args.FunctionReturns) == 0 {
		return ""
	}

	if len(args.FunctionReturns) == 1 {
		return args.FunctionReturns[0]
	}

	return fmt.Sprintf("(%s)", strings.Join(args.FunctionReturns, ", "))
}

func GenerateImplementation(args ImplementationArgs) (string, error) {
	buff := new(bytes.Buffer)
	tt := template.New("impl_gen")
	tt.Funcs(sprig.TxtFuncMap())

	if _, err := tt.Parse(`func ({{.ReceiverName}} *{{.ReceiverStructName}}) {{.FunctionName}}{{.FunctionParams}}{{- if .FunctionReturns }} {{.FunctionReturns}} {{- end}} {
	if {{.ReceiverName}}.{{.MockFunctionName}} != nil {
		{{.ReceiverName}}.registerCall("{{.FunctionName}}" {{- if .FunctionCallArgs}}, {{end}} {{- .FunctionCallArgs | join ", " | replace "..." "" }})
		{{if .FunctionReturns}}return {{end}}{{.ReceiverName}}.{{.MockFunctionName}}({{.FunctionCallArgs | join ", "}})
	}
  panic("{{.ReceiverStructName}}: method '{{.FunctionName}}' not implemented, yet")
}`); err != nil {
		return "", err
	}

	if err := tt.ExecuteTemplate(buff, "impl_gen", map[string]any{
		"ReceiverName":       args.ReceiverName,
		"StructName":         args.StructName,
		"ReceiverStructName": args.ReceiverStructName,
		"FunctionName":       args.FunctionName,
		"FunctionParams":     args.FunctionParams,
		"FunctionReturns":    args.FormatReturnValues(),
		"FunctionCallArgs":   args.FunctionCallArgs,
		"MockFunctionName":   args.MockFunctionName,
	}); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func extractPackageImports(pkg *packages.Package) (map[string]ImportInfo, error) {
	imports := make(map[string]ImportInfo)

	for _, file := range pkg.Syntax {
		for _, importSpec := range file.Imports {
			pkgPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return nil, err
			}
			importInfo := ImportInfo{PackagePath: pkgPath}
			if importSpec.Name != nil && importSpec.Name.Name != "" {
				importInfo.Alias = importSpec.Name.Name
			}
			if _, ok := imports[pkgPath]; !ok {
				imports[pkgPath] = importInfo
			}
		}
	}

	return imports, nil
}

func findNamedType(packagePath string, interfaceName string) (types.Type, map[string]ImportInfo, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
		// Mode: packages.LoadSyntax,
	}

	pkgs, err := packages.Load(cfg, packagePath)
	if err != nil {
		return nil, nil, err
	}

	if len(pkgs) == 0 {
		return nil, nil, fmt.Errorf("no packages were loaded")
	}

	for _, pkg := range pkgs {
		pkgImports, err := extractPackageImports(pkg)
		if err != nil {
			return nil, nil, err
		}

		scope := pkg.Types.Scope()
		obj := scope.Lookup(interfaceName)

		if obj == nil {
			continue
		}

		return obj.Type(), pkgImports, nil
	}

	return nil, nil, fmt.Errorf("could not find named type %q in package path %q", interfaceName, packagePath)
}

func (p *Parser) FindAndParseInterface(packagePath string, interfaceName string) (*Info, error) {
	info := Info{}

	namedType, pkgImports, err := findNamedType(packagePath, interfaceName)
	if err != nil {
		return nil, err
	}

	interfaceType, ok := namedType.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("not an interface: %s", interfaceName)
	}

	p.pkgImports = pkgImports

	genericConstraints, constraintVars, err := p.ExtractGenericConstraints(namedType)
	if err != nil {
		return nil, err
	}

	info.StructName = interfaceName
	if len(genericConstraints) > 0 {
		info.StructName = info.StructName + "[" + strings.Join(genericConstraints, ", ") + "]"
	}

	receiverName := fmt.Sprintf("%sMock", strings.ToLower(string(info.StructName[0])))
	info.ReceiverStructName = interfaceName

	// can be of format: Sample, Sample[T any, K any] etc.
	if len(constraintVars) > 0 {
		info.ReceiverStructName = fmt.Sprintf("%s[%s]", interfaceName, strings.Join(constraintVars, ", "))
	}

	for i := 0; i < interfaceType.NumMethods(); i++ {
		method := interfaceType.Method(i)
		if !method.Exported() {
			continue
		}

		signature, ok := method.Type().(*types.Signature)
		if !ok {
			return nil, fmt.Errorf("not a signature type: %s", method.Type().String())
		}

		parameters, paramCallArgs, err := p.ExtractParameters(signature)
		if err != nil {
			return nil, err
		}

		results, err := p.ExtractResults(signature)
		if err != nil {
			return nil, err
		}

		iArgs := ImplementationArgs{
			ReceiverName:       receiverName,
			StructName:         info.StructName,
			ReceiverStructName: info.ReceiverStructName,
			FunctionName:       method.Name(),
			FunctionParams:     fmt.Sprintf("(%s)", strings.Join(parameters, ", ")),
			FunctionReturns:    results,
			FunctionCallArgs:   paramCallArgs,
			MockFunctionName:   fmt.Sprintf("Mock%s", method.Name()),
		}

		implementation, err := GenerateImplementation(iArgs)
		if err != nil {
			return nil, err
		}

		info.Implementations = append(info.Implementations, implementation)
		info.MockFunctions = append(info.MockFunctions, func() string {
			if iArgs.FormatReturnValues() == "" {
				return fmt.Sprintf("%s func%s", "Mock"+method.Name(), iArgs.FunctionParams)
			}
			return fmt.Sprintf("%s func%s %s", "Mock"+method.Name(), iArgs.FunctionParams, iArgs.FormatReturnValues())
		}())
	}

	info.Imports = p.Imports
	return &info, nil
}
