import { BaseRepository, type ListOptions } from './base';
import type {
  Environment,
  EnvironmentList,
  EnvironmentState,
  ResourceCount,
} from '../types/environment';
import { buildLabelSelector } from '../utils';
import { parseK8sError, NotFoundError } from '../errors';

/**
 * Environment repository for managing Environment custom resources
 * Implements all CRUD operations plus environment-specific operations
 */
export class EnvironmentRepository extends BaseRepository<Environment> {
  constructor() {
    super('environments.kloudlite.io', 'v1', 'environments', true);
  }

  /**
   * Get environment by target namespace
   * Searches across all namespaces since targetNamespace is unique
   */
  async getByTargetNamespace(targetNamespace: string): Promise<Environment> {
    try {
      // Search across all namespaces using label selector
      const result = await this.list('', {
        labelSelector: buildLabelSelector({ 'kloudlite.io/target-namespace': targetNamespace }),
      });

      if (result.items.length === 0) {
        throw new NotFoundError('Environment', `with target namespace ${targetNamespace}`);
      }

      if (result.items.length > 1) {
        throw new Error(
          `Multiple environments found with target namespace ${targetNamespace}`
        );
      }

      return result.items[0];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List all active environments in a namespace
   */
  async listActive(namespace: string, options?: ListOptions): Promise<EnvironmentList> {
    return this.list(namespace, {
      ...options,
      labelSelector: buildLabelSelector({ 'kloudlite.io/activated': 'true' }),
    }) as Promise<EnvironmentList>;
  }

  /**
   * List all inactive environments in a namespace
   */
  async listInactive(namespace: string, options?: ListOptions): Promise<EnvironmentList> {
    return this.list(namespace, {
      ...options,
      labelSelector: buildLabelSelector({ 'kloudlite.io/activated': 'false' }),
    }) as Promise<EnvironmentList>;
  }

  /**
   * Activate an environment (scales deployments/statefulsets back up)
   */
  async activate(namespace: string, name: string): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      // Check if already activated
      if (environment.spec?.activated) {
        return environment; // Already activated
      }

      // Update spec.activated to true
      environment.spec = {
        ...environment.spec!,
        activated: true,
      };

      // Update status.state to activating
      if (!environment.status) {
        environment.status = {};
      }
      environment.status.state = 'activating';

      // Update the environment (this will trigger controller reconciliation)
      const updated = await this.update(namespace, name, environment);

      // Update status separately
      return await this.updateStatus(namespace, name, updated);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Deactivate an environment (scales deployments/statefulsets to 0)
   */
  async deactivate(namespace: string, name: string): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      // Check if already deactivated
      if (!environment.spec?.activated) {
        return environment; // Already deactivated
      }

      // Update spec.activated to false
      environment.spec = {
        ...environment.spec!,
        activated: false,
      };

      // Update status.state to deactivating
      if (!environment.status) {
        environment.status = {};
      }
      environment.status.state = 'deactivating';

      // Update the environment (this will trigger controller reconciliation)
      const updated = await this.update(namespace, name, environment);

      // Update status separately
      return await this.updateStatus(namespace, name, updated);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update environment state in status (for controller use)
   */
  async updateState(namespace: string, name: string, state: EnvironmentState): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      // Update status.state
      if (!environment.status) {
        environment.status = {};
      }
      environment.status.state = state;

      // Use status subresource to update only status
      return await this.updateStatus(namespace, name, environment);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List environments by owner
   */
  async listByOwner(namespace: string, owner: string, options?: ListOptions): Promise<EnvironmentList> {
    return this.list(namespace, {
      ...options,
      labelSelector: buildLabelSelector({ 'kloudlite.io/owned-by': owner }),
    }) as Promise<EnvironmentList>;
  }

  /**
   * List environments by visibility
   */
  async listByVisibility(
    namespace: string,
    visibility: 'private' | 'shared' | 'open',
    options?: ListOptions
  ): Promise<EnvironmentList> {
    return this.list(namespace, {
      ...options,
      fieldSelector: `spec.visibility=${visibility}`,
    }) as Promise<EnvironmentList>;
  }

  /**
   * List environments shared with a specific user
   */
  async listSharedWith(namespace: string, username: string, options?: ListOptions): Promise<EnvironmentList> {
    // Note: This requires custom filtering as sharedWith is an array
    // For now, we fetch all and filter client-side
    const allEnvironments = await this.list(namespace, options);

    const filteredItems = allEnvironments.items.filter(env =>
      env.spec?.sharedWith?.includes(username)
    );

    return {
      ...allEnvironments,
      items: filteredItems,
    } as EnvironmentList;
  }

  /**
   * Share environment with a user
   */
  async shareWith(namespace: string, name: string, username: string): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      if (!environment.spec!.sharedWith) {
        environment.spec!.sharedWith = [];
      }

      if (!environment.spec!.sharedWith.includes(username)) {
        environment.spec!.sharedWith.push(username);
      }

      // Update visibility to shared if not already
      if (environment.spec!.visibility !== 'shared') {
        environment.spec!.visibility = 'shared';
      }

      return await this.update(namespace, name, environment);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Unshare environment with a user
   */
  async unshareWith(namespace: string, name: string, username: string): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      if (environment.spec!.sharedWith) {
        environment.spec!.sharedWith = environment.spec!.sharedWith.filter(
          u => u !== username
        );

        // If no more shared users, change visibility to private
        if (environment.spec!.sharedWith.length === 0) {
          environment.spec!.visibility = 'private';
        }
      }

      return await this.update(namespace, name, environment);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update resource quotas for an environment
   */
  async updateResourceQuotas(
    namespace: string,
    name: string,
    quotas: Environment['spec']['resourceQuotas']
  ): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      environment.spec!.resourceQuotas = quotas;

      return await this.update(namespace, name, environment);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update network policies for an environment
   */
  async updateNetworkPolicies(
    namespace: string,
    name: string,
    policies: Environment['spec']['networkPolicies']
  ): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      environment.spec!.networkPolicies = policies;

      return await this.update(namespace, name, environment);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update namespace labels
   */
  async updateNamespaceLabels(
    namespace: string,
    name: string,
    labels: Record<string, string>
  ): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      environment.spec!.labels = {
        ...environment.spec!.labels,
        ...labels,
      };

      return await this.update(namespace, name, environment);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update namespace annotations
   */
  async updateNamespaceAnnotations(
    namespace: string,
    name: string,
    annotations: Record<string, string>
  ): Promise<Environment> {
    try {
      const environment = await this.get(namespace, name);

      environment.spec!.annotations = {
        ...environment.spec!.annotations,
        ...annotations,
      };

      return await this.update(namespace, name, environment);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get resource count for an environment
   */
  async getResourceCount(namespace: string, name: string): Promise<ResourceCount | null> {
    try {
      const environment = await this.get(namespace, name);
      return environment.status?.resourceCount || null;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List all environments across all namespaces
   */
  async listAll(options?: ListOptions): Promise<EnvironmentList> {
    // Empty namespace means list across all namespaces
    return this.list('', options) as Promise<EnvironmentList>;
  }

  /**
   * List environments by state
   */
  async listByState(
    namespace: string,
    state: EnvironmentState,
    options?: ListOptions
  ): Promise<EnvironmentList> {
    // This requires custom filtering as state is in status
    // For now, we fetch all and filter client-side
    const allEnvironments = await this.list(namespace, options);

    const filteredItems = allEnvironments.items.filter(
      env => env.status?.state === state
    );

    return {
      ...allEnvironments,
      items: filteredItems,
    } as EnvironmentList;
  }

  /**
   * Check if environment is activated
   */
  async isActivated(namespace: string, name: string): Promise<boolean> {
    try {
      const environment = await this.get(namespace, name);
      return environment.spec?.activated ?? false;
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const environmentRepository = new EnvironmentRepository();
