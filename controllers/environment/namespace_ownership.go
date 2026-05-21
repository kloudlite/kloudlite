package environment

import (
	"fmt"

	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	environmentManagedLabel        = "kloudlite.io/managed"
	environmentNameLabel           = "kloudlite.io/environment"
	environmentNamespaceLabel      = "kloudlite.io/environment-namespace"
	environmentUIDAnnotation       = "kloudlite.io/environment-uid"
	environmentCreatedByAnnotation = "kloudlite.io/created-by"
)

func applyEnvironmentNamespaceOwnership(namespace *corev1.Namespace, environment *environmentsv1.Environment) {
	if namespace.Labels == nil {
		namespace.Labels = map[string]string{}
	}
	if namespace.Annotations == nil {
		namespace.Annotations = map[string]string{}
	}

	namespace.Labels[environmentManagedLabel] = "true"
	namespace.Labels[environmentNameLabel] = environment.Name
	namespace.Labels[environmentNamespaceLabel] = environment.Namespace
	namespace.Annotations[environmentUIDAnnotation] = string(environment.UID)
}

func isEnvironmentNamespaceManagedMetadataKey(key string) bool {
	switch key {
	case environmentManagedLabel,
		environmentNameLabel,
		environmentNamespaceLabel,
		environmentUIDAnnotation,
		environmentCreatedByAnnotation:
		return true
	default:
		return false
	}
}

func validateEnvironmentNamespaceOwnership(namespace *corev1.Namespace, environment *environmentsv1.Environment) error {
	if namespace.Labels[environmentManagedLabel] != "true" {
		return fmt.Errorf("namespace %s is not managed by kloudlite", namespace.Name)
	}
	if namespace.Labels[environmentNameLabel] != environment.Name {
		return fmt.Errorf("namespace %s belongs to environment %q, want %q", namespace.Name, namespace.Labels[environmentNameLabel], environment.Name)
	}
	if namespace.Labels[environmentNamespaceLabel] != environment.Namespace {
		return fmt.Errorf("namespace %s belongs to environment namespace %q, want %q", namespace.Name, namespace.Labels[environmentNamespaceLabel], environment.Namespace)
	}
	if namespace.Annotations[environmentUIDAnnotation] != string(environment.UID) {
		return fmt.Errorf("namespace %s belongs to environment uid %q, want %q", namespace.Name, namespace.Annotations[environmentUIDAnnotation], environment.UID)
	}
	return nil
}
