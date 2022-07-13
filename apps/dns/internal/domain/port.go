package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetRecords(ctx context.Context, host string) ([]*Record, error)
	DeleteRecords(ctx context.Context, host string) error
	AddARecords(ctx context.Context, host string, aRecords []string) error
	VerifySite(ctx context.Context, claimId repos.ID) error
	GetSites(ctx context.Context, accountId string) ([]*Site, error)
	GetSite(ctx context.Context, siteId string) (*Site, error)
	GetSiteFromDomain(ctx context.Context, domain string) (*Site, error)
	GetAccountEdgeCName(ctx context.Context, accountId string) (string, error)
	CreateSite(ctx context.Context, domain string, accountId repos.ID) error
	CreateRecord(
		ctx context.Context,
		recordType string,
		host string,
		answer string,
		ttl uint32,
		priority int64,
	) (*Record, error)
}
