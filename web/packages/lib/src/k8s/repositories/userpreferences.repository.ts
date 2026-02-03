import { BaseRepository } from './base';
import type { UserPreferences, ResourceReference } from '../types/user';
import { parseK8sError, NotFoundError } from '../errors';

/**
 * UserPreferences repository for managing UserPreferences custom resources
 * UserPreferences are cluster-scoped (not namespaced)
 * The resource name matches the username (User.metadata.name)
 */
export class UserPreferencesRepository extends BaseRepository<UserPreferences> {
  constructor() {
    super('platform.kloudlite.io', 'v1alpha1', 'userpreferences', false); // false = cluster-scoped
  }

  /**
   * Get preferences for a specific user (by username)
   */
  async getByUser(username: string): Promise<UserPreferences> {
    try {
      return await this.get(username);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get preferences or create empty ones if not found
   */
  async getOrCreate(username: string): Promise<UserPreferences> {
    try {
      return await this.get(username);
    } catch (err) {
      if (err instanceof NotFoundError) {
        // Create new empty preferences
        const newPrefs: Partial<UserPreferences> = {
          metadata: {
            name: username,
          },
          spec: {
            pinnedWorkspaces: [],
            pinnedEnvironments: [],
          },
        };

        return await this.create(newPrefs as UserPreferences);
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Add a workspace to pinned list
   */
  async addPinnedWorkspace(username: string, wsRef: ResourceReference): Promise<UserPreferences> {
    try {
      const prefs = await this.getOrCreate(username);

      if (!prefs.spec!.pinnedWorkspaces) {
        prefs.spec!.pinnedWorkspaces = [];
      }

      // Check if already pinned
      const alreadyPinned = prefs.spec!.pinnedWorkspaces.some(
        pw => pw.name === wsRef.name && pw.namespace === wsRef.namespace
      );

      if (!alreadyPinned) {
        prefs.spec!.pinnedWorkspaces.push(wsRef);

        // Update status
        if (!prefs.status) {
          prefs.status = {};
        }
        prefs.status.lastUpdated = new Date().toISOString();
      }

      return await this.update(username, prefs);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Remove a workspace from pinned list
   */
  async removePinnedWorkspace(username: string, wsRef: ResourceReference): Promise<UserPreferences> {
    try {
      const prefs = await this.getByUser(username);

      if (prefs.spec!.pinnedWorkspaces) {
        prefs.spec!.pinnedWorkspaces = prefs.spec!.pinnedWorkspaces.filter(
          pw => !(pw.name === wsRef.name && pw.namespace === wsRef.namespace)
        );

        // Update status
        if (!prefs.status) {
          prefs.status = {};
        }
        prefs.status.lastUpdated = new Date().toISOString();

        return await this.update(username, prefs);
      }

      return prefs;
    } catch (err) {
      if (err instanceof NotFoundError) {
        // No preferences exist, nothing to remove
        return this.getOrCreate(username);
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Add an environment to pinned list
   */
  async addPinnedEnvironment(username: string, envName: string): Promise<UserPreferences> {
    try {
      const prefs = await this.getOrCreate(username);

      if (!prefs.spec!.pinnedEnvironments) {
        prefs.spec!.pinnedEnvironments = [];
      }

      // Check if already pinned
      if (!prefs.spec!.pinnedEnvironments.includes(envName)) {
        prefs.spec!.pinnedEnvironments.push(envName);

        // Update status
        if (!prefs.status) {
          prefs.status = {};
        }
        prefs.status.lastUpdated = new Date().toISOString();
      }

      return await this.update(username, prefs);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Remove an environment from pinned list
   */
  async removePinnedEnvironment(username: string, envName: string): Promise<UserPreferences> {
    try {
      const prefs = await this.getByUser(username);

      if (prefs.spec!.pinnedEnvironments) {
        prefs.spec!.pinnedEnvironments = prefs.spec!.pinnedEnvironments.filter(
          pe => pe !== envName
        );

        // Update status
        if (!prefs.status) {
          prefs.status = {};
        }
        prefs.status.lastUpdated = new Date().toISOString();

        return await this.update(username, prefs);
      }

      return prefs;
    } catch (err) {
      if (err instanceof NotFoundError) {
        // No preferences exist, nothing to remove
        return this.getOrCreate(username);
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Get all pinned workspaces for a user
   */
  async getPinnedWorkspaces(username: string): Promise<ResourceReference[]> {
    try {
      const prefs = await this.getByUser(username);
      return prefs.spec?.pinnedWorkspaces || [];
    } catch (err) {
      if (err instanceof NotFoundError) {
        return [];
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Get all pinned environments for a user
   */
  async getPinnedEnvironments(username: string): Promise<string[]> {
    try {
      const prefs = await this.getByUser(username);
      return prefs.spec?.pinnedEnvironments || [];
    } catch (err) {
      if (err instanceof NotFoundError) {
        return [];
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Check if a workspace is pinned
   */
  async isWorkspacePinned(username: string, wsRef: ResourceReference): Promise<boolean> {
    try {
      const prefs = await this.getByUser(username);
      return prefs.spec?.pinnedWorkspaces?.some(
        pw => pw.name === wsRef.name && pw.namespace === wsRef.namespace
      ) ?? false;
    } catch (err) {
      if (err instanceof NotFoundError) {
        return false;
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Check if an environment is pinned
   */
  async isEnvironmentPinned(username: string, envName: string): Promise<boolean> {
    try {
      const prefs = await this.getByUser(username);
      return prefs.spec?.pinnedEnvironments?.includes(envName) ?? false;
    } catch (err) {
      if (err instanceof NotFoundError) {
        return false;
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Clear all pinned workspaces
   */
  async clearPinnedWorkspaces(username: string): Promise<UserPreferences> {
    try {
      const prefs = await this.getByUser(username);

      prefs.spec!.pinnedWorkspaces = [];

      // Update status
      if (!prefs.status) {
        prefs.status = {};
      }
      prefs.status.lastUpdated = new Date().toISOString();

      return await this.update(username, prefs);
    } catch (err) {
      if (err instanceof NotFoundError) {
        return this.getOrCreate(username);
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Clear all pinned environments
   */
  async clearPinnedEnvironments(username: string): Promise<UserPreferences> {
    try {
      const prefs = await this.getByUser(username);

      prefs.spec!.pinnedEnvironments = [];

      // Update status
      if (!prefs.status) {
        prefs.status = {};
      }
      prefs.status.lastUpdated = new Date().toISOString();

      return await this.update(username, prefs);
    } catch (err) {
      if (err instanceof NotFoundError) {
        return this.getOrCreate(username);
      }
      throw parseK8sError(err);
    }
  }

  /**
   * Set pinned workspaces (replaces existing list)
   */
  async setPinnedWorkspaces(username: string, workspaces: ResourceReference[]): Promise<UserPreferences> {
    try {
      const prefs = await this.getOrCreate(username);

      prefs.spec!.pinnedWorkspaces = workspaces;

      // Update status
      if (!prefs.status) {
        prefs.status = {};
      }
      prefs.status.lastUpdated = new Date().toISOString();

      return await this.update(username, prefs);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Set pinned environments (replaces existing list)
   */
  async setPinnedEnvironments(username: string, environments: string[]): Promise<UserPreferences> {
    try {
      const prefs = await this.getOrCreate(username);

      prefs.spec!.pinnedEnvironments = environments;

      // Update status
      if (!prefs.status) {
        prefs.status = {};
      }
      prefs.status.lastUpdated = new Date().toISOString();

      return await this.update(username, prefs);
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const userPreferencesRepository = new UserPreferencesRepository();
