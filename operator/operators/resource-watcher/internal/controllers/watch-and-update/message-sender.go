package watch_and_update

import (
	"context"
	"log/slog"

	t "github.com/kloudlite/operator/operators/resource-watcher/types"
)

type MessageSenderContext struct {
	context.Context
	logger *slog.Logger
}

type MessageSender interface {
	DispatchConsoleResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
	DispatchInfraResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
	DispatchContainerRegistryResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
	DispatchIotConsoleResourceUpdates(ctx MessageSenderContext, stu t.ResourceUpdate) error
}
