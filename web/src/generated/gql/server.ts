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
  Json: { input: any; output: any };
  ProviderDetail: { input: any; output: any };
  Any: { input: any; output: any };
  URL: { input: any; output: any };
};

export type Github__Com___Kloudlite___Api___Apps___Iam___Types__Role =
  | 'account_admin'
  | 'account_member'
  | 'account_owner'
  | 'project_admin'
  | 'project_member'
  | 'resource_owner';

export type ConsoleResType =
  | 'app'
  | 'config'
  | 'environment'
  | 'managed_resource'
  | 'managed_service'
  | 'router'
  | 'secret'
  | 'vpn_device';

export type Github__Com___Kloudlite___Api___Pkg___Types__SyncAction =
  | 'APPLY'
  | 'DELETE';

export type Github__Com___Kloudlite___Api___Pkg___Types__SyncState =
  | 'APPLIED_AT_AGENT'
  | 'DELETED_AT_AGENT'
  | 'DELETING_AT_AGENT'
  | 'ERRORED_AT_AGENT'
  | 'IDLE'
  | 'IN_QUEUE'
  | 'UPDATED_AT_AGENT';

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitProvider =
  'github' | 'gitlab';

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__BuildStatus =
  'error' | 'failed' | 'idle' | 'pending' | 'queued' | 'running' | 'success';

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret =
  'config' | 'pvc' | 'secret';

export type K8s__Io___Api___Core___V1__TaintEffect =
  | 'NoExecute'
  | 'NoSchedule'
  | 'PreferNoSchedule';

export type K8s__Io___Api___Core___V1__TolerationOperator = 'Equal' | 'Exists';

export type K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator =
  'DoesNotExist' | 'Exists' | 'In' | 'NotIn';

export type K8s__Io___Api___Core___V1__UnsatisfiableConstraintAction =
  | 'DoNotSchedule'
  | 'ScheduleAnyway';

