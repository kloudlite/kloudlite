package parser_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	parser "kloudlite.io/cmd/mocki/internal/parser"
)

func TestFindAndParseInterface(t *testing.T) {
	type args struct {
		packagePath   string
		interfaceName string
	}
	tests := []struct {
		name    string
		args    args
		want    *parser.Info
		wantErr bool
	}{
		{
			name: "test 1: normal interface",
			args: args{
				packagePath:   "kloudlite.io/cmd/mocki/internal/parser/test_data",
				interfaceName: "Type1",
			},
			want: &parser.Info{
				Imports: map[string]parser.ImportInfo{},
				Implementations: []string{
					`func (tMock *Type1) Method1() string {
	if tMock.MockMethod1 != nil {
		tMock.registerCall("Method1")
		return tMock.MockMethod1()
	}
  panic("Type1: method 'Method1' not implemented, yet")
}`,
				},
				MockFunctions: []string{
					`MockMethod1 func() string`,
				},
				StructName:         "Type1",
				ReceiverStructName: "Type1",
			},
			wantErr: false,
		},
		{
			name: "test 2: generic interface, with any constraint",
			args: args{
				packagePath:   "kloudlite.io/cmd/mocki/internal/parser/test_data",
				interfaceName: "Type2",
			},
			want: &parser.Info{
				Imports: map[string]parser.ImportInfo{
					"kloudlite.io/cmd/mocki/internal/parser/test_data": {
						Alias:       "test_data",
						PackagePath: "kloudlite.io/cmd/mocki/internal/parser/test_data",
					},
					"kloudlite.io/pkg/repos": {
						Alias:       "repos",
						PackagePath: "kloudlite.io/pkg/repos",
					},
				},
				Implementations: []string{
					`func (tMock *Type2[T]) Method1() T {
	if tMock.MockMethod1 != nil {
		tMock.registerCall("Method1")
		return tMock.MockMethod1()
	}
  panic("Type2[T]: method 'Method1' not implemented, yet")
}`,

					`func (tMock *Type2[T]) Method2(x int, y ...[]byte) string {
	if tMock.MockMethod2 != nil {
		tMock.registerCall("Method2", x, y)
		return tMock.MockMethod2(x, y...)
	}
  panic("Type2[T]: method 'Method2' not implemented, yet")
}`,

					`func (tMock *Type2[T]) Method3(x int, y *int, z T, p *repos.DbRepo[test_data.X], q map[string]test_data.X, r *test_data.X, s []int, u ...test_data.X) string {
	if tMock.MockMethod3 != nil {
		tMock.registerCall("Method3", x, y, z, p, q, r, s, u)
		return tMock.MockMethod3(x, y, z, p, q, r, s, u...)
	}
  panic("Type2[T]: method 'Method3' not implemented, yet")
}`,
				},
				MockFunctions: []string{
					`MockMethod1 func() T`,
					`MockMethod2 func(x int, y ...[]byte) string`,
					`MockMethod3 func(x int, y *int, z T, p *repos.DbRepo[test_data.X], q map[string]test_data.X, r *test_data.X, s []int, u ...test_data.X) string`,
				},
				StructName:         "Type2[T any]",
				ReceiverStructName: "Type2[T]",
			},
			wantErr: false,
		},
		{
			name: "test 3: generic interface with another interface as constraint",
			args: args{
				packagePath:   "kloudlite.io/cmd/mocki/internal/parser/test_data",
				interfaceName: "Type3",
			},
			want: &parser.Info{
				Imports: map[string]parser.ImportInfo{
					"kloudlite.io/cmd/mocki/internal/parser/test_data": {
						Alias:       "test_data",
						PackagePath: "kloudlite.io/cmd/mocki/internal/parser/test_data",
					},
				},
				Implementations: []string{
					`func (tMock *Type3[T]) Method1() T {
	if tMock.MockMethod1 != nil {
		tMock.registerCall("Method1")
		return tMock.MockMethod1()
	}
  panic("Type3[T]: method 'Method1' not implemented, yet")
}`,
				},
				MockFunctions: []string{
					`MockMethod1 func() T`,
				},
				StructName:         "Type3[T test_data.Entity]",
				ReceiverStructName: "Type3[T]",
			},
			wantErr: false,
		},
		{
			name: "test 4: normal interface with methods having no named arguments",
			args: args{
				packagePath:   "kloudlite.io/cmd/mocki/internal/parser/test_data",
				interfaceName: "Type4",
			},
			want: &parser.Info{
				Imports: map[string]parser.ImportInfo{
					"context": {
						Alias:       "context",
						PackagePath: "context",
					},
				},
				Implementations: []string{
					`func (tMock *Type4) Method1(ka context.Context, kb int) string {
	if tMock.MockMethod1 != nil {
		tMock.registerCall("Method1", ka, kb)
		return tMock.MockMethod1(ka, kb)
	}
  panic("Type4: method 'Method1' not implemented, yet")
}`,
				},
				MockFunctions: []string{
					`MockMethod1 func(ka context.Context, kb int) string`,
				},
				StructName:         "Type4",
				ReceiverStructName: "Type4",
			},
			wantErr: false,
		},
		{
			name: "test 5: interface with methods having no return values",
			args: args{
				packagePath:   "kloudlite.io/cmd/mocki/internal/parser/test_data",
				interfaceName: "Type5",
			},
			want: &parser.Info{
				Imports: map[string]parser.ImportInfo{},
				Implementations: []string{
					`func (tMock *Type5) Method1() {
	if tMock.MockMethod1 != nil {
		tMock.registerCall("Method1")
		tMock.MockMethod1()
	}
  panic("Type5: method 'Method1' not implemented, yet")
}`,

					`func (tMock *Type5) Method2(x int) {
	if tMock.MockMethod2 != nil {
		tMock.registerCall("Method2", x)
		tMock.MockMethod2(x)
	}
  panic("Type5: method 'Method2' not implemented, yet")
}`,
				},
				MockFunctions: []string{
					`MockMethod1 func()`,
					`MockMethod2 func(x int)`,
				},
				StructName:         "Type5",
				ReceiverStructName: "Type5",
			},
			wantErr: false,
		},

		{
			name: "test 6: interface with methods having aliases imported types",
			args: args{
				packagePath:   "kloudlite.io/cmd/mocki/internal/parser/test_data",
				interfaceName: "Type6",
			},
			want: &parser.Info{
				Imports: map[string]parser.ImportInfo{
					"io": {
						Alias:       "io2",
						PackagePath: "io",
					},
				},
				Implementations: []string{
					`func (tMock *Type6) Method1(writer io2.Writer) {
	if tMock.MockMethod1 != nil {
		tMock.registerCall("Method1", writer)
		tMock.MockMethod1(writer)
	}
  panic("Type6: method 'Method1' not implemented, yet")
}`,
				},
				MockFunctions: []string{
					`MockMethod1 func(writer io2.Writer)`,
				},
				StructName:         "Type6",
				ReceiverStructName: "Type6",
			},
			wantErr: false,
		},

		{
			name: "test 7: non existent interface, should throw error",
			args: args{
				packagePath:   "kloudlite.io/cmd/mocki/internal/parser/test_data",
				interfaceName: "DoesNotExist",
			},
			want: &parser.Info{
				Imports: map[string]parser.ImportInfo{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser()
			got, err := p.FindAndParseInterface(tt.args.packagePath, tt.args.interfaceName)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("FindAndParseInterface() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				return
			}

			for i := range tt.want.Implementations {
				tt.want.Implementations[i] = strings.Replace(tt.want.Implementations[i], "\n", "  ", -1)
			}

			for i := range got.Implementations {
				got.Implementations[i] = strings.Replace(got.Implementations[i], "\n", "  ", -1)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("FindAndParseInterface(), got=\n%s\n\nwant:\n%s\n", got, tt.want)
			}
		})
	}
}
