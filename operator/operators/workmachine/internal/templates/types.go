package templates

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkMachineLifecycleVars struct {
	JobMetadata  metav1.ObjectMeta
	NodeSelector map[string]string
	Tolerations  []corev1.Toleration
	JobImage     string

	TFWorkspaceName      string
	TfWorkspaceNamespace string

	CloudProvider string

	ValuesJSON string

	OutputSecretName      string
	OutputSecretNamespace string

	NodeName string
}

type JumpServerDeploymentSpecTemplateArgs struct {
	SSHAuthorizedKeysSecretName string
	SSHAuthorizedKeysSecretKey  string
	SelectorLabels              map[string]string
}
