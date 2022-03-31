package domain

const (
	JOB_SERVICE_ACCOUNT    = "hotspot-cluster-svc-account"
	JOB_IMAGE              = "harbor.dev.madhouselabs.io/kloudlite/jobs/jobber:latest"
	JOB_IMAGE_CONFIG       = "harbor.dev.madhouselabs.io/kloudlite/jobs/config:latest"
	JOB_IMAGE_SECRET       = "harbor.dev.madhouselabs.io/kloudlite/jobs/secret:latest"
	JOB_IMAGE_PROJECT      = "harbor.dev.madhouselabs.io/kloudlite/jobs/project:latest"
	JOB_IMAGE_ROUTER       = "harbor.dev.madhouselabs.io/kloudlite/jobs/router:latest"
	JOB_IMAGE_APP          = "harbor.dev.madhouselabs.io/kloudlite/jobs/app:latest"
	JOB_IMAGE_GIT_PIPELINE = "harbor.dev.madhouselabs.io/kloudlite/jobs/git-pipeline:latest"
)

const (
	DEFAULT_IMAGE_PULL_POLICY = "Always"
)

const (
	RESOURCE_CONFIG           = "config"
	RESOURCE_SECRET           = "secret"
	RESOURCE_APP              = "app"
	RESOURCE_PROJECT          = "project"
	RESOURCE_ROUTER           = "router"
	RESOURCE_MANAGED_SERVICE  = "mService"
	RESOURCE_MANAGED_RESOURCE = "mResource"
	RESOURCE_GIT_PIPELINE     = "gitPipeline"
)

var ResourceImageMap = map[string]string{
	RESOURCE_CONFIG:       JOB_IMAGE_CONFIG,
	RESOURCE_SECRET:       JOB_IMAGE_SECRET,
	RESOURCE_ROUTER:       JOB_IMAGE_ROUTER,
	RESOURCE_APP:          JOB_IMAGE_APP,
	RESOURCE_GIT_PIPELINE: JOB_IMAGE_GIT_PIPELINE,
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
	RESOURCE_APP:              map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: true},
	RESOURCE_PROJECT:          map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: false},
	RESOURCE_MANAGED_SERVICE:  map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: false},
	RESOURCE_MANAGED_RESOURCE: map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: false},
	RESOURCE_GIT_PIPELINE:     map[string]bool{ACTION_CREATE: true, ACTION_UPDATE: true, ACTION_DELETE: true},
}
