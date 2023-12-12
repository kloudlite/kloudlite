package beacon

import (
	"context"
	"github.com/kloudlite/api/pkg/repos"
)

type Beacon interface {
	TriggerEvent(ctx context.Context, accountId repos.ID, event *AuditLogEvent) error
	TriggerWithUserCtx(ctx context.Context, accountId repos.ID, act EventAction)
}
