package reconciler

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Get[T client.Object](ctx context.Context, cli client.Client, nn types.NamespacedName, resource T) error {
	if err := cli.Get(ctx, nn, resource); err != nil {
		// return obj, err
		return err
	}
	return nil
}
