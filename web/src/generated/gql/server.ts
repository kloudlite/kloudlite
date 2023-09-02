export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Date: { input: any; output: any; }
  Map: { input: any; output: any; }
  Any: { input: any; output: any; }
  Json: { input: any; output: any; }
  ProviderDetail: { input: any; output: any; }
  URL: { input: any; output: any; }
};

export type Query = {
  accounts_checkNameAvailability: AccountsCheckNameAvailabilityOutput;
  accounts_getAccount?: Maybe<Account>;
  accounts_getAccountMembership?: Maybe<AccountMembership>;
  accounts_getInvitation?: Maybe<Invitation>;
  accounts_listAccounts?: Maybe<Array<Maybe<Account>>>;
  accounts_listInvitations?: Maybe<Array<Invitation>>;
  accounts_listMembershipsForAccount?: Maybe<Array<AccountMembership>>;
  accounts_listMembershipsForUser?: Maybe<Array<AccountMembership>>;
  accounts_resyncAccount: Scalars['Boolean']['output'];
  auth_findByEmail?: Maybe<User>;
  auth_getRemoteLogin?: Maybe<RemoteLogin>;
  auth_listOAuthProviders?: Maybe<Array<OAuthProviderStatus>>;
  auth_me?: Maybe<User>;
  core_checkNameAvailability: ConsoleCheckNameAvailabilityOutput;
  core_getApp?: Maybe<App>;
  core_getConfig?: Maybe<Config>;
  core_getEnvironment?: Maybe<Workspace>;
  core_getImagePullSecret?: Maybe<ImagePullSecret>;
  core_getManagedResource?: Maybe<ManagedResource>;
  core_getManagedService?: Maybe<ManagedService>;
  core_getManagedServiceTemplate?: Maybe<Kloudlite_Io__Apps__Console__Internal__Entities_MsvcTemplateEntry>;
  core_getProject?: Maybe<Project>;
  core_getRouter?: Maybe<Router>;
  core_getSecret?: Maybe<Secret>;
  core_getVPNDevice?: Maybe<VpnDevice>;
  core_getWorkspace?: Maybe<Workspace>;
  core_listApps?: Maybe<AppPaginatedRecords>;
  core_listConfigs?: Maybe<ConfigPaginatedRecords>;
  core_listEnvironments?: Maybe<WorkspacePaginatedRecords>;
  core_listImagePullSecrets?: Maybe<ImagePullSecretPaginatedRecords>;
  core_listManagedResources?: Maybe<ManagedResourcePaginatedRecords>;
  core_listManagedServices?: Maybe<ManagedServicePaginatedRecords>;
  core_listManagedServiceTemplates?: Maybe<Array<MsvcTemplate>>;
  core_listProjects?: Maybe<ProjectPaginatedRecords>;
  core_listRouters?: Maybe<RouterPaginatedRecords>;
  core_listSecrets?: Maybe<SecretPaginatedRecords>;
  core_listVPNDevices?: Maybe<VpnDevicePaginatedRecords>;
  core_listWorkspaces?: Maybe<WorkspacePaginatedRecords>;
  core_resyncApp: Scalars['Boolean']['output'];
  core_resyncConfig: Scalars['Boolean']['output'];
  core_resyncEnvironment: Scalars['Boolean']['output'];
  core_resyncImagePullSecret: Scalars['Boolean']['output'];
  core_resyncManagedResource: Scalars['Boolean']['output'];
  core_resyncManagedService: Scalars['Boolean']['output'];
  core_resyncProject: Scalars['Boolean']['output'];
  core_resyncRouter: Scalars['Boolean']['output'];
  core_resyncSecret: Scalars['Boolean']['output'];
  core_resyncWorkspace: Scalars['Boolean']['output'];
  cr_listArtifacts: Array<Artifact>;
  cr_listRepos: Array<Repo>;
  cr_listRobots: Array<HarborRobotUser>;
  infra_checkNameAvailability: CheckNameAvailabilityOutput;
  infra_getBYOCCluster?: Maybe<ByocCluster>;
  infra_getCluster?: Maybe<Cluster>;
  infra_getNode?: Maybe<Node>;
  infra_getNodePool?: Maybe<NodePool>;
  infra_getProviderSecret?: Maybe<CloudProviderSecret>;
  infra_listBYOCClusters?: Maybe<ByocClusterPaginatedRecords>;
  infra_listClusters?: Maybe<ClusterPaginatedRecords>;
  infra_listNodePools?: Maybe<NodePoolPaginatedRecords>;
  infra_listNodes?: Maybe<NodePaginatedRecords>;
  infra_listProviderSecrets?: Maybe<CloudProviderSecretPaginatedRecords>;
  oAuth_requestLogin: Scalars['URL']['output'];
};


export type QueryAccounts_CheckNameAvailabilityArgs = {
  name: Scalars['String']['input'];
};


export type QueryAccounts_GetAccountArgs = {
  accountName: Scalars['String']['input'];
};


export type QueryAccounts_GetAccountMembershipArgs = {
  accountName: Scalars['String']['input'];
};


export type QueryAccounts_GetInvitationArgs = {
  accountName: Scalars['String']['input'];
  invitationId: Scalars['String']['input'];
};


export type QueryAccounts_ListInvitationsArgs = {
  accountName: Scalars['String']['input'];
};


export type QueryAccounts_ListMembershipsForAccountArgs = {
  accountName: Scalars['String']['input'];
};


export type QueryAccounts_ResyncAccountArgs = {
  accountName: Scalars['String']['input'];
};


export type QueryAuth_FindByEmailArgs = {
  email: Scalars['String']['input'];
};


export type QueryAuth_GetRemoteLoginArgs = {
  loginId: Scalars['String']['input'];
  secret: Scalars['String']['input'];
};


export type QueryCore_CheckNameAvailabilityArgs = {
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
  resType: ConsoleResType;
};


export type QueryCore_GetAppArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_GetConfigArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_GetEnvironmentArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
};


export type QueryCore_GetImagePullSecretArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope?: InputMaybe<WorkspaceOrEnvId>;
};


export type QueryCore_GetManagedResourceArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_GetManagedServiceArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_GetManagedServiceTemplateArgs = {
  category: Scalars['String']['input'];
  name: Scalars['String']['input'];
};


export type QueryCore_GetProjectArgs = {
  name: Scalars['String']['input'];
};


export type QueryCore_GetRouterArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_GetSecretArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_GetVpnDeviceArgs = {
  name: Scalars['String']['input'];
};


export type QueryCore_GetWorkspaceArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
};


export type QueryCore_ListAppsArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchApps>;
};


export type QueryCore_ListConfigsArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchConfigs>;
};


export type QueryCore_ListEnvironmentsArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  search?: InputMaybe<SearchWorkspaces>;
};


export type QueryCore_ListImagePullSecretsArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  scope?: InputMaybe<WorkspaceOrEnvId>;
  search?: InputMaybe<SearchImagePullSecrets>;
};


export type QueryCore_ListManagedResourcesArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchManagedResources>;
};


export type QueryCore_ListManagedServicesArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchManagedServices>;
};


