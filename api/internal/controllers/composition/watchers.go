package composition

import (
	"context"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// findCompositionsForConfigMap triggers reconciliation of all Compositions when env-config changes
func (r *CompositionReconciler) findCompositionsForConfigMap(ctx context.Context, obj client.Object) []reconcile.Request {
	configMap := obj.(*corev1.ConfigMap)

	// Only trigger for env-config ConfigMap
	if configMap.Name != envConfigConfigMapName {
		return []reconcile.Request{}
	}

	return r.listCompositionsInNamespace(ctx, configMap.Namespace)
}

// findCompositionsForSecret triggers reconciliation of all Compositions when env-secret changes
func (r *CompositionReconciler) findCompositionsForSecret(ctx context.Context, obj client.Object) []reconcile.Request {
	secret := obj.(*corev1.Secret)

	// Only trigger for env-secret Secret
	if secret.Name != envSecretSecretName {
		return []reconcile.Request{}
	}

	return r.listCompositionsInNamespace(ctx, secret.Namespace)
}

// findCompositionsForEnvironment triggers reconciliation of all Compositions when environment activation changes
func (r *CompositionReconciler) findCompositionsForEnvironment(ctx context.Context, obj client.Object) []reconcile.Request {
	environment := obj.(*compositionsv1.Environment)

	// Get the target namespace for this environment
	targetNamespace := environment.Spec.TargetNamespace
	if targetNamespace == "" {
		return []reconcile.Request{}
	}

	return r.listCompositionsInNamespace(ctx, targetNamespace)
}

// findCompositionsForPod triggers reconciliation when a pod's status changes
// This enables faster feedback for ImagePullBackOff, CrashLoopBackOff, etc.
func (r *CompositionReconciler) findCompositionsForPod(ctx context.Context, obj client.Object) []reconcile.Request {
	pod := obj.(*corev1.Pod)

	// Check if this pod belongs to a composition by looking at labels
	compositionName, ok := pod.Labels[dockerCompositionLabel]
	if !ok {
		return []reconcile.Request{}
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      compositionName,
				Namespace: pod.Namespace,
			},
		},
	}
}

// listCompositionsInNamespace lists all compositions in a given namespace and returns reconcile requests
func (r *CompositionReconciler) listCompositionsInNamespace(ctx context.Context, namespace string) []reconcile.Request {
	compositionList := &compositionsv1.CompositionList{}
	if err := r.List(ctx, compositionList, client.InNamespace(namespace)); err != nil {
		return []reconcile.Request{}
	}

	// Pre-allocate requests slice with known capacity
	requests := make([]reconcile.Request, 0, len(compositionList.Items))
	for _, composition := range compositionList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      composition.Name,
				Namespace: composition.Namespace,
			},
		})
	}

	return requests
}
