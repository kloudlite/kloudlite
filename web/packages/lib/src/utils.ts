import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Formats a workspace display name as {ownedBy}/{workspaceName}
 * @param owner - The owner/username from spec.ownedBy
 * @param name - The workspace name
 * @returns Formatted string like "username/workspace-name"
 */
export function formatWorkspaceName(owner: string, name: string): string {
  return `${owner}/${name}`
}

/**
 * Formats a resource name by extracting username and resource name from
 * `{username}--{resourceName}` into `{username}/{resourceName}`.
 */
export function formatResourceName(fullName: string): string {
  const parts = fullName.split('--')
  if (parts.length === 2) {
    return `${parts[0]}/${parts[1]}`
  }
  return fullName
}

/**
 * Extracts only the resource name portion from `{username}--{resourceName}`.
 */
export function getResourceName(fullName: string): string {
  const parts = fullName.split('--')
  if (parts.length === 2) {
    return parts[1]
  }
  return fullName
}

/**
 * Extracts the owner portion from `{username}--{resourceName}`.
 */
export function getResourceOwner(fullName: string): string | null {
  const parts = fullName.split('--')
  if (parts.length === 2) {
    return parts[0]
  }
  return null
}
