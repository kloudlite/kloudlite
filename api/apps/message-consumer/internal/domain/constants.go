package domain

const (
	JOB_SERVICE_ACCOUNT = "hotspot-cluster-svc-account"
	JOB_IMAGE           = "harbor.dev.madhouselabs.io/kloudlite/jobs/jobber:latest"
	JOB_IMAGE_CONFIG    = "harbor.dev.madhouselabs.io/kloudlite/jobs/config:latest"
	JOB_IMAGE_SECRET    = "harbor.dev.madhouselabs.io/kloudlite/jobs/secret:latest"
	JOB_IMAGE_PROJECT   = "harbor.dev.madhouselabs.io/kloudlite/jobs/project:latest"
	JOB_IMAGE_ROUTER    = "harbor.dev.madhouselabs.io/kloudlite/jobs/router:latest"
	JOB_IMAGE_APP       = "harbor.dev.madhouselabs.io/kloudlite/jobs/app:latest"
)

const (
	DEFAULT_IMAGE_PULL_POLICY = "Always"
)

const (
	RESOURCE_CONFIG           = "config"
	RESOURCE_SECRET           = "secret"
	RESOURCE_PROJECT          = "project"
	RESOURCE_MANAGED_SERVICE  = "mService"
	RESOURCE_MANAGED_RESOURCE = "mResource"
)

var ResourceImageMap = map[string]string{
	RESOURCE_CONFIG: JOB_IMAGE_CONFIG,
}

type ops struct {
	Commit string
	Undo   string
}

const (
	ACTION_CREATE = "create"
	ACTION_UPDATE = "update"
	ACTION_DELETE = "delete"
)

var ActionOpsMap = map[string]ops{
	ACTION_CREATE: ops{Commit: "commitCreate", Undo: "undoCreate"},
	ACTION_UPDATE: ops{Commit: "commitUpdate", Undo: "undoUpdate"},
	ACTION_DELETE: ops{Commit: "commitDelete", Undo: "undoDelete"},
}

var ReverseActionMap = map[string]string{
	"create": "delete",
	"delete": "create",
	"update": "update",
}

var ResourceActionUndoMap = map[string]map[string]bool{
	RESOURCE_CONFIG:           map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: true},
	RESOURCE_SECRET:           map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: true},
	RESOURCE_PROJECT:          map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: false},
	RESOURCE_MANAGED_SERVICE:  map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: false},
	RESOURCE_MANAGED_RESOURCE: map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: false},
}
