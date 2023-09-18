package parser

import (
	"fmt"
	"go/constant"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strings"
)

func parseConstantsFromPkg(packagePath string, typeName string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}

	pkgs, err := packages.Load(cfg, packagePath)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages were loaded")
	}

	var enums []string

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()

		obj := scope.Lookup(typeName)
		if obj == nil {
			continue
		}

		for _, name := range scope.Names() {
			varObj := scope.Lookup(name)
			if varObj == nil || varObj.Type() == nil {
				continue
			}

			if varObj.Type().String() == obj.Type().String() {
				if val, ok := varObj.(*types.Const); ok {
					vstr := val.Val().ExactString()
					if strings.HasPrefix(vstr, "\"") {
						vstr = constant.StringVal(val.Val())
					}
					enums = append(enums, vstr)
				}
			}
		}
	}

	return enums, nil
}
