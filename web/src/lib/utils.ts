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
