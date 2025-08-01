package templates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	EnableTTYD bool
	ImageTTYD  string

	EnableJupyterNotebook bool
	ImageJupyterNotebook  string

	EnableCodeServer bool
	ImageCodeServer  string

	EnableVSCodeServer bool
	ImageVSCodeServer  string

	ImagePullPolicy string

	KloudliteDeviceFQDN string

	SSHSecretName string
}

type ServiceTemplateArgs struct {
	Metadata                metav1.ObjectMeta
	HeadlessServiceMetadata metav1.ObjectMeta
	Selector                map[string]string

	EnableJupyterNotebook bool
	EnableCodeServer      bool
	EnableTTYD            bool

	TTYDPort       int32
	SSHPort        int32
	NotebookPort   int32
	CodeServerPort int32
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

// type WorkspaceTemplateArgs struct {
// 	Metadata                 metav1.ObjectMeta
// 	WorkMachineName          string
// 	WorkMachineTolerationKey string
//
// 	Paused             bool
// 	ServiceAccountName string
//
// 	ImageInitContainer string
// 	ImageSSH           string
//
// 	EnableTTYD bool
// 	ImageTTYD  string
//
// 	EnableJupyterNotebook bool
// 	ImageJupyterNotebook  string
//
// 	EnableCodeServer bool
// 	ImageCodeServer  string
//
// 	EnableVSCodeServer bool
// 	ImageVscodeServer  string
//
// 	ImagePullPolicy string
//
// 	KloudliteDeviceFQDN string
//
// 	KloudliteDomain string
//
// 	RouterSpec       crdsv1.RouterSpec
// 	PortConfig       PortConfig
// 	IngressClassName string
// }
