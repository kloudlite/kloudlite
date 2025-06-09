package domain

import "context"

type Domain interface {
	PushEvent(ctx context.Context, el *EventLog) (*EventLog, error)
}
