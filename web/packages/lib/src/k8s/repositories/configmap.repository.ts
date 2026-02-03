import { getK8sClient } from "../client";
import { parseK8sError } from "../errors";
import type { V1ConfigMap } from "@kubernetes/client-node";

/**
 * Repository for standard Kubernetes ConfigMaps (core/v1)
 */
class ConfigMapRepository {
  private client = getK8sClient();

  /**
   * List ConfigMaps in a namespace
   */
  async list(namespace: string): Promise<V1ConfigMap[]> {
    try {
      const response = await this.client.core.listNamespacedConfigMap({
        namespace,
      });
      return response.items || [];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get a single ConfigMap by name
   */
  async get(namespace: string, name: string): Promise<V1ConfigMap> {
    try {
      const response = await this.client.core.readNamespacedConfigMap({
        namespace,
        name,
      });
      return response;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Create a ConfigMap
   */
  async create(namespace: string, configMap: V1ConfigMap): Promise<V1ConfigMap> {
    try {
      const response = await this.client.core.createNamespacedConfigMap({
        namespace,
        body: configMap,
      });
      return response;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update a ConfigMap
   */
  async update(namespace: string, name: string, configMap: V1ConfigMap): Promise<V1ConfigMap> {
    try {
      const response = await this.client.core.replaceNamespacedConfigMap({
        namespace,
        name,
        body: configMap,
      });
      return response;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Delete a ConfigMap
   */
  async delete(namespace: string, name: string): Promise<void> {
    try {
      await this.client.core.deleteNamespacedConfigMap({
        namespace,
        name,
      });
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const configMapRepository = new ConfigMapRepository();
export { ConfigMapRepository };
