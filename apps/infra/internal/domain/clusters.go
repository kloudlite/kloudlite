package domain

import (
	"fmt"
	"time"

	message_office_internal "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	ct "github.com/kloudlite/operator/apis/common-types"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	keyClusterToken = "cluster-token"
)

type ErrClusterAlreadyExists struct {
	ClusterName string
	AccountName string
}

func (e ErrClusterAlreadyExists) Error() string {
	return fmt.Sprintf("cluster with name %q already exists for account: %s", e.ClusterName, e.AccountName)
}

func (d *domain) createTokenSecret(ctx InfraContext, ps *entities.CloudProviderSecret, clusterName string, clusterNamespace string) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: clusterNamespace,
		},
	}

	tout, err := d.messageOfficeInternalClient.GenerateClusterToken(ctx, &message_office_internal.GenerateClusterTokenIn{
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	secret.StringData = map[string]string{
		keyClusterToken: tout.ClusterToken,
	}

	return secret, nil
}

func (d *domain) GetClusterAdminKubeconfig(ctx InfraContext, clusterName string) (*string, error) {
	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster.Spec.Output == nil {
		return fn.New(""), nil
	}

	kscrt := corev1.Secret{}
	if err := d.k8sClient.Get(ctx.Context, fn.NN(cluster.Namespace, cluster.Spec.Output.SecretName), &kscrt); err != nil {
		return nil, errors.NewE(err)
	}

	kubeconfig, ok := kscrt.Data[cluster.Spec.Output.KeyKubeconfig]
	if !ok {
		return nil, errors.Newf("kubeconfig key %q not found in secret %q", cluster.Spec.Output.KeyKubeconfig, cluster.Spec.Output.SecretName)
	}

	return fn.New(string(kubeconfig)), nil
}

func (d *domain) CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCluster); err != nil {
		return nil, errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	cluster.EnsureGVK()
	cluster.Namespace = accNs

	cps, err := d.findProviderSecret(ctx, cluster.Spec.CredentialsRef.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cps.IsMarkedForDeletion() {
		return nil, errors.Newf("cloud provider secret %q is marked for deletion, aborting cluster creation", cps.Name)
	}

	existing, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		"metadata.name":      cluster.Name,
		"metadata.namespace": cluster.Namespace,
		"accountName":        ctx.AccountName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, ErrClusterAlreadyExists{ClusterName: cluster.Name, AccountName: ctx.AccountName}
	}

	cluster.AccountName = ctx.AccountName
	out, err := d.accountsSvc.GetAccount(ctx, string(ctx.UserId), ctx.AccountName)
	if err != nil {
		return nil, errors.NewEf(err, "failed to get account %q", ctx.AccountName)
	}

	cluster.Spec.AccountId = out.AccountId

	if cluster.Spec.CredentialsRef.Namespace == "" {
		cluster.Spec.CredentialsRef.Namespace = cps.Namespace
	}

	tokenScrt, err := d.createTokenSecret(ctx, cps, cluster.Name, cluster.Namespace)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.k8sClient.ValidateObject(ctx, &cluster.Cluster); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, tokenScrt, 1); err != nil {
		return nil, errors.NewE(err)
	}

	cluster.Spec = clustersv1.ClusterSpec{
		AccountName: ctx.AccountName,
		AccountId:   out.AccountId,
		ClusterTokenRef: ct.SecretKeyRef{
			Name:      tokenScrt.Name,
			Namespace: tokenScrt.Namespace,
			Key:       keyClusterToken,
		},
		CredentialsRef: cluster.Spec.CredentialsRef,
		CredentialKeys: &clustersv1.CloudProviderCredentialKeys{
			KeyAWSAccountId:              entities.AWSAccountId,
			KeyAWSAssumeRoleExternalID:   entities.AWSAssumeRoleExternalId,
			KeyAWSAssumeRoleRoleARN:      entities.AWAssumeRoleRoleARN,
			KeyAWSIAMInstanceProfileRole: entities.AWSInstanceProfileName,
			KeyAccessKey:                 entities.AccessKey,
			KeySecretKey:                 entities.SecretKey,
		},
		AvailabilityMode: cluster.Spec.AvailabilityMode,

		// PublicDNSHost is <cluster-name>.<account-name>.tenants.<public-dns-host-suffix>
		PublicDNSHost:          fmt.Sprintf("%s.%s.tenants.%s", cluster.Name, ctx.AccountName, d.env.PublicDNSHostSuffix),
		ClusterInternalDnsHost: fn.New("cluster.local"),
		CloudflareEnabled:      fn.New(true),
		TaintMasterNodes:       true,
		BackupToS3Enabled:      false,

		CloudProvider: cluster.Spec.CloudProvider,
		AWS: func() *clustersv1.AWSClusterConfig {
			if cluster.Spec.CloudProvider != ct.CloudProviderAWS {
				return nil
			}
			return &clustersv1.AWSClusterConfig{
				Region: cluster.Spec.AWS.Region,
				K3sMasters: clustersv1.AWSK3sMastersConfig{
					ImageId:          "ami-06d146e85d1709abb",
					ImageSSHUsername: "ubuntu",
					InstanceType:     cluster.Spec.AWS.K3sMasters.InstanceType,
					NvidiaGpuEnabled: cluster.Spec.AWS.K3sMasters.NvidiaGpuEnabled,
					RootVolumeType:   "gp3",
					RootVolumeSize: func() int {
						if cluster.Spec.AWS.K3sMasters.NvidiaGpuEnabled {
							return 80
						}
						return 50
					}(),
					IAMInstanceProfileRole: &cps.AWS.CfParamInstanceProfileName,
					Nodes: func() map[string]clustersv1.MasterNodeProps {
						if cluster.Spec.AvailabilityMode == "dev" {
							return map[string]clustersv1.MasterNodeProps{
								"master-1": {
									Role: "primary-master",
								},
							}
						}
						return map[string]clustersv1.MasterNodeProps{
							"master-1": {
								Role: "primary-master",
							},
							"master-2": {
								Role: "secondary-master",
							},
							"master-3": {
								Role: "secondary-master",
							},
						}
					}(),
				},
			}
		}(),
		// MessageQueueTopicName: fmt.Sprintf("kl-acc-%s-clus-%s", ctx.AccountName, cluster.Name),
		MessageQueueTopicName: common.GetTenantClusterMessagingTopic(ctx.AccountName, cluster.Name),
		KloudliteRelease:      "v1.0.5-nightly",
		Output:                nil,
	}

	cluster.IncrementRecordVersion()
	cluster.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	cluster.LastUpdatedBy = cluster.CreatedBy

	cluster.Spec.AccountId = out.AccountId
	cluster.Spec.AccountName = ctx.AccountName
	cluster.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cluster.RecordVersion)

	nCluster, err := d.clusterRepo.Create(ctx, &cluster)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, errors.Newf("cluster with name %q already exists in namespace %q", cluster.Name, cluster.Namespace)
		}
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, &nCluster.Cluster, nCluster.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nCluster, nil
}

