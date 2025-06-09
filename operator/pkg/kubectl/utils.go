package kubectl

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PaginatedList[T client.Object](ctx context.Context, kcli client.Client, list client.ObjectList, opts *client.ListOptions) (chan T, error) {
	if opts == nil {
		opts = &client.ListOptions{}
	}

	if opts.Limit == 0 {
		opts.Limit = 10
	}

	resultsCh := make(chan T)

	go func() {
		defer close(resultsCh)
		for {
			if err := kcli.List(ctx, list, opts); err != nil {
				fmt.Printf("[ERROR] listing resources: %v", err)
				return
			}

			meta.EachListItem(list, func(obj runtime.Object) error {
				o, ok := obj.(T)
				if !ok {
					return nil
				}
				resultsCh <- o
				return nil
			})

			opts.Continue = list.GetContinue()
			if opts.Continue == "" {
				break
			}
		}
	}()

	return resultsCh, nil
}
