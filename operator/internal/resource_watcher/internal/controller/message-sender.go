package controller

import (
	t "github.com/kloudlite/operator/operators/resource-watcher/types"
)

type MessageSender interface {
	DispatchResourceUpdates(ctx Context, ru t.ResourceUpdate) error
}
