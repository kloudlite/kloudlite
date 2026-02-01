/**
 * Native Kubernetes resource type definitions
 */

import type {
  K8sResource,
  K8sList,
  ObjectMeta,
  ResourceRequirements,
  EnvVar,
  VolumeMount,
  Volume,
  Toleration,
  Affinity,
  LabelSelector,
  Condition,
} from './common';

// Pod Types
export interface Pod extends K8sResource<PodSpec, PodStatus> {
  kind: 'Pod';
}

export interface PodList extends K8sList<Pod> {
  kind: 'PodList';
}

export interface PodSpec {
  containers: Container[];
  initContainers?: Container[];
  volumes?: Volume[];
  restartPolicy?: 'Always' | 'OnFailure' | 'Never';
  terminationGracePeriodSeconds?: number;
  activeDeadlineSeconds?: number;
  dnsPolicy?: 'ClusterFirst' | 'ClusterFirstWithHostNet' | 'Default' | 'None';
  nodeSelector?: Record<string, string>;
  serviceAccountName?: string;
  serviceAccount?: string;
  nodeName?: string;
  hostNetwork?: boolean;
  hostPID?: boolean;
  hostIPC?: boolean;
  tolerations?: Toleration[];
  affinity?: Affinity;
  schedulerName?: string;
  priorityClassName?: string;
  priority?: number;
}

export interface Container {
  name: string;
  image: string;
  imagePullPolicy?: 'Always' | 'Never' | 'IfNotPresent';
  command?: string[];
  args?: string[];
  workingDir?: string;
  ports?: ContainerPort[];
  env?: EnvVar[];
  resources?: ResourceRequirements;
  volumeMounts?: VolumeMount[];
  livenessProbe?: Probe;
  readinessProbe?: Probe;
  startupProbe?: Probe;
  securityContext?: SecurityContext;
}

export interface ContainerPort {
  name?: string;
  containerPort: number;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
  hostPort?: number;
  hostIP?: string;
}

export interface Probe {
  exec?: ExecAction;
  httpGet?: HTTPGetAction;
  tcpSocket?: TCPSocketAction;
  initialDelaySeconds?: number;
  timeoutSeconds?: number;
  periodSeconds?: number;
  successThreshold?: number;
  failureThreshold?: number;
}

export interface ExecAction {
  command?: string[];
}

export interface HTTPGetAction {
  path?: string;
  port: number | string;
  host?: string;
  scheme?: 'HTTP' | 'HTTPS';
  httpHeaders?: HTTPHeader[];
}

export interface HTTPHeader {
  name: string;
  value: string;
}

export interface TCPSocketAction {
  port: number | string;
  host?: string;
}

export interface SecurityContext {
  capabilities?: Capabilities;
  privileged?: boolean;
  seLinuxOptions?: SELinuxOptions;
  windowsOptions?: WindowsSecurityContextOptions;
  runAsUser?: number;
  runAsGroup?: number;
  runAsNonRoot?: boolean;
  readOnlyRootFilesystem?: boolean;
  allowPrivilegeEscalation?: boolean;
  procMount?: string;
  seccompProfile?: SeccompProfile;
}

export interface Capabilities {
  add?: string[];
  drop?: string[];
}

export interface SELinuxOptions {
  user?: string;
  role?: string;
  type?: string;
  level?: string;
}

export interface WindowsSecurityContextOptions {
  gmsaCredentialSpecName?: string;
  gmsaCredentialSpec?: string;
  runAsUserName?: string;
  hostProcess?: boolean;
}

export interface SeccompProfile {
  type: string;
  localhostProfile?: string;
}

export interface PodStatus {
  phase?: 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Unknown';
  conditions?: PodCondition[];
  message?: string;
  reason?: string;
  hostIP?: string;
  podIP?: string;
  podIPs?: PodIP[];
  startTime?: string;
  containerStatuses?: ContainerStatus[];
  initContainerStatuses?: ContainerStatus[];
  qosClass?: 'BestEffort' | 'Burstable' | 'Guaranteed';
}

export interface PodCondition extends Condition {
  type: 'PodScheduled' | 'Ready' | 'Initialized' | 'ContainersReady';
}

export interface PodIP {
  ip: string;
}

export interface ContainerStatus {
  name: string;
  state?: ContainerState;
  lastState?: ContainerState;
  ready: boolean;
  restartCount: number;
  image: string;
  imageID: string;
  containerID?: string;
  started?: boolean;
}

export interface ContainerState {
  waiting?: ContainerStateWaiting;
  running?: ContainerStateRunning;
  terminated?: ContainerStateTerminated;
}

export interface ContainerStateWaiting {
  reason?: string;
  message?: string;
}

export interface ContainerStateRunning {
  startedAt?: string;
}

export interface ContainerStateTerminated {
  exitCode: number;
  signal?: number;
  reason?: string;
  message?: string;
  startedAt?: string;
  finishedAt?: string;
  containerID?: string;
}

// ConfigMap Types
export interface ConfigMap extends K8sResource<never, never> {
  kind: 'ConfigMap';
  data?: Record<string, string>;
  binaryData?: Record<string, string>;
}