func (d *domain) ListClusters(ctx InfraContext, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusters); err != nil {
		return nil, errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.namespace": accNs,
	}

	pr, err := d.clusterRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	edges := make([]repos.RecordEdge[*entities.Cluster], 0, len(pr.Edges))

	for i := range pr.Edges {
		c, found, err := d.readClusterK8sResource(ctx, pr.Edges[i].Node.Namespace, pr.Edges[i].Node.Name)
		if err != nil {
			return nil, errors.NewE(err)
		}

		if found {
			pr.Edges[i].Node.Status = c.Status
		}

		if !found && pr.Edges[i].Node.MarkedForDeletion != nil && *pr.Edges[i].Node.MarkedForDeletion {
			if err := d.clusterRepo.DeleteById(ctx, pr.Edges[i].Node.Id); err != nil {
				return nil, errors.NewE(err)
			}
			continue
		}
		edges = append(edges, pr.Edges[i])
	}

	pr.Edges = edges
	return pr, nil
}

func (d *domain) GetCluster(ctx InfraContext, name string) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCluster); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findCluster(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	clus, found, err := d.readClusterK8sResource(ctx, c.Namespace, c.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if found {
		c.Status = clus.Status
	}

	if !found && c.MarkedForDeletion != nil && *c.MarkedForDeletion {
		if err := d.clusterRepo.DeleteById(ctx, c.Id); err != nil {
			return nil, errors.NewE(err)
		}
		return nil, nil
	}

	return c, nil
}

func (d *domain) UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCluster); err != nil {
		return nil, errors.NewE(err)
	}
	cluster.EnsureGVK()
	clus, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if clus.IsMarkedForDeletion() {
		return nil, errors.Newf("cluster %q in namespace %q is marked for deletion, could not perform any update operation", clus.Name, clus.Namespace)
	}

	cps, err := d.findProviderSecret(ctx, cluster.Spec.CredentialsRef.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cps.IsMarkedForDeletion() {
		return nil, errors.Newf("cloud provider secret %q is marked for deletion, aborting cluster update", cps.Name)
	}

	cluster.Spec.CredentialsRef.Namespace = cps.Namespace

	clus.IncrementRecordVersion()
	clus.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	clus.Spec = cluster.Spec
	clus.Labels = cluster.Labels
	clus.Annotations = cluster.Annotations
	clus.SyncStatus = t.GenSyncStatus(t.SyncActionApply, clus.RecordVersion)

	uCluster, err := d.clusterRepo.UpdateById(ctx, clus.Id, clus)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, &uCluster.Cluster, uCluster.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return uCluster, nil
}

