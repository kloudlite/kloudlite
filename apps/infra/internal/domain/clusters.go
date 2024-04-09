package domain

import (
	"encoding/json"
	"fmt"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/domain/templates"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	message_office_internal "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"sigs.k8s.io/yaml"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
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

var ErrClusterNotFound error = fmt.Errorf("cluster not found")

func (e ErrClusterAlreadyExists) Error() string {
	return fmt.Sprintf("cluster with name %q already exists for account: %s", e.ClusterName, e.AccountName)
}

func (d *domain) createTokenSecret(ctx InfraContext, clusterName string, clusterNamespace string) (*corev1.Secret, error) {
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

	secret.Data = map[string][]byte{
		keyClusterToken: []byte(tout.ClusterToken),
	}

	return secret, nil
}

func (d *domain) GetClusterAdminKubeconfig(ctx InfraContext, clusterName string) (*string, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCluster); err != nil {
		return nil, errors.NewE(err)
	}

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

func (d *domain) applyCluster(ctx InfraContext, cluster *entities.Cluster) error {
	addTrackingId(&cluster.Cluster, cluster.Id)
	return d.applyK8sResource(ctx, &cluster.Cluster, cluster.RecordVersion)
}

func (d *domain) CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCluster); err != nil {
		return nil, errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return nil, errors.NewE(err)
	}

	cluster.EnsureGVK()
	cluster.Namespace = accNs

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
	out, err := d.accountsSvc.GetAccount(ctx, string(ctx.UserId), ctx.AccountName)
	if err != nil {
		return nil, errors.NewEf(err, "failed to get account %q", ctx.AccountName)
	}

	cluster.Spec.AccountId = out.AccountId

	tokenScrt, err := d.createTokenSecret(ctx, cluster.Name, cluster.Namespace)
	if err != nil {
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

			cps, err := d.findProviderSecret(ctx, cluster.Spec.AWS.Credentials.SecretRef.Name)
			if err != nil {
				return nil
			}

			return &clustersv1.AWSClusterConfig{
				Credentials: clustersv1.AwsCredentials{
					AuthMechanism: cps.AWS.AuthMechanism,
					SecretRef: ct.SecretRef{
						Name:      cps.Name,
						Namespace: cps.Namespace,
					},
				},

				Region: cluster.Spec.AWS.Region,
				K3sMasters: clustersv1.AWSK3sMastersConfig{
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
									Role:             "primary-master",
									KloudliteRelease: d.env.KloudliteRelease,
								},
							}
						}
						return map[string]clustersv1.MasterNodeProps{
							"master-1": {
								Role:             "primary-master",
								KloudliteRelease: d.env.KloudliteRelease,
							},
							"master-2": {
								Role:             "secondary-master",
								KloudliteRelease: d.env.KloudliteRelease,
							},
							"master-3": {
								Role:             "secondary-master",
								KloudliteRelease: d.env.KloudliteRelease,
							},
						}
					}(),
				},
			}
		}(),
		GCP: func() *clustersv1.GCPClusterConfig {
			if cluster.Spec.CloudProvider != ct.CloudProviderGCP {
				return nil
			}

			cps, err := d.findProviderSecret(ctx, cluster.Spec.GCP.CredentialsRef.Name)
			if err != nil {
				return nil
			}

			var gcpServiceAccountJSON struct {
				ProjectID string `json:"project_id"`
			}

			if cps.GCP != nil {
				if err := json.Unmarshal([]byte(cps.GCP.ServiceAccountJSON), &gcpServiceAccountJSON); err != nil {
					return nil
				}
			}

			return &clustersv1.GCPClusterConfig{
				Region:       cluster.Spec.GCP.Region,
				GCPProjectID: gcpServiceAccountJSON.ProjectID,
				CredentialsRef: ct.SecretRef{
					Name:      cps.Name,
					Namespace: cps.Namespace,
				},
				MasterNodes: clustersv1.GCPMasterNodesConfig{
					RootVolumeType: "pd-ssd",
					RootVolumeSize: 50,
					Nodes: func() map[string]clustersv1.MasterNodeProps {
						if cluster.Spec.AvailabilityMode == "dev" {
							return map[string]clustersv1.MasterNodeProps{
								"master-1": {
									Role:             "primary-master",
									AvailabilityZone: fmt.Sprintf("%s-a", cluster.Spec.GCP.Region), // defaults to {{.region}}-a zone
									KloudliteRelease: d.env.KloudliteRelease,
								},
							}
						}
						return map[string]clustersv1.MasterNodeProps{
							"master-1": {
								Role:             "primary-master",
								AvailabilityZone: fmt.Sprintf("%s-a", cluster.Spec.GCP.Region),
								KloudliteRelease: d.env.KloudliteRelease,
							},
							"master-2": {
								Role:             "secondary-master",
								AvailabilityZone: fmt.Sprintf("%s-a", cluster.Spec.GCP.Region),
								KloudliteRelease: d.env.KloudliteRelease,
							},
							"master-3": {
								Role:             "secondary-master",
								AvailabilityZone: fmt.Sprintf("%s-a", cluster.Spec.GCP.Region),
								KloudliteRelease: d.env.KloudliteRelease,
							},
						}
					}(),
				},
			}
		}(),
		MessageQueueTopicName: common.GetTenantClusterMessagingTopic(ctx.AccountName, cluster.Name),
		KloudliteRelease:      d.env.KloudliteRelease,
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
	cluster.SyncStatus = t.GenSyncStatus(t.SyncActionApply, 0)

	if err := d.k8sClient.ValidateObject(ctx, &cluster.Cluster); err != nil {
		return nil, errors.NewE(err)
	}

	nCluster, err := d.clusterRepo.Create(ctx, &cluster)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, errors.Newf("cluster with name %q already exists in namespace %q", cluster.Name, cluster.Namespace)
		}
		return nil, errors.NewE(err)
	}

	if err := d.applyCluster(ctx, nCluster); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, nCluster.Name, PublishAdd)

	if err := d.applyHelmKloudliteAgent(ctx, nCluster.Name, string(tokenScrt.Data[keyClusterToken]), nCluster.Spec.PublicDNSHost, string(nCluster.Spec.CloudProvider)); err != nil {
		return nil, errors.NewE(err)
	}

	return nCluster, nil
}

