package templates

import (
	corev1 "k8s.io/api/core/v1"
)

type DBLifecycleVars struct {
	NodeSelector map[string]string   `json:"nodeSelector"`
	Tolerations  []corev1.Toleration `json:"tolerations"`

	PostgressRootCredentialsSecret string `json:"postgressRootCredentialsSecret"`
	PostgressNewCredentialsSecret  string
}
