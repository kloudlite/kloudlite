package domain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/domain/templates"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	message_office_internal "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/kubectl"
	"sigs.k8s.io/yaml"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/api/pkg/wgutils"
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

const (
	DefaultGlobalVPNName = "default"
)

func (d *domain) generateClusterToken(ctx InfraContext, clusterName string) (string, error) {
	tout, err := d.messageOfficeInternalClient.GenerateClusterToken(ctx, &message_office_internal.GenerateClusterTokenIn{
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
	})
	if err != nil {
		return "", errors.NewE(err)
	}

	return tout.ClusterToken, nil
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

	clusterToken, err := d.generateClusterToken(ctx, clusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	secret.Data = map[string][]byte{
		keyClusterToken: []byte(clusterToken),
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

	if cluster.GlobalVPN == nil {
		cluster.GlobalVPN = fn.New(DefaultGlobalVPNName)
	}

	if _, err := d.ensureGlobalVPN(ctx, *cluster.GlobalVPN); err != nil {
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
				// FIXME: once, we allow gcp service account for clusters via UI
				ServiceAccount: clustersv1.GCPServiceAccount{
					Enabled: false,
					Email:   nil,
					Scopes:  nil,
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

	// FIXME: removing public DNS host for now
	gvpnConn, err := d.ensureGlobalVPNConnection(ctx, cluster.Name, *cluster.GlobalVPN)
	if err != nil {
		return nil, errors.NewE(err)
	}

	cluster.Spec.ClusterServiceCIDR = gvpnConn.ClusterCIDR

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

	if err := d.applyHelmKloudliteAgent(ctx, string(tokenScrt.Data[keyClusterToken]), nCluster); err != nil {
		return nil, errors.NewE(err)
	}

	return nCluster, nil
}

func (d *domain) syncKloudliteGatewayDevice(ctx InfraContext, gvpnName string) error {
	// 1. parse deployment template
	b, err := templates.Read(templates.GlobalVPNKloudliteDeviceTemplate)
	if err != nil {
		return errors.NewE(err)
	}

	svcTemplate, err := templates.Read(templates.GatewayServiceTemplate)
	if err != nil {
		return errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return errors.NewE(err)
	}

	gv, err := d.findGlobalVPN(ctx, gvpnName)
	if err != nil {
		return err
	}

	if gv.KloudliteGatewayDevice.Name == "" {
		return nil
	}

	gvpnConns, err := d.listGlobalVPNConnections(ctx, gvpnName)
	if err != nil {
		return err
	}

	klDevice, err := d.findGlobalVPNDevice(ctx, gvpnName, gv.KloudliteGatewayDevice.Name)
	if err != nil {
		return err
	}

	clDevice, err := d.findGlobalVPNDevice(ctx, gvpnName, gv.KloudliteClusterLocalDevice.Name)
	if err != nil {
		return err
	}

	wgParams, deviceHosts, err := d.buildGlobalVPNDeviceWgBaseParams(ctx, gvpnConns, klDevice)
	if err != nil {
		return err
	}

	publicPeers := make([]wgutils.PublicPeer, 0, len(wgParams.PublicPeers))
	for _, p := range wgParams.PublicPeers {
	  if p.PublicKey != clDevice.PublicKey {
	    publicPeers = append(publicPeers, p)
	  }
	}


	deviceSvcHosts := make([]string, 0, len(deviceHosts))
	for k, v := range deviceHosts {
		deviceSvcHosts = append(deviceSvcHosts, fmt.Sprintf("%s=%s", k, v))
	}

  wgParams.PublicPeers = publicPeers
	wgParams.DNS = klDevice.IPAddr
	wgParams.ListenPort = 31820

	dnsServerArgs := make([]string, 0, len(gvpnConns))
	for _, gvpnConn := range gvpnConns {
		if gvpnConn.Spec.GlobalIP != "" {
			dnsServerArgs = append(dnsServerArgs, fmt.Sprintf("%s=%s:53", gvpnConn.Spec.DNSSuffix, gvpnConn.Spec.GlobalIP))
		}
	}

	resourceName := fmt.Sprintf("kloudlite-device-%s", gv.Name)
	resourceNamespace := accNs
	selector := map[string]string{
		"app": resourceName,
	}

	// wgEndpoint := d.env.KloudliteGlobalVPNDeviceHost

	gao, err := d.accountsSvc.GetAccount(ctx, string(ctx.UserId), ctx.AccountName)
	if err != nil {
		return errors.NewE(err)
	}

	gwRegion, ok := d.env.AvailableKloudliteRegions[gao.KloudliteGatewayRegion]
	if !ok {
		return errors.Newf("invalid gateway region %q", gao.KloudliteGatewayRegion)
	}

	wgEndpoint := gwRegion.PublicDNSHost

	c, err := k8s.RestConfigFromKubeConfig([]byte(gwRegion.Kubeconfig))
	if err != nil {
		return errors.NewE(err)
	}

	yc, err := kubectl.NewYAMLClient(c, kubectl.YAMLClientOpts{})
	if err != nil {
		return errors.NewE(err)
	}

	service := &corev1.Service{}

	wgSvcName := fmt.Sprintf("%s-wg", resourceName)

	svcBytes, err := templates.ParseBytes(svcTemplate, templates.GatewayServiceTemplateVars{
		Name:          wgSvcName,
		Namespace:     resourceNamespace,
		WireguardPort: wgParams.ListenPort,
		Selector:      selector,
	})
	if err != nil {
		return errors.NewE(err)
	}

	ctx2, cf := func() (context.Context, context.CancelFunc) {
		if d.env.IsDev {
			return context.WithCancel(ctx)
		}
		return context.WithTimeout(ctx, 5*time.Second)
	}()
	defer cf()

	for {
		if ctx2.Err() != nil {
			return ctx2.Err()
		}
		service, err = yc.Client().CoreV1().Services(resourceNamespace).Get(ctx, wgSvcName, metav1.GetOptions{})
		if err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}
			if _, err := yc.ApplyYAML(ctx, svcBytes); err != nil {
				return errors.NewE(err)
			}
			continue
		}

		if service.Spec.Ports[0].NodePort != 0 {
			wgEndpoint = fmt.Sprintf("%s:%d", wgEndpoint, service.Spec.Ports[0].NodePort)
			break
		}
	}

	if _, err := d.gvpnDevicesRepo.PatchById(ctx, klDevice.Id, repos.Document{
		fc.GlobalVPNDevicePublicEndpoint: wgEndpoint,
	}); err != nil {
		return err
	}

	wgConfig, err := wgutils.GenerateWireguardConfig(*wgParams)
	if err != nil {
		return err
	}

	deploymentBytes, err := templates.ParseBytes(b, templates.GVPNKloudliteDeviceTemplateVars{
		Name:                  resourceName,
		Namespace:             accNs,
		WgConfig:              wgConfig,
		EnableKubeReverseProxy: false,
		KubeReverseProxyImage: d.env.GlobalVPNKubeReverseProxyImage,
		AuthzToken:            d.env.GlobalVPNKubeReverseProxyAuthzToken,
		GatewayDNSServers:     strings.Join(dnsServerArgs, ","),
		GatewayServiceHosts:   strings.Join(deviceSvcHosts, ","),
		WireguardPort:         wgParams.ListenPort,

    KloudliteAccount: gv.AccountName,
	})
	if err != nil {
		return err
	}

	if _, err := yc.ApplyYAML(ctx, deploymentBytes); err != nil {
		return errors.NewE(err)
	}

	return nil
}

/*
syncKloudliteDeviceOnPlatform:
  - creates a specific device for each global VPN reserved for kloudlite internal use
  - need to use that device as a kube-proxy to all the clusters
  - we can read their logs, and everything on demand
*/
func (d *domain) syncKloudliteDeviceOnPlatform(ctx InfraContext, gvpnName string) error {
	// 1. parse deployment template
	b, err := templates.Read(templates.GlobalVPNKloudliteDeviceTemplate)
	if err != nil {
		return errors.NewE(err)
	}

	svcTemplate, err := templates.Read(templates.GatewayServiceTemplate)
	if err != nil {
		return errors.NewE(err)
	}

	accNs, err := d.getAccNamespace(ctx)
	if err != nil {
		return errors.NewE(err)
	}

	gv, err := d.findGlobalVPN(ctx, gvpnName)
	if err != nil {
		return err
	}

	if gv.KloudliteClusterLocalDevice.Name == "" {
		return nil
	}

	gvpnConns, err := d.listGlobalVPNConnections(ctx, gvpnName)
	if err != nil {
		return err
	}

	clDevice, err := d.findGlobalVPNDevice(ctx, gvpnName, gv.KloudliteClusterLocalDevice.Name)
	if err != nil {
		return err
	}


	wgParams, deviceHosts, err := d.buildGlobalVPNDeviceWgBaseParams(ctx, gvpnConns, clDevice)
	if err != nil {
		return err
	}

	klDevice, err := d.findGlobalVPNDevice(ctx, gvpnName, gv.KloudliteGatewayDevice.Name)
	if err != nil {
		return err
	}

	publicPeers := make([]wgutils.PublicPeer, 0, len(wgParams.PublicPeers))
	for _, p := range wgParams.PublicPeers {
	  if p.PublicKey != klDevice.PublicKey {
	    publicPeers = append(publicPeers, p)
	  }
	}

	deviceSvcHosts := make([]string, 0, len(deviceHosts))
	for k, v := range deviceHosts {
		deviceSvcHosts = append(deviceSvcHosts, fmt.Sprintf("%s=%s", k, v))
	}

  wgParams.PublicPeers = publicPeers
	wgParams.DNS = clDevice.IPAddr
	wgParams.ListenPort = 31820

	dnsServerArgs := make([]string, 0, len(gvpnConns))
	for _, gvpnConn := range gvpnConns {
		if gvpnConn.Spec.GlobalIP != "" {
			dnsServerArgs = append(dnsServerArgs, fmt.Sprintf("%s=%s:53", gvpnConn.Spec.DNSSuffix, gvpnConn.Spec.GlobalIP))
		}
	}

	resourceName := fmt.Sprintf("kloudlite-device-%s", gv.Name)
	resourceNamespace := accNs
	selector := map[string]string{
		"app": resourceName,
	}

	service := &corev1.Service{}
	ctx2, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()

	wgEndpoint := d.env.KloudliteGlobalVPNDeviceHost

	wgSvcName := fmt.Sprintf("%s-wg", resourceName)
	svcBytes, err := templates.ParseBytes(svcTemplate, templates.GatewayServiceTemplateVars{
		Name:          wgSvcName,
		Namespace:     resourceNamespace,
		WireguardPort: wgParams.ListenPort,
		Selector:      selector,
	})
	if err != nil {
		return errors.NewE(err)
	}

	for {
		if ctx2.Err() != nil {
			return ctx2.Err()
		}
		if err := d.k8sClient.Get(ctx, fn.NN(resourceNamespace, wgSvcName), service); err != nil {
			if !apiErrors.IsNotFound(err) {
				return err
			}

			if err := d.k8sClient.ApplyYAML(ctx, svcBytes); err != nil {
				return errors.NewE(err)
			}

			continue
		}

		if service.Spec.Ports[0].NodePort != 0 {
			wgEndpoint = fmt.Sprintf("%s:%d", wgEndpoint, service.Spec.Ports[0].NodePort)
			break
		}
	}

	if _, err := d.gvpnDevicesRepo.PatchById(ctx, clDevice.Id, repos.Document{
		fc.GlobalVPNDevicePublicEndpoint: wgEndpoint,
	}); err != nil {
		return err
	}

	wgConfig, err := wgutils.GenerateWireguardConfig(*wgParams)
	if err != nil {
		return err
	}

	deploymentBytes, err := templates.ParseBytes(b, templates.GVPNKloudliteDeviceTemplateVars{
		Name:                  resourceName,
		Namespace:             accNs,
		WgConfig:              wgConfig,
		EnableKubeReverseProxy: true,
		KubeReverseProxyImage: d.env.GlobalVPNKubeReverseProxyImage,
		AuthzToken:            d.env.GlobalVPNKubeReverseProxyAuthzToken,
		GatewayDNSServers:     strings.Join(dnsServerArgs, ","),
		GatewayServiceHosts:   strings.Join(deviceSvcHosts, ","),
		WireguardPort:         wgParams.ListenPort,

    KloudliteAccount: gv.AccountName,
	})
	if err != nil {
		return err
	}

	if err := d.k8sClient.ApplyYAML(ctx, deploymentBytes); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) applyHelmKloudliteAgent(ctx InfraContext, clusterToken string, cluster *entities.Cluster) error {
	b, err := templates.Read(templates.HelmKloudliteAgent)
	if err != nil {
		return errors.NewE(err)
	}

	values := map[string]any{
		"account-name": ctx.AccountName,

		"cluster-name":  cluster.Name,
		"cluster-token": clusterToken,

		"kloudlite-release":        d.env.KloudliteRelease,
		"message-office-grpc-addr": d.env.MessageOfficeExternalGrpcAddr,

		"public-dns-host": cluster.Spec.PublicDNSHost,
		"cloudprovider":   cluster.Spec.CloudProvider,
	}

	if cluster.Spec.CloudProvider == ct.CloudProviderGCP {
		var credsSecret corev1.Secret
		if err := d.k8sClient.Get(ctx, fn.NN(cluster.Spec.GCP.CredentialsRef.Namespace, cluster.Spec.GCP.CredentialsRef.Name), &credsSecret); err != nil {
			return err
		}

		m := make(map[string]string)
		for k, v := range credsSecret.Data {
			m[k] = string(v)
		}

		gcpCreds, err := fn.JsonConvert[clustersv1.GCPCredentials](m)
		if err != nil {
			return err
		}

		values["gcp-service-account-json"] = base64.StdEncoding.EncodeToString([]byte(gcpCreds.ServiceAccountJSON))
	}

	b2, err := templates.ParseBytes(b, values)
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
		ClusterName: cluster.Name,
		SyncStatus:  t.GenSyncStatus(t.SyncActionApply, 0),
	}

	hr.IncrementRecordVersion()

	uhr, err := d.upsertHelmRelease(ctx, cluster.Name, &hr)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, cluster.Name, &uhr.HelmChart, uhr.RecordVersion); err != nil {
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

	if err := d.applyHelmKloudliteAgent(ctx, out.ClusterToken, cluster); err != nil {
		return errors.NewE(err)
	}

	if cluster.GlobalVPN != nil {
		gvpn, err := d.findGlobalVPNConnection(ctx, cluster.Name, *cluster.GlobalVPN)
		if err != nil {
			return errors.NewE(err)
		}
		if err := d.applyGlobalVPNConnection(ctx, gvpn); err != nil {
			return errors.NewE(err)
		}
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

	pr, err := d.clusterRepo.FindPaginated(ctx, d.clusterRepo.MergeMatchFilters(f, mf), pagination)
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
		if errors.Is(err, ErrClusterNotFound) {
			byokCluster, err := d.findBYOKCluster(ctx, name)
			if err != nil {
				return nil, err
			}
			return &entities.Cluster{
				Cluster: clustersv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: byokCluster.Name,
					},
					Spec: clustersv1.ClusterSpec{
						ClusterServiceCIDR: byokCluster.ClusterSvcCIDR,
						PublicDNSHost:      "",
					},
				},
				ResourceMetadata: byokCluster.ResourceMetadata,
				AccountName:      byokCluster.AccountName,
				GlobalVPN:        &byokCluster.GlobalVPN,
				LastOnlineAt:     byokCluster.LastOnlineAt,
			}, nil
		}
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

	if _, err := d.consoleClient.ArchiveEnvironmentsForCluster(ctx, &console.ArchiveEnvironmentsForClusterIn{
		UserId:      string(ctx.UserId),
		UserName:    ctx.UserName,
		UserEmail:   ctx.UserEmail,
		AccountName: ctx.AccountName,
		ClusterName: name,
	}); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, ucluster.Name, PublishUpdate)
	if err := d.deleteK8sResource(ctx, &ucluster.Cluster); err != nil {
		if !apiErrors.IsNotFound(err) {
			return errors.NewE(err)
		}

		return d.OnClusterDeleteMessage(ctx, *ucluster)
	}

	return nil
}

