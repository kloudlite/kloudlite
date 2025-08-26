package templates

import (
	v1 "github.com/kloudlite/operator/api/v1"
)

type IngressTemplateArgs struct {
	CertSecretNamePrefix string
	IngressClassName     string
	HttpsEnabled         bool
	Hosts                []string
	Routes               []v1.RouterRoute
}
