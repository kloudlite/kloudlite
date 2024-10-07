package domain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/helm"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) clusterAlreadyExists(ctx InfraContext, name string) (*bool, error) {
	exists, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		fc.AccountName:  ctx.AccountName,
		fc.MetadataName: name,
	})
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return fn.New(true), nil
	}

	existsBYOK, err := d.byokClusterRepo.FindOne(ctx, repos.Filter{
		fc.AccountName:  ctx.AccountName,
		fc.MetadataName: name,
	})
	if err != nil {
		return nil, err
	}

	if existsBYOK != nil {
		return fn.New(true), nil
	}

	return fn.New(false), nil
}

func (d *domain) CreateBYOKCluster(ctx InfraContext, cluster entities.BYOKCluster) (*entities.BYOKCluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCluster); err != nil {
		return nil, errors.NewE(err)
	}

	exists, err := d.clusterAlreadyExists(ctx, cluster.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if exists != nil && *exists {
		return nil, errors.Newf("cluster/byok cluster with name (%s) already exists", cluster.Name)
	}

	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if s, ok := cluster.GetLabels()[constants.ClusterLabelOwnedBy]; ok {
		if s != string(ctx.UserId) {
			return nil, errors.Newf("provided wrong owner for cluster %q, expected %q", cluster.Name, ctx.UserId)
		}

		cluster.OwnedBy = &s
	}

	cluster.Namespace = accNs

	if cluster.GlobalVPN == "" {
		cluster.GlobalVPN = DefaultGlobalVPNName
	}

	if _, err := d.ensureGlobalVPN(ctx, cluster.GlobalVPN); err != nil {
		return nil, errors.NewE(err)
	}

	ctoken, err := d.generateClusterToken(ctx, cluster.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	cluster.ClusterToken = ctoken

	cluster.MessageQueueTopicName = common.SendToAgentSubjectPrefix(ctx.AccountName, cluster.Name)

	gvpnConn, err := d.ensureGlobalVPNConnection(ctx, ensureGlobalVPNConnectionArgs{
		ClusterName:   cluster.Name,
		GlobalVPNName: cluster.GlobalVPN,
		DispatchAddr:  nil,
		Visibility:    cluster.Visibility,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	cluster.ClusterSvcCIDR = gvpnConn.ClusterCIDR

	existing, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		fields.MetadataName:      cluster.Name,
		fields.MetadataNamespace: cluster.Namespace,
		fields.AccountName:       ctx.AccountName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, ErrClusterAlreadyExists{ClusterName: cluster.Name, AccountName: ctx.AccountName}
	}

	cluster.AccountName = ctx.AccountName

	cluster.IncrementRecordVersion()
	cluster.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	cluster.LastUpdatedBy = cluster.CreatedBy

	cluster.SyncStatus = t.GenSyncStatus(t.SyncActionApply, 0)

	nCluster, err := d.byokClusterRepo.Create(ctx, &cluster)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, errors.NewEf(err, "cluster with name %q already exists in namespace %q", cluster.Name, cluster.Namespace)
		}
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, nCluster.Name, PublishAdd)

	return nCluster, nil
}

func (d *domain) UpdateBYOKCluster(ctx InfraContext, clusterName string, displayName string) (*entities.BYOKCluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCluster); err != nil {
		return nil, errors.NewE(err)
	}

	updated, err := d.byokClusterRepo.PatchOne(ctx, entities.UniqueBYOKClusterFilter(ctx.AccountName, clusterName), repos.Document{
		fc.DisplayName: displayName,
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (d *domain) ListBYOKCluster(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BYOKCluster], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusters); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		fields.AccountName: ctx.AccountName,
		// "$or": []map[string]any{
		// 	{fc.BYOKClusterOwnedBy: ctx.UserId},
		// 	{fc.BYOKClusterOwnedBy: nil},
		// },
	}

	pRecords, err := d.byokClusterRepo.FindPaginated(ctx, d.byokClusterRepo.MergeMatchFilters(f, search), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pRecords, nil
}

