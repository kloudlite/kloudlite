package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func main() {
	packagePath := "kloudlite.io/apps/iam/types"

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
		// Mode: packages.LoadSyntax,
	}

	pkgs, err := packages.Load(cfg, packagePath)
	if err != nil {
		panic(err)
	}

	if len(pkgs) == 0 {
		panic(fmt.Errorf("no packages were loaded"))
	}

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()

		obj := scope.Lookup("Role")
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
					//fmt.Println(varObj.Name())
					fmt.Println(val.Val())
				}
			}
		}
	}
}
