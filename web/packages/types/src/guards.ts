import type { Composition } from './composition'
import type { Environment } from './environment'
import type { MachineType } from './machine'
import type { K8sService } from './service'
import type { ServiceIntercept } from './serviceintercept'
import type { UserPreferences } from './user-preferences'
import type { WorkMachine } from './work-machine'
import type { Workspace, PackageRequest } from './workspace'

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function hasString(value: Record<string, unknown>, key: string): boolean {
  return typeof value[key] === 'string' && value[key].length > 0
}

/**
 * Runtime guard for `Workspace`.
 */
export function isWorkspace(value: unknown): value is Workspace {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name') && hasString(metadata, 'namespace') && hasString(spec, 'ownedBy')
}

/**
 * Runtime guard for `Environment`.
 */
export function isEnvironment(value: unknown): value is Environment {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name') && typeof spec.activated === 'boolean'
}

/**
 * Runtime guard for `PackageRequest`.
 */
export function isPackageRequest(value: unknown): value is PackageRequest {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name') && Array.isArray(spec.packages) && hasString(spec, 'workspaceRef')
}

/**
 * Runtime guard for `WorkMachine`.
 */
export function isWorkMachine(value: unknown): value is WorkMachine {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name') && hasString(spec, 'ownedBy') && hasString(spec, 'targetNamespace')
}

/**
 * Runtime guard for `MachineType`.
 */
export function isMachineType(value: unknown): value is MachineType {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name')
}

/**
 * Runtime guard for `Composition`.
 */
export function isComposition(value: unknown): value is Composition {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name') && hasString(spec, 'displayName') && hasString(spec, 'composeContent')
}

/**
 * Runtime guard for `K8sService`.
 */
export function isK8sService(value: unknown): value is K8sService {
  if (!isRecord(value)) return false
  return (
    hasString(value, 'name') &&
    hasString(value, 'namespace') &&
    hasString(value, 'type') &&
    Array.isArray(value.ports)
  )
}

/**
 * Runtime guard for `ServiceIntercept`.
 */
export function isServiceIntercept(value: unknown): value is ServiceIntercept {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name') && isRecord(spec.workspaceRef) && isRecord(spec.serviceRef)
}

/**
 * Runtime guard for `UserPreferences`.
 */
export function isUserPreferences(value: unknown): value is UserPreferences {
  if (!isRecord(value)) return false
  const metadata = value.metadata
  const spec = value.spec
  if (!isRecord(metadata) || !isRecord(spec)) return false
  return hasString(metadata, 'name')
}
