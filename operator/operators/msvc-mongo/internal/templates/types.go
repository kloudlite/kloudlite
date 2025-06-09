package templates

import (
	corev1 "k8s.io/api/core/v1"
)

type DBLifecycleVars struct {
	NodeSelector map[string]string
	Tolerations  []corev1.Toleration

	RootCredentialsSecret string
	NewCredentialsSecret  string
}
