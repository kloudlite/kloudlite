package beacon

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type Beacon interface {
	TriggerEvent(ctx context.Context, accountId repos.ID, event *AuditLogEvent) error
	TriggerWithUserCtx(ctx context.Context, accountId repos.ID, act EventAction)
}
