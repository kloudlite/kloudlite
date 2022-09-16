package app

import (
	"context"
	"kloudlite.io/apps/dns/internal/domain"
	kldns "kloudlite.io/grpc-interfaces/kloudlite.io/rpc/dns"
)

type dnsServerI struct {
	kldns.UnimplementedDNSServer
	d domain.Domain
}

func (d *dnsServerI) GetAccountDomains(ctx context.Context, in *kldns.GetAccountDomainsIn) (*kldns.GetAccountDomainsOut, error) {
	_sites, err := d.d.GetSites(ctx, in.AccountId)
	if err != nil {
		return nil, err
	}
	return &kldns.GetAccountDomainsOut{
		Domains: func() []string {
			sites := make([]string, 0)
			for _, site := range _sites {
				sites = append(sites, site.Domain)
			}
			return sites
		}(),
	}, nil
}

func fxDNSGrpcServer(d domain.Domain) kldns.DNSServer {
	return &dnsServerI{
		d: d,
	}
}
