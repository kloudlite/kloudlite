/**
 * Kubernetes repositories index
 */

export * from './base';
export * from './workspace.repository';
export * from './environment.repository';
export * from './workmachine.repository';
export * from './machinetype.repository';
export * from './user.repository';
export * from './userpreferences.repository';
export * from './snapshot.repository';
export * from './packagerequest.repository';
export * from './service.repository';
export * from './configmap.repository';
export * from './secret.repository';

// Export singleton instances
export { workspaceRepository } from './workspace.repository';
export { environmentRepository } from './environment.repository';
export { workMachineRepository } from './workmachine.repository';
export { machineTypeRepository } from './machinetype.repository';
export { userRepository } from './user.repository';
export { userPreferencesRepository } from './userpreferences.repository';
export { snapshotRepository } from './snapshot.repository';
export { packageRequestRepository } from './packagerequest.repository';
export { serviceRepository } from './service.repository';
export { configMapRepository } from './configmap.repository';
export { secretRepository } from './secret.repository';
