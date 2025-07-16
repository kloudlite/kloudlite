package e2e_test

import (
	"context"
	"fmt"
	
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	
	v1 "github.com/kloudlite/operator/api/v1"
)

// createTestService creates a test service that routers can reference
func createTestService(ctx context.Context, k8sClient client.Client, namespace, name string) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": fmt.Sprintf("%s-app", name),
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     80,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	
	if err := k8sClient.Create(ctx, service); err != nil {
		return nil, err
	}
	
	return service, nil
}

// getRouter retrieves a router by name and namespace
func getRouter(ctx context.Context, k8sClient client.Client, namespace, name string) (*v1.Router, error) {
	router := &v1.Router{}
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	
	if err := k8sClient.Get(ctx, key, router); err != nil {
		return nil, err
	}
	
	return router, nil
}

// isRouterReady checks if a router has the ready status
func isRouterReady(router *v1.Router) bool {
	return router.Status.IsReady
}