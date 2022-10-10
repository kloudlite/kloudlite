package domain

import (
	"context"
	"fmt"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/pkg/config"
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
	env               *Env
	consoleClient     console.ConsoleClient
}

func (d *domainI) UpsertARecords(ctx context.Context, host string, records []string) error {
	err := d.deleteRecords(ctx, host)
	if err != nil {
		return err
	}
	return d.addARecords(ctx, host, records)
}

func (d *domainI) UpdateNodeIPs(ctx context.Context, regionPart string, accountId string, clusterPart string, ips []string) bool {
	accountCName, err := d.accountCNamesRepo.FindOne(ctx, repos.Filter{
		"accountId": accountId,
	})
	if err != nil {
		return false
	}
	one, err := d.nodeIpsRepo.FindOne(ctx, repos.Filter{
		"regionPart":  regionPart,
		"accountPart": accountCName.CName,
		"clusterPart": clusterPart,
	})
	if err != nil {
		return false
	}
	if one == nil {
		one, err = d.nodeIpsRepo.Create(ctx, &NodeIps{
			RegionPart:  regionPart,
			AccountPart: accountCName.CName,
			ClusterPart: clusterPart,
			Ips:         ips,
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

func (d *domainI) GetNodeIps(ctx context.Context,
	regionPart *string, accountPart *string, clusterPart string,
) ([]string, error) {
	filter := repos.Filter{
		"clusterPart": clusterPart,
	}
	if regionPart != nil {
		filter["regionPart"] = *regionPart
	}
	if accountPart != nil {
		filter["accountPart"] = *accountPart
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
	site, err := d.sitesRepo.FindById(ctx, siteId)
	if err != nil {
		return err
	}
	err = d.sitesRepo.DeleteById(ctx, siteId)
	if err != nil {
		return err
	}
	_, err = d.consoleClient.SetupAccount(ctx, &console.AccountSetupIn{
		AccountId: string(site.AccountId),
	})
	if err != nil {
		return err
	}
	return err
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

func (d *domainI) GetVerifiedSites(ctx context.Context, accountId string) ([]*Site, error) {
	return d.sitesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountId": accountId,
			"verified":  true,
		},
	})
}

func (d *domainI) CreateSite(ctx context.Context, domain string, accountId repos.ID) error {
	one, err := d.sitesRepo.FindOne(ctx, repos.Filter{
		"host":      domain,
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

	if cname != fmt.Sprintf("%s.%s.", accountCnameIdentity, d.env.EdgeCnameBaseDomain) {
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
	if err != nil {
		return err
	}
	_, err = d.consoleClient.SetupAccount(ctx, &console.AccountSetupIn{
		AccountId: string(site.AccountId),
	})
	if err != nil {
		return err
	}
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
	return fmt.Sprintf("%s.%s", name, d.env.EdgeCnameBaseDomain), nil
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

func (d *domainI) GetRecord(ctx context.Context, host string) (*Record, error) {
	one, err := d.recordsRepo.FindOne(ctx, repos.Filter{
		"host": host,
	})
	if err != nil {
		return nil, err
	}
	return one, nil
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
	_, err = d.recordsRepo.Create(ctx, &Record{
		Host:    host,
		Answers: aRecords,
	})
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
	consoleclient console.ConsoleClient,
	env *Env,
) Domain {
	return &domainI{
		recordsRepo,
		sitesRepo,
		recordsCache,
		accountDNSRepo,
		nodeIpsRepo,
		env,
		consoleclient,
	}
}

type Env struct {
	EdgeCnameBaseDomain string `env:"EDGE_CNAME_BASE_DOMAIN" required:"true"`
	MongoUri            string `env:"MONGO_URI" required:"true"`
}

var Module = fx.Module(
	"domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
)
