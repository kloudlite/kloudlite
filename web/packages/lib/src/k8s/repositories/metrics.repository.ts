import { getK8sClient } from '../client';
import { parseK8sError } from '../errors';
import type { PodMetrics } from '../types/metrics';

/**
 * Metrics repository for querying Kubernetes metrics-server API
 * Provides access to CPU and memory metrics for pods and nodes
 */
export class MetricsRepository {
  protected client = getK8sClient();

  /**
   * Get metrics for a specific pod
   * @param namespace - The namespace of the pod
   * @param name - The name of the pod
   */
  async getPodMetrics(namespace: string, name: string): Promise<PodMetrics> {
    try {
      const metricsClient = this.client.metrics;
      // The k8s.Metrics client returns a list of all pods in namespace
      const response = await metricsClient.getPodMetrics(namespace);

      // Find the specific pod by name
      const podMetrics = response.items.find(
        (item: { metadata?: { name?: string } }) => item.metadata?.name === name
      );

      if (!podMetrics) {
        throw new Error(`Pod metrics not found for ${name} in namespace ${namespace}`);
      }

      return podMetrics as unknown as PodMetrics;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List all pod metrics in a namespace
   * @param namespace - The namespace to list pod metrics for
   */
  async listPodMetrics(namespace: string): Promise<{ items: PodMetrics[] }> {
    try {
      const metricsClient = this.client.metrics;
      const response = await metricsClient.getPodMetrics(namespace);
      return { items: response.items as unknown as PodMetrics[] };
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const metricsRepository = new MetricsRepository();
