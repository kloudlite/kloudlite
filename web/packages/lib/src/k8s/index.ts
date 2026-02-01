/**
 * Kubernetes client library for Kloudlite
 *
 * Provides TypeScript client for interacting with Kubernetes API
 * including custom Kloudlite resources (Workspace, Environment, etc.)
 *
 * Usage:
 * ```typescript
 * import { getK8sClient, workspaceRepository } from '@kloudlite/lib/k8s';
 *
 * // Get client
 * const client = getK8sClient();
 *
 * // Use repositories
 * const workspaces = await workspaceRepository.list('my-namespace');
 * ```
 */

export * from './client';
export * from './auth';
export * from './errors';
export * from './utils';
export * from './types';
export * from './repositories';
