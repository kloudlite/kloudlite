/**
 * Common Kubernetes type definitions
 */

export interface ObjectMeta {
  name?: string;
  namespace?: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  resourceVersion?: string;
  creationTimestamp?: string;
  deletionTimestamp?: string;
  finalizers?: string[];
  uid?: string;
  generation?: number;
  ownerReferences?: OwnerReference[];
}

export interface OwnerReference {
  apiVersion: string;
  kind: string;
  name: string;
  uid: string;
  controller?: boolean;
  blockOwnerDeletion?: boolean;
}

export interface TypeMeta {
  apiVersion: string;
  kind: string;
}

export interface ListMeta {
  resourceVersion?: string;
  continue?: string;
  remainingItemCount?: number;
}

export interface K8sResource<TSpec = any, TStatus = any> {
  apiVersion: string;
  kind: string;
  metadata: ObjectMeta;
  spec?: TSpec;
  status?: TStatus;
}

export interface K8sList<T> {
  apiVersion: string;
  kind: string;
  metadata: ListMeta;
  items: T[];
}

export interface Condition {
  type: string;
  status: 'True' | 'False' | 'Unknown';
  lastTransitionTime?: string;
  lastUpdateTime?: string;
  reason?: string;
  message?: string;
}

export interface ResourceRequirements {
  limits?: {
    cpu?: string;
    memory?: string;
    [key: string]: string | undefined;
  };
  requests?: {
    cpu?: string;
    memory?: string;
    [key: string]: string | undefined;
  };
}

export interface LabelSelector {
  matchLabels?: Record<string, string>;
  matchExpressions?: LabelSelectorRequirement[];
}

export interface LabelSelectorRequirement {
  key: string;
  operator: 'In' | 'NotIn' | 'Exists' | 'DoesNotExist';
  values?: string[];
}

export interface LocalObjectReference {
  name: string;
}

export interface EnvVar {
  name: string;
  value?: string;
  valueFrom?: EnvVarSource;
}

export interface EnvVarSource {
  configMapKeyRef?: ConfigMapKeySelector;
  secretKeyRef?: SecretKeySelector;
  fieldRef?: ObjectFieldSelector;
  resourceFieldRef?: ResourceFieldSelector;
}

export interface ConfigMapKeySelector {
  name: string;
  key: string;
  optional?: boolean;
}

export interface SecretKeySelector {
  name: string;
  key: string;
  optional?: boolean;
}

export interface ObjectFieldSelector {
  apiVersion?: string;
  fieldPath: string;
}

export interface ResourceFieldSelector {
  containerName?: string;
  resource: string;
  divisor?: string;
}

export interface VolumeMount {
  name: string;
  mountPath: string;
  subPath?: string;
  readOnly?: boolean;
}

export interface Volume {
  name: string;
  configMap?: ConfigMapVolumeSource;
  secret?: SecretVolumeSource;
  persistentVolumeClaim?: PersistentVolumeClaimVolumeSource;
  emptyDir?: EmptyDirVolumeSource;
  hostPath?: HostPathVolumeSource;
}

export interface ConfigMapVolumeSource {
  name: string;
  items?: KeyToPath[];
  defaultMode?: number;
  optional?: boolean;
}

export interface SecretVolumeSource {
  secretName: string;
  items?: KeyToPath[];
  defaultMode?: number;
  optional?: boolean;
}

export interface PersistentVolumeClaimVolumeSource {
  claimName: string;
  readOnly?: boolean;
}

export interface EmptyDirVolumeSource {
  medium?: string;
  sizeLimit?: string;
}

export interface HostPathVolumeSource {
  path: string;
  type?: string;
}

export interface KeyToPath {
  key: string;
  path: string;
  mode?: number;
}

export interface Toleration {
  key?: string;
  operator?: 'Exists' | 'Equal';
  value?: string;
  effect?: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute';
  tolerationSeconds?: number;
}

export interface Affinity {
  nodeAffinity?: NodeAffinity;
  podAffinity?: PodAffinity;
  podAntiAffinity?: PodAntiAffinity;
}

export interface NodeAffinity {
  requiredDuringSchedulingIgnoredDuringExecution?: NodeSelector;
  preferredDuringSchedulingIgnoredDuringExecution?: PreferredSchedulingTerm[];
}

export interface NodeSelector {
  nodeSelectorTerms: NodeSelectorTerm[];
}

export interface NodeSelectorTerm {
  matchExpressions?: NodeSelectorRequirement[];
  matchFields?: NodeSelectorRequirement[];
}

export interface NodeSelectorRequirement {
  key: string;
  operator: 'In' | 'NotIn' | 'Exists' | 'DoesNotExist' | 'Gt' | 'Lt';
  values?: string[];
}

export interface PreferredSchedulingTerm {
  weight: number;
  preference: NodeSelectorTerm;
}

export interface PodAffinity {
  requiredDuringSchedulingIgnoredDuringExecution?: PodAffinityTerm[];
  preferredDuringSchedulingIgnoredDuringExecution?: WeightedPodAffinityTerm[];
}

export interface PodAntiAffinity {
  requiredDuringSchedulingIgnoredDuringExecution?: PodAffinityTerm[];
  preferredDuringSchedulingIgnoredDuringExecution?: WeightedPodAffinityTerm[];
}

export interface PodAffinityTerm {
  labelSelector?: LabelSelector;
  namespaces?: string[];
  topologyKey: string;
}

export interface WeightedPodAffinityTerm {
  weight: number;
  podAffinityTerm: PodAffinityTerm;
}
