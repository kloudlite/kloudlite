package node_controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func drainNode(ctx context.Context, cli client.Client, node *corev1.Node) error {
	if !node.Spec.Unschedulable {
		node.Spec.Unschedulable = true
		return cli.Update(ctx, node)
	}
	return nil
}
