import type { V1ObjectMeta, V1LabelSelector } from '@kubernetes/client-node';

/**
 * Utility functions for Kubernetes operations
 */

/**
 * Build label selector string from object
 */
export function buildLabelSelector(labels: Record<string, string>): string {
  return Object.entries(labels)
    .map(([key, value]) => `${key}=${value}`)
    .join(',');
}

/**
 * Parse label selector string into object
 */
export function parseLabelSelector(selector: string): Record<string, string> {
  const labels: Record<string, string> = {};
  if (!selector) return labels;

  selector.split(',').forEach((pair) => {
    const [key, value] = pair.split('=');
    if (key && value) {
      labels[key.trim()] = value.trim();
    }
  });

  return labels;
}

/**
 * Check if labels match selector
 */
export function matchesSelector(
  labels: Record<string, string> | undefined,
  selector: V1LabelSelector | undefined
): boolean {
  if (!selector || !selector.matchLabels) return true;
  if (!labels) return false;

  return Object.entries(selector.matchLabels).every(
    ([key, value]) => labels[key] === value
  );
}

/**
 * Generate metadata for new resources
 */
export function generateMetadata(
  name: string,
  namespace?: string,
  labels?: Record<string, string>,
  annotations?: Record<string, string>
): V1ObjectMeta {
  return {
    name,
    namespace,
    labels: labels || {},
    annotations: annotations || {},
  };
}

/**
 * Create JSON Patch for updating resource
 */
export function createJsonPatch(operations: Array<{
  op: 'add' | 'remove' | 'replace' | 'move' | 'copy' | 'test';
  path: string;
  value?: any;
  from?: string;
}>): object[] {
  return operations;
}

/**
 * Create Strategic Merge Patch for updating resource
 */
export function createMergePatch(updates: object): object {
  return updates;
}

/**
 * Extract namespace from resource or use default
 */
export function getNamespace(resource: { metadata?: V1ObjectMeta }, defaultNamespace = 'default'): string {
  return resource.metadata?.namespace || defaultNamespace;
}

/**
 * Extract name from resource
 */
export function getName(resource: { metadata?: V1ObjectMeta }): string | undefined {
  return resource.metadata?.name;
}

/**
 * Check if resource has finalizers
 */
export function hasFinalizers(resource: { metadata?: V1ObjectMeta }): boolean {
  return !!resource.metadata?.finalizers && resource.metadata.finalizers.length > 0;
}

/**
 * Add finalizer to resource metadata
 */
export function addFinalizer(metadata: V1ObjectMeta, finalizer: string): V1ObjectMeta {
  const finalizers = metadata.finalizers || [];
  if (!finalizers.includes(finalizer)) {
    finalizers.push(finalizer);
  }
  return { ...metadata, finalizers };
}

/**
 * Remove finalizer from resource metadata
 */
export function removeFinalizer(metadata: V1ObjectMeta, finalizer: string): V1ObjectMeta {
  const finalizers = (metadata.finalizers || []).filter(f => f !== finalizer);
  return { ...metadata, finalizers };
}

/**
 * Check if resource is being deleted
 */
export function isBeingDeleted(resource: { metadata?: V1ObjectMeta }): boolean {
  return !!resource.metadata?.deletionTimestamp;
}

/**
 * Generate unique name with random suffix
 */
export function generateName(prefix: string): string {
  const suffix = Math.random().toString(36).substring(2, 8);
  return `${prefix}-${suffix}`;
}

/**
 * Convert bytes to human-readable format
 */
export function formatBytes(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }

  return `${value.toFixed(2)} ${units[unitIndex]}`;
}

/**
 * Parse Kubernetes quantity string to number (e.g., "100m" -> 0.1, "2Gi" -> 2147483648)
 */
export function parseQuantity(quantity: string): number {
  const match = quantity.match(/^(\d+(?:\.\d+)?)(.*?)$/);
  if (!match) return 0;

  const [, value, unit] = match;
  const numValue = parseFloat(value);

  const multipliers: Record<string, number> = {
    // CPU units
    m: 0.001,
    // Memory units
    Ki: 1024,
    Mi: 1024 * 1024,
    Gi: 1024 * 1024 * 1024,
    Ti: 1024 * 1024 * 1024 * 1024,
    // Decimal units
    k: 1000,
    M: 1000 * 1000,
    G: 1000 * 1000 * 1000,
    T: 1000 * 1000 * 1000 * 1000,
  };

  return numValue * (multipliers[unit] || 1);
}
