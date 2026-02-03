import { getK8sClient } from "../client";
import { parseK8sError } from "../errors";
import type { V1Secret } from "@kubernetes/client-node";

/**
 * Repository for standard Kubernetes Secrets (core/v1)
 */
class SecretRepository {
  private client = getK8sClient();

  /**
   * List Secrets in a namespace
   */
  async list(namespace: string): Promise<V1Secret[]> {
    try {
      const response = await this.client.core.listNamespacedSecret({
        namespace,
      });
      return response.items || [];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get a single Secret by name
   */
  async get(namespace: string, name: string): Promise<V1Secret> {
    try {
      const response = await this.client.core.readNamespacedSecret({
        namespace,
        name,
      });
      return response;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Create a Secret
   */
  async create(namespace: string, secret: V1Secret): Promise<V1Secret> {
    try {
      const response = await this.client.core.createNamespacedSecret({
        namespace,
        body: secret,
      });
      return response;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update a Secret
   */
  async update(namespace: string, name: string, secret: V1Secret): Promise<V1Secret> {
    try {
      const response = await this.client.core.replaceNamespacedSecret({
        namespace,
        name,
        body: secret,
      });
      return response;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Delete a Secret
   */
  async delete(namespace: string, name: string): Promise<void> {
    try {
      await this.client.core.deleteNamespacedSecret({
        namespace,
        name,
      });
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const secretRepository = new SecretRepository();
export { SecretRepository };
