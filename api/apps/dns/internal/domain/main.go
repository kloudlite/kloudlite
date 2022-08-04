package domain

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/goombaio/namegenerator"
	"go.uber.org/fx"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type domainI struct {
	recordsRepo       repos.DbRepo[*Record]
	sitesRepo         repos.DbRepo[*Site]
	recordsCache      cache.Repo[[]*Record]
	accountCNamesRepo repos.DbRepo[*AccountCName]
	nodeIpsRepo       repos.DbRepo[*NodeIps]
}

func (d *domainI) UpsertARecords(ctx context.Context, host string, records []string) error {
	err := d.deleteRecords(ctx, host)
	if err != nil {
		return err
	}
	return d.AddARecords(ctx, host, records)
}

func (d *domainI) UpdateNodeIPs(ctx context.Context, region string, ips []string) bool {
	one, err := d.nodeIpsRepo.FindOne(ctx, repos.Filter{
		"region": region,
	})
	if err != nil {
		return false
	}
	if one == nil {
		one, err = d.nodeIpsRepo.Create(ctx, &NodeIps{
			Region: region,
			Ips:    ips,
		})
		if err != nil {
			return false
		}
	} else {
		one.Ips = ips
		_, err = d.nodeIpsRepo.UpdateById(ctx, one.Id, one)
		if err != nil {
			return false
		}
	}
	return true
}

func (d *domainI) GetNodeIps(ctx context.Context, region *string) ([]string, error) {
	filter := repos.Filter{}
	if region != nil {
		filter["region"] = *region
	}
	all, err := d.nodeIpsRepo.Find(ctx, repos.Query{
		Filter: filter,
	})
	out := make([]string, 0)
	for _, nodeIps := range all {
		out = append(out, nodeIps.Ips...)
	}
	return out, err
}

func (d *domainI) DeleteSite(ctx context.Context, siteId repos.ID) error {
	return d.sitesRepo.DeleteById(ctx, siteId)
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

func (d *domainI) GetSites(ctx context.Context, accountId string) ([]*Site, error) {
	return d.sitesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountId": accountId,
		},
	})
}

func (d *domainI) CreateSite(ctx context.Context, domain string, accountId repos.ID) error {
	one, err := d.sitesRepo.FindOne(ctx, repos.Filter{
		"domain":    domain,
		"accountId": accountId,
	})
	if err != nil {
		return err
	}
	if one != nil {
		return errors.New("site already exists")
	}
	if one == nil {
		one, err = d.sitesRepo.Create(ctx, &Site{
			Domain:    domain,
			AccountId: accountId,
			Verified:  false,
		})
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *domainI) VerifySite(ctx context.Context, siteId repos.ID) error {
	site, err := d.sitesRepo.FindById(ctx, siteId)
	if err != nil {
		return err
	}
	if site == nil {
		return errors.New("site not found")
	}
	if site.Verified {
		return errors.New("site already verified")
	}
	cname, err := net.LookupCNAME(site.Domain)
	fmt.Println(site.Domain)
	if err != nil {
		return errors.New("Unable to verify CName. Please wait for a while and try again.")
	}
	accountCnameIdentity, err := d.getAccountCName(ctx, string(site.AccountId))
	if err != nil {
		return err
	}

	if cname != fmt.Sprintf("%s.edgenet.khost.dev.", accountCnameIdentity) {
		return errors.New("cname does not match")
	}
	err = d.sitesRepo.UpdateMany(ctx, repos.Filter{
		"host": site.Domain,
	}, map[string]any{
		"verified": false,
	})
	if err != nil {
		return err
	}
	site.Verified = true
	_, err = d.sitesRepo.UpdateById(ctx, site.Id, site)
	return err
}

func (d *domainI) GetSite(ctx context.Context, siteId string) (*Site, error) {
	return d.sitesRepo.FindById(ctx, repos.ID(siteId))
}

func (d *domainI) GetAccountEdgeCName(ctx context.Context, accountId string) (string, error) {
	name, err := d.getAccountCName(ctx, accountId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.edgenet.khost.dev", name), nil
}

func (d *domainI) getAccountCName(ctx context.Context, accountId string) (string, error) {
	accountDNS, err := d.accountCNamesRepo.FindOne(ctx, repos.Filter{
		"accountId": accountId,
	})
	if err != nil {
		return "", err
	}
	if accountDNS == nil {
		seed := time.Now().UTC().UnixNano()
		nameGenerator := namegenerator.NewNameGenerator(seed)
		name1 := nameGenerator.Generate()
		name2 := nameGenerator.Generate()
		create, err := d.accountCNamesRepo.Create(ctx, &AccountCName{
			AccountId: repos.ID(accountId),
			CName:     fmt.Sprintf("%s-%s", name1, name2),
		})
		if err != nil {
			return "", err
		}
		return create.CName, nil
	}
	return accountDNS.CName, nil
}

func (d *domainI) GetRecords(ctx context.Context, host string) ([]*Record, error) {
	find, err := d.recordsRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"host": host,
		},
	})
	if err != nil {
		return nil, err
	}
	return find, nil
}

func (d *domainI) CreateRecord(
	ctx context.Context,
	recordType string,
	host string,
	answer string,
	ttl uint32,
	priority int64,
) (*Record, error) {
	create, err := d.recordsRepo.Create(ctx, &Record{
		Type:     recordType,
		Host:     host,
		Answer:   answer,
		TTL:      ttl,
		Priority: priority,
	})
	return create, err
}

func (d *domainI) deleteRecords(ctx context.Context, host string) error {
	d.recordsCache.Drop(ctx, host)
	return d.recordsRepo.DeleteMany(ctx, repos.Filter{
		"host": host,
	})
}

func (d *domainI) DeleteRecords(ctx context.Context, host string) error {
	return d.deleteRecords(ctx, host)
}

func (d *domainI) addARecords(ctx context.Context, host string, aRecords []string) error {
	var err error
	d.recordsCache.Drop(ctx, host)
	for _, aRecord := range aRecords {
		_, err = d.recordsRepo.Create(ctx, &Record{
			Type:     "A",
			Host:     host,
			Answer:   aRecord,
			TTL:      30,
			Priority: 0,
		})
	}
	return err
}

func (d *domainI) AddARecords(ctx context.Context, host string, aRecords []string) error {
	return d.addARecords(ctx, host, aRecords)
}

func fxDomain(
	recordsRepo repos.DbRepo[*Record],
	sitesRepo repos.DbRepo[*Site],
	nodeIpsRepo repos.DbRepo[*NodeIps],
	accountDNSRepo repos.DbRepo[*AccountCName],
	recordsCache cache.Repo[[]*Record],
) Domain {
	return &domainI{
		recordsRepo,
		sitesRepo,
		recordsCache,
		accountDNSRepo,
		nodeIpsRepo,
	}
}

var Module = fx.Module(
	"domain",
	fx.Provide(fxDomain),
)
