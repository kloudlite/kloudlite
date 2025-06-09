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
		{name: "seed with camelcase", args: args{seed: "helloWorld"}, want: "helloworld-"},
		{name: "seed with hyphen", args: args{seed: "hello-world"}, want: "hello-world"},
		{name: "seed with dot", args: args{seed: "hello.world"}, want: "hello-"},
		{name: "seed with underscore", args: args{seed: "hello_world"}, want: "hello-"},
		{name: "seed with numbers", args: args{seed: "hello123"}, want: "hello123-"},

		// from chatgpt
		{name: "seed with single word", args: args{seed: "hello"}, want: "hello-"},
		{name: "empty seed", args: args{seed: ""}, want: ""},
		{name: "super long seed", args: args{seed: "supercalifragilisticexpialidocious"}, want: "supercalifragilistic-"},
	}
	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GenReadableName(tt.args.seed)
			if got != "" && !IsValidK8sResourceName(got) {
				t.Errorf("got %q, which is not a valid k8s resource name", got)
			}
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("GenReadableName() = %v, want %q as prefix", got, tt.want)
			}
		})
	}
}