func (d *domain) applyHelmKloudliteAgent(ctx InfraContext, clusterName string, clusterToken string, clusterPublicHost string, cloudprovider string) error {
	b, err := templates.Read(templates.HelmKloudliteAgent)
	if err != nil {
		return errors.NewE(err)
	}

	b2, err := templates.ParseBytes(b, map[string]any{
		"account-name": ctx.AccountName,

		"cluster-name":  clusterName,
		"cluster-token": clusterToken,

		"kloudlite-release":        d.env.KloudliteRelease,
		"message-office-grpc-addr": d.env.MessageOfficeExternalGrpcAddr,

		"public-dns-host": clusterPublicHost,
		"cloudprovider":   cloudprovider,
	})
	if err != nil {
		return errors.NewE(err)
	}

	var m map[string]any
	if err := yaml.Unmarshal(b2, &m); err != nil {
		return errors.NewE(err)
	}

	helmChart, err := fn.JsonConvert[crdsv1.HelmChart](m)
	if err != nil {
		return errors.NewE(err)
	}

	hr := entities.HelmRelease{
		HelmChart: helmChart,
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: fmt.Sprintf("kloudlite agent %s", d.env.KloudliteRelease),
			CreatedBy: common.CreatedOrUpdatedBy{
				UserId:    "kloudlite-platform",
				UserName:  "kloudlite-platform",
				UserEmail: "kloudlite-platform",
			},
			LastUpdatedBy: common.CreatedOrUpdatedBy{
				UserId:    "kloudlite-platform",
				UserName:  "kloudlite-platform",
				UserEmail: "kloudlite-platform",
			},
		},
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
		SyncStatus:  t.GenSyncStatus(t.SyncActionApply, 0),
	}

	hr.IncrementRecordVersion()

	uhr, err := d.upsertHelmRelease(ctx, clusterName, &hr)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &uhr.HelmChart, uhr.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) UpgradeHelmKloudliteAgent(ctx InfraContext, clusterName string) error {
	out, err := d.messageOfficeInternalClient.GetClusterToken(ctx, &message_office_internal.GetClusterTokenIn{
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
	})
	if err != nil {
		return errors.NewE(err)
	}

	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.applyHelmKloudliteAgent(ctx, clusterName, out.ClusterToken, cluster.Spec.PublicDNSHost, string(cluster.Spec.CloudProvider)); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) ListClusters(ctx InfraContext, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusters); err != nil {
		return nil, errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		fields.AccountName:       ctx.AccountName,
		fields.MetadataNamespace: accNs,
	}

	pr, err := d.clusterRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

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

	return c, nil
}

