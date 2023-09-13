export type Maybe<T> = T;
export type InputMaybe<T> = T;
export type Exact<T extends { [key: string]: unknown }> = {
  [K in keyof T]: T[K];
};
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]?: Maybe<T[SubKey]>;
};
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]: Maybe<T[SubKey]>;
};
export type MakeEmpty<
  T extends { [key: string]: unknown },
  K extends keyof T
> = { [_ in K]?: never };
export type Incremental<T> =
  | T
  | {
      [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never;
    };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string };
  String: { input: string; output: string };
  Boolean: { input: boolean; output: boolean };
  Int: { input: number; output: number };
  Float: { input: number; output: number };
  Date: { input: any; output: any };
  Map: { input: any; output: any };
  Any: { input: any; output: any };
  Json: { input: any; output: any };
  ProviderDetail: { input: any; output: any };
  URL: { input: any; output: any };
};

export type ConsoleResType =
  | 'app'
  | 'config'
  | 'environment'
  | 'managed_resource'
  | 'managed_service'
  | 'project'
  | 'router'
  | 'secret'
  | 'vpn_device'
  | 'workspace';

export type ProjectId = {
  type: ProjectIdType;
  value: Scalars['String']['input'];
};

export type ProjectIdType = 'name' | 'targetNamespace';

export type WorkspaceOrEnvId = {
  type: WorkspaceOrEnvIdType;
  value: Scalars['String']['input'];
};

export type WorkspaceOrEnvIdType =
  | 'environmentName'
  | 'environmentTargetNamespace'
  | 'workspaceName'
  | 'workspaceTargetNamespace';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvType =
  'config' | 'secret';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromType =
  'config' | 'secret';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeType =
  'httpGet' | 'shell' | 'tcp';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeType =
  'httpGet' | 'shell' | 'tcp';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesType =
  'config' | 'secret';

export type Kloudlite_Io__Pkg__Types_SyncStatusAction = 'APPLY' | 'DELETE';

export type Kloudlite_Io__Pkg__Types_SyncStatusState =
  | 'APPLIED_AT_AGENT'
  | 'ERRORED_AT_AGENT'
  | 'IDLE'
  | 'IN_QUEUE'
  | 'RECEIVED_UPDATE_FROM_AGENT';

export type CursorPaginationIn = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  orderBy?: InputMaybe<Scalars['String']['input']>;
  sortDirection?: InputMaybe<CursorPaginationSortDirection>;
};

export type CursorPaginationSortDirection = 'ASC' | 'DESC';

export type SearchApps = {
  text?: InputMaybe<MatchFilterIn>;
};

export type MatchFilterIn = {
  array?: InputMaybe<Array<Scalars['Any']['input']>>;
  exact?: InputMaybe<Scalars['Any']['input']>;
  matchType: MatchFilterMatchType;
  regex?: InputMaybe<Scalars['String']['input']>;
};

export type MatchFilterMatchType = 'array' | 'exact' | 'regex';

