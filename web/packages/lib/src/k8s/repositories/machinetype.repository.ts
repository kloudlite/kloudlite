import { BaseRepository, type ListOptions } from './base';
import type { MachineType, MachineTypeList, MachineCategory } from '../types/workmachine';
import { buildLabelSelector } from '../utils';
import { parseK8sError, NotFoundError } from '../errors';

/**
 * MachineType repository for managing MachineType custom resources
 * MachineTypes are cluster-scoped (not namespaced)
 */
export class MachineTypeRepository extends BaseRepository<MachineType> {
  constructor() {
    super('machines.kloudlite.io', 'v1', 'machinetypes', false); // false = cluster-scoped
  }

  /**
   * List only active machine types (spec.active = true)
   */
  async listActive(options?: ListOptions): Promise<MachineTypeList> {
    try {
      const all = await this.list(options);

      // Filter for active machine types (client-side since it's in spec)
      const filtered = all.items.filter(mt => mt.spec?.active === true);

      return {
        ...all,
        items: filtered,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get machine types by category
   */
  async getByCategory(category: MachineCategory, options?: ListOptions): Promise<MachineTypeList> {
    try {
      const all = await this.list(options);

      // Filter by category (client-side since it's in spec)
      const filtered = all.items.filter(mt => mt.spec?.category === category);

      return {
        ...all,
        items: filtered,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get the default machine type
   */
  async getDefault(): Promise<MachineType> {
    try {
      const result = await this.list({
        labelSelector: buildLabelSelector({ 'kloudlite.io/machinetype.default': 'true' }),
      });

      if (result.items.length === 0) {
        throw new NotFoundError('MachineType', 'default');
      }

      return result.items[0];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List active machine types sorted by priority
   */
  async listActiveSorted(options?: ListOptions): Promise<MachineTypeList> {
    try {
      const activeList = await this.listActive(options);

      // Sort by priority (lower numbers first)
      const sorted = [...activeList.items].sort((a, b) => {
        const priorityA = a.spec?.priority ?? 100;
        const priorityB = b.spec?.priority ?? 100;
        return priorityA - priorityB;
      });

      return {
        ...activeList,
        items: sorted,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get machine types by category, sorted by priority
   */
  async getByCategorySorted(category: MachineCategory, options?: ListOptions): Promise<MachineTypeList> {
    try {
      const categoryList = await this.getByCategory(category, options);

      // Sort by priority (lower numbers first)
      const sorted = [...categoryList.items].sort((a, b) => {
        const priorityA = a.spec?.priority ?? 100;
        const priorityB = b.spec?.priority ?? 100;
        return priorityA - priorityB;
      });

      return {
        ...categoryList,
        items: sorted,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Activate a machine type (sets spec.active = true)
   */
  async activate(name: string): Promise<MachineType> {
    try {
      const machineType = await this.get(name);

      machineType.spec = {
        ...machineType.spec!,
        active: true,
      };

      return await this.update(name, machineType);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Deactivate a machine type (sets spec.active = false)
   */
  async deactivate(name: string): Promise<MachineType> {
    try {
      const machineType = await this.get(name);

      machineType.spec = {
        ...machineType.spec!,
        active: false,
      };

      return await this.update(name, machineType);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Set a machine type as default
   * Note: This should unset other defaults first in a production system
   */
  async setDefault(name: string): Promise<MachineType> {
    try {
      const machineType = await this.get(name);

      machineType.spec = {
        ...machineType.spec!,
        isDefault: true,
      };

      // Add the default label
      if (!machineType.metadata.labels) {
        machineType.metadata.labels = {};
      }
      machineType.metadata.labels['kloudlite.io/machinetype.default'] = 'true';

      return await this.update(name, machineType);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Unset a machine type as default
   */
  async unsetDefault(name: string): Promise<MachineType> {
    try {
      const machineType = await this.get(name);

      machineType.spec = {
        ...machineType.spec!,
        isDefault: false,
      };

      // Remove the default label
      if (machineType.metadata.labels) {
        delete machineType.metadata.labels['kloudlite.io/machinetype.default'];
      }

      return await this.update(name, machineType);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update priority for a machine type
   */
  async updatePriority(name: string, priority: number): Promise<MachineType> {
    try {
      const machineType = await this.get(name);

      machineType.spec = {
        ...machineType.spec!,
        priority,
      };

      return await this.update(name, machineType);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List machine types with GPU support
   */
  async listWithGPU(options?: ListOptions): Promise<MachineTypeList> {
    try {
      const all = await this.list(options);

      // Filter for machine types with GPU
      const filtered = all.items.filter(mt => mt.spec?.resources?.gpu);

      return {
        ...all,
        items: filtered,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get in-use count for a machine type
   */
  async getInUseCount(name: string): Promise<number> {
    try {
      const machineType = await this.get(name);
      return machineType.status?.inUseCount || 0;
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const machineTypeRepository = new MachineTypeRepository();
