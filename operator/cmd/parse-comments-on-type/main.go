package main

import (
	"fmt"
	// "go/constant"
	"go/ast"
	"go/parser"
	"go/token"
	// "strings"

	"golang.org/x/tools/go/packages"
)

func parseStructsFromPkg(packagePath string, typeName string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedEmbedPatterns,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			return parser.ParseFile(fset, filename, src, parser.ParseComments)
		},
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
		//scope := pkg.Types.Scope()
		//
		//obj := scope.Lookup(typeName)
		//if obj == nil {
		//	continue
		//}
		//
		//objStruct, ok := obj.Type().Underlying().(*types.Struct)
		//if !ok {
		//	continue
		//}
		//
		//fmt.Println(objStruct)

		//commentStart := 0
		//commentEnd := 0

		for i := range pkg.Syntax {
			ast.Inspect(pkg.Syntax[i], func(n ast.Node) bool {
				comment, ok := n.(*ast.Comment)
				if !ok {
					return true
				}


				pos := cfg.Fset.Position(comment.Pos())

				fmt.Println(comment)
				fmt.Println(pos)

				// Check if the node is a struct type
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				// Check if the struct name is "Sample"
				if typeSpec.Name.Name == typeName {
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						return true
					}

					fmt.Println(structType)

					// Print the comments above the struct
					//if structType != nil {
					//	fmt.Println("Comments for struct Sample:")
					//	for _, comment := range structType.Doc.List {
					//		fmt.Println(comment.Text)
					//	}
					//}

					return false
				}

				return true
			})
		}

		// for _, name := range scope.Names() {
		// 	varObj := scope.Lookup(name)
		// 	fmt.Println(varObj)
		// 	// if varObj == nil || varObj.Type() == nil {
		// 	// 	continue
		// 	// }
		// 	//
		// 	if varObj.Type().String() == obj.Type().String() {
		// 		if val, ok := varObj.(*types.Struct); ok {
		// 			vstr := val.Val().ExactString()
		// 		}
		// 	}
		// }
	}

	return enums, nil
}

func main() {
	parseStructsFromPkg("github.com/kloudlite/operator/apis/mongodb.msvc/v1", "StandaloneServiceSpec")
}
