import { getK8sClient } from "../client";
import { parseK8sError } from "../errors";
import type { V1Service } from "@kubernetes/client-node";
import type { K8sService } from "@kloudlite/types";

/**
 * Transform V1Service to K8sService format
 */
function transformService(v1Service: V1Service): K8sService {
  return {
    name: v1Service.metadata?.name || '',
    namespace: v1Service.metadata?.namespace || '',
    type: v1Service.spec?.type || 'ClusterIP',
    clusterIP: v1Service.spec?.clusterIP || '',
    ports: (v1Service.spec?.ports || []).map((port) => ({
      name: port.name || '',
      port: port.port,
      targetPort: String(port.targetPort || port.port),
      protocol: port.protocol || 'TCP',
    })),
    selector: v1Service.spec?.selector || undefined,
    replicas: 0, // Services don't have replicas - this comes from Deployments
    image: undefined, // Services don't have images - this comes from Pods
  };
}

/**
 * Repository for standard Kubernetes Services (core/v1)
 */
class ServiceRepository {
  private client = getK8sClient();

  /**
   * List services in a namespace
   */
  async list(namespace: string): Promise<K8sService[]> {
    try {
      const response = await this.client.core.listNamespacedService({
        namespace,
      });
      return (response.items || []).map(transformService);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get a single service by name
   */
  async get(namespace: string, name: string): Promise<K8sService> {
    try {
      const response = await this.client.core.readNamespacedService({
        namespace,
        name,
      });
      return transformService(response);
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const serviceRepository = new ServiceRepository();
export { ServiceRepository };