func (d *domain) readClusterK8sResource(ctx InfraContext, namespace string, name string) (cluster *clustersv1.Cluster, found bool, err error) {
	var clus entities.Cluster
	if err := d.k8sClient.Get(ctx, fn.NN(namespace, name), &clus.Cluster); err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, false, nil
		}
	}

	return &clus.Cluster, true, nil
}

func (d *domain) DeleteCluster(ctx InfraContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCluster); err != nil {
		return errors.NewE(err)
	}
	c, err := d.findCluster(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	if c.MarkedForDeletion == nil || *c.MarkedForDeletion {
		c.MarkedForDeletion = fn.New(true)
		c.SyncStatus = t.GetSyncStatusForDeletion(c.Generation)
		upC, err := d.clusterRepo.UpdateById(ctx, c.Id, c)
		if err != nil {
			return errors.NewE(err)
		}

		return d.deleteK8sResource(ctx, &upC.Cluster)
	}

	return nil
}

func (d *domain) OnDeleteClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return errors.NewE(err)
	}

	return d.clusterRepo.DeleteOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.name":      cluster.Name,
		"metadata.namespace": accNs,
	})
}

func (d *domain) OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	c, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.matchRecordVersion(cluster.Annotations, c.RecordVersion); err != nil {
		return nil
	}

	c.Labels = cluster.Labels
	c.Annotations = cluster.Annotations
	c.Spec = cluster.Spec

	c.SyncStatus.LastSyncedAt = time.Now()
	c.SyncStatus.Error = nil
	c.SyncStatus.RecordVersion = c.RecordVersion
	c.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent

	c.Status = cluster.Status

	_, err = d.clusterRepo.UpdateById(ctx, c.Id, c)
	return errors.NewE(err)
}

func (d *domain) findCluster(ctx InfraContext, clusterName string) (*entities.Cluster, error) {
	accNs, err := d.getAccNamespace(ctx, ctx.AccountName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	cluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"metadata.name":      clusterName,
		"metadata.namespace": accNs,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return nil, errors.Newf("cluster with name %q not found", clusterName)
	}
	return cluster, nil
}
