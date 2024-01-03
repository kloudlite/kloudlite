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

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret =
  'config' | 'secret';

export type K8s__Io___Api___Core___V1__TaintEffect =
  | 'NoExecute'
  | 'NoSchedule'
  | 'PreferNoSchedule';

export type K8s__Io___Api___Core___V1__TolerationOperator = 'Equal' | 'Exists';

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
  | 'RECEIVED_UPDATE_FROM_AGENT'
  | 'UPDATED_AT_AGENT';

export type K8s__Io___Api___Core___V1__SecretType =
  | 'bootstrap__kubernetes__io___token'
  | 'kubernetes__io___basic____auth'
  | 'kubernetes__io___dockercfg'
  | 'kubernetes__io___dockerconfigjson'
  | 'kubernetes__io___service____account____token'
  | 'kubernetes__io___ssh____auth'
  | 'kubernetes__io___tls'
  | 'Opaque';

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
  matchType: MatchFilterMatchType;
  regex?: InputMaybe<Scalars['String']['input']>;
};

export type MatchFilterMatchType = 'array' | 'exact' | 'regex';

export type SearchConfigs = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchWorkspaces = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  projectName?: InputMaybe<MatchFilterIn>;
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

export type SearchProjects = {
  isReady?: InputMaybe<MatchFilterIn>;
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

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__GitProvider =
  'github' | 'gitlab';

export type Github__Com___Kloudlite___Api___Apps___Container____Registry___Internal___Domain___Entities__BuildStatus =
  'error' | 'failed' | 'idle' | 'pending' | 'queued' | 'running' | 'success';

export type SearchBuildCacheKeys = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchBuildRuns = {
  text?: InputMaybe<MatchFilterIn>;
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
  | 'cluster'
  | 'helm_release'
  | 'nodepool'
  | 'providersecret'
  | 'vpn_device';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode =
  'dev' | 'HA';

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider =
  'aws' | 'azure' | 'do' | 'gcp';

export type K8s__Io___Api___Core___V1__NodeSelectorOperator =
  | 'DoesNotExist'
  | 'Exists'
  | 'Gt'
  | 'In'
  | 'Lt'
  | 'NotIn';

export type K8s__Io___Apimachinery___Pkg___Apis___Meta___V1__LabelSelectorOperator =
  'DoesNotExist' | 'Exists' | 'In' | 'NotIn';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType =
  'ec2' | 'spot';

export type K8s__Io___Api___Core___V1__ConditionStatus =
  | 'False'
  | 'True'
  | 'Unknown';

export type K8s__Io___Api___Core___V1__PersistentVolumeClaimConditionType =
  | 'FileSystemResizePending'
  | 'Resizing';

export type K8s__Io___Api___Core___V1__PersistentVolumeClaimPhase =
  | 'Bound'
  | 'Lost'
  | 'Pending';

export type SearchClusterManagedService = {
  isReady?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchCluster = {
  cloudProviderName?: InputMaybe<MatchFilterIn>;
  isReady?: InputMaybe<MatchFilterIn>;
  region?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchDomainEntry = {
  clusterName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchHelmRelease = {
  isReady?: InputMaybe<MatchFilterIn>;
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

export type SearchVpnDevices = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type AccountIn = {
  contactEmail: Scalars['String']['input'];
  description?: InputMaybe<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  isActive?: InputMaybe<Scalars['Boolean']['input']>;
  logo?: InputMaybe<Scalars['String']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__AccountSpecIn;
};

export type MetadataIn = {
  annotations?: InputMaybe<Scalars['Map']['input']>;
  labels?: InputMaybe<Scalars['Map']['input']>;
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__AccountSpecIn =
  {
    targetNamespace?: InputMaybe<Scalars['String']['input']>;
  };

export type InvitationIn = {
  userEmail?: InputMaybe<Scalars['String']['input']>;
  userName?: InputMaybe<Scalars['String']['input']>;
  userRole: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
};

export type AppIn = {
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
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
  serviceAccount?: InputMaybe<Scalars['String']['input']>;
  services?: InputMaybe<
    Array<Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppSvcIn>
  >;
  tolerations?: InputMaybe<Array<K8s__Io___Api___Core___V1__TolerationIn>>;
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
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  maxReplicas?: InputMaybe<Scalars['Int']['input']>;
  minReplicas?: InputMaybe<Scalars['Int']['input']>;
  thresholdCpu?: InputMaybe<Scalars['Int']['input']>;
  thresholdMemory?: InputMaybe<Scalars['Int']['input']>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__InterceptIn =
  {
    enabled: Scalars['Boolean']['input'];
    toDevice: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__AppSvcIn = {
  name?: InputMaybe<Scalars['String']['input']>;
  port: Scalars['Int']['input'];
  targetPort?: InputMaybe<Scalars['Int']['input']>;
  type?: InputMaybe<Scalars['String']['input']>;
};

export type K8s__Io___Api___Core___V1__TolerationIn = {
  effect?: InputMaybe<K8s__Io___Api___Core___V1__TaintEffect>;
  key?: InputMaybe<Scalars['String']['input']>;
  operator?: InputMaybe<K8s__Io___Api___Core___V1__TolerationOperator>;
  tolerationSeconds?: InputMaybe<Scalars['Int']['input']>;
  value?: InputMaybe<Scalars['String']['input']>;
};

export type ConfigIn = {
  data?: InputMaybe<Scalars['Map']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  metadata?: InputMaybe<MetadataIn>;
};

export type WorkspaceIn = {
  displayName: Scalars['String']['input'];
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__WorkspaceSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__WorkspaceSpecIn =
  {
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
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedResourceSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedResourceSpecIn =
  {
    resourceTemplate: Github__Com___Kloudlite___Operator___Apis___Crds___V1__MresResourceTemplateIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__MresResourceTemplateIn =
  {
    msvcRef: Github__Com___Kloudlite___Operator___Apis___Crds___V1__MsvcNamedRefIn;
    spec: Scalars['Map']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__MsvcNamedRefIn =
  {
    apiVersion: Scalars['String']['input'];
    kind: Scalars['String']['input'];
    name: Scalars['String']['input'];
    namespace: Scalars['String']['input'];
  };

export type ProjectIn = {
  displayName: Scalars['String']['input'];
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ProjectSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ProjectSpecIn =
  {
    displayName?: InputMaybe<Scalars['String']['input']>;
    logo?: InputMaybe<Scalars['String']['input']>;
    targetNamespace: Scalars['String']['input'];
  };

export type RouterIn = {
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
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
    region?: InputMaybe<Scalars['String']['input']>;
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
  app?: InputMaybe<Scalars['String']['input']>;
  lambda?: InputMaybe<Scalars['String']['input']>;
  path: Scalars['String']['input'];
  port: Scalars['Int']['input'];
  rewrite?: InputMaybe<Scalars['Boolean']['input']>;
};

export type SecretIn = {
  data?: InputMaybe<Scalars['Map']['input']>;
  displayName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  metadata?: InputMaybe<MetadataIn>;
  stringData?: InputMaybe<Scalars['Map']['input']>;
  type?: InputMaybe<K8s__Io___Api___Core___V1__SecretType>;
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
    cacheKeyName?: InputMaybe<Scalars['String']['input']>;
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

export type BuildCacheKeyIn = {
  displayName: Scalars['String']['input'];
  name: Scalars['String']['input'];
  volumeSizeInGB: Scalars['Float']['input'];
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

export type ClusterIn = {
  displayName: Scalars['String']['input'];
  metadata: MetadataIn;
  spec: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecIn =
  {
    availabilityMode: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode;
    aws?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsClusterConfigIn>;
    cloudflareEnabled?: InputMaybe<Scalars['Boolean']['input']>;
    cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
    credentialsRef: Github__Com___Kloudlite___Operator___Apis___Common____Types__SecretRefIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsClusterConfigIn =
  {
    k3sMasters?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__Awsk3sMastersConfigIn>;
    region: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__Awsk3sMastersConfigIn =
  {
    instanceType: Scalars['String']['input'];
    nvidiaGpuEnabled: Scalars['Boolean']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__SecretRefIn =
  {
    name: Scalars['String']['input'];
    namespace?: InputMaybe<Scalars['String']['input']>;
  };

export type ClusterManagedServiceIn = {
  displayName: Scalars['String']['input'];
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__ClusterManagedServiceSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ClusterManagedServiceSpecIn =
  {
    msvcSpec: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedServiceSpecIn;
    namespace: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ManagedServiceSpecIn =
  {
    serviceTemplate: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ServiceTemplateIn;
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ServiceTemplateIn =
  {
    apiVersion: Scalars['String']['input'];
    kind: Scalars['String']['input'];
    spec: Scalars['Map']['input'];
  };

export type DomainEntryIn = {
  clusterName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  domainName: Scalars['String']['input'];
};

export type HelmReleaseIn = {
  displayName: Scalars['String']['input'];
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__HelmChartSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__HelmChartSpecIn =
  {
    chartName: Scalars['String']['input'];
    chartRepo: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ChartRepoIn;
    chartVersion: Scalars['String']['input'];
    jobVars?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Crds___V1__JobVarsIn>;
    postInstall?: InputMaybe<Scalars['String']['input']>;
    postUninstall?: InputMaybe<Scalars['String']['input']>;
    preInstall?: InputMaybe<Scalars['String']['input']>;
    preUninstall?: InputMaybe<Scalars['String']['input']>;
    values: Scalars['Map']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Crds___V1__ChartRepoIn =
  {
    name: Scalars['String']['input'];
    url: Scalars['String']['input'];
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

export type K8s__Io___Api___Core___V1__PodAntiAffinityIn = {
  preferredDuringSchedulingIgnoredDuringExecution?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__WeightedPodAffinityTermIn>
  >;
  requiredDuringSchedulingIgnoredDuringExecution?: InputMaybe<
    Array<K8s__Io___Api___Core___V1__PodAffinityTermIn>
  >;
};

export type NodePoolIn = {
  displayName: Scalars['String']['input'];
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodePoolSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodePoolSpecIn =
  {
    aws?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsNodePoolConfigIn>;
    cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
    maxCount: Scalars['Int']['input'];
    minCount: Scalars['Int']['input'];
    nodeLabels?: InputMaybe<Scalars['Map']['input']>;
    nodeTaints?: InputMaybe<Array<K8s__Io___Api___Core___V1__TaintIn>>;
    targetCount: Scalars['Int']['input'];
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
  metadata: MetadataIn;
};

export type Github__Com___Kloudlite___Api___Apps___Infra___Internal___Entities__AwsSecretCredentialsIn =
  {
    accessKey?: InputMaybe<Scalars['String']['input']>;
    awsAccountId?: InputMaybe<Scalars['String']['input']>;
    secretKey?: InputMaybe<Scalars['String']['input']>;
  };

export type VpnDeviceIn = {
  displayName: Scalars['String']['input'];
  metadata?: InputMaybe<MetadataIn>;
  spec?: InputMaybe<Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__DeviceSpecIn>;
};

export type Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__DeviceSpecIn =
  {
    cnameRecords?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__CNameRecordIn>
    >;
    deviceNamespace?: InputMaybe<Scalars['String']['input']>;
    nodeSelector?: InputMaybe<Scalars['Map']['input']>;
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

export type AccountMembershipIn = {
  accountName: Scalars['String']['input'];
  role: Github__Com___Kloudlite___Api___Apps___Iam___Types__Role;
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

export type Github__Com___Kloudlite___Operator___Pkg___Operator__CheckIn = {
  generation?: InputMaybe<Scalars['Int']['input']>;
  message?: InputMaybe<Scalars['String']['input']>;
  status: Scalars['Boolean']['input'];
};

export type Github__Com___Kloudlite___Operator___Pkg___Operator__ResourceRefIn =
  {
    name: Scalars['String']['input'];
    namespace: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Pkg___Operator__StatusIn = {
  checks?: InputMaybe<Scalars['Map']['input']>;
  isReady: Scalars['Boolean']['input'];
  lastReadyGeneration?: InputMaybe<Scalars['Int']['input']>;
  lastReconcileTime?: InputMaybe<Scalars['Date']['input']>;
  message?: InputMaybe<Github__Com___Kloudlite___Operator___Pkg___Raw____Json__RawJsonIn>;
  resources?: InputMaybe<
    Array<Github__Com___Kloudlite___Operator___Pkg___Operator__ResourceRefIn>
  >;
};

export type Github__Com___Kloudlite___Operator___Pkg___Raw____Json__RawJsonIn =
  {
    RawMessage?: InputMaybe<Scalars['Any']['input']>;
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

export type K8s__Io___Apimachinery___Pkg___Api___Resource__Format =
  | 'BinarySI'
  | 'DecimalExponent'
  | 'DecimalSI';

export type K8s__Io___Apimachinery___Pkg___Api___Resource__QuantityIn = {
  Format: K8s__Io___Apimachinery___Pkg___Api___Resource__Format;
};

export type NodeIn = {
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodeSpecIn;
};

export type SearchEnvironments = {
  isReady?: InputMaybe<MatchFilterIn>;
  markedForDeletion?: InputMaybe<MatchFilterIn>;
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
  namespace?: InputMaybe<Scalars['String']['input']>;
}>;

export type ConsoleCoreCheckNameAvailabilityQuery = {
  core_checkNameAvailability: {
    result: boolean;
    suggestedNames?: Array<string>;
  };
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
    updateTime: any;
    contactEmail: string;
    displayName: string;
    metadata?: { name: string; annotations?: any };
    spec: { targetNamespace?: string };
  };
};

export type ConsoleCreateProjectMutationVariables = Exact<{
  project: ProjectIn;
}>;

export type ConsoleCreateProjectMutation = {
  core_createProject?: { id: string };
};

export type ConsoleUpdateProjectMutationVariables = Exact<{
  project: ProjectIn;
}>;

export type ConsoleUpdateProjectMutation = {
  core_updateProject?: { id: string };
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
    metadata?: {
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
        kind: string;
        apiVersion: string;
        namespace: string;
      }>;
      message?: { RawMessage?: any };
    };
    syncStatus: {
      syncScheduledAt?: any;
      state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
      recordVersion: number;
      lastSyncedAt?: any;
      error?: string;
      action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
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
        createdBy: { userName: string; userEmail: string; userId: string };
        lastUpdatedBy: { userName: string; userId: string; userEmail: string };
        metadata?: {
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
            kind: string;
            apiVersion: string;
            namespace: string;
          }>;
          message?: { RawMessage?: any };
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
        markedForDeletion?: boolean;
        creationTime: any;
        updateTime: any;
        recordVersion: number;
        metadata: { name: string; annotations?: any };
        lastUpdatedBy: { userId: string; userName: string; userEmail: string };
        createdBy: { userEmail: string; userId: string; userName: string };
        syncStatus: {
          syncScheduledAt?: any;
          lastSyncedAt?: any;
          recordVersion: number;
          error?: string;
        };
        status?: {
          lastReconcileTime?: any;
          isReady: boolean;
          checks?: any;
          resources?: Array<{
            namespace: string;
            name: string;
            kind: string;
            apiVersion: string;
          }>;
          message?: { RawMessage?: any };
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
          credentialsRef: { namespace?: string; name: string };
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
    apiVersion: string;
    creationTime: any;
    displayName: string;
    id: string;
    kind: string;
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
          imageId: string;
          imageSSHUsername: string;
          instanceType: string;
          nodes?: any;
          nvidiaGpuEnabled: boolean;
          rootVolumeSize: number;
          rootVolumeType: string;
        };
      };
      clusterTokenRef?: { key: string; name: string; namespace?: string };
      credentialKeys?: {
        keyAccessKey: string;
        keyAWSAccountId: string;
        keyAWSAssumeRoleExternalID: string;
        keyAWSAssumeRoleRoleARN: string;
        keyIAMInstanceProfileRole: string;
        keySecretKey: string;
      };
      credentialsRef: { name: string; namespace?: string };
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
        aws?: { awsAccountId?: string };
        createdBy: { userEmail: string; userId: string; userName: string };
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
    aws?: { awsAccountId?: string };
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
    clusterName: string;
    creationTime: any;
    displayName: string;
    kind: string;
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
      targetCount: number;
      aws?: {
        availabilityZone: string;
        iamInstanceProfileRole?: string;
        imageId: string;
        imageSSHUsername: string;
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
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      startCursor?: string;
    };
    edges: Array<{
      cursor: string;
      node: {
        clusterName: string;
        creationTime: any;
        displayName: string;
        markedForDeletion?: boolean;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: { name: string };
        spec: {
          cloudProvider: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
          maxCount: number;
          minCount: number;
          targetCount: number;
          aws?: {
            availabilityZone: string;
            nvidiaGpuEnabled: boolean;
            poolType: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType;
            ec2Pool?: { instanceType: string; nodes?: any };
            spotPool?: {
              nodes?: any;
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
    }>;
  };
};

export type ConsoleDeleteNodePoolMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  poolName: Scalars['String']['input'];
}>;

export type ConsoleDeleteNodePoolMutation = { infra_deleteNodePool: boolean };

export type ConsoleGetWorkspaceQueryVariables = Exact<{
  project: ProjectId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetWorkspaceQuery = {
  core_getWorkspace?: {
    displayName: string;
    clusterName: string;
    updateTime: any;
    metadata?: {
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

export type ConsoleGetEnvironmentQueryVariables = Exact<{
  project: ProjectId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetEnvironmentQuery = {
  core_getEnvironment?: {
    accountName: string;
    apiVersion: string;
    clusterName: string;
    creationTime: any;
    displayName: string;
    id: string;
    kind: string;
    markedForDeletion?: boolean;
    projectName: string;
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
      isEnvironment?: boolean;
      projectName: string;
      targetNamespace: string;
    };
    status?: {
      checks?: any;
      isReady: boolean;
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
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        apiVersion: string;
        clusterName: string;
        creationTime: any;
        displayName: string;
        id: string;
        kind: string;
        markedForDeletion?: boolean;
        projectName: string;
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
          isEnvironment?: boolean;
          projectName: string;
          targetNamespace: string;
        };
        status?: {
          checks?: any;
          isReady: boolean;
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
    }>;
    pageInfo: {
      endCursor?: string;
      hasNextPage?: boolean;
      hasPreviousPage?: boolean;
      startCursor?: string;
    };
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
    metadata?: { name: string; namespace?: string; annotations?: any };
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
        operator?: K8s__Io___Api___Core___V1__TolerationOperator;
        key?: string;
        effect?: K8s__Io___Api___Core___V1__TaintEffect;
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
          type?: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
          value?: string;
        }>;
        envFrom?: Array<{
          type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
          refName: string;
        }>;
        livenessProbe?: {
          type: string;
          interval?: number;
          initialDelay?: number;
          failureThreshold?: number;
          tcp?: { port: number };
          shell?: { command?: Array<string> };
          httpGet?: { httpHeaders?: any; path: string; port: number };
        };
        readinessProbe?: {
          type: string;
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
          type: Github__Com___Kloudlite___Operator___Apis___Crds___V1__ConfigOrSecret;
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
      message?: { RawMessage?: any };
    };
    syncStatus: {
      syncScheduledAt?: any;
      state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
      recordVersion: number;
      lastSyncedAt?: any;
      error?: string;
      action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
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
        metadata?: { name: string };
        spec: { freeze?: boolean; displayName?: string };
        status?: {
          lastReconcileTime?: any;
          isReady: boolean;
          checks?: any;
          message?: { RawMessage?: any };
        };
        syncStatus: {
          syncScheduledAt?: any;
          state: Github__Com___Kloudlite___Api___Pkg___Types__SyncState;
          recordVersion: number;
          lastSyncedAt?: any;
          error?: string;
          action: Github__Com___Kloudlite___Api___Pkg___Types__SyncAction;
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
        metadata?: {
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
    metadata?: {
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
  pq?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchConfigs>;
}>;

export type ConsoleListConfigsQuery = {
  core_listConfigs?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        apiVersion: string;
        clusterName: string;
        creationTime: any;
        data?: any;
        displayName: string;
        enabled?: boolean;
        id: string;
        kind: string;
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
        status?: {
          checks?: any;
          isReady: boolean;
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

export type ConsoleCreateConfigMutationVariables = Exact<{
  config: ConfigIn;
}>;

export type ConsoleCreateConfigMutation = {
  core_createConfig?: { id: string };
};

export type ConsoleListSecretsQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchSecrets>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListSecretsQuery = {
  core_listSecrets?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        accountName: string;
        apiVersion: string;
        clusterName: string;
        creationTime: any;
        data?: any;
        displayName: string;
        enabled?: boolean;
        id: string;
        kind: string;
        markedForDeletion?: boolean;
        recordVersion: number;
        stringData?: any;
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
        status?: {
          checks?: any;
          isReady: boolean;
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
    metadata?: {
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

export type ConsoleDeleteSecretMutationVariables = Exact<{
  namespace: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleDeleteSecretMutation = { core_deleteSecret: boolean };

export type ConsoleCreateVpnDeviceMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  vpnDevice: VpnDeviceIn;
}>;

export type ConsoleCreateVpnDeviceMutation = {
  infra_createVPNDevice?: { id: string };
};

export type ConsoleUpdateVpnDeviceMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  vpnDevice: VpnDeviceIn;
}>;

export type ConsoleUpdateVpnDeviceMutation = {
  infra_updateVPNDevice?: { id: string };
};

export type ConsoleListVpnDevicesQueryVariables = Exact<{
  clusterName?: InputMaybe<Scalars['String']['input']>;
  search?: InputMaybe<SearchVpnDevices>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListVpnDevicesQuery = {
  infra_listVPNDevices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        clusterName: string;
        creationTime: any;
        displayName: string;
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
        spec?: { ports?: Array<{ port?: number; targetPort?: number }> };
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

export type ConsoleGetVpnDeviceQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetVpnDeviceQuery = {
  infra_getVPNDevice?: {
    clusterName: string;
    creationTime: any;
    displayName: string;
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
    spec?: { ports?: Array<{ port?: number; targetPort?: number }> };
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
    wireguardConfig?: { value: string; encoding: string };
  };
};

export type ConsoleDeleteVpnDeviceMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  deviceName: Scalars['String']['input'];
}>;

export type ConsoleDeleteVpnDeviceMutation = { infra_deleteVPNDevice: boolean };

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

export type ConsoleGetManagedResourceQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetManagedResourceQuery = {
  core_getManagedResource?: {
    accountName: string;
    apiVersion: string;
    clusterName: string;
    creationTime: any;
    displayName: string;
    enabled?: boolean;
    id: string;
    kind: string;
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
        msvcRef: { name: string; kind: string; apiVersion: string };
      };
    };
    status?: {
      checks?: any;
      isReady: boolean;
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
};

export type ConsoleListManagedResourceQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchManagedResources>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListManagedResourceQuery = {
  core_listManagedResources?: {
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
        kind: string;
        displayName: string;
        creationTime: any;
        metadata?: { name: string };
        lastUpdatedBy: { userEmail: string; userName: string };
        createdBy: { userEmail: string; userName: string };
      };
    }>;
  };
};

export type ConsoleCreateManagedResourceMutationVariables = Exact<{
  mres: ManagedResourceIn;
}>;

export type ConsoleCreateManagedResourceMutation = {
  core_createManagedResource?: { id: string };
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
          accountName: string;
          cacheKeyName?: string;
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

export type ConsoleListBuildCachesQueryVariables = Exact<{
  pq?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchBuildCacheKeys>;
}>;

export type ConsoleListBuildCachesQuery = {
  cr_listBuildCacheKeys?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        id: string;
        creationTime: any;
        displayName: string;
        name: string;
        updateTime: any;
        volumeSizeInGB: number;
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

export type ConsoleCreateBuildCacheMutationVariables = Exact<{
  buildCacheKey: BuildCacheKeyIn;
}>;

export type ConsoleCreateBuildCacheMutation = {
  cr_addBuildCacheKey?: { id: string };
};

export type ConsoleUpdateBuildCachesMutationVariables = Exact<{
  crUpdateBuildCacheKeyId: Scalars['ID']['input'];
  buildCacheKey: BuildCacheKeyIn;
}>;

export type ConsoleUpdateBuildCachesMutation = {
  cr_updateBuildCacheKey?: { id: string };
};

export type ConsoleDeleteBuildCacheMutationVariables = Exact<{
  crDeleteBuildCacheKeyId: Scalars['ID']['input'];
}>;

export type ConsoleDeleteBuildCacheMutation = {
  cr_deleteBuildCacheKey: boolean;
};

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

export type ConsoleListBuildRunsQueryVariables = Exact<{
  repoName: Scalars['String']['input'];
  search?: InputMaybe<SearchBuildRuns>;
  pq?: InputMaybe<CursorPaginationIn>;
}>;

export type ConsoleListBuildRunsQuery = {
  cr_listBuildRuns?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
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
          accountName: string;
          cacheKeyName?: string;
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

export type ConsoleGetBuildRunQueryVariables = Exact<{
  repoName: Scalars['String']['input'];
  buildRunName: Scalars['String']['input'];
}>;

export type ConsoleGetBuildRunQuery = {
  cr_getBuildRun?: {
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
      accountName: string;
      cacheKeyName?: string;
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

export type ConsoleGetClusterMSvQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetClusterMSvQuery = {
  infra_getClusterManagedService?: {
    displayName: string;
    creationTime: any;
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
      namespace: string;
      msvcSpec: {
        serviceTemplate: { apiVersion: string; kind: string; spec: any };
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

export type ConsoleCreateClusterMSvMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  service: ClusterManagedServiceIn;
}>;

export type ConsoleCreateClusterMSvMutation = {
  infra_createClusterManagedService?: { id: string };
};

export type ConsoleUpdateClusterMSvMutationVariables = Exact<{
  clusterName: Scalars['String']['input'];
  service: ClusterManagedServiceIn;
}>;

export type ConsoleUpdateClusterMSvMutation = {
  infra_updateClusterManagedService?: { id: string };
};

export type ConsoleListClusterMSvsQueryVariables = Exact<{
  clusterName: Scalars['String']['input'];
}>;

export type ConsoleListClusterMSvsQuery = {
  infra_listClusterManagedServices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
        markedForDeletion?: boolean;
        updateTime: any;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
        metadata?: { name: string };
        spec?: {
          msvcSpec: {
            serviceTemplate: { apiVersion: string; kind: string; spec: any };
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
  clusterName: Scalars['String']['input'];
  serviceName: Scalars['String']['input'];
}>;

export type ConsoleDeleteClusterMSvMutation = {
  infra_deleteClusterManagedService: boolean;
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
      }>;
    }>;
  }>;
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
      chartName: string;
      chartVersion: string;
      postInstall?: string;
      postUninstall?: string;
      preInstall?: string;
      preUninstall?: string;
      values: any;
      chartRepo: { name: string; url: string };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReadyGeneration?: number;
      lastReconcileTime?: any;
      releaseNotes: string;
      releaseStatus: string;
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
}>;

export type ConsoleListHelmChartQuery = {
  infra_listHelmReleases?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
        creationTime: any;
        displayName: string;
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
          chartName: string;
          chartVersion: string;
          postInstall?: string;
          postUninstall?: string;
          preInstall?: string;
          preUninstall?: string;
          values: any;
          chartRepo: { name: string; url: string };
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReadyGeneration?: number;
          lastReconcileTime?: any;
          releaseNotes: string;
          releaseStatus: string;
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
