package domain

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"go.uber.org/fx"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type domainI struct {
	recordsRepo    repos.DbRepo[*Record]
	sitesRepo      repos.DbRepo[*Site]
	siteClaimRepo  repos.DbRepo[*SiteClaim]
	recordsCache   cache.Repo[[]*Record]
	accountDNSRepo repos.DbRepo[*AccountDNS]
}

func (d *domainI) GetSiteFromDomain(ctx context.Context, domain string) (*Site, error) {
	one, err := d.sitesRepo.FindOne(ctx, repos.Filter{
		"host": domain,
	})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("site not found")
	}
	return one, nil
}

func (d *domainI) GetNameServers(ctx context.Context, accountId repos.ID) ([]string, error) {
	return d.GetAccountHostNames(ctx, string(accountId))
}

func (d *domainI) GetSiteClaims(ctx context.Context, accountId repos.ID) ([]*SiteClaim, error) {
	return d.siteClaimRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountId": accountId,
		},
	})
}

func (d *domainI) GetSiteClaim(ctx context.Context, accountId repos.ID, siteId repos.ID) (*SiteClaim, error) {
	return d.siteClaimRepo.FindOne(ctx, repos.Filter{
		"accountId": accountId,
		"siteId":    siteId,
	})
}

func (d *domainI) GetSites(ctx context.Context, accountId string) ([]*Site, error) {
	return d.sitesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountId": accountId,
		},
	})
}

func (d *domainI) CreateSite(ctx context.Context, domain string, accountId repos.ID) error {
	one, err := d.sitesRepo.FindOne(ctx, repos.Filter{
		"domain": domain,
	})
	if err != nil {
		return err
	}
	if one == nil {
		one, err = d.sitesRepo.Create(ctx, &Site{
			Domain: domain,
		})
		if err != nil {
			return err
		}
	}
	_, err = d.siteClaimRepo.Create(ctx, &SiteClaim{
		AccountId: accountId,
		SiteId:    one.Id,
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *domainI) VerifySite(ctx context.Context, claimId repos.ID) error {
	claim, err := d.siteClaimRepo.FindById(ctx, claimId)
	if err != nil {
		return err
	}
	if claim == nil {
		return errors.New("claim not found")
	}
	siteId := claim.SiteId
	accountId := claim.AccountId
	matchedSite, err := d.sitesRepo.FindById(ctx, siteId)
	if err != nil {
		return err
	}
	txtRecords, err := net.LookupTXT(fmt.Sprintf("klcheck.%s", matchedSite.Domain))
	if err != nil {
		return err
	}

	//nsDomainNames := make([]string, 0)
	//for _, txtEntry := range txtRecords {
	//	if txtEntry ==
	//	nsDomainNames = append(nsDomainNames, nsEntry.Host)
	//}
	//sort.Strings(nsDomainNames)

	names, err := d.GetAccountHostNames(ctx, string(accountId))
	verified := false
	for _, txt := range txtRecords {
		if txt == strings.Join(names, ",") {
			verified = true
			break
		}
	}
	//verified := true
	//if len(nsDomainNames) != len(names) {
	//	return errors.New("DNSDomainNamesMismatch")
	//}
	//for i, x := range nsDomainNames {
	//	verified = verified && x == names[i]
	//}
	if !verified {
		return errors.New("DNSDomainNamesMismatch")
	}
	matchedSite.AccountId = accountId
	_, err = d.sitesRepo.UpdateById(ctx, matchedSite.Id, matchedSite)
	if err != nil {
		return err
	}
	err = d.siteClaimRepo.UpdateMany(ctx, &repos.Filter{
		"siteId": siteId,
	}, map[string]any{
		"attached": "false",
	})
	if err != nil {
		return err
	}
	claim.Attached = true
	d.siteClaimRepo.UpdateById(ctx, claimId, claim)
	_, err = d.siteClaimRepo.UpdateById(ctx, claimId, claim)
	if err != nil {
		return err
	}
	return nil
}

func (d *domainI) GetSite(ctx context.Context, siteId string) (*Site, error) {
	return d.sitesRepo.FindById(ctx, repos.ID(siteId))
}

func (d *domainI) GetAccountHostNames(ctx context.Context, accountId string) ([]string, error) {
	accountDNS, err := d.accountDNSRepo.FindOne(ctx, repos.Filter{
		"accountId": accountId,
	})
	if err != nil {
		return nil, err
	}
	if accountDNS == nil {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		name1 := nameGenerator.Generate()
		name2 := nameGenerator.Generate()
		create, err := d.accountDNSRepo.Create(ctx, &AccountDNS{
			AccountId: repos.ID(accountId),
			Hosts: func() []string {
				x := []string{
					fmt.Sprintf("%s.ns.kloudlite.io", name1),
					fmt.Sprintf("%s.ns.kloudlite.io", name2),
				}
				sort.Strings(x)
				return x
			}(),
		})
		if err != nil {
			return nil, err
		}
		return create.Hosts, nil
	}
	return accountDNS.Hosts, nil
}

func (d *domainI) GetRecords(ctx context.Context, host string) ([]*Record, error) {

	if matchedRecords, err := d.recordsCache.Get(ctx, host); err == nil && matchedRecords != nil {
		return matchedRecords, nil
	}

	domainSplits := strings.Split(strings.TrimSpace(host), ".")
	filters := make([]repos.Filter, 0)
	for i := range domainSplits {
		filters = append(filters, repos.Filter{
			"host": strings.Join(domainSplits[i:], "."),
		})
	}
	matchedSites, err := d.sitesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"verified": true,
			"$or":      filters,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(matchedSites) == 0 {
		return nil, errors.New("NoSitesFound")
	}
	var site *Site
	for _, s := range matchedSites {
		if site != nil {
			if len(s.Domain) > len(site.Domain) {
				site = s
			}
		} else {
			site = s
		}
	}

	recordFilters := make([]repos.Filter, 0)

	for i := range domainSplits {
		recordFilters = append(recordFilters, repos.Filter{
			"host": fmt.Sprintf("*.%v", strings.Join(domainSplits[i:], ".")),
		}, repos.Filter{
			"host": strings.Join(domainSplits[i:], "."),
		})
	}

	rec, err := d.recordsRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"siteId": site.Id,
			"$or":    recordFilters,
		},
		Sort: map[string]interface{}{
			"priority": 1,
		},
	})

	if err != nil {
		return nil, err
	}

	err = d.recordsCache.Set(ctx, host, rec)
	if err != nil {
		fmt.Println(err)
	}

	return rec, nil
}

