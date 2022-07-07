package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetRecords(ctx context.Context, host string) ([]*Record, error)
	DeleteRecords(ctx context.Context, host string, siteId string) error
	AddARecords(ctx context.Context, host string, aRecords []string, siteId string) error
	VerifySite(ctx context.Context, vid repos.ID) error
	GetVerification(ctx context.Context, accountId repos.ID, siteId repos.ID) (*Verification, error)
	GetVerifications(ctx context.Context, accountId repos.ID) ([]*Verification, error)
	GetSites(ctx context.Context, accountId string) ([]*Site, error)
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
