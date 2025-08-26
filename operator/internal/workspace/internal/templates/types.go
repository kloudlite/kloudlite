package templates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerParams struct {
	Enable bool
	Image  string
	Port   int32
}

type StatefulSetTemplateArgs struct {
	Metadata metav1.ObjectMeta

	Selector  map[string]string
	PodLabels map[string]string

	WorkMachineName          string
	WorkMachineTolerationKey string

	Paused             bool
	ServiceAccountName string

	ImageInitContainer string
	ImageSSH           string

	SSHContainer             ContainerParams
	TTYDContainer            ContainerParams
	JupyterNotebookContainer ContainerParams
	VSCodeServerContainer    ContainerParams
	VSCodeTunnelContainer    ContainerParams

	ImagePullPolicy string

	KloudliteDeviceFQDN string

	SSHSecretName string
}

type ServiceTemplateArgs struct {
	Metadata                metav1.ObjectMeta
	HeadlessServiceMetadata metav1.ObjectMeta
	Selector                map[string]string

	EnableJupyterNotebook bool
	EnableVSCodeServer    bool
	EnableVSCodeTunnel    bool
	EnableTTYD            bool

	TTYDPort         int32
	SSHPort          int32
	NotebookPort     int32
	VSCodeServerPort int32
}

type RouterTemplateArgs struct {
	Metadata metav1.ObjectMeta

	WorkMachineName string
	KloudliteDomain string

	EnableJupyterNotebook bool
	EnableCodeServer      bool
	EnableTTYD            bool

	TTYDPort       int32
	NotebookPort   int32
	CodeServerPort int32

	ServiceName string
	ServicePath string
}
