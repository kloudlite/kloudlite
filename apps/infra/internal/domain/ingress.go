package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	networkingv1 "k8s.io/api/networking/v1"
)

func (d *domain) OnIngressUpdateMessage(ctx InfraContext, clusterName string, ingress networkingv1.Ingress, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	for i := range ingress.Spec.Rules {
		domainName := ingress.Spec.Rules[i].Host
		de, err := d.domainEntryRepo.Upsert(ctx, repos.Filter{
			"accountName": ctx.AccountName,
			"clusterName": clusterName,
			"domainName":  domainName,
		}, &entities.DomainEntry{
			ResourceMetadata: common.ResourceMetadata{
				DisplayName:   domainName,
				CreatedBy:     common.CreatedOrUpdatedByResourceSync,
				LastUpdatedBy: common.CreatedOrUpdatedByResourceSync,
			},
			DomainName:  domainName,
			AccountName: ctx.AccountName,
			ClusterName: clusterName,
		})
		if err != nil {
			return err
		}

		d.resourceEventPublisher.PublishDomainResEvent(de, PublishUpdate)
	}

	return nil
}

func (d *domain) OnIngressDeleteMessage(ctx InfraContext, clusterName string, ingress networkingv1.Ingress) error {
	domainNames := make([]any, 0, len(ingress.Spec.Rules))
	for i := range ingress.Spec.Rules {
		domainNames = append(domainNames, ingress.Spec.Rules[i].Host)
	}

	filter := repos.Filter{
		"accountName": ctx.AccountName,
		"clusterName": clusterName,
	}

	filters := d.domainEntryRepo.MergeMatchFilters(filter, map[string]repos.MatchFilter{
		"domainName": {
			MatchType: repos.MatchTypeArray,
			Array:     domainNames,
		},
	})

	err := d.domainEntryRepo.DeleteMany(ctx, filters)
	if err != nil {
		return err
	}

	for i := range domainNames {
		d.resourceEventPublisher.PublishDomainResEvent(&entities.DomainEntry{
			DomainName:  domainNames[i].(string),
			AccountName: ctx.AccountName,
			ClusterName: clusterName,
		}, PublishDelete)
	}
	return nil
}
