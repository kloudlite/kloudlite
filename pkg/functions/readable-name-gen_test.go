package functions

import (
	"strings"
	"testing"
)

func TestGenReadableName(t *testing.T) {
	type args struct {
		seed string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "seed with space", args: args{seed: "hello world"}, want: "hello-"},
		{name: "seed with camelcase", args: args{seed: "helloWorld"}, want: "hello-"},
		{name: "seed with hyphen", args: args{seed: "hello-world"}, want: "hello-"},
		{name: "seed with dot", args: args{seed: "hello.world"}, want: "hello-"},
		{name: "seed with underscore", args: args{seed: "hello_world"}, want: "hello-"},
		{name: "seed with numbers", args: args{seed: "hello123"}, want: "hello-"},

		// from chatgpt
		{name: "seed with single word", args: args{seed: "hello"}, want: "hello-"},
		{name: "empty seed", args: args{seed: ""}},
		{name: "super long seed", args: args{seed: "supercalifragilisticexpialidocious"}, want: "supercalif-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenReadableName(tt.args.seed)
			if strings.HasPrefix(got, "-") {
				t.Errorf("GenReadableName() = %v, want %q as prefix, can not get strings starting with -", got, tt.want)
			}
			if tt.args.seed == "" {
				if len(got) == 0 {
					t.Errorf("GenReadableName(%q) = %q, must give some generated string, instead of empty string", tt.args.seed, got)
				}
			}
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("GenReadableName() = %v, want %q as prefix", got, tt.want)
			}
		})
	}
}