func (d *domain) OnClusterDeleteMessage(ctx InfraContext, cluster entities.Cluster) error {
	xcluster, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err = d.clusterRepo.DeleteById(ctx, xcluster.Id); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeCluster, cluster.Name, PublishDelete)

	if xcluster.GlobalVPN != nil {
		if err := d.claimClusterSvcCIDRRepo.DeleteOne(ctx, repos.Filter{
			fc.ClaimClusterSvcCIDRClaimedByCluster: xcluster.Name,
			fc.AccountName:                         ctx.AccountName,
			fc.ClaimClusterSvcCIDRGlobalVPNName:    xcluster.GlobalVPN,
		}); err != nil {
			return errors.NewE(err)
		}

		if _, err := d.freeClusterSvcCIDRRepo.Create(ctx, &entities.FreeClusterSvcCIDR{
			AccountName:    ctx.AccountName,
			GlobalVPNName:  *xcluster.GlobalVPN,
			ClusterSvcCIDR: xcluster.Spec.ClusterServiceCIDR,
		}); err != nil {
			return errors.NewE(err)
		}

		gv, err := d.findGlobalVPNConnection(ctx, xcluster.Name, *xcluster.GlobalVPN)
		if err != nil {
			return errors.NewE(err)
		}

		if err := d.OnGlobalVPNConnectionDeleteMessage(ctx, xcluster.Name, *gv); err != nil {
			return errors.NewE(err)
		}
	}

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

	if cluster.Spec.GCP != nil && cluster.Spec.GCP.VPC != nil {
		patchDoc[fc.ClusterSpecGcpVpc] = cluster.Spec.GCP.VPC
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

func (d *domain) MarkClusterOnlineAt(ctx InfraContext, clusterName string, timestamp *time.Time) error {
	if d.isBYOKCluster(ctx, clusterName) {
		if _, err := d.byokClusterRepo.Patch(ctx, entities.UniqueBYOKClusterFilter(ctx.AccountName, clusterName), repos.Document{
			fc.BYOKClusterLastOnlineAt: timestamp,
		}); err != nil {
			return errors.NewEf(err, "failed to patch last online time for byok cluster %q,", clusterName)
		}
		return nil
	}

	if _, err := d.clusterRepo.Patch(ctx, repos.Filter{
		fc.AccountName:  ctx.AccountName,
		fc.MetadataName: clusterName,
	}, repos.Document{
		fc.ClusterLastOnlineAt: timestamp,
	}); err != nil {
		return errors.NewEf(err, "failed to patch last online time for cluster %q", clusterName)
	}

	return nil
}
