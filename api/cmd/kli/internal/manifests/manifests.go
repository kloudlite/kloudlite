package manifests

import _ "embed"

//go:embed crds.yaml
var CRDs string

//go:embed api-server-rbac.yaml
var APIServerRBAC string

//go:embed api-server.yaml
var APIServer string

//go:embed frontend-rbac.yaml
var FrontendRBAC string

//go:embed webhooks.yaml
var Webhooks string

//go:embed frontend.yaml
var Frontend string

//go:embed aws-machine-types.yaml
var AWSMachineTypes string

//go:embed azure-machine-types.yaml
var AzureMachineTypes string

//go:embed gcp-machine-types.yaml
var GCPMachineTypes string

//go:embed oci-machine-types.yaml
var OCIMachineTypes string

//go:embed image-registry.yaml
var ImageRegistry string

//go:embed ingress-proxy.yaml
var IngressProxy string

//go:embed local-path-provisioner-config.yaml
var LocalPathProvisionerConfig string

//go:embed local-path-storageclass.yaml
var LocalPathStorageClass string

//go:embed local-path-simple-storageclass.yaml
var LocalPathSimpleStorageClass string