func (d *domainI) CreateRecord(
	ctx context.Context,
	siteId repos.ID,
	recordType string,
	host string,
	answer string,
	ttl uint32,
	priority int64,
) (*Record, error) {
	create, err := d.recordsRepo.Create(ctx, &Record{
		SiteId:   siteId,
		Type:     recordType,
		Host:     host,
		Answer:   answer,
		TTL:      ttl,
		Priority: priority,
	})
	return create, err
}

func (d *domainI) DeleteRecords(ctx context.Context, host string, siteId string) error {

	d.recordsCache.Drop(ctx, host)

	return d.recordsRepo.DeleteMany(ctx, repos.Filter{
		"host": host,
	})
}

func (d *domainI) AddARecords(ctx context.Context, host string, aRecords []string, siteId string) error {
	var err error

	// fmt.Println(aRecords, host, siteId)
	d.recordsCache.Drop(ctx, host)

	for _, aRecord := range aRecords {
		_, err = d.recordsRepo.Create(ctx, &Record{
			SiteId:   repos.ID(siteId),
			Type:     "A",
			Host:     host,
			Answer:   aRecord,
			TTL:      30,
			Priority: 0,
		})

	}
	return err
}

func fxDomain(
	recordsRepo repos.DbRepo[*Record],
	sitesRepo repos.DbRepo[*Site],
	siteClaimRepo repos.DbRepo[*SiteClaim],
	accountDNSRepo repos.DbRepo[*AccountDNS],
	recordsCache cache.Repo[[]*Record],
) Domain {
	return &domainI{
		recordsRepo,
		sitesRepo,
		siteClaimRepo,
		recordsCache,
		accountDNSRepo,
	}
}

var Module = fx.Module(
	"domain",
	fx.Provide(fxDomain),
)
