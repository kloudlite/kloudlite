package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetNodeIps(ctx context.Context, regionPart *string, accountPart *string, clusterPart string) ([]string, error)
	GetRecord(ctx context.Context, host string) (*Record, error)
	DeleteRecords(ctx context.Context, host string) error
	AddARecords(ctx context.Context, host string, aRecords []string) error

	UpsertARecords(ctx context.Context, host string, records []string) error
	VerifySite(ctx context.Context, claimId repos.ID) error
	GetSites(ctx context.Context, accountId string) ([]*Site, error)
	GetVerifiedSites(ctx context.Context, accountId string) ([]*Site, error)
	GetSite(ctx context.Context, siteId string) (*Site, error)
	GetSiteFromDomain(ctx context.Context, domain string) (*Site, error)
	GetAccountEdgeCName(ctx context.Context, accountId string, regionId repos.ID) (string, error)

	CreateSite(ctx context.Context, domain string, accountId, regionId repos.ID) error
	DeleteSite(ctx context.Context, siteId repos.ID) error
	UpdateNodeIPs(
		ctx context.Context,
		regionId string,
		accountId string,
		clusterPart string,
		ips []string,
	) bool
}
