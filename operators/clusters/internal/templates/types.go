package templates

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AWSClusterJobParams struct {
	AccessKeyID     string
	AccessKeySecret string
}

type ClusterJobVars struct {
	JobMetadata  metav1.ObjectMeta
	NodeSelector map[string]string
	Tolerations  []corev1.Toleration
	JobImage     string

	TFWorkspaceName            string
	TFWorkspaceSecretNamespace string

	ClusterSecretName      string
	ClusterSecretNamespace string

	ValuesJSON string

	CloudProvider string
	AWS           *AWSClusterJobParams
}

type AwsVPCJobVars struct {
	JobMetadata  metav1.ObjectMeta
	NodeSelector map[string]string
	Tolerations  []corev1.Toleration
	JobImage     string

	TFWorkspaceName            string
	TFWorkspaceSecretNamespace string

	ValuesJSON string

	AWS                      AWSClusterJobParams
	VPCOutputSecretName      string
	VPCOutputSecretNamespace string
}
