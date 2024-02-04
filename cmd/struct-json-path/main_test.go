package main

import (
	ast_parser "github.com/kloudlite/api/cmd/struct-json-path/ast-parser"
	"github.com/kloudlite/api/cmd/struct-json-path/test_data"
	"reflect"
	"sort"
	"testing"
)

func Test_extractTag(t *testing.T) {
	type args struct {
		tagstr string
	}
	tests := []struct {
		name string
		args args
		want map[string]Tag
	}{
		{
			name: "1. json tag with name",
			args: args{tagstr: `json:"hello"`},
			want: map[string]Tag{
				"json": {Value: "hello", Params: map[string]struct{}{}},
			},
		},

		{
			name: "2. json tag with name and inline",
			args: args{tagstr: `json:"hello,inline"`},
			want: map[string]Tag{
				"json": {Value: "hello", Params: map[string]struct{}{"inline": {}}},
			},
		},

		{
			name: "3. json tag with empty name but inline",
			args: args{tagstr: `json:",inline"`},
			want: map[string]Tag{
				"json": {Value: "", Params: map[string]struct{}{"inline": {}}},
			},
		},

		{
			name: "3. json tag with empty value",
			args: args{tagstr: `json:""`},
			want: map[string]Tag{
				"json": {Value: "", Params: map[string]struct{}{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTag(tt.args.tagstr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_flattenChildKeys(t *testing.T) {
	type args struct {
		child map[string][]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenChildKeys(tt.args.child); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenChildKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_traverseStruct(t *testing.T) {
	type args struct {
		pkgPath       string
		structName    string
		ignoreNesting map[string]struct{}
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		// TODO: Add test cases.
		{
			name: "1. struct with fields with no json tag",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test1Input,
			},
			want: test_data.Test1Output,
		},
		{
			name: "2. struct with fields with json tag, with value",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test2Input,
			},
			want: test_data.Test2Output,
		},
		{
			name: "3. struct with fields with both a json tag, and without json tag",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test3Input,
			},
			want: test_data.Test3Output,
		},
		{
			name: "4. struct with an embedded struct, without json tag",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test4Input,
			},
			want: test_data.Test4Output,
		},
		{
			name: "5. struct with an embedded struct, with one without a tag, and other one with a json tag value",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test5Input,
			},
			want: test_data.Test5Output,
		},
		{
			name: "6. struct with an embedded struct, with one inline struct",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test6Input,
			},
			want: test_data.Test6Output,
		},
		{
			name: "7. struct with an embedded struct, with 2 inline structs with a common field",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test7Input,
			},
			want: test_data.Test7Output,
		},
		{
			name: "8. struct with non-embedded struct with json tag",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test8Input,
			},
			want: test_data.Test8Output,
		},
		{
			name: "9. struct with non-embedded struct with json tag, but with one value with json tag -",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test9Input,
			},
			want: test_data.Test9Output,
		},
		{
			name: "9. struct with non-embedded struct with json tag, but with one value with json tag -",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test9Input,
			},
			want: test_data.Test9Output,
		},
		{
			name: "10. struct with struct embedded without field name",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test10Input,
			},
			want: test_data.Test10Output,
		},
		{
			name: "11. struct with field having struct-json-path set to ignore",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test11Input,
			},
			want: test_data.Test11Output,
		},
		{
			name: "12. struct with field having ignored nesting for packages",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test12Input,
				ignoreNesting: map[string]struct{}{
					"time.Time": {},
				},
			},
			want: test_data.Test12Output,
		},
		{
			name: "13. struct with fields having ignored-nesting tag on struct field",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test13Input,
			},
			want: test_data.Test13Output,
		},
		{
			name: "14. struct with fields having a custom string like type",
			args: args{
				pkgPath:    test_data.PkgPath,
				structName: test_data.Test14Input,
			},
			want: test_data.Test14Output,
		},
	}

	parser := ast_parser.NewASTParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			structObj, err := parser.FindStruct(tt.args.pkgPath, tt.args.structName)
			if err != nil {
				t.Errorf("no struct found with pkgpath (%s) and type (%s)", tt.args.pkgPath, tt.args.structName)
			}

			got := traverseStruct(structObj, tt.args.ignoreNesting)
			for k := range got {
				sort.Strings(got[k])
			}

			for k := range tt.want {
				sort.Strings(tt.want[k])
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("traverseStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}
