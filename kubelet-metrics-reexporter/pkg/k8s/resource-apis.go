package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) GetNode(ctx context.Context, nodename string) (*corev1.Node, error) {
	return c.CoreV1().Nodes().Get(context.TODO(), nodename, metav1.GetOptions{})
}

func (c *Client) ListPodsOnNode(ctx context.Context, nodename string) ([]corev1.Pod, error) {
	pl, err := c.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodename),
	})
	if err != nil {
		return nil, err
	}

	return pl.Items, nil
}