export type QueryCore_ListProjectsArgs = {
  clusterName?: InputMaybe<Scalars['String']['input']>;
  pq?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchProjects>;
};


export type QueryCore_ListRoutersArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchRouters>;
};


export type QueryCore_ListSecretsArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  scope: WorkspaceOrEnvId;
  search?: InputMaybe<SearchSecrets>;
};


export type QueryCore_ListVpnDevicesArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchVpnDevices>;
};


export type QueryCore_ListWorkspacesArgs = {
  pq?: InputMaybe<CursorPaginationIn>;
  project: ProjectId;
  search?: InputMaybe<SearchWorkspaces>;
};


export type QueryCore_ResyncAppArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_ResyncConfigArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_ResyncEnvironmentArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
};


export type QueryCore_ResyncImagePullSecretArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope?: InputMaybe<WorkspaceOrEnvId>;
};


export type QueryCore_ResyncManagedResourceArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_ResyncManagedServiceArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_ResyncProjectArgs = {
  name: Scalars['String']['input'];
};


export type QueryCore_ResyncRouterArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_ResyncSecretArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
  scope: WorkspaceOrEnvId;
};


export type QueryCore_ResyncWorkspaceArgs = {
  name: Scalars['String']['input'];
  project: ProjectId;
};


export type QueryCr_ListArtifactsArgs = {
  repoName: Scalars['String']['input'];
};


export type QueryInfra_CheckNameAvailabilityArgs = {
  clusterName?: InputMaybe<Scalars['String']['input']>;
  name: Scalars['String']['input'];
  resType: ResType;
};


export type QueryInfra_GetByocClusterArgs = {
  name: Scalars['String']['input'];
};


export type QueryInfra_GetClusterArgs = {
  name: Scalars['String']['input'];
};


export type QueryInfra_GetNodeArgs = {
  clusterName: Scalars['String']['input'];
  nodeName: Scalars['String']['input'];
  poolName: Scalars['String']['input'];
};


export type QueryInfra_GetNodePoolArgs = {
  clusterName: Scalars['String']['input'];
  poolName: Scalars['String']['input'];
};


export type QueryInfra_GetProviderSecretArgs = {
  name: Scalars['String']['input'];
};


export type QueryInfra_ListByocClustersArgs = {
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchCluster>;
};


export type QueryInfra_ListClustersArgs = {
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchCluster>;
};


export type QueryInfra_ListNodePoolsArgs = {
  clusterName: Scalars['String']['input'];
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchNodepool>;
};


export type QueryInfra_ListNodesArgs = {
  clusterName: Scalars['String']['input'];
  pagination?: InputMaybe<CursorPaginationIn>;
  poolName: Scalars['String']['input'];
};


export type QueryInfra_ListProviderSecretsArgs = {
  pagination?: InputMaybe<CursorPaginationIn>;
  search?: InputMaybe<SearchProviderSecret>;
};


export type QueryOAuth_RequestLoginArgs = {
  provider: Scalars['String']['input'];
  state?: InputMaybe<Scalars['String']['input']>;
};

export type AccountsCheckNameAvailabilityOutput = {
  result: Scalars['Boolean']['output'];
  suggestedNames?: Maybe<Array<Scalars['String']['output']>>;
};

export type Account = {
  apiVersion: Scalars['String']['output'];
  contactEmail: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  id: Scalars['String']['output'];
  isActive?: Maybe<Scalars['Boolean']['output']>;
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Scalars['Map']['output'];
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  updateTime: Scalars['Date']['output'];
};

