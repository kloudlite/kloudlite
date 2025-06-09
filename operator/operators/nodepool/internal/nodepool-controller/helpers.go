package nodepool_controller

import (
	"context"
	"encoding/json"
	"sort"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/pkg/constants"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	fn "github.com/kloudlite/operator/pkg/functions"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

func nodesBelongingToNodepool(ctx context.Context, cli client.Client, name string) ([]clustersv1.Node, error) {
	var nodesList clustersv1.NodeList
	if err := cli.List(ctx, &nodesList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.NodePoolNameKey: name,
		}),
	}); err != nil {
		return nil, err
	}

	return nodesList.Items, nil
}

func realNodesBelongingToNodepool(ctx context.Context, cli client.Client, name string) ([]corev1.Node, error) {
	var nodesList corev1.NodeList
	if err := cli.List(ctx, &nodesList, &client.ListOptions{
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			constants.NodePoolNameKey: name,
		}),
	}); err != nil {
		return nil, err
	}

	return nodesList.Items, nil
}

func filterNodesMarkedForDeletion(nodes []clustersv1.Node) map[string]clustersv1.Node {
	nl := make(map[string]clustersv1.Node, len(nodes))
	for i := range nodes {
		if nodes[i].GetDeletionTimestamp() != nil {
			nl[nodes[i].GetName()] = nodes[i]
		}
	}

	return nl
}

func addFinalizersOnNodes(ctx context.Context, cli client.Client, nodes []clustersv1.Node, finalizer string) error {
	for i := range nodes {
		if !controllerutil.ContainsFinalizer(&nodes[i], finalizer) {
			controllerutil.AddFinalizer(&nodes[i], finalizer)
			if err := cli.Update(ctx, &nodes[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func deleteFinalizersOnNodes(ctx context.Context, cli client.Client, nodes []clustersv1.Node, finalizer string) error {
	for i := range nodes {
		controllerutil.RemoveFinalizer(&nodes[i], finalizer)
		if err := cli.Update(ctx, &nodes[i]); err != nil {
			return err
		}
	}
	return nil
}

func deleteNodes(ctx context.Context, cli client.Client, nodes ...clustersv1.Node) error {
	for i := range nodes {
		if err := cli.Delete(ctx, &nodes[i]); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func nodesChecksum(nodesMap map[string]clustersv1.NodeProps) string {
	names := fn.MapKeys(nodesMap)
	sort.Strings(names)

	b, _ := json.Marshal(names)
	return fn.Md5(b)
}
