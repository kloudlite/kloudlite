import { BaseRepository, type ListOptions } from './base';
import type { Snapshot, SnapshotList, SnapshotState } from '../types/snapshot';
import { buildLabelSelector } from '../utils';
import { parseK8sError } from '../errors';

/**
 * Snapshot repository for managing Snapshot custom resources
 * Snapshots are namespaced resources
 */
export class SnapshotRepository extends BaseRepository<Snapshot> {
  constructor() {
    super('snapshots.kloudlite.io', 'v1', 'snapshots', true); // true = namespaced
  }

  /**
   * List snapshots by environment name
   */
  async listByEnvironment(namespace: string, envName: string, options?: ListOptions): Promise<SnapshotList> {
    return this.list(namespace, {
      ...options,
      labelSelector: buildLabelSelector({ 'snapshots.kloudlite.io/environment': envName }),
    }) as Promise<SnapshotList>;
  }

  /**
   * List snapshots by workspace name
   */
  async listByWorkspace(namespace: string, workspaceName: string, options?: ListOptions): Promise<SnapshotList> {
    return this.list(namespace, {
      ...options,
      labelSelector: buildLabelSelector({ 'snapshots.kloudlite.io/workspace': workspaceName }),
    }) as Promise<SnapshotList>;
  }

  /**
   * List snapshots by owner
   */
  async listByOwner(namespace: string, owner: string, options?: ListOptions): Promise<SnapshotList> {
    return this.list(namespace, {
      ...options,
      labelSelector: buildLabelSelector({ 'kloudlite.io/owned-by': owner }),
    }) as Promise<SnapshotList>;
  }

  /**
   * List all snapshots across all namespaces
   */
  async listAll(options?: ListOptions): Promise<SnapshotList> {
    return this.list('', options) as Promise<SnapshotList>;
  }

  /**
   * List snapshots by state
   */
  async listByState(namespace: string, state: SnapshotState, options?: ListOptions): Promise<SnapshotList> {
    try {
      const all = await this.list(namespace, options);

      // Filter by state (client-side since it's in status)
      const filtered = all.items.filter(snapshot => snapshot.status?.state === state);

      return {
        ...all,
        items: filtered,
      } as SnapshotList;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get snapshot size information
   */
  async getSize(namespace: string, name: string): Promise<{ bytes: number; human: string } | null> {
    try {
      const snapshot = await this.get(namespace, name);

      if (!snapshot.status?.sizeBytes) {
        return null;
      }

      return {
        bytes: snapshot.status.sizeBytes,
        human: snapshot.status.sizeHuman || `${snapshot.status.sizeBytes} bytes`,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get snapshot lineage (parent chain)
   */
  async getLineage(namespace: string, name: string): Promise<string[]> {
    try {
      const snapshot = await this.get(namespace, name);
      return snapshot.status?.lineage || [];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if snapshot has parent
   */
  async hasParent(namespace: string, name: string): Promise<boolean> {
    try {
      const snapshot = await this.get(namespace, name);
      return !!snapshot.spec?.parentSnapshot;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get child snapshots (snapshots that have this one as parent)
   */
  async getChildren(namespace: string, name: string, options?: ListOptions): Promise<SnapshotList> {
    try {
      const all = await this.list(namespace, options);

      // Filter for snapshots that have this snapshot as parent
      const children = all.items.filter(
        snapshot => snapshot.spec?.parentSnapshot === name
      );

      return {
        ...all,
        items: children,
      } as SnapshotList;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update retention policy
   */
  async updateRetentionPolicy(
    namespace: string,
    name: string,
    retentionPolicy: Snapshot['spec']['retentionPolicy']
  ): Promise<Snapshot> {
    try {
      const snapshot = await this.get(namespace, name);

      snapshot.spec = {
        ...snapshot.spec!,
        retentionPolicy,
      };

      return await this.update(namespace, name, snapshot);
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const snapshotRepository = new SnapshotRepository();
