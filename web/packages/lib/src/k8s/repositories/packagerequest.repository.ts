import { BaseRepository, type ListOptions } from './base';
import type { PackageRequest, PackageRequestList, PackageSpec, PackageRequestPhase } from '../types/packages';
import { buildLabelSelector } from '../utils';
import { parseK8sError } from '../errors';

/**
 * PackageRequest repository for managing PackageRequest custom resources
 * PackageRequests are namespaced resources
 */
export class PackageRequestRepository extends BaseRepository<PackageRequest> {
  constructor() {
    super('packages.kloudlite.io', 'v1', 'packagerequests', true); // true = namespaced
  }

  /**
   * Get PackageRequest by workspace reference
   */
  async getByWorkspace(namespace: string, workspaceRef: string): Promise<PackageRequest | null> {
    try {
      const result = await this.list(namespace, {
        labelSelector: buildLabelSelector({ 'kloudlite.io/workspace': workspaceRef }),
      });

      if (result.items.length === 0) {
        return null;
      }

      return result.items[0];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List PackageRequests by phase
   */
  async listByPhase(
    namespace: string,
    phase: PackageRequestPhase,
    options?: ListOptions
  ): Promise<PackageRequestList> {
    try {
      const all = await this.list(namespace, options);

      // Filter by phase (client-side since it's in status)
      const filtered = all.items.filter(pr => pr.status?.phase === phase);

      return {
        ...all,
        items: filtered,
      } as PackageRequestList;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update packages list
   */
  async updatePackages(namespace: string, name: string, packages: PackageSpec[]): Promise<PackageRequest> {
    try {
      const packageRequest = await this.get(namespace, name);

      packageRequest.spec = {
        ...packageRequest.spec!,
        packages,
      };

      return await this.update(namespace, name, packageRequest);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Add a package to the request
   */
  async addPackage(namespace: string, name: string, pkg: PackageSpec): Promise<PackageRequest> {
    try {
      const packageRequest = await this.get(namespace, name);

      if (!packageRequest.spec!.packages) {
        packageRequest.spec!.packages = [];
      }

      // Check if package already exists
      const exists = packageRequest.spec!.packages.some(p => p.name === pkg.name);

      if (!exists) {
        packageRequest.spec!.packages.push(pkg);
      }

      return await this.update(namespace, name, packageRequest);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Remove a package from the request
   */
  async removePackage(namespace: string, name: string, packageName: string): Promise<PackageRequest> {
    try {
      const packageRequest = await this.get(namespace, name);

      if (packageRequest.spec!.packages) {
        packageRequest.spec!.packages = packageRequest.spec!.packages.filter(
          p => p.name !== packageName
        );
      }

      return await this.update(namespace, name, packageRequest);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if a package is in the request
   */
  async hasPackage(namespace: string, name: string, packageName: string): Promise<boolean> {
    try {
      const packageRequest = await this.get(namespace, name);
      return packageRequest.spec?.packages?.some(p => p.name === packageName) ?? false;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get package count
   */
  async getPackageCount(namespace: string, name: string): Promise<number> {
    try {
      const packageRequest = await this.get(namespace, name);
      return packageRequest.status?.packageCount || 0;
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if installation is ready
   */
  async isReady(namespace: string, name: string): Promise<boolean> {
    try {
      const packageRequest = await this.get(namespace, name);
      return packageRequest.status?.phase === 'Ready';
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if installation failed
   */
  async isFailed(namespace: string, name: string): Promise<boolean> {
    try {
      const packageRequest = await this.get(namespace, name);
      return packageRequest.status?.phase === 'Failed';
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get failed package name (if any)
   */
  async getFailedPackage(namespace: string, name: string): Promise<string | null> {
    try {
      const packageRequest = await this.get(namespace, name);
      return packageRequest.status?.failedPackage || null;
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const packageRequestRepository = new PackageRequestRepository();