export type SearchConfigs = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchWorkspaces = {
  projectName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchImagePullSecrets = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchManagedResources = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchManagedServices = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchProjects = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchRouters = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchSecrets = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchVpnDevices = {
  text?: InputMaybe<MatchFilterIn>;
};

export type ResType = 'cluster' | 'nodepool' | 'providersecret';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode =
  'dev' | 'HA';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider =
  'aws' | 'azure' | 'do' | 'gcp';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode =
  'on_demand' | 'reserved' | 'spot';

export type CloudProviderSecretCloudProviderName =
  | 'aws'
  | 'azure'
  | 'do'
  | 'gcp'
  | 'oci'
  | 'openstack'
  | 'vmware';

export type SearchCluster = {
  cloudProviderName?: InputMaybe<MatchFilterIn>;
  isReady?: InputMaybe<MatchFilterIn>;
  region?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchNodepool = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchProviderSecret = {
  cloudProviderName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type AccountIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  contactEmail: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  isActive?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Scalars['Map']['input'];
};

export type MetadataIn = {
  annotations?: InputMaybe<Scalars['Map']['input']>;
  labels?: InputMaybe<Scalars['Map']['input']>;
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type InvitationIn = {
  accountName: Scalars['String']['input'];
  userEmail?: InputMaybe<Scalars['String']['input']>;
  userName?: InputMaybe<Scalars['String']['input']>;
  userRole: Scalars['String']['input'];
};

export type AppIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecIn = {
  containers: Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersIn>;
  displayName?: InputMaybe<Scalars['String']['input']>;
  freeze?: InputMaybe<Scalars['Boolean']['input']>;
  hpa?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecHpaIn>;
  intercept?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecInterceptIn>;
  nodeSelector?: InputMaybe<Scalars['Map']['input']>;
  region?: InputMaybe<Scalars['String']['input']>;
  replicas?: InputMaybe<Scalars['Int']['input']>;
  serviceAccount?: InputMaybe<Scalars['String']['input']>;
  services?: InputMaybe<
    Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecServicesIn>
  >;
  tolerations?: InputMaybe<
    Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecTolerationsIn>
  >;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersIn =
  {
    args?: InputMaybe<Array<Scalars['String']['input']>>;
    command?: InputMaybe<Array<Scalars['String']['input']>>;
    env?: InputMaybe<
      Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvIn>
    >;
    envFrom?: InputMaybe<
      Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromIn>
    >;
    image: Scalars['String']['input'];
    imagePullPolicy?: InputMaybe<Scalars['String']['input']>;
    livenessProbe?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeIn>;
    name: Scalars['String']['input'];
    readinessProbe?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeIn>;
    resourceCpu?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceCpuIn>;
    resourceMemory?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceMemoryIn>;
    volumes?: InputMaybe<
      Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesIn>
    >;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvIn =
  {
    key: Scalars['String']['input'];
    optional?: InputMaybe<Scalars['Boolean']['input']>;
    refKey?: InputMaybe<Scalars['String']['input']>;
    refName?: InputMaybe<Scalars['String']['input']>;
    type?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvType>;
    value?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromIn =
  {
    refName: Scalars['String']['input'];
    type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromType;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeIn =
  {
    failureThreshold?: InputMaybe<Scalars['Int']['input']>;
    httpGet?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeHttpGetIn>;
    initialDelay?: InputMaybe<Scalars['Int']['input']>;
    interval?: InputMaybe<Scalars['Int']['input']>;
    shell?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeShellIn>;
    tcp?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeTcpIn>;
    type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeType;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeHttpGetIn =
  {
    httpHeaders?: InputMaybe<Scalars['Map']['input']>;
    path: Scalars['String']['input'];
    port: Scalars['Int']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeShellIn =
  {
    command?: InputMaybe<Array<Scalars['String']['input']>>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeTcpIn =
  {
    port: Scalars['Int']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeIn =
  {
    failureThreshold?: InputMaybe<Scalars['Int']['input']>;
    httpGet?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeHttpGetIn>;
    initialDelay?: InputMaybe<Scalars['Int']['input']>;
    interval?: InputMaybe<Scalars['Int']['input']>;
    shell?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeShellIn>;
    tcp?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeTcpIn>;
    type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeType;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeHttpGetIn =
  {
    httpHeaders?: InputMaybe<Scalars['Map']['input']>;
    path: Scalars['String']['input'];
    port: Scalars['Int']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeShellIn =
  {
    command?: InputMaybe<Array<Scalars['String']['input']>>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeTcpIn =
  {
    port: Scalars['Int']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceCpuIn =
  {
    max?: InputMaybe<Scalars['String']['input']>;
    min?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceMemoryIn =
  {
    max?: InputMaybe<Scalars['String']['input']>;
    min?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesIn =
  {
    items?: InputMaybe<
      Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesItemsIn>
    >;
    mountPath: Scalars['String']['input'];
    refName: Scalars['String']['input'];
    type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesType;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesItemsIn =
  {
    fileName?: InputMaybe<Scalars['String']['input']>;
    key: Scalars['String']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecHpaIn = {
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  maxReplicas?: InputMaybe<Scalars['Int']['input']>;
  minReplicas?: InputMaybe<Scalars['Int']['input']>;
  thresholdCpu?: InputMaybe<Scalars['Int']['input']>;
  thresholdMemory?: InputMaybe<Scalars['Int']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecInterceptIn =
  {
    enabled: Scalars['Boolean']['input'];
    toDevice: Scalars['String']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecServicesIn =
  {
    name?: InputMaybe<Scalars['String']['input']>;
    port: Scalars['Int']['input'];
    targetPort?: InputMaybe<Scalars['Int']['input']>;
    type?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecTolerationsIn =
  {
    effect?: InputMaybe<Scalars['String']['input']>;
    key?: InputMaybe<Scalars['String']['input']>;
    operator?: InputMaybe<Scalars['String']['input']>;
    tolerationSeconds?: InputMaybe<Scalars['Int']['input']>;
    value?: InputMaybe<Scalars['String']['input']>;
  };

export type ConfigIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  data?: InputMaybe<Scalars['Map']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
};

export type WorkspaceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_WorkspaceSpecIn>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_WorkspaceSpecIn = {
  isEnvironment?: InputMaybe<Scalars['Boolean']['input']>;
  projectName: Scalars['String']['input'];
  targetNamespace: Scalars['String']['input'];
};

export type ImagePullSecretIn = {
  accountName: Scalars['String']['input'];
  dockerConfigJson?: InputMaybe<Scalars['String']['input']>;
  dockerPassword?: InputMaybe<Scalars['String']['input']>;
  dockerRegistryEndpoint?: InputMaybe<Scalars['String']['input']>;
  dockerUsername?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
};

export type ManagedResourceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecIn =
  {
    inputs?: InputMaybe<Scalars['Map']['input']>;
    mresKind: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMresKindIn;
    msvcRef: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMsvcRefIn;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMresKindIn =
  {
    kind: Scalars['String']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMsvcRefIn =
  {
    apiVersion: Scalars['String']['input'];
    kind?: InputMaybe<Scalars['String']['input']>;
    name: Scalars['String']['input'];
  };

export type ManagedServiceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecIn =
  {
    inputs?: InputMaybe<Scalars['Map']['input']>;
    msvcKind: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecMsvcKindIn;
    nodeSelector?: InputMaybe<Scalars['Map']['input']>;
    region?: InputMaybe<Scalars['String']['input']>;
    tolerations?: InputMaybe<
      Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecTolerationsIn>
    >;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecMsvcKindIn =
  {
    apiVersion: Scalars['String']['input'];
    kind?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecTolerationsIn =
  {
    effect?: InputMaybe<Scalars['String']['input']>;
    key?: InputMaybe<Scalars['String']['input']>;
    operator?: InputMaybe<Scalars['String']['input']>;
    tolerationSeconds?: InputMaybe<Scalars['Int']['input']>;
    value?: InputMaybe<Scalars['String']['input']>;
  };

export type ProjectIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ProjectSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ProjectSpecIn = {
  accountName: Scalars['String']['input'];
  clusterName: Scalars['String']['input'];
  displayName?: InputMaybe<Scalars['String']['input']>;
  logo?: InputMaybe<Scalars['String']['input']>;
  targetNamespace: Scalars['String']['input'];
};

export type RouterIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecIn = {
  backendProtocol?: InputMaybe<Scalars['String']['input']>;
  basicAuth?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecBasicAuthIn>;
  cors?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecCorsIn>;
  domains: Array<Scalars['String']['input']>;
  https?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecHttpsIn>;
  ingressClass?: InputMaybe<Scalars['String']['input']>;
  maxBodySizeInMB?: InputMaybe<Scalars['Int']['input']>;
  rateLimit?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRateLimitIn>;
  region?: InputMaybe<Scalars['String']['input']>;
  routes?: InputMaybe<
    Array<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRoutesIn>
  >;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecBasicAuthIn =
  {
    enabled: Scalars['Boolean']['input'];
    secretName?: InputMaybe<Scalars['String']['input']>;
    username?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecCorsIn = {
  allowCredentials?: InputMaybe<Scalars['Boolean']['input']>;
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  origins?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecHttpsIn =
  {
    clusterIssuer?: InputMaybe<Scalars['String']['input']>;
    enabled: Scalars['Boolean']['input'];
    forceRedirect?: InputMaybe<Scalars['Boolean']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRateLimitIn =
  {
    connections?: InputMaybe<Scalars['Int']['input']>;
    enabled?: InputMaybe<Scalars['Boolean']['input']>;
    rpm?: InputMaybe<Scalars['Int']['input']>;
    rps?: InputMaybe<Scalars['Int']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRoutesIn =
  {
    app?: InputMaybe<Scalars['String']['input']>;
    lambda?: InputMaybe<Scalars['String']['input']>;
    path: Scalars['String']['input'];
    port: Scalars['Int']['input'];
    rewrite?: InputMaybe<Scalars['Boolean']['input']>;
  };

export type SecretIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  data?: InputMaybe<Scalars['Map']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  stringData?: InputMaybe<Scalars['Map']['input']>;
  type?: InputMaybe<Scalars['String']['input']>;
};

export type VpnDeviceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecIn>;
};

export type Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecIn =
  {
    offset: Scalars['Int']['input'];
    ports?: InputMaybe<
      Array<Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecPortsIn>
    >;
    serverName: Scalars['String']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecPortsIn =
  {
    port?: InputMaybe<Scalars['Int']['input']>;
    targetPort?: InputMaybe<Scalars['Int']['input']>;
  };

export type HarborRobotUserIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Artifacts__V1_HarborUserAccountSpecIn>;
};

export type Github_Com__Kloudlite__Operator__Apis__Artifacts__V1_HarborUserAccountSpecIn =
  {
    accountName: Scalars['String']['input'];
    enabled?: InputMaybe<Scalars['Boolean']['input']>;
    harborProjectName: Scalars['String']['input'];
    permissions?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    targetSecret?: InputMaybe<Scalars['String']['input']>;
  };

export type HarborPermission = 'PullRepository' | 'PushRepository';

export type ByocClusterIn = {
  accountName: Scalars['String']['input'];
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ByocSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ByocSpecIn = {
  accountName: Scalars['String']['input'];
  displayName?: InputMaybe<Scalars['String']['input']>;
  incomingKafkaTopic: Scalars['String']['input'];
  ingressClasses?: InputMaybe<Array<Scalars['String']['input']>>;
  provider: Scalars['String']['input'];
  publicIps?: InputMaybe<Array<Scalars['String']['input']>>;
  region: Scalars['String']['input'];
  storageClasses?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type ClusterIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecIn>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecIn =
  {
    accountName: Scalars['String']['input'];
    agentHelmValuesRef?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAgentHelmValuesRefIn>;
    availabilityMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode;
    cloudProvider: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider;
    credentialsRef: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCredentialsRefIn;
    nodeIps?: InputMaybe<Array<Scalars['String']['input']>>;
    operatorsHelmValuesRef?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecOperatorsHelmValuesRefIn>;
    region: Scalars['String']['input'];
    vpc?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAgentHelmValuesRefIn =
  {
    key: Scalars['String']['input'];
    name: Scalars['String']['input'];
    namespace?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCredentialsRefIn =
  {
    name: Scalars['String']['input'];
    namespace?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecOperatorsHelmValuesRefIn =
  {
    key: Scalars['String']['input'];
    name: Scalars['String']['input'];
    namespace?: InputMaybe<Scalars['String']['input']>;
  };

export type NodePoolIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecIn =
  {
    awsNodeConfig?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigIn>;
    labels?: InputMaybe<Scalars['Map']['input']>;
    maxCount: Scalars['Int']['input'];
    minCount: Scalars['Int']['input'];
    taints?: InputMaybe<Array<Scalars['String']['input']>>;
    targetCount: Scalars['Int']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigIn =
  {
    imageId?: InputMaybe<Scalars['String']['input']>;
    isGpu?: InputMaybe<Scalars['Boolean']['input']>;
    onDemandSpecs?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigOnDemandSpecsIn>;
    provisionMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode;
    region?: InputMaybe<Scalars['String']['input']>;
    spotSpecs?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigSpotSpecsIn>;
    vpc?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigOnDemandSpecsIn =
  {
    instanceType: Scalars['String']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigSpotSpecsIn =
  {
    cpuMax: Scalars['Int']['input'];
    cpuMin: Scalars['Int']['input'];
    memMax: Scalars['Int']['input'];
    memMin: Scalars['Int']['input'];
  };

export type CloudProviderSecretIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  cloudProviderName: CloudProviderSecretCloudProviderName;
  data?: InputMaybe<Scalars['Map']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  stringData?: InputMaybe<Scalars['Map']['input']>;
  type?: InputMaybe<Scalars['String']['input']>;
};

export type AccountMembershipIn = {
  accountName: Scalars['String']['input'];
  role: Scalars['String']['input'];
  userId: Scalars['String']['input'];
};

export type EnvOrWorkspaceOrProjectId = {
  name: Scalars['String']['input'];
  type: EnvOrWorkspaceOrProjectIdType;
};

export type EnvOrWorkspaceOrProjectIdType =
  | 'environmentName'
  | 'environmentTargetNamespace'
  | 'projectName'
  | 'projectTargetNamespace'
  | 'workspaceName'
  | 'workspaceTargetNamespace';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecNodeType =
  'cluster' | 'master' | 'worker';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecIn = {
  clusterName?: InputMaybe<Scalars['String']['input']>;
  labels?: InputMaybe<Scalars['Map']['input']>;
  nodePoolName?: InputMaybe<Scalars['String']['input']>;
  nodeType: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecNodeType;
  taints?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type HarborProjectIn = {
  accountName: Scalars['String']['input'];
  credentials: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_HarborCredentialsIn;
  harborProjectName: Scalars['String']['input'];
};

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_HarborCredentialsIn =
  {
    password: Scalars['String']['input'];
    username: Scalars['String']['input'];
  };

export type MembershipIn = {
  accountName: Scalars['String']['input'];
  role: Scalars['String']['input'];
  userId: Scalars['String']['input'];
};

export type NodeIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecIn;
};

export type SearchEnvironments = {
  projectName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type ConsoleAccountCheckNameAvailabilityQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleAccountCheckNameAvailabilityQuery = {
  accounts_checkNameAvailability: {
    result: boolean;
    suggestedNames?: Array<string>;
  };
};

export type ConsoleInfraCheckNameAvailabilityQueryVariables = Exact<{
  resType: ResType;
  name: Scalars['String']['input'];
  clusterName?: InputMaybe<Scalars['String']['input']>;
}>;

export type ConsoleInfraCheckNameAvailabilityQuery = {
  infra_checkNameAvailability: {
    suggestedNames: Array<string>;
    result: boolean;
  };
};

export type ConsoleCoreCheckNameAvailabilityQueryVariables = Exact<{
  resType: ConsoleResType;
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
}>;

export type ConsoleCoreCheckNameAvailabilityQuery = {
  core_checkNameAvailability: {
    result: boolean;
    suggestedNames?: Array<string>;
  };
};

export type ConsoleWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type ConsoleWhoAmIQuery = { auth_me?: { id: string; email: string } };

export type ConsoleCreateAccountMutationVariables = Exact<{
  account: AccountIn;
}>;

export type ConsoleCreateAccountMutation = {
  accounts_createAccount: { displayName: string };
};

export type ConsoleListAccountsQueryVariables = Exact<{ [key: string]: never }>;

export type ConsoleListAccountsQuery = {
  accounts_listAccounts?: Array<{
    id: string;
    updateTime: any;
    displayName: string;
    metadata: { name: string; annotations?: any };
  }>;
};

export type ConsoleGetAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleGetAccountQuery = {
  accounts_getAccount?: {
    updateTime: any;
    contactEmail: string;
    displayName: string;
    metadata: { name: string; annotations?: any };
  };
};

export type ConsoleCreateProjectMutationVariables = Exact<{
  project: ProjectIn;
}>;

export type ConsoleCreateProjectMutation = {
  core_createProject?: { id: string };
};

export type ConsoleGetProjectQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetProjectQuery = {
  core_getProject?: {
    id: string;
    displayName: string;
    creationTime: any;
    clusterName: string;
    apiVersion: string;
    kind: string;
    recordVersion: number;
    updateTime: any;
    accountName: string;
    metadata: {
      namespace?: string;
      name: string;
      labels?: any;
      deletionTimestamp?: any;
      generation: number;
      creationTimestamp: any;
      annotations?: any;
    };
    spec: {
      targetNamespace: string;
      logo?: string;
      displayName?: string;
      clusterName: string;
      accountName: string;
    };
    status?: {
      lastReconcileTime?: any;
      isReady: boolean;
      checks?: any;
      resources?: Array<{
        name: string;
        kind?: string;
        apiVersion?: string;
        namespace: string;
      }>;
      message?: { RawMessage?: any };
    };
    syncStatus: {
      syncScheduledAt?: any;
      state: Kloudlite_Io__Pkg__Types_SyncStatusState;
      recordVersion: number;
      lastSyncedAt?: any;
      error?: string;
      action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
    };
  };
};

export type ConsoleListProjectsQueryVariables = Exact<{
  clusterName?: InputMaybe<Scalars['String']['input']>;
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchProjects>;
}>;

export type ConsoleListProjectsQuery = {
  core_listProjects?: {
    totalCount: number;
    edges: Array<{
      node: {
        id: string;
        displayName: string;
        creationTime: any;
        clusterName: string;
        apiVersion: string;
        kind: string;
        recordVersion: number;
        updateTime: any;
        accountName: string;
        metadata: {
          namespace?: string;
          name: string;
          labels?: any;
          deletionTimestamp?: any;
          generation: number;
          creationTimestamp: any;
          annotations?: any;
        };
        spec: {
          targetNamespace: string;
          logo?: string;
          displayName?: string;
          clusterName: string;
          accountName: string;
        };
        status?: {
          lastReconcileTime?: any;
          isReady: boolean;
          checks?: any;
          resources?: Array<{
            name: string;
            kind?: string;
            apiVersion?: string;
            namespace: string;
          }>;
          message?: { RawMessage?: any };
        };
        syncStatus: {
          syncScheduledAt?: any;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
          recordVersion: number;
          lastSyncedAt?: any;
          error?: string;
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
        };
      };
    }>;
    pageInfo: {
      startCursor?: string;
      hasNextPage?: boolean;
      endCursor?: string;
      hasPreviousPage?: boolean;
    };
  };
};

export type ConsoleCreateClusterMutationVariables = Exact<{
  cluster: ClusterIn;
}>;

export type ConsoleCreateClusterMutation = {
  infra_createCluster?: { id: string };
};

export type ConsoleClustersCountQueryVariables = Exact<{
  [key: string]: never;
}>;

export type ConsoleClustersCountQuery = {
  infra_listClusters?: { totalCount: number };
};

export type ConsoleListClustersQueryVariables = Exact<{
  search?: InputMaybe<SearchCluster>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListClustersQuery = {
  infra_listClusters?: {
    totalCount: number;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        displayName: string;
        updateTime: any;
        recordVersion: number;
        metadata: { name: string; annotations?: any };
        syncStatus: {
          syncScheduledAt?: any;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
          error?: string;
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
        };
        status?: {
          lastReconcileTime?: any;
          isReady: boolean;
          checks?: any;
          resources?: Array<{
            namespace: string;
            name: string;
            kind?: string;
            apiVersion?: string;
          }>;
          message?: { RawMessage?: any };
        };
        spec?: {
          vpc?: string;
          region: string;
          cloudProvider: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider;
          availabilityMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode;
          credentialsRef: { namespace?: string; name: string };
        };
      };
    }>;
  };
};

export type ConsoleGetClusterQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetClusterQuery = {
  infra_getCluster?: {
    displayName: string;
    updateTime: any;
    recordVersion: number;
    metadata: { name: string; annotations?: any };
    syncStatus: {
      syncScheduledAt?: any;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Kloudlite_Io__Pkg__Types_SyncStatusState;
      error?: string;
      action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
    };
    status?: {
      lastReconcileTime?: any;
      isReady: boolean;
      checks?: any;
      resources?: Array<{
        namespace: string;
        name: string;
        kind?: string;
        apiVersion?: string;
      }>;
      message?: { RawMessage?: any };
    };
    spec?: {
      vpc?: string;
      region: string;
      cloudProvider: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider;
      availabilityMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode;
      credentialsRef: { namespace?: string; name: string };
    };
  };
};

export type ConsoleListProviderSecretsQueryVariables = Exact<{
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchProviderSecret>;
}>;

export type ConsoleListProviderSecretsQuery = {
  infra_listProviderSecrets?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        enabled?: boolean;
        stringData?: any;
        cloudProviderName: CloudProviderSecretCloudProviderName;
        creationTime: any;
        updateTime: any;
        metadata: { annotations?: any; name: string };
        status?: {
          lastReconcileTime?: any;
          isReady: boolean;
          checks?: any;
          resources?: Array<{
            namespace: string;
            name: string;
            kind?: string;
            apiVersion?: string;
          }>;
          message?: { RawMessage?: any };
        };
      };
    }>;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
  };
};

export type ConsoleCreateProviderSecretMutationVariables = Exact<{
  secret: CloudProviderSecretIn;
}>;

export type ConsoleCreateProviderSecretMutation = {
  infra_createProviderSecret?: { metadata: { name: string } };
};

export type ConsoleUpdateProviderSecretMutationVariables = Exact<{
  secret: CloudProviderSecretIn;
}>;

export type ConsoleUpdateProviderSecretMutation = {
  infra_updateProviderSecret?: { id: string };
};

export type ConsoleDeleteProviderSecretMutationVariables = Exact<{
  secretName: Scalars['String']['input'];
}>;

export type ConsoleDeleteProviderSecretMutation = {
  infra_deleteProviderSecret: boolean;
};

export type ConsoleGetProviderSecretQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetProviderSecretQuery = {
  infra_getProviderSecret?: {
    enabled?: boolean;
    stringData?: any;
    cloudProviderName: CloudProviderSecretCloudProviderName;
    creationTime: any;
    updateTime: any;
    metadata: { annotations?: any; name: string };
    status?: {
      lastReconcileTime?: any;
      isReady: boolean;
      checks?: any;
      resources?: Array<{
        namespace: string;
        name: string;
        kind?: string;
        apiVersion?: string;
      }>;
      message?: { RawMessage?: any };
    };
  };
};

export type ConsoleGetNodePoolQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  poolName: Scalars['String']['input'];
}>;

export type ConsoleGetNodePoolQuery = {
  infra_getNodePool?: {
    updateTime: any;
    clusterName: string;
    spec: {
      targetCount: number;
      minCount: number;
      maxCount: number;
      awsNodeConfig?: {
        vpc?: string;
        region?: string;
        provisionMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode;
        isGpu?: boolean;
        imageId?: string;
        spotSpecs?: {
          memMin: number;
          memMax: number;
          cpuMin: number;
          cpuMax: number;
        };
        onDemandSpecs?: { instanceType: string };
      };
    };
    metadata: { name: string; annotations?: any };
    status?: { isReady: boolean; checks?: any; message?: { RawMessage?: any } };
  };
};

export type ConsoleCreateNodePoolMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  pool: NodePoolIn;
}>;

export type ConsoleCreateNodePoolMutation = {
  infra_createNodePool?: { id: string };
};

export type ConsoleListNodePoolsQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchNodepool>;
}>;

export type ConsoleListNodePoolsQuery = {
  infra_listNodePools?: {
    totalCount: number;
    edges: Array<{
      node: {
        updateTime: any;
        clusterName: string;
        spec: {
          targetCount: number;
          minCount: number;
          maxCount: number;
          awsNodeConfig?: {
            vpc?: string;
            region?: string;
            provisionMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode;
            isGpu?: boolean;
            imageId?: string;
            spotSpecs?: {
              memMin: number;
              memMax: number;
              cpuMin: number;
              cpuMax: number;
            };
            onDemandSpecs?: { instanceType: string };
          };
        };
        metadata: { name: string; annotations?: any };
        status?: {
          isReady: boolean;
          checks?: any;
          message?: { RawMessage?: any };
        };
      };
    }>;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
  };
};

export type ConsoleGetWorkspaceQueryVariables = Exact<{
  project: ProjectId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetWorkspaceQuery = {
  core_getWorkspace?: {
    displayName: string;
    clusterName: string;
    updateTime: any;
    metadata: {
      name: string;
      namespace?: string;
      labels?: any;
      annotations?: any;
    };
    spec?: { targetNamespace: string; projectName: string };
  };
};

export type ConsoleCreateWorkspaceMutationVariables = Exact<{
  env: WorkspaceIn;
}>;

export type ConsoleCreateWorkspaceMutation = {
  core_createWorkspace?: { id: string };
};

export type ConsoleUpdateWorkspaceMutationVariables = Exact<{
  env: WorkspaceIn;
}>;

export type ConsoleUpdateWorkspaceMutation = {
  core_updateWorkspace?: { id: string };
};

export type ConsoleListWorkspacesQueryVariables = Exact<{
  project: ProjectId;
  search?: InputMaybe<SearchWorkspaces>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListWorkspacesQuery = {
  core_listWorkspaces?: {
    totalCount: number;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
    edges: Array<{
      node: {
        displayName: string;
        clusterName: string;
        updateTime: any;
        creationTime: any;
        metadata: {
          name: string;
          namespace?: string;
          labels?: any;
          annotations?: any;
        };
        spec?: { targetNamespace: string; projectName: string };
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
      };
    }>;
  };
};

export type ConsoleGetEnvironmentQueryVariables = Exact<{
  project: ProjectId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetEnvironmentQuery = {
  core_getEnvironment?: {
    displayName: string;
    clusterName: string;
    updateTime: any;
    metadata: {
      name: string;
      namespace?: string;
      labels?: any;
      annotations?: any;
    };
    spec?: { targetNamespace: string; projectName: string };
  };
};

export type ConsoleCreateEnvironmentMutationVariables = Exact<{
  env: WorkspaceIn;
}>;

export type ConsoleCreateEnvironmentMutation = {
  core_createEnvironment?: { id: string };
};

export type ConsoleUpdateEnvironmentMutationVariables = Exact<{
  env: WorkspaceIn;
}>;

export type ConsoleUpdateEnvironmentMutation = {
  core_updateEnvironment?: { id: string };
};

export type ConsoleListEnvironmentsQueryVariables = Exact<{
  project: ProjectId;
  search?: InputMaybe<SearchWorkspaces>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListEnvironmentsQuery = {
  core_listEnvironments?: {
    totalCount: number;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        displayName: string;
        clusterName: string;
        updateTime: any;
        metadata: {
          name: string;
          namespace?: string;
          labels?: any;
          annotations?: any;
        };
        spec?: { targetNamespace: string; projectName: string };
      };
    }>;
  };
};

export type ConsoleCreateAppMutationVariables = Exact<{
  app: AppIn;
}>;

export type ConsoleCreateAppMutation = { core_createApp?: { id: string } };

export type ConsoleUpdateAppMutationVariables = Exact<{
  app: AppIn;
}>;

export type ConsoleUpdateAppMutation = { core_updateApp?: { id: string } };

export type ConsoleGetAppQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetAppQuery = {
  core_getApp?: {
    creationTime: any;
    accountName: string;
    displayName: string;
    markedForDeletion?: boolean;
    updateTime: any;
    createdBy: { userName: string; userId: string; userEmail: string };
    lastUpdatedBy: { userName: string; userId: string; userEmail: string };
    metadata: { name: string; namespace?: string; annotations?: any };
    spec: {
      serviceAccount?: string;
      replicas?: number;
      region?: string;
      nodeSelector?: any;
      freeze?: boolean;
      displayName?: string;
      tolerations?: Array<{
        value?: string;
        tolerationSeconds?: number;
        operator?: string;
        key?: string;
        effect?: string;
      }>;
      services?: Array<{
        type?: string;
        targetPort?: number;
        port: number;
        name?: string;
      }>;
      intercept?: { enabled: boolean; toDevice: string };
      hpa?: {
        maxReplicas?: number;
        enabled?: boolean;
        minReplicas?: number;
        thresholdCpu?: number;
        thresholdMemory?: number;
      };
      containers: Array<{
        args?: Array<string>;
        command?: Array<string>;
        image: string;
        imagePullPolicy?: string;
        name: string;
        env?: Array<{
          refName?: string;
          refKey?: string;
          optional?: boolean;
          key: string;
          type?: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvType;
          value?: string;
        }>;
        envFrom?: Array<{
          type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromType;
          refName: string;
        }>;
        livenessProbe?: {
          type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeType;
          interval?: number;
          initialDelay?: number;
          failureThreshold?: number;
          tcp?: { port: number };
          shell?: { command?: Array<string> };
          httpGet?: { httpHeaders?: any; path: string; port: number };
        };
        readinessProbe?: {
          type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeType;
          interval?: number;
          initialDelay?: number;
          failureThreshold?: number;
          tcp?: { port: number };
          shell?: { command?: Array<string> };
          httpGet?: { httpHeaders?: any; path: string; port: number };
        };
        resourceCpu?: { min?: string; max?: string };
        resourceMemory?: { min?: string; max?: string };
        volumes?: Array<{
          type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesType;
          refName: string;
          mountPath: string;
          items?: Array<{ fileName?: string; key: string }>;
        }>;
      }>;
    };
    status?: {
      lastReconcileTime?: any;
      isReady: boolean;
      checks?: any;
      resources?: Array<{
        namespace: string;
        name: string;
        kind?: string;
        apiVersion?: string;
      }>;
      message?: { RawMessage?: any };
    };
    syncStatus: {
      syncScheduledAt?: any;
      state: Kloudlite_Io__Pkg__Types_SyncStatusState;
      recordVersion: number;
      lastSyncedAt?: any;
      error?: string;
      action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
    };
  };
};

export type ConsoleListAppsQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchApps>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListAppsQuery = {
  core_listApps?: {
    totalCount: number;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        markedForDeletion?: boolean;
        updateTime: any;
        createdBy: { userName: string; userId: string; userEmail: string };
        lastUpdatedBy: { userName: string; userId: string; userEmail: string };
        metadata: { name: string };
        spec: { freeze?: boolean; displayName?: string };
        status?: {
          lastReconcileTime?: any;
          isReady: boolean;
          checks?: any;
          resources?: Array<{
            namespace: string;
            name: string;
            kind?: string;
            apiVersion?: string;
          }>;
          message?: { RawMessage?: any };
        };
        syncStatus: {
          syncScheduledAt?: any;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
          recordVersion: number;
          lastSyncedAt?: any;
          error?: string;
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
        };
      };
    }>;
  };
};

export type ConsoleListRoutersQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchRouters>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListRoutersQuery = {
  core_listRouters?: {
    edges: Array<{
      cursor: string;
      node: {
        metadata: {
          name: string;
          namespace?: string;
          annotations?: any;
          labels?: any;
        };
        spec: {
          routes?: Array<{ app?: string; lambda?: string; path: string }>;
        };
      };
    }>;
  };
};

export type ConsoleUpdateConfigMutationVariables = Exact<{
  config: ConfigIn;
}>;

export type ConsoleUpdateConfigMutation = {
  core_updateConfig?: { id: string };
};

export type ConsoleGetConfigQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetConfigQuery = {
  core_getConfig?: {
    displayName: string;
    updateTime: any;
    data?: any;
    metadata: {
      namespace?: string;
      name: string;
      annotations?: any;
      labels?: any;
    };
  };
};

export type ConsoleListConfigsQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchConfigs>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListConfigsQuery = {
  core_listConfigs?: {
    totalCount: number;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
    edges: Array<{
      node: {
        displayName: string;
        updateTime: any;
        data?: any;
        metadata: {
          namespace?: string;
          name: string;
          annotations?: any;
          labels?: any;
        };
      };
    }>;
  };
};

export type ConsoleCreateConfigMutationVariables = Exact<{
  config: ConfigIn;
}>;

export type ConsoleCreateConfigMutation = {
  core_createConfig?: { id: string };
};

export type ConsoleListSecretsQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  pq?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchSecrets>;
}>;

export type ConsoleListSecretsQuery = {
  core_listSecrets?: {
    totalCount: number;
    pageInfo: {
      startCursor?: string;
      hasPreviousPage?: boolean;
      hasNextPage?: boolean;
      endCursor?: string;
    };
    edges: Array<{
      node: {
        stringData?: any;
        updateTime: any;
        displayName: string;
        metadata: {
          name: string;
          namespace?: string;
          annotations?: any;
          labels?: any;
        };
      };
    }>;
  };
};

export type ConsoleCreateSecretMutationVariables = Exact<{
  secret: SecretIn;
}>;

export type ConsoleCreateSecretMutation = {
  core_createSecret?: { id: string };
};

export type ConsoleGetSecretQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetSecretQuery = {
  core_getSecret?: {
    stringData?: any;
    updateTime: any;
    displayName: string;
    metadata: {
      name: string;
      namespace?: string;
      annotations?: any;
      labels?: any;
    };
  };
};

export type ConsoleUpdateSecretMutationVariables = Exact<{
  secret: SecretIn;
}>;

export type ConsoleUpdateSecretMutation = {
  core_updateSecret?: { id: string };
};

export type ConsoleCreateVpnDeviceMutationVariables = Exact<{
  vpnDevice: VpnDeviceIn;
}>;

export type ConsoleCreateVpnDeviceMutation = {
  core_createVPNDevice?: { id: string };
};

export type ConsoleUpdateVpnDeviceMutationVariables = Exact<{
  vpnDevice: VpnDeviceIn;
}>;

export type ConsoleUpdateVpnDeviceMutation = {
  core_updateVPNDevice?: { id: string };
};

export type ConsoleListVpnDevicesQueryVariables = Exact<{
  search?: InputMaybe<SearchVpnDevices>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListVpnDevicesQuery = {
  core_listVPNDevices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        clusterName: string;
        displayName: string;
        metadata: { name: string };
        spec?: {
          serverName: string;
          offset: number;
          ports?: Array<{ port?: number; targetPort?: number }>;
        };
        createdBy: { userId: string; userName: string; userEmail: string };
      };
    }>;
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPreviousPage?: boolean;
      startCursor?: string;
    };
  };
};

export type ConsoleGetVpnDeviceQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetVpnDeviceQuery = {
  core_getVPNDevice?: {
    updateTime: any;
    clusterName: string;
    displayName: string;
    spec?: {
      serverName: string;
      offset: number;
      ports?: Array<{ port?: number; targetPort?: number }>;
    };
    metadata: { name: string };
  };
};

export type ConsoleListInvitationsForAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleListInvitationsForAccountQuery = {
  accounts_listInvitations?: Array<{
    accepted?: boolean;
    accountName: string;
    creationTime: any;
    id: string;
    inviteToken: string;
    invitedBy: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    rejected?: boolean;
    updateTime: any;
    userEmail?: string;
    userName?: string;
    userRole: string;
  }>;
};

export type ConsoleListMembershipsForAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleListMembershipsForAccountQuery = {
  accounts_listMembershipsForAccount?: Array<{
    role: string;
    user: { verified: boolean; name: string; joined: any; email: string };
  }>;
};

export type ConsoleDeleteAccountInvitationMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  invitationId: Scalars['String']['input'];
}>;

export type ConsoleDeleteAccountInvitationMutation = {
  accounts_deleteInvitation: boolean;
};

export type ConsoleInviteMemberForAccountMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  invitation: InvitationIn;
}>;

export type ConsoleInviteMemberForAccountMutation = {
  accounts_inviteMember: { id: string };
};

export type ConsoleUpdateAccountMembershipMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  memberId: Scalars['ID']['input'];
  role: Scalars['String']['input'];
}>;

export type ConsoleUpdateAccountMembershipMutation = {
  accounts_updateAccountMembership: boolean;
};

export type AuthRequestResetPasswordMutationVariables = Exact<{
  email: Scalars['String']['input'];
}>;

export type AuthRequestResetPasswordMutation = {
  auth_requestResetPassword: boolean;
};

export type AuthResetPasswordMutationVariables = Exact<{
  token: Scalars['String']['input'];
  password: Scalars['String']['input'];
}>;

export type AuthResetPasswordMutation = { auth_resetPassword: boolean };

export type AuthOauthLoginMutationVariables = Exact<{
  code: Scalars['String']['input'];
  provider: Scalars['String']['input'];
  state?: InputMaybe<Scalars['String']['input']>;
}>;

export type AuthOauthLoginMutation = { oAuth_login: { id: string } };

export type AuthVerifyEmailMutationVariables = Exact<{
  token: Scalars['String']['input'];
}>;

export type AuthVerifyEmailMutation = { auth_verifyEmail: { id: string } };

export type AuthLoginPageInitUrlsQueryVariables = Exact<{
  [key: string]: never;
}>;

export type AuthLoginPageInitUrlsQuery = {
  githubLoginUrl: any;
  gitlabLoginUrl: any;
  googleLoginUrl: any;
};

export type AuthLoginMutationVariables = Exact<{
  email: Scalars['String']['input'];
  password: Scalars['String']['input'];
}>;

export type AuthLoginMutation = { auth_login?: { id: string } };

export type AuthLogoutMutationVariables = Exact<{ [key: string]: never }>;

export type AuthLogoutMutation = { auth_logout: boolean };

export type AuthSignUpWithEmailMutationVariables = Exact<{
  name: Scalars['String']['input'];
  password: Scalars['String']['input'];
  email: Scalars['String']['input'];
}>;

export type AuthSignUpWithEmailMutation = { auth_signup?: { id: string } };

export type AuthWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type AuthWhoAmIQuery = {
  auth_me?: { id: string; email: string; verified: boolean };
};

export type LibWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type LibWhoAmIQuery = {
  auth_me?: { verified: boolean; name: string; id: string; email: string };
};
