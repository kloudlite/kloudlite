export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
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
  | 'managedresource'
  | 'managedservice'
  | 'project'
  | 'router'
  | 'secret'
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

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeType =
  'httpGet' | 'shell' | 'tcp';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeType =
  'httpGet' | 'shell' | 'tcp';

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
  containers: Array<
    InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersIn>
  >;
  displayName?: InputMaybe<Scalars['String']['input']>;
  freeze?: InputMaybe<Scalars['Boolean']['input']>;
  hpa?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecHpaIn>;
  intercept?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecInterceptIn>;
  nodeSelector?: InputMaybe<Scalars['Map']['input']>;
  region?: InputMaybe<Scalars['String']['input']>;
  replicas?: InputMaybe<Scalars['Int']['input']>;
  serviceAccount?: InputMaybe<Scalars['String']['input']>;
  services?: InputMaybe<
    Array<
      InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecServicesIn>
    >
  >;
  tolerations?: InputMaybe<
    Array<
      InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecTolerationsIn>
    >
  >;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersIn =
  {
    args?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    command?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
    env?: InputMaybe<
      Array<
        InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvIn>
      >
    >;
    envFrom?: InputMaybe<
      Array<
        InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromIn>
      >
    >;
    image: Scalars['String']['input'];
    imagePullPolicy?: InputMaybe<Scalars['String']['input']>;
    livenessProbe?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeIn>;
    name: Scalars['String']['input'];
    readinessProbe?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeIn>;
    resourceCpu?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceCpuIn>;
    resourceMemory?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceMemoryIn>;
    volumes?: InputMaybe<
      Array<
        InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesIn>
      >
    >;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvIn =
  {
    key: Scalars['String']['input'];
    optional?: InputMaybe<Scalars['Boolean']['input']>;
    refKey?: InputMaybe<Scalars['String']['input']>;
    refName?: InputMaybe<Scalars['String']['input']>;
    type?: InputMaybe<Scalars['String']['input']>;
    value?: InputMaybe<Scalars['String']['input']>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromIn =
  {
    refName: Scalars['String']['input'];
    type: Scalars['String']['input'];
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
    command?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
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
    command?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
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
      Array<
        InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesItemsIn>
      >
    >;
    mountPath: Scalars['String']['input'];
    refName: Scalars['String']['input'];
    type: Scalars['String']['input'];
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
      Array<
        InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecTolerationsIn>
      >
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
  domains: Array<InputMaybe<Scalars['String']['input']>>;
  https?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecHttpsIn>;
  ingressClass?: InputMaybe<Scalars['String']['input']>;
  maxBodySizeInMB?: InputMaybe<Scalars['Int']['input']>;
  rateLimit?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRateLimitIn>;
  region?: InputMaybe<Scalars['String']['input']>;
  routes?: InputMaybe<
    Array<
      InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRoutesIn>
    >
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
  origins?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
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
      Array<
        InputMaybe<Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecPortsIn>
      >
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
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ByocSpecIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ByocSpecIn = {
  accountName: Scalars['String']['input'];
  displayName?: InputMaybe<Scalars['String']['input']>;
  incomingKafkaTopic: Scalars['String']['input'];
  ingressClasses?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  provider: Scalars['String']['input'];
  publicIps?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  region: Scalars['String']['input'];
  storageClasses?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
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
    nodeIps?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
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
    taints?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
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
  taints?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
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
    suggestedNames?: Array<string> | null;
  };
};

export type ConsoleInfraCheckNameAvailabilityQueryVariables = Exact<{
  resType: ResType;
  name: Scalars['String']['input'];
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
    suggestedNames?: Array<string> | null;
  };
};

export type ConsoleWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type ConsoleWhoAmIQuery = {
  auth_me?: { id: string; email: string } | null;
};

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
    metadata: { name: string; annotations?: any | null };
  } | null> | null;
};

export type ConsoleGetAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleGetAccountQuery = {
  accounts_getAccount?: {
    updateTime: any;
    contactEmail: string;
    displayName: string;
    metadata: { name: string; annotations?: any | null };
  } | null;
};

export type ConsoleCreateProjectMutationVariables = Exact<{
  project: ProjectIn;
}>;

export type ConsoleCreateProjectMutation = {
  core_createProject?: { id: string } | null;
};

export type ConsoleGetProjectQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetProjectQuery = {
  core_getProject?: {
    metadata: {
      name: string;
      annotations?: any | null;
      namespace?: string | null;
    };
    spec: { targetNamespace: string; displayName?: string | null };
  } | null;
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
        creationTime: any;
        clusterName: string;
        apiVersion: string;
        kind: string;
        recordVersion: number;
        updateTime: any;
        accountName: string;
        metadata: {
          namespace?: string | null;
          name: string;
          labels?: any | null;
          deletionTimestamp?: any | null;
          generation: number;
          creationTimestamp: any;
          annotations?: any | null;
        };
        spec: {
          targetNamespace: string;
          logo?: string | null;
          displayName?: string | null;
          clusterName: string;
          accountName: string;
        };
        status?: {
          lastReconcileTime?: any | null;
          isReady: boolean;
          checks?: any | null;
          resources?: Array<{
            name: string;
            kind?: string | null;
            apiVersion?: string | null;
            namespace: string;
          }> | null;
          message?: { RawMessage?: any | null } | null;
        } | null;
        syncStatus: {
          syncScheduledAt?: any | null;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
          recordVersion: number;
          lastSyncedAt?: any | null;
          error?: string | null;
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
        };
      };
    }>;
    pageInfo: {
      startCursor?: string | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
      hasPreviousPage?: boolean | null;
    };
  } | null;
};

