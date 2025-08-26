package templates

import (
	v1 "github.com/kloudlite/operator/api/v1"
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

type JumpServerDeploymentTemplateArgs struct {
	Metadata       metav1.ObjectMeta
	SelectorLabels map[string]string

	SSH v1.WorkmachineSSH

	ImageSSHServer string

	WorkMachineName          string
	WorkMachineTolerationKey string
}
