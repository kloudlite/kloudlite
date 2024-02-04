package watch_and_update

import (
	"context"

	t "github.com/kloudlite/operator/operators/resource-watcher/types"
)

type MessageSender interface {
	DispatchConsoleResourceUpdates(ctx context.Context, stu t.ResourceUpdate) error
	DispatchInfraResourceUpdates(ctx context.Context, stu t.ResourceUpdate) error
	DispatchContainerRegistryResourceUpdates(ctx context.Context, stu t.ResourceUpdate) error
}
