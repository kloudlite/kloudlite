package domain

import (
	"context"
	"go.uber.org/fx"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	eventsRepo repos.DbRepo[*EventLog]
}

func (d domain) PushEvent(ctx context.Context, el *EventLog) (*EventLog, error) {
	return d.eventsRepo.Create(ctx, el)
}

var Module = fx.Module("domain",
	fx.Provide(func(eventsRepo repos.DbRepo[*EventLog]) Domain {
		return &domain{eventsRepo: eventsRepo}
	}),
)
