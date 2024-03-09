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

// func evictPods(clientset *kubernetes.Clientset, nodeName string) {
// 	// This example does not consider DaemonSets, mirror pods, or PodDisruptionBudgets for simplicity.
// 	podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
// 		FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName}).String(),
// 	})
// 	if err != nil {
// 		panic(err.Error())
// 	}
//
// 	for _, pod := range podList.Items {
// 		if pod.Namespace == "kube-system" {
// 			// Skip kube-system pods or add other checks as needed
// 			continue
// 		}
//
// 		eviction := &types.Eviction{
// 			TypeMeta:      metav1.TypeMeta{Kind: "Eviction", APIVersion: "policy/v1beta1"},
// 			ObjectMeta:    metav1.ObjectMeta{Name: pod.Name, Namespace: pod.Namespace},
// 			DeleteOptions: &metav1.DeleteOptions{},
// 		}
//
// 		// Attempt to evict the pod
// 		err := clientset.PolicyV1beta1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
// 		if err != nil {
// 			fmt.Printf("Failed to evict pod %s: %v\n", pod.Name, err)
// 			continue
// 		}
//
// 		fmt.Printf("Pod %s evicted\n", pod.Name)
// 	}
// }
