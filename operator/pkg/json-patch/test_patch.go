package json_patch

import "testing"

func TestApplyPatch(t *testing.T) {
	type arg struct {
		JsonDoc any
		patch   []PatchOperation
	}

	type test struct {
		name string
		args arg
		want bool
	}

	tests := []test{
		{
			name: "dest is nil and source is nil",
			args: arg{},
			//args: arg{JsonDoc: },
			want: true,
		},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
		})
	}

}