export type ConfigKeyRefIn = {
  configName: Scalars['String']['input'];
  key: Scalars['String']['input'];
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentRoutingMode =
  'private' | 'public';

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ExternalAppRecordType =
  'CNAME' | 'IPAddr';

export type Github__Com___Kloudlite___Api___Apps___Console___Internal___Entities__PullSecretFormat =
  'dockerConfigJson' | 'params';

export type ManagedResourceKeyRefIn = {
  key: Scalars['String']['input'];
  mresName: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__SecretType =
  | 'bootstrap__kubernetes__io___token'
  | 'kubernetes__io___basic____auth'
  | 'kubernetes__io___dockercfg'
  | 'kubernetes__io___dockerconfigjson'
  | 'kubernetes__io___service____account____token'
  | 'kubernetes__io___ssh____auth'
  | 'kubernetes__io___tls'
  | 'Opaque';

export type SecretKeyRefIn = {
  key: Scalars['String']['input'];
  secretName: Scalars['String']['input'];
};

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
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type MatchFilterIn = {
  array?: InputMaybe<Array<Scalars['Any']['input']>>;
  exact?: InputMaybe<Scalars['Any']['input']>;
  matchType: Scalars['String']['input'];
  notInArray?: InputMaybe<Array<Scalars['Any']['input']>>;
  regex?: InputMaybe<Scalars['String']['input']>;
};

export type SearchConfigs = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchEnvironments = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchExternalApps = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchImagePullSecrets = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchManagedResources = {
  isReady?: InputMaybe<MatchFilterIn>;
  managedServiceName?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchRouters = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchSecrets = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type CoreSearchVpnDevices = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchBuildRuns = {
  buildId?: InputMaybe<Scalars['ID']['input']>;
  repoName?: InputMaybe<MatchFilterIn>;
};

export type SearchBuilds = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchCreds = {
  text?: InputMaybe<MatchFilterIn>;
};

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__RepoAccess =
  'read' | 'read_write';

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__ExpirationUnit =
  'd' | 'h' | 'm' | 'w' | 'y';

export type SearchRepos = {
  text?: InputMaybe<MatchFilterIn>;
};

export type PaginationIn = {
  page?: InputMaybe<Scalars['Int']['input']>;
  per_page?: InputMaybe<Scalars['Int']['input']>;
};

export type ResType =
  | 'byok_cluster'
  | 'cluster'
  | 'cluster_managed_service'
  | 'global_vpn_device'
  | 'helm_release'
  | 'nodepool'
  | 'providersecret';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode =
  'dev' | 'HA';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsAuthMechanism =
  'assume_role' | 'secret_keys';

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider =
  'aws' | 'azure' | 'digitalocean' | 'gcp';

export type K8s__Io___Api___Core___V1__NodeSelectorOperator =
  | 'DoesNotExist'
  | 'Exists'
  | 'Gt'
  | 'In'
  | 'Lt'
  | 'NotIn';

export type K8s__Io___Api___Core___V1__ConditionStatus =
  | 'False'
  | 'True'
  | 'Unknown';

export type K8s__Io___Api___Core___V1__NamespaceConditionType =
  | 'NamespaceContentRemaining'
  | 'NamespaceDeletionContentFailure'
  | 'NamespaceDeletionDiscoveryFailure'
  | 'NamespaceDeletionGroupVersionParsingFailure'
  | 'NamespaceFinalizersRemaining';

export type K8s__Io___Api___Core___V1__NamespacePhase =
  | 'Active'
  | 'Terminating';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType =
  'ec2' | 'spot';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpPoolType =
  'SPOT' | 'STANDARD';

export type K8s__Io___Api___Core___V1__PersistentVolumeReclaimPolicy =
  | 'Delete'
  | 'Recycle'
  | 'Retain';

export type K8s__Io___Api___Core___V1__PersistentVolumePhase =
  | 'Available'
  | 'Bound'
  | 'Failed'
  | 'Pending'
  | 'Released';

export type K8s__Io___Api___Core___V1__PersistentVolumeClaimConditionType =
  | 'FileSystemResizePending'
  | 'Resizing';

export type K8s__Io___Api___Core___V1__PersistentVolumeClaimPhase =
  | 'Bound'
  | 'Lost'
  | 'Pending';

export type SearchCluster = {
  cloudProviderName?: InputMaybe<MatchFilterIn>;
  isReady?: InputMaybe<MatchFilterIn>;
  region?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchClusterManagedService = {
  isReady?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchDomainEntry = {
  clusterName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchGlobalVpnDevices = {
  creationMethod?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchGlobalVpNs = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchHelmRelease = {
  isReady?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchNamespaces = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchNodepool = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchProviderSecret = {
  cloudProviderName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchPersistentVolumeClaims = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchPersistentVolumes = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchVolumeAttachments = {
  text?: InputMaybe<MatchFilterIn>;
};

export type Github__Com___Kloudlite___Api___Apps___Iot____Console___Internal___Entities__BluePrintType =
  'group_blueprint' | 'singleton_blueprint';

export type SearchIotApps = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchIotDeployments = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchIotDeviceBlueprints = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchIotDevices = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchIotProjects = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type AccountIn = {
  contactEmail?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  isActive?: InputMaybe<Scalars['Boolean']['input']>;
  logo?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
};

export type MetadataIn = {
  annotations?: InputMaybe<Scalars['Map']['input']>;
  labels?: InputMaybe<Scalars['Map']['input']>;
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type InvitationIn = {
  userEmail?: InputMaybe<Scalars['String']['input']>;
  userName?: InputMaybe<Scalars['String']['input']>;
  userRole: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
};

export type AppIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  ciBuildId?: InputMaybe<Scalars['ID']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppSpecIn = {
  containers: Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppContainerIn>;
  displayName?: InputMaybe<Scalars['String']['input']>;
  freeze?: InputMaybe<Scalars['Boolean']['input']>;
  hpa?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__HpaIn>;
  intercept?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__InterceptIn>;
  nodeSelector?: InputMaybe<Scalars['Map']['input']>;
  region?: InputMaybe<Scalars['String']['input']>;
  replicas?: InputMaybe<Scalars['Int']['input']>;
  router?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppRouterIn>;
  serviceAccount?: InputMaybe<Scalars['String']['input']>;
  services?: InputMaybe<
    Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppSvcIn>
  >;
  tolerations?: InputMaybe<Array<K8s__Io___Api___Core___V1__TolerationIn>>;
  topologySpreadConstraints?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__TopologySpreadConstraintIn>
  >;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppContainerIn =
  {
    args?: InputMaybe<Array<Scalars['String']['input']>>;
    command?: InputMaybe<Array<Scalars['String']['input']>>;
    env?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerEnvIn>
    >;
    envFrom?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvFromIn>
    >;
    image: Scalars['String']['input'];
    imagePullPolicy?: InputMaybe<Scalars['String']['input']>;
    livenessProbe?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ProbeIn>;
    name: Scalars['String']['input'];
    readinessProbe?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ProbeIn>;
    resourceCpu?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerResourceIn>;
    resourceMemory?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerResourceIn>;
    volumes?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerVolumeIn>
    >;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerEnvIn =
  {
    key: Scalars['String']['input'];
    optional?: InputMaybe<Scalars['Boolean']['input']>;
    refKey?: InputMaybe<Scalars['String']['input']>;
    refName?: InputMaybe<Scalars['String']['input']>;
    type?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret>;
    value?: InputMaybe<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvFromIn = {
  refName: Scalars['String']['input'];
  type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ProbeIn = {
  failureThreshold?: InputMaybe<Scalars['Int']['input']>;
  httpGet?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__HttpGetProbeIn>;
  initialDelay?: InputMaybe<Scalars['Int']['input']>;
  interval?: InputMaybe<Scalars['Int']['input']>;
  shell?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ShellProbeIn>;
  tcp?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__TcpProbeIn>;
  type: Scalars['String']['input'];
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__HttpGetProbeIn =
  {
    httpHeaders?: InputMaybe<Scalars['Map']['input']>;
    path: Scalars['String']['input'];
    port: Scalars['Int']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ShellProbeIn =
  {
    command?: InputMaybe<Array<Scalars['String']['input']>>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__TcpProbeIn =
  {
    port: Scalars['Int']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerResourceIn =
  {
    max?: InputMaybe<Scalars['String']['input']>;
    min?: InputMaybe<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerVolumeIn =
  {
    items?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerVolumeItemIn>
    >;
    mountPath: Scalars['String']['input'];
    refName: Scalars['String']['input'];
    type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ContainerVolumeItemIn =
  {
    fileName?: InputMaybe<Scalars['String']['input']>;
    key: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__HpaIn = {
  enabled: Scalars['Boolean']['input'];
  maxReplicas?: InputMaybe<Scalars['Int']['input']>;
  minReplicas?: InputMaybe<Scalars['Int']['input']>;
  thresholdCpu?: InputMaybe<Scalars['Int']['input']>;
  thresholdMemory?: InputMaybe<Scalars['Int']['input']>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__InterceptIn =
  {
    enabled: Scalars['Boolean']['input'];
    portMappings?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn>
    >;
    toDevice: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn =
  {
    appPort: Scalars['Int']['input'];
    devicePort: Scalars['Int']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppRouterIn =
  {
    backendProtocol?: InputMaybe<Scalars['String']['input']>;
    basicAuth?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__BasicAuthIn>;
    cors?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__CorsIn>;
    domains: Array<Scalars['String']['input']>;
    https?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__HttpsIn>;
    ingressClass?: InputMaybe<Scalars['String']['input']>;
    maxBodySizeInMB?: InputMaybe<Scalars['Int']['input']>;
    rateLimit?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__RateLimitIn>;
    routes?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__RouteIn>
    >;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__BasicAuthIn =
  {
    enabled: Scalars['Boolean']['input'];
    secretName?: InputMaybe<Scalars['String']['input']>;
    username?: InputMaybe<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__CorsIn = {
  allowCredentials?: InputMaybe<Scalars['Boolean']['input']>;
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  origins?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__HttpsIn = {
  clusterIssuer?: InputMaybe<Scalars['String']['input']>;
  enabled: Scalars['Boolean']['input'];
  forceRedirect?: InputMaybe<Scalars['Boolean']['input']>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__RateLimitIn =
  {
    connections?: InputMaybe<Scalars['Int']['input']>;
    enabled?: InputMaybe<Scalars['Boolean']['input']>;
    rpm?: InputMaybe<Scalars['Int']['input']>;
    rps?: InputMaybe<Scalars['Int']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__RouteIn = {
  app: Scalars['String']['input'];
  path: Scalars['String']['input'];
  port: Scalars['Int']['input'];
  rewrite?: InputMaybe<Scalars['Boolean']['input']>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppSvcIn = {
  port: Scalars['Int']['input'];
  protocol?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__TolerationIn = {
  effect?: InputMaybe<K8s__Io___Api___Core___V1__TaintEffect>;
  key?: InputMaybe<Scalars['String']['input']>;
  operator?: InputMaybe<K8s__Io___Api___Core___V1__TolerationOperator>;
  tolerationSeconds?: InputMaybe<Scalars['Int']['input']>;
  value?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__TopologySpreadConstraintIn = {
  labelSelector?: InputMaybe<K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorIn>;
  matchLabelKeys?: InputMaybe<Array<Scalars['String']['input']>>;
  maxSkew: Scalars['Int']['input'];
  minDomains?: InputMaybe<Scalars['Int']['input']>;
  nodeAffinityPolicy?: InputMaybe<Scalars['String']['input']>;
  nodeTaintsPolicy?: InputMaybe<Scalars['String']['input']>;
  topologyKey: Scalars['String']['input'];
  whenUnsatisfiable: K8s__Io___Api___Core___V1__UnsatisfiableConstraintAction;
};

export type K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorIn = {
  matchExpressions?: InputMaybe<
    Array<K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorRequirementIn>
  >;
  matchLabels?: InputMaybe<Scalars['Map']['input']>;
};

export type K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorRequirementIn =
  {
    key: Scalars['String']['input'];
    operator: K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator;
    values?: InputMaybe<Array<Scalars['String']['input']>>;
  };

export type ConfigIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  binaryData?: InputMaybe<Scalars['Map']['input']>;
  data?: InputMaybe<Scalars['Map']['input']>;
  displayName: Scalars['String']['input'];
  immutable?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
};

export type EnvironmentIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  clusterName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentSpecIn =
  {
    routing?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentRoutingIn>;
    targetNamespace?: InputMaybe<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentRoutingIn =
  {
    mode?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentRoutingMode>;
  };

export type ExternalAppIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ExternalAppSpecIn>;
  status?: InputMaybe<Github__Com___Kloudlite___Operator___Pkg___Operator__StatusIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ExternalAppSpecIn =
  {
    intercept?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__InterceptIn>;
    record: Scalars['String']['input'];
    recordType: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ExternalAppRecordType;
  };

export type Github__Com___Kloudlite___Operator___Pkg___Operator__StatusIn = {
  checkList?: InputMaybe<
    Array<Github__Com___Kloudlite___Operator___Pkg___Operator__CheckMetaIn>
  >;
  checks?: InputMaybe<Scalars['Map']['input']>;
  isReady: Scalars['Boolean']['input'];
  lastReadyGeneration?: InputMaybe<Scalars['Int']['input']>;
  lastReconcileTime?: InputMaybe<Scalars['Date']['input']>;
  message?: InputMaybe<Github__Com___Kloudlite___Operator___Pkg___Raw____Json__RawJsonIn>;
  resources?: InputMaybe<
    Array<Github__Com___Kloudlite___Operator___Pkg___Operator__ResourceRefIn>
  >;
};

export type Github__Com___Kloudlite___Operator___Pkg___Operator__CheckMetaIn = {
  debug?: InputMaybe<Scalars['Boolean']['input']>;
  description?: InputMaybe<Scalars['String']['input']>;
  hide?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  title: Scalars['String']['input'];
};

export type Github__Com___Kloudlite___Operator___Pkg___Raw____Json__RawJsonIn =
  {
    RawMessage?: InputMaybe<Scalars['Any']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Pkg___Operator__ResourceRefIn =
  {
    apiVersion: Scalars['String']['input'];
    kind: Scalars['String']['input'];
    name: Scalars['String']['input'];
    namespace: Scalars['String']['input'];
  };

export type ImagePullSecretIn = {
  displayName: Scalars['String']['input'];
  dockerConfigJson?: InputMaybe<Scalars['String']['input']>;
  environments?: InputMaybe<Array<Scalars['String']['input']>>;
  format: Github__Com___Kloudlite___Api___Apps___Console___Internal___Entities__PullSecretFormat;
  metadata: MetadataIn;
  registryPassword?: InputMaybe<Scalars['String']['input']>;
  registryURL?: InputMaybe<Scalars['String']['input']>;
  registryUsername?: InputMaybe<Scalars['String']['input']>;
};

export type ManagedResourceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedResourceSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedResourceSpecIn =
  {
    resourceNamePrefix?: InputMaybe<Scalars['String']['input']>;
    resourceTemplate: Github__Com___Kloudlite___Operator___Apis___Crds___V1__MresResourceTemplateIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__MresResourceTemplateIn =
  {
    apiVersion: Scalars['String']['input'];
    kind: Scalars['String']['input'];
    msvcRef: Github__Com___Kloudlite___Operator___Apis___Common____Types__MsvcRefIn;
    spec: Scalars['Map']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__MsvcRefIn =
  {
    apiVersion?: InputMaybe<Scalars['String']['input']>;
    clusterName?: InputMaybe<Scalars['String']['input']>;
    kind?: InputMaybe<Scalars['String']['input']>;
    name: Scalars['String']['input'];
    namespace: Scalars['String']['input'];
  };

export type RouterIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__RouterSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__RouterSpecIn =
  {
    backendProtocol?: InputMaybe<Scalars['String']['input']>;
    basicAuth?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__BasicAuthIn>;
    cors?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__CorsIn>;
    domains: Array<Scalars['String']['input']>;
    https?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__HttpsIn>;
    ingressClass?: InputMaybe<Scalars['String']['input']>;
    maxBodySizeInMB?: InputMaybe<Scalars['Int']['input']>;
    rateLimit?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__RateLimitIn>;
    routes?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__RouteIn>
    >;
  };

export type SecretIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  data?: InputMaybe<Scalars['Map']['input']>;
  displayName: Scalars['String']['input'];
  immutable?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  stringData?: InputMaybe<Scalars['Map']['input']>;
  type?: InputMaybe<K8s__Io___Api___Core___V1__SecretType>;
};

export type ConsoleVpnDeviceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  clusterName?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  environmentName?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__DeviceSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__DeviceSpecIn =
  {
    activeNamespace?: InputMaybe<Scalars['String']['input']>;
    cnameRecords?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__CNameRecordIn>
    >;
    ports?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__PortIn>
    >;
  };

export type Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__CNameRecordIn =
  {
    host?: InputMaybe<Scalars['String']['input']>;
    target?: InputMaybe<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__PortIn =
  {
    port?: InputMaybe<Scalars['Int']['input']>;
    targetPort?: InputMaybe<Scalars['Int']['input']>;
  };

export type PortIn = {
  port?: InputMaybe<Scalars['Int']['input']>;
  targetPort?: InputMaybe<Scalars['Int']['input']>;
};

export type BuildIn = {
  buildClusterName: Scalars['String']['input'];
  name: Scalars['String']['input'];
  source: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitSourceIn;
  spec: Github__Com___Kloudlite___Operator___Apis___Distribution___V1__BuildRunSpecIn;
};

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitSourceIn =
  {
    branch: Scalars['String']['input'];
    provider: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitProvider;
    repository: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Distribution___V1__BuildRunSpecIn =
  {
    buildOptions?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Distribution___V1__BuildOptionsIn>;
    caches?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Distribution___V1__CacheIn>
    >;
    registry: Github__Com___Kloudlite___Operator___Apis___Distribution___V1__RegistryIn;
    resource: Github__Com___Kloudlite___Operator___Apis___Distribution___V1__ResourceIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Distribution___V1__BuildOptionsIn =
  {
    buildArgs?: InputMaybe<Scalars['Map']['input']>;
    buildContexts?: InputMaybe<Scalars['Map']['input']>;
    contextDir?: InputMaybe<Scalars['String']['input']>;
    dockerfileContent?: InputMaybe<Scalars['String']['input']>;
    dockerfilePath?: InputMaybe<Scalars['String']['input']>;
    targetPlatforms?: InputMaybe<Array<Scalars['String']['input']>>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Distribution___V1__CacheIn =
  {
    name: Scalars['String']['input'];
    path: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Distribution___V1__RegistryIn =
  {
    repo: Github__Com___Kloudlite___Operator___Apis___Distribution___V1__RepoIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Distribution___V1__RepoIn =
  {
    name: Scalars['String']['input'];
    tags: Array<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Distribution___V1__ResourceIn =
  {
    cpu: Scalars['Int']['input'];
    memoryInMb: Scalars['Int']['input'];
  };

export type CredentialIn = {
  access: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__RepoAccess;
  expiration: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__ExpirationIn;
  name: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__ExpirationIn =
  {
    unit: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__ExpirationUnit;
    value: Scalars['Int']['input'];
  };

export type RepositoryIn = {
  name: Scalars['String']['input'];
};

export type ByokClusterIn = {
  displayName: Scalars['String']['input'];
  metadata: MetadataIn;
};

export type ClusterIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  globalVPN?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecIn =
  {
    availabilityMode: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode;
    aws?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsClusterConfigIn>;
    cloudflareEnabled?: InputMaybe<Scalars['Boolean']['input']>;
    cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
    gcp?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpClusterConfigIn>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsClusterConfigIn =
  {
    credentials: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsCredentialsIn;
    k3sMasters?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__Awsk3sMastersConfigIn>;
    region: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsCredentialsIn =
  {
    authMechanism: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsAuthMechanism;
    secretRef: Github__Com___Kloudlite___Operator___Apis___Common____Types__SecretRefIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__SecretRefIn =
  {
    name: Scalars['String']['input'];
    namespace?: InputMaybe<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__Awsk3sMastersConfigIn =
  {
    instanceType: Scalars['String']['input'];
    nvidiaGpuEnabled: Scalars['Boolean']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpClusterConfigIn =
  {
    credentialsRef: Github__Com___Kloudlite___Operator___Apis___Common____Types__SecretRefIn;
    region: Scalars['String']['input'];
  };

export type ClusterManagedServiceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  clusterName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ClusterManagedServiceSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ClusterManagedServiceSpecIn =
  {
    msvcSpec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedServiceSpecIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedServiceSpecIn =
  {
    nodeSelector?: InputMaybe<Scalars['Map']['input']>;
    serviceTemplate: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ServiceTemplateIn;
    tolerations?: InputMaybe<Array<K8s__Io___Api___Core___V1__TolerationIn>>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ServiceTemplateIn =
  {
    apiVersion: Scalars['String']['input'];
    kind: Scalars['String']['input'];
    spec?: InputMaybe<Scalars['Map']['input']>;
  };

export type DomainEntryIn = {
  clusterName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  domainName: Scalars['String']['input'];
};

export type GlobalVpnIn = {
  allocatableCIDRSuffix: Scalars['Int']['input'];
  CIDR: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  kloudliteDevice: GlobalVpnKloudliteDeviceIn;
  metadata: MetadataIn;
  numAllocatedClusterCIDRs: Scalars['Int']['input'];
  numAllocatedDevices: Scalars['Int']['input'];
  numReservedIPsForNonClusterUse: Scalars['Int']['input'];
  wgInterface: Scalars['String']['input'];
};

export type GlobalVpnKloudliteDeviceIn = {
  ipAddr: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type GlobalVpnDeviceIn = {
  creationMethod?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  globalVPNName: Scalars['String']['input'];
  metadata: MetadataIn;
};

export type HelmReleaseIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__HelmChartSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__HelmChartSpecIn =
  {
    chartName: Scalars['String']['input'];
    chartRepoURL: Scalars['String']['input'];
    chartVersion: Scalars['String']['input'];
    jobVars?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__JobVarsIn>;
    postInstall?: InputMaybe<Scalars['String']['input']>;
    postUninstall?: InputMaybe<Scalars['String']['input']>;
    preInstall?: InputMaybe<Scalars['String']['input']>;
    preUninstall?: InputMaybe<Scalars['String']['input']>;
    values: Scalars['Map']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__JobVarsIn = {
  affinity?: InputMaybe<K8s__Io___Api___Core___V1__AffinityIn>;
  backOffLimit?: InputMaybe<Scalars['Int']['input']>;
  nodeSelector?: InputMaybe<Scalars['Map']['input']>;
  tolerations?: InputMaybe<Array<K8s__Io___Api___Core___V1__TolerationIn>>;
};

export type K8s__Io___Api___Core___V1__AffinityIn = {
  nodeAffinity?: InputMaybe<K8s__Io___Api___Core___V1__NodeAffinityIn>;
  podAffinity?: InputMaybe<K8s__Io___Api___Core___V1__PodAffinityIn>;
  podAntiAffinity?: InputMaybe<K8s__Io___Api___Core___V1__PodAntiAffinityIn>;
};

export type K8s__Io___Api___Core___V1__NodeAffinityIn = {
  preferredDuringSchedulingIgnoredDuringExecution?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__PreferredSchedulingTermIn>
  >;
  requiredDuringSchedulingIgnoredDuringExecution?: InputMaybe<K8s__Io___Api___Core___V1__NodeSelectorIn>;
};

export type K8s__Io___Api___Core___V1__PreferredSchedulingTermIn = {
  preference: K8s__Io___Api___Core___V1__NodeSelectorTermIn;
  weight: Scalars['Int']['input'];
};

export type K8s__Io___Api___Core___V1__NodeSelectorTermIn = {
  matchExpressions?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__NodeSelectorRequirementIn>
  >;
  matchFields?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__NodeSelectorRequirementIn>
  >;
};

export type K8s__Io___Api___Core___V1__NodeSelectorRequirementIn = {
  key: Scalars['String']['input'];
  operator: K8s__Io___Api___Core___V1__NodeSelectorOperator;
  values?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type K8s__Io___Api___Core___V1__NodeSelectorIn = {
  nodeSelectorTerms: Array<K8s__Io___Api___Core___V1__NodeSelectorTermIn>;
};

export type K8s__Io___Api___Core___V1__PodAffinityIn = {
  preferredDuringSchedulingIgnoredDuringExecution?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__WeightedPodAffinityTermIn>
  >;
  requiredDuringSchedulingIgnoredDuringExecution?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__PodAffinityTermIn>
  >;
};

export type K8s__Io___Api___Core___V1__WeightedPodAffinityTermIn = {
  podAffinityTerm: K8s__Io___Api___Core___V1__PodAffinityTermIn;
  weight: Scalars['Int']['input'];
};

export type K8s__Io___Api___Core___V1__PodAffinityTermIn = {
  labelSelector?: InputMaybe<K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorIn>;
  namespaces?: InputMaybe<Array<Scalars['String']['input']>>;
  namespaceSelector?: InputMaybe<K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorIn>;
  topologyKey: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__PodAntiAffinityIn = {
  preferredDuringSchedulingIgnoredDuringExecution?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__WeightedPodAffinityTermIn>
  >;
  requiredDuringSchedulingIgnoredDuringExecution?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__PodAffinityTermIn>
  >;
};

export type NodePoolIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodePoolSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodePoolSpecIn =
  {
    aws?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsNodePoolConfigIn>;
    cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
    gcp?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpNodePoolConfigIn>;
    maxCount: Scalars['Int']['input'];
    minCount: Scalars['Int']['input'];
    nodeLabels?: InputMaybe<Scalars['Map']['input']>;
    nodeTaints?: InputMaybe<Array<K8s__Io___Api___Core___V1__TaintIn>>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsNodePoolConfigIn =
  {
    availabilityZone: Scalars['String']['input'];
    ec2Pool?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsEc2PoolConfigIn>;
    nvidiaGpuEnabled: Scalars['Boolean']['input'];
    poolType: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType;
    spotPool?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsSpotPoolConfigIn>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsEc2PoolConfigIn =
  {
    instanceType: Scalars['String']['input'];
    nodes?: InputMaybe<Scalars['Map']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsSpotPoolConfigIn =
  {
    cpuNode?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsSpotCpuNodeIn>;
    gpuNode?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsSpotGpuNodeIn>;
    nodes?: InputMaybe<Scalars['Map']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsSpotCpuNodeIn =
  {
    memoryPerVcpu?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Common____Types__MinMaxFloatIn>;
    vcpu: Github__Com___Kloudlite___Operator___Apis___Common____Types__MinMaxFloatIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__MinMaxFloatIn =
  {
    max: Scalars['String']['input'];
    min: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsSpotGpuNodeIn =
  {
    instanceTypes: Array<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpNodePoolConfigIn =
  {
    availabilityZone: Scalars['String']['input'];
    machineType: Scalars['String']['input'];
    poolType: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpPoolType;
  };

export type K8s__Io___Api___Core___V1__TaintIn = {
  effect: K8s__Io___Api___Core___V1__TaintEffect;
  key: Scalars['String']['input'];
  timeAdded?: InputMaybe<Scalars['Date']['input']>;
  value?: InputMaybe<Scalars['String']['input']>;
};

export type CloudProviderSecretIn = {
  aws?: InputMaybe<Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__AwsSecretCredentialsIn>;
  cloudProviderName: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
  displayName: Scalars['String']['input'];
  gcp?: InputMaybe<Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__GcpSecretCredentialsIn>;
  metadata: MetadataIn;
};

export type Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__AwsSecretCredentialsIn =
  {
    assumeRoleParams?: InputMaybe<Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__AwsAssumeRoleParamsIn>;
    authMechanism: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsAuthMechanism;
    authSecretKeys?: InputMaybe<Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__AwsAuthSecretKeysIn>;
  };

export type Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__AwsAssumeRoleParamsIn =
  {
    awsAccountId: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__AwsAuthSecretKeysIn =
  {
    accessKey: Scalars['String']['input'];
    secretKey: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__GcpSecretCredentialsIn =
  {
    serviceAccountJSON: Scalars['String']['input'];
  };

export type IotAppIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppSpecIn;
};

export type IotDeploymentIn = {
  CIDR: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  exposedServices: Array<Github__Com___Kloudlite___Api___Apps___Iot____Console___Internal___Entities__ExposedServiceIn>;
  name: Scalars['String']['input'];
};

export type Github__Com___Kloudlite___Api___Apps___Iot____Console___Internal___Entities__ExposedServiceIn =
  {
    ip: Scalars['String']['input'];
    name: Scalars['String']['input'];
  };

export type IotDeviceIn = {
  displayName: Scalars['String']['input'];
  ip: Scalars['String']['input'];
  name: Scalars['String']['input'];
  podCIDR: Scalars['String']['input'];
  publicKey: Scalars['String']['input'];
  serviceCIDR: Scalars['String']['input'];
  version: Scalars['String']['input'];
};

export type IotDeviceBlueprintIn = {
  bluePrintType: Github__Com___Kloudlite___Api___Apps___Iot____Console___Internal___Entities__BluePrintType;
  displayName: Scalars['String']['input'];
  name: Scalars['String']['input'];
  version: Scalars['String']['input'];
};

export type IotProjectIn = {
  displayName: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type AccountMembershipIn = {
  accountName: Scalars['String']['input'];
  role: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
  userId: Scalars['String']['input'];
};

export type BuildRunIn = {
  displayName: Scalars['String']['input'];
};

export type ConfigKeyValueRefIn = {
  configName: Scalars['String']['input'];
  key: Scalars['String']['input'];
  value: Scalars['String']['input'];
};

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GithubUserAccountIn =
  {
    avatarUrl?: InputMaybe<Scalars['String']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    login?: InputMaybe<Scalars['String']['input']>;
    nodeId?: InputMaybe<Scalars['String']['input']>;
    type?: InputMaybe<Scalars['String']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodePropsIn =
  {
    lastRecreatedAt?: InputMaybe<Scalars['Date']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodeSpecIn =
  {
    nodepoolName: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Pkg___Operator__State =
  | 'errored____during____reconcilation'
  | 'finished____reconcilation'
  | 'under____reconcilation'
  | 'yet____to____be____reconciled';

export type Github__Com___Kloudlite___Operator___Pkg___Operator__CheckIn = {
  debug?: InputMaybe<Scalars['String']['input']>;
  error?: InputMaybe<Scalars['String']['input']>;
  generation?: InputMaybe<Scalars['Int']['input']>;
  info?: InputMaybe<Scalars['String']['input']>;
  message?: InputMaybe<Scalars['String']['input']>;
  startedAt?: InputMaybe<Scalars['Date']['input']>;
  state?: InputMaybe<Github__Com___Kloudlite___Operator___Pkg___Operator__State>;
  status: Scalars['Boolean']['input'];
};

export type IotEnvironmentIn = {
  displayName: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__AwsElasticBlockStoreVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  partition?: InputMaybe<Scalars['Int']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  volumeID: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__AzureDiskVolumeSourceIn = {
  cachingMode?: InputMaybe<Scalars['String']['input']>;
  diskName: Scalars['String']['input'];
  diskURI: Scalars['String']['input'];
  fsType?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
};

export type K8s__Io___Api___Core___V1__AzureFilePersistentVolumeSourceIn = {
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretName: Scalars['String']['input'];
  secretNamespace?: InputMaybe<Scalars['String']['input']>;
  shareName: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__CephFsPersistentVolumeSourceIn = {
  monitors: Array<Scalars['String']['input']>;
  path?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretFile?: InputMaybe<Scalars['String']['input']>;
  secretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  user?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__SecretReferenceIn = {
  name?: InputMaybe<Scalars['String']['input']>;
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__CinderPersistentVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  volumeID: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__CsiPersistentVolumeSourceIn = {
  controllerExpandSecretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  controllerPublishSecretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  driver: Scalars['String']['input'];
  fsType?: InputMaybe<Scalars['String']['input']>;
  nodeExpandSecretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  nodePublishSecretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  nodeStageSecretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  volumeAttributes?: InputMaybe<Scalars['Map']['input']>;
  volumeHandle: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__FcVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  lun?: InputMaybe<Scalars['Int']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  targetWWNs?: InputMaybe<Array<Scalars['String']['input']>>;
  wwids?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type K8s__Io___Api___Core___V1__FlexPersistentVolumeSourceIn = {
  driver: Scalars['String']['input'];
  fsType?: InputMaybe<Scalars['String']['input']>;
  options?: InputMaybe<Scalars['Map']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
};

export type K8s__Io___Api___Core___V1__FlockerVolumeSourceIn = {
  datasetName?: InputMaybe<Scalars['String']['input']>;
  datasetUUID?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__GcePersistentDiskVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  partition?: InputMaybe<Scalars['Int']['input']>;
  pdName: Scalars['String']['input'];
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
};

export type K8s__Io___Api___Core___V1__GlusterfsPersistentVolumeSourceIn = {
  endpoints: Scalars['String']['input'];
  endpointsNamespace?: InputMaybe<Scalars['String']['input']>;
  path: Scalars['String']['input'];
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
};

export type K8s__Io___Api___Core___V1__HostPathVolumeSourceIn = {
  path: Scalars['String']['input'];
  type?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__IscsiPersistentVolumeSourceIn = {
  chapAuthDiscovery?: InputMaybe<Scalars['Boolean']['input']>;
  chapAuthSession?: InputMaybe<Scalars['Boolean']['input']>;
  fsType?: InputMaybe<Scalars['String']['input']>;
  initiatorName?: InputMaybe<Scalars['String']['input']>;
  iqn: Scalars['String']['input'];
  iscsiInterface?: InputMaybe<Scalars['String']['input']>;
  lun: Scalars['Int']['input'];
  portals?: InputMaybe<Array<Scalars['String']['input']>>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  targetPortal: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__LocalVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  path: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__NamespaceConditionIn = {
  lastTransitionTime?: InputMaybe<Scalars['Date']['input']>;
  message?: InputMaybe<Scalars['String']['input']>;
  reason?: InputMaybe<Scalars['String']['input']>;
  status: K8s__Io___Api___Core___V1__ConditionStatus;
  type: K8s__Io___Api___Core___V1__NamespaceConditionType;
};

export type K8s__Io___Api___Core___V1__NamespaceSpecIn = {
  finalizers?: InputMaybe<Array<Scalars['String']['input']>>;
};

export type K8s__Io___Api___Core___V1__NamespaceStatusIn = {
  conditions?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__NamespaceConditionIn>
  >;
  phase?: InputMaybe<K8s__Io___Api___Core___V1__NamespacePhase>;
};

export type K8s__Io___Api___Core___V1__NfsVolumeSourceIn = {
  path: Scalars['String']['input'];
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  server: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__ObjectReferenceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  fieldPath?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  namespace?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  uid?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__PersistentVolumeClaimConditionIn = {
  lastProbeTime?: InputMaybe<Scalars['Date']['input']>;
  lastTransitionTime?: InputMaybe<Scalars['Date']['input']>;
  message?: InputMaybe<Scalars['String']['input']>;
  reason?: InputMaybe<Scalars['String']['input']>;
  status: K8s__Io___Api___Core___V1__ConditionStatus;
  type: K8s__Io___Api___Core___V1__PersistentVolumeClaimConditionType;
};

export type K8s__Io___Api___Core___V1__PersistentVolumeClaimSpecIn = {
  accessModes?: InputMaybe<Array<Scalars['String']['input']>>;
  dataSource?: InputMaybe<K8s__Io___Api___Core___V1__TypedLocalObjectReferenceIn>;
  dataSourceRef?: InputMaybe<K8s__Io___Api___Core___V1__TypedObjectReferenceIn>;
  resources?: InputMaybe<K8s__Io___Api___Core___V1__ResourceRequirementsIn>;
  selector?: InputMaybe<K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorIn>;
  storageClassName?: InputMaybe<Scalars['String']['input']>;
  volumeMode?: InputMaybe<Scalars['String']['input']>;
  volumeName?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__TypedLocalObjectReferenceIn = {
  apiGroup?: InputMaybe<Scalars['String']['input']>;
  kind: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__TypedObjectReferenceIn = {
  apiGroup?: InputMaybe<Scalars['String']['input']>;
  kind: Scalars['String']['input'];
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__ResourceRequirementsIn = {
  claims?: InputMaybe<Array<K8s__Io___Api___Core___V1__ResourceClaimIn>>;
  limits?: InputMaybe<Scalars['Map']['input']>;
  requests?: InputMaybe<Scalars['Map']['input']>;
};

export type K8s__Io___Api___Core___V1__ResourceClaimIn = {
  name: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__PersistentVolumeClaimStatusIn = {
  accessModes?: InputMaybe<Array<Scalars['String']['input']>>;
  allocatedResources?: InputMaybe<Scalars['Map']['input']>;
  allocatedResourceStatuses?: InputMaybe<Scalars['Map']['input']>;
  capacity?: InputMaybe<Scalars['Map']['input']>;
  conditions?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__PersistentVolumeClaimConditionIn>
  >;
  phase?: InputMaybe<K8s__Io___Api___Core___V1__PersistentVolumeClaimPhase>;
};

export type K8s__Io___Api___Core___V1__PersistentVolumeSpecIn = {
  accessModes?: InputMaybe<Array<Scalars['String']['input']>>;
  awsElasticBlockStore?: InputMaybe<K8s__Io___Api___Core___V1__AwsElasticBlockStoreVolumeSourceIn>;
  azureDisk?: InputMaybe<K8s__Io___Api___Core___V1__AzureDiskVolumeSourceIn>;
  azureFile?: InputMaybe<K8s__Io___Api___Core___V1__AzureFilePersistentVolumeSourceIn>;
  capacity?: InputMaybe<Scalars['Map']['input']>;
  cephfs?: InputMaybe<K8s__Io___Api___Core___V1__CephFsPersistentVolumeSourceIn>;
  cinder?: InputMaybe<K8s__Io___Api___Core___V1__CinderPersistentVolumeSourceIn>;
  claimRef?: InputMaybe<K8s__Io___Api___Core___V1__ObjectReferenceIn>;
  csi?: InputMaybe<K8s__Io___Api___Core___V1__CsiPersistentVolumeSourceIn>;
  fc?: InputMaybe<K8s__Io___Api___Core___V1__FcVolumeSourceIn>;
  flexVolume?: InputMaybe<K8s__Io___Api___Core___V1__FlexPersistentVolumeSourceIn>;
  flocker?: InputMaybe<K8s__Io___Api___Core___V1__FlockerVolumeSourceIn>;
  gcePersistentDisk?: InputMaybe<K8s__Io___Api___Core___V1__GcePersistentDiskVolumeSourceIn>;
  glusterfs?: InputMaybe<K8s__Io___Api___Core___V1__GlusterfsPersistentVolumeSourceIn>;
  hostPath?: InputMaybe<K8s__Io___Api___Core___V1__HostPathVolumeSourceIn>;
  iscsi?: InputMaybe<K8s__Io___Api___Core___V1__IscsiPersistentVolumeSourceIn>;
  local?: InputMaybe<K8s__Io___Api___Core___V1__LocalVolumeSourceIn>;
  mountOptions?: InputMaybe<Array<Scalars['String']['input']>>;
  nfs?: InputMaybe<K8s__Io___Api___Core___V1__NfsVolumeSourceIn>;
  nodeAffinity?: InputMaybe<K8s__Io___Api___Core___V1__VolumeNodeAffinityIn>;
  persistentVolumeReclaimPolicy?: InputMaybe<K8s__Io___Api___Core___V1__PersistentVolumeReclaimPolicy>;
  photonPersistentDisk?: InputMaybe<K8s__Io___Api___Core___V1__PhotonPersistentDiskVolumeSourceIn>;
  portworxVolume?: InputMaybe<K8s__Io___Api___Core___V1__PortworxVolumeSourceIn>;
  quobyte?: InputMaybe<K8s__Io___Api___Core___V1__QuobyteVolumeSourceIn>;
  rbd?: InputMaybe<K8s__Io___Api___Core___V1__RbdPersistentVolumeSourceIn>;
  scaleIO?: InputMaybe<K8s__Io___Api___Core___V1__ScaleIoPersistentVolumeSourceIn>;
  storageClassName?: InputMaybe<Scalars['String']['input']>;
  storageos?: InputMaybe<K8s__Io___Api___Core___V1__StorageOsPersistentVolumeSourceIn>;
  volumeMode?: InputMaybe<Scalars['String']['input']>;
  vsphereVolume?: InputMaybe<K8s__Io___Api___Core___V1__VsphereVirtualDiskVolumeSourceIn>;
};

export type K8s__Io___Api___Core___V1__VolumeNodeAffinityIn = {
  required?: InputMaybe<K8s__Io___Api___Core___V1__NodeSelectorIn>;
};

export type K8s__Io___Api___Core___V1__PhotonPersistentDiskVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  pdID: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__PortworxVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  volumeID: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__QuobyteVolumeSourceIn = {
  group?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  registry: Scalars['String']['input'];
  tenant?: InputMaybe<Scalars['String']['input']>;
  user?: InputMaybe<Scalars['String']['input']>;
  volume: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__RbdPersistentVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  image: Scalars['String']['input'];
  keyring?: InputMaybe<Scalars['String']['input']>;
  monitors: Array<Scalars['String']['input']>;
  pool?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  user?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__ScaleIoPersistentVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  gateway: Scalars['String']['input'];
  protectionDomain?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretRef?: InputMaybe<K8s__Io___Api___Core___V1__SecretReferenceIn>;
  sslEnabled?: InputMaybe<Scalars['Boolean']['input']>;
  storageMode?: InputMaybe<Scalars['String']['input']>;
  storagePool?: InputMaybe<Scalars['String']['input']>;
  system: Scalars['String']['input'];
  volumeName?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__StorageOsPersistentVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  readOnly?: InputMaybe<Scalars['Boolean']['input']>;
  secretRef?: InputMaybe<K8s__Io___Api___Core___V1__ObjectReferenceIn>;
  volumeName?: InputMaybe<Scalars['String']['input']>;
  volumeNamespace?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__VsphereVirtualDiskVolumeSourceIn = {
  fsType?: InputMaybe<Scalars['String']['input']>;
  storagePolicyID?: InputMaybe<Scalars['String']['input']>;
  storagePolicyName?: InputMaybe<Scalars['String']['input']>;
  volumePath: Scalars['String']['input'];
};

export type K8s__Io___Api___Core___V1__PersistentVolumeStatusIn = {
  lastPhaseTransitionTime?: InputMaybe<Scalars['Date']['input']>;
  message?: InputMaybe<Scalars['String']['input']>;
  phase?: InputMaybe<K8s__Io___Api___Core___V1__PersistentVolumePhase>;
  reason?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Storage___V1__VolumeAttachmentSourceIn = {
  inlineVolumeSpec?: InputMaybe<K8s__Io___Api___Core___V1__PersistentVolumeSpecIn>;
  persistentVolumeName?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Storage___V1__VolumeAttachmentSpecIn = {
  attacher: Scalars['String']['input'];
  nodeName: Scalars['String']['input'];
  source: K8s__Io___Api___Storage___V1__VolumeAttachmentSourceIn;
};

export type K8s__Io___Api___Storage___V1__VolumeAttachmentStatusIn = {
  attached: Scalars['Boolean']['input'];
  attachError?: InputMaybe<K8s__Io___Api___Storage___V1__VolumeErrorIn>;
  attachmentMetadata?: InputMaybe<Scalars['Map']['input']>;
  detachError?: InputMaybe<K8s__Io___Api___Storage___V1__VolumeErrorIn>;
};

export type K8s__Io___Api___Storage___V1__VolumeErrorIn = {
  message?: InputMaybe<Scalars['String']['input']>;
  time?: InputMaybe<Scalars['Date']['input']>;
};

export type K8s__Io___Apimachinery___Pkg___Api___Resource__Format =
  | 'BinarySI'
  | 'DecimalExponent'
  | 'DecimalSI';

export type K8s__Io___Apimachinery___Pkg___Api___Resource__QuantityIn = {
  Format: K8s__Io___Apimachinery___Pkg___Api___Resource__Format;
};

export type ManagedResourceKeyValueRefIn = {
  key: Scalars['String']['input'];
  mresName: Scalars['String']['input'];
  value: Scalars['String']['input'];
};

export type NamespaceIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<K8s__Io___Api___Core___V1__NamespaceSpecIn>;
  status?: InputMaybe<K8s__Io___Api___Core___V1__NamespaceStatusIn>;
};

export type NodeIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodeSpecIn;
};

export type PersistentVolumeIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<K8s__Io___Api___Core___V1__PersistentVolumeSpecIn>;
  status?: InputMaybe<K8s__Io___Api___Core___V1__PersistentVolumeStatusIn>;
};

export type SearchProjectManagedService = {
  isReady?: InputMaybe<MatchFilterIn>;
  managedServiceName?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchProjects = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SecretKeyValueRefIn = {
  key: Scalars['String']['input'];
  secretName: Scalars['String']['input'];
  value: Scalars['String']['input'];
};

export type VolumeAttachmentIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: K8s__Io___Api___Storage___V1__VolumeAttachmentSpecIn;
  status?: InputMaybe<K8s__Io___Api___Storage___V1__VolumeAttachmentStatusIn>;
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

export type ConsoleCrCheckNameAvailabilityQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleCrCheckNameAvailabilityQuery = {
  cr_checkUserNameAvailability: {
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
  envName?: InputMaybe<Scalars['String']['input']>;
}>;

export type ConsoleCoreCheckNameAvailabilityQuery = {
  core_checkNameAvailability: { result: boolean };
};

export type ConsoleWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type ConsoleWhoAmIQuery = {
  auth_me?: {
    id: string;
    email: string;
    providerGitlab?: any;
    providerGithub?: any;
    providerGoogle?: any;
  };
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
    metadata?: { name: string; annotations?: any };
  }>;
};

export type ConsoleUpdateAccountMutationVariables = Exact<{
  account: AccountIn;
}>;

export type ConsoleUpdateAccountMutation = {
  accounts_updateAccount: { id: string };
};

export type ConsoleGetAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleGetAccountQuery = {
  accounts_getAccount?: {
    targetNamespace?: string;
    updateTime: any;
    contactEmail?: string;
    displayName: string;
    metadata?: { name: string; annotations?: any };
  };
};

export type ConsoleDeleteAccountMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleDeleteAccountMutation = { accounts_deleteAccount: boolean };

export type ConsoleListDnsHostsQueryVariables = Exact<{ [key: string]: never }>;

export type ConsoleListDnsHostsQuery = {
  infra_listClusters?: {
    edges: Array<{
      node: {
        metadata: { name: string; namespace?: string };
        spec: { publicDNSHost: string };
      };
    }>;
  };
};

export type ConsoleCreateClusterMutationVariables = Exact<{
  cluster: ClusterIn;
}>;

export type ConsoleCreateClusterMutation = {
  infra_createCluster?: { id: string };
};

export type ConsoleDeleteClusterMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleDeleteClusterMutation = { infra_deleteCluster: boolean };

export type ConsoleClustersCountQueryVariables = Exact<{
  [key: string]: never;
}>;

export type ConsoleClustersCountQuery = {
  infra_listClusters?: { totalCount: number };
};

export type ConsoleListAllClustersQueryVariables = Exact<{
  search?: InputMaybe<SearchCluster>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListAllClustersQuery = {
  byok_clusters?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        clusterPublicEndpoint: string;
        clusterSvcCIDR: string;
        creationTime: any;
        displayName: string;
        globalVPN: string;
        id: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
      };
    }>;
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPreviousPage?: boolean;
      startCursor?: string;
    };
  };
  clusters?: {
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
        id: string;
        displayName: string;
        markedForDeletion?: boolean;
        creationTime: any;
        updateTime: any;
        recordVersion: number;
        metadata: { name: string; annotations?: any; generation: number };
        lastUpdatedBy: { userId: string; userName: string; userEmail: string };
        createdBy: { userEmail: string; userId: string; userName: string };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
        spec: {
          messageQueueTopicName: string;
          kloudliteRelease: string;
          accountId: string;
          accountName: string;
          availabilityMode: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode;
          cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
          backupToS3Enabled: boolean;
          cloudflareEnabled?: boolean;
          clusterInternalDnsHost?: string;
          clusterServiceCIDR?: string;
          publicDNSHost: string;
          taintMasterNodes: boolean;
          clusterTokenRef?: { key: string; name: string; namespace?: string };
          aws?: {
            nodePools?: any;
            region: string;
            spotNodePools?: any;
            k3sMasters?: {
              iamInstanceProfileRole?: string;
              instanceType: string;
              nodes?: any;
              nvidiaGpuEnabled: boolean;
              rootVolumeSize: number;
              rootVolumeType: string;
            };
          };
          gcp?: {
            gcpProjectID: string;
            region: string;
            credentialsRef: { name: string; namespace?: string };
          };
          output?: {
            keyK3sAgentJoinToken: string;
            keyK3sServerJoinToken: string;
            keyKubeconfig: string;
            secretName: string;
          };
        };
      };
    }>;
  };
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
        id: string;
        displayName: string;
        markedForDeletion?: boolean;
        creationTime: any;
        updateTime: any;
        recordVersion: number;
        metadata: { name: string; annotations?: any; generation: number };
        lastUpdatedBy: { userId: string; userName: string; userEmail: string };
        createdBy: { userEmail: string; userId: string; userName: string };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
        spec: {
          messageQueueTopicName: string;
          kloudliteRelease: string;
          accountId: string;
          accountName: string;
          availabilityMode: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode;
          cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
          backupToS3Enabled: boolean;
          cloudflareEnabled?: boolean;
          clusterInternalDnsHost?: string;
          publicDNSHost: string;
          taintMasterNodes: boolean;
          clusterTokenRef?: { key: string; name: string; namespace?: string };
          aws?: {
            nodePools?: any;
            region: string;
            spotNodePools?: any;
            k3sMasters?: {
              iamInstanceProfileRole?: string;
              instanceType: string;
              nodes?: any;
              nvidiaGpuEnabled: boolean;
              rootVolumeSize: number;
              rootVolumeType: string;
            };
          };
          gcp?: {
            gcpProjectID: string;
            region: string;
            credentialsRef: { name: string; namespace?: string };
          };
          output?: {
            keyK3sAgentJoinToken: string;
            keyK3sServerJoinToken: string;
            keyKubeconfig: string;
            secretName: string;
          };
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
    accountName: string;
    apiVersion?: string;
    creationTime: any;
    displayName: string;
    id: string;
    kind?: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec: {
      accountId: string;
      accountName: string;
      availabilityMode: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode;
      backupToS3Enabled: boolean;
      cloudflareEnabled?: boolean;
      cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
      clusterInternalDnsHost?: string;
      kloudliteRelease: string;
      messageQueueTopicName: string;
      publicDNSHost: string;
      taintMasterNodes: boolean;
      aws?: {
        nodePools?: any;
        region: string;
        spotNodePools?: any;
        k3sMasters?: {
          iamInstanceProfileRole?: string;
          instanceType: string;
          nodes?: any;
          nvidiaGpuEnabled: boolean;
          rootVolumeSize: number;
          rootVolumeType: string;
        };
      };
      clusterTokenRef?: { key: string; name: string; namespace?: string };
      output?: {
        keyK3sAgentJoinToken: string;
        keyK3sServerJoinToken: string;
        keyKubeconfig: string;
        secretName: string;
      };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        description?: string;
        debug?: boolean;
        name: string;
        title: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      syncScheduledAt?: any;
    };
  };
};

export type ConsoleGetKubeConfigQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetKubeConfigQuery = {
  infra_getCluster?: { adminKubeconfig?: { encoding: string; value: string } };
};

export type ConsoleUpdateClusterMutationVariables = Exact<{
  cluster: ClusterIn;
}>;

export type ConsoleUpdateClusterMutation = {
  infra_updateCluster?: { id: string };
};

export type ConsoleCheckAwsAccessQueryVariables = Exact<{
  cloudproviderName: Scalars['String']['input'];
}>;

export type ConsoleCheckAwsAccessQuery = {
  infra_checkAwsAccess: { result: boolean; installationUrl?: string };
};

export type ConsoleListProviderSecretsQueryVariables = Exact<{
  search?: InputMaybe<SearchProviderSecret>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListProviderSecretsQuery = {
  infra_listProviderSecrets?: {
    totalCount: number;
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPreviousPage?: boolean;
      startCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        cloudProviderName: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
        creationTime: any;
        displayName: string;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        aws?: {
          authMechanism: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsAuthMechanism;
        };
        gcp?: { serviceAccountJSON: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata: {
          namespace?: string;
          name: string;
          labels?: any;
          annotations?: any;
        };
      };
    }>;
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
    cloudProviderName: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
    creationTime: any;
    displayName: string;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata: { namespace?: string; name: string; labels?: any };
  };
};

export type ConsoleGetNodePoolQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  poolName: Scalars['String']['input'];
}>;

export type ConsoleGetNodePoolQuery = {
  infra_getNodePool?: {
    id: string;
    clusterName: string;
    creationTime: any;
    displayName: string;
    kind?: string;
    markedForDeletion?: boolean;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec: {
      cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
      maxCount: number;
      minCount: number;
      nodeLabels?: any;
      gcp?: {
        availabilityZone: string;
        machineType: string;
        poolType: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpPoolType;
      };
      aws?: {
        availabilityZone: string;
        iamInstanceProfileRole?: string;
        nvidiaGpuEnabled: boolean;
        poolType: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType;
        rootVolumeSize: number;
        rootVolumeType: string;
        ec2Pool?: { instanceType: string; nodes?: any };
        spotPool?: {
          nodes?: any;
          spotFleetTaggingRoleName: string;
          cpuNode?: {
            memoryPerVcpu?: { max: string; min: string };
            vcpu: { max: string; min: string };
          };
          gpuNode?: { instanceTypes: Array<string> };
        };
      };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
  };
};

export type ConsoleCreateNodePoolMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  pool: NodePoolIn;
}>;

export type ConsoleCreateNodePoolMutation = {
  infra_createNodePool?: { id: string };
};

export type ConsoleUpdateNodePoolMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  pool: NodePoolIn;
}>;

export type ConsoleUpdateNodePoolMutation = {
  infra_updateNodePool?: { id: string };
};

export type ConsoleListNodePoolsQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  search?: InputMaybe<SearchNodepool>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListNodePoolsQuery = {
  infra_listNodePools?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        id: string;
        clusterName: string;
        creationTime: any;
        displayName: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: { generation: number; name: string; namespace?: string };
        spec: {
          cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
          maxCount: number;
          minCount: number;
          nodeLabels?: any;
          gcp?: {
            availabilityZone: string;
            machineType: string;
            poolType: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__GcpPoolType;
          };
          aws?: {
            availabilityZone: string;
            nvidiaGpuEnabled: boolean;
            poolType: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType;
            ec2Pool?: { instanceType: string; nodes?: any };
            spotPool?: {
              nodes?: any;
              spotFleetTaggingRoleName: string;
              cpuNode?: {
                memoryPerVcpu?: { max: string; min: string };
                vcpu: { max: string; min: string };
              };
              gpuNode?: { instanceTypes: Array<string> };
            };
          };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleDeleteNodePoolMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  poolName: Scalars['String']['input'];
}>;

export type ConsoleDeleteNodePoolMutation = { infra_deleteNodePool: boolean };

export type ConsoleGetEnvironmentQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetEnvironmentQuery = {
  core_getEnvironment?: {
    creationTime: any;
    displayName: string;
    clusterName: string;
    markedForDeletion?: boolean;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec?: {
      targetNamespace?: string;
      routing?: {
        mode?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentRoutingMode;
        privateIngressClass?: string;
        publicIngressClass?: string;
      };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        description?: string;
        debug?: boolean;
        name: string;
        title: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
  };
};

export type ConsoleCreateEnvironmentMutationVariables = Exact<{
  env: EnvironmentIn;
}>;

export type ConsoleCreateEnvironmentMutation = {
  core_createEnvironment?: { id: string };
};

export type ConsoleUpdateEnvironmentMutationVariables = Exact<{
  env: EnvironmentIn;
}>;

export type ConsoleUpdateEnvironmentMutation = {
  core_updateEnvironment?: { id: string };
};

export type ConsoleDeleteEnvironmentMutationVariables = Exact<{
  envName: Scalars['String']['input'];
}>;

export type ConsoleDeleteEnvironmentMutation = {
  core_deleteEnvironment: boolean;
};

export type ConsoleListEnvironmentsQueryVariables = Exact<{
  search?: InputMaybe<SearchEnvironments>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListEnvironmentsQuery = {
  core_listEnvironments?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        clusterName: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: { generation: number; name: string; namespace?: string };
        spec?: {
          targetNamespace?: string;
          routing?: {
            mode?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentRoutingMode;
            privateIngressClass?: string;
            publicIngressClass?: string;
          };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleCloneEnvironmentMutationVariables = Exact<{
  sourceEnvName: Scalars['String']['input'];
  destinationEnvName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  environmentRoutingMode: Github__Com___Kloudlite___Operator___Apis___Crds___V1__EnvironmentRoutingMode;
}>;

export type ConsoleCloneEnvironmentMutation = {
  core_cloneEnvironment?: { id: string };
};

export type ConsoleRestartAppQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  appName: Scalars['String']['input'];
}>;

export type ConsoleRestartAppQuery = { core_restartApp: boolean };

export type ConsoleCreateAppMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  app: AppIn;
}>;

export type ConsoleCreateAppMutation = { core_createApp?: { id: string } };

export type ConsoleUpdateAppMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  app: AppIn;
}>;

export type ConsoleUpdateAppMutation = { core_updateApp?: { id: string } };

export type ConsoleInterceptAppMutationVariables = Exact<{
  portMappings?: InputMaybe<
    | Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn>
    | Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn
  >;
  intercept: Scalars['Boolean']['input'];
  deviceName: Scalars['String']['input'];
  appname: Scalars['String']['input'];
  envName: Scalars['String']['input'];
}>;

export type ConsoleInterceptAppMutation = { core_interceptApp: boolean };

export type ConsoleDeleteAppMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  appName: Scalars['String']['input'];
}>;

export type ConsoleDeleteAppMutation = { core_deleteApp: boolean };

export type ConsoleGetAppQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetAppQuery = {
  core_getApp?: {
    id: string;
    recordVersion: number;
    creationTime: any;
    displayName: string;
    enabled?: boolean;
    environmentName: string;
    markedForDeletion?: boolean;
    ciBuildId?: string;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: { annotations?: any; name: string; namespace?: string };
    spec: {
      displayName?: string;
      freeze?: boolean;
      nodeSelector?: any;
      region?: string;
      replicas?: number;
      serviceAccount?: string;
      router?: { domains: Array<string> };
      containers: Array<{
        args?: Array<string>;
        command?: Array<string>;
        image: string;
        imagePullPolicy?: string;
        name: string;
        env?: Array<{
          key: string;
          optional?: boolean;
          refKey?: string;
          refName?: string;
          type?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
          value?: string;
        }>;
        envFrom?: Array<{
          refName: string;
          type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
        }>;
        livenessProbe?: {
          failureThreshold?: number;
          initialDelay?: number;
          interval?: number;
          type: string;
          httpGet?: { httpHeaders?: any; path: string; port: number };
          shell?: { command?: Array<string> };
          tcp?: { port: number };
        };
        readinessProbe?: {
          failureThreshold?: number;
          initialDelay?: number;
          interval?: number;
          type: string;
        };
        resourceCpu?: { max?: string; min?: string };
        resourceMemory?: { max?: string; min?: string };
        volumes?: Array<{
          mountPath: string;
          refName: string;
          type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
          items?: Array<{ fileName?: string; key: string }>;
        }>;
      }>;
      hpa?: {
        enabled: boolean;
        maxReplicas?: number;
        minReplicas?: number;
        thresholdCpu?: number;
        thresholdMemory?: number;
      };
      intercept?: {
        enabled: boolean;
        toDevice: string;
        portMappings?: Array<{ devicePort: number; appPort: number }>;
      };
      services?: Array<{ port: number }>;
      tolerations?: Array<{
        effect?: K8s__Io___Api___Core___V1__TaintEffect;
        key?: string;
        operator?: K8s__Io___Api___Core___V1__TolerationOperator;
        tolerationSeconds?: number;
        value?: string;
      }>;
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        description?: string;
        debug?: boolean;
        title: string;
        name: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
    build?: {
      id: string;
      buildClusterName: string;
      name: string;
      source: {
        branch: string;
        provider: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitProvider;
        repository: string;
      };
      spec: {
        buildOptions?: {
          buildArgs?: any;
          buildContexts?: any;
          contextDir?: string;
          dockerfileContent?: string;
          dockerfilePath?: string;
          targetPlatforms?: Array<string>;
        };
        registry: { repo: { name: string; tags: Array<string> } };
        resource: { cpu: number; memoryInMb: number };
      };
    };
  };
};

export type ConsoleListAppsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchApps>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListAppsQuery = {
  core_listApps?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        enabled?: boolean;
        environmentName: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: { generation: number; name: string; namespace?: string };
        spec: {
          displayName?: string;
          freeze?: boolean;
          nodeSelector?: any;
          region?: string;
          replicas?: number;
          serviceAccount?: string;
          router?: { domains: Array<string> };
          containers: Array<{
            args?: Array<string>;
            command?: Array<string>;
            image: string;
            imagePullPolicy?: string;
            name: string;
            env?: Array<{
              key: string;
              optional?: boolean;
              refKey?: string;
              refName?: string;
              type?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
              value?: string;
            }>;
            envFrom?: Array<{
              refName: string;
              type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
            }>;
            readinessProbe?: {
              failureThreshold?: number;
              initialDelay?: number;
              interval?: number;
              type: string;
            };
            resourceCpu?: { max?: string; min?: string };
            resourceMemory?: { max?: string; min?: string };
          }>;
          hpa?: {
            enabled: boolean;
            maxReplicas?: number;
            minReplicas?: number;
            thresholdCpu?: number;
            thresholdMemory?: number;
          };
          intercept?: {
            enabled: boolean;
            toDevice: string;
            portMappings?: Array<{ devicePort: number; appPort: number }>;
          };
          services?: Array<{ port: number }>;
          tolerations?: Array<{
            effect?: K8s__Io___Api___Core___V1__TaintEffect;
            key?: string;
            operator?: K8s__Io___Api___Core___V1__TolerationOperator;
            tolerationSeconds?: number;
            value?: string;
          }>;
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            title: string;
            name: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleCreateExternalAppMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  externalApp: ExternalAppIn;
}>;

export type ConsoleCreateExternalAppMutation = {
  core_createExternalApp?: { id: string };
};

export type ConsoleUpdateExternalAppMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  externalApp: ExternalAppIn;
}>;

export type ConsoleUpdateExternalAppMutation = {
  core_updateExternalApp?: { id: string };
};

export type ConsoleInterceptExternalAppMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  externalAppName: Scalars['String']['input'];
  deviceName: Scalars['String']['input'];
  intercept: Scalars['Boolean']['input'];
  portMappings?: InputMaybe<
    | Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn>
    | Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn
  >;
}>;

export type ConsoleInterceptExternalAppMutation = {
  core_interceptExternalApp: boolean;
};

export type ConsoleDeleteExternalAppMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  externalAppName: Scalars['String']['input'];
}>;

export type ConsoleDeleteExternalAppMutation = {
  core_deleteExternalApp: boolean;
};

export type ConsoleGetExternalAppQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetExternalAppQuery = {
  core_getExternalApp?: {
    accountName: string;
    apiVersion?: string;
    creationTime: any;
    displayName: string;
    environmentName: string;
    id: string;
    kind?: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec?: {
      record: string;
      recordType: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ExternalAppRecordType;
      intercept?: {
        enabled: boolean;
        toDevice: string;
        portMappings?: Array<{ appPort: number; devicePort: number }>;
      };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        debug?: boolean;
        description?: string;
        hide?: boolean;
        name: string;
        title: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
      syncScheduledAt?: any;
    };
  };
};

export type ConsoleListExternalAppsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchExternalApps>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListExternalAppsQuery = {
  core_listExternalApps?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        apiVersion?: string;
        creationTime: any;
        displayName: string;
        environmentName: string;
        id: string;
        kind?: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec?: {
          record: string;
          recordType: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ExternalAppRecordType;
          intercept?: {
            enabled: boolean;
            toDevice: string;
            portMappings?: Array<{ appPort: number; devicePort: number }>;
          };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            debug?: boolean;
            description?: string;
            hide?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleCreateRouterMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  router: RouterIn;
}>;

export type ConsoleCreateRouterMutation = {
  core_createRouter?: { id: string };
};

export type ConsoleUpdateRouterMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  router: RouterIn;
}>;

export type ConsoleUpdateRouterMutation = {
  core_updateRouter?: { id: string };
};

export type ConsoleDeleteRouterMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  routerName: Scalars['String']['input'];
}>;

export type ConsoleDeleteRouterMutation = { core_deleteRouter: boolean };

export type ConsoleListRoutersQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchRouters>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListRoutersQuery = {
  core_listRouters?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        enabled?: boolean;
        environmentName: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: { generation: number; name: string; namespace?: string };
        spec: {
          backendProtocol?: string;
          domains: Array<string>;
          ingressClass?: string;
          maxBodySizeInMB?: number;
          basicAuth?: {
            enabled: boolean;
            secretName?: string;
            username?: string;
          };
          cors?: {
            allowCredentials?: boolean;
            enabled?: boolean;
            origins?: Array<string>;
          };
          https?: {
            clusterIssuer?: string;
            enabled: boolean;
            forceRedirect?: boolean;
          };
          rateLimit?: {
            connections?: number;
            enabled?: boolean;
            rpm?: number;
            rps?: number;
          };
          routes?: Array<{
            app: string;
            path: string;
            port: number;
            rewrite?: boolean;
          }>;
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleGetRouterQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetRouterQuery = {
  core_getRouter?: {
    creationTime: any;
    displayName: string;
    enabled?: boolean;
    environmentName: string;
    markedForDeletion?: boolean;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: { name: string; namespace?: string };
    spec: {
      backendProtocol?: string;
      domains: Array<string>;
      ingressClass?: string;
      maxBodySizeInMB?: number;
      basicAuth?: { enabled: boolean; secretName?: string; username?: string };
      cors?: {
        allowCredentials?: boolean;
        enabled?: boolean;
        origins?: Array<string>;
      };
      https?: {
        clusterIssuer?: string;
        enabled: boolean;
        forceRedirect?: boolean;
      };
      rateLimit?: {
        connections?: number;
        enabled?: boolean;
        rpm?: number;
        rps?: number;
      };
      routes?: Array<{
        app: string;
        path: string;
        port: number;
        rewrite?: boolean;
      }>;
    };
    status?: {
      checks?: any;
      isReady: boolean;
      checkList?: Array<{
        description?: string;
        debug?: boolean;
        name: string;
        title: string;
      }>;
    };
  };
};

export type ConsoleUpdateConfigMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  config: ConfigIn;
}>;

export type ConsoleUpdateConfigMutation = {
  core_updateConfig?: { id: string };
};

export type ConsoleDeleteConfigMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  configName: Scalars['String']['input'];
}>;

export type ConsoleDeleteConfigMutation = { core_deleteConfig: boolean };

export type ConsoleGetConfigQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetConfigQuery = {
  core_getConfig?: {
    binaryData?: any;
    data?: any;
    displayName: string;
    environmentName: string;
    immutable?: boolean;
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
  };
};

export type ConsoleListConfigsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchConfigs>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListConfigsQuery = {
  core_listConfigs?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        data?: any;
        environmentName: string;
        immutable?: boolean;
        markedForDeletion?: boolean;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
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

export type ConsoleCreateConfigMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  config: ConfigIn;
}>;

export type ConsoleCreateConfigMutation = {
  core_createConfig?: { id: string };
};

export type ConsoleListSecretsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchSecrets>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListSecretsQuery = {
  core_listSecrets?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        stringData?: any;
        environmentName: string;
        isReadyOnly: boolean;
        immutable?: boolean;
        markedForDeletion?: boolean;
        type?: K8s__Io___Api___Core___V1__SecretType;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
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

export type ConsoleCreateSecretMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  secret: SecretIn;
}>;

export type ConsoleCreateSecretMutation = {
  core_createSecret?: { id: string };
};

export type ConsoleGetSecretQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetSecretQuery = {
  core_getSecret?: {
    data?: any;
    displayName: string;
    environmentName: string;
    immutable?: boolean;
    markedForDeletion?: boolean;
    stringData?: any;
    type?: K8s__Io___Api___Core___V1__SecretType;
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
  };
};

export type ConsoleUpdateSecretMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  secret: SecretIn;
}>;

export type ConsoleUpdateSecretMutation = {
  core_updateSecret?: { id: string };
};

export type ConsoleDeleteSecretMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  secretName: Scalars['String']['input'];
}>;

export type ConsoleDeleteSecretMutation = { core_deleteSecret: boolean };

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
    userRole: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
  }>;
};

export type ConsoleListMembershipsForAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleListMembershipsForAccountQuery = {
  accounts_listMembershipsForAccount?: Array<{
    role: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
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

export type ConsoleInviteMembersForAccountMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  invitations: Array<InvitationIn> | InvitationIn;
}>;

export type ConsoleInviteMembersForAccountMutation = {
  accounts_inviteMembers?: Array<{ id: string }>;
};

export type ConsoleListInvitationsForUserQueryVariables = Exact<{
  onlyPending: Scalars['Boolean']['input'];
}>;

export type ConsoleListInvitationsForUserQuery = {
  accounts_listInvitationsForUser?: Array<{
    accountName: string;
    id: string;
    updateTime: any;
    inviteToken: string;
  }>;
};

export type ConsoleAcceptInvitationMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  inviteToken: Scalars['String']['input'];
}>;

export type ConsoleAcceptInvitationMutation = {
  accounts_acceptInvitation: boolean;
};

export type ConsoleRejectInvitationMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  inviteToken: Scalars['String']['input'];
}>;

export type ConsoleRejectInvitationMutation = {
  accounts_rejectInvitation: boolean;
};

export type ConsoleUpdateAccountMembershipMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  memberId: Scalars['ID']['input'];
  role: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
}>;

export type ConsoleUpdateAccountMembershipMutation = {
  accounts_updateAccountMembership: boolean;
};

export type ConsoleDeleteAccountMembershipMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  memberId: Scalars['ID']['input'];
}>;

export type ConsoleDeleteAccountMembershipMutation = {
  accounts_removeAccountMembership: boolean;
};

export type ConsoleGetCredTokenQueryVariables = Exact<{
  username: Scalars['String']['input'];
}>;

export type ConsoleGetCredTokenQuery = { cr_getCredToken: string };

export type ConsoleListCredQueryVariables = Exact<{
  search?: InputMaybe<SearchCreds>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListCredQuery = {
  cr_listCreds?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        access: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__RepoAccess;
        accountName: string;
        creationTime: any;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        recordVersion: number;
        updateTime: any;
        username: string;
        createdBy: { userEmail: string; userId: string; userName: string };
        expiration: {
          unit: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__ExpirationUnit;
          value: number;
        };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
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

export type ConsoleCreateCredMutationVariables = Exact<{
  credential: CredentialIn;
}>;

export type ConsoleCreateCredMutation = { cr_createCred?: { id: string } };

export type ConsoleDeleteCredMutationVariables = Exact<{
  username: Scalars['String']['input'];
}>;

export type ConsoleDeleteCredMutation = { cr_deleteCred: boolean };

export type ConsoleListRepoQueryVariables = Exact<{
  search?: InputMaybe<SearchRepos>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListRepoQuery = {
  cr_listRepos?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        creationTime: any;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
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

export type ConsoleCreateRepoMutationVariables = Exact<{
  repository: RepositoryIn;
}>;

export type ConsoleCreateRepoMutation = { cr_createRepo?: { id: string } };

export type ConsoleDeleteRepoMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleDeleteRepoMutation = { cr_deleteRepo: boolean };

export type ConsoleListDigestQueryVariables = Exact<{
  repoName: Scalars['String']['input'];
  search?: InputMaybe<SearchRepos>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListDigestQuery = {
  cr_listDigests?: {
    totalCount: number;
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPreviousPage?: boolean;
      startCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        url: string;
        updateTime: any;
        tags: Array<string>;
        size: number;
        repository: string;
        digest: string;
        creationTime: any;
      };
    }>;
  };
};

export type ConsoleDeleteDigestMutationVariables = Exact<{
  repoName: Scalars['String']['input'];
  digest: Scalars['String']['input'];
}>;

export type ConsoleDeleteDigestMutation = { cr_deleteDigest: boolean };

export type ConsoleGetGitConnectionsQueryVariables = Exact<{
  state?: InputMaybe<Scalars['String']['input']>;
}>;

export type ConsoleGetGitConnectionsQuery = {
  githubLoginUrl: any;
  gitlabLoginUrl: any;
  auth_me?: {
    providerGitlab?: any;
    providerGithub?: any;
    providerGoogle?: any;
  };
};

export type ConsoleGetLoginsQueryVariables = Exact<{ [key: string]: never }>;

export type ConsoleGetLoginsQuery = {
  auth_me?: { providerGithub?: any; providerGitlab?: any };
};

export type ConsoleLoginUrlsQueryVariables = Exact<{ [key: string]: never }>;

export type ConsoleLoginUrlsQuery = {
  githubLoginUrl: any;
  gitlabLoginUrl: any;
};

export type ConsoleListGithubReposQueryVariables = Exact<{
  installationId: Scalars['Int']['input'];
  pagination?: InputMaybe<PaginationIn>;
}>;

export type ConsoleListGithubReposQuery = {
  cr_listGithubRepos?: {
    totalCount?: number;
    repositories: Array<{
      cloneUrl?: string;
      defaultBranch?: string;
      fullName?: string;
      private?: boolean;
      updatedAt?: any;
    }>;
  };
};

export type ConsoleListGithubInstalltionsQueryVariables = Exact<{
  pagination?: InputMaybe<PaginationIn>;
}>;

export type ConsoleListGithubInstalltionsQuery = {
  cr_listGithubInstallations?: Array<{
    appId?: number;
    id?: number;
    nodeId?: string;
    repositoriesUrl?: string;
    targetId?: number;
    targetType?: string;
    account?: {
      avatarUrl?: string;
      id?: number;
      login?: string;
      nodeId?: string;
      type?: string;
    };
  }>;
};

export type ConsoleListGithubBranchesQueryVariables = Exact<{
  repoUrl: Scalars['String']['input'];
  pagination?: InputMaybe<PaginationIn>;
}>;

export type ConsoleListGithubBranchesQuery = {
  cr_listGithubBranches?: Array<{ name?: string }>;
};

export type ConsoleSearchGithubReposQueryVariables = Exact<{
  organization: Scalars['String']['input'];
  search: Scalars['String']['input'];
  pagination?: InputMaybe<PaginationIn>;
}>;

export type ConsoleSearchGithubReposQuery = {
  cr_searchGithubRepos?: {
    repositories: Array<{
      cloneUrl?: string;
      defaultBranch?: string;
      fullName?: string;
      private?: boolean;
      updatedAt?: any;
    }>;
  };
};

export type ConsoleListGitlabGroupsQueryVariables = Exact<{
  query?: InputMaybe<Scalars['String']['input']>;
  pagination?: InputMaybe<PaginationIn>;
}>;

export type ConsoleListGitlabGroupsQuery = {
  cr_listGitlabGroups?: Array<{ fullName: string; id: string }>;
};

export type ConsoleListGitlabReposQueryVariables = Exact<{
  query?: InputMaybe<Scalars['String']['input']>;
  pagination?: InputMaybe<PaginationIn>;
  groupId: Scalars['String']['input'];
}>;

export type ConsoleListGitlabReposQuery = {
  cr_listGitlabRepositories?: Array<{
    createdAt?: any;
    name: string;
    id: number;
    public: boolean;
    httpUrlToRepo: string;
  }>;
};

export type ConsoleListGitlabBranchesQueryVariables = Exact<{
  repoId: Scalars['String']['input'];
  query?: InputMaybe<Scalars['String']['input']>;
  pagination?: InputMaybe<PaginationIn>;
}>;

export type ConsoleListGitlabBranchesQuery = {
  cr_listGitlabBranches?: Array<{ name?: string; protected?: boolean }>;
};

export type ConsoleGetDomainQueryVariables = Exact<{
  domainName: Scalars['String']['input'];
}>;

export type ConsoleGetDomainQuery = {
  infra_getDomainEntry?: {
    updateTime: any;
    id: string;
    domainName: string;
    displayName: string;
    creationTime: any;
    clusterName: string;
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    createdBy: { userEmail: string; userId: string; userName: string };
  };
};

export type ConsoleCreateDomainMutationVariables = Exact<{
  domainEntry: DomainEntryIn;
}>;

export type ConsoleCreateDomainMutation = {
  infra_createDomainEntry?: { id: string };
};

export type ConsoleUpdateDomainMutationVariables = Exact<{
  domainEntry: DomainEntryIn;
}>;

export type ConsoleUpdateDomainMutation = {
  infra_updateDomainEntry?: { id: string };
};

export type ConsoleDeleteDomainMutationVariables = Exact<{
  domainName: Scalars['String']['input'];
}>;

export type ConsoleDeleteDomainMutation = { infra_deleteDomainEntry: boolean };

export type ConsoleListDomainsQueryVariables = Exact<{
  search?: InputMaybe<SearchDomainEntry>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListDomainsQuery = {
  infra_listDomainEntries?: {
    totalCount: number;
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPreviousPage?: boolean;
      startCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        updateTime: any;
        id: string;
        domainName: string;
        displayName: string;
        creationTime: any;
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        createdBy: { userEmail: string; userId: string; userName: string };
      };
    }>;
  };
};

export type ConsoleListBuildsQueryVariables = Exact<{
  repoName: Scalars['String']['input'];
  search?: InputMaybe<SearchBuilds>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListBuildsQuery = {
  cr_listBuilds?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        buildClusterName: string;
        errorMessages: any;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        status: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__BuildStatus;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        credUser: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        source: {
          branch: string;
          provider: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitProvider;
          repository: string;
          webhookId?: number;
        };
        spec: {
          buildOptions?: {
            buildArgs?: any;
            buildContexts?: any;
            contextDir?: string;
            dockerfileContent?: string;
            dockerfilePath?: string;
            targetPlatforms?: Array<string>;
          };
          registry: { repo: { name: string; tags: Array<string> } };
          resource: { cpu: number; memoryInMb: number };
          caches?: Array<{ name: string; path: string }>;
        };
        latestBuildRun?: {
          recordVersion: number;
          markedForDeletion?: boolean;
          status?: {
            checks?: any;
            isReady: boolean;
            lastReadyGeneration?: number;
            lastReconcileTime?: any;
            checkList?: Array<{
              debug?: boolean;
              description?: string;
              name: string;
              title: string;
            }>;
            message?: { RawMessage?: any };
            resources?: Array<{
              apiVersion: string;
              kind: string;
              name: string;
              namespace: string;
            }>;
          };
          syncStatus: {
            action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
            error?: string;
            lastSyncedAt?: any;
            recordVersion: number;
            state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
            syncScheduledAt?: any;
          };
        };
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

export type ConsoleCreateBuildMutationVariables = Exact<{
  build: BuildIn;
}>;

export type ConsoleCreateBuildMutation = { cr_addBuild?: { id: string } };

export type ConsoleUpdateBuildMutationVariables = Exact<{
  crUpdateBuildId: Scalars['ID']['input'];
  build: BuildIn;
}>;

export type ConsoleUpdateBuildMutation = { cr_updateBuild?: { id: string } };

export type ConsoleDeleteBuildMutationVariables = Exact<{
  crDeleteBuildId: Scalars['ID']['input'];
}>;

export type ConsoleDeleteBuildMutation = { cr_deleteBuild: boolean };

export type ConsoleTriggerBuildMutationVariables = Exact<{
  crTriggerBuildId: Scalars['ID']['input'];
}>;

export type ConsoleTriggerBuildMutation = { cr_triggerBuild: boolean };

export type ConsoleGetPvcQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetPvcQuery = {
  infra_getPVC?: {
    clusterName: string;
    creationTime: any;
    markedForDeletion?: boolean;
    updateTime: any;
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec?: {
      accessModes?: Array<string>;
      storageClassName?: string;
      volumeMode?: string;
      volumeName?: string;
      dataSource?: { apiGroup?: string; kind: string; name: string };
      dataSourceRef?: {
        apiGroup?: string;
        kind: string;
        name: string;
        namespace?: string;
      };
      resources?: {
        limits?: any;
        requests?: any;
        claims?: Array<{ name: string }>;
      };
      selector?: {
        matchLabels?: any;
        matchExpressions?: Array<{
          key: string;
          operator: K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator;
          values?: Array<string>;
        }>;
      };
    };
    status?: {
      accessModes?: Array<string>;
      allocatedResources?: any;
      capacity?: any;
      phase?: K8s__Io___Api___Core___V1__PersistentVolumeClaimPhase;
      conditions?: Array<{
        lastProbeTime?: any;
        lastTransitionTime?: any;
        message?: string;
        reason?: string;
        status: K8s__Io___Api___Core___V1__ConditionStatus;
        type: K8s__Io___Api___Core___V1__PersistentVolumeClaimConditionType;
      }>;
    };
  };
};

export type ConsoleListPvcsQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  search?: InputMaybe<SearchPersistentVolumeClaims>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListPvcsQuery = {
  infra_listPVCs?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        id: string;
        markedForDeletion?: boolean;
        updateTime: any;
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec?: {
          accessModes?: Array<string>;
          storageClassName?: string;
          volumeMode?: string;
          volumeName?: string;
          dataSource?: { apiGroup?: string; kind: string; name: string };
          dataSourceRef?: {
            apiGroup?: string;
            kind: string;
            name: string;
            namespace?: string;
          };
          resources?: {
            limits?: any;
            requests?: any;
            claims?: Array<{ name: string }>;
          };
          selector?: {
            matchLabels?: any;
            matchExpressions?: Array<{
              key: string;
              operator: K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator;
              values?: Array<string>;
            }>;
          };
        };
        status?: {
          accessModes?: Array<string>;
          allocatedResources?: any;
          capacity?: any;
          phase?: K8s__Io___Api___Core___V1__PersistentVolumeClaimPhase;
          conditions?: Array<{
            lastProbeTime?: any;
            lastTransitionTime?: any;
            message?: string;
            reason?: string;
            status: K8s__Io___Api___Core___V1__ConditionStatus;
            type: K8s__Io___Api___Core___V1__PersistentVolumeClaimConditionType;
          }>;
        };
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

export type ConsoleGetPvQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetPvQuery = {
  infra_getPV?: {
    clusterName: string;
    creationTime: any;
    displayName: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec?: {
      accessModes?: Array<string>;
      capacity?: any;
      mountOptions?: Array<string>;
      persistentVolumeReclaimPolicy?: K8s__Io___Api___Core___V1__PersistentVolumeReclaimPolicy;
      storageClassName?: string;
      volumeMode?: string;
      awsElasticBlockStore?: {
        fsType?: string;
        partition?: number;
        readOnly?: boolean;
        volumeID: string;
      };
      azureDisk?: {
        cachingMode?: string;
        diskName: string;
        diskURI: string;
        fsType?: string;
        kind?: string;
        readOnly?: boolean;
      };
      azureFile?: {
        readOnly?: boolean;
        secretName: string;
        secretNamespace?: string;
        shareName: string;
      };
      cephfs?: {
        monitors: Array<string>;
        path?: string;
        readOnly?: boolean;
        secretFile?: string;
        user?: string;
        secretRef?: { name?: string; namespace?: string };
      };
      cinder?: { fsType?: string; readOnly?: boolean; volumeID: string };
      claimRef?: {
        apiVersion?: string;
        fieldPath?: string;
        kind?: string;
        name?: string;
        namespace?: string;
        resourceVersion?: string;
        uid?: string;
      };
      csi?: {
        driver: string;
        fsType?: string;
        readOnly?: boolean;
        volumeAttributes?: any;
        volumeHandle: string;
        controllerExpandSecretRef?: { name?: string; namespace?: string };
        controllerPublishSecretRef?: { name?: string; namespace?: string };
        nodeExpandSecretRef?: { name?: string; namespace?: string };
        nodePublishSecretRef?: { name?: string; namespace?: string };
        nodeStageSecretRef?: { name?: string; namespace?: string };
      };
      fc?: {
        fsType?: string;
        lun?: number;
        readOnly?: boolean;
        targetWWNs?: Array<string>;
        wwids?: Array<string>;
      };
      flexVolume?: {
        driver: string;
        fsType?: string;
        options?: any;
        readOnly?: boolean;
      };
      flocker?: { datasetName?: string; datasetUUID?: string };
      gcePersistentDisk?: {
        fsType?: string;
        partition?: number;
        pdName: string;
        readOnly?: boolean;
      };
      glusterfs?: {
        endpoints: string;
        endpointsNamespace?: string;
        path: string;
        readOnly?: boolean;
      };
      hostPath?: { path: string; type?: string };
      iscsi?: {
        chapAuthDiscovery?: boolean;
        chapAuthSession?: boolean;
        fsType?: string;
        initiatorName?: string;
        iqn: string;
        iscsiInterface?: string;
        lun: number;
        portals?: Array<string>;
        readOnly?: boolean;
        targetPortal: string;
      };
      local?: { fsType?: string; path: string };
      nfs?: { path: string; readOnly?: boolean; server: string };
      nodeAffinity?: {
        required?: {
          nodeSelectorTerms: Array<{
            matchExpressions?: Array<{
              key: string;
              operator: K8s__Io___Api___Core___V1__NodeSelectorOperator;
              values?: Array<string>;
            }>;
            matchFields?: Array<{
              key: string;
              operator: K8s__Io___Api___Core___V1__NodeSelectorOperator;
              values?: Array<string>;
            }>;
          }>;
        };
      };
      photonPersistentDisk?: { fsType?: string; pdID: string };
      portworxVolume?: {
        fsType?: string;
        readOnly?: boolean;
        volumeID: string;
      };
      quobyte?: {
        group?: string;
        readOnly?: boolean;
        registry: string;
        tenant?: string;
        user?: string;
        volume: string;
      };
      rbd?: {
        fsType?: string;
        image: string;
        keyring?: string;
        monitors: Array<string>;
        pool?: string;
        readOnly?: boolean;
        user?: string;
      };
      scaleIO?: {
        fsType?: string;
        gateway: string;
        protectionDomain?: string;
        readOnly?: boolean;
        sslEnabled?: boolean;
        storageMode?: string;
        storagePool?: string;
        system: string;
        volumeName?: string;
      };
      storageos?: {
        fsType?: string;
        readOnly?: boolean;
        volumeName?: string;
        volumeNamespace?: string;
        secretRef?: {
          apiVersion?: string;
          fieldPath?: string;
          kind?: string;
          name?: string;
          namespace?: string;
          resourceVersion?: string;
          uid?: string;
        };
      };
      vsphereVolume?: {
        fsType?: string;
        storagePolicyID?: string;
        storagePolicyName?: string;
        volumePath: string;
      };
    };
    status?: {
      lastPhaseTransitionTime?: any;
      message?: string;
      phase?: K8s__Io___Api___Core___V1__PersistentVolumePhase;
      reason?: string;
    };
  };
};

export type ConsoleListPvsQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  search?: InputMaybe<SearchPersistentVolumes>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListPvsQuery = {
  infra_listPVs?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        clusterName: string;
        creationTime: any;
        displayName: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec?: {
          accessModes?: Array<string>;
          capacity?: any;
          mountOptions?: Array<string>;
          persistentVolumeReclaimPolicy?: K8s__Io___Api___Core___V1__PersistentVolumeReclaimPolicy;
          storageClassName?: string;
          volumeMode?: string;
          awsElasticBlockStore?: {
            fsType?: string;
            partition?: number;
            readOnly?: boolean;
            volumeID: string;
          };
          azureDisk?: {
            cachingMode?: string;
            diskName: string;
            diskURI: string;
            fsType?: string;
            kind?: string;
            readOnly?: boolean;
          };
          azureFile?: {
            readOnly?: boolean;
            secretName: string;
            secretNamespace?: string;
            shareName: string;
          };
          cephfs?: {
            monitors: Array<string>;
            path?: string;
            readOnly?: boolean;
            secretFile?: string;
            user?: string;
            secretRef?: { name?: string; namespace?: string };
          };
          cinder?: { fsType?: string; readOnly?: boolean; volumeID: string };
          claimRef?: {
            apiVersion?: string;
            fieldPath?: string;
            kind?: string;
            name?: string;
            namespace?: string;
            resourceVersion?: string;
            uid?: string;
          };
          csi?: {
            driver: string;
            fsType?: string;
            readOnly?: boolean;
            volumeAttributes?: any;
            volumeHandle: string;
            controllerExpandSecretRef?: { name?: string; namespace?: string };
            controllerPublishSecretRef?: { name?: string; namespace?: string };
            nodeExpandSecretRef?: { name?: string; namespace?: string };
            nodePublishSecretRef?: { name?: string; namespace?: string };
            nodeStageSecretRef?: { name?: string; namespace?: string };
          };
          fc?: {
            fsType?: string;
            lun?: number;
            readOnly?: boolean;
            targetWWNs?: Array<string>;
            wwids?: Array<string>;
          };
          flexVolume?: {
            driver: string;
            fsType?: string;
            options?: any;
            readOnly?: boolean;
          };
          flocker?: { datasetName?: string; datasetUUID?: string };
          gcePersistentDisk?: {
            fsType?: string;
            partition?: number;
            pdName: string;
            readOnly?: boolean;
          };
          glusterfs?: {
            endpoints: string;
            endpointsNamespace?: string;
            path: string;
            readOnly?: boolean;
          };
          hostPath?: { path: string; type?: string };
          iscsi?: {
            chapAuthDiscovery?: boolean;
            chapAuthSession?: boolean;
            fsType?: string;
            initiatorName?: string;
            iqn: string;
            iscsiInterface?: string;
            lun: number;
            portals?: Array<string>;
            readOnly?: boolean;
            targetPortal: string;
          };
          local?: { fsType?: string; path: string };
          nfs?: { path: string; readOnly?: boolean; server: string };
          photonPersistentDisk?: { fsType?: string; pdID: string };
          portworxVolume?: {
            fsType?: string;
            readOnly?: boolean;
            volumeID: string;
          };
          quobyte?: {
            group?: string;
            readOnly?: boolean;
            registry: string;
            tenant?: string;
            user?: string;
            volume: string;
          };
          rbd?: {
            fsType?: string;
            image: string;
            keyring?: string;
            monitors: Array<string>;
            pool?: string;
            readOnly?: boolean;
            user?: string;
          };
          scaleIO?: {
            fsType?: string;
            gateway: string;
            protectionDomain?: string;
            readOnly?: boolean;
            sslEnabled?: boolean;
            storageMode?: string;
            storagePool?: string;
            system: string;
            volumeName?: string;
          };
          storageos?: {
            fsType?: string;
            readOnly?: boolean;
            volumeName?: string;
            volumeNamespace?: string;
            secretRef?: {
              apiVersion?: string;
              fieldPath?: string;
              kind?: string;
              name?: string;
              namespace?: string;
              resourceVersion?: string;
              uid?: string;
            };
          };
          vsphereVolume?: {
            fsType?: string;
            storagePolicyID?: string;
            storagePolicyName?: string;
            volumePath: string;
          };
        };
        status?: {
          lastPhaseTransitionTime?: any;
          message?: string;
          phase?: K8s__Io___Api___Core___V1__PersistentVolumePhase;
          reason?: string;
        };
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

export type ConsoleDeletePvMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  pvName: Scalars['String']['input'];
}>;

export type ConsoleDeletePvMutation = { infra_deletePV: boolean };

export type ConsoleListBuildRunsQueryVariables = Exact<{
  search?: InputMaybe<SearchBuildRuns>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListBuildRunsQuery = {
  cr_listBuildRuns?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        id: string;
        clusterName: string;
        creationTime: any;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec?: {
          accountName: string;
          buildOptions?: {
            buildArgs?: any;
            buildContexts?: any;
            contextDir?: string;
            dockerfileContent?: string;
            dockerfilePath?: string;
            targetPlatforms?: Array<string>;
          };
          registry: { repo: { name: string; tags: Array<string> } };
          resource: { cpu: number; memoryInMb: number };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleGetBuildRunQueryVariables = Exact<{
  buildId: Scalars['ID']['input'];
  buildRunName: Scalars['String']['input'];
}>;

export type ConsoleGetBuildRunQuery = {
  cr_getBuildRun?: {
    clusterName: string;
    creationTime: any;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec?: {
      accountName: string;
      buildOptions?: {
        buildArgs?: any;
        buildContexts?: any;
        contextDir?: string;
        dockerfileContent?: string;
        dockerfilePath?: string;
        targetPlatforms?: Array<string>;
      };
      registry: { repo: { name: string; tags: Array<string> } };
      resource: { cpu: number; memoryInMb: number };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        description?: string;
        debug?: boolean;
        name: string;
        title: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
      syncScheduledAt?: any;
    };
  };
};

export type ConsoleGetClusterMSvQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetClusterMSvQuery = {
  infra_getClusterManagedService?: {
    clusterName: string;
    creationTime: any;
    displayName: string;
    id: string;
    kind?: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec?: {
      targetNamespace: string;
      msvcSpec: {
        nodeSelector?: any;
        serviceTemplate: { apiVersion: string; kind: string; spec?: any };
        tolerations?: Array<{
          effect?: K8s__Io___Api___Core___V1__TaintEffect;
          key?: string;
          operator?: K8s__Io___Api___Core___V1__TolerationOperator;
          tolerationSeconds?: number;
          value?: string;
        }>;
      };
    };
  };
};

export type ConsoleCreateClusterMSvMutationVariables = Exact<{
  service: ClusterManagedServiceIn;
}>;

export type ConsoleCreateClusterMSvMutation = {
  infra_createClusterManagedService?: { id: string };
};

export type ConsoleUpdateClusterMSvMutationVariables = Exact<{
  service: ClusterManagedServiceIn;
}>;

export type ConsoleUpdateClusterMSvMutation = {
  infra_updateClusterManagedService?: { id: string };
};

export type ConsoleListClusterMSvsQueryVariables = Exact<{
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchClusterManagedService>;
}>;

export type ConsoleListClusterMSvsQuery = {
  infra_listClusterManagedServices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        apiVersion?: string;
        clusterName: string;
        creationTime: any;
        displayName: string;
        id: string;
        kind?: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec?: {
          targetNamespace: string;
          msvcSpec: {
            nodeSelector?: any;
            serviceTemplate: { apiVersion: string; kind: string; spec?: any };
            tolerations?: Array<{
              effect?: K8s__Io___Api___Core___V1__TaintEffect;
              key?: string;
              operator?: K8s__Io___Api___Core___V1__TolerationOperator;
              tolerationSeconds?: number;
              value?: string;
            }>;
          };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            debug?: boolean;
            description?: string;
            hide?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleDeleteClusterMSvMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleDeleteClusterMSvMutation = {
  infra_deleteClusterManagedService: boolean;
};

export type ConsoleDeleteByokClusterMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleDeleteByokClusterMutation = {
  infra_deleteBYOKCluster: boolean;
};

export type ConsoleCreateByokClusterMutationVariables = Exact<{
  cluster: ByokClusterIn;
}>;

export type ConsoleCreateByokClusterMutation = {
  infra_createBYOKCluster?: { id: string };
};

export type ConsoleUpdateByokClusterMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
}>;

export type ConsoleUpdateByokClusterMutation = {
  infra_updateBYOKCluster?: { id: string };
};

export type ConsoleGetByokClusterInstructionsQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetByokClusterInstructionsQuery = {
  infrat_getBYOKClusterSetupInstructions?: string;
};

export type ConsoleGetByokClusterQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetByokClusterQuery = {
  infra_getBYOKCluster?: {
    accountName: string;
    creationTime: any;
    displayName: string;
    id: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    clusterPublicEndpoint: string;
    clusterSvcCIDR: string;
    globalVPN: string;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    syncStatus: {
      action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
      syncScheduledAt?: any;
    };
  };
};

export type ConsoleListByokClustersQueryVariables = Exact<{
  search?: InputMaybe<SearchCluster>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListByokClustersQuery = {
  infra_listBYOKClusters?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        clusterPublicEndpoint: string;
        clusterSvcCIDR: string;
        creationTime: any;
        displayName: string;
        globalVPN: string;
        id: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleGetMSvTemplateQueryVariables = Exact<{
  category: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetMSvTemplateQuery = {
  infra_getManagedServiceTemplate?: {
    active: boolean;
    apiVersion?: string;
    description: string;
    displayName: string;
    kind?: string;
    logoUrl: string;
    name: string;
    fields: Array<{
      defaultValue?: any;
      inputType: string;
      label: string;
      max?: number;
      min?: number;
      name: string;
      required?: boolean;
      unit?: string;
      displayUnit?: string;
      multiplier?: number;
    }>;
    outputs: Array<{ description: string; label: string; name: string }>;
    resources: Array<{
      apiVersion?: string;
      description: string;
      displayName: string;
      kind?: string;
      name: string;
      fields: Array<{
        defaultValue?: any;
        displayUnit?: string;
        inputType: string;
        label: string;
        max?: number;
        min?: number;
        multiplier?: number;
        name: string;
        required?: boolean;
        unit?: string;
      }>;
    }>;
  };
};

export type ConsoleListMSvTemplatesQueryVariables = Exact<{
  [key: string]: never;
}>;

export type ConsoleListMSvTemplatesQuery = {
  infra_listManagedServiceTemplates?: Array<{
    category: string;
    displayName: string;
    items: Array<{
      active: boolean;
      apiVersion?: string;
      description: string;
      displayName: string;
      kind?: string;
      logoUrl: string;
      name: string;
      fields: Array<{
        defaultValue?: any;
        inputType: string;
        label: string;
        max?: number;
        min?: number;
        name: string;
        required?: boolean;
        unit?: string;
        displayUnit?: string;
        multiplier?: number;
      }>;
      outputs: Array<{ description: string; label: string; name: string }>;
      resources: Array<{
        apiVersion?: string;
        description: string;
        displayName: string;
        kind?: string;
        name: string;
        fields: Array<{
          defaultValue?: any;
          displayUnit?: string;
          inputType: string;
          label: string;
          max?: number;
          min?: number;
          multiplier?: number;
          name: string;
          required?: boolean;
          unit?: string;
        }>;
      }>;
    }>;
  }>;
};

export type ConsoleGetManagedResourceQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetManagedResourceQuery = {
  core_getManagedResource?: {
    displayName: string;
    enabled?: boolean;
    environmentName: string;
    markedForDeletion?: boolean;
    updateTime: any;
    metadata?: { name: string; namespace?: string };
    spec: {
      resourceTemplate: {
        apiVersion: string;
        kind: string;
        spec: any;
        msvcRef: {
          apiVersion?: string;
          kind?: string;
          name: string;
          namespace: string;
        };
      };
    };
  };
};

export type ConsoleCreateManagedResourceMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  mres: ManagedResourceIn;
}>;

export type ConsoleCreateManagedResourceMutation = {
  core_createManagedResource?: { id: string };
};

export type ConsoleUpdateManagedResourceMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  mres: ManagedResourceIn;
}>;

export type ConsoleUpdateManagedResourceMutation = {
  core_updateManagedResource?: { id: string };
};

export type ConsoleListManagedResourcesQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchManagedResources>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListManagedResourcesQuery = {
  core_listManagedResources?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec: {
          resourceTemplate: {
            apiVersion: string;
            kind: string;
            spec: any;
            msvcRef: {
              apiVersion?: string;
              kind?: string;
              name: string;
              namespace: string;
              clusterName?: string;
            };
          };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncedOutputSecretRef?: {
          metadata?: { name: string; namespace?: string };
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleDeleteManagedResourceMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  mresName: Scalars['String']['input'];
}>;

export type ConsoleDeleteManagedResourceMutation = {
  core_deleteManagedResource: boolean;
};

export type ConsoleGetHelmChartQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetHelmChartQuery = {
  infra_getHelmRelease?: {
    creationTime: any;
    displayName: string;
    markedForDeletion?: boolean;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: { name: string; namespace?: string };
    spec?: {
      chartName: string;
      chartRepoURL: string;
      chartVersion: string;
      values: any;
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      releaseNotes: string;
      releaseStatus: string;
      checkList?: Array<{
        description?: string;
        debug?: boolean;
        title: string;
        name: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
  };
};

export type ConsoleListHelmChartQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  search?: InputMaybe<SearchHelmRelease>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListHelmChartQuery = {
  infra_listHelmReleases?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        clusterName: string;
        creationTime: any;
        displayName: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          generation: number;
          name: string;
          namespace?: string;
          annotations?: any;
        };
        spec?: {
          chartName: string;
          chartRepoURL: string;
          chartVersion: string;
          values: any;
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          releaseNotes: string;
          releaseStatus: string;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            title: string;
            name: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleCreateHelmChartMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  release: HelmReleaseIn;
}>;

export type ConsoleCreateHelmChartMutation = {
  infra_createHelmRelease?: { id: string };
};

export type ConsoleUpdateHelmChartMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  release: HelmReleaseIn;
}>;

export type ConsoleUpdateHelmChartMutation = {
  infra_updateHelmRelease?: { id: string };
};

export type ConsoleDeleteHelmChartMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  releaseName: Scalars['String']['input'];
}>;

export type ConsoleDeleteHelmChartMutation = {
  infra_deleteHelmRelease: boolean;
};

export type ConsoleListNamespacesQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
}>;

export type ConsoleListNamespacesQuery = {
  infra_listNamespaces?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        apiVersion?: string;
        clusterName: string;
        creationTime: any;
        displayName: string;
        id: string;
        kind?: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec?: { finalizers?: Array<string> };
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

export type ConsoleCreateConsoleVpnDeviceMutationVariables = Exact<{
  vpnDevice: ConsoleVpnDeviceIn;
}>;

export type ConsoleCreateConsoleVpnDeviceMutation = {
  core_createVPNDevice?: { id: string };
};

export type ConsoleUpdateConsoleVpnDeviceMutationVariables = Exact<{
  vpnDevice: ConsoleVpnDeviceIn;
}>;

export type ConsoleUpdateConsoleVpnDeviceMutation = {
  core_updateVPNDevice?: { id: string };
};

export type ConsoleListConsoleVpnDevicesQueryVariables = Exact<{
  search?: InputMaybe<CoreSearchVpnDevices>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListConsoleVpnDevicesQuery = {
  core_listVPNDevices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        environmentName?: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: { generation: number; name: string; namespace?: string };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            debug?: boolean;
            description?: string;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
        spec?: {
          activeNamespace?: string;
          disabled?: boolean;
          nodeSelector?: any;
          cnameRecords?: Array<{ host?: string; target?: string }>;
          ports?: Array<{ port?: number; targetPort?: number }>;
        };
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

export type ConsoleGetConsoleVpnDeviceQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleGetConsoleVpnDeviceQuery = {
  core_getVPNDevice?: {
    displayName: string;
    environmentName?: string;
    recordVersion: number;
    metadata?: { name: string; namespace?: string };
    spec?: {
      activeNamespace?: string;
      disabled?: boolean;
      nodeSelector?: any;
      cnameRecords?: Array<{ host?: string; target?: string }>;
      ports?: Array<{ port?: number; targetPort?: number }>;
    };
    wireguardConfig?: { encoding: string; value: string };
  };
};

export type ConsoleListConsoleVpnDevicesForUserQueryVariables = Exact<{
  [key: string]: never;
}>;

export type ConsoleListConsoleVpnDevicesForUserQuery = {
  core_listVPNDevicesForUser?: Array<{
    creationTime: any;
    displayName: string;
    environmentName?: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: { generation: number; name: string; namespace?: string };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        debug?: boolean;
        description?: string;
        name: string;
        title: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
      syncScheduledAt?: any;
    };
    spec?: {
      activeNamespace?: string;
      disabled?: boolean;
      nodeSelector?: any;
      cnameRecords?: Array<{ host?: string; target?: string }>;
      ports?: Array<{ port?: number; targetPort?: number }>;
    };
  }>;
};

export type ConsoleDeleteConsoleVpnDeviceMutationVariables = Exact<{
  deviceName: Scalars['String']['input'];
}>;

export type ConsoleDeleteConsoleVpnDeviceMutation = {
  core_deleteVPNDevice: boolean;
};

export type ConsoleCreateImagePullSecretMutationVariables = Exact<{
  pullSecret: ImagePullSecretIn;
}>;

export type ConsoleCreateImagePullSecretMutation = {
  core_createImagePullSecret?: { id: string };
};

export type ConsoleUpdateImagePullSecretMutationVariables = Exact<{
  pullSecret: ImagePullSecretIn;
}>;

export type ConsoleUpdateImagePullSecretMutation = {
  core_updateImagePullSecret?: { id: string };
};

export type ConsoleDeleteImagePullSecretsMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type ConsoleDeleteImagePullSecretsMutation = {
  core_deleteImagePullSecret: boolean;
};

export type ConsoleListImagePullSecretsQueryVariables = Exact<{
  search?: InputMaybe<SearchImagePullSecrets>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListImagePullSecretsQuery = {
  core_listImagePullSecrets?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        creationTime: any;
        displayName: string;
        dockerConfigJson?: string;
        format: Github__Com___Kloudlite___Api___Apps___Console___Internal___Entities__PullSecretFormat;
        id: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        registryPassword?: string;
        registryURL?: string;
        registryUsername?: string;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type ConsoleDeleteGlobalVpnDeviceMutationVariables = Exact<{
  gvpn: Scalars['String']['input'];
  deviceName: Scalars['String']['input'];
}>;

export type ConsoleDeleteGlobalVpnDeviceMutation = {
  infra_deleteGlobalVPNDevice: boolean;
};

export type ConsoleCreateGlobalVpnDeviceMutationVariables = Exact<{
  gvpnDevice: GlobalVpnDeviceIn;
}>;

export type ConsoleCreateGlobalVpnDeviceMutation = {
  infra_createGlobalVPNDevice?: { id: string };
};

export type ConsoleUpdateGlobalVpnDeviceMutationVariables = Exact<{
  gvpnDevice: GlobalVpnDeviceIn;
}>;

export type ConsoleUpdateGlobalVpnDeviceMutation = {
  infra_updateGlobalVPNDevice?: { id: string };
};

export type ConsoleGetGlobalVpnDeviceQueryVariables = Exact<{
  gvpn: Scalars['String']['input'];
  deviceName: Scalars['String']['input'];
}>;

export type ConsoleGetGlobalVpnDeviceQuery = {
  infra_getGlobalVPNDevice?: {
    accountName: string;
    creationTime: any;
    displayName: string;
    globalVPNName: string;
    id: string;
    ipAddr: string;
    markedForDeletion?: boolean;
    privateKey: string;
    publicKey: string;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    wireguardConfig?: { value: string; encoding: string };
  };
};

export type ConsoleListGlobalVpnDevicesQueryVariables = Exact<{
  gvpn: Scalars['String']['input'];
  search?: InputMaybe<SearchGlobalVpnDevices>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListGlobalVpnDevicesQuery = {
  infra_listGlobalVPNDevices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        creationMethod?: string;
        creationTime: any;
        displayName: string;
        globalVPNName: string;
        id: string;
        ipAddr: string;
        markedForDeletion?: boolean;
        privateKey: string;
        publicKey: string;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
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

export type IotconsoleAccountCheckNameAvailabilityQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type IotconsoleAccountCheckNameAvailabilityQuery = {
  accounts_checkNameAvailability: {
    result: boolean;
    suggestedNames?: Array<string>;
  };
};

export type IotconsoleCrCheckNameAvailabilityQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type IotconsoleCrCheckNameAvailabilityQuery = {
  cr_checkUserNameAvailability: {
    result: boolean;
    suggestedNames?: Array<string>;
  };
};

export type IotconsoleInfraCheckNameAvailabilityQueryVariables = Exact<{
  resType: ResType;
  name: Scalars['String']['input'];
  clusterName?: InputMaybe<Scalars['String']['input']>;
}>;

export type IotconsoleInfraCheckNameAvailabilityQuery = {
  infra_checkNameAvailability: {
    suggestedNames: Array<string>;
    result: boolean;
  };
};

export type IotconsoleCoreCheckNameAvailabilityQueryVariables = Exact<{
  resType: ConsoleResType;
  name: Scalars['String']['input'];
  projectName?: InputMaybe<Scalars['String']['input']>;
  envName?: InputMaybe<Scalars['String']['input']>;
}>;

export type IotconsoleCoreCheckNameAvailabilityQuery = {
  core_checkNameAvailability: { result: boolean };
};

export type IotconsoleWhoAmIQueryVariables = Exact<{ [key: string]: never }>;

export type IotconsoleWhoAmIQuery = {
  auth_me?: {
    id: string;
    email: string;
    providerGitlab?: any;
    providerGithub?: any;
    providerGoogle?: any;
  };
};

export type IotconsoleCreateAccountMutationVariables = Exact<{
  account: AccountIn;
}>;

export type IotconsoleCreateAccountMutation = {
  accounts_createAccount: { displayName: string };
};

export type IotconsoleListAccountsQueryVariables = Exact<{
  [key: string]: never;
}>;

export type IotconsoleListAccountsQuery = {
  accounts_listAccounts?: Array<{
    id: string;
    updateTime: any;
    displayName: string;
    metadata?: { name: string; annotations?: any };
  }>;
};

export type IotconsoleUpdateAccountMutationVariables = Exact<{
  account: AccountIn;
}>;

export type IotconsoleUpdateAccountMutation = {
  accounts_updateAccount: { id: string };
};

export type IotconsoleGetAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type IotconsoleGetAccountQuery = {
  accounts_getAccount?: {
    targetNamespace?: string;
    updateTime: any;
    contactEmail?: string;
    displayName: string;
    metadata?: { name: string; annotations?: any };
  };
};

export type IotconsoleDeleteAccountMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type IotconsoleDeleteAccountMutation = {
  accounts_deleteAccount: boolean;
};

export type IotconsoleDeleteIotProjectMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type IotconsoleDeleteIotProjectMutation = { iot_deleteProject: boolean };

export type IotconsoleCreateIotProjectMutationVariables = Exact<{
  project: IotProjectIn;
}>;

export type IotconsoleCreateIotProjectMutation = {
  iot_createProject?: { id: string };
};

export type IotconsoleUpdateIotProjectMutationVariables = Exact<{
  project: IotProjectIn;
}>;

export type IotconsoleUpdateIotProjectMutation = {
  iot_updateProject?: { id: string };
};

export type IotconsoleGetIotProjectQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type IotconsoleGetIotProjectQuery = {
  iot_getProject?: {
    accountName: string;
    creationTime: any;
    displayName: string;
    id: string;
    markedForDeletion?: boolean;
    name: string;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
  };
};

export type IotconsoleListIotProjectsQueryVariables = Exact<{
  search?: InputMaybe<SearchIotProjects>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListIotProjectsQuery = {
  iot_listProjects?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        displayName: string;
        name: string;
        creationTime: any;
        markedForDeletion?: boolean;
        updateTime: any;
        createdBy: { userEmail: string; userName: string; userId: string };
        lastUpdatedBy: { userEmail: string; userName: string; userId: string };
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

export type IotconsoleDeleteIotDeviceBlueprintMutationVariables = Exact<{
  projectName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleDeleteIotDeviceBlueprintMutation = {
  iot_deleteDeviceBlueprint: boolean;
};

export type IotconsoleCreateIotDeviceBlueprintMutationVariables = Exact<{
  deviceBlueprint: IotDeviceBlueprintIn;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleCreateIotDeviceBlueprintMutation = {
  iot_createDeviceBlueprint?: { id: string };
};

export type IotconsoleUpdateIotDeviceBlueprintMutationVariables = Exact<{
  deviceBlueprint: IotDeviceBlueprintIn;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleUpdateIotDeviceBlueprintMutation = {
  iot_updateDeviceBlueprint?: { id: string };
};

export type IotconsoleGetIotDeviceBlueprintQueryVariables = Exact<{
  projectName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleGetIotDeviceBlueprintQuery = {
  iot_getDeviceBlueprint?: {
    accountName: string;
    bluePrintType: Github__Com___Kloudlite___Api___Apps___Iot____Console___Internal___Entities__BluePrintType;
    creationTime: any;
    displayName: string;
    id: string;
    markedForDeletion?: boolean;
    name: string;
    recordVersion: number;
    updateTime: any;
    version: string;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
  };
};

export type IotconsoleListIotDeviceBlueprintsQueryVariables = Exact<{
  search?: InputMaybe<SearchIotDeviceBlueprints>;
  pq?: InputMaybe<CursorPaginationIn>;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleListIotDeviceBlueprintsQuery = {
  iot_listDeviceBlueprints?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        bluePrintType: Github__Com___Kloudlite___Api___Apps___Iot____Console___Internal___Entities__BluePrintType;
        creationTime: any;
        displayName: string;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        recordVersion: number;
        updateTime: any;
        version: string;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
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

export type IotconsoleDeleteIotDeploymentMutationVariables = Exact<{
  projectName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleDeleteIotDeploymentMutation = {
  iot_deleteDeployment: boolean;
};

export type IotconsoleCreateIotDeploymentMutationVariables = Exact<{
  projectName: Scalars['String']['input'];
  deployment: IotDeploymentIn;
}>;

export type IotconsoleCreateIotDeploymentMutation = {
  iot_createDeployment?: { id: string };
};

export type IotconsoleUpdateIotDeploymentMutationVariables = Exact<{
  projectName: Scalars['String']['input'];
  deployment: IotDeploymentIn;
}>;

export type IotconsoleUpdateIotDeploymentMutation = {
  iot_updateDeployment?: { id: string };
};

export type IotconsoleGetIotDeploymentQueryVariables = Exact<{
  projectName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleGetIotDeploymentQuery = {
  iot_getDeployment?: {
    accountName: string;
    CIDR: string;
    creationTime: any;
    displayName: string;
    id: string;
    markedForDeletion?: boolean;
    name: string;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    exposedServices: Array<{ ip: string; name: string }>;
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
  };
};

export type IotconsoleListIotDeploymentsQueryVariables = Exact<{
  search?: InputMaybe<SearchIotDeployments>;
  pq?: InputMaybe<CursorPaginationIn>;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleListIotDeploymentsQuery = {
  iot_listDeployments?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        CIDR: string;
        creationTime: any;
        displayName: string;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        exposedServices: Array<{ ip: string; name: string }>;
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
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

export type IotconsoleDeleteIotAppMutationVariables = Exact<{
  projectName: Scalars['String']['input'];
  deviceBlueprintName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleDeleteIotAppMutation = { iot_deleteApp: boolean };

export type IotconsoleCreateIotAppMutationVariables = Exact<{
  deviceBlueprintName: Scalars['String']['input'];
  app: IotAppIn;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleCreateIotAppMutation = { iot_createApp?: { id: string } };

export type IotconsoleUpdateIotAppMutationVariables = Exact<{
  projectName: Scalars['String']['input'];
  deviceBlueprintName: Scalars['String']['input'];
  app: IotAppIn;
}>;

export type IotconsoleUpdateIotAppMutation = { iot_updateApp?: { id: string } };

export type IotconsoleGetIotAppQueryVariables = Exact<{
  projectName: Scalars['String']['input'];
  deviceBlueprintName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleGetIotAppQuery = {
  iot_getApp?: {
    accountName: string;
    apiVersion?: string;
    creationTime: any;
    deviceBlueprintName: string;
    displayName: string;
    enabled?: boolean;
    id: string;
    kind?: string;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
    metadata?: {
      annotations?: any;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec: {
      displayName?: string;
      freeze?: boolean;
      nodeSelector?: any;
      region?: string;
      replicas?: number;
      serviceAccount?: string;
      containers: Array<{
        args?: Array<string>;
        command?: Array<string>;
        image: string;
        imagePullPolicy?: string;
        name: string;
        env?: Array<{
          key: string;
          optional?: boolean;
          refKey?: string;
          refName?: string;
          type?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
          value?: string;
        }>;
        envFrom?: Array<{
          refName: string;
          type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
        }>;
        livenessProbe?: {
          failureThreshold?: number;
          initialDelay?: number;
          interval?: number;
          type: string;
          httpGet?: { httpHeaders?: any; path: string; port: number };
          shell?: { command?: Array<string> };
          tcp?: { port: number };
        };
        readinessProbe?: {
          failureThreshold?: number;
          initialDelay?: number;
          interval?: number;
          type: string;
        };
        resourceCpu?: { max?: string; min?: string };
        resourceMemory?: { max?: string; min?: string };
        volumes?: Array<{
          mountPath: string;
          refName: string;
          type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
          items?: Array<{ fileName?: string; key: string }>;
        }>;
      }>;
      hpa?: {
        enabled: boolean;
        maxReplicas?: number;
        minReplicas?: number;
        thresholdCpu?: number;
        thresholdMemory?: number;
      };
      intercept?: { enabled: boolean; toDevice: string };
      services?: Array<{ port: number }>;
      tolerations?: Array<{
        effect?: K8s__Io___Api___Core___V1__TaintEffect;
        key?: string;
        operator?: K8s__Io___Api___Core___V1__TolerationOperator;
        tolerationSeconds?: number;
        value?: string;
      }>;
      topologySpreadConstraints?: Array<{
        matchLabelKeys?: Array<string>;
        maxSkew: number;
        minDomains?: number;
        nodeAffinityPolicy?: string;
        nodeTaintsPolicy?: string;
        topologyKey: string;
        whenUnsatisfiable: K8s__Io___Api___Core___V1__UnsatisfiableConstraintAction;
        labelSelector?: {
          matchLabels?: any;
          matchExpressions?: Array<{
            key: string;
            operator: K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator;
            values?: Array<string>;
          }>;
        };
      }>;
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        debug?: boolean;
        description?: string;
        name: string;
        title: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
  };
};

export type IotconsoleListIotAppsQueryVariables = Exact<{
  deviceBlueprintName: Scalars['String']['input'];
  search?: InputMaybe<SearchIotApps>;
  pq?: InputMaybe<CursorPaginationIn>;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleListIotAppsQuery = {
  iot_listApps?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        apiVersion?: string;
        creationTime: any;
        deviceBlueprintName: string;
        displayName: string;
        enabled?: boolean;
        id: string;
        kind?: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec: {
          displayName?: string;
          freeze?: boolean;
          nodeSelector?: any;
          region?: string;
          replicas?: number;
          serviceAccount?: string;
          containers: Array<{
            args?: Array<string>;
            command?: Array<string>;
            image: string;
            imagePullPolicy?: string;
            name: string;
            env?: Array<{
              key: string;
              optional?: boolean;
              refKey?: string;
              refName?: string;
              type?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
              value?: string;
            }>;
            envFrom?: Array<{
              refName: string;
              type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
            }>;
            livenessProbe?: {
              failureThreshold?: number;
              initialDelay?: number;
              interval?: number;
              type: string;
              httpGet?: { httpHeaders?: any; path: string; port: number };
              shell?: { command?: Array<string> };
              tcp?: { port: number };
            };
            readinessProbe?: {
              failureThreshold?: number;
              initialDelay?: number;
              interval?: number;
              type: string;
            };
            resourceCpu?: { max?: string; min?: string };
            resourceMemory?: { max?: string; min?: string };
            volumes?: Array<{
              mountPath: string;
              refName: string;
              type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
              items?: Array<{ fileName?: string; key: string }>;
            }>;
          }>;
          hpa?: {
            enabled: boolean;
            maxReplicas?: number;
            minReplicas?: number;
            thresholdCpu?: number;
            thresholdMemory?: number;
          };
          intercept?: { enabled: boolean; toDevice: string };
          services?: Array<{ port: number }>;
          tolerations?: Array<{
            effect?: K8s__Io___Api___Core___V1__TaintEffect;
            key?: string;
            operator?: K8s__Io___Api___Core___V1__TolerationOperator;
            tolerationSeconds?: number;
            value?: string;
          }>;
          topologySpreadConstraints?: Array<{
            matchLabelKeys?: Array<string>;
            maxSkew: number;
            minDomains?: number;
            nodeAffinityPolicy?: string;
            nodeTaintsPolicy?: string;
            topologyKey: string;
            whenUnsatisfiable: K8s__Io___Api___Core___V1__UnsatisfiableConstraintAction;
            labelSelector?: {
              matchLabels?: any;
              matchExpressions?: Array<{
                key: string;
                operator: K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator;
                values?: Array<string>;
              }>;
            };
          }>;
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            debug?: boolean;
            description?: string;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
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

export type IotconsoleDeleteIotDeviceMutationVariables = Exact<{
  projectName: Scalars['String']['input'];
  deploymentName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleDeleteIotDeviceMutation = { iot_deleteDevice: boolean };

export type IotconsoleCreateIotDeviceMutationVariables = Exact<{
  deploymentName: Scalars['String']['input'];
  device: IotDeviceIn;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleCreateIotDeviceMutation = {
  iot_createDevice?: { id: string };
};

export type IotconsoleUpdateIotDeviceMutationVariables = Exact<{
  deploymentName: Scalars['String']['input'];
  device: IotDeviceIn;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleUpdateIotDeviceMutation = {
  iot_updateDevice?: { id: string };
};

export type IotconsoleGetIotDeviceQueryVariables = Exact<{
  projectName: Scalars['String']['input'];
  deploymentName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleGetIotDeviceQuery = {
  iot_getDevice?: {
    accountName: string;
    creationTime: any;
    deploymentName: string;
    displayName: string;
    id: string;
    ip: string;
    markedForDeletion?: boolean;
    name: string;
    podCIDR: string;
    publicKey: string;
    recordVersion: number;
    serviceCIDR: string;
    updateTime: any;
    version: string;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userEmail: string; userId: string; userName: string };
  };
};

export type IotconsoleListIotDevicesQueryVariables = Exact<{
  deploymentName: Scalars['String']['input'];
  search?: InputMaybe<SearchIotDevices>;
  pq?: InputMaybe<CursorPaginationIn>;
  projectName: Scalars['String']['input'];
}>;

export type IotconsoleListIotDevicesQuery = {
  iot_listDevices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        creationTime: any;
        deploymentName: string;
        displayName: string;
        id: string;
        ip: string;
        markedForDeletion?: boolean;
        name: string;
        podCIDR: string;
        publicKey: string;
        recordVersion: number;
        serviceCIDR: string;
        updateTime: any;
        version: string;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
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

export type IotconsoleListRepoQueryVariables = Exact<{
  search?: InputMaybe<SearchRepos>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListRepoQuery = {
  cr_listRepos?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        creationTime: any;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        recordVersion: number;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
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

export type IotconsoleCreateRepoMutationVariables = Exact<{
  repository: RepositoryIn;
}>;

export type IotconsoleCreateRepoMutation = { cr_createRepo?: { id: string } };

export type IotconsoleDeleteRepoMutationVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type IotconsoleDeleteRepoMutation = { cr_deleteRepo: boolean };

export type IotconsoleListDigestQueryVariables = Exact<{
  repoName: Scalars['String']['input'];
  search?: InputMaybe<SearchRepos>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListDigestQuery = {
  cr_listDigests?: {
    totalCount: number;
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPreviousPage?: boolean;
      startCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        url: string;
        updateTime: any;
        tags: Array<string>;
        size: number;
        repository: string;
        digest: string;
        creationTime: any;
      };
    }>;
  };
};

export type IotconsoleDeleteDigestMutationVariables = Exact<{
  repoName: Scalars['String']['input'];
  digest: Scalars['String']['input'];
}>;

export type IotconsoleDeleteDigestMutation = { cr_deleteDigest: boolean };

export type IotconsoleUpdateConfigMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  config: ConfigIn;
}>;

export type IotconsoleUpdateConfigMutation = {
  core_updateConfig?: { id: string };
};

export type IotconsoleDeleteConfigMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  configName: Scalars['String']['input'];
}>;

export type IotconsoleDeleteConfigMutation = { core_deleteConfig: boolean };

export type IotconsoleGetConfigQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleGetConfigQuery = {
  core_getConfig?: {
    binaryData?: any;
    data?: any;
    displayName: string;
    environmentName: string;
    immutable?: boolean;
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
  };
};

export type IotconsoleListConfigsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchConfigs>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListConfigsQuery = {
  core_listConfigs?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        data?: any;
        environmentName: string;
        immutable?: boolean;
        markedForDeletion?: boolean;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
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

export type IotconsoleCreateConfigMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  config: ConfigIn;
}>;

export type IotconsoleCreateConfigMutation = {
  core_createConfig?: { id: string };
};

export type IotconsoleListSecretsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  search?: InputMaybe<SearchSecrets>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListSecretsQuery = {
  core_listSecrets?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        stringData?: any;
        environmentName: string;
        isReadyOnly: boolean;
        immutable?: boolean;
        markedForDeletion?: boolean;
        type?: K8s__Io___Api___Core___V1__SecretType;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
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

export type IotconsoleCreateSecretMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  secret: SecretIn;
}>;

export type IotconsoleCreateSecretMutation = {
  core_createSecret?: { id: string };
};

export type IotconsoleGetSecretQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type IotconsoleGetSecretQuery = {
  core_getSecret?: {
    data?: any;
    displayName: string;
    environmentName: string;
    immutable?: boolean;
    markedForDeletion?: boolean;
    stringData?: any;
    type?: K8s__Io___Api___Core___V1__SecretType;
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
  };
};

export type IotconsoleUpdateSecretMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  secret: SecretIn;
}>;

export type IotconsoleUpdateSecretMutation = {
  core_updateSecret?: { id: string };
};

export type IotconsoleDeleteSecretMutationVariables = Exact<{
  envName: Scalars['String']['input'];
  secretName: Scalars['String']['input'];
}>;

export type IotconsoleDeleteSecretMutation = { core_deleteSecret: boolean };

export type IotconsoleGetCredTokenQueryVariables = Exact<{
  username: Scalars['String']['input'];
}>;

export type IotconsoleGetCredTokenQuery = { cr_getCredToken: string };

export type IotconsoleListCredQueryVariables = Exact<{
  search?: InputMaybe<SearchCreds>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListCredQuery = {
  cr_listCreds?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        access: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__RepoAccess;
        accountName: string;
        creationTime: any;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        recordVersion: number;
        updateTime: any;
        username: string;
        createdBy: { userEmail: string; userId: string; userName: string };
        expiration: {
          unit: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__ExpirationUnit;
          value: number;
        };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
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

export type IotconsoleCreateCredMutationVariables = Exact<{
  credential: CredentialIn;
}>;

export type IotconsoleCreateCredMutation = { cr_createCred?: { id: string } };

export type IotconsoleDeleteCredMutationVariables = Exact<{
  username: Scalars['String']['input'];
}>;

export type IotconsoleDeleteCredMutation = { cr_deleteCred: boolean };

export type IotconsoleGetGitConnectionsQueryVariables = Exact<{
  state?: InputMaybe<Scalars['String']['input']>;
}>;

export type IotconsoleGetGitConnectionsQuery = {
  githubLoginUrl: any;
  gitlabLoginUrl: any;
  auth_me?: {
    providerGitlab?: any;
    providerGithub?: any;
    providerGoogle?: any;
  };
};

export type IotconsoleGetLoginsQueryVariables = Exact<{ [key: string]: never }>;

export type IotconsoleGetLoginsQuery = {
  auth_me?: { providerGithub?: any; providerGitlab?: any };
};

export type IotconsoleLoginUrlsQueryVariables = Exact<{ [key: string]: never }>;

export type IotconsoleLoginUrlsQuery = {
  githubLoginUrl: any;
  gitlabLoginUrl: any;
};

export type IotconsoleListGithubReposQueryVariables = Exact<{
  installationId: Scalars['Int']['input'];
  pagination?: InputMaybe<PaginationIn>;
}>;

export type IotconsoleListGithubReposQuery = {
  cr_listGithubRepos?: {
    totalCount?: number;
    repositories: Array<{
      cloneUrl?: string;
      defaultBranch?: string;
      fullName?: string;
      private?: boolean;
      updatedAt?: any;
    }>;
  };
};

export type IotconsoleListGithubInstalltionsQueryVariables = Exact<{
  pagination?: InputMaybe<PaginationIn>;
}>;

export type IotconsoleListGithubInstalltionsQuery = {
  cr_listGithubInstallations?: Array<{
    appId?: number;
    id?: number;
    nodeId?: string;
    repositoriesUrl?: string;
    targetId?: number;
    targetType?: string;
    account?: {
      avatarUrl?: string;
      id?: number;
      login?: string;
      nodeId?: string;
      type?: string;
    };
  }>;
};

export type IotconsoleListGithubBranchesQueryVariables = Exact<{
  repoUrl: Scalars['String']['input'];
  pagination?: InputMaybe<PaginationIn>;
}>;

export type IotconsoleListGithubBranchesQuery = {
  cr_listGithubBranches?: Array<{ name?: string }>;
};

export type IotconsoleSearchGithubReposQueryVariables = Exact<{
  organization: Scalars['String']['input'];
  search: Scalars['String']['input'];
  pagination?: InputMaybe<PaginationIn>;
}>;

export type IotconsoleSearchGithubReposQuery = {
  cr_searchGithubRepos?: {
    repositories: Array<{
      cloneUrl?: string;
      defaultBranch?: string;
      fullName?: string;
      private?: boolean;
      updatedAt?: any;
    }>;
  };
};

export type IotconsoleListGitlabGroupsQueryVariables = Exact<{
  query?: InputMaybe<Scalars['String']['input']>;
  pagination?: InputMaybe<PaginationIn>;
}>;

export type IotconsoleListGitlabGroupsQuery = {
  cr_listGitlabGroups?: Array<{ fullName: string; id: string }>;
};

export type IotconsoleListGitlabReposQueryVariables = Exact<{
  query?: InputMaybe<Scalars['String']['input']>;
  pagination?: InputMaybe<PaginationIn>;
  groupId: Scalars['String']['input'];
}>;

export type IotconsoleListGitlabReposQuery = {
  cr_listGitlabRepositories?: Array<{
    createdAt?: any;
    name: string;
    id: number;
    public: boolean;
    httpUrlToRepo: string;
  }>;
};

export type IotconsoleListGitlabBranchesQueryVariables = Exact<{
  repoId: Scalars['String']['input'];
  query?: InputMaybe<Scalars['String']['input']>;
  pagination?: InputMaybe<PaginationIn>;
}>;

export type IotconsoleListGitlabBranchesQuery = {
  cr_listGitlabBranches?: Array<{ name?: string; protected?: boolean }>;
};

export type IotconsoleListBuildsQueryVariables = Exact<{
  repoName: Scalars['String']['input'];
  search?: InputMaybe<SearchBuilds>;
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListBuildsQuery = {
  cr_listBuilds?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        buildClusterName: string;
        errorMessages: any;
        id: string;
        markedForDeletion?: boolean;
        name: string;
        status: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__BuildStatus;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        credUser: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        source: {
          branch: string;
          provider: Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitProvider;
          repository: string;
          webhookId?: number;
        };
        spec: {
          buildOptions?: {
            buildArgs?: any;
            buildContexts?: any;
            contextDir?: string;
            dockerfileContent?: string;
            dockerfilePath?: string;
            targetPlatforms?: Array<string>;
          };
          registry: { repo: { name: string; tags: Array<string> } };
          resource: { cpu: number; memoryInMb: number };
          caches?: Array<{ name: string; path: string }>;
        };
        latestBuildRun?: {
          recordVersion: number;
          markedForDeletion?: boolean;
          status?: {
            checks?: any;
            isReady: boolean;
            lastReadyGeneration?: number;
            lastReconcileTime?: any;
            checkList?: Array<{
              debug?: boolean;
              description?: string;
              name: string;
              title: string;
            }>;
            message?: { RawMessage?: any };
            resources?: Array<{
              apiVersion: string;
              kind: string;
              name: string;
              namespace: string;
            }>;
          };
          syncStatus: {
            action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
            error?: string;
            lastSyncedAt?: any;
            recordVersion: number;
            state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
            syncScheduledAt?: any;
          };
        };
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

export type IotconsoleCreateBuildMutationVariables = Exact<{
  build: BuildIn;
}>;

export type IotconsoleCreateBuildMutation = { cr_addBuild?: { id: string } };

export type IotconsoleUpdateBuildMutationVariables = Exact<{
  crUpdateBuildId: Scalars['ID']['input'];
  build: BuildIn;
}>;

export type IotconsoleUpdateBuildMutation = { cr_updateBuild?: { id: string } };

export type IotconsoleDeleteBuildMutationVariables = Exact<{
  crDeleteBuildId: Scalars['ID']['input'];
}>;

export type IotconsoleDeleteBuildMutation = { cr_deleteBuild: boolean };

export type IotconsoleTriggerBuildMutationVariables = Exact<{
  crTriggerBuildId: Scalars['ID']['input'];
}>;

export type IotconsoleTriggerBuildMutation = { cr_triggerBuild: boolean };

export type IotconsoleListBuildRunsQueryVariables = Exact<{
  search?: InputMaybe<SearchBuildRuns>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type IotconsoleListBuildRunsQuery = {
  cr_listBuildRuns?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        id: string;
        clusterName: string;
        creationTime: any;
        markedForDeletion?: boolean;
        recordVersion: number;
        updateTime: any;
        metadata?: {
          annotations?: any;
          creationTimestamp: any;
          deletionTimestamp?: any;
          generation: number;
          labels?: any;
          name: string;
          namespace?: string;
        };
        spec?: {
          accountName: string;
          buildOptions?: {
            buildArgs?: any;
            buildContexts?: any;
            contextDir?: string;
            dockerfileContent?: string;
            dockerfilePath?: string;
            targetPlatforms?: Array<string>;
          };
          registry: { repo: { name: string; tags: Array<string> } };
          resource: { cpu: number; memoryInMb: number };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          checkList?: Array<{
            description?: string;
            debug?: boolean;
            name: string;
            title: string;
          }>;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion: string;
            kind: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          syncScheduledAt?: any;
        };
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

export type IotconsoleGetBuildRunQueryVariables = Exact<{
  buildId: Scalars['ID']['input'];
  buildRunName: Scalars['String']['input'];
}>;

export type IotconsoleGetBuildRunQuery = {
  cr_getBuildRun?: {
    clusterName: string;
    creationTime: any;
    markedForDeletion?: boolean;
    recordVersion: number;
    updateTime: any;
    metadata?: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    spec?: {
      accountName: string;
      buildOptions?: {
        buildArgs?: any;
        buildContexts?: any;
        contextDir?: string;
        dockerfileContent?: string;
        dockerfilePath?: string;
        targetPlatforms?: Array<string>;
      };
      registry: { repo: { name: string; tags: Array<string> } };
      resource: { cpu: number; memoryInMb: number };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      checkList?: Array<{
        description?: string;
        debug?: boolean;
        name: string;
        title: string;
      }>;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion: string;
        kind: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
      syncScheduledAt?: any;
    };
  };
};

export type IotconsoleListInvitationsForAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type IotconsoleListInvitationsForAccountQuery = {
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
    userRole: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
  }>;
};

export type IotconsoleListMembershipsForAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type IotconsoleListMembershipsForAccountQuery = {
  accounts_listMembershipsForAccount?: Array<{
    role: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
    user: { verified: boolean; name: string; joined: any; email: string };
  }>;
};

export type IotconsoleDeleteAccountInvitationMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  invitationId: Scalars['String']['input'];
}>;

export type IotconsoleDeleteAccountInvitationMutation = {
  accounts_deleteInvitation: boolean;
};

export type IotconsoleInviteMembersForAccountMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  invitations: Array<InvitationIn> | InvitationIn;
}>;

export type IotconsoleInviteMembersForAccountMutation = {
  accounts_inviteMembers?: Array<{ id: string }>;
};

export type IotconsoleListInvitationsForUserQueryVariables = Exact<{
  onlyPending: Scalars['Boolean']['input'];
}>;

export type IotconsoleListInvitationsForUserQuery = {
  accounts_listInvitationsForUser?: Array<{
    accountName: string;
    id: string;
    updateTime: any;
    inviteToken: string;
  }>;
};

export type IotconsoleAcceptInvitationMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  inviteToken: Scalars['String']['input'];
}>;

export type IotconsoleAcceptInvitationMutation = {
  accounts_acceptInvitation: boolean;
};

export type IotconsoleRejectInvitationMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  inviteToken: Scalars['String']['input'];
}>;

export type IotconsoleRejectInvitationMutation = {
  accounts_rejectInvitation: boolean;
};

export type IotconsoleUpdateAccountMembershipMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  memberId: Scalars['ID']['input'];
  role: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
}>;

export type IotconsoleUpdateAccountMembershipMutation = {
  accounts_updateAccountMembership: boolean;
};

export type IotconsoleDeleteAccountMembershipMutationVariables = Exact<{
  accountName: Scalars['String']['input'];
  memberId: Scalars['ID']['input'];
}>;

export type IotconsoleDeleteAccountMembershipMutation = {
  accounts_removeAccountMembership: boolean;
};

export type AuthCli_UpdateDeviceClusterMutationVariables = Exact<{
  deviceName: Scalars['String']['input'];
  clusterName: Scalars['String']['input'];
}>;

export type AuthCli_UpdateDeviceClusterMutation = {
  core_updateVpnClusterName: boolean;
};

export type AuthCli_UpdateDeviceNsMutationVariables = Exact<{
  deviceName: Scalars['String']['input'];
  ns: Scalars['String']['input'];
}>;

export type AuthCli_UpdateDeviceNsMutation = {
  core_updateVpnDeviceNs: boolean;
};

export type AuthCli_UpdateDevicePortsMutationVariables = Exact<{
  deviceName: Scalars['String']['input'];
  ports: Array<PortIn> | PortIn;
}>;

export type AuthCli_UpdateDevicePortsMutation = {
  core_updateVPNDevicePorts: boolean;
};

export type AuthCli_UpdateDeviceEnvMutationVariables = Exact<{
  deviceName: Scalars['String']['input'];
  envName: Scalars['String']['input'];
}>;

export type AuthCli_UpdateDeviceEnvMutation = {
  core_updateVPNDeviceEnv: boolean;
};

export type AuthCli_ListDevicesQueryVariables = Exact<{ [key: string]: never }>;

export type AuthCli_ListDevicesQuery = {
  core_listVPNDevicesForUser?: Array<{
    displayName: string;
    environmentName?: string;
    clusterName?: string;
    metadata?: { name: string };
    status?: { isReady: boolean; message?: { RawMessage?: any } };
    spec?: {
      activeNamespace?: string;
      disabled?: boolean;
      cnameRecords?: Array<{ host?: string; target?: string }>;
      ports?: Array<{ port?: number; targetPort?: number }>;
    };
  }>;
};

export type AuthCli_GetDeviceQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type AuthCli_GetDeviceQuery = {
  core_getVPNDevice?: {
    displayName: string;
    clusterName?: string;
    environmentName?: string;
    metadata?: { name: string };
    spec?: {
      activeNamespace?: string;
      disabled?: boolean;
      ports?: Array<{ port?: number; targetPort?: number }>;
    };
    wireguardConfig?: { encoding: string; value: string };
  };
};

export type AuthCli_CreateDeviceMutationVariables = Exact<{
  vpnDevice: ConsoleVpnDeviceIn;
}>;

export type AuthCli_CreateDeviceMutation = {
  core_createVPNDevice?: {
    metadata?: { name: string };
    wireguardConfig?: { encoding: string; value: string };
  };
};

export type AuthCli_CreateGlobalVpnDeviceMutationVariables = Exact<{
  gvpnDevice: GlobalVpnDeviceIn;
}>;

export type AuthCli_CreateGlobalVpnDeviceMutation = {
  infra_createGlobalVPNDevice?: {
    accountName: string;
    creationTime: any;
    displayName: string;
    globalVPNName: string;
    id: string;
    ipAddr: string;
    markedForDeletion?: boolean;
    privateKey: string;
    publicKey: string;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userName: string; userId: string; userEmail: string };
    metadata: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    wireguardConfig?: { value: string; encoding: string };
  };
};

export type AuthCli_GetMresOutputKeyValuesQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  keyrefs?: InputMaybe<
    | Array<InputMaybe<ManagedResourceKeyRefIn>>
    | InputMaybe<ManagedResourceKeyRefIn>
  >;
}>;

export type AuthCli_GetMresOutputKeyValuesQuery = {
  core_getManagedResouceOutputKeyValues: Array<{
    key: string;
    mresName: string;
    value: string;
  }>;
};

export type AuthCli_GetGlobalVpnDeviceQueryVariables = Exact<{
  gvpn: Scalars['String']['input'];
  deviceName: Scalars['String']['input'];
}>;

export type AuthCli_GetGlobalVpnDeviceQuery = {
  infra_getGlobalVPNDevice?: {
    accountName: string;
    creationTime: any;
    displayName: string;
    globalVPNName: string;
    id: string;
    ipAddr: string;
    markedForDeletion?: boolean;
    privateKey: string;
    publicKey: string;
    recordVersion: number;
    updateTime: any;
    createdBy: { userEmail: string; userId: string; userName: string };
    lastUpdatedBy: { userName: string; userId: string; userEmail: string };
    metadata: {
      annotations?: any;
      creationTimestamp: any;
      deletionTimestamp?: any;
      generation: number;
      labels?: any;
      name: string;
      namespace?: string;
    };
    wireguardConfig?: { value: string; encoding: string };
  };
};

export type AuthCli_CoreCheckNameAvailabilityQueryVariables = Exact<{
  resType: ConsoleResType;
  name: Scalars['String']['input'];
}>;

export type AuthCli_CoreCheckNameAvailabilityQuery = {
  core_checkNameAvailability: {
    result: boolean;
    suggestedNames?: Array<string>;
  };
};

export type AuthCli_GetMresKeysQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type AuthCli_GetMresKeysQuery = {
  core_getManagedResouceOutputKeys: Array<string>;
};

export type AuthCli_ListMresesQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type AuthCli_ListMresesQuery = {
  core_listManagedResources?: {
    edges: Array<{
      node: {
        displayName: string;
        metadata?: { name: string; namespace?: string };
      };
    }>;
  };
};

export type AuthCli_GetMresConfigsValuesQueryVariables = Exact<{
  keyrefs?: InputMaybe<
    | Array<InputMaybe<ManagedResourceKeyRefIn>>
    | InputMaybe<ManagedResourceKeyRefIn>
  >;
  envName: Scalars['String']['input'];
}>;

export type AuthCli_GetMresConfigsValuesQuery = {
  core_getManagedResouceOutputKeyValues: Array<{
    key: string;
    mresName: string;
    value: string;
  }>;
};

export type AuthCli_InfraCheckNameAvailabilityQueryVariables = Exact<{
  resType: ResType;
  name: Scalars['String']['input'];
  clusterName?: InputMaybe<Scalars['String']['input']>;
}>;

export type AuthCli_InfraCheckNameAvailabilityQuery = {
  infra_checkNameAvailability: {
    result: boolean;
    suggestedNames: Array<string>;
  };
};

export type AuthCli_GetConfigSecretMapQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  configQueries?: InputMaybe<
    Array<InputMaybe<ConfigKeyRefIn>> | InputMaybe<ConfigKeyRefIn>
  >;
  secretQueries?: InputMaybe<Array<SecretKeyRefIn> | SecretKeyRefIn>;
  mresQueries?: InputMaybe<
    | Array<InputMaybe<ManagedResourceKeyRefIn>>
    | InputMaybe<ManagedResourceKeyRefIn>
  >;
}>;

export type AuthCli_GetConfigSecretMapQuery = {
  configs?: Array<{ configName: string; key: string; value: string }>;
  secrets?: Array<{ key: string; secretName: string; value: string }>;
  mreses: Array<{ key: string; mresName: string; value: string }>;
};

export type AuthCli_InterceptAppMutationVariables = Exact<{
  portMappings?: InputMaybe<
    | Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn>
    | Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppInterceptPortMappingsIn
  >;
  intercept: Scalars['Boolean']['input'];
  deviceName: Scalars['String']['input'];
  appname: Scalars['String']['input'];
  envName: Scalars['String']['input'];
}>;

export type AuthCli_InterceptAppMutation = { core_interceptApp: boolean };

export type AuthCli_GetEnvironmentQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type AuthCli_GetEnvironmentQuery = {
  core_getEnvironment?: {
    displayName: string;
    clusterName: string;
    status?: { isReady: boolean; message?: { RawMessage?: any } };
    metadata?: { name: string };
    spec?: { targetNamespace?: string };
  };
};

export type AuthCli_GetSecretQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type AuthCli_GetSecretQuery = {
  core_getSecret?: {
    displayName: string;
    stringData?: any;
    metadata?: { name: string; namespace?: string };
  };
};

export type AuthCli_GetConfigQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type AuthCli_GetConfigQuery = {
  core_getConfig?: {
    data?: any;
    displayName: string;
    metadata?: { name: string; namespace?: string };
  };
};

export type AuthCli_ListAppsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
}>;

export type AuthCli_ListAppsQuery = {
  core_listApps?: {
    edges: Array<{
      cursor: string;
      node: {
        displayName: string;
        environmentName: string;
        markedForDeletion?: boolean;
        metadata?: { annotations?: any; name: string; namespace?: string };
        spec: {
          displayName?: string;
          nodeSelector?: any;
          replicas?: number;
          serviceAccount?: string;
          containers: Array<{
            args?: Array<string>;
            command?: Array<string>;
            image: string;
            name: string;
            env?: Array<{
              key: string;
              optional?: boolean;
              refKey?: string;
              refName?: string;
              type?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
              value?: string;
            }>;
            envFrom?: Array<{
              refName: string;
              type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
            }>;
          }>;
          intercept?: {
            enabled: boolean;
            toDevice: string;
            portMappings?: Array<{ appPort: number; devicePort: number }>;
          };
          services?: Array<{ port: number }>;
        };
        status?: {
          checks?: any;
          isReady: boolean;
          message?: { RawMessage?: any };
        };
      };
    }>;
  };
};

export type AuthCli_ListConfigsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
}>;

export type AuthCli_ListConfigsQuery = {
  core_listConfigs?: {
    totalCount: number;
    edges: Array<{
      node: {
        data?: any;
        displayName: string;
        metadata?: { name: string; namespace?: string };
      };
    }>;
  };
};

export type AuthCli_ListSecretsQueryVariables = Exact<{
  envName: Scalars['String']['input'];
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type AuthCli_ListSecretsQuery = {
  core_listSecrets?: {
    edges: Array<{
      cursor: string;
      node: {
        displayName: string;
        markedForDeletion?: boolean;
        stringData?: any;
        metadata?: { name: string; namespace?: string };
      };
    }>;
  };
};

export type AuthCli_ListEnvironmentsQueryVariables = Exact<{
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type AuthCli_ListEnvironmentsQuery = {
  core_listEnvironments?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        displayName: string;
        markedForDeletion?: boolean;
        clusterName: string;
        metadata?: { name: string; namespace?: string };
        spec?: { targetNamespace?: string };
        status?: { isReady: boolean; message?: { RawMessage?: any } };
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

export type AuthCli_GetKubeConfigQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;

export type AuthCli_GetKubeConfigQuery = {
  infra_getCluster?: {
    adminKubeconfig?: { encoding: string; value: string };
    status?: { isReady: boolean };
  };
};

export type AuthCli_ListClustersQueryVariables = Exact<{
  pagination?: InputMaybe<CursorPaginationIn>;
}>;

export type AuthCli_ListClustersQuery = {
  infra_listClusters?: {
    edges: Array<{
      node: {
        displayName: string;
        metadata: { name: string };
        status?: { isReady: boolean };
      };
    }>;
  };
};

export type AuthCli_ListAccountsQueryVariables = Exact<{
  [key: string]: never;
}>;

export type AuthCli_ListAccountsQuery = {
  accounts_listAccounts?: Array<{
    displayName: string;
    metadata?: { name: string };
  }>;
};

export type AuthCli_GetCurrentUserQueryVariables = Exact<{
  [key: string]: never;
}>;

export type AuthCli_GetCurrentUserQuery = {
  auth_me?: { id: string; email: string; name: string };
};

export type AuthCli_CreateRemoteLoginMutationVariables = Exact<{
  secret?: InputMaybe<Scalars['String']['input']>;
}>;

export type AuthCli_CreateRemoteLoginMutation = {
  auth_createRemoteLogin: string;
};

export type AuthCli_GetRemoteLoginQueryVariables = Exact<{
  loginId: Scalars['String']['input'];
  secret: Scalars['String']['input'];
}>;

export type AuthCli_GetRemoteLoginQuery = {
  auth_getRemoteLogin?: { authHeader?: string; status: string };
};

export type AuthSetRemoteAuthHeaderMutationVariables = Exact<{
  loginId: Scalars['String']['input'];
  authHeader?: InputMaybe<Scalars['String']['input']>;
}>;

export type AuthSetRemoteAuthHeaderMutation = {
  auth_setRemoteAuthHeader: boolean;
};

export type AuthCheckOauthEnabledQueryVariables = Exact<{
  [key: string]: never;
}>;

export type AuthCheckOauthEnabledQuery = {
  auth_listOAuthProviders?: Array<{ enabled: boolean; provider: string }>;
};

export type AuthAddOauthCredientialsMutationVariables = Exact<{
  provider: Scalars['String']['input'];
  state: Scalars['String']['input'];
  code: Scalars['String']['input'];
}>;

export type AuthAddOauthCredientialsMutation = { oAuth_addLogin: boolean };

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
  auth_me?: {
    verified: boolean;
    name: string;
    id: string;
    email: string;
    providerGitlab?: any;
    providerGithub?: any;
    providerGoogle?: any;
  };
};
