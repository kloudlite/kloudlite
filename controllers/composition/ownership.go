package composition

import (
	compositionsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	DockerCompositionLabel    = "kloudlite.io/docker-composition"
	EnvironmentNamespaceLabel = "kloudlite.io/environment-namespace"
	ManagedLabel              = "kloudlite.io/managed"
)

type NodePlacement struct {
	NodeSelector map[string]string
	Tolerations  []corev1.Toleration
}

func CompositionOwnershipLabels(comp *compositionsv1.Composition, env *compositionsv1.Environment) map[string]string {
	labels := map[string]string{
		DockerCompositionLabel: comp.Name,
		ManagedLabel:           "true",
	}
	if env != nil {
		labels[EnvironmentNamespaceLabel] = env.Namespace
	}
	return labels
}

func ApplyEnvironmentOwnershipLabels(labels map[string]string, env *compositionsv1.Environment) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	labels[DockerCompositionLabel] = env.Name
	labels[EnvironmentNamespaceLabel] = env.Namespace
	return labels
}

func EnvironmentNodePlacement(env *compositionsv1.Environment) NodePlacement {
	if env == nil || env.Spec.NodeName == "" {
		return NodePlacement{}
	}
	return NodePlacement{
		NodeSelector: map[string]string{
			"kubernetes.io/hostname": env.Spec.NodeName,
		},
		Tolerations: []corev1.Toleration{
			{
				Key:      "kloudlite.io/workmachine",
				Operator: corev1.TolerationOpEqual,
				Value:    env.Spec.NodeName,
				Effect:   corev1.TaintEffectNoSchedule,
			},
		},
	}
}
