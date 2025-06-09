package environments

import (
	"time"

	"github.com/kloudlite/api/apps/console/internal/domain/types"
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	watcher_types "github.com/kloudlite/operator/operators/resource-watcher/types"
)

type CloneEnvironmentArgs struct {
	SourceEnvName      string
	DestinationEnvName string
	DisplayName        string
	EnvRoutingMode     crdsv1.EnvironmentRoutingMode
	ClusterName        string
}

type UpdateAndDeleteOpts struct {
	MessageTimestamp time.Time
	ClusterName      string
}

type Domain interface {
	ListEnvironments(ctx types.ConsoleContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Environment], error)
	GetEnvironment(ctx types.ConsoleContext, name string) (*entities.Environment, error)

	CreateEnvironment(ctx types.ConsoleContext, env entities.Environment) (*entities.Environment, error)
	CloneEnvironment(ctx types.ConsoleContext, args CloneEnvironmentArgs) (*entities.Environment, error)
	UpdateEnvironment(ctx types.ConsoleContext, env entities.Environment) (*entities.Environment, error)
	DeleteEnvironment(ctx types.ConsoleContext, name string) error
	ArchiveEnvironmentsForCluster(ctx types.ConsoleContext, clusterName string) (bool, error)
}

type Sync interface {
	OnEnvironmentApplyError(ctx types.ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error
	OnEnvironmentDeleteMessage(ctx types.ConsoleContext, env entities.Environment) error
	OnEnvironmentUpdateMessage(ctx types.ConsoleContext, env entities.Environment, status watcher_types.ResourceStatus, opts UpdateAndDeleteOpts) error
}
