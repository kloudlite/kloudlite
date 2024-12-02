package reconciler

import (
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetLocal[T any, V Resource](r *Request[V], key string) (T, bool) {
	x := r.locals[key]
	t, ok := x.(T)
	if !ok {
		return *new(T), ok
	}
	return t, ok
}

func SetLocal[T any, V Resource](r *Request[V], key string, value T) {
	if r.locals == nil {
		r.locals = map[string]any{}
	}
	r.locals[key] = value
}

func Get[T client.Object](ctx context.Context, cli client.Client, nn types.NamespacedName, obj T) (T, error) {
	if err := cli.Get(ctx, nn, obj); err != nil {
		// return obj, err
		return *new(T), err
	}
	return obj, nil
}

func GetRaw[T any](ctx context.Context, cli client.Client, nn types.NamespacedName, obj *unstructured.Unstructured) (*T, error) {
	// b, err := json.Marshal(obj)
	// if err != nil {
	// 	return nil, err
	// }
	// var m map[string]any
	// if err := json.Unmarshal(b, &m); err != nil {
	// 	return nil, err
	// }
	//
	// k := unstructured.Unstructured{
	// 	Object: obj,
	// }
	if err := cli.Get(ctx, nn, obj); err != nil {
		return nil, err
	}

	b, err := json.Marshal(obj.Object)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
