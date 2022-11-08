package domain

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	"go.uber.org/fx"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type domainI struct {
	recordsRepo       repos.DbRepo[*Record]
	sitesRepo         repos.DbRepo[*Site]
	recordsCache      cache.Repo[[]*Record]
	accountCNamesRepo repos.DbRepo[*AccountCName]
	regionCNamesRepo  repos.DbRepo[*RegionCName]
	nodeIpsRepo       repos.DbRepo[*NodeIps]
	env               *Env
	consoleClient     console.ConsoleClient
	financeClient     finance.FinanceClient
}

func (d *domainI) UpsertARecords(ctx context.Context, host string, records []string) error {
	err := d.deleteRecords(ctx, host)
	if err != nil {
		return err
	}
	return d.addARecords(ctx, host, records)
}

func (d *domainI) UpdateNodeIPs(ctx context.Context, regionId string, accountId string, clusterPart string, ips []string) bool {
	accountCname, err := d.getAccountCName(ctx, accountId)
	if err != nil {
		return false
	}
	regionCname, err := d.getRegionCName(ctx, regionId)
	if err != nil {
		return false
	}
	one, err := d.nodeIpsRepo.FindOne(
		ctx, repos.Filter{
			"regionPart":  regionCname,
			"accountPart": accountCname,
			"clusterPart": clusterPart,
		},
	)
	if err != nil {
		return false
	}
	if one == nil {
		_, err = d.nodeIpsRepo.Create(
			ctx, &NodeIps{
				RegionPart:  regionCname,
				AccountPart: accountCname,
				ClusterPart: clusterPart,
				Ips:         ips,
			},
		)
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
	all, err := d.nodeIpsRepo.Find(
		ctx, repos.Query{
			Filter: filter,
		},
	)
	out := make([]string, 0)
	for _, nodeIps := range all {
		out = append(out, nodeIps.Ips...)
	}
	if len(out) == 0 && regionPart != nil {
		// TODO: (abdhesh), check if this is really required ?
		result, e := d.regionCNamesRepo.FindOne(
			ctx, repos.Filter{
				"cName": regionPart,
			},
		)
		if e == nil && result.IsShared {
			filter := repos.Filter{
				"regionPart": regionPart,
			}
			all, e2 := d.nodeIpsRepo.Find(
				ctx, repos.Query{
					Filter: filter,
				},
			)
			if e2 == nil {
				for _, nodeIps := range all {
					out = append(out, nodeIps.Ips...)
				}
			}
		}
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
	_, err = d.consoleClient.SetupAccount(
		ctx, &console.AccountSetupIn{
			AccountId: string(site.AccountId),
		},
	)
	if err != nil {
		return err
	}
	return err
}

func (d *domainI) GetSiteFromDomain(ctx context.Context, domain string) (*Site, error) {
	one, err := d.sitesRepo.FindOne(
		ctx, repos.Filter{
			"host": domain,
		},
	)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("site not found")
	}
	return one, nil
}

func (d *domainI) GetSites(ctx context.Context, accountId string) ([]*Site, error) {
	return d.sitesRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"accountId": accountId,
			},
		},
	)
}

func (d *domainI) GetVerifiedSites(ctx context.Context, accountId string) ([]*Site, error) {
	return d.sitesRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"accountId": accountId,
				"verified":  true,
			},
		},
	)
}

