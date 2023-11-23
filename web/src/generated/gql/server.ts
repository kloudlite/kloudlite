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

export type Kloudlite__Io___Apps___Iam___Types__Role =
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

export type SearchManagedServices = {
  isReady?: InputMaybe<MatchFilterIn>;
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

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_GitProvider =
  'github' | 'gitlab';

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_BuildStatus =
  'error' | 'failed' | 'idle' | 'pending' | 'queued' | 'running' | 'success';

export type SearchBuildCacheKeys = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchBuilds = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchCreds = {
  text?: InputMaybe<MatchFilterIn>;
};

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_RepoAccess =
  'read' | 'read_write';

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_ExpirationUnit =
  'd' | 'h' | 'm' | 'w' | 'y';

export type SearchRepos = {
  text?: InputMaybe<MatchFilterIn>;
};

export type PaginationIn = {
  page?: InputMaybe<Scalars['Int']['input']>;
  per_page?: InputMaybe<Scalars['Int']['input']>;
};

export type ResType = 'cluster' | 'nodepool' | 'providersecret' | 'vpn_device';

export type Kloudlite__Io___Pkg___Types__SyncStatusAction = 'APPLY' | 'DELETE';

export type Kloudlite__Io___Pkg___Types__SyncStatusState =
  | 'APPLIED_AT_AGENT'
  | 'ERRORED_AT_AGENT'
  | 'IDLE'
  | 'IN_QUEUE'
  | 'RECEIVED_UPDATE_FROM_AGENT';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ClusterSpecAvailabilityMode =
  'dev' | 'HA';

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider =
  'aws' | 'azure' | 'do' | 'gcp';

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__AwsPoolType =
  'ec2' | 'spot';

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

export type SearchNodepool = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SearchProviderSecret = {
  cloudProviderName?: InputMaybe<MatchFilterIn>;
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
  userRole: Kloudlite__Io___Apps___Iam___Types__Role;
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

export type BuildIn = {
  name: Scalars['String']['input'];
  source: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_GitSourceIn;
  spec: Github_Com__Kloudlite__Operator__Apis__Distribution__V1_BuildRunSpecIn;
};

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_GitSourceIn =
  {
    branch: Scalars['String']['input'];
    provider: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_GitProvider;
    repository: Scalars['String']['input'];
  };

export type Github_Com__Kloudlite__Operator__Apis__Distribution__V1_BuildRunSpecIn =
  {
    buildOptions?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Distribution__V1_BuildOptionsIn>;
    cacheKeyName?: InputMaybe<Scalars['String']['input']>;
    registry: Github_Com__Kloudlite__Operator__Apis__Distribution__V1_RegistryIn;
    resource: Github_Com__Kloudlite__Operator__Apis__Distribution__V1_ResourceIn;
  };

export type Github_Com__Kloudlite__Operator__Apis__Distribution__V1_BuildOptionsIn =
  {
    buildArgs?: InputMaybe<Scalars['Map']['input']>;
    buildContexts?: InputMaybe<Scalars['Map']['input']>;
    contextDir?: InputMaybe<Scalars['String']['input']>;
    dockerfileContent?: InputMaybe<Scalars['String']['input']>;
    dockerfilePath?: InputMaybe<Scalars['String']['input']>;
    targetPlatforms?: InputMaybe<Array<Scalars['String']['input']>>;
  };

export type Github_Com__Kloudlite__Operator__Apis__Distribution__V1_RegistryIn =
  {
    repo: Github_Com__Kloudlite__Operator__Apis__Distribution__V1_RepoIn;
  };

export type Github_Com__Kloudlite__Operator__Apis__Distribution__V1_RepoIn = {
  name: Scalars['String']['input'];
  tags: Array<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Distribution__V1_ResourceIn =
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
  access: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_RepoAccess;
  expiration: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_ExpirationIn;
  name: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_ExpirationIn =
  {
    unit: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_ExpirationUnit;
    value: Scalars['Int']['input'];
  };

export type RepositoryIn = {
  name: Scalars['String']['input'];
};

export type ByocClusterIn = {
  accountName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  metadata?: InputMaybe<MetadataIn>;
  spec: Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ByocSpecIn;
};

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__ByocSpecIn =
  {
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
  };

export type Github__Com___Kloudlite___Operator___Apis___Common____Types__SecretRefIn =
  {
    name: Scalars['String']['input'];
    namespace?: InputMaybe<Scalars['String']['input']>;
  };

export type DomainEntryIn = {
  clusterName: Scalars['String']['input'];
  displayName: Scalars['String']['input'];
  domainName: Scalars['String']['input'];
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

export type CloudProviderSecretIn = {
  aws?: InputMaybe<Kloudlite__Io___Apps___Infra___Internal___Entities__AwsSecretCredentialsIn>;
  cloudProviderName: Github__Com___Kloudlite___Operator___Apis___Common____Types__CloudProvider;
  displayName: Scalars['String']['input'];
  metadata: MetadataIn;
};

export type Kloudlite__Io___Apps___Infra___Internal___Entities__AwsSecretCredentialsIn =
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
    offset?: InputMaybe<Scalars['Int']['input']>;
    ports?: InputMaybe<
      Array<Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__PortIn>
    >;
    serverName: Scalars['String']['input'];
  };

export type Github__Com___Kloudlite___Operator___Apis___Wireguard___V1__PortIn =
  {
    port?: InputMaybe<Scalars['Int']['input']>;
    targetPort?: InputMaybe<Scalars['Int']['input']>;
  };

export type AccountMembershipIn = {
  accountName: Scalars['String']['input'];
  role: Kloudlite__Io___Apps___Iam___Types__Role;
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

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodePropsIn =
  {
    lastRecreatedAt?: InputMaybe<Scalars['Date']['input']>;
  };

export type Github__Com___Kloudlite___Operator___Apis___Clusters___V1__NodeSpecIn =
  {
    nodepoolName: Scalars['String']['input'];
  };

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_GithubUserAccountIn =
  {
    avatarUrl?: InputMaybe<Scalars['String']['input']>;
    id?: InputMaybe<Scalars['Int']['input']>;
    login?: InputMaybe<Scalars['String']['input']>;
    nodeId?: InputMaybe<Scalars['String']['input']>;
    type?: InputMaybe<Scalars['String']['input']>;
  };

export type MembershipIn = {
  accountName: Scalars['String']['input'];
  role: Scalars['String']['input'];
  userId: Scalars['String']['input'];
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
        createdBy: { userName: string; userEmail: string; userId: string };
        lastUpdatedBy: { userName: string; userId: string; userEmail: string };
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
          state: Kloudlite__Io___Pkg___Types__SyncStatusState;
          error?: string;
          action: Kloudlite__Io___Pkg___Types__SyncStatusAction;
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
        spec?: {
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
    clusterToken: string;
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
    spec?: {
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
      action: Kloudlite__Io___Pkg___Types__SyncStatusAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Kloudlite__Io___Pkg___Types__SyncStatusState;
      syncScheduledAt?: any;
    };
    adminKubeconfig?: { value: string; encoding: string };
  };
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
    updateTime: any;
    metadata: { annotations?: any; name: string };
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
    syncStatus: {
      action: Kloudlite__Io___Pkg___Types__SyncStatusAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Kloudlite__Io___Pkg___Types__SyncStatusState;
      syncScheduledAt?: any;
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
        syncStatus: {
          action: Kloudlite__Io___Pkg___Types__SyncStatusAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Kloudlite__Io___Pkg___Types__SyncStatusState;
          syncScheduledAt?: any;
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
    metadata: {
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
        apiVersion?: string;
        kind?: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Kloudlite_Io__Pkg__Types_SyncStatusState;
      syncScheduledAt?: any;
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
        metadata: {
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
            apiVersion?: string;
            kind?: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
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
        metadata: {
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
            apiVersion?: string;
            kind?: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
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
        type?: string;
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
        status?: {
          checks?: any;
          isReady: boolean;
          lastReconcileTime?: any;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion?: string;
            kind?: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
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
        accountName: string;
        apiVersion: string;
        clusterName: string;
        creationTime: any;
        displayName: string;
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
        spec?: {
          offset?: number;
          serverName: string;
          ports?: Array<{ port?: number; targetPort?: number }>;
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
          action: Kloudlite__Io___Pkg___Types__SyncStatusAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Kloudlite__Io___Pkg___Types__SyncStatusState;
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
    accountName: string;
    apiVersion: string;
    clusterName: string;
    creationTime: any;
    displayName: string;
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
    spec?: {
      offset?: number;
      serverName: string;
      ports?: Array<{ port?: number; targetPort?: number }>;
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
      action: Kloudlite__Io___Pkg___Types__SyncStatusAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Kloudlite__Io___Pkg___Types__SyncStatusState;
      syncScheduledAt?: any;
    };
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
    userRole: Kloudlite__Io___Apps___Iam___Types__Role;
  }>;
};

export type ConsoleListMembershipsForAccountQueryVariables = Exact<{
  accountName: Scalars['String']['input'];
}>;

export type ConsoleListMembershipsForAccountQuery = {
  accounts_listMembershipsForAccount?: Array<{
    role: Kloudlite__Io___Apps___Iam___Types__Role;
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
  role: Kloudlite__Io___Apps___Iam___Types__Role;
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

export type ConsoleGetTemplateQueryVariables = Exact<{
  category: Scalars['String']['input'];
  name: Scalars['String']['input'];
}>;

export type ConsoleGetTemplateQuery = {
  core_getManagedServiceTemplate?: {
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
        inputType: string;
        label: string;
        max?: number;
        min?: number;
        name: string;
        required?: boolean;
        unit?: string;
      }>;
      outputs: Array<{ description: string; label: string; name: string }>;
    }>;
  };
};

export type ConsoleListTemplatesQueryVariables = Exact<{
  [key: string]: never;
}>;

export type ConsoleListTemplatesQuery = {
  core_listManagedServiceTemplates?: Array<{
    category: string;
    displayName: string;
    items: Array<{
      description: string;
      active: boolean;
      displayName: string;
      logoUrl: string;
      name: string;
      kind?: string;
      apiVersion?: string;
      fields: Array<{
        defaultValue?: any;
        inputType: string;
        label: string;
        max?: number;
        min?: number;
        name: string;
        required?: boolean;
        unit?: string;
      }>;
      outputs: Array<{ name: string; label: string; description: string }>;
      resources: Array<{
        description: string;
        displayName: string;
        name: string;
        kind?: string;
        apiVersion?: string;
        fields: Array<{
          defaultValue?: any;
          inputType: string;
          label: string;
          max?: number;
          min?: number;
          name: string;
          required?: boolean;
          unit?: string;
        }>;
        outputs: Array<{ description: string; label: string; name: string }>;
      }>;
    }>;
  }>;
};

export type ConsoleGetManagedServiceQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  name: Scalars['String']['input'];
}>;

export type ConsoleGetManagedServiceQuery = {
  core_getManagedService?: {
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
      inputs?: any;
      nodeSelector?: any;
      region?: string;
      msvcKind: { apiVersion: string; kind?: string };
      tolerations?: Array<{
        effect?: string;
        key?: string;
        operator?: string;
        tolerationSeconds?: number;
        value?: string;
      }>;
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReconcileTime?: any;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion?: string;
        kind?: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Kloudlite_Io__Pkg__Types_SyncStatusState;
      syncScheduledAt?: any;
    };
  };
};

export type ConsoleListManagedServicesQueryVariables = Exact<{
  project: ProjectId;
  scope: WorkspaceOrEnvId;
}>;

export type ConsoleListManagedServicesQuery = {
  core_listManagedServices?: {
    totalCount: number;
    edges: Array<{
      cursor: string;
      node: {
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
          inputs?: any;
          nodeSelector?: any;
          region?: string;
          msvcKind: { apiVersion: string; kind?: string };
          tolerations?: Array<{
            effect?: string;
            key?: string;
            operator?: string;
            tolerationSeconds?: number;
            value?: string;
          }>;
        };
        status?: {
          checks?: any;
          isReady: boolean;
          lastReconcileTime?: any;
          message?: { RawMessage?: any };
          resources?: Array<{
            apiVersion?: string;
            kind?: string;
            name: string;
            namespace: string;
          }>;
        };
        syncStatus: {
          action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
          error?: string;
          lastSyncedAt?: any;
          recordVersion: number;
          state: Kloudlite_Io__Pkg__Types_SyncStatusState;
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

export type ConsoleCreateManagedServiceMutationVariables = Exact<{
  msvc: ManagedServiceIn;
}>;

export type ConsoleCreateManagedServiceMutation = {
  core_createManagedService?: { id: string };
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
      inputs?: any;
      mresKind: { kind: string };
      msvcRef: { apiVersion: string; kind?: string; name: string };
    };
    status?: {
      checks?: any;
      isReady: boolean;
      lastReconcileTime?: any;
      message?: { RawMessage?: any };
      resources?: Array<{
        apiVersion?: string;
        kind?: string;
        name: string;
        namespace: string;
      }>;
    };
    syncStatus: {
      action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
      error?: string;
      lastSyncedAt?: any;
      recordVersion: number;
      state: Kloudlite_Io__Pkg__Types_SyncStatusState;
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
        metadata: { name: string };
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
        access: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_RepoAccess;
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
          unit: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_ExpirationUnit;
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
        updateTime: any;
        status: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_BuildStatus;
        name: string;
        id: string;
        createdBy: { userEmail: string; userId: string; userName: string };
        lastUpdatedBy: { userEmail: string; userId: string; userName: string };
      };
    }>;
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
