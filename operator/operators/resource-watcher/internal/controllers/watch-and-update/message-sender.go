package watch_and_update

import (
	"context"

	t "github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/logging"
)

type MessageSenderContext struct {
	context.Context
	logger logging.Logger
}

type MessageSender interface {
	DispatchConsoleResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
	DispatchInfraResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
	DispatchContainerRegistryResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
	DispatchIotConsoleResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
}