func (d *domainI) CreateSite(ctx context.Context, domain string, accountId, regionId repos.ID) error {
	one, err := d.sitesRepo.FindOne(
		ctx, repos.Filter{
			"host":      domain,
			"accountId": accountId,
		},
	)
	if err != nil {
		return err
	}
	if one != nil {
		return errors.New("site already exists")
	}
	if one == nil {
		_, err = d.sitesRepo.Create(
			ctx, &Site{
				Domain:    domain,
				AccountId: accountId,
				RegionId:  regionId,
				Verified:  false,
			},
		)
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
	if err != nil {
		fmt.Println(err)
		return errors.New("Unable to verify CName. Please wait for a while and try again.")
	}

	fmt.Println(site.Domain, cname)
	accountCnameIdentity, err := d.getAccountCName(ctx, string(site.AccountId))
	if err != nil {
		return err
	}

	regionCnameIdentity, err := d.getRegionCName(ctx, string(site.RegionId))
	if err != nil {
		return err
	}

	accountId := string(site.AccountId)

	cluster, err := d.getClusterFromAccount(ctx, err, accountId)
	if err != nil {
		return err
	}

	if cname != fmt.Sprintf("%s.%s.%s.%s.", regionCnameIdentity, accountCnameIdentity, cluster, d.env.EdgeCnameBaseDomain) {
		return errors.New("cname does not match")
	}
	err = d.sitesRepo.UpdateMany(
		ctx, repos.Filter{
			"host": site.Domain,
		}, map[string]any{
			"verified": false,
		},
	)
	if err != nil {
		return err
	}
	site.Verified = true
	_, err = d.sitesRepo.UpdateById(ctx, site.Id, site)
	if err != nil {
		return err
	}
	_, err = d.consoleClient.SetupAccount(
		ctx, &console.AccountSetupIn{
			AccountId: string(site.AccountId),
		},
	)
	if err != nil {
		return err
	}
	return err
}

func (d *domainI) getClusterFromAccount(ctx context.Context, err error, accountId string) (*finance.GetAttachedClusterOut, error) {
	cluster, err := d.financeClient.GetAttachedCluster(
		ctx, &finance.GetAttachedClusterIn{
			AccountId: accountId,
		},
	)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (d *domainI) GetSite(ctx context.Context, siteId string) (*Site, error) {
	return d.sitesRepo.FindById(ctx, repos.ID(siteId))
}

func (d *domainI) GetAccountEdgeCName(ctx context.Context, accountId string, regionId repos.ID) (string, error) {
	name, err := d.getAccountCName(ctx, accountId)
	if err != nil {
		return "", err
	}

	regionCnameIdentity, err := d.getRegionCName(ctx, string(regionId))
	if err != nil {
		return "", err
	}

	cluster, err := d.getClusterFromAccount(ctx, err, accountId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s.%s.%s", regionCnameIdentity, name, cluster, d.env.EdgeCnameBaseDomain), nil
}

func generateName() string {
	randomAdjective := ADJECTIVES[rand.Intn(len(ADJECTIVES))]
	randomNoun := NOUNS[rand.Intn(len(NOUNS))]
	return fmt.Sprintf("%v-%v", randomAdjective, randomNoun)
}

func (d *domainI) getRegionCName(ctx context.Context, regionId string) (string, error) {
	regionDNS, err := d.regionCNamesRepo.FindOne(
		ctx, repos.Filter{
			"regionId": regionId,
		},
	)
	if err != nil {
		return "", err
	}
	if regionDNS != nil {
		return regionDNS.CName, nil
	}
	var genUniqueName func() (string, error)
	genUniqueName = func() (string, error) {
		name := generateName()
		regionDNS, err = d.regionCNamesRepo.FindOne(
			ctx, repos.Filter{
				"cName": name,
			},
		)
		if err != nil {
			return "", err
		}
		if regionDNS != nil {
			return genUniqueName()
		}
		return name, nil
	}

	generatedName, err := genUniqueName()
	if err != nil {
		return "", err
	}
	if regionDNS == nil {
		create, err := d.regionCNamesRepo.Create(
			ctx, &RegionCName{
				RegionId: repos.ID(regionId),
				CName:    generatedName,
			},
		)
		if err != nil {
			return "", err
		}
		return create.CName, nil
	}
	return regionDNS.CName, nil
}

func (d *domainI) getAccountCName(ctx context.Context, accountId string) (string, error) {
	accountDNS, err := d.accountCNamesRepo.FindOne(
		ctx, repos.Filter{
			"accountId": accountId,
		},
	)
	if err != nil {
		return "", err
	}
	if err == nil && accountDNS != nil {
		return accountDNS.CName, nil
	}
	var genUniqueName func() (string, error)
	genUniqueName = func() (string, error) {
		name := generateName()
		accountDNS, err = d.accountCNamesRepo.FindOne(
			ctx, repos.Filter{
				"cName": name,
			},
		)
		if err != nil {
			return "", err
		}
		if accountDNS != nil {
			return genUniqueName()
		}
		return name, nil
	}

	generatedName, err := genUniqueName()
	if err != nil {
		return "", err
	}
	if accountDNS == nil {
		create, err := d.accountCNamesRepo.Create(
			ctx, &AccountCName{
				AccountId: repos.ID(accountId),
				CName:     generatedName,
			},
		)
		if err != nil {
			return "", err
		}
		return create.CName, nil
	}
	return accountDNS.CName, nil
}

func (d *domainI) GetRecord(ctx context.Context, host string) (*Record, error) {
	one, err := d.recordsRepo.FindOne(
		ctx, repos.Filter{
			"host": host,
		},
	)
	if err != nil {
		return nil, err
	}
	return one, nil
}

func (d *domainI) deleteRecords(ctx context.Context, host string) error {
	d.recordsCache.Drop(ctx, host)
	return d.recordsRepo.DeleteMany(
		ctx, repos.Filter{
			"host": host,
		},
	)
}

func (d *domainI) DeleteRecords(ctx context.Context, host string) error {
	return d.deleteRecords(ctx, host)
}

func (d *domainI) addARecords(ctx context.Context, host string, aRecords []string) error {
	var err error
	d.recordsCache.Drop(ctx, host)
	_, err = d.recordsRepo.Create(
		ctx, &Record{
			Host:    host,
			Answers: aRecords,
		},
	)
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
	regionDNSRepo repos.DbRepo[*RegionCName],
	recordsCache cache.Repo[[]*Record],
	consoleclient console.ConsoleClient,
	financeClient finance.FinanceClient,
	env *Env,
) Domain {
	return &domainI{
		recordsRepo,
		sitesRepo,
		recordsCache,
		accountDNSRepo,
		regionDNSRepo,
		nodeIpsRepo,
		env,
		consoleclient,
		financeClient,
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
