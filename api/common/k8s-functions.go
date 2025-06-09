package common

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/constants"
	corev1 "k8s.io/api/core/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindNamespaceForAccount(ctx context.Context, k8sCli client.Client, accountName string) (*corev1.Namespace, error) {
	var nsList corev1.NamespaceList
	if err := k8sCli.List(ctx, &nsList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.AccountNameKey: accountName,
		}),
	}); err != nil {
		return nil, err
	}

	if len(nsList.Items) > 1 {
		return nil, fmt.Errorf("multiple namespaces with label (%s: %s) found, there should have been only one", constants.AccountNameKey, accountName)
	}

	return &nsList.Items[0], nil
}
