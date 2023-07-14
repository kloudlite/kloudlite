package utils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func Drain(kubeconfig []byte, nodeName string) error {

	clientset, err := getClientSet(kubeconfig)
	if err != nil {
		return err
	}

	// Cordon the node
	if err := cordonNode(clientset, nodeName); err != nil {
		return fmt.Errorf("Error cordoning node: %s\n", err.Error())
	}

	// Delete all pods on the node
	if err := deletePodsOnNode(clientset, nodeName); err != nil {
		return fmt.Errorf("Error deleting pods on node: %s\n", err.Error())
	}

	return nil
}

func cordonNode(clientset *kubernetes.Clientset, nodeName string) error {
	ctx := context.TODO()
	node, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if node.Spec.Unschedulable {
		fmt.Printf("Node '%s' is already cordoned\n", nodeName)
		return nil
	}

	node.Spec.Unschedulable = true
	_, err = clientset.CoreV1().Nodes().Update(ctx, node, v1.UpdateOptions{})
	return err
}

func deletePodsOnNode(clientset *kubernetes.Clientset, nodeName string) error {
	ctx := context.TODO()
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		err = clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteNode(kubeconfig []byte, nodeName string) error {

	clientset, err := getClientSet(kubeconfig)
	if err != nil {
		return err
	}

	// Delete the node
	err = clientset.CoreV1().Nodes().Delete(context.TODO(), nodeName, v1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func getClientSet(kubeconfig []byte) (*kubernetes.Clientset, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