func (d *domain) GetBYOKCluster(ctx InfraContext, name string) (*entities.BYOKCluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCluster); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findBYOKCluster(ctx, ctx.AccountName, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

type BYOKSetupInstruction struct {
	Title   string `json:"title"`
	Command string `json:"command"`
}

func (d *domain) GetBYOKClusterSetupInstructions(ctx InfraContext, name string, onlyHelmValues bool) ([]BYOKSetupInstruction, error) {
	cluster, err := d.findBYOKCluster(ctx, ctx.AccountName, name)
	if err != nil {
		return nil, err
	}

	if onlyHelmValues {
		b, err := json.Marshal(map[string]any{
			"crds-url": fmt.Sprintf("https://github.com/kloudlite/helm-charts/releases/download/%s/crds-all.yml", d.env.KloudliteRelease),

			"chart-repo":    "https://kloudlite.github.io/helm-charts",
			"chart-version": d.env.KloudliteRelease,

			"helm-values": map[string]any{
				"accountName":           ctx.AccountName,
				"clusterName":           name,
				"clusterToken":          cluster.ClusterToken,
				"messageOfficeGRPCAddr": d.env.MessageOfficeExternalGrpcAddr,
				"kloudliteDNSSuffix":    fmt.Sprintf("%s.%s", ctx.AccountName, d.env.KloudliteDNSSuffix),
			},
		})
		if err != nil {
			return nil, err
		}

		return []BYOKSetupInstruction{
			{
				Title:   "Helm Values",
				Command: string(b),
			},
		}, nil
	}

	return []BYOKSetupInstruction{
		{Title: "Add Helm Repo", Command: "helm repo add kloudlite https://kloudlite.github.io/helm-charts"},
		{Title: "Update Kloudlite Repo", Command: "helm repo update kloudlite"},
		{Title: "Install kloudlite CRDs", Command: fmt.Sprintf("kubectl apply -f https://github.com/kloudlite/helm-charts/releases/download/%s/crds-all.yml --server-side", d.env.KloudliteRelease)},
		{Title: "Install Kloudlite Agent", Command: fmt.Sprintf(`helm upgrade --install kloudlite --namespace kloudlite --create-namespace kloudlite/kloudlite-agent --version %s --set accountName="%s" --set clusterName="%s" --set clusterToken="%s" --set messageOfficeGRPCAddr="%s" --set kloudliteDNSSuffix="%s"`, d.env.KloudliteRelease, ctx.AccountName, name, cluster.ClusterToken, d.env.MessageOfficeExternalGrpcAddr, fmt.Sprintf("%s.%s", ctx.AccountName, d.env.KloudliteDNSSuffix))},
	}, nil
}

func (d *domain) DeleteBYOKCluster(ctx InfraContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCluster); err != nil {
		return errors.NewE(err)
	}

	cluster, err := d.findBYOKCluster(ctx, ctx.AccountName, name)
	if err != nil {
		return errors.NewE(err)
	}

	if cluster.GlobalVPN != "" {
		if err := d.deleteGlobalVPNConnection(ctx, cluster.Name, cluster.GlobalVPN); err != nil {
			return errors.NewE(err)
		}
	}

	if _, err := d.consoleClient.ArchiveResourcesForCluster(ctx, &console.ArchiveResourcesForClusterIn{
		UserId:      string(ctx.UserId),
		UserName:    ctx.UserName,
		UserEmail:   ctx.UserEmail,
		AccountName: ctx.AccountName,
		ClusterName: name,
	}); err != nil {
		return errors.NewE(err)
	}

	if err := d.byokClusterRepo.DeleteOne(ctx, entities.UniqueBYOKClusterFilter(ctx.AccountName, name)); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) RenderHelmKloudliteAgent(ctx context.Context, accountName string, clusterName string, clusterToken string) ([]byte, error) {
	cluster, err := d.byokClusterRepo.FindOne(ctx, entities.UniqueBYOKClusterFilter(accountName, clusterName))
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return nil, ErrClusterNotFound
	}

	if clusterToken != cluster.ClusterToken {
		return nil, fmt.Errorf("UnAuthorized")
	}

	values, err := json.Marshal(map[string]any{
		"accountName":           accountName,
		"clusterName":           clusterName,
		"clusterToken":          cluster.ClusterToken,
		"messageOfficeGRPCAddr": d.env.MessageOfficeExternalGrpcAddr,
		"kloudliteDNSSuffix":    fmt.Sprintf("%s.%s", accountName, d.env.KloudliteDNSSuffix),
	})
	if err != nil {
		return nil, err
	}

	b, err := d.helmClient.TemplateChart(ctx, &helm.ChartSpec{
		ReleaseName: "kloudlite-agent",
		Namespace:   "kloudlite",
		ChartName:   "kloudlite/kloudlite-agent",
		Version:     d.env.KloudliteRelease,
		ValuesYaml:  string(values),
	})
	if err != nil {
		return nil, err
	}

	namespace := `
apiVersion: v1
kind: Namespace
metadata:
  name: kloudlite
`

	return []byte(fmt.Sprintf("%s\n---\n%s", namespace, b)), nil
}

func (d *domain) findBYOKCluster(ctx context.Context, accountName, clusterName string) (*entities.BYOKCluster, error) {
	cluster, err := d.byokClusterRepo.FindOne(ctx, entities.UniqueBYOKClusterFilter(accountName, clusterName))
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return nil, ErrClusterNotFound
	}
	return cluster, nil
}

func (d *domain) UpsertBYOKClusterKubeconfig(ctx InfraContext, clusterName string, kubeconfig []byte) error {
	byokCluster, err := d.findBYOKCluster(ctx, ctx.AccountName, clusterName)
	if err != nil {
		return err
	}

	if _, err := d.byokClusterRepo.PatchById(ctx, byokCluster.Id, repos.Document{
		fc.BYOKClusterKubeconfig: t.EncodedString{
			Value:    base64.StdEncoding.EncodeToString(kubeconfig),
			Encoding: "base64",
		},
	}); err != nil {
		return err
	}

	return nil
}

func (d *domain) isBYOKCluster(ctx InfraContext, name string) bool {
	cluster, err := d.findBYOKCluster(ctx, ctx.AccountName, name)
	if err != nil {
		return false
	}

	return cluster != nil
}
