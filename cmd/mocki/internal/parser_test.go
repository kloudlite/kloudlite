package parser_test

import (
	parser "kloudlite.io/cmd/mocki/internal"
	"reflect"
	"testing"
)

func TestExtractCallArgs(t *testing.T) {
	type args struct {
		definitionArgs string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "test 1: no arguments in function declaration",
			args: args{
				definitionArgs: "func()",
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "test 2: single argument in function declaration, with no return",
			args: args{
				definitionArgs: "func(x int)",
			},
			want:    []string{"x"},
			wantErr: false,
		},
		{
			name: "test 3: single argument in function declaration, with one return type",
			args: args{
				definitionArgs: "func(x int) int",
			},
			want:    []string{"x"},
			wantErr: false,
		},
		{
			name: "test 4: single argument in function declaration, with one named return type",
			args: args{
				definitionArgs: "func(x int) int",
			},
			want:    []string{"x"},
			wantErr: false,
		},
		{
			name: "test 5: multiple arguments in function declaration, with no return",
			args: args{
				definitionArgs: "func(x int, y string, z int)",
			},
			want:    []string{"x", "y", "z"},
			wantErr: false,
		},
		{
			name: "test 6: multiple arguments in function declaration, with one return type",
			args: args{
				definitionArgs: "func(x int, y string, z int) int",
			},
			want:    []string{"x", "y", "z"},
			wantErr: false,
		},
		{
			name: "test 6: multiple arguments in function declaration, with multiple return type",
			args: args{
				definitionArgs: "func(x int, y string, z int) (int, error)",
			},
			want:    []string{"x", "y", "z"},
			wantErr: false,
		},
		{
			name: "test 7: multiple arguments in function declaration, with pointer return type",
			args: args{
				definitionArgs: "func(x int, y string, z int) (*int, error)",
			},
			want:    []string{"x", "y", "z"},
			wantErr: false,
		},
		{
			name: "test 8: multiple arguments (pointer and non-pointer mix) in function declaration, with pointer return type",
			args: args{
				definitionArgs: "func(x int, y string, z *int) (*int, error)",
			},
			want:    []string{"x", "y", "z"},
			wantErr: false,
		},
		{
			name: "test 9: multiple arguments (custom type) in function declaration, with pointer return type",
			args: args{
				definitionArgs: "func(x Sample, y string, z *int) (*int, error)",
			},
			want:    []string{"x", "y", "z"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ExtractCallArgs(tt.args.definitionArgs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractCallArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractCallArgs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractDefinitionArgs(t *testing.T) {
	type args struct {
		funcDecl string
		imports  map[string]parser.ImportInfo
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   map[string]parser.ImportInfo
		wantErr bool
	}{
		{
			name: "test 1: no arguments in function declaration",
			args: args{
				funcDecl: "func()",
				imports:  nil,
			},
			want:    "()",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "test 2: one argument in function declaration",
			args: args{
				funcDecl: "func(x int)",
				imports:  nil,
			},
			want:    "(x int)",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "test 3: one pointer argument in function declaration",
			args: args{
				funcDecl: "func(x *int)",
				imports:  nil,
			},
			want:    "(x *int)",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "test 4: multiple non-pointer argument in function declaration",
			args: args{
				funcDecl: "func(x int, y string) int",
				imports:  nil,
			},
			want:    "(x int, y string)",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "test 5: multiple pointer argument in function declaration",
			args: args{
				funcDecl: "func(x *int, y *string) int",
				imports:  nil,
			},
			want:    "(x *int, y *string)",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "test 6: mixed pointer and non-pointer argument in function declaration",
			args: args{
				funcDecl: "func(x *int, y *string, z int) int",
				imports:  nil,
			},
			want:    "(x *int, y *string, z int)",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "test 7: custom type argument in function declaration",
			args: args{
				funcDecl: "func(x *int, y *string, z Sample) int",
				imports:  nil,
			},
			want:    "(x *int, y *string, z Sample)",
			want1:   nil,
			wantErr: false,
		},
		{
			name: "test 8: arguments without name",
			args: args{
				funcDecl: "func(*int, *string, Sample) int",
				imports:  nil,
			},
			want:    "(ka *int, kb *string, kc Sample)",
			want1:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parser.ExtractDefinitionArgs(tt.args.funcDecl, tt.args.imports)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractDefinitionArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractDefinitionArgs() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ExtractDefinitionArgs() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestExtractReturnValues(t *testing.T) {
	type args struct {
		funcDecl string
		imports  map[string]parser.ImportInfo
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   map[string]parser.ImportInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parser.ExtractReturnValues(tt.args.funcDecl, tt.args.imports)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractReturnValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractReturnValues() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ExtractReturnValues() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.FindAndParseInterface(tt.args.packagePath, tt.args.interfaceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAndParseInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAndParseInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}
