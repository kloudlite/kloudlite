package types

import (
	corev1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
)

type PodsMap map[types.NamespacedName]corev1.Pod

func ToPodsMap(pods []corev1.Pod) PodsMap {
	m := make(PodsMap, len(pods))
	for _, pod := range pods {
		m[types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}] = pod
	}
	return m
}

func (p PodsMap) GetPod(namespace, name string) corev1.Pod {
	return p[types.NamespacedName{Name: name, Namespace: namespace}]
}

func (p PodsMap) PodTrackingId(namespace, name string) string {
	if v, ok := p[types.NamespacedName{Name: name, Namespace: namespace}]; ok {
		return v.GetAnnotations()["kloudlite.io/observability.tracking.id"]
	}
	return ""
}

func (p PodsMap) PodAccountName(namespace, name string) string {
	if v, ok := p[types.NamespacedName{Name: name, Namespace: namespace}]; ok {
		return v.GetAnnotations()["kloudlite.io/observability.account.name"]
	}
	return ""
}

func (p PodsMap) PodClusterName(namespace, name string) string {
	if v, ok := p[types.NamespacedName{Name: name, Namespace: namespace}]; ok {
		return v.GetAnnotations()["kloudlite.io/observability.cluster.name"]
	}
	return ""
}