func (d *domain) UpdateCluster(ctx InfraContext, clusterIn entities.Cluster) (*entities.Cluster, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCluster); err != nil {
		return nil, errors.NewE(err)
	}
	clusterIn.EnsureGVK()

	uCluster, err := d.clusterRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: clusterIn.Name,
		},
		common.PatchForUpdate(ctx, &clusterIn),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyCluster(ctx, uCluster); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, uCluster.Name, PublishUpdate)
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

	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: name,
	}

	npCount, err := d.nodePoolRepo.Count(ctx, filter)
	if err != nil {
		return errors.NewE(err)
	}
	if npCount != 0 {
		return errors.Newf("delete nodepool first, aborting cluster deletion")
	}

	pvCount, err := d.pvRepo.Count(ctx, filter)
	if err != nil {
		return errors.NewE(err)
	}
	if pvCount != 0 {
		return errors.Newf("delete pvs first, aborting cluster deletion")
	}

	ucluster, err := d.clusterRepo.Patch(
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

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, ucluster.Name, PublishUpdate)
	if err := d.deleteK8sResource(ctx, &ucluster.Cluster); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnClusterDeleteMessage(ctx InfraContext, cluster entities.Cluster) error {
	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return errors.NewE(err)
	}
	err = d.clusterRepo.DeleteOne(ctx, repos.Filter{
		fields.AccountName:       ctx.AccountName,
		fields.MetadataName:      cluster.Name,
		fields.MetadataNamespace: accNs,
	})
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, cluster.Name, PublishDelete)

	return nil
}

func (d *domain) OnClusterUpdateMessage(ctx InfraContext, cluster entities.Cluster, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xCluster, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return errors.NewE(err)
	}

	recordVersion, err := d.matchRecordVersion(cluster.Annotations, xCluster.RecordVersion)
	if err != nil {
		return nil
	}

	patchDoc := repos.Document{}
	if cluster.Spec.Output != nil {
		patchDoc[fc.ClusterSpecOutput] = cluster.Spec.Output
	}

	if cluster.Spec.AWS != nil && cluster.Spec.AWS.VPC != nil {
		patchDoc[fc.ClusterSpecAwsVpc] = cluster.Spec.AWS.VPC
	}

	uCluster, err := d.clusterRepo.PatchById(
		ctx,
		xCluster.Id,
		common.PatchForSyncFromAgent(&cluster, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
			XPatch:           patchDoc,
		}))
	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, uCluster.GetName(), PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) findCluster(ctx InfraContext, clusterName string) (*entities.Cluster, error) {
	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return nil, errors.NewE(err)
	}

	cluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:       ctx.AccountName,
		fields.MetadataName:      clusterName,
		fields.MetadataNamespace: accNs,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return nil, ErrClusterNotFound
	}
	return cluster, nil
}
