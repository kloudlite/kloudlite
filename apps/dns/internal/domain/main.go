package domain

import (
	"context"
	"fmt"
	nanoid "github.com/matoous/go-nanoid/v2"
	"go.uber.org/fx"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"net"
	"strings"
)

type domainI struct {
	recordsRepo repos.DbRepo[*Record]
	sitesRepo   repos.DbRepo[*Site]
	verifyRepo  repos.DbRepo[*Verification]
}

func (d *domainI) GetVerification(ctx context.Context, accountId repos.ID, siteId repos.ID) (*Verification, error) {
	return d.verifyRepo.FindOne(ctx, repos.Filter{
		"accountId": accountId,
		"siteId":    siteId,
	})
}

func (d *domainI) GetSites(ctx context.Context, accountId string) ([]*Site, error) {
	fmt.Println(accountId)
	return d.sitesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountId": accountId,
		},
	})
}

func (d *domainI) CreateSite(ctx context.Context, domain string, accountId repos.ID) (*Verification, error) {
	one, err := d.sitesRepo.FindOne(ctx, repos.Filter{
		"domain": domain,
	})
	if err != nil {
		return nil, err
	}
	if one != nil {
		if one.AccountId == accountId {
			return nil, errors.New("DomainAlreadyExists")
		}

		if err != nil {
			return nil, err
		}

		verifyTxt, err := nanoid.New(32)

		if err != nil {
			return nil, err
		}

		vtext := fmt.Sprintf("kloudlite_verify_%v", verifyTxt)
		verification, err := d.verifyRepo.Create(ctx, &Verification{
			AccountId:  accountId,
			SiteId:     one.Id,
			VerifyText: vtext,
		})

		if err != nil {
			return nil, err
		}

		return verification, nil
	}

	create, err := d.sitesRepo.Create(ctx, &Site{
		AccountId: accountId,
		Domain:    domain,
		Verified:  false,
	})

	if err != nil {
		return nil, err
	}

	verifyTxt, err := nanoid.New(32)

	if err != nil {
		return nil, err
	}

	vtext := fmt.Sprintf("kloudlite_verify_%v", verifyTxt)
	verification, err := d.verifyRepo.Create(ctx, &Verification{
		AccountId:  accountId,
		SiteId:     create.Id,
		VerifyText: vtext,
	})
	if err != nil {
		return nil, err
	}
	return verification, nil
}

func (d *domainI) VerifySite(ctx context.Context, vid repos.ID) error {
	matchedVerificaton, err := d.verifyRepo.FindById(ctx, vid)
	if err != nil {
		return err
	}

	if matchedVerificaton == nil {
		return errors.New("NoVerificationFound")
	}

	matchedSite, err := d.sitesRepo.FindById(ctx, matchedVerificaton.SiteId)
	if err != nil {
		return err
	}

	txts, err := net.LookupTXT(matchedSite.Domain)
	if err != nil {
		return err
	}
	for _, txt := range txts {
		if matchedVerificaton.VerifyText == strings.TrimSpace(txt) {
			matchedSite.Verified = true
			matchedSite.AccountId = matchedVerificaton.AccountId
			_, err := d.sitesRepo.UpdateById(ctx, matchedSite.Id, matchedSite)
			if err != nil {
				return err
			}
			err = d.verifyRepo.DeleteById(ctx, matchedVerificaton.Id)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("NoTxtRecordFound")
}

func (d *domainI) GetRecords(ctx context.Context, host string) ([]*Record, error) {
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

	return d.recordsRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"siteId": site.Id,
			"$or":    recordFilters,
		},
		Sort: map[string]interface{}{
			"priority": 1,
		},
	})

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

	return d.recordsRepo.DeleteMany(ctx, repos.Filter{
		"host": host,
	})
}

func (d *domainI) AddARecords(ctx context.Context, host string, aRecords []string, siteId string) error {
	var err error

	fmt.Println(aRecords, host, siteId)

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
	verifyRepo repos.DbRepo[*Verification],

) Domain {
	return &domainI{
		recordsRepo,
		sitesRepo,
		verifyRepo,
	}
}

var Module = fx.Module(
	"domain",
	fx.Provide(fxDomain),
)