export type ConsoleCreateClusterMutationVariables = Exact<{
  cluster: ClusterIn;
}>;

export type ConsoleCreateClusterMutation = {
  infra_createCluster?: { id: string } | null;
};

export type ConsoleClustersCountQueryVariables = Exact<{
  [key: string]: never;
}>;

export type ConsoleClustersCountQuery = {
  infra_listClusters?: { totalCount: number } | null;
};

export type ConsoleListClustersQueryVariables = Exact<{
  search?: InputMaybe<SearchCluster>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListClustersQuery = {
  infra_listClusters?: {
    totalCount: number;
    pageInfo: {
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
    edges: Array<{
      cursor: string;
      node: {
        updateTime: any;
        recordVersion: number;
        metadata: { name: string; annotations?: any | null };
        syncStatus: {
          syncScheduledAt?: any | null;
          lastSyncedAt?: any | null;
          recordVersion: number;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
          error?: string | null;
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
        };
        status?: {
          lastReconcileTime?: any | null;
          isReady: boolean;
          checks?: any | null;
          resources?: Array<{
            namespace: string;
            name: string;
            kind?: string | null;
            apiVersion?: string | null;
          }> | null;
          message?: { RawMessage?: any | null } | null;
        } | null;
        spec?: {
          vpc?: string | null;
          region: string;
          cloudProvider: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider;
          availabilityMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode;
          credentialsRef: { namespace?: string | null; name: string };
        } | null;
      };
    }>;
  } | null;
};

export type ConsoleGetClusterQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetClusterQuery = {
  infra_getCluster?: {
    metadata: { name: string; annotations?: any | null };
    spec?: {
      vpc?: string | null;
      region: string;
      nodeIps?: Array<string | null> | null;
      cloudProvider: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider;
      availabilityMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode;
    } | null;
  } | null;
};

export type ConsoleListProviderSecretsQueryVariables = Exact<{
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchProviderSecret>;
}>;

export type ConsoleListProviderSecretsQuery = {
  infra_listProviderSecrets?: {
    totalCount: number;
    edges: Array<{
      node: {
        enabled?: boolean | null;
        stringData?: any | null;
        cloudProviderName: CloudProviderSecretCloudProviderName;
        creationTime: any;
        updateTime: any;
        metadata: { annotations?: any | null; name: string };
        status?: {
          lastReconcileTime?: any | null;
          isReady: boolean;
          checks?: any | null;
          resources?: Array<{
            namespace: string;
            name: string;
            kind?: string | null;
            apiVersion?: string | null;
          }> | null;
          message?: { RawMessage?: any | null } | null;
        } | null;
      };
    }>;
    pageInfo: {
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
  } | null;
};

export type ConsoleCreateProviderSecretMutationVariables = Exact<{
  secret: CloudProviderSecretIn;
}>;

export type ConsoleCreateProviderSecretMutation = {
  infra_createProviderSecret?: { metadata: { name: string } } | null;
};

export type ConsoleUpdateProviderSecretMutationVariables = Exact<{
  secret: CloudProviderSecretIn;
}>;

export type ConsoleUpdateProviderSecretMutation = {
  infra_updateProviderSecret?: { id: string } | null;
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
    cloudProviderName: CloudProviderSecretCloudProviderName;
    metadata: { name: string; annotations?: any | null };
  } | null;
};

export type ConsoleCreateNodePoolMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  pool: NodePoolIn;
}>;

export type ConsoleCreateNodePoolMutation = {
  infra_createNodePool?: { id: string } | null;
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
            vpc?: string | null;
            region?: string | null;
            provisionMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode;
            isGpu?: boolean | null;
            imageId?: string | null;
            spotSpecs?: {
              memMin: number;
              memMax: number;
              cpuMin: number;
              cpuMax: number;
            } | null;
            onDemandSpecs?: { instanceType: string } | null;
          } | null;
        };
        metadata: { name: string; annotations?: any | null };
        status?: {
          isReady: boolean;
          checks?: any | null;
          message?: { RawMessage?: any | null } | null;
        } | null;
      };
    }>;
    pageInfo: {
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
  } | null;
};

