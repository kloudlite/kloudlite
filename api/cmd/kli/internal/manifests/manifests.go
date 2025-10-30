package manifests

import _ "embed"

//go:embed crds.yaml
var CRDs string

//go:embed api-server-rbac.yaml
var APIServerRBAC string

//go:embed api-server.yaml
var APIServer string

//go:embed frontend.yaml
var Frontend string
