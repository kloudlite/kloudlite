package manifests

import _ "embed"

//go:embed crds.yaml
var CRDs string

//go:embed api-server-rbac.yaml
var APIServerRBAC string

//go:embed api-server.yaml
var APIServer string

//go:embed webhooks.yaml
var Webhooks string

//go:embed frontend.yaml
var Frontend string

//go:embed aws-machine-types.yaml
var AWSMachineTypes string

//go:embed kloudlite-ingress-namespace.yaml
var KloudliteIngressNamespace string

//go:embed kloudlite-hostmanager-namespace.yaml
var KloudliteHostmanagerNamespace string

//go:embed image-registry.yaml
var ImageRegistry string