export type ConsoleGetWorkspaceQueryVariables = Exact<{
  project: ProjectId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetWorkspaceQuery = {
  core_getWorkspace?: {
    displayName: string;
    updateTime: any;
    spec?: { targetNamespace: string; projectName: string } | null;
    metadata: {
      namespace?: string | null;
      name: string;
      annotations?: any | null;
      labels?: any | null;
    };
  } | null;
};

export type ConsoleCreateWorkspaceMutationVariables = Exact<{
  env: WorkspaceIn;
}>;

export type ConsoleCreateWorkspaceMutation = {
  core_createWorkspace?: { id: string } | null;
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
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
    edges: Array<{
      node: {
        displayName: string;
        clusterName: string;
        updateTime: any;
        metadata: {
          name: string;
          namespace?: string | null;
          labels?: any | null;
          annotations?: any | null;
        };
        spec?: { targetNamespace: string; projectName: string } | null;
      };
    }>;
  } | null;
};

export type ConsoleGetEnvironmentQueryVariables = Exact<{
  project: ProjectId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetEnvironmentQuery = {
  core_getEnvironment?: {
    updateTime: any;
    displayName: string;
    spec?: { targetNamespace: string; projectName: string } | null;
    metadata: {
      namespace?: string | null;
      name: string;
      annotations?: any | null;
      labels?: any | null;
    };
  } | null;
};

export type ConsoleCreateEnvironmentMutationVariables = Exact<{
  env: WorkspaceIn;
}>;

export type ConsoleCreateEnvironmentMutation = {
  core_createEnvironment?: { id: string } | null;
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
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
    edges: Array<{
      node: {
        displayName: string;
        clusterName: string;
        updateTime: any;
        metadata: {
          name: string;
          namespace?: string | null;
          labels?: any | null;
          annotations?: any | null;
        };
        spec?: { targetNamespace: string; projectName: string } | null;
      };
    }>;
  } | null;
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
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
    edges: Array<{
      cursor: string;
      node: {
        clusterName: string;
        spec: { displayName?: string | null };
        metadata: {
          namespace?: string | null;
          name: string;
          labels?: any | null;
          annotations?: any | null;
        };
      };
    }>;
  } | null;
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
      node: {
        metadata: {
          name: string;
          namespace?: string | null;
          annotations?: any | null;
          labels?: any | null;
        };
        spec: {
          routes?: Array<{
            app?: string | null;
            lambda?: string | null;
            path: string;
          } | null> | null;
        };
      };
    }>;
  } | null;
};

export type ConsoleUpdateConfigMutationVariables = Exact<{
  config: ConfigIn;
}>;

export type ConsoleUpdateConfigMutation = {
  core_updateConfig?: { id: string } | null;
};

export type ConsoleGetConfigQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetConfigQuery = {
  core_getConfig?: {
    data?: any | null;
    updateTime: any;
    displayName: string;
    metadata: {
      name: string;
      namespace?: string | null;
      annotations?: any | null;
      labels?: any | null;
    };
  } | null;
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
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
    edges: Array<{
      node: {
        displayName: string;
        updateTime: any;
        data?: any | null;
        metadata: {
          namespace?: string | null;
          name: string;
          annotations?: any | null;
          labels?: any | null;
        };
      };
    }>;
  } | null;
};

export type ConsoleCreateConfigMutationVariables = Exact<{
  config: ConfigIn;
}>;

export type ConsoleCreateConfigMutation = {
  core_createConfig?: { id: string } | null;
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
      startCursor?: string | null;
      hasPreviousPage?: boolean | null;
      hasNextPage?: boolean | null;
      endCursor?: string | null;
    };
    edges: Array<{
      node: {
        updateTime: any;
        stringData?: any | null;
        metadata: {
          namespace?: string | null;
          name: string;
          annotations?: any | null;
          labels?: any | null;
        };
      };
    }>;
  } | null;
};

export type ConsoleCreateSecretMutationVariables = Exact<{
  secret: SecretIn;
}>;

export type ConsoleCreateSecretMutation = {
  core_createSecret?: { id: string } | null;
};

export type ConsoleGetSecretQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetSecretQuery = {
  core_getSecret?: {
    stringData?: any | null;
    updateTime: any;
    displayName: string;
    metadata: {
      name: string;
      namespace?: string | null;
      annotations?: any | null;
      labels?: any | null;
    };
  } | null;
};

export type ConsoleUpdateSecretMutationVariables = Exact<{
  secret: SecretIn;
}>;

export type ConsoleUpdateSecretMutation = {
  core_updateSecret?: { id: string } | null;
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

export type AuthLoginMutation = { auth_login?: { id: string } | null };

export type AuthLogoutMutationVariables = Exact<{ [key: string]: never }>;

export type AuthLogoutMutation = { auth_logout: boolean };

export type AuthSignUpWithEmailMutationVariables = Exact<{
  name: Scalars['String']['input'];
  password: Scalars['String']['input'];
  email: Scalars['String']['input'];
}>;

export type AuthSignUpWithEmailMutation = {
  auth_signup?: { id: string } | null;
};

export type AuthWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type AuthWhoAmIQuery = {
  auth_me?: { id: string; email: string; verified: boolean } | null;
};

export type LibWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type LibWhoAmIQuery = {
  auth_me?: {
    verified: boolean;
    name: string;
    id: string;
    email: string;
  } | null;
};
