package watch_and_update

import (
	"context"

	t "github.com/kloudlite/operator/operators/resource-watcher/types"
)

type MessageSender interface {
	DispatchResourceUpdates(ctx context.Context, stu t.ResourceUpdate) error
	DispatchInfraUpdates(ctx context.Context, stu t.ResourceUpdate) error
}

