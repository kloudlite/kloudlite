import { BaseRepository, type ListOptions } from './base';
import type {
  Workspace,
  WorkspaceList,
  WorkspacePhase,
} from '../types/workspace';
import { buildLabelSelector } from '../utils';
import { parseK8sError } from '../errors';

/**
 * Workspace repository for managing Workspace custom resources
 * Implements all CRUD operations plus workspace-specific operations
 */
export class WorkspaceRepository extends BaseRepository<Workspace> {
  constructor() {
    super('workspaces.kloudlite.io', 'v1', 'workspaces', true);
  }

  /**
   * Get all workspaces owned by a specific user
   */
  async getByOwner(namespace: string, owner: string, options?: ListOptions): Promise<WorkspaceList> {
    const labelSelector = buildLabelSelector({ 'kloudlite.io/owned-by': owner });
    return this.list(namespace, {
      ...options,
      labelSelector,
    }) as Promise<WorkspaceList>;
  }

  /**
   * Get all workspaces in a WorkMachine's namespace
   * Note: With 1:1 namespace-to-WorkMachine relationship, this returns all workspaces in the namespace
   */
  async getByWorkMachine(namespace: string, _workmachineName: string, options?: ListOptions): Promise<WorkspaceList> {
    // Since there's 1:1 relationship between namespace and WorkMachine,
    // all workspaces in the namespace belong to the same WorkMachine
    return this.list(namespace, options) as Promise<WorkspaceList>;
  }

  /**
   * List all workspaces across all namespaces
   */
  async listAll(options?: ListOptions): Promise<WorkspaceList> {
    // Empty namespace means list across all namespaces
    return this.list('', options) as Promise<WorkspaceList>;
  }

  /**
   * List all active workspaces in a namespace
   */
  async listActive(namespace: string, options?: ListOptions): Promise<WorkspaceList> {
    return this.list(namespace, {
      ...options,
      fieldSelector: 'spec.status=active',
    }) as Promise<WorkspaceList>;
  }

  /**
   * List all suspended workspaces in a namespace
   */
  async listSuspended(namespace: string, options?: ListOptions): Promise<WorkspaceList> {
    return this.list(namespace, {
      ...options,
      fieldSelector: 'spec.status=suspended',
    }) as Promise<WorkspaceList>;
  }

  /**
   * List all archived workspaces in a namespace
   */
  async listArchived(namespace: string, options?: ListOptions): Promise<WorkspaceList> {
    return this.list(namespace, {
      ...options,
      fieldSelector: 'spec.status=archived',
    }) as Promise<WorkspaceList>;
  }

  /**
   * Suspend a workspace (stops the workspace pod to save resources)
   */
  async suspend(namespace: string, name: string): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      // Update spec.status to suspended
      workspace.spec!.status = 'suspended';

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Activate a workspace (starts the workspace pod)
   */
  async activate(namespace: string, name: string): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      // Update spec.status to active
      workspace.spec!.status = 'active';

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Archive a workspace (preserves workspace but makes it inaccessible)
   */
  async archive(namespace: string, name: string): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      // Update spec.status to archived
      workspace.spec!.status = 'archived';

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update workspace phase in status (for controller use)
   */
  async updatePhase(namespace: string, name: string, phase: WorkspacePhase): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      // Update status.phase
      if (!workspace.status) {
        workspace.status = {};
      }
      workspace.status.phase = phase;

      // Use status subresource to update only status
      return await this.updateStatus(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Connect workspace to an environment
   */
  async connectToEnvironment(
    namespace: string,
    name: string,
    environmentName: string,
    environmentNamespace?: string
  ): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      // Set environment connection
      workspace.spec!.environmentConnection = {
        environmentRef: {
          kind: 'Environment',
          apiVersion: 'environments.kloudlite.io/v1',
          name: environmentName,
          namespace: environmentNamespace || namespace,
        },
      };

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Disconnect workspace from environment
   */
  async disconnectFromEnvironment(namespace: string, name: string): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      // Remove environment connection
      workspace.spec!.environmentConnection = undefined;

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update workspace settings
   */
  async updateSettings(
    namespace: string,
    name: string,
    settings: Workspace['spec']['settings']
  ): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      // Merge settings
      workspace.spec!.settings = {
        ...workspace.spec!.settings,
        ...settings,
      };

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get workspace metrics (CPU, memory usage)
   * This is a placeholder - actual implementation would query metrics-server
   */
  async getMetrics(namespace: string, name: string): Promise<{
    cpu: string;
    memory: string;
  } | null> {
    try {
      const workspace = await this.get(namespace, name);

      if (!workspace.status?.resourceUsage) {
        return null;
      }

      return {
        cpu: workspace.status.resourceUsage.cpu || '0',
        memory: workspace.status.resourceUsage.memory || '0',
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get workspace logs from the workspace pod
   * Returns the pod name - actual log fetching should use Pod API
   */
  async getPodName(namespace: string, name: string): Promise<string | null> {
    try {
      const workspace = await this.get(namespace, name);
      return workspace.status?.podName || null;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List workspaces by visibility type
   */
  async listByVisibility(
    namespace: string,
    visibility: 'private' | 'shared' | 'open',
    options?: ListOptions
  ): Promise<WorkspaceList> {
    return this.list(namespace, {
      ...options,
      fieldSelector: `spec.visibility=${visibility}`,
    }) as Promise<WorkspaceList>;
  }

  /**
   * List workspaces shared with a specific user
   */
  async listSharedWith(namespace: string, username: string, options?: ListOptions): Promise<WorkspaceList> {
    // Note: This requires custom filtering as sharedWith is an array
    // For now, we fetch all and filter client-side
    const allWorkspaces = await this.list(namespace, options);

    const filteredItems = allWorkspaces.items.filter(workspace =>
      workspace.spec?.sharedWith?.includes(username)
    );

    return {
      ...allWorkspaces,
      items: filteredItems,
    } as WorkspaceList;
  }

  /**
   * Add user to workspace's sharedWith list
   */
  async shareWith(namespace: string, name: string, username: string): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      if (!workspace.spec!.sharedWith) {
        workspace.spec!.sharedWith = [];
      }

      if (!workspace.spec!.sharedWith.includes(username)) {
        workspace.spec!.sharedWith.push(username);
      }

      // Update visibility to shared if not already
      if (workspace.spec!.visibility !== 'shared') {
        workspace.spec!.visibility = 'shared';
      }

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Remove user from workspace's sharedWith list
   */
  async unshareWith(namespace: string, name: string, username: string): Promise<Workspace> {
    try {
      const workspace = await this.get(namespace, name);

      if (workspace.spec!.sharedWith) {
        workspace.spec!.sharedWith = workspace.spec!.sharedWith.filter(u => u !== username);

        // If no more shared users, change visibility to private
        if (workspace.spec!.sharedWith.length === 0) {
          workspace.spec!.visibility = 'private';
        }
      }

      return await this.update(namespace, name, workspace);
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const workspaceRepository = new WorkspaceRepository();
