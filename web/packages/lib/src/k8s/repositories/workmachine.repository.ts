import { BaseRepository, type ListOptions } from './base';
import type { WorkMachine, WorkMachineList, MachineState } from '../types/workmachine';
import { buildLabelSelector } from '../utils';
import { parseK8sError, NotFoundError } from '../errors';

/**
 * WorkMachine repository for managing WorkMachine custom resources
 * WorkMachines are cluster-scoped (not namespaced)
 */
export class WorkMachineRepository extends BaseRepository<WorkMachine> {
  constructor() {
    super('machines.kloudlite.io', 'v1', 'workmachines', false); // false = cluster-scoped
  }

  /**
   * Get WorkMachine by owner (each user has one WorkMachine)
   */
  async getByOwner(owner: string): Promise<WorkMachine> {
    try {
      const result = await this.list({
        labelSelector: buildLabelSelector({ 'kloudlite.io/owned-by': owner }),
      });

      if (result.items.length === 0) {
        throw new NotFoundError('WorkMachine', `for owner ${owner}`);
      }

      // Each user should have only one WorkMachine
      return result.items[0];
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Start a WorkMachine (sets spec.state to running)
   */
  async start(name: string): Promise<WorkMachine> {
    try {
      // Retry up to 3 times on conflict
      for (let i = 0; i < 3; i++) {
        try {
          const machine = await this.get(name);

          // Update desired state to running
          machine.spec = {
            ...machine.spec!,
            state: 'running',
          };

          return await this.update(name, machine);
        } catch (err: any) {
          // Retry on conflict (409)
          if (err.code === 409 && i < 2) {
            console.log(`Conflict updating machine ${name}, retrying (${i + 1}/3)`);
            continue;
          }
          throw err;
        }
      }

      throw new Error(`Failed to update machine ${name} after 3 retries due to conflicts`);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Stop a WorkMachine (sets spec.state to stopped)
   */
  async stop(name: string): Promise<WorkMachine> {
    try {
      // Retry up to 3 times on conflict
      for (let i = 0; i < 3; i++) {
        try {
          const machine = await this.get(name);

          // Update desired state to stopped
          machine.spec = {
            ...machine.spec!,
            state: 'stopped',
          };

          return await this.update(name, machine);
        } catch (err: any) {
          // Retry on conflict (409)
          if (err.code === 409 && i < 2) {
            console.log(`Conflict updating machine ${name}, retrying (${i + 1}/3)`);
            continue;
          }
          throw err;
        }
      }

      throw new Error(`Failed to update machine ${name} after 3 retries due to conflicts`);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List WorkMachines using a specific machine type
   */
  async listByMachineType(machineType: string, options?: ListOptions): Promise<WorkMachineList> {
    try {
      const all = await this.list(options);

      // Filter by machine type (client-side since it's in spec)
      const filtered = all.items.filter(machine => machine.spec?.machineType === machineType);

      return {
        ...all,
        items: filtered,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List WorkMachines by state
   */
  async listByState(state: MachineState, options?: ListOptions): Promise<WorkMachineList> {
    try {
      const all = await this.list(options);

      // Filter by state (client-side since it's in status)
      const filtered = all.items.filter(machine => machine.status?.state === state);

      return {
        ...all,
        items: filtered,
      };
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * List all running WorkMachines
   */
  async listRunning(options?: ListOptions): Promise<WorkMachineList> {
    return this.listByState('running', options);
  }

  /**
   * List all stopped WorkMachines
   */
  async listStopped(options?: ListOptions): Promise<WorkMachineList> {
    return this.listByState('stopped', options);
  }

  /**
   * Update machine type (triggers instance type change)
   */
  async updateMachineType(name: string, machineType: string): Promise<WorkMachine> {
    try {
      const machine = await this.get(name);

      machine.spec = {
        ...machine.spec!,
        machineType,
      };

      return await this.update(name, machine);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update volume size
   */
  async updateVolumeSize(name: string, volumeSize: number): Promise<WorkMachine> {
    try {
      const machine = await this.get(name);

      machine.spec = {
        ...machine.spec!,
        volumeSize,
      };

      return await this.update(name, machine);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update SSH public keys
   */
  async updateSSHKeys(name: string, sshPublicKeys: string[]): Promise<WorkMachine> {
    try {
      const machine = await this.get(name);

      machine.spec = {
        ...machine.spec!,
        sshPublicKeys,
      };

      return await this.update(name, machine);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Update auto-shutdown configuration
   */
  async updateAutoShutdown(
    name: string,
    autoShutdown: WorkMachine['spec']['autoShutdown']
  ): Promise<WorkMachine> {
    try {
      const machine = await this.get(name);

      machine.spec = {
        ...machine.spec!,
        autoShutdown,
      };

      return await this.update(name, machine);
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Check if machine is running
   */
  async isRunning(name: string): Promise<boolean> {
    try {
      const machine = await this.get(name);
      return machine.status?.state === 'running';
    } catch (err) {
      throw parseK8sError(err);
    }
  }

  /**
   * Get machine metrics (CPU, memory, GPU usage)
   * Returns the resource allocation information
   */
  async getMetrics(name: string): Promise<WorkMachine['status']['allocatedResources'] | null> {
    try {
      const machine = await this.get(name);
      return machine.status?.allocatedResources || null;
    } catch (err) {
      throw parseK8sError(err);
    }
  }
}

// Export singleton instance
export const workMachineRepository = new WorkMachineRepository();
