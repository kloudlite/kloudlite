package templates

import (
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  corev1 "k8s.io/api/core/v1"
)

type DBLifecycleVars struct {
  Metadata metav1.ObjectMeta `json:"metadata"`
  NodeSelector map[string]string `json:"nodeSelector"`
  Tolerations []corev1.Toleration `json:"tolerations"`

  PostgressRootCredentialsSecret string `json:"postgressRootCredentialsSecret"`
  PostgressNewCredentialsSecret string
}
