package templates

import (
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  corev1 "k8s.io/api/core/v1"
)

type DBLifecycleVars struct {
  Metadata metav1.ObjectMeta
  NodeSelector map[string]string
  Tolerations []corev1.Toleration

  RootCredentialsSecret string
  NewCredentialsSecret string
}