export interface ConfigMapList extends K8sList<ConfigMap> {
  kind: 'ConfigMapList';
}

// Secret Types
export interface Secret extends K8sResource<never, never> {
  kind: 'Secret';
  type?: string;
  data?: Record<string, string>;
  stringData?: Record<string, string>;
}

export interface SecretList extends K8sList<Secret> {
  kind: 'SecretList';
}

// Service Types
export interface Service extends K8sResource<ServiceSpec, ServiceStatus> {
  kind: 'Service';
}

export interface ServiceList extends K8sList<Service> {
  kind: 'ServiceList';
}

export interface ServiceSpec {
  type?: 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';
  clusterIP?: string;
  clusterIPs?: string[];
  externalIPs?: string[];
  sessionAffinity?: 'ClientIP' | 'None';
  loadBalancerIP?: string;
  loadBalancerSourceRanges?: string[];
  externalName?: string;
  externalTrafficPolicy?: 'Cluster' | 'Local';
  healthCheckNodePort?: number;
  publishNotReadyAddresses?: boolean;
  selector?: Record<string, string>;
  ports?: ServicePort[];
}

export interface ServicePort {
  name?: string;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
  port: number;
  targetPort?: number | string;
  nodePort?: number;
}

export interface ServiceStatus {
  loadBalancer?: LoadBalancerStatus;
  conditions?: Condition[];
}

export interface LoadBalancerStatus {
  ingress?: LoadBalancerIngress[];
}

export interface LoadBalancerIngress {
  ip?: string;
  hostname?: string;
  ports?: PortStatus[];
}

export interface PortStatus {
  port: number;
  protocol: string;
  error?: string;
}

// Deployment Types
export interface Deployment extends K8sResource<DeploymentSpec, DeploymentStatus> {
  kind: 'Deployment';
}

export interface DeploymentList extends K8sList<Deployment> {
  kind: 'DeploymentList';
}

export interface DeploymentSpec {
  replicas?: number;
  selector: LabelSelector;
  template: PodTemplateSpec;
  strategy?: DeploymentStrategy;
  minReadySeconds?: number;
  revisionHistoryLimit?: number;
  paused?: boolean;
  progressDeadlineSeconds?: number;
}

export interface PodTemplateSpec {
  metadata?: ObjectMeta;
  spec?: PodSpec;
}

export interface DeploymentStrategy {
  type?: 'Recreate' | 'RollingUpdate';
  rollingUpdate?: RollingUpdateDeployment;
}

export interface RollingUpdateDeployment {
  maxUnavailable?: number | string;
  maxSurge?: number | string;
}

export interface DeploymentStatus {
  observedGeneration?: number;
  replicas?: number;
  updatedReplicas?: number;
  readyReplicas?: number;
  availableReplicas?: number;
  unavailableReplicas?: number;
  conditions?: DeploymentCondition[];
  collisionCount?: number;
}

export interface DeploymentCondition extends Condition {
  type: 'Available' | 'Progressing' | 'ReplicaFailure';
}

// Node Types
export interface Node extends K8sResource<NodeSpec, NodeStatus> {
  kind: 'Node';
}

export interface NodeList extends K8sList<Node> {
  kind: 'NodeList';
}

export interface NodeSpec {
  podCIDR?: string;
  podCIDRs?: string[];
  providerID?: string;
  unschedulable?: boolean;
  taints?: NodeTaint[];
}

export interface NodeTaint {
  key: string;
  value?: string;
  effect: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute';
}

export interface NodeStatus {
  capacity?: Record<string, string>;
  allocatable?: Record<string, string>;
  phase?: 'Pending' | 'Running' | 'Terminated';
  conditions?: NodeCondition[];
  addresses?: NodeAddress[];
  daemonEndpoints?: NodeDaemonEndpoints;
  nodeInfo?: NodeSystemInfo;
  images?: ContainerImage[];
}

export interface NodeCondition extends Condition {
  type: 'Ready' | 'MemoryPressure' | 'DiskPressure' | 'PIDPressure' | 'NetworkUnavailable';
}

export interface NodeAddress {
  type: 'Hostname' | 'ExternalIP' | 'InternalIP' | 'ExternalDNS' | 'InternalDNS';
  address: string;
}

export interface NodeDaemonEndpoints {
  kubeletEndpoint?: DaemonEndpoint;
}

export interface DaemonEndpoint {
  Port: number;
}

export interface NodeSystemInfo {
  machineID: string;
  systemUUID: string;
  bootID: string;
  kernelVersion: string;
  osImage: string;
  containerRuntimeVersion: string;
  kubeletVersion: string;
  kubeProxyVersion: string;
  operatingSystem: string;
  architecture: string;
}

export interface ContainerImage {
  names: string[];
  sizeBytes?: number;
}

// Namespace Types
export interface Namespace extends K8sResource<NamespaceSpec, NamespaceStatus> {
  kind: 'Namespace';
}

export interface NamespaceList extends K8sList<Namespace> {
  kind: 'NamespaceList';
}

export interface NamespaceSpec {
  finalizers?: string[];
}

export interface NamespaceStatus {
  phase?: 'Active' | 'Terminating';
  conditions?: Condition[];
}