export type Metadata = {
  annotations?: Maybe<Scalars['Map']['output']>;
  creationTimestamp: Scalars['Date']['output'];
  deletionTimestamp?: Maybe<Scalars['Date']['output']>;
  generation: Scalars['Int']['output'];
  labels?: Maybe<Scalars['Map']['output']>;
  name: Scalars['String']['output'];
  namespace?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Pkg__Operator_Status = {
  checks?: Maybe<Scalars['Map']['output']>;
  isReady: Scalars['Boolean']['output'];
  lastReconcileTime?: Maybe<Scalars['Date']['output']>;
  message?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Raw___Json_RawJson>;
  resources?: Maybe<Array<Github_Com__Kloudlite__Operator__Pkg__Operator_ResourceRef>>;
};

export type Github_Com__Kloudlite__Operator__Pkg__Raw___Json_RawJson = {
  RawMessage?: Maybe<Scalars['Any']['output']>;
};

export type Github_Com__Kloudlite__Operator__Pkg__Operator_ResourceRef = {
  apiVersion?: Maybe<Scalars['String']['output']>;
  kind?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  namespace: Scalars['String']['output'];
};

export type AccountMembership = {
  accountName: Scalars['String']['output'];
  role: Scalars['String']['output'];
  userId: Scalars['String']['output'];
};

export type Invitation = {
  accepted?: Maybe<Scalars['Boolean']['output']>;
  accountName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  id: Scalars['String']['output'];
  invitedBy: Scalars['String']['output'];
  inviteToken: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  recordVersion: Scalars['Int']['output'];
  rejected?: Maybe<Scalars['Boolean']['output']>;
  updateTime: Scalars['Date']['output'];
  userEmail?: Maybe<Scalars['String']['output']>;
  userName?: Maybe<Scalars['String']['output']>;
  userRole: Scalars['String']['output'];
};

export type User = {
  accountMembership?: Maybe<AccountMembership>;
  accountMemberships?: Maybe<Array<AccountMembership>>;
  avatar?: Maybe<Scalars['String']['output']>;
  email: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  invite: Scalars['String']['output'];
  joined: Scalars['Date']['output'];
  metadata?: Maybe<Scalars['Json']['output']>;
  name: Scalars['String']['output'];
  providerGithub?: Maybe<Scalars['ProviderDetail']['output']>;
  providerGitlab?: Maybe<Scalars['ProviderDetail']['output']>;
  providerGoogle?: Maybe<Scalars['ProviderDetail']['output']>;
  verified: Scalars['Boolean']['output'];
};


export type UserAccountMembershipArgs = {
  accountName: Scalars['String']['input'];
};

export type RemoteLogin = {
  authHeader?: Maybe<Scalars['String']['output']>;
  status: Scalars['String']['output'];
};

export type OAuthProviderStatus = {
  enabled: Scalars['Boolean']['output'];
  provider: Scalars['String']['output'];
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

export type ConsoleCheckNameAvailabilityOutput = {
  result: Scalars['Boolean']['output'];
  suggestedNames?: Maybe<Array<Scalars['String']['output']>>;
};

export type ProjectId = {
  type: ProjectIdType;
  value: Scalars['String']['input'];
};

export type ProjectIdType =
  | 'name'
  | 'targetNamespace';

export type WorkspaceOrEnvId = {
  type: WorkspaceOrEnvIdType;
  value: Scalars['String']['input'];
};

export type WorkspaceOrEnvIdType =
  | 'environmentName'
  | 'environmentTargetNamespace'
  | 'workspaceName'
  | 'workspaceTargetNamespace';

export type App = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  projectName: Scalars['String']['output'];
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
  workspaceName: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpec = {
  containers: Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainers>>;
  displayName?: Maybe<Scalars['String']['output']>;
  freeze?: Maybe<Scalars['Boolean']['output']>;
  hpa?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecHpa>;
  intercept?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecIntercept>;
  nodeSelector?: Maybe<Scalars['Map']['output']>;
  region?: Maybe<Scalars['String']['output']>;
  replicas?: Maybe<Scalars['Int']['output']>;
  serviceAccount?: Maybe<Scalars['String']['output']>;
  services?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecServices>>>;
  tolerations?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecTolerations>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainers = {
  args?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  command?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  env?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnv>>>;
  envFrom?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFrom>>>;
  image: Scalars['String']['output'];
  imagePullPolicy?: Maybe<Scalars['String']['output']>;
  livenessProbe?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbe>;
  name: Scalars['String']['output'];
  readinessProbe?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbe>;
  resourceCpu?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceCpu>;
  resourceMemory?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceMemory>;
  volumes?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumes>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnv = {
  key: Scalars['String']['output'];
  optional?: Maybe<Scalars['Boolean']['output']>;
  refKey?: Maybe<Scalars['String']['output']>;
  refName?: Maybe<Scalars['String']['output']>;
  type?: Maybe<Scalars['String']['output']>;
  value?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFrom = {
  refName: Scalars['String']['output'];
  type: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbe = {
  failureThreshold?: Maybe<Scalars['Int']['output']>;
  httpGet?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeHttpGet>;
  initialDelay?: Maybe<Scalars['Int']['output']>;
  interval?: Maybe<Scalars['Int']['output']>;
  shell?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeShell>;
  tcp?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeTcp>;
  type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeType;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeHttpGet = {
  httpHeaders?: Maybe<Scalars['Map']['output']>;
  path: Scalars['String']['output'];
  port: Scalars['Int']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeShell = {
  command?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeTcp = {
  port: Scalars['Int']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeType =
  | 'httpGet'
  | 'shell'
  | 'tcp';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbe = {
  failureThreshold?: Maybe<Scalars['Int']['output']>;
  httpGet?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeHttpGet>;
  initialDelay?: Maybe<Scalars['Int']['output']>;
  interval?: Maybe<Scalars['Int']['output']>;
  shell?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeShell>;
  tcp?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeTcp>;
  type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeType;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeHttpGet = {
  httpHeaders?: Maybe<Scalars['Map']['output']>;
  path: Scalars['String']['output'];
  port: Scalars['Int']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeShell = {
  command?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeTcp = {
  port: Scalars['Int']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeType =
  | 'httpGet'
  | 'shell'
  | 'tcp';

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceCpu = {
  max?: Maybe<Scalars['String']['output']>;
  min?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceMemory = {
  max?: Maybe<Scalars['String']['output']>;
  min?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumes = {
  items?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesItems>>>;
  mountPath: Scalars['String']['output'];
  refName: Scalars['String']['output'];
  type: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesItems = {
  fileName?: Maybe<Scalars['String']['output']>;
  key: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecHpa = {
  enabled?: Maybe<Scalars['Boolean']['output']>;
  maxReplicas?: Maybe<Scalars['Int']['output']>;
  minReplicas?: Maybe<Scalars['Int']['output']>;
  thresholdCpu?: Maybe<Scalars['Int']['output']>;
  thresholdMemory?: Maybe<Scalars['Int']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecIntercept = {
  enabled: Scalars['Boolean']['output'];
  toDevice: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecServices = {
  name?: Maybe<Scalars['String']['output']>;
  port: Scalars['Int']['output'];
  targetPort?: Maybe<Scalars['Int']['output']>;
  type?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecTolerations = {
  effect?: Maybe<Scalars['String']['output']>;
  key?: Maybe<Scalars['String']['output']>;
  operator?: Maybe<Scalars['String']['output']>;
  tolerationSeconds?: Maybe<Scalars['Int']['output']>;
  value?: Maybe<Scalars['String']['output']>;
};

export type Kloudlite_Io__Pkg__Types_SyncStatus = {
  action: Kloudlite_Io__Pkg__Types_SyncStatusAction;
  error?: Maybe<Scalars['String']['output']>;
  lastSyncedAt?: Maybe<Scalars['Date']['output']>;
  recordVersion: Scalars['Int']['output'];
  state: Kloudlite_Io__Pkg__Types_SyncStatusState;
  syncScheduledAt?: Maybe<Scalars['Date']['output']>;
};

export type Kloudlite_Io__Pkg__Types_SyncStatusAction =
  | 'APPLY'
  | 'DELETE';

export type Kloudlite_Io__Pkg__Types_SyncStatusState =
  | 'APPLIED_AT_AGENT'
  | 'ERRORED_AT_AGENT'
  | 'IDLE'
  | 'IN_QUEUE'
  | 'RECEIVED_UPDATE_FROM_AGENT';

export type Config = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  data?: Maybe<Scalars['Map']['output']>;
  displayName: Scalars['String']['output'];
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Workspace = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  projectName: Scalars['String']['output'];
  recordVersion: Scalars['Int']['output'];
  spec?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_WorkspaceSpec>;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_WorkspaceSpec = {
  projectName: Scalars['String']['output'];
  targetNamespace: Scalars['String']['output'];
};

export type ImagePullSecret = {
  accountName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  dockerConfigJson?: Maybe<Scalars['String']['output']>;
  dockerPassword?: Maybe<Scalars['String']['output']>;
  dockerRegistryEndpoint?: Maybe<Scalars['String']['output']>;
  dockerUsername?: Maybe<Scalars['String']['output']>;
  id: Scalars['String']['output'];
  name: Scalars['String']['output'];
  updateTime: Scalars['Date']['output'];
};

export type ManagedResource = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpec = {
  inputs?: Maybe<Scalars['Map']['output']>;
  mresKind: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMresKind;
  msvcRef: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMsvcRef;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMresKind = {
  kind: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMsvcRef = {
  apiVersion: Scalars['String']['output'];
  kind?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
};

export type ManagedService = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpec = {
  inputs?: Maybe<Scalars['Map']['output']>;
  msvcKind: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecMsvcKind;
  nodeSelector?: Maybe<Scalars['Map']['output']>;
  region?: Maybe<Scalars['String']['output']>;
  tolerations?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecTolerations>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecMsvcKind = {
  apiVersion: Scalars['String']['output'];
  kind?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecTolerations = {
  effect?: Maybe<Scalars['String']['output']>;
  key?: Maybe<Scalars['String']['output']>;
  operator?: Maybe<Scalars['String']['output']>;
  tolerationSeconds?: Maybe<Scalars['Int']['output']>;
  value?: Maybe<Scalars['String']['output']>;
};

export type Kloudlite_Io__Apps__Console__Internal__Entities_MsvcTemplateEntry = {
  active: Scalars['Boolean']['output'];
  description: Scalars['String']['output'];
  displayName: Scalars['String']['output'];
  fields: Array<Kloudlite_Io__Apps__Console__Internal__Entities_InputField>;
  logoUrl: Scalars['String']['output'];
  name: Scalars['String']['output'];
  outputs: Array<Kloudlite_Io__Apps__Console__Internal__Entities_OutputField>;
  resources: Array<Kloudlite_Io__Apps__Console__Internal__Entities_MresTemplate>;
};

export type Kloudlite_Io__Apps__Console__Internal__Entities_InputField = {
  defaultValue: Scalars['Any']['output'];
  inputType: Scalars['String']['output'];
  label: Scalars['String']['output'];
  max?: Maybe<Scalars['Float']['output']>;
  min?: Maybe<Scalars['Float']['output']>;
  name: Scalars['String']['output'];
  required?: Maybe<Scalars['Boolean']['output']>;
  unit?: Maybe<Scalars['String']['output']>;
};

export type Kloudlite_Io__Apps__Console__Internal__Entities_OutputField = {
  description: Scalars['String']['output'];
  label: Scalars['String']['output'];
  name: Scalars['String']['output'];
};

export type Kloudlite_Io__Apps__Console__Internal__Entities_MresTemplate = {
  description: Scalars['String']['output'];
  displayName: Scalars['String']['output'];
  fields: Array<Kloudlite_Io__Apps__Console__Internal__Entities_InputField>;
  name: Scalars['String']['output'];
  outputs: Array<Kloudlite_Io__Apps__Console__Internal__Entities_OutputField>;
};

export type Project = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ProjectSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ProjectSpec = {
  accountName: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  displayName?: Maybe<Scalars['String']['output']>;
  logo?: Maybe<Scalars['String']['output']>;
  targetNamespace: Scalars['String']['output'];
};

export type Router = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpec = {
  backendProtocol?: Maybe<Scalars['String']['output']>;
  basicAuth?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecBasicAuth>;
  cors?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecCors>;
  domains: Array<Maybe<Scalars['String']['output']>>;
  https?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecHttps>;
  ingressClass?: Maybe<Scalars['String']['output']>;
  maxBodySizeInMB?: Maybe<Scalars['Int']['output']>;
  rateLimit?: Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRateLimit>;
  region?: Maybe<Scalars['String']['output']>;
  routes?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRoutes>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecBasicAuth = {
  enabled: Scalars['Boolean']['output'];
  secretName?: Maybe<Scalars['String']['output']>;
  username?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecCors = {
  allowCredentials?: Maybe<Scalars['Boolean']['output']>;
  enabled?: Maybe<Scalars['Boolean']['output']>;
  origins?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecHttps = {
  clusterIssuer?: Maybe<Scalars['String']['output']>;
  enabled: Scalars['Boolean']['output'];
  forceRedirect?: Maybe<Scalars['Boolean']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRateLimit = {
  connections?: Maybe<Scalars['Int']['output']>;
  enabled?: Maybe<Scalars['Boolean']['output']>;
  rpm?: Maybe<Scalars['Int']['output']>;
  rps?: Maybe<Scalars['Int']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRoutes = {
  app?: Maybe<Scalars['String']['output']>;
  lambda?: Maybe<Scalars['String']['output']>;
  path: Scalars['String']['output'];
  port: Scalars['Int']['output'];
  rewrite?: Maybe<Scalars['Boolean']['output']>;
};

export type Secret = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  data?: Maybe<Scalars['Map']['output']>;
  displayName: Scalars['String']['output'];
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  stringData?: Maybe<Scalars['Map']['output']>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  type?: Maybe<Scalars['String']['output']>;
  updateTime: Scalars['Date']['output'];
};

export type VpnDevice = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  createdBy: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  displayName: Scalars['String']['output'];
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec?: Maybe<Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpec>;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpec = {
  offset: Scalars['Int']['output'];
  ports?: Maybe<Array<Maybe<Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecPorts>>>;
  serverName: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecPorts = {
  port?: Maybe<Scalars['Int']['output']>;
  targetPort?: Maybe<Scalars['Int']['output']>;
};

export type CursorPaginationIn = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  orderBy?: InputMaybe<Scalars['String']['input']>;
  sortDirection?: InputMaybe<CursorPaginationSortDirection>;
};

export type CursorPaginationSortDirection =
  | 'ASC'
  | 'DESC';

export type SearchApps = {
  text?: InputMaybe<MatchFilterIn>;
};

export type MatchFilterIn = {
  array?: InputMaybe<Array<Scalars['Any']['input']>>;
  exact?: InputMaybe<Scalars['Any']['input']>;
  matchType: MatchFilterMatchType;
  regex?: InputMaybe<Scalars['String']['input']>;
};

export type MatchFilterMatchType =
  | 'array'
  | 'exact'
  | 'regex';

export type AppPaginatedRecords = {
  edges: Array<AppEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type AppEdge = {
  cursor: Scalars['String']['output'];
  node: App;
};

export type PageInfo = {
  endCursor?: Maybe<Scalars['String']['output']>;
  hasNextPage?: Maybe<Scalars['Boolean']['output']>;
  hasPreviousPage?: Maybe<Scalars['Boolean']['output']>;
  startCursor?: Maybe<Scalars['String']['output']>;
};

export type SearchConfigs = {
  text?: InputMaybe<MatchFilterIn>;
};

export type ConfigPaginatedRecords = {
  edges: Array<ConfigEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ConfigEdge = {
  cursor: Scalars['String']['output'];
  node: Config;
};

export type SearchWorkspaces = {
  projectName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type WorkspacePaginatedRecords = {
  edges: Array<WorkspaceEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type WorkspaceEdge = {
  cursor: Scalars['String']['output'];
  node: Workspace;
};

export type SearchImagePullSecrets = {
  text?: InputMaybe<MatchFilterIn>;
};

export type ImagePullSecretPaginatedRecords = {
  edges: Array<ImagePullSecretEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ImagePullSecretEdge = {
  cursor: Scalars['String']['output'];
  node: ImagePullSecret;
};

export type SearchManagedResources = {
  text?: InputMaybe<MatchFilterIn>;
};

export type ManagedResourcePaginatedRecords = {
  edges: Array<ManagedResourceEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ManagedResourceEdge = {
  cursor: Scalars['String']['output'];
  node: ManagedResource;
};

export type SearchManagedServices = {
  text?: InputMaybe<MatchFilterIn>;
};

export type ManagedServicePaginatedRecords = {
  edges: Array<ManagedServiceEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ManagedServiceEdge = {
  cursor: Scalars['String']['output'];
  node: ManagedService;
};

export type MsvcTemplate = {
  category: Scalars['String']['output'];
  displayName: Scalars['String']['output'];
  items: Array<Kloudlite_Io__Apps__Console__Internal__Entities_MsvcTemplateEntry>;
};

export type SearchProjects = {
  text?: InputMaybe<MatchFilterIn>;
};

export type ProjectPaginatedRecords = {
  edges: Array<ProjectEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ProjectEdge = {
  cursor: Scalars['String']['output'];
  node: Project;
};

export type SearchRouters = {
  text?: InputMaybe<MatchFilterIn>;
};

export type RouterPaginatedRecords = {
  edges: Array<RouterEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type RouterEdge = {
  cursor: Scalars['String']['output'];
  node: Router;
};

export type SearchSecrets = {
  text?: InputMaybe<MatchFilterIn>;
};

export type SecretPaginatedRecords = {
  edges: Array<SecretEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type SecretEdge = {
  cursor: Scalars['String']['output'];
  node: Secret;
};

export type SearchVpnDevices = {
  text?: InputMaybe<MatchFilterIn>;
};

export type VpnDevicePaginatedRecords = {
  edges: Array<VpnDeviceEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type VpnDeviceEdge = {
  cursor: Scalars['String']['output'];
  node: VpnDevice;
};

export type Artifact = {
  size: Scalars['Int']['output'];
  tags: Array<ImageTag>;
};

export type ImageTag = {
  immutable: Scalars['Boolean']['output'];
  name: Scalars['String']['output'];
  pushedAt: Scalars['String']['output'];
  signed: Scalars['Boolean']['output'];
};

export type Repo = {
  artifactCount: Scalars['Int']['output'];
  id: Scalars['Int']['output'];
  name: Scalars['String']['output'];
  pullCount: Scalars['Int']['output'];
};

export type HarborRobotUser = {
  apiVersion?: Maybe<Scalars['String']['output']>;
  creationTime: Scalars['Date']['output'];
  id: Scalars['String']['output'];
  kind?: Maybe<Scalars['String']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec?: Maybe<Github_Com__Kloudlite__Operator__Apis__Artifacts__V1_HarborUserAccountSpec>;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Artifacts__V1_HarborUserAccountSpec = {
  accountName: Scalars['String']['output'];
  enabled?: Maybe<Scalars['Boolean']['output']>;
  harborProjectName: Scalars['String']['output'];
  permissions?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  targetSecret?: Maybe<Scalars['String']['output']>;
};

export type ResType =
  | 'cluster'
  | 'nodepool'
  | 'providersecret';

export type CheckNameAvailabilityOutput = {
  result: Scalars['Boolean']['output'];
  suggestedNames: Array<Scalars['String']['output']>;
};

export type ByocCluster = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterToken: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  helmStatus: Scalars['Map']['output'];
  id: Scalars['String']['output'];
  incomingKafkaTopicName: Scalars['String']['output'];
  isConnected: Scalars['Boolean']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ByocSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ByocSpec = {
  accountName: Scalars['String']['output'];
  displayName?: Maybe<Scalars['String']['output']>;
  incomingKafkaTopic: Scalars['String']['output'];
  ingressClasses?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  provider: Scalars['String']['output'];
  publicIps?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  region: Scalars['String']['output'];
  storageClasses?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type Cluster = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterToken: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec?: Maybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpec>;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpec = {
  accountName: Scalars['String']['output'];
  agentHelmValuesRef?: Maybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAgentHelmValuesRef>;
  availabilityMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode;
  cloudProvider: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider;
  credentialsRef: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCredentialsRef;
  nodeIps?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  operatorsHelmValuesRef?: Maybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecOperatorsHelmValuesRef>;
  region: Scalars['String']['output'];
  vpc?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAgentHelmValuesRef = {
  key: Scalars['String']['output'];
  name: Scalars['String']['output'];
  namespace?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAvailabilityMode =
  | 'dev'
  | 'HA';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCloudProvider =
  | 'aws'
  | 'azure'
  | 'do'
  | 'gcp';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCredentialsRef = {
  name: Scalars['String']['output'];
  namespace?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecOperatorsHelmValuesRef = {
  key: Scalars['String']['output'];
  name: Scalars['String']['output'];
  namespace?: Maybe<Scalars['String']['output']>;
};

export type Node = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpec = {
  clusterName?: Maybe<Scalars['String']['output']>;
  nodePoolName?: Maybe<Scalars['String']['output']>;
  nodeType: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecNodeType;
  taints?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecNodeType =
  | 'cluster'
  | 'master'
  | 'worker';

export type NodePool = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  clusterName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  spec: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpec;
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  syncStatus: Kloudlite_Io__Pkg__Types_SyncStatus;
  updateTime: Scalars['Date']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpec = {
  awsNodeConfig?: Maybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfig>;
  maxCount: Scalars['Int']['output'];
  minCount: Scalars['Int']['output'];
  targetCount: Scalars['Int']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfig = {
  imageId?: Maybe<Scalars['String']['output']>;
  isGpu?: Maybe<Scalars['Boolean']['output']>;
  onDemandSpecs?: Maybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigOnDemandSpecs>;
  provisionMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode;
  region?: Maybe<Scalars['String']['output']>;
  spotSpecs?: Maybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigSpotSpecs>;
  vpc?: Maybe<Scalars['String']['output']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigOnDemandSpecs = {
  instanceType: Scalars['String']['output'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode =
  | 'on_demand'
  | 'reserved'
  | 'spot';

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigSpotSpecs = {
  cpuMax: Scalars['Int']['output'];
  cpuMin: Scalars['Int']['output'];
  memMax: Scalars['Int']['output'];
  memMin: Scalars['Int']['output'];
};

export type CloudProviderSecret = {
  accountName: Scalars['String']['output'];
  apiVersion: Scalars['String']['output'];
  cloudProviderName: CloudProviderSecretCloudProviderName;
  creationTime: Scalars['Date']['output'];
  data?: Maybe<Scalars['Map']['output']>;
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  markedForDeletion?: Maybe<Scalars['Boolean']['output']>;
  metadata: Metadata;
  recordVersion: Scalars['Int']['output'];
  status?: Maybe<Github_Com__Kloudlite__Operator__Pkg__Operator_Status>;
  stringData?: Maybe<Scalars['Map']['output']>;
  type?: Maybe<Scalars['String']['output']>;
  updateTime: Scalars['Date']['output'];
};

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

export type ByocClusterPaginatedRecords = {
  edges: Array<ByocClusterEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ByocClusterEdge = {
  cursor: Scalars['String']['output'];
  node: ByocCluster;
};

export type ClusterPaginatedRecords = {
  edges: Array<ClusterEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type ClusterEdge = {
  cursor: Scalars['String']['output'];
  node: Cluster;
};

export type SearchNodepool = {
  text?: InputMaybe<MatchFilterIn>;
};

export type NodePoolPaginatedRecords = {
  edges: Array<NodePoolEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type NodePoolEdge = {
  cursor: Scalars['String']['output'];
  node: NodePool;
};

export type NodePaginatedRecords = {
  edges: Array<NodeEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type NodeEdge = {
  cursor: Scalars['String']['output'];
  node: Node;
};

export type SearchProviderSecret = {
  cloudProviderName?: InputMaybe<MatchFilterIn>;
  text?: InputMaybe<MatchFilterIn>;
};

export type CloudProviderSecretPaginatedRecords = {
  edges: Array<CloudProviderSecretEdge>;
  pageInfo: PageInfo;
  totalCount: Scalars['Int']['output'];
};

export type CloudProviderSecretEdge = {
  cursor: Scalars['String']['output'];
  node: CloudProviderSecret;
};

export type Mutation = {
  accounts_acceptInvitation: Scalars['Boolean']['output'];
  accounts_activateAccount: Scalars['Boolean']['output'];
  accounts_createAccount: Account;
  accounts_deactivateAccount: Scalars['Boolean']['output'];
  accounts_deleteAccount: Scalars['Boolean']['output'];
  accounts_deleteInvitation: Scalars['Boolean']['output'];
  accounts_inviteMember: Invitation;
  accounts_rejectInvitation: Scalars['Boolean']['output'];
  accounts_removeAccountMembership: Scalars['Boolean']['output'];
  accounts_resendInviteMail: Scalars['Boolean']['output'];
  accounts_updateAccount: Account;
  accounts_updateAccountMembership: Scalars['Boolean']['output'];
  auth_changeEmail: Scalars['Boolean']['output'];
  auth_changePassword: Scalars['Boolean']['output'];
  auth_clearMetadata: User;
  auth_createRemoteLogin: Scalars['String']['output'];
  auth_login?: Maybe<Session>;
  auth_logout: Scalars['Boolean']['output'];
  auth_requestResetPassword: Scalars['Boolean']['output'];
  auth_resendVerificationEmail: Scalars['Boolean']['output'];
  auth_resetPassword: Scalars['Boolean']['output'];
  auth_setMetadata: User;
  auth_setRemoteAuthHeader: Scalars['Boolean']['output'];
  auth_signup?: Maybe<Session>;
  auth_verifyEmail: Session;
  core_createApp?: Maybe<App>;
  core_createConfig?: Maybe<Config>;
  core_createEnvironment?: Maybe<Workspace>;
  core_createImagePullSecret?: Maybe<ImagePullSecret>;
  core_createManagedResource?: Maybe<ManagedResource>;
  core_createManagedService?: Maybe<ManagedService>;
  core_createProject?: Maybe<Project>;
  core_createRouter?: Maybe<Router>;
  core_createSecret?: Maybe<Secret>;
  core_createVPNDevice?: Maybe<VpnDevice>;
  core_createWorkspace?: Maybe<Workspace>;
  core_deleteApp: Scalars['Boolean']['output'];
  core_deleteConfig: Scalars['Boolean']['output'];
  core_deleteEnvironment: Scalars['Boolean']['output'];
  core_deleteImagePullSecret: Scalars['Boolean']['output'];
  core_deleteManagedResource: Scalars['Boolean']['output'];
  core_deleteManagedService: Scalars['Boolean']['output'];
  core_deleteProject: Scalars['Boolean']['output'];
  core_deleteRouter: Scalars['Boolean']['output'];
  core_deleteSecret: Scalars['Boolean']['output'];
  core_deleteVPNDevice: Scalars['Boolean']['output'];
  core_deleteWorkspace: Scalars['Boolean']['output'];
  core_updateApp?: Maybe<App>;
  core_updateConfig?: Maybe<Config>;
  core_updateEnvironment?: Maybe<Workspace>;
  core_updateManagedResource?: Maybe<ManagedResource>;
  core_updateManagedService?: Maybe<ManagedService>;
  core_updateProject?: Maybe<Project>;
  core_updateRouter?: Maybe<Router>;
  core_updateSecret?: Maybe<Secret>;
  core_updateVPNDevice?: Maybe<VpnDevice>;
  core_updateWorkspace?: Maybe<Workspace>;
  cr_createRobot?: Maybe<HarborRobotUser>;
  cr_deleteRepo: Scalars['Boolean']['output'];
  cr_deleteRobot: Scalars['Boolean']['output'];
  cr_resyncRobot: Scalars['Boolean']['output'];
  cr_updateRobot?: Maybe<HarborRobotUser>;
  generateClusterToken: Scalars['String']['output'];
  infra_createBYOCCluster?: Maybe<ByocCluster>;
  infra_createCluster?: Maybe<Cluster>;
  infra_createNodePool?: Maybe<NodePool>;
  infra_createProviderSecret?: Maybe<CloudProviderSecret>;
  infra_deleteBYOCCluster: Scalars['Boolean']['output'];
  infra_deleteCluster: Scalars['Boolean']['output'];
  infra_deleteNodePool: Scalars['Boolean']['output'];
  infra_deleteProviderSecret: Scalars['Boolean']['output'];
  infra_updateBYOCCluster?: Maybe<ByocCluster>;
  infra_updateCluster?: Maybe<Cluster>;
  infra_updateNodePool?: Maybe<NodePool>;
  infra_updateProviderSecret?: Maybe<CloudProviderSecret>;
  oAuth_addLogin: Scalars['Boolean']['output'];
  oAuth_login: Session;
};


export type MutationAccounts_AcceptInvitationArgs = {
  accountName: Scalars['String']['input'];
  inviteToken: Scalars['String']['input'];
};


export type MutationAccounts_ActivateAccountArgs = {
  accountName: Scalars['String']['input'];
};


export type MutationAccounts_CreateAccountArgs = {
  account: AccountIn;
};


export type MutationAccounts_DeactivateAccountArgs = {
  accountName: Scalars['String']['input'];
};


export type MutationAccounts_DeleteAccountArgs = {
  accountName: Scalars['String']['input'];
};


export type MutationAccounts_DeleteInvitationArgs = {
  accountName: Scalars['String']['input'];
  invitationId: Scalars['String']['input'];
};


export type MutationAccounts_InviteMemberArgs = {
  accountName: Scalars['String']['input'];
  invitation: InvitationIn;
};


export type MutationAccounts_RejectInvitationArgs = {
  accountName: Scalars['String']['input'];
  inviteToken: Scalars['String']['input'];
};


export type MutationAccounts_RemoveAccountMembershipArgs = {
  accountName: Scalars['String']['input'];
  memberId: Scalars['ID']['input'];
};


export type MutationAccounts_ResendInviteMailArgs = {
  accountName: Scalars['String']['input'];
  invitationId: Scalars['String']['input'];
};


export type MutationAccounts_UpdateAccountArgs = {
  account: AccountIn;
};


export type MutationAccounts_UpdateAccountMembershipArgs = {
  accountName: Scalars['String']['input'];
  memberId: Scalars['ID']['input'];
  role: Scalars['String']['input'];
};


export type MutationAuth_ChangeEmailArgs = {
  email: Scalars['String']['input'];
};


export type MutationAuth_ChangePasswordArgs = {
  currentPassword: Scalars['String']['input'];
  newPassword: Scalars['String']['input'];
};


export type MutationAuth_CreateRemoteLoginArgs = {
  secret?: InputMaybe<Scalars['String']['input']>;
};


export type MutationAuth_LoginArgs = {
  email: Scalars['String']['input'];
  password: Scalars['String']['input'];
};


export type MutationAuth_RequestResetPasswordArgs = {
  email: Scalars['String']['input'];
};


export type MutationAuth_ResetPasswordArgs = {
  password: Scalars['String']['input'];
  token: Scalars['String']['input'];
};


export type MutationAuth_SetMetadataArgs = {
  values: Scalars['Json']['input'];
};


export type MutationAuth_SetRemoteAuthHeaderArgs = {
  authHeader?: InputMaybe<Scalars['String']['input']>;
  loginId: Scalars['String']['input'];
};


export type MutationAuth_SignupArgs = {
  email: Scalars['String']['input'];
  name: Scalars['String']['input'];
  password: Scalars['String']['input'];
};


export type MutationAuth_VerifyEmailArgs = {
  token: Scalars['String']['input'];
};


export type MutationCore_CreateAppArgs = {
  app: AppIn;
};


export type MutationCore_CreateConfigArgs = {
  config: ConfigIn;
};


export type MutationCore_CreateEnvironmentArgs = {
  env: WorkspaceIn;
};


export type MutationCore_CreateImagePullSecretArgs = {
  imagePullSecretIn: ImagePullSecretIn;
};


export type MutationCore_CreateManagedResourceArgs = {
  mres: ManagedResourceIn;
};


export type MutationCore_CreateManagedServiceArgs = {
  msvc: ManagedServiceIn;
};


export type MutationCore_CreateProjectArgs = {
  project: ProjectIn;
};


export type MutationCore_CreateRouterArgs = {
  router: RouterIn;
};


export type MutationCore_CreateSecretArgs = {
  secret: SecretIn;
};


export type MutationCore_CreateVpnDeviceArgs = {
  vpnDevice: VpnDeviceIn;
};


export type MutationCore_CreateWorkspaceArgs = {
  env: WorkspaceIn;
};


export type MutationCore_DeleteAppArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteConfigArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteEnvironmentArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteImagePullSecretArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteManagedResourceArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteManagedServiceArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteProjectArgs = {
  name: Scalars['String']['input'];
};


export type MutationCore_DeleteRouterArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteSecretArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_DeleteVpnDeviceArgs = {
  deviceName: Scalars['String']['input'];
};


export type MutationCore_DeleteWorkspaceArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
};


export type MutationCore_UpdateAppArgs = {
  app: AppIn;
};


export type MutationCore_UpdateConfigArgs = {
  config: ConfigIn;
};


export type MutationCore_UpdateEnvironmentArgs = {
  env: WorkspaceIn;
};


export type MutationCore_UpdateManagedResourceArgs = {
  mres: ManagedResourceIn;
};


export type MutationCore_UpdateManagedServiceArgs = {
  msvc: ManagedServiceIn;
};


export type MutationCore_UpdateProjectArgs = {
  project: ProjectIn;
};


export type MutationCore_UpdateRouterArgs = {
  router: RouterIn;
};


export type MutationCore_UpdateSecretArgs = {
  secret: SecretIn;
};


export type MutationCore_UpdateVpnDeviceArgs = {
  vpnDevice: VpnDeviceIn;
};


export type MutationCore_UpdateWorkspaceArgs = {
  env: WorkspaceIn;
};


export type MutationCr_CreateRobotArgs = {
  robotUser: HarborRobotUserIn;
};


export type MutationCr_DeleteRepoArgs = {
  repoId: Scalars['Int']['input'];
};


export type MutationCr_DeleteRobotArgs = {
  robotId: Scalars['Int']['input'];
};


export type MutationCr_ResyncRobotArgs = {
  name: Scalars['String']['input'];
};


export type MutationCr_UpdateRobotArgs = {
  name: Scalars['String']['input'];
  permissions?: InputMaybe<Array<HarborPermission>>;
};


export type MutationGenerateClusterTokenArgs = {
  accountName: Scalars['String']['input'];
  clusterName: Scalars['String']['input'];
};


export type MutationInfra_CreateByocClusterArgs = {
  byocCluster: ByocClusterIn;
};


export type MutationInfra_CreateClusterArgs = {
  cluster: ClusterIn;
};


export type MutationInfra_CreateNodePoolArgs = {
  clusterName: Scalars['String']['input'];
  pool: NodePoolIn;
};


export type MutationInfra_CreateProviderSecretArgs = {
  secret: CloudProviderSecretIn;
};


export type MutationInfra_DeleteByocClusterArgs = {
  name: Scalars['String']['input'];
};


export type MutationInfra_DeleteClusterArgs = {
  name: Scalars['String']['input'];
};


export type MutationInfra_DeleteNodePoolArgs = {
  clusterName: Scalars['String']['input'];
  poolName: Scalars['String']['input'];
};


export type MutationInfra_DeleteProviderSecretArgs = {
  secretName: Scalars['String']['input'];
};


export type MutationInfra_UpdateByocClusterArgs = {
  byocCluster: ByocClusterIn;
};


export type MutationInfra_UpdateClusterArgs = {
  cluster: ClusterIn;
};


export type MutationInfra_UpdateNodePoolArgs = {
  clusterName: Scalars['String']['input'];
  pool: NodePoolIn;
};


export type MutationInfra_UpdateProviderSecretArgs = {
  secret: CloudProviderSecretIn;
};


export type MutationOAuth_AddLoginArgs = {
  code: Scalars['String']['input'];
  provider: Scalars['String']['input'];
  state: Scalars['String']['input'];
};


export type MutationOAuth_LoginArgs = {
  code: Scalars['String']['input'];
  provider: Scalars['String']['input'];
  state?: InputMaybe<Scalars['String']['input']>;
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

export type Session = {
  id: Scalars['ID']['output'];
  loginMethod: Scalars['String']['output'];
  userEmail: Scalars['String']['output'];
  userId: Scalars['ID']['output'];
  userVerified: Scalars['Boolean']['output'];
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
  containers: Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersIn>>;
  displayName?: InputMaybe<Scalars['String']['input']>;
  freeze?: InputMaybe<Scalars['Boolean']['input']>;
  hpa?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecHpaIn>;
  intercept?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecInterceptIn>;
  nodeSelector?: InputMaybe<Scalars['Map']['input']>;
  region?: InputMaybe<Scalars['String']['input']>;
  replicas?: InputMaybe<Scalars['Int']['input']>;
  serviceAccount?: InputMaybe<Scalars['String']['input']>;
  services?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecServicesIn>>>;
  tolerations?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecTolerationsIn>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersIn = {
  args?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  command?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  env?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvIn>>>;
  envFrom?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromIn>>>;
  image: Scalars['String']['input'];
  imagePullPolicy?: InputMaybe<Scalars['String']['input']>;
  livenessProbe?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeIn>;
  name: Scalars['String']['input'];
  readinessProbe?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeIn>;
  resourceCpu?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceCpuIn>;
  resourceMemory?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceMemoryIn>;
  volumes?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesIn>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvIn = {
  key: Scalars['String']['input'];
  optional?: InputMaybe<Scalars['Boolean']['input']>;
  refKey?: InputMaybe<Scalars['String']['input']>;
  refName?: InputMaybe<Scalars['String']['input']>;
  type?: InputMaybe<Scalars['String']['input']>;
  value?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersEnvFromIn = {
  refName: Scalars['String']['input'];
  type: Scalars['String']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeIn = {
  failureThreshold?: InputMaybe<Scalars['Int']['input']>;
  httpGet?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeHttpGetIn>;
  initialDelay?: InputMaybe<Scalars['Int']['input']>;
  interval?: InputMaybe<Scalars['Int']['input']>;
  shell?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeShellIn>;
  tcp?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeTcpIn>;
  type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeType;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeHttpGetIn = {
  httpHeaders?: InputMaybe<Scalars['Map']['input']>;
  path: Scalars['String']['input'];
  port: Scalars['Int']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeShellIn = {
  command?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersLivenessProbeTcpIn = {
  port: Scalars['Int']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeIn = {
  failureThreshold?: InputMaybe<Scalars['Int']['input']>;
  httpGet?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeHttpGetIn>;
  initialDelay?: InputMaybe<Scalars['Int']['input']>;
  interval?: InputMaybe<Scalars['Int']['input']>;
  shell?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeShellIn>;
  tcp?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeTcpIn>;
  type: Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeType;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeHttpGetIn = {
  httpHeaders?: InputMaybe<Scalars['Map']['input']>;
  path: Scalars['String']['input'];
  port: Scalars['Int']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeShellIn = {
  command?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersReadinessProbeTcpIn = {
  port: Scalars['Int']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceCpuIn = {
  max?: InputMaybe<Scalars['String']['input']>;
  min?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersResourceMemoryIn = {
  max?: InputMaybe<Scalars['String']['input']>;
  min?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesIn = {
  items?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesItemsIn>>>;
  mountPath: Scalars['String']['input'];
  refName: Scalars['String']['input'];
  type: Scalars['String']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecContainersVolumesItemsIn = {
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

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecInterceptIn = {
  enabled: Scalars['Boolean']['input'];
  toDevice: Scalars['String']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecServicesIn = {
  name?: InputMaybe<Scalars['String']['input']>;
  port: Scalars['Int']['input'];
  targetPort?: InputMaybe<Scalars['Int']['input']>;
  type?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_AppSpecTolerationsIn = {
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

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecIn = {
  inputs?: InputMaybe<Scalars['Map']['input']>;
  mresKind: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMresKindIn;
  msvcRef: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMsvcRefIn;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMresKindIn = {
  kind: Scalars['String']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedResourceSpecMsvcRefIn = {
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

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecIn = {
  inputs?: InputMaybe<Scalars['Map']['input']>;
  msvcKind: Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecMsvcKindIn;
  nodeSelector?: InputMaybe<Scalars['Map']['input']>;
  region?: InputMaybe<Scalars['String']['input']>;
  tolerations?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecTolerationsIn>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecMsvcKindIn = {
  apiVersion: Scalars['String']['input'];
  kind?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_ManagedServiceSpecTolerationsIn = {
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
  routes?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRoutesIn>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecBasicAuthIn = {
  enabled: Scalars['Boolean']['input'];
  secretName?: InputMaybe<Scalars['String']['input']>;
  username?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecCorsIn = {
  allowCredentials?: InputMaybe<Scalars['Boolean']['input']>;
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  origins?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecHttpsIn = {
  clusterIssuer?: InputMaybe<Scalars['String']['input']>;
  enabled: Scalars['Boolean']['input'];
  forceRedirect?: InputMaybe<Scalars['Boolean']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRateLimitIn = {
  connections?: InputMaybe<Scalars['Int']['input']>;
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  rpm?: InputMaybe<Scalars['Int']['input']>;
  rps?: InputMaybe<Scalars['Int']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Crds__V1_RouterSpecRoutesIn = {
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

export type Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecIn = {
  offset: Scalars['Int']['input'];
  ports?: InputMaybe<Array<InputMaybe<Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecPortsIn>>>;
  serverName: Scalars['String']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Wireguard__V1_DeviceSpecPortsIn = {
  port?: InputMaybe<Scalars['Int']['input']>;
  targetPort?: InputMaybe<Scalars['Int']['input']>;
};

export type HarborRobotUserIn = {
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Artifacts__V1_HarborUserAccountSpecIn>;
};

export type Github_Com__Kloudlite__Operator__Apis__Artifacts__V1_HarborUserAccountSpecIn = {
  accountName: Scalars['String']['input'];
  enabled?: InputMaybe<Scalars['Boolean']['input']>;
  harborProjectName: Scalars['String']['input'];
  permissions?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  targetSecret?: InputMaybe<Scalars['String']['input']>;
};

export type HarborPermission =
  | 'PullRepository'
  | 'PushRepository';

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
  kind?: InputMaybe<Scalars['String']['input']>;
  metadata: MetadataIn;
  spec?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecIn>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecIn = {
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

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecAgentHelmValuesRefIn = {
  key: Scalars['String']['input'];
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecCredentialsRefIn = {
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_ClusterSpecOperatorsHelmValuesRefIn = {
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

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecIn = {
  awsNodeConfig?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigIn>;
  maxCount: Scalars['Int']['input'];
  minCount: Scalars['Int']['input'];
  targetCount: Scalars['Int']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigIn = {
  imageId?: InputMaybe<Scalars['String']['input']>;
  isGpu?: InputMaybe<Scalars['Boolean']['input']>;
  onDemandSpecs?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigOnDemandSpecsIn>;
  provisionMode: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigProvisionMode;
  region?: InputMaybe<Scalars['String']['input']>;
  spotSpecs?: InputMaybe<Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigSpotSpecsIn>;
  vpc?: InputMaybe<Scalars['String']['input']>;
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigOnDemandSpecsIn = {
  instanceType: Scalars['String']['input'];
};

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodePoolSpecAwsNodeConfigSpotSpecsIn = {
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

export type CursorPagination = {
  after?: Maybe<Scalars['String']['output']>;
  before?: Maybe<Scalars['String']['output']>;
  first?: Maybe<Scalars['Int']['output']>;
  last?: Maybe<Scalars['Int']['output']>;
  orderBy?: Maybe<Scalars['String']['output']>;
  sortDirection?: Maybe<CursorPaginationSortDirection>;
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

export type Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecIn = {
  clusterName?: InputMaybe<Scalars['String']['input']>;
  nodePoolName?: InputMaybe<Scalars['String']['input']>;
  nodeType: Github_Com__Kloudlite__Operator__Apis__Clusters__V1_NodeSpecNodeType;
  taints?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
};

export type Github_Com__Kloudlite__Operator__Pkg__Operator_Check = {
  generation?: Maybe<Scalars['Int']['output']>;
  message?: Maybe<Scalars['String']['output']>;
  status: Scalars['Boolean']['output'];
};

export type HarborProject = {
  accountName: Scalars['String']['output'];
  creationTime: Scalars['Date']['output'];
  credentials: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_HarborCredentials;
  harborProjectName: Scalars['String']['output'];
  id: Scalars['String']['output'];
  recordVersion: Scalars['Int']['output'];
  updateTime: Scalars['Date']['output'];
};

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_HarborCredentials = {
  password: Scalars['String']['output'];
  username: Scalars['String']['output'];
};

export type HarborProjectIn = {
  accountName: Scalars['String']['input'];
  credentials: Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_HarborCredentialsIn;
  harborProjectName: Scalars['String']['input'];
};

export type Kloudlite_Io__Apps__Container___Registry__Internal__Domain__Entities_HarborCredentialsIn = {
  password: Scalars['String']['input'];
  username: Scalars['String']['input'];
};

export type Kloudlite_Io__Apps__Infra__Internal__Entities_HelmStatusVal = {
  isReady?: Maybe<Scalars['Boolean']['output']>;
  message: Scalars['String']['output'];
};

export type MatchFilter = {
  array?: Maybe<Array<Scalars['Any']['output']>>;
  exact?: Maybe<Scalars['Any']['output']>;
  matchType: MatchFilterMatchType;
  regex?: Maybe<Scalars['String']['output']>;
};

export type Membership = {
  accountName: Scalars['String']['output'];
  role: Scalars['String']['output'];
  userId: Scalars['String']['output'];
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
