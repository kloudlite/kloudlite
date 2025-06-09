package ast_parser

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
)

type Parser struct {
	cfg *packages.Config
}

func (p *Parser) FindStruct(pkgPath string, structName string) (*types.Struct, error) {
	pkgs, err := packages.Load(p.cfg, pkgPath)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages were loaded")
	}

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()

		obj := scope.Lookup(structName)
		if obj == nil {
			continue
		}

		structObj, ok := obj.Type().Underlying().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("not a struct type")
		}

		return structObj, nil
	}

	return nil, fmt.Errorf("not found")
}

type Struct struct {
	Name string
	*types.Struct
}

func (p *Parser) FindAllStructs(pkgPath string) ([]Struct, error) {
	pkgs, err := packages.Load(p.cfg, pkgPath)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages were loaded")
	}

	var structObjs []Struct

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()

		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			structObj, ok := obj.Type().Underlying().(*types.Struct)
			if !ok {
				continue
			}

			structObjs = append(structObjs, Struct{
				Name:   name,
				Struct: structObj,
			})
		}
	}

	return structObjs, nil
}

func NewASTParser() *Parser {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}

	return &Parser{
		cfg: cfg,
	}
}
