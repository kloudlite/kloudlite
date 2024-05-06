package environment

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func findResourceBelongingToEnvironment(ctx context.Context, kclient client.Client, resources client.ObjectList, envTargetNamespace string) error {
	if err := kclient.List(ctx, resources, &client.ListOptions{
		Namespace: envTargetNamespace,
	}); err != nil {
		return err
	}

	return nil
}
