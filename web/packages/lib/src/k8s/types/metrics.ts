/**
 * Kubernetes metrics types (from metrics-server)
 */

import type { K8sResource, K8sList, ObjectMeta } from './common';

export interface PodMetrics extends K8sResource<never, never> {
  kind: 'PodMetrics';
  timestamp: string;
  window: string;
  containers: ContainerMetrics[];
}

export interface PodMetricsList extends K8sList<PodMetrics> {
  kind: 'PodMetricsList';
}

export interface ContainerMetrics {
  name: string;
  usage: ResourceUsage;
}

export interface ResourceUsage {
  cpu: string;
  memory: string;
}

export interface NodeMetrics extends K8sResource<never, never> {
  kind: 'NodeMetrics';
  timestamp: string;
  window: string;
  usage: ResourceUsage;
}

export interface NodeMetricsList extends K8sList<NodeMetrics> {
  kind: 'NodeMetricsList';
}
