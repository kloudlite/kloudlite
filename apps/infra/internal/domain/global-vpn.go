package domain

import (
	"fmt"
	"math"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/pkg/iputils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) CreateGlobalVPN(ctx InfraContext, gvpn entities.GlobalVPN) (*entities.GlobalVPN, error) {
	return d.createGlobalVPN(ctx, gvpn)
}

const (
	kloudliteGlobalVPNDevice             = "kloudlite-global-vpn-device"
	kloudliteClusterLocalGlobalVPNDevice = "kloudlite-cluster-local-device"
)

func (d *domain) createGlobalVPN(ctx InfraContext, gvpn entities.GlobalVPN) (*entities.GlobalVPN, error) {
	if gvpn.CIDR == "" {
		gvpn.CIDR = d.env.BaseCIDR
	}

	if gvpn.AllocatableCIDRSuffix == 0 {
		gvpn.AllocatableCIDRSuffix = d.env.AllocatableCIDRSuffix
	}

	if gvpn.NumReservedIPsForNonClusterUse == 0 {
		gvpn.NumReservedIPsForNonClusterUse = d.env.ClustersOffset * int(math.Pow(2, float64(32-gvpn.AllocatableCIDRSuffix)))
		gvpn.NonClusterUseAllowedIPs = make([]string, 0, gvpn.NumReservedIPsForNonClusterUse)
		for i := 0; i < gvpn.NumReservedIPsForNonClusterUse; i++ {
			numIPsPerCluster := int(math.Pow(2, float64(32-gvpn.AllocatableCIDRSuffix)))
			ipv4StartingAddr, err := iputils.GenIPAddr(gvpn.CIDR, i*numIPsPerCluster)
			if err != nil {
				break
			}
			gvpn.NonClusterUseAllowedIPs = append(gvpn.NonClusterUseAllowedIPs, fmt.Sprintf("%s/%d", ipv4StartingAddr, gvpn.AllocatableCIDRSuffix))
		}
	}

	if gvpn.WgInterface == "" {
		gvpn.WgInterface = "kl0"
	}

	gv, err := d.gvpnRepo.Create(ctx, &gvpn)
	if err != nil {
		return nil, err
	}

	device, err := d.createGlobalVPNDevice(ctx, entities.GlobalVPNDevice{
		ObjectMeta: metav1.ObjectMeta{
			Name: kloudliteGlobalVPNDevice,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName:   kloudliteGlobalVPNDevice,
			CreatedBy:     common.CreatedOrUpdatedByKloudlite,
			LastUpdatedBy: common.CreatedOrUpdatedByKloudlite,
		},
		AccountName:    ctx.AccountName,
		GlobalVPNName:  gv.Name,
		PublicEndpoint: nil,
		CreationMethod: kloudliteGlobalVPNDevice,
	})
	if err != nil {
		return nil, err
	}

	clDevice, err := d.createGlobalVPNDevice(ctx, entities.GlobalVPNDevice{
		ObjectMeta: metav1.ObjectMeta{
			Name: kloudliteClusterLocalGlobalVPNDevice,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName:   kloudliteGlobalVPNDevice,
			CreatedBy:     common.CreatedOrUpdatedByKloudlite,
			LastUpdatedBy: common.CreatedOrUpdatedByKloudlite,
		},
		AccountName:    ctx.AccountName,
		GlobalVPNName:  gv.Name,
		PublicEndpoint: nil,
		CreationMethod: kloudliteGlobalVPNDevice,
	})
	if err != nil {
		return nil, err
	}

	return d.gvpnRepo.PatchById(ctx, gv.Id, repos.Document{
		fc.GlobalVPNKloudliteGatewayDeviceName:        device.Name,
		fc.GlobalVPNKloudliteGatewayDeviceIpAddr:      device.IPAddr,
		fc.GlobalVPNKloudliteClusterLocalDeviceName:   clDevice.Name,
		fc.GlobalVPNKloudliteClusterLocalDeviceIpAddr: clDevice.IPAddr,
	})
}

func (d *domain) ensureGlobalVPN(ctx InfraContext, gvpnName string) (*entities.GlobalVPN, error) {
	gvpn, err := d.gvpnRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: gvpnName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if gvpn != nil {
		return gvpn, nil
	}

	return d.createGlobalVPN(ctx, entities.GlobalVPN{
		ObjectMeta: metav1.ObjectMeta{
			Name: gvpnName,
		},
		ResourceMetadata: common.ResourceMetadata{
			DisplayName:   gvpnName,
			CreatedBy:     common.CreatedOrUpdatedByKloudlite,
			LastUpdatedBy: common.CreatedOrUpdatedByKloudlite,
		},
		AccountName: ctx.AccountName,
	})
}

func (d *domain) UpdateGlobalVPN(ctx InfraContext, cgIn entities.GlobalVPN) (*entities.GlobalVPN, error) {
	return nil, errors.New("not implemented")
}

func (d *domain) DeleteGlobalVPN(ctx InfraContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCluster); err != nil {
		return errors.NewE(err)
	}

	filter := repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fc.ClusterGlobalVPN: name,
	}

	cCount, err := d.clusterRepo.Count(ctx, filter)
	if err != nil {
		return errors.NewE(err)
	}
	if cCount != 0 {
		return errors.Newf("delete clusters first, aborting cluster group deletion")
	}

	ucg, err := d.gvpnRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: name,
		},
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeClusterGroup, ucg.Name, PublishUpdate)
	return nil
}

func (d *domain) ListGlobalVPN(ctx InfraContext, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.GlobalVPN], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusters); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}

	pr, err := d.gvpnRepo.FindPaginated(ctx, d.gvpnRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) GetGlobalVPN(ctx InfraContext, name string) (*entities.GlobalVPN, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCluster); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findGlobalVPN(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) findGlobalVPN(ctx InfraContext, gvpnName string) (*entities.GlobalVPN, error) {
	cg, err := d.gvpnRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: gvpnName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cg == nil {
		return nil, ErrClusterNotFound
	}
	return cg, nil
}
