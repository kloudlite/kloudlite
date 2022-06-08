package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetRecords(ctx context.Context, host string) ([]*Record, error)
	CreateSite(ctx context.Context, domain string, accountId repos.ID) (*Verification, error)
	CreateRecord(
		ctx context.Context,
		siteId repos.ID,
		recordType string,
		host string,
		answer string,
		ttl uint32,
		priority int64,
	) (*Record, error)
}
